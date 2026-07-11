package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"money-planner/backend/internal/api/middleware"
	"money-planner/backend/internal/statement"
)

// ListHandler handles statement list requests
type ListHandler struct {
	service *statement.StatementService
}

// NewListHandler creates a new list handler
func NewListHandler(service *statement.StatementService) *ListHandler {
	return &ListHandler{
		service: service,
	}
}

// StatementListResponse represents a statement in the list
type StatementListResponse struct {
	StatementID      string `json:"statement_id"`
	FileName         string `json:"file_name"`
	FileFormat       string `json:"file_format"`
	BankCode         string `json:"bank_code"`
	TransactionCount int    `json:"transaction_count"`
	Status           string `json:"status"`
	UploadedAt       string `json:"uploaded_at"`
}

// ListStatementsResponse is the API response for listing statements
type ListStatementsResponse struct {
	Data       []StatementListResponse `json:"data"`
	Pagination PaginationMeta          `json:"pagination"`
	Message    string                  `json:"message,omitempty"`
}

// List handles GET /api/statements
func (h *ListHandler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.WriteJSONError(w, http.StatusMethodNotAllowed, "method not allowed", "METHOD_NOT_ALLOWED")
		return
	}

	// Get user ID from context
	userID, err := middleware.GetUserID(r)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "user not authenticated", "UNAUTHORIZED")
		return
	}

	// Parse pagination parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	offset := 0
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Fetch statements from service
	statements, err := h.service.ListStatements(userID, limit, offset)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to fetch statements", "FETCH_ERROR")
		return
	}

	// Convert to response format
	var data []StatementListResponse
	if statements != nil {
		for _, stmt := range statements {
			data = append(data, StatementListResponse{
				StatementID:      stmt.StatementID.String(),
				FileName:         stmt.FileName,
				FileFormat:       stmt.FileFormat,
				BankCode:         stmt.BankCode,
				TransactionCount: stmt.TransactionCount,
				Status:           stmt.Status,
				UploadedAt:       stmt.UploadedAt.Format("2006-01-02 15:04:05"),
			})
		}
	}

	resp := &ListStatementsResponse{
		Data: data,
		Pagination: PaginationMeta{
			Total:  len(data),
			Limit:  limit,
			Offset: offset,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
