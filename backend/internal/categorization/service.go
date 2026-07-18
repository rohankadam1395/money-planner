package categorization

import (
	"context"
	"fmt"
	"log"
)

// CategorizationService handles transaction categorization
type CategorizationService struct {
	merchantDict *MerchantDictionary
	confidencer  *ConfidenceScorer
	llmProvider  LLMProvider
}

// NewCategorizationService creates a new categorization service
func NewCategorizationService(merchantDict *MerchantDictionary, confidencer *ConfidenceScorer) *CategorizationService {
	return &CategorizationService{
		merchantDict: merchantDict,
		confidencer:  confidencer,
		llmProvider:  nil,
	}
}

// WithLLMProvider adds an LLM provider to the service
func (s *CategorizationService) WithLLMProvider(provider LLMProvider) *CategorizationService {
	s.llmProvider = provider
	return s
}

// CategorizeTransaction categorizes a single transaction with rule-based or LLM fallback
func (s *CategorizationService) CategorizeTransaction(ctx context.Context, merchant string, amount float64) *CategorizationResult {
	if merchant == "" {
		return &CategorizationResult{
			Category:   "Uncategorized",
			Method:     "none",
			Confidence: 0.0,
			Reason:     "Empty merchant name",
		}
	}

	// Try exact match first
	result := s.merchantDict.LookupExact(merchant)
	if result != nil {
		result.Confidence = s.confidencer.ScoreExactMatch()
		result.Method = "rule_based"
		result.Reason = fmt.Sprintf("Known merchant: %s", merchant)
		return result
	}

	// Try fuzzy match
	result = s.merchantDict.LookupFuzzy(merchant)
	if result != nil {
		confidence := s.confidencer.ScoreFuzzyMatch(result.matchDistance)
		result.Confidence = confidence
		result.Method = "fuzzy"
		result.Reason = fmt.Sprintf("Fuzzy match: %s (distance: %.2f)", result.Category, result.matchDistance)
		return result
	}

	// Try LLM categorization if available
	if s.llmProvider != nil {
		return s.CategorizeLLM(ctx, merchant, amount)
	}

	// No match found and no LLM available
	return &CategorizationResult{
		Category:   "Uncategorized",
		Method:     "none",
		Confidence: 0.0,
		Reason:     fmt.Sprintf("No matching merchant for: %s", merchant),
	}
}

// CategorizeLLM categorizes a transaction using the LLM provider with graceful degradation
func (s *CategorizationService) CategorizeLLM(ctx context.Context, merchant string, amount float64) *CategorizationResult {
	if s.llmProvider == nil {
		return &CategorizationResult{
			Category:   "Uncategorized",
			Method:     "none",
			Confidence: 0.0,
			Reason:     "LLM provider not available",
		}
	}

	category, confidence, explanation, err := s.llmProvider.Categorize(ctx, merchant, amount)
	if err != nil {
		log.Printf("LLM categorization failed for merchant %s: %v", merchant, err)
		return &CategorizationResult{
			Category:   "Uncategorized",
			Method:     "none",
			Confidence: 0.0,
			Reason:     fmt.Sprintf("LLM error (graceful degradation): %v", err),
		}
	}

	return &CategorizationResult{
		Category:      category,
		Method:        "llm",
		Confidence:    confidence,
		Reason:        explanation,
		LLMProvider:   s.llmProvider.Name(),
		matchDistance: 0.0,
	}
}

// CategorizeTransactions categorizes multiple transactions
func (s *CategorizationService) CategorizeTransactions(ctx context.Context, transactions []TransactionInput) []CategorizationResult {
	results := make([]CategorizationResult, len(transactions))
	for i, txn := range transactions {
		results[i] = *s.CategorizeTransaction(ctx, txn.Merchant, txn.Amount)
	}
	return results
}

// TransactionInput represents a transaction to be categorized
type TransactionInput struct {
	ID        string
	Merchant  string
	Amount    float64
	Timestamp int64
}

// CategorizationResult represents the result of categorization
type CategorizationResult struct {
	Category      string
	Method        string
	Confidence    float64
	Reason        string
	LLMProvider   string
	matchDistance float64 // internal field for fuzzy matching
}

// UpdateCategoryStats updates category_stats for a category after recategorization
// Called when a transaction is recategorized to both the old and new category
func (s *CategorizationService) UpdateCategoryStats(ctx context.Context, statsUpdate *CategoryStatsUpdate) error {
	// This method is a placeholder for database operations
	// In production, this would call database methods to upsert category_stats
	// For now, the actual stats updates happen via SQL queries in the API handlers
	return nil
}

// CategoryStatsUpdate represents statistics to be aggregated for a category
type CategoryStatsUpdate struct {
	UserID      string
	CategoryID  string
	Period      string // YYYY-MM format
	Amount      float64
	TotalCount  int
	MinAmount   float64
	MaxAmount   float64
}
