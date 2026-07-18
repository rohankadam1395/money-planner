package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/go-chi/chi/v5"
	"money-planner/backend/internal/api/middleware"
	"money-planner/backend/internal/categorization"
	"money-planner/backend/internal/statement"
)

// ConfirmHandler handles statement import confirmation
type ConfirmHandler struct {
	service         *statement.StatementService
	categService    *categorization.CategorizationService
	dbConn          *sql.DB
}

// NewConfirmHandler creates a new confirm handler
func NewConfirmHandler(service *statement.StatementService) *ConfirmHandler {
	return &ConfirmHandler{
		service: service,
	}
}

// WithCategorization adds categorization service
func (h *ConfirmHandler) WithCategorization(categService *categorization.CategorizationService, dbConn *sql.DB) *ConfirmHandler {
	h.categService = categService
	h.dbConn = dbConn
	return h
}

// Confirm handles POST /api/statements/{id}/confirm
func (h *ConfirmHandler) Confirm(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.WriteJSONError(w, http.StatusMethodNotAllowed, "method not allowed", "METHOD_NOT_ALLOWED")
		return
	}

	userID, err := middleware.GetUserID(r)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "user not authenticated", "UNAUTHORIZED")
		return
	}

	statementID := chi.URLParam(r, "id")
	if statementID == "" {
		middleware.WriteJSONError(w, http.StatusBadRequest, "statement ID is required", "MISSING_STATEMENT_ID")
		return
	}

	ctx := r.Context()

	// Fetch statement from database
	stmt, err := h.service.GetStatement(statementID)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to fetch statement", "DB_ERROR")
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
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to fetch transactions", "DB_ERROR")
		return
	}

	userIDUUID, _ := uuid.Parse(userID)
	period := time.Now().Format("2006-01")
	txnCount := 0

	// Categorize and save each transaction
	if h.categService != nil && h.dbConn != nil {
		for _, txn := range transactions {
			// Categorize the transaction
			result := h.categService.CategorizeTransaction(ctx, txn.Merchant, txn.Amount)

			// Get category ID from name
			var categoryID string
			err := h.dbConn.QueryRowContext(ctx,
				`SELECT id FROM categories WHERE name = $1`,
				result.Category,
			).Scan(&categoryID)

			if err != nil && err != sql.ErrNoRows {
				continue
			}

			if categoryID == "" {
				categoryID = "uncategorized" // fallback
			}

			// Insert into transaction_categories
			_, err = h.dbConn.ExecContext(ctx,
				`INSERT INTO transaction_categories
				 (id, user_id, transaction_id, category_id, method, confidence, assigned_at, updated_at)
				 VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
				 ON CONFLICT (transaction_id) DO UPDATE SET
				 category_id = $4, method = $5, confidence = $6, updated_at = NOW()`,
				uuid.New(), userIDUUID, txn.TransactionID, categoryID, result.Method, result.Confidence,
			)

			if err == nil {
				txnCount++

				// Update category_stats
				h.dbConn.ExecContext(ctx,
					`INSERT INTO category_stats
					 (id, user_id, category_id, period, total_spent, transaction_count, average_transaction, created_at, updated_at)
					 VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
					 ON CONFLICT (user_id, category_id, period) DO UPDATE SET
					 total_spent = category_stats.total_spent + $5,
					 transaction_count = category_stats.transaction_count + 1,
					 average_transaction = (category_stats.total_spent + $5) / (category_stats.transaction_count + 1),
					 updated_at = NOW()`,
					uuid.New(), userIDUUID, categoryID, period, txn.Amount, 1, txn.Amount,
				)
			}
		}
	}

	confirmResp := &statement.ConfirmImportResponse{
		StatementID:      statementID,
		Status:           "SUCCESS",
		TransactionCount: txnCount,
		Message:          fmt.Sprintf("%d transactions confirmed and categorized", txnCount),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(confirmResp); err != nil {
		fmt.Printf("error encoding response: %v\n", err)
	}
}
