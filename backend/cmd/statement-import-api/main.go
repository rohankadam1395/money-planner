package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sirupsen/logrus"

	dbpkg "money-planner/backend/internal/db"
	"money-planner/backend/internal/api"
	apimiddleware "money-planner/backend/internal/api/middleware"
	"money-planner/backend/internal/statement"
)

func main() {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Load configuration from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		logger.Fatal("DATABASE_URL environment variable not set")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		logger.Fatal("JWT_SECRET environment variable not set")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Initialize database
	db, err := dbpkg.New(dbURL)
	if err != nil {
		logger.WithError(err).Fatal("failed to initialize database")
	}
	defer db.Close()

	// Set up router
	router := chi.NewRouter()

	// Global middleware
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(apimiddleware.NewLoggingMiddleware(logger).Handler)
	router.Use(middleware.Recoverer)

	// Health check endpoint (no auth required)
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"ok"}`)
	})

	// Initialize statement service (TODO: inject database connection)
	// For now, create with nil repositories as placeholders
	stmtService := &statement.StatementService{}

	// Protected API routes
	router.Route("/api", func(r chi.Router) {
		authMiddleware := apimiddleware.NewAuthMiddleware(jwtSecret)
		r.Use(authMiddleware.Handler)

		// Setup statement routes
		api.SetupRoutes(r, stmtService)
	})

	// Start server
	logger.WithFields(logrus.Fields{
		"port": port,
	}).Info("starting statement import API server")

	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), router); err != nil {
		logger.WithError(err).Fatal("failed to start server")
	}
}
