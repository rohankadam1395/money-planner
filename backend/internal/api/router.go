package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
	"money-planner/backend/internal/statement"
)

// SetupRoutes configures all API routes
func SetupRoutes(router chi.Router, service *statement.StatementService, logger *logrus.Logger) {
	router.Route("/statements", func(sr chi.Router) {
		// List statements
		listHandler := NewListHandler(service)
		sr.Get("/", listHandler.List)

		// Upload endpoint
		uploadHandler := NewUploadHandler(service, logger)
		sr.Post("/upload", uploadHandler.Upload)

		// Preview and confirm endpoints
		sr.Route("/{id}", func(idr chi.Router) {
			previewHandler := NewPreviewHandler(service)
			confirmHandler := NewConfirmHandler(service)

			idr.Get("/preview", previewHandler.Preview)
			idr.Post("/confirm", confirmHandler.Confirm)
		})
	})
}
