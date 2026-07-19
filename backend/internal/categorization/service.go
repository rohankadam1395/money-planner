package categorization

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/google/uuid"
)

// CategorizationService handles transaction categorization
type CategorizationService struct {
	merchantDict *MerchantDictionary
	confidencer  *ConfidenceScorer
	llmProvider  LLMProvider
	dbConn       *sql.DB
}

// NewCategorizationService creates a new categorization service
func NewCategorizationService(merchantDict *MerchantDictionary, confidencer *ConfidenceScorer, dbConn *sql.DB) *CategorizationService {
	return &CategorizationService{
		merchantDict: merchantDict,
		confidencer:  confidencer,
		llmProvider:  nil,
		dbConn:       dbConn,
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
		// Accept fuzzy match if confidence is high (≥75%), skip expensive LLM call
		if confidence >= 0.75 {
			return result
		}
		// Low confidence fuzzy match - try LLM for better accuracy
	}

	// Try LLM categorization if available
	if s.llmProvider != nil {
		return s.CategorizeLLM(ctx, merchant, amount)
	}

	// Low-confidence fuzzy match with no LLM available
	if result != nil {
		return result
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

// CategorizeTransactions categorizes multiple transactions (with batched LLM fallback if available)
func (s *CategorizationService) CategorizeTransactions(ctx context.Context, transactions []TransactionInput) []CategorizationResult {
	results := make([]CategorizationResult, len(transactions))
	var needsLLM []int // indices that still need LLM categorization

	// First pass: resolve rule-based/fuzzy matches
	for i, txn := range transactions {
		results[i] = *s.resolveRuleBased(txn.Merchant, txn.Amount)
		if results[i].Method == "none" && s.llmProvider != nil {
			needsLLM = append(needsLLM, i)
		}
	}

	// If nothing needs LLM or no provider, return early
	if len(needsLLM) == 0 || s.llmProvider == nil {
		return results
	}

	// Try to use categorization batch interface if available
	type batchLLMProvider interface {
		CategorizeBatch(ctx context.Context, items []BatchItem) ([]BatchResult, error)
	}

	if batchProv, ok := s.llmProvider.(batchLLMProvider); ok {
		s.categorizeBatch(ctx, batchProv, transactions, results, needsLLM)
		return results
	}

	// Fallback: sequential single-transaction LLM calls (for providers that don't support batch)
	for _, i := range needsLLM {
		results[i] = *s.CategorizeLLM(ctx, transactions[i].Merchant, transactions[i].Amount)
	}

	return results
}

// resolveRuleBased resolves a transaction using rule-based and fuzzy matching only
func (s *CategorizationService) resolveRuleBased(merchant string, amount float64) *CategorizationResult {
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
		// Accept fuzzy match if confidence is high (≥75%)
		if confidence >= 0.75 {
			return result
		}
		// Low confidence fuzzy match stays tentative; LLM will be tried if available
		return result
	}

	// No rule-based match found
	return &CategorizationResult{
		Category:   "Uncategorized",
		Method:     "none",
		Confidence: 0.0,
		Reason:     fmt.Sprintf("No matching merchant for: %s", merchant),
	}
}

// categorizeBatch calls the provider's batch method and maps results back by original index
// batchProvider must implement CategorizeBatch(ctx, []BatchItem) ([]BatchResult, error)
func (s *CategorizationService) categorizeBatch(ctx context.Context, batchProvider interface {
	CategorizeBatch(ctx context.Context, items []BatchItem) ([]BatchResult, error)
}, transactions []TransactionInput, results []CategorizationResult, needsLLM []int) {
	// Build batch items for LLM
	batchItems := make([]BatchItem, len(needsLLM))
	for j, i := range needsLLM {
		batchItems[j] = BatchItem{
			Merchant: transactions[i].Merchant,
			Amount:   transactions[i].Amount,
		}
	}

	// Call batch categorization
	batchResults, err := batchProvider.CategorizeBatch(ctx, batchItems)
	if err != nil {
		// On batch call error, degrade all needs-LLM indices to Uncategorized
		log.Printf("Batch LLM categorization failed: %v", err)
		for _, i := range needsLLM {
			results[i] = CategorizationResult{
				Category:   "Uncategorized",
				Method:     "none",
				Confidence: 0.0,
				Reason:     fmt.Sprintf("LLM error (graceful degradation): %v", err),
			}
		}
		return
	}

	// Map batch results back to original indices
	for j, i := range needsLLM {
		if j >= len(batchResults) {
			// Short result array; degrade this item
			results[i] = CategorizationResult{
				Category:   "Uncategorized",
				Method:     "none",
				Confidence: 0.0,
				Reason:     "Batch result mismatch",
			}
			continue
		}

		batchResult := batchResults[j]
		if batchResult.Err != nil {
			// Per-item error in batch result
			results[i] = CategorizationResult{
				Category:   "Uncategorized",
				Method:     "none",
				Confidence: 0.0,
				Reason:     fmt.Sprintf("LLM error (graceful degradation): %v", batchResult.Err),
			}
			continue
		}

		// Successful batch result
		results[i] = CategorizationResult{
			Category:    batchResult.Category,
			Method:      "llm",
			Confidence:  batchResult.Confidence,
			Reason:      batchResult.Explanation,
			LLMProvider: s.llmProvider.Name(),
		}
	}
}

// CategorizeTransactionsRuleBasedOnly categorizes using only exact + fuzzy matching (no LLM)
func (s *CategorizationService) CategorizeTransactionsRuleBasedOnly(ctx context.Context, transactions []TransactionInput) []CategorizationResult {
	results := make([]CategorizationResult, len(transactions))
	for i, txn := range transactions {
		results[i] = *s.categorizeRuleBasedOnly(ctx, txn.Merchant, txn.Amount)
	}
	return results
}

// categorizeRuleBasedOnly is the rule-based-only variant of CategorizeTransaction
func (s *CategorizationService) categorizeRuleBasedOnly(ctx context.Context, merchant string, amount float64) *CategorizationResult {
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

	// Try fuzzy match (accept any fuzzy match, no LLM fallback)
	result = s.merchantDict.LookupFuzzy(merchant)
	if result != nil {
		confidence := s.confidencer.ScoreFuzzyMatch(result.matchDistance)
		result.Confidence = confidence
		result.Method = "fuzzy"
		result.Reason = fmt.Sprintf("Fuzzy match: %s (distance: %.2f)", result.Category, result.matchDistance)
		return result
	}

	// No rule-based match found
	return &CategorizationResult{
		Category:   "Uncategorized",
		Method:     "none",
		Confidence: 0.0,
		Reason:     fmt.Sprintf("No rule-based match for: %s", merchant),
	}
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

// BatchItem represents a single transaction for batch LLM categorization
type BatchItem struct {
	Merchant string
	Amount   float64
}

// BatchResult represents the result of one transaction in batch categorization
type BatchResult struct {
	Category    string  // Category name or "Uncategorized" on error
	Confidence  float64 // 0.0-1.0
	Explanation string  // Reason for categorization
	Err         error   // Per-item error; nil means success
}

// UpdateCategoryStats updates category_stats for a category after recategorization
// Called when a transaction is recategorized to both the old and new category
func (s *CategorizationService) UpdateCategoryStats(ctx context.Context, userID, categoryID, period string) error {
	if s.dbConn == nil {
		return fmt.Errorf("database connection not initialized")
	}

	var totalSpent sql.NullFloat64
	var count sql.NullInt64
	var avgTransaction sql.NullFloat64

	// Calculate current stats from transactions
	err := s.dbConn.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(t.amount), 0), COUNT(*), COALESCE(AVG(t.amount), 0)
		 FROM transactions t
		 JOIN transaction_categories tc ON t.transaction_id = tc.transaction_id
		 WHERE t.user_id = $1 AND tc.category_id = $2 AND DATE_TRUNC('month', t.transaction_date::timestamp)::text LIKE $3 || '%'`,
		userID, categoryID, period,
	).Scan(&totalSpent, &count, &avgTransaction)

	if err != nil && err != sql.ErrNoRows {
		log.Printf("Error calculating category stats for user %s, category %s, period %s: %v", userID, categoryID, period, err)
		return err
	}

	// Upsert into category_stats
	_, err = s.dbConn.ExecContext(ctx,
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

	if err != nil {
		log.Printf("Error updating category_stats for user %s, category %s, period %s: %v", userID, categoryID, period, err)
		return err
	}

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
