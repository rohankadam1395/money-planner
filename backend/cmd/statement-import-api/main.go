package main

import (
	"database/sql"
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
	"money-planner/backend/internal/categorization/providers"
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
	router.Post("/api/auth/register", api.RegisterHandler(authService))
	router.Post("/api/auth/login", api.LoginHandler(authService))

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

		// Initialize LLM provider if configured
		llmProvider := os.Getenv("LLM_PROVIDER")
		if llmProvider != "" {
			switch llmProvider {
			case "ollama":
				ollamaURL := os.Getenv("OLLAMA_URL")
				if ollamaURL == "" {
					ollamaURL = "http://localhost:11434"
				}
				ollamaModel := os.Getenv("OLLAMA_MODEL")
				if ollamaModel == "" {
					ollamaModel = "mistral"
				}
				provider := providers.NewOllamaProvider(ollamaURL, ollamaModel)
				categService.WithLLMProvider(provider)
				logger.WithFields(logrus.Fields{
					"provider": "ollama",
					"url":      ollamaURL,
					"model":    ollamaModel,
				}).Info("LLM provider initialized")

			case "claude":
				claudeAPIKey := os.Getenv("ANTHROPIC_API_KEY")
				if claudeAPIKey == "" {
					logger.Warn("ANTHROPIC_API_KEY not set, Claude provider will be unavailable")
				} else {
					claudeModel := os.Getenv("CLAUDE_MODEL")
					provider := providers.NewClaudeProvider(claudeAPIKey, claudeModel)
					categService.WithLLMProvider(provider)
					logger.WithFields(logrus.Fields{
						"provider": "claude",
						"model":    claudeModel,
					}).Info("LLM provider initialized")
				}

			default:
				logger.WithField("provider", llmProvider).Warn("unknown LLM provider, categorization will use rule-based only")
			}
		} else {
			logger.Info("no LLM provider configured (LLM_PROVIDER env var not set), using rule-based categorization only")
		}

		// Load merchants from database into memory
		conn := db.GetConnection()

		// Check if merchants table is empty and auto-seed if needed
		var merchantCount int64
		err = conn.QueryRow("SELECT COUNT(*) FROM merchant_dictionary").Scan(&merchantCount)
		if err != nil {
			logger.WithError(err).Warn("failed to check merchant count")
		} else if merchantCount == 0 {
			logger.Info("merchant dictionary is empty, auto-seeding merchants...")
			if err := seedMerchants(conn, logger); err != nil {
				logger.WithError(err).Warn("failed to auto-seed merchants")
			}
		}

		// Load merchants into memory cache
		rows, err := conn.Query(`
			SELECT m.merchant_name, c.name
			FROM merchant_dictionary m
			JOIN categories c ON m.category_id = c.id
		`)
		if err != nil {
			logger.WithError(err).Warn("failed to load merchants from database")
		} else {
			defer rows.Close()
			loadedCount := 0
			for rows.Next() {
				var merchantName, categoryName string
				if err := rows.Scan(&merchantName, &categoryName); err != nil {
					logger.WithError(err).Warn("failed to scan merchant row")
					continue
				}
				merchantDict.Insert(merchantName, categoryName)
				loadedCount++
			}
			logger.WithField("count", loadedCount).Info("merchants loaded into memory cache")
		}

		logger.Info("categorization service initialized")
	}

	// Protected API routes
	router.Route("/api/v1", func(r chi.Router) {
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

// seedMerchants auto-seeds categories and merchants on first startup
func seedMerchants(conn *sql.DB, logger *logrus.Logger) error {
	// Seed categories first
	categories := []struct {
		name  string
		desc  string
		color string
		icon  string
	}{
		{"Food & Dining", "Restaurants, food delivery, groceries", "#FF6B6B", "🍔"},
		{"Shopping", "Retail, clothing, online marketplaces", "#4ECDC4", "🛍️"},
		{"Transport", "Ride-sharing, fuel, transport", "#45B7D1", "🚗"},
		{"Housing", "Rent, property, home maintenance", "#F7B731", "🏠"},
		{"Utilities", "Electricity, water, internet, phone", "#5F27CD", "💡"},
		{"Entertainment", "Movies, streaming, games, events", "#EE5A6F", "🎬"},
		{"Income", "Salary, freelance, refunds", "#2ECC71", "💰"},
		{"Healthcare", "Medical, pharmacy, gym, insurance", "#FF4757", "🏥"},
		{"Education", "Tuition, courses, books", "#1E90FF", "📚"},
		{"Miscellaneous", "Gifts, charity, other", "#95A5A6", "📌"},
	}

	for _, cat := range categories {
		_, err := conn.Exec(
			`INSERT INTO categories (id, name, description, color, icon, is_predefined, created_at, updated_at)
			 VALUES (gen_random_uuid(), $1, $2, $3, $4, true, NOW(), NOW())
			 ON CONFLICT (name) DO NOTHING`,
			cat.name, cat.desc, cat.color, cat.icon)
		if err != nil {
			logger.WithError(err).Warnf("failed to insert category %s", cat.name)
		}
	}

	// Sample merchants to seed
	merchants := []struct {
		name     string
		category string
	}{
		{"Swiggy", "Food & Dining"},
		{"Zomato", "Food & Dining"},
		{"Amazon", "Shopping"},
		{"Flipkart", "Shopping"},
		{"Uber", "Transport"},
		{"Ola", "Transport"},
		{"Netflix", "Entertainment"},
		{"Spotify", "Entertainment"},
		{"BSNL", "Utilities"},
		{"Airtel", "Utilities"},
		{"Apollo Hospital", "Healthcare"},
		{"Coursera", "Education"},
	}

	for _, m := range merchants {
		_, err := conn.Exec(
			`INSERT INTO merchant_dictionary (id, merchant_name, category_id, source, confidence, match_type, frequency, created_at, updated_at)
			 SELECT gen_random_uuid(), $1, id, 'auto-seed', 100, 'exact', 0, NOW(), NOW()
			 FROM categories WHERE name = $2
			 ON CONFLICT (merchant_name, category_id) DO NOTHING`,
			m.name, m.category)
		if err != nil {
			logger.WithError(err).Warnf("failed to insert merchant %s", m.name)
		}
	}

	logger.Info("auto-seeding completed: categories and sample merchants inserted")
	return nil
}
