package categorization

import "time"

// Category represents a predefined transaction category
type Category struct {
	ID          string    `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	Description string    `db:"description" json:"description"`
	Color       string    `db:"color" json:"color"`
	Icon        string    `db:"icon" json:"icon"`
	IsPredefined bool     `db:"is_predefined" json:"is_predefined"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

// MerchantDictionaryEntry maps merchant names to categories (database entity)
type MerchantDictionaryEntry struct {
	ID             string    `db:"id" json:"id"`
	MerchantName   string    `db:"merchant_name" json:"merchant_name"`
	MerchantPattern string   `db:"merchant_pattern" json:"merchant_pattern"`
	CategoryID     string    `db:"category_id" json:"category_id"`
	Source         string    `db:"source" json:"source"`
	Confidence     int       `db:"confidence" json:"confidence"`
	MatchType      string    `db:"match_type" json:"match_type"`
	Frequency      int       `db:"frequency" json:"frequency"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time `db:"updated_at" json:"updated_at"`
}

// TransactionCategory links a transaction to its category
type TransactionCategory struct {
	ID               string    `db:"id" json:"id"`
	UserID           string    `db:"user_id" json:"user_id"`
	TransactionID    string    `db:"transaction_id" json:"transaction_id"`
	CategoryID       string    `db:"category_id" json:"category_id"`
	Method           string    `db:"method" json:"method"`
	LLMProvider      *string   `db:"llm_provider" json:"llm_provider"`
	Confidence       float64   `db:"confidence" json:"confidence"`
	LLMExplanation   *string   `db:"llm_explanation" json:"llm_explanation"`
	AssignedByUserID *string   `db:"assigned_by_user_id" json:"assigned_by_user_id"`
	AssignedAt       time.Time `db:"assigned_at" json:"assigned_at"`
	UpdatedAt        time.Time `db:"updated_at" json:"updated_at"`
}

// CategoryStats holds pre-computed analytics
type CategoryStats struct {
	ID                 string    `db:"id" json:"id"`
	UserID             string    `db:"user_id" json:"user_id"`
	CategoryID         string    `db:"category_id" json:"category_id"`
	Period             string    `db:"period" json:"period"`
	TotalSpent         float64   `db:"total_spent" json:"total_spent"`
	TransactionCount   int       `db:"transaction_count" json:"transaction_count"`
	AverageTransaction float64   `db:"average_transaction" json:"average_transaction"`
	MinTransaction     *float64  `db:"min_transaction" json:"min_transaction"`
	MaxTransaction     *float64  `db:"max_transaction" json:"max_transaction"`
	LastTransactionAt  *time.Time `db:"last_transaction_at" json:"last_transaction_at"`
	CreatedAt          time.Time `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time `db:"updated_at" json:"updated_at"`
}

// CategorizeRequest is the API request payload
type CategorizeRequest struct {
	Transactions []CategorizeTransactionInput `json:"transactions"`
}

// CategorizeTransactionInput represents a transaction to categorize
type CategorizeTransactionInput struct {
	ID        string  `json:"id"`
	Merchant  string  `json:"merchant"`
	Amount    float64 `json:"amount"`
	Timestamp int64   `json:"timestamp"`
}

// CategorizeResponse is the API response payload
type CategorizeResponse struct {
	Transactions []CategorizeTransactionResult `json:"transactions"`
	Stats        CategorizationStats           `json:"stats"`
}

// CategorizeTransactionResult represents a categorized transaction
type CategorizeTransactionResult struct {
	ID          string  `json:"id"`
	Category    string  `json:"category"`
	Confidence  float64 `json:"confidence"`
	Method      string  `json:"method"`
	LLMProvider *string `json:"llm_provider,omitempty"`
	Explanation string  `json:"explanation"`
}

// CategorizationStats provides summary of categorization results
type CategorizationStats struct {
	Total           int            `json:"total"`
	Categorized     int            `json:"categorized"`
	Uncategorized   int            `json:"uncategorized"`
	ByMethod        map[string]int `json:"by_method"`
	LLMProviders    map[string]int `json:"llm_providers,omitempty"`
	AvgConfidence   float64        `json:"avg_confidence"`
}
