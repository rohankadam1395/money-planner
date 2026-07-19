package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"money-planner/backend/internal/api/middleware"
	"money-planner/backend/internal/categorization"
	"money-planner/backend/internal/statement"
	"github.com/go-chi/chi/v5"
)

// PreviewHandler handles statement preview requests
type PreviewHandler struct {
	service               *statement.StatementService
	categorizationService *categorization.CategorizationService
	dbConn                *sql.DB
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

// WithDatabase adds database connection
func (h *PreviewHandler) WithDatabase(dbConn *sql.DB) *PreviewHandler {
	h.dbConn = dbConn
	return h
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

	// Fetch saved categories from database first
	savedCategories := make(map[string]struct {
		Name       string
		Confidence float64
		Method     string
	})

	if h.dbConn != nil && len(transactions) > 0 {
		for _, txn := range transactions {
			var categoryName string
			var method string
			var confidence float64

			// Query saved category if it exists
			err := h.dbConn.QueryRowContext(r.Context(),
				`SELECT c.name, tc.method, tc.confidence
				 FROM transaction_categories tc
				 JOIN categories c ON tc.category_id = c.id
				 WHERE tc.transaction_id = $1
				 LIMIT 1`,
				txn.TransactionID,
			).Scan(&categoryName, &method, &confidence)

			if err == nil && categoryName != "" {
				// Saved category found
				savedCategories[txn.TransactionID] = struct {
					Name       string
					Confidence float64
					Method     string
				}{Name: categoryName, Confidence: confidence, Method: method}
			}
		}
	}

	// Categorize transactions with rule-based matching only (fast path)
	var categorizationStats *categorization.CategorizationStats
	var categResults []categorization.CategorizationResult
	if h.categorizationService != nil && len(transactions) > 0 {
		categInput := make([]categorization.TransactionInput, len(transactions))
		for i, t := range transactions {
			categInput[i] = categorization.TransactionInput{
				ID:       t.TransactionID,
				Merchant: t.Merchant,
				Amount:   t.Amount,
				Timestamp: 0,
			}
		}

		// Rule-based categorization only (fast): exact + fuzzy matching
		categResults = h.categorizationService.CategorizeTransactionsRuleBasedOnly(r.Context(), categInput)
		categorizationStats = &categorization.CategorizationStats{
			Total:         len(categResults),
			Categorized:   0,
			Uncategorized: 0,
			ByMethod:      make(map[string]int),
		}

		// Merge categorization results with transactions
		for _, result := range categResults {
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

	// Build response with proper transaction structure including categories
	responseTransactions := make([]map[string]interface{}, len(transactions))
	for i, t := range transactions {
		txnMap := map[string]interface{}{
			"transaction_id":   t.TransactionID,
			"statement_id":     t.StatementID,
			"user_id":          t.UserID,
			"transaction_date": t.TransactionDate.Format("2006-01-02T15:04:05Z"),
			"merchant":         t.Merchant,
			"amount":           t.Amount,
			"type":             t.Type,
			"currency":         t.Currency,
			"imported_at":      t.ImportedAt.Format("2006-01-02T15:04:05Z"),
			"created_at":       t.CreatedAt.Format("2006-01-02T15:04:05Z"),
			"updated_at":       t.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}

		if t.Description != "" {
			txnMap["description"] = t.Description
		}
		if t.Balance != nil {
			txnMap["balance"] = *t.Balance
		}

		// Add category - prefer saved category over rule-based
		var categoryName string
		var confidence float64
		var method string

		if saved, ok := savedCategories[t.TransactionID]; ok {
			// Use saved category
			categoryName = saved.Name
			confidence = saved.Confidence
			method = saved.Method
		} else if len(categResults) > i {
			// Fall back to rule-based categorization
			result := categResults[i]
			categoryName = result.Category
			confidence = result.Confidence
			method = result.Method
		}

		if categoryName != "" {
			categoryMap := map[string]interface{}{
				"id":         "cat_" + strings.ToLower(strings.ReplaceAll(categoryName, " & ", "_")),
				"name":       categoryName,
				"confidence": confidence,
				"method":     method,
				"color":      getCategoryColor(categoryName),
				"icon":       getCategoryIcon(categoryName),
			}
			txnMap["category"] = categoryMap
		}

		responseTransactions[i] = txnMap
	}

	// Build preview response
	previewResp := map[string]interface{}{
		"transactions": responseTransactions,
		"categorization": categorizationStats,
		"validation_summary": map[string]interface{}{
			"total_rows":              len(transactions),
			"valid_transactions":      len(transactions),
			"invalid_transactions":    0,
			"errors":                  []interface{}{},
		},
		"status": "SUCCESS",
	}

	// Prevent caching - always fetch fresh from server
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(previewResp)
}

// Helper to get category color
func getCategoryColor(categoryName string) string {
	colors := map[string]string{
		"Food & Dining":      "#FF6B6B",
		"Shopping":           "#4ECDC4",
		"Transport":          "#45B7D1",
		"Housing":            "#F7B731",
		"Utilities":          "#5F27CD",
		"Entertainment":      "#EE5A6F",
		"Income":             "#2ECC71",
		"Healthcare":         "#FF4757",
		"Education":          "#1E90FF",
		"Miscellaneous":      "#95A5A6",
		"Uncategorized":      "#95A5A6",
	}
	if color, ok := colors[categoryName]; ok {
		return color
	}
	return "#95A5A6"
}

// Helper to get category icon
func getCategoryIcon(categoryName string) string {
	icons := map[string]string{
		"Food & Dining":      "🍔",
		"Shopping":           "🛍️",
		"Transport":          "🚗",
		"Housing":            "🏠",
		"Utilities":          "💡",
		"Entertainment":      "🎬",
		"Income":             "💰",
		"Healthcare":         "🏥",
		"Education":          "📚",
		"Miscellaneous":      "📌",
		"Uncategorized":      "❓",
	}
	if icon, ok := icons[categoryName]; ok {
		return icon
	}
	return "📌"
}
