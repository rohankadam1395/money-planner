package api

import (
	"encoding/json"
	"net/http"

	"money-planner/backend/internal/api/middleware"
	"money-planner/backend/internal/statement"
	"github.com/go-chi/chi/v5"
)

// PreviewHandler handles statement preview requests
type PreviewHandler struct {
	service *statement.StatementService
}

// NewPreviewHandler creates a new preview handler
func NewPreviewHandler(service *statement.StatementService) *PreviewHandler {
	return &PreviewHandler{
		service: service,
	}
}

// Preview handles GET /api/statements/{id}/preview
func (h *PreviewHandler) Preview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.WriteJSONError(w, http.StatusMethodNotAllowed, "method not allowed", "METHOD_NOT_ALLOWED")
		return
	}

	// Get user ID from context (verify authentication)
	_, err := middleware.GetUserID(r)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "user not authenticated", "UNAUTHORIZED")
		return
	}

	// Get statement ID from URL
	statementID := chi.URLParam(r, "id")
	if statementID == "" {
		middleware.WriteJSONError(w, http.StatusBadRequest, "statement ID is required", "MISSING_STATEMENT_ID")
		return
	}

	// TODO: Fetch statement from database and verify ownership
	// For now, return placeholder preview response

	// In a production system, this would:
	// 1. Fetch the statement record from DB
	// 2. Check if it belongs to the authenticated user
	// 3. If status is PENDING, fetch cached extracted transactions
	// 4. If status is SUCCESS, fetch persisted transactions
	// 5. Return the preview response

	previewResp := &statement.PreviewResponse{
		Transactions: []*statement.Transaction{},
		ValidationSummary: &statement.ValidationSummary{
			TotalTransactions:   0,
			ValidTransactions:   0,
			InvalidTransactions: 0,
			Errors:              []map[string]interface{}{},
		},
		Status:  "PROCESSING",
		Message: "Statement is being processed. Check back in a few seconds.",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(previewResp)
}

// PreviewTransactionResponse represents a single transaction in preview
type PreviewTransactionResponse struct {
	TransactionID string  `json:"transaction_id"`
	Date          string  `json:"transaction_date"`
	Merchant      string  `json:"merchant"`
	Amount        float64 `json:"amount"`
	Type          string  `json:"type"`
	Balance       *float64 `json:"balance,omitempty"`
	Description   string  `json:"description,omitempty"`
	Currency      string  `json:"currency"`
}
