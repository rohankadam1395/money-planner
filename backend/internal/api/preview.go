package api

import (
	"encoding/json"
	"net/http"

	"money-planner/backend/internal/api/middleware"
	"money-planner/backend/internal/categorization"
	"money-planner/backend/internal/statement"
	"github.com/go-chi/chi/v5"
)

// PreviewHandler handles statement preview requests
type PreviewHandler struct {
	service              *statement.StatementService
	categorizationService *categorization.CategorizationService
}

// NewPreviewHandler creates a new preview handler
func NewPreviewHandler(
	service *statement.StatementService,
	categorizationService *categorization.CategorizationService,
) *PreviewHandler {
	return &PreviewHandler{
		service:               service,
		categorizationService: categorizationService,
	}
}

// Preview handles GET /api/statements/{id}/preview
func (h *PreviewHandler) Preview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.WriteJSONError(w, http.StatusMethodNotAllowed, "method not allowed", "METHOD_NOT_ALLOWED")
		return
	}

	// Get user ID from context (verify authentication)
	userID, err := middleware.GetUserID(r)
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

	// Fetch statement from database
	stmt, err := h.service.GetStatement(statementID)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to fetch statement", "FETCH_ERROR")
		return
	}

	if stmt == nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "statement not found", "NOT_FOUND")
		return
	}

	// Verify ownership
	if stmt.UserID.String() != userID {
		middleware.WriteJSONError(w, http.StatusForbidden, "access denied", "FORBIDDEN")
		return
	}

	// Fetch transactions for this statement
	transactions, err := h.service.GetTransactions(statementID)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to fetch transactions", "FETCH_ERROR")
		return
	}

	// Categorize transactions if service available
	var categorizationStats *categorization.CategorizationStats
	if h.categorizationService != nil && len(transactions) > 0 {
		categInput := make([]categorization.TransactionInput, len(transactions))
		for i, t := range transactions {
			categInput[i] = categorization.TransactionInput{
				ID:       t.TransactionID,
				Merchant: t.Merchant,
				Amount:   t.Amount,
				Timestamp: 0, // Optional for rule-based categorization
			}
		}

		categResults := h.categorizationService.CategorizeTransactions(r.Context(), categInput)
		categorizationStats = &categorization.CategorizationStats{
			Total:         len(categResults),
			Categorized:   0,
			Uncategorized: 0,
			ByMethod:      make(map[string]int),
		}

		// Merge categorization results with transactions
		for i, result := range categResults {
			transactions[i].Category = result.Category
			transactions[i].CategoryConfidence = result.Confidence
			transactions[i].CategoryMethod = result.Method

			categorizationStats.ByMethod[result.Method]++
			if result.Category != "Uncategorized" {
				categorizationStats.Categorized++
			} else {
				categorizationStats.Uncategorized++
			}
		}

		if categorizationStats.Categorized > 0 {
			totalConfidence := 0.0
			for _, result := range categResults {
				if result.Category != "Uncategorized" {
					totalConfidence += result.Confidence
				}
			}
			categorizationStats.AvgConfidence = totalConfidence / float64(categorizationStats.Categorized)
		}
	}

	// Build preview response
	validCount := len(transactions)
	previewResp := &statement.PreviewResponse{
		Transactions: transactions,
		Categorization: categorizationStats,
		ValidationSummary: &statement.ValidationSummary{
			TotalTransactions:   len(transactions),
			ValidTransactions:   validCount,
			InvalidTransactions: 0,
			Errors:              []map[string]interface{}{},
		},
		Status: "SUCCESS",
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
