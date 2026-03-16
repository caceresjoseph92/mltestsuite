package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"

	appauth "mltestsuite/internal/application/auth"
	apptesting "mltestsuite/internal/application/testing"
	appuser "mltestsuite/internal/application/user"
	"mltestsuite/internal/infrastructure/cloudinary"
	infrapostgres "mltestsuite/internal/infrastructure/postgres"

	httphandler "mltestsuite/internal/interface/http"
)

func main() {
	// Configurar slog
	if os.Getenv("ENV") == "production" {
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	} else {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))
	}

	ctx := context.Background()

	// -- Infraestructura -----------------------------------------------------------
	pool, err := infrapostgres.NewPool(ctx)
	if err != nil {
		log.Fatalf("Error conectando a la base de datos: %v", err)
	}
	defer pool.Close()

	// Cloudinary
	uploader := cloudinary.New(cloudinary.ConfigFromEnv())

	// -- Repositorios (adaptadores) ------------------------------------------------
	userRepo := infrapostgres.NewUserRepository(pool)
	teamRepo := infrapostgres.NewTeamRepository(pool)
	reportRepo := infrapostgres.NewReportRepository(pool)
	testCaseRepo := infrapostgres.NewTestCaseRepository(pool)
	releaseRepo := infrapostgres.NewReleaseRepository(pool)
	execRepo := infrapostgres.NewExecutionRepository(pool)
	knowledgeRepo := infrapostgres.NewKnowledgeRepository(pool)

	// -- Servicios de aplicacion (casos de uso) ------------------------------------
	authService := appauth.NewService(userRepo)
	userService := appuser.NewService(userRepo)
	testingService := apptesting.NewService(teamRepo, reportRepo, testCaseRepo, releaseRepo, execRepo, knowledgeRepo)

	// -- Templates HTML ------------------------------------------------------------
	renderer, err := httphandler.NewRenderer()
	if err != nil {
		log.Fatalf("Error cargando templates: %v", err)
	}

	// -- Handlers HTTP -------------------------------------------------------------
	authHandler := httphandler.NewAuthHandler(authService, renderer)
	testingHandler := httphandler.NewTestingHandler(testingService, userService, uploader, renderer)
	userHandler := httphandler.NewUserHandler(userService, testingService, renderer)

	// -- Router -------------------------------------------------------------------
	router := httphandler.NewRouter(authHandler, testingHandler, userHandler)

	// -- Servidor -----------------------------------------------------------------
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	slog.Info("servidor iniciado", "port", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatalf("Error iniciando servidor: %v", err)
	}
}
