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

	// CORS middleware
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Max-Age", "300")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	// Health check endpoint (no auth required)
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"ok"}`)
	})

	// Test login endpoint (no auth required) - for local testing only
	router.Post("/api/auth/login", handleTestLogin(jwtSecret))

	// Initialize statement service with stub repositories for testing
	// TODO: Connect to actual database and use real repositories
	stmtService := statement.NewStatementService(
		&statement.StatementRepository{},
		&statement.TransactionRepository{},
		&statement.ImportJobRepository{},
	)

	// Protected API routes
	router.Route("/api", func(r chi.Router) {
		authMiddleware := apimiddleware.NewAuthMiddleware(jwtSecret)
		r.Use(authMiddleware.Handler)

		// Setup statement routes
		api.SetupRoutes(r, stmtService, logger)
	})

	// Start server
	logger.WithFields(logrus.Fields{
		"port": port,
	}).Info("starting statement import API server")

	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), router); err != nil {
		logger.WithError(err).Fatal("failed to start server")
	}
}
