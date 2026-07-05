package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"money-planner/backend/internal/api/middleware"
	"money-planner/backend/internal/statement"
	"github.com/go-chi/chi/v5"
)

// ConfirmHandler handles statement import confirmation
type ConfirmHandler struct {
	service *statement.StatementService
}

// NewConfirmHandler creates a new confirm handler
func NewConfirmHandler(service *statement.StatementService) *ConfirmHandler {
	return &ConfirmHandler{
		service: service,
	}
}

// Confirm handles POST /api/statements/{id}/confirm
func (h *ConfirmHandler) Confirm(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
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

	// TODO: In production:
	// 1. Fetch statement from database
	// 2. Verify ownership (user_id matches authenticated user)
	// 3. Check status (should be PENDING or READY)
	// 4. Fetch previously extracted/previewed transactions
	// 5. Call service.ConfirmImport() to persist
	// 6. Update statement status to SUCCESS
	// 7. Return confirmation response

	// For now, return placeholder response
	confirmResp := &statement.ConfirmImportResponse{
		StatementID:      statementID,
		Status:           "SUCCESS",
		TransactionCount: 0,
		Message:          "Statement confirmed (placeholder response)",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(confirmResp); err != nil {
		fmt.Printf("error encoding response: %v\n", err)
	}
}
