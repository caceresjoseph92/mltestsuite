package http

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"mltestsuite/internal/domain/user"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type contextKey string

const (
	contextKeyUserID   contextKey = "userID"
	contextKeyUserRole contextKey = "userRole"
	contextKeyUserName contextKey = "userName"
	contextKeyReqID    contextKey = "requestID"
)

type Claims struct {
	UserID string `json:"user_id"`
	Name   string `json:"name"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = uuid.New().String()
		}
		ctx := context.WithValue(r.Context(), contextKeyReqID, id)
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(s int) {
	rw.status = s
	rw.ResponseWriter.WriteHeader(s)
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := &responseWriter{ResponseWriter: w, status: 200}
		next.ServeHTTP(ww, r)
		reqID, _ := r.Context().Value(contextKeyReqID).(string)
		slog.Info("request",
			"request_id", reqID,
			"method", r.Method,
			"path", r.URL.Path,
			"status", ww.status,
			"ms", time.Since(start).Milliseconds(),
		)
	})
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := extractToken(r)
		if token == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		claims, err := parseToken(token)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		ctx := context.WithValue(r.Context(), contextKeyUserID, claims.UserID)
		ctx = context.WithValue(ctx, contextKeyUserRole, claims.Role)
		ctx = context.WithValue(ctx, contextKeyUserName, claims.Name)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if GetUserRole(r.Context()) != string(user.RoleAdmin) {
			http.Error(w, "Acceso denegado", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func MethodOverride(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			if m := r.FormValue("_method"); m != "" {
				r.Method = strings.ToUpper(m)
			}
		}
		next.ServeHTTP(w, r)
	})
}

func RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("panic", "error", err)
				http.Error(w, "Error interno", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func GetUserID(ctx context.Context) string {
	s, _ := ctx.Value(contextKeyUserID).(string)
	return s
}

func GetUserRole(ctx context.Context) string {
	s, _ := ctx.Value(contextKeyUserRole).(string)
	return s
}

func GetUserName(ctx context.Context) string {
	s, _ := ctx.Value(contextKeyUserName).(string)
	return s
}

func IsAdmin(ctx context.Context) bool {
	return GetUserRole(ctx) == string(user.RoleAdmin)
}

func extractToken(r *http.Request) string {
	if c, err := r.Cookie("auth_token"); err == nil && c.Value != "" {
		return c.Value
	}
	if b := r.Header.Get("Authorization"); strings.HasPrefix(b, "Bearer ") {
		return strings.TrimPrefix(b, "Bearer ")
	}
	return ""
}

func parseToken(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	t, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil || !t.Valid {
		return nil, err
	}
	return claims, nil
}

func generateToken(userID, name, role string) (string, error) {
	claims := &Claims{
		UserID: userID, Name: name, Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(os.Getenv("JWT_SECRET")))
}
