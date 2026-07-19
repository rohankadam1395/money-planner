package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/go-chi/chi/v5"
	"money-planner/backend/internal/api/middleware"
	"money-planner/backend/internal/categorization"
)

// RecategorizeHandler holds dependencies for recategorization
type RecategorizeHandler struct {
	service *categorization.CategorizationService
	dbConn  *sql.DB
}

// NewRecategorizeHandler creates a new recategorize handler
func NewRecategorizeHandler(service *categorization.CategorizationService, dbConn *sql.DB) *RecategorizeHandler {
	return &RecategorizeHandler{
		service: service,
		dbConn:  dbConn,
	}
}

// RecategorizeRequest is the API request for recategorization
type RecategorizeRequest struct {
	TransactionID   string `json:"transaction_id"`
	NewCategoryID   string `json:"new_category_id"`
	LearnCorrection bool   `json:"learn_correction,omitempty"`
}

// RecategorizeResponse is the API response after recategorization
type RecategorizeResponse struct {
	TransactionID    string `json:"transaction_id"`
	OldCategoryID    string `json:"old_category_id"`
	OldCategoryName  string `json:"old_category_name"`
	NewCategoryID    string `json:"new_category_id"`
	NewCategoryName  string `json:"new_category_name"`
	LearnedAsCorrect bool   `json:"learned_as_correct"`
	UpdatedAt        string `json:"updated_at"`
}

