package contract

import (
	"testing"

	"money-planner/backend/internal/categorization"
	"github.com/stretchr/testify/assert"
)

// T032: Test known merchant (Swiggy) → correct category (Food) with confidence 1.0
func TestCategorizeKnownMerchant(t *testing.T) {
	dict := categorization.NewMerchantDictionary()
	dict.Insert("Swiggy", "Food")

	scorer := categorization.NewConfidenceScorer()
	svc := categorization.NewCategorizationService(dict, scorer)

	result := svc.CategorizeTransaction(nil, "Swiggy", 500.0)

	assert.NotNil(t, result)
	assert.Equal(t, "Food", result.Category)
	assert.Equal(t, 1.0, result.Confidence)
	assert.Equal(t, "rule_based", result.Method)
}

// T033: Test fuzzy match (SWIGGY FD) → Food with confidence 0.85-0.99
func TestCategorizeFuzzyMatch(t *testing.T) {
	dict := categorization.NewMerchantDictionary()
	dict.Insert("Swiggy", "Food")

	scorer := categorization.NewConfidenceScorer()
	svc := categorization.NewCategorizationService(dict, scorer)

	result := svc.CategorizeTransaction(nil, "SWIGGY FD", 500.0)

	assert.NotNil(t, result)
	assert.Equal(t, "Food", result.Category)
	assert.GreaterOrEqual(t, result.Confidence, 0.85)
	assert.LessOrEqual(t, result.Confidence, 0.99)
	assert.Equal(t, "fuzzy", result.Method)
}

// T034: Test unknown merchant → "Uncategorized", method "none"
func TestCategorizeUnknownMerchant(t *testing.T) {
	dict := categorization.NewMerchantDictionary()
	dict.Insert("Swiggy", "Food")

	scorer := categorization.NewConfidenceScorer()
	svc := categorization.NewCategorizationService(dict, scorer)

	result := svc.CategorizeTransaction(nil, "UnknownShop XYZ", 500.0)

	assert.NotNil(t, result)
	assert.Equal(t, "Uncategorized", result.Category)
	assert.Equal(t, 0.0, result.Confidence)
	assert.Equal(t, "none", result.Method)
}

// T035: Test batch categorization (10 txns) → partial results, correct stats
func TestCategorizeBatch(t *testing.T) {
	dict := categorization.NewMerchantDictionary()
	dict.Insert("Swiggy", "Food")
	dict.Insert("Amazon", "Shopping")
	dict.Insert("Uber", "Transport")

	scorer := categorization.NewConfidenceScorer()
	svc := categorization.NewCategorizationService(dict, scorer)

	transactions := []categorization.TransactionInput{
		{ID: "1", Merchant: "Swiggy", Amount: 500},
		{ID: "2", Merchant: "SWIGGY", Amount: 450},           // Fuzzy match
		{ID: "3", Merchant: "Amazon", Amount: 2000},
		{ID: "4", Merchant: "Uber", Amount: 300},
		{ID: "5", Merchant: "UnknownMerchant", Amount: 100},
		{ID: "6", Merchant: "Swiggy FD", Amount: 600},       // Fuzzy match
		{ID: "7", Merchant: "Amazon.in", Amount: 1500},      // Fuzzy match
		{ID: "8", Merchant: "Uber Trip", Amount: 250},       // Fuzzy match
		{ID: "9", Merchant: "", Amount: 0},                  // Empty merchant
		{ID: "10", Merchant: "RandomStore", Amount: 100},
	}

	results := svc.CategorizeTransactions(nil, transactions)

	assert.Equal(t, 10, len(results))

	// Check known merchants
	assert.Equal(t, "Food", results[0].Category)
	assert.Equal(t, 1.0, results[0].Confidence)

	// Check fuzzy matches
	assert.Equal(t, "Food", results[1].Category)
	assert.GreaterOrEqual(t, results[1].Confidence, 0.85)

	// Check uncategorized
	uncategorizedCount := 0
	for _, r := range results {
		if r.Category == "Uncategorized" {
			uncategorizedCount++
		}
	}
	assert.Greater(t, uncategorizedCount, 0)

	// At least 70% should be categorized (7+ out of 10)
	categorizedCount := len(results) - uncategorizedCount
	assert.GreaterOrEqual(t, categorizedCount, 7)
}

// Test confidence scoring extremes
func TestConfidenceScoring(t *testing.T) {
	scorer := categorization.NewConfidenceScorer()

	// Exact match
	assert.Equal(t, 1.0, scorer.ScoreExactMatch())

	// Uncategorized
	assert.Equal(t, 0.0, scorer.ScoreUncategorized())

	// Fuzzy matches at boundaries
	fuzzyMin := scorer.ScoreFuzzyMatch(0.85)
	assert.GreaterOrEqual(t, fuzzyMin, 0.85)

	fuzzyMax := scorer.ScoreFuzzyMatch(1.0)
	assert.LessOrEqual(t, fuzzyMax, 0.99)

	// Validate all scores
	scores := []float64{
		scorer.ScoreExactMatch(),
		scorer.ScoreUncategorized(),
		scorer.ScoreFuzzyMatch(0.90),
	}
	for _, score := range scores {
		assert.True(t, scorer.Validate(score))
	}
}
