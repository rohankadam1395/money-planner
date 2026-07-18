package api

import (
	"encoding/json"
	"net/http"

	"money-planner/backend/internal/categorization"
)

// CategoriesHandler handles category-related endpoints
type CategoriesHandler struct {
	service *categorization.CategorizationService
}

// NewCategoriesHandler creates a new categories handler
func NewCategoriesHandler(service *categorization.CategorizationService) *CategoriesHandler {
	return &CategoriesHandler{
		service: service,
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

	// TODO: In production, this would:
	// 1. Get all categories from database
	// 2. Query category_stats table for the requested period
	// 3. Aggregate transaction counts and spending

	// Stub response with all 10 categories
	categories := []CategoryWithStats{
		{
			ID:              "cat_food",
			Name:            "Food & Dining",
			Description:     "Restaurants, food delivery, groceries",
			Color:           "#FF6B6B",
			Icon:            "🍔",
			IsPredefined:    true,
			TotalSpent:      4500.00,
			TransactionCount: 24,
			AverageTransaction: 187.50,
		},
		{
			ID:              "cat_shopping",
			Name:            "Shopping",
			Description:     "Retail, clothing, online marketplaces",
			Color:           "#4ECDC4",
			Icon:            "🛍️",
			IsPredefined:    true,
			TotalSpent:      8000.00,
			TransactionCount: 15,
			AverageTransaction: 533.33,
		},
		{
			ID:              "cat_transport",
			Name:            "Transport",
			Description:     "Ride-sharing, fuel, transport",
			Color:           "#45B7D1",
			Icon:            "🚗",
			IsPredefined:    true,
			TotalSpent:      2100.00,
			TransactionCount: 18,
			AverageTransaction: 116.67,
		},
		{
			ID:              "cat_housing",
			Name:            "Housing",
			Description:     "Rent, property, home maintenance",
			Color:           "#F7B731",
			Icon:            "🏠",
			IsPredefined:    true,
			TotalSpent:      15000.00,
			TransactionCount: 3,
			AverageTransaction: 5000.00,
		},
		{
			ID:              "cat_utilities",
			Name:            "Utilities",
			Description:     "Electricity, water, internet, phone",
			Color:           "#5F27CD",
			Icon:            "💡",
			IsPredefined:    true,
			TotalSpent:      1200.00,
			TransactionCount: 4,
			AverageTransaction: 300.00,
		},
		{
			ID:              "cat_entertainment",
			Name:            "Entertainment",
			Description:     "Movies, streaming, games, events",
			Color:           "#EE5A6F",
			Icon:            "🎬",
			IsPredefined:    true,
			TotalSpent:      800.00,
			TransactionCount: 6,
			AverageTransaction: 133.33,
		},
		{
			ID:              "cat_income",
			Name:            "Income",
			Description:     "Salary, freelance, refunds",
			Color:           "#2ECC71",
			Icon:            "💰",
			IsPredefined:    true,
			TotalSpent:      0.00,
			TransactionCount: 0,
			AverageTransaction: 0.00,
		},
		{
			ID:              "cat_healthcare",
			Name:            "Healthcare",
			Description:     "Medical, pharmacy, gym, insurance",
			Color:           "#FF4757",
			Icon:            "🏥",
			IsPredefined:    true,
			TotalSpent:      600.00,
			TransactionCount: 2,
			AverageTransaction: 300.00,
		},
		{
			ID:              "cat_education",
			Name:            "Education",
			Description:     "Tuition, courses, books",
			Color:           "#1E90FF",
			Icon:            "📚",
			IsPredefined:    true,
			TotalSpent:      2000.00,
			TransactionCount: 1,
			AverageTransaction: 2000.00,
		},
		{
			ID:              "cat_misc",
			Name:            "Miscellaneous",
			Description:     "Gifts, charity, other",
			Color:           "#95A5A6",
			Icon:            "📌",
			IsPredefined:    true,
			TotalSpent:      300.00,
			TransactionCount: 3,
			AverageTransaction: 100.00,
		},
	}

	totalSpent := 0.0
	for _, cat := range categories {
		totalSpent += cat.TotalSpent
	}

	response := CategoriesResponse{
		Categories: categories,
		Period:     "2024-07",
		TotalSpent: totalSpent,
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

	// TODO: In production, this would:
	// 1. Extract category_id from URL path
	// 2. Query transaction_categories with filtering/sorting
	// 3. Join with transactions and categories tables

	response := CategoryTransactionsResponse{
		CategoryID:   "cat_food",
		CategoryName: "Food & Dining",
		Transactions: []CategoryTransaction{},
		Total:        0,
		TotalSpent:   0.0,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
