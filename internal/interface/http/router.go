package http

import (
	"net/http"
)

// NewRouter configura todas las rutas de la aplicacion.
func NewRouter(
	authHandler *AuthHandler,
	testingHandler *TestingHandler,
	userHandler *UserHandler,
) http.Handler {
	mux := http.NewServeMux()

	// -- Rutas publicas (auth) ---------------------------------------------------
	mux.HandleFunc("GET /register", authHandler.ShowRegister)
	mux.HandleFunc("POST /register", authHandler.Register)
	mux.HandleFunc("GET /login", authHandler.ShowLogin)
	mux.HandleFunc("POST /login", authHandler.Login)
	mux.HandleFunc("POST /logout", authHandler.Logout)

	// -- Redirect root to releases -----------------------------------------------
	mux.Handle("GET /", requireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/releases", http.StatusSeeOther)
	})))

	// -- Releases ----------------------------------------------------------------
	mux.Handle("GET /releases", requireAuth(http.HandlerFunc(testingHandler.ListReleases)))
	mux.Handle("GET /releases/new", requireAuth(http.HandlerFunc(testingHandler.ShowCreateRelease)))
	mux.Handle("POST /releases", requireAuth(http.HandlerFunc(testingHandler.CreateRelease)))
	mux.Handle("GET /releases/{id}", requireAuth(http.HandlerFunc(testingHandler.ShowRelease)))
	mux.Handle("GET /releases/{id}/export", requireAuth(http.HandlerFunc(testingHandler.ExportExcel)))

	// -- TestCases ---------------------------------------------------------------
	mux.Handle("GET /testcases", requireAuth(http.HandlerFunc(testingHandler.ListTestCases)))
	mux.Handle("GET /testcases/report/{reportID}", requireAuth(http.HandlerFunc(testingHandler.ListTestCasesByReport)))
	mux.Handle("GET /testcases/new", requireAuth(http.HandlerFunc(testingHandler.ShowCreateTestCase)))
	mux.Handle("POST /testcases", requireAuth(http.HandlerFunc(testingHandler.CreateTestCase)))
	mux.Handle("GET /testcases/import", requireAuth(http.HandlerFunc(testingHandler.ShowImport)))
	mux.Handle("POST /testcases/import", requireAuth(http.HandlerFunc(testingHandler.ImportMarkdown)))
	mux.Handle("GET /testcases/{id}", requireAuth(http.HandlerFunc(testingHandler.ShowTestCase)))
	mux.Handle("GET /testcases/{id}/edit", requireAuth(http.HandlerFunc(testingHandler.ShowEditTestCase)))
	mux.Handle("PUT /testcases/{id}", requireAuth(http.HandlerFunc(testingHandler.UpdateTestCase)))
	mux.Handle("DELETE /testcases/{id}", requireAuth(http.HandlerFunc(testingHandler.DeleteTestCase)))
	mux.Handle("POST /testcases/{id}/image", requireAuth(http.HandlerFunc(testingHandler.UploadImage)))
	mux.Handle("POST /testcases/{id}/fields", requireAuth(http.HandlerFunc(testingHandler.SaveField)))
	mux.Handle("DELETE /testcases/fields/{fieldID}", requireAuth(http.HandlerFunc(testingHandler.DeleteField)))

	// -- Reports -----------------------------------------------------------------
	mux.Handle("GET /reports", requireAuth(http.HandlerFunc(testingHandler.ListReports)))
	mux.Handle("GET /reports/new", requireAuth(http.HandlerFunc(testingHandler.ShowCreateReport)))
	mux.Handle("POST /reports", requireAuth(http.HandlerFunc(testingHandler.CreateReport)))
	mux.Handle("GET /reports/{id}/edit", requireAuth(http.HandlerFunc(testingHandler.ShowEditReport)))
	mux.Handle("PUT /reports/{id}", requireAuth(http.HandlerFunc(testingHandler.UpdateReport)))
	mux.Handle("GET /reports/{id}/claude-prompt", requireAuth(http.HandlerFunc(testingHandler.ExportClaudePrompt)))

	// -- Executions --------------------------------------------------------------
	mux.Handle("GET /executions/{id}", requireAuth(http.HandlerFunc(testingHandler.ShowExecution)))
	mux.Handle("PATCH /executions/{id}/status", requireAuth(http.HandlerFunc(testingHandler.UpdateExecution)))
	mux.Handle("POST /executions/{id}/compare", requireAuth(http.HandlerFunc(testingHandler.CompareJSON)))

	// -- Knowledge ---------------------------------------------------------------
	mux.Handle("GET /knowledge", requireAuth(http.HandlerFunc(testingHandler.ShowKnowledge)))
	mux.Handle("POST /knowledge", requireAuth(http.HandlerFunc(testingHandler.UpdateKnowledge)))

	// -- Teams -------------------------------------------------------------------
	mux.Handle("GET /teams", requireAuth(http.HandlerFunc(testingHandler.ListTeams)))
	mux.Handle("GET /teams/new", requireAuth(http.HandlerFunc(testingHandler.ShowCreateTeam)))
	mux.Handle("POST /teams", requireAuth(http.HandlerFunc(testingHandler.CreateTeam)))
	mux.Handle("GET /teams/{id}/edit", requireAuth(http.HandlerFunc(testingHandler.ShowEditTeam)))
	mux.Handle("PUT /teams/{id}", requireAuth(http.HandlerFunc(testingHandler.UpdateTeam)))

	// -- Admin Users -------------------------------------------------------------
	mux.Handle("GET /admin/users", requireAdmin(http.HandlerFunc(userHandler.List)))
	mux.Handle("GET /admin/users/new", requireAdmin(http.HandlerFunc(userHandler.ShowCreate)))
	mux.Handle("POST /admin/users", requireAdmin(http.HandlerFunc(userHandler.Create)))
	mux.Handle("GET /admin/users/{id}/edit", requireAdmin(http.HandlerFunc(userHandler.ShowEdit)))
	mux.Handle("PUT /admin/users/{id}", requireAdmin(http.HandlerFunc(userHandler.Update)))
	mux.Handle("DELETE /admin/users/{id}", requireAdmin(http.HandlerFunc(userHandler.Delete)))

	// -- Middlewares globales -----------------------------------------------------
	var handler http.Handler = mux
	handler = MethodOverride(handler)
	handler = RecoverMiddleware(handler)
	handler = LoggingMiddleware(handler)
	handler = RequestIDMiddleware(handler)

	return handler
}

// requireAuth aplica AuthMiddleware a un handler individual.
func requireAuth(next http.Handler) http.Handler {
	return AuthMiddleware(next)
}

// requireAdmin aplica AuthMiddleware + RequireAdmin a un handler individual.
func requireAdmin(next http.Handler) http.Handler {
	return AuthMiddleware(RequireAdmin(next))
}
