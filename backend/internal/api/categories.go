package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"money-planner/backend/internal/api/middleware"
	"money-planner/backend/internal/categorization"
)

// CategoriesHandler handles category-related endpoints
type CategoriesHandler struct {
	service *categorization.CategorizationService
	dbConn  *sql.DB
}

// NewCategoriesHandler creates a new categories handler
func NewCategoriesHandler(service *categorization.CategorizationService, dbConn *sql.DB) *CategoriesHandler {
	return &CategoriesHandler{
		service: service,
		dbConn:  dbConn,
	}
}

// CategoryWithStats represents a category with aggregated stats
type CategoryWithStats struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	Description     string  `json:"description"`
	Color           string  `json:"color"`
	Icon            string  `json:"icon"`
	IsPredefined    bool    `json:"is_predefined"`
	TotalSpent      float64 `json:"total_spent"`
	TransactionCount int    `json:"transaction_count"`
	AverageTransaction float64 `json:"average_transaction"`
}

// CategoriesResponse is the response for listing categories
type CategoriesResponse struct {
	Categories []CategoryWithStats `json:"categories"`
	Period     string              `json:"period,omitempty"`
	TotalSpent float64             `json:"total_spent"`
}

// HandleGetCategories returns all categories with stats
// GET /api/v1/categories
func (h *CategoriesHandler) HandleGetCategories(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := middleware.GetUserID(r)
	if err != nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	// Get period from query params (defaults to current month YYYY-MM)
	period := r.URL.Query().Get("period")
	if period == "" {
		// Default to current month
		period = time.Now().Format("2006-01")
	}

	ctx := r.Context()

	// Get all categories from database (T097)
	rows, err := h.dbConn.QueryContext(ctx,
		`SELECT id, name, description, color, icon FROM categories ORDER BY name ASC`)
	if err != nil {
		http.Error(w, "Failed to retrieve categories", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	categories := []CategoryWithStats{}
	totalSpentAll := 0.0

	for rows.Next() {
		var id, name, desc, color, icon string
		if err := rows.Scan(&id, &name, &desc, &color, &icon); err != nil {
			continue
		}

		// Query category_stats for this category in the requested period
		var totalSpent sql.NullFloat64
		var transactionCount sql.NullInt64
		var avgTransaction sql.NullFloat64

		err = h.dbConn.QueryRowContext(ctx,
			`SELECT COALESCE(total_spent, 0),
                    COALESCE(transaction_count, 0),
                    COALESCE(average_transaction, 0)
             FROM category_stats
             WHERE user_id = $1 AND category_id = $2 AND period = $3`,
			userID, id, period,
		).Scan(&totalSpent, &transactionCount, &avgTransaction)

		// If no stats found, that's okay - use zeros
		spent := 0.0
		count := 0
		avg := 0.0

		if err == nil {
			if totalSpent.Valid {
				spent = totalSpent.Float64
			}
			if transactionCount.Valid {
				count = int(transactionCount.Int64)
			}
			if avgTransaction.Valid {
				avg = avgTransaction.Float64
			}
		} else if err != sql.ErrNoRows {
			// Log unexpected errors but continue
			continue
		}

		categories = append(categories, CategoryWithStats{
			ID:                  id,
			Name:                name,
			Description:         desc,
			Color:               color,
			Icon:                icon,
			IsPredefined:        true,
			TotalSpent:          spent,
			TransactionCount:    count,
			AverageTransaction:  avg,
		})

		totalSpentAll += spent
	}

	// Add uncategorized transactions card
	var uncatCount int
	var uncatSpent float64
	err = h.dbConn.QueryRowContext(ctx,
		`SELECT COUNT(*), COALESCE(SUM(amount), 0)
		 FROM transactions t
		 WHERE t.user_id = $1 AND NOT EXISTS (
		   SELECT 1 FROM transaction_categories tc WHERE tc.transaction_id = t.transaction_id AND tc.user_id = $1
		 )`,
		userID,
	).Scan(&uncatCount, &uncatSpent)
	if err == nil && uncatCount > 0 {
		categories = append(categories, CategoryWithStats{
			ID:               "uncategorized",
			Name:             "Uncategorized",
			Description:      "Transactions without a category",
			Color:            "#9CA3AF",
			Icon:             "question",
			IsPredefined:     false,
			TotalSpent:       uncatSpent,
			TransactionCount: uncatCount,
			AverageTransaction: uncatSpent / float64(uncatCount),
		})
		totalSpentAll += uncatSpent
	}

	response := CategoriesResponse{
		Categories: categories,
		Period:     period,
		TotalSpent: totalSpentAll,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CategoryTransactionsResponse represents transactions in a category
type CategoryTransactionsResponse struct {
	CategoryID   string                `json:"category_id"`
	CategoryName string                `json:"category_name"`
	Transactions []CategoryTransaction `json:"transactions"`
	Total        int                   `json:"total"`
	TotalSpent   float64               `json:"total_spent"`
}

// CategoryTransaction represents a single transaction in a category
type CategoryTransaction struct {
	TransactionID   string  `json:"transaction_id"`
	Date            string  `json:"date"`
	Merchant        string  `json:"merchant"`
	Amount          float64 `json:"amount"`
	Method          string  `json:"method"`
	LLMProvider     *string `json:"llm_provider,omitempty"`
	Confidence      float64 `json:"confidence"`
}

// HandleGetCategoryTransactions returns transactions in a specific category
// GET /api/v1/categories/{id}/transactions
func (h *CategoriesHandler) HandleGetCategoryTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := middleware.GetUserID(r)
	if err != nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	// Extract category_id from URL
	categoryID := chi.URLParam(r, "id")
	if categoryID == "" {
		http.Error(w, "Category ID is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	transactions := []CategoryTransaction{}
	totalSpent := 0.0
	var categoryName string

	if categoryID == "uncategorized" {
		categoryName = "Uncategorized"
		// Query transactions with no category assignment
		rows, err := h.dbConn.QueryContext(ctx,
			`SELECT t.transaction_id, t.transaction_date, t.merchant, t.amount
			 FROM transactions t
			 WHERE t.user_id = $1 AND NOT EXISTS (
			   SELECT 1 FROM transaction_categories tc WHERE tc.transaction_id = t.transaction_id AND tc.user_id = $1
			 )
			 ORDER BY t.transaction_date DESC`,
			userID,
		)
		if err != nil {
			log.Printf("Error querying uncategorized transactions for user %s: %v", userID, err)
			http.Error(w, "Failed to fetch transactions", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var txnID, date, merchant string
			var amount float64

			if err := rows.Scan(&txnID, &date, &merchant, &amount); err != nil {
				continue
			}

			transactions = append(transactions, CategoryTransaction{
				TransactionID: txnID,
				Date:          date,
				Merchant:      merchant,
				Amount:        amount,
				Method:        "none",
				Confidence:    0,
			})

			totalSpent += amount
		}
	} else {
		// Get category name for regular categories
		err = h.dbConn.QueryRowContext(ctx, `SELECT name FROM categories WHERE id = $1`, categoryID).Scan(&categoryName)
		if err != nil {
			log.Printf("Error fetching category %s: %v", categoryID, err)
			http.Error(w, "Category not found", http.StatusNotFound)
			return
		}

		// Query transactions in this category
		rows, err := h.dbConn.QueryContext(ctx,
			`SELECT tc.transaction_id, t.transaction_date, t.merchant, t.amount, tc.method, tc.llm_provider, tc.confidence
			 FROM transaction_categories tc
			 JOIN transactions t ON tc.transaction_id = t.transaction_id
			 WHERE tc.user_id = $1 AND tc.category_id = $2
			 ORDER BY t.transaction_date DESC`,
			userID, categoryID,
		)
		if err != nil {
			log.Printf("Error querying transactions for category %s, user %s: %v", categoryID, userID, err)
			http.Error(w, "Failed to fetch transactions", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var txnID, date, merchant, method string
			var amount, confidence float64
			var llmProvider sql.NullString

			if err := rows.Scan(&txnID, &date, &merchant, &amount, &method, &llmProvider, &confidence); err != nil {
				continue
			}

			var llmProviderPtr *string
			if llmProvider.Valid {
				llmProviderPtr = &llmProvider.String
			}

			transactions = append(transactions, CategoryTransaction{
				TransactionID: txnID,
				Date:          date,
				Merchant:      merchant,
				Amount:        amount,
				Method:        method,
				LLMProvider:   llmProviderPtr,
				Confidence:    confidence,
			})

			totalSpent += amount
		}
	}

	response := CategoryTransactionsResponse{
		CategoryID:   categoryID,
		CategoryName: categoryName,
		Transactions: transactions,
		Total:        len(transactions),
		TotalSpent:   totalSpent,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
