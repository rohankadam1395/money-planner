package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"money-planner/backend/internal/categorization"
)

// CategorizationHandler holds dependencies for categorization endpoints
type CategorizationHandler struct {
	service *categorization.CategorizationService
}

// NewCategorizationHandler creates a new categorization handler
func NewCategorizationHandler(service *categorization.CategorizationService) *CategorizationHandler {
	return &CategorizationHandler{
		service: service,
	}
}

// HandleCategorize processes a batch of transactions for categorization
// POST /api/v1/transactions/categorize
func (h *CategorizationHandler) HandleCategorize(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req categorization.CategorizeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.Transactions) == 0 {
		http.Error(w, "Transactions list is empty", http.StatusBadRequest)
		return
	}

	// Convert input to internal format
	txns := make([]categorization.TransactionInput, len(req.Transactions))
	for i, t := range req.Transactions {
		txns[i] = categorization.TransactionInput{
			ID:        t.ID,
			Merchant:  t.Merchant,
			Amount:    t.Amount,
			Timestamp: t.Timestamp,
		}
	}

	// Categorize transactions
	fmt.Fprintf(os.Stderr, "DEBUG: Starting categorization for %d transactions\n", len(txns))
	results := h.service.CategorizeTransactions(r.Context(), txns)
	fmt.Fprintf(os.Stderr, "DEBUG: Categorization completed, got %d results\n", len(results))

	// Build response with statistics
	responseResults := make([]categorization.CategorizeTransactionResult, len(results))
	stats := categorization.CategorizationStats{
		Total:        len(results),
		ByMethod:     make(map[string]int),
		LLMProviders: make(map[string]int),
		Categorized:  0,
	}

	totalConfidence := 0.0
	for i, result := range results {
		var llmProvider *string
		if result.LLMProvider != "" {
			llmProvider = &result.LLMProvider
		}

		responseResults[i] = categorization.CategorizeTransactionResult{
			ID:          txns[i].ID,
			Category:    result.Category,
			Confidence:  result.Confidence,
			Method:      result.Method,
			LLMProvider: llmProvider,
			Explanation: result.Reason,
		}

		stats.ByMethod[result.Method]++
		if result.LLMProvider != "" {
			stats.LLMProviders[result.LLMProvider]++
		}
		if result.Category != "Uncategorized" {
			stats.Categorized++
			totalConfidence += result.Confidence
		} else {
			stats.Uncategorized++
		}
	}

	if stats.Categorized > 0 {
		stats.AvgConfidence = totalConfidence / float64(stats.Categorized)
	}

	response := categorization.CategorizeResponse{
		Transactions: responseResults,
		Stats:        stats,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
