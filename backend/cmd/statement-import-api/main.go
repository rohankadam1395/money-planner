package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sirupsen/logrus"

	"money-planner/backend/internal/api"
	apimiddleware "money-planner/backend/internal/api/middleware"
	"money-planner/backend/internal/auth"
	"money-planner/backend/internal/categorization"
	"money-planner/backend/internal/config"
	dbpkg "money-planner/backend/internal/db"
	"money-planner/backend/internal/db/migrations"
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

	// Run migrations
	if err := migrations.RunMigrations(db.GetConnection()); err != nil {
		logger.WithError(err).Fatal("failed to run database migrations")
	}

	// Set up router
	router := chi.NewRouter()

	// Global middleware
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(apimiddleware.NewLoggingMiddleware(logger).Handler)
	router.Use(middleware.Recoverer)

	// CORS middleware
	corsOrigin := os.Getenv("CORS_ORIGIN")
	if corsOrigin == "" {
		corsOrigin = "http://localhost:3000"
	}
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", corsOrigin)
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

	// Initialize auth service
	userRepo := auth.NewUserRepository(db.GetConnection())
	authService := auth.NewAuthService(userRepo, jwtSecret)

	// Auth endpoints (no auth required)
	router.Post("/api/auth/register", RegisterHandler(authService))
	router.Post("/api/auth/login", LoginHandler(authService))

	// Initialize statement service with database-backed repositories
	stmtService := statement.NewStatementService(
		statement.NewStatementRepository(db.GetConnection()),
		statement.NewTransactionRepository(db.GetConnection()),
		statement.NewImportJobRepository(db.GetConnection()),
	)

	// Initialize categorization service
	var categService *categorization.CategorizationService
	categConfig, err := config.LoadMerchantsConfig()
	if err != nil {
		logger.WithError(err).Warn("failed to load merchants config, categorization disabled")
	} else {
		// Create merchant dictionary and confidence scorer
		merchantDict := categorization.NewMerchantDictionary()
		confidencer := categorization.NewConfidenceScorer()
		categService = categorization.NewCategorizationService(merchantDict, confidencer)
		config.LogConfig(categConfig)

		// Load merchants from database into memory
		conn := db.GetConnection()
		rows, err := conn.Query(`
			SELECT m.merchant_name, c.name
			FROM merchant_dictionary m
			JOIN categories c ON m.category_id = c.id
		`)
		if err != nil {
			logger.WithError(err).Warn("failed to load merchants from database")
		} else {
			defer rows.Close()
			merchantCount := 0
			for rows.Next() {
				var merchantName, categoryName string
				if err := rows.Scan(&merchantName, &categoryName); err != nil {
					logger.WithError(err).Warn("failed to scan merchant row")
					continue
				}
				merchantDict.Insert(merchantName, categoryName)
				merchantCount++
			}
			logger.WithField("count", merchantCount).Info("merchants loaded into memory cache")
		}

		logger.Info("categorization service initialized")
	}

	// Protected API routes
	router.Route("/api", func(r chi.Router) {
		authMiddleware := apimiddleware.NewAuthMiddleware(jwtSecret)
		r.Use(authMiddleware.Handler)

		// Setup statement routes
		api.SetupRoutes(r, stmtService, categService, logger)
	})

	// Start server
	logger.WithFields(logrus.Fields{
		"port": port,
	}).Info("starting statement import API server")

	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), router); err != nil {
		logger.WithError(err).Fatal("failed to start server")
	}
}