// HandleRecategorize updates a transaction's category
// POST /api/v1/transactions/{id}/recategorize
func (h *RecategorizeHandler) HandleRecategorize(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.WriteJSONError(w, http.StatusMethodNotAllowed, "method not allowed", "METHOD_NOT_ALLOWED")
		return
	}

	userID, err := middleware.GetUserID(r)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "user not authenticated", "UNAUTHORIZED")
		return
	}

	txnID := chi.URLParam(r, "id")
	if txnID == "" {
		middleware.WriteJSONError(w, http.StatusBadRequest, "transaction ID is required", "MISSING_TXN_ID")
		return
	}

	var req RecategorizeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid request body", "INVALID_REQUEST")
		return
	}

	if req.NewCategoryID == "" {
		middleware.WriteJSONError(w, http.StatusBadRequest, "new_category_id is required", "MISSING_CATEGORY_ID")
		return
	}

	// "uncategorized" is a special case - mark transaction as having no category
	if req.NewCategoryID != "uncategorized" {
		// Validate that the new category exists in the database
		var testCat string
		err = h.dbConn.QueryRowContext(r.Context(), `SELECT id FROM categories WHERE id = $1`, req.NewCategoryID).Scan(&testCat)
		if err != nil {
			if err == sql.ErrNoRows {
				middleware.WriteJSONError(w, http.StatusBadRequest, "new category not found", "INVALID_CATEGORY")
			} else {
				middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to validate new category", "DB_ERROR")
			}
			return
		}
	}

	// Parse IDs
	txnIDUUID, err := uuid.Parse(txnID)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid transaction ID format", "INVALID_UUID")
		return
	}

	ctx := r.Context()

	// Get old category ID (if exists - uncategorized transactions won't have one)
	var oldCatID sql.NullString
	err = h.dbConn.QueryRowContext(ctx,
		`SELECT category_id FROM transaction_categories WHERE transaction_id = $1`,
		txnIDUUID,
	).Scan(&oldCatID)

	var oldCatName string
	if err != nil && err != sql.ErrNoRows {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to retrieve transaction category", "DB_ERROR")
		return
	}

	// Get old category name if it exists
	if oldCatID.Valid {
		err = h.dbConn.QueryRowContext(ctx, `SELECT name FROM categories WHERE id = $1`, oldCatID.String).Scan(&oldCatName)
		if err != nil {
			middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to retrieve old category", "DB_ERROR")
			return
		}
	} else {
		oldCatName = "Uncategorized"
	}

	var newCatName string
	if req.NewCategoryID == "uncategorized" {
		newCatName = "Uncategorized"
	} else {
		// Get category name for real categories
		err = h.dbConn.QueryRowContext(ctx, `SELECT name FROM categories WHERE id = $1`, req.NewCategoryID).Scan(&newCatName)
		if err != nil {
			middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to retrieve new category", "DB_ERROR")
			return
		}
	}

	userIDUUID, _ := uuid.Parse(userID)
	period := time.Now().Format("2006-01")

	if req.NewCategoryID == "uncategorized" {
		// Remove category assignment (set as uncategorized)
		_, err = h.dbConn.ExecContext(ctx,
			`DELETE FROM transaction_categories WHERE transaction_id = $1`,
			txnIDUUID,
		)
		if err != nil {
			middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to remove transaction category", "DB_ERROR")
			return
		}

		// Recalculate stats for old category
		if oldCatID.Valid {
			h.updateCategoryStats(ctx, userIDUUID, oldCatID.String, period)
		}
	} else {
		// Assign to a real category
		if !oldCatID.Valid {
			// Uncategorized transaction - INSERT new category assignment
			_, err = h.dbConn.ExecContext(ctx,
				`INSERT INTO transaction_categories (user_id, transaction_id, category_id, method, confidence, assigned_by_user_id, updated_at)
				 VALUES ($1, $2, $3, $4, $5, $6, NOW())`,
				userIDUUID, txnIDUUID, req.NewCategoryID, "manual", 1.0, userIDUUID,
			)
		} else {
			// Already categorized - UPDATE existing assignment
			_, err = h.dbConn.ExecContext(ctx,
				`UPDATE transaction_categories SET category_id = $1, method = $2, confidence = $3, assigned_by_user_id = $4, updated_at = NOW()
				 WHERE transaction_id = $5`,
				req.NewCategoryID, "manual", 1.0, userIDUUID, txnIDUUID,
			)
		}

		if err != nil {
			middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to update transaction category", "DB_ERROR")
			return
		}

		// Recalculate category stats for both old and new categories via service layer
		if oldCatID.Valid {
			h.service.UpdateCategoryStats(ctx, userID, oldCatID.String, period)
		}
		h.service.UpdateCategoryStats(ctx, userID, req.NewCategoryID, period)
	}

	// Handle merchant dictionary learning if requested (T095)
	if req.LearnCorrection {
		merchantName := oldCatName
		if len(merchantName) > 255 {
			merchantName = merchantName[:255]
		}

		_, err := h.dbConn.ExecContext(ctx,
			`INSERT INTO merchant_dictionary (id, merchant_name, category_id, source, confidence, match_type, frequency, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())`,
			uuid.New(), merchantName, req.NewCategoryID, "user_correction", 100, "manual", 0,
		)
		if err != nil {
			// Log error but don't fail the request - learning is best-effort
			log.Printf("Warning: Failed to learn correction for merchant %s: %v", merchantName, err)
		}
	}

	oldCatIDStr := ""
	if oldCatID.Valid {
		oldCatIDStr = oldCatID.String
	}

	response := RecategorizeResponse{
		TransactionID:    txnID,
		OldCategoryID:    oldCatIDStr,
		OldCategoryName:  oldCatName,
		NewCategoryID:    req.NewCategoryID,
		NewCategoryName:  newCatName,
		LearnedAsCorrect: req.LearnCorrection,
		UpdatedAt:        time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// updateCategoryStats recalculates and updates category stats for a period
func (h *RecategorizeHandler) updateCategoryStats(ctx context.Context, userID uuid.UUID, categoryID string, period string) {
	var totalSpent sql.NullFloat64
	var count sql.NullInt64
	var avgTransaction sql.NullFloat64

	// Calculate current stats from transactions
	err := h.dbConn.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(t.amount), 0), COUNT(*), COALESCE(AVG(t.amount), 0)
		 FROM transactions t
		 JOIN transaction_categories tc ON t.transaction_id = tc.transaction_id
		 WHERE t.user_id = $1 AND tc.category_id = $2 AND DATE_TRUNC('month', t.transaction_date::timestamp)::text LIKE $3 || '%'`,
		userID, categoryID, period,
	).Scan(&totalSpent, &count, &avgTransaction)

	if err != nil {
		return
	}

	// Upsert into category_stats
	h.dbConn.ExecContext(ctx,
		`INSERT INTO category_stats (id, user_id, category_id, period, total_spent, transaction_count, average_transaction, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		 ON CONFLICT(user_id, category_id, period) DO UPDATE SET
		   total_spent = EXCLUDED.total_spent,
		   transaction_count = EXCLUDED.transaction_count,
		   average_transaction = EXCLUDED.average_transaction,
		   updated_at = NOW()`,
		uuid.New(), userID, categoryID, period,
		totalSpent.Float64, count.Int64, avgTransaction.Float64,
	)
}
