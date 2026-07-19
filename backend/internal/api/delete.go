package api

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"money-planner/backend/internal/api/middleware"
	"money-planner/backend/internal/statement"
)

// DeleteHandler handles statement deletion requests
type DeleteHandler struct {
	service *statement.StatementService
}

// NewDeleteHandler creates a new delete handler
func NewDeleteHandler(service *statement.StatementService) *DeleteHandler {
	return &DeleteHandler{
		service: service,
	}
}

// Delete handles DELETE /api/statements/{id}
func (h *DeleteHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		middleware.WriteJSONError(w, http.StatusMethodNotAllowed, "method not allowed", "METHOD_NOT_ALLOWED")
		return
	}

	// Get user ID from context
	userID, err := middleware.GetUserID(r)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "user not authenticated", "UNAUTHORIZED")
		return
	}

	// Get statement ID from URL
	statementIDStr := chi.URLParam(r, "id")
	if statementIDStr == "" {
		middleware.WriteJSONError(w, http.StatusBadRequest, "statement id is required", "MISSING_ID")
		return
	}

	statementID, err := uuid.Parse(statementIDStr)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid statement id format", "INVALID_ID")
		return
	}

	// Verify statement belongs to user
	stmt, err := h.service.GetStatement(statementID.String())
	if err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to fetch statement", "FETCH_ERROR")
		return
	}

	if stmt == nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "statement not found", "NOT_FOUND")
		return
	}

	if stmt.UserID.String() != userID {
		middleware.WriteJSONError(w, http.StatusForbidden, "you don't have permission to delete this statement", "FORBIDDEN")
		return
	}

	// Delete statement and its transactions
	if err := h.service.DeleteStatement(statementID); err != nil {
		log.Printf("Error deleting statement %s for user %s: %v", statementID, userID, err)
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to delete statement", "DELETE_ERROR")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
