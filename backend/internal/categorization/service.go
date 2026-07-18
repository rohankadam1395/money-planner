package categorization

import (
	"context"
	"fmt"
)

// CategorizationService handles transaction categorization
type CategorizationService struct {
	merchantDict *MerchantDictionary
	confidencer  *ConfidenceScorer
}

// NewCategorizationService creates a new categorization service
func NewCategorizationService(merchantDict *MerchantDictionary, confidencer *ConfidenceScorer) *CategorizationService {
	return &CategorizationService{
		merchantDict: merchantDict,
		confidencer:  confidencer,
	}
}

// CategorizeTransaction categorizes a single transaction
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

	// No match found
	return &CategorizationResult{
		Category:   "Uncategorized",
		Method:     "none",
		Confidence: 0.0,
		Reason:     fmt.Sprintf("No matching merchant for: %s", merchant),
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
	matchDistance float64 // internal field for fuzzy matching
}
