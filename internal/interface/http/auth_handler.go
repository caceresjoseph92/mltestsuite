package http

import (
	"net/http"
	"time"

	appauth "mltestsuite/internal/application/auth"
)

// AuthHandler maneja autenticacion (login/logout/register).
type AuthHandler struct {
	authService *appauth.Service
	tmpl        *Renderer
}

// NewAuthHandler crea el handler de autenticacion.
func NewAuthHandler(authService *appauth.Service, tmpl *Renderer) *AuthHandler {
	return &AuthHandler{authService: authService, tmpl: tmpl}
}

// ShowRegister muestra el formulario de registro.
func (h *AuthHandler) ShowRegister(w http.ResponseWriter, r *http.Request) {
	h.tmpl.ExecuteTemplate(w, "auth/register.html", nil)
}

// Register procesa el registro de un nuevo usuario.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	email := r.FormValue("email")
	password := r.FormValue("password")

	if name == "" || email == "" || password == "" {
		h.tmpl.ExecuteTemplate(w, "auth/register.html", map[string]any{
			"Error": "Todos los campos son obligatorios",
			"Name":  name,
			"Email": email,
		})
		return
	}

	u, err := h.authService.Register(r.Context(), name, email, password)
	if err != nil {
		h.tmpl.ExecuteTemplate(w, "auth/register.html", map[string]any{
			"Error": err.Error(),
			"Name":  name,
			"Email": email,
		})
		return
	}

	// Auto-login despues del registro
	token, err := generateToken(u.ID.String(), u.Name, string(u.Role))
	if err != nil {
		http.Error(w, "Error interno", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(24 * time.Hour),
	})

	setFlash(w, "success", "Registro exitoso. Bienvenido/a!")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// ShowLogin muestra el formulario de login.
func (h *AuthHandler) ShowLogin(w http.ResponseWriter, r *http.Request) {
	h.tmpl.ExecuteTemplate(w, "auth/login.html", nil)
}

// Login procesa las credenciales y emite el JWT en una cookie.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")

	u, err := h.authService.Login(r.Context(), email, password)
	if err != nil {
		h.tmpl.ExecuteTemplate(w, "auth/login.html", map[string]any{
			"Error": "Credenciales inválidas",
			"Email": email,
		})
		return
	}

	token, err := generateToken(u.ID.String(), u.Name, string(u.Role))
	if err != nil {
		http.Error(w, "Error interno", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(24 * time.Hour),
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Logout elimina la cookie de sesion.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:    "auth_token",
		Value:   "",
		Path:    "/",
		MaxAge:  -1,
		Expires: time.Unix(0, 0),
	})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
