package api

import (
	"database/sql"

	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
	"money-planner/backend/internal/categorization"
	"money-planner/backend/internal/statement"
)

// SetupRoutes configures all API routes
func SetupRoutes(
	router chi.Router,
	service *statement.StatementService,
	categService *categorization.CategorizationService,
	logger *logrus.Logger,
	dbConn *sql.DB,
) {
	// Transaction categorization endpoint
	categHandler := NewCategorizationHandler(categService)
	router.Post("/transactions/categorize", categHandler.HandleCategorize)

	// Category analytics endpoints
	categoriesHandler := NewCategoriesHandler(categService, dbConn)
	router.Get("/categories", categoriesHandler.HandleGetCategories)
	router.Get("/categories/{id}/transactions", categoriesHandler.HandleGetCategoryTransactions)

	// Recategorization endpoint
	recategorizeHandler := NewRecategorizeHandler(categService, dbConn)
	router.Post("/transactions/{id}/recategorize", recategorizeHandler.HandleRecategorize)

	router.Route("/statements", func(sr chi.Router) {
		// List statements
		listHandler := NewListHandler(service)
		sr.Get("/", listHandler.List)

		// Upload endpoint
		uploadHandler := NewUploadHandler(service, logger)
		sr.Post("/upload", uploadHandler.Upload)

		// Preview and confirm endpoints
		sr.Route("/{id}", func(idr chi.Router) {
			previewHandler := NewPreviewHandler(service, categService)
			confirmHandler := NewConfirmHandler(service).WithCategorization(categService, dbConn)
			deleteHandler := NewDeleteHandler(service)

			idr.Get("/preview", previewHandler.Preview)
			idr.Post("/confirm", confirmHandler.Confirm)
			idr.Delete("/", deleteHandler.Delete)
		})
	})
}
