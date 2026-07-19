package integration

import (
	"context"
	"testing"

	"money-planner/backend/internal/categorization"
	"money-planner/backend/internal/categorization/providers"
)

func TestLLMCategorizationEndToEnd(t *testing.T) {
	mockProvider := providers.NewMockProvider("mock")
	merchantDict := categorization.NewMerchantDictionary()
	confidencer := categorization.NewConfidenceScorer()

	service := categorization.NewCategorizationService(merchantDict, confidencer, nil).
		WithLLMProvider(mockProvider)

	merchantDict.Insert("Swiggy", "Food")
	merchantDict.Insert("Amazon", "Shopping")
	merchantDict.Insert("Uber", "Transport")

	tests := []struct {
		name                  string
		merchant              string
		amount                float64
		expectedCategory      string
		expectedMinConfidence float64
		expectedMethod        string
		expectLLMProvider     bool
	}{
		{
			name:                  "Known merchant uses rule-based",
			merchant:              "Swiggy",
			amount:                300.0,
			expectedCategory:      "Food",
			expectedMinConfidence: 0.9,
			expectedMethod:        "rule_based",
			expectLLMProvider:     false,
		},
		{
			name:                  "Unknown merchant falls back to LLM",
			merchant:              "Aashish Restaurant",
			amount:                250.0,
			expectedCategory:      "Uncategorized",
			expectedMinConfidence: 0.0,
			expectedMethod:        "llm",
			expectLLMProvider:     true,
		},
		{
			name:                  "Partial merchant name falls back to fuzzy match before LLM",
			merchant:              "SWIGGY FD",
			amount:                320.0,
			expectedCategory:      "Food",
			expectedMinConfidence: 0.85,
			expectedMethod:        "fuzzy",
			expectLLMProvider:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.CategorizeTransaction(context.Background(), tt.merchant, tt.amount)

			if result.Category != tt.expectedCategory {
				t.Errorf("expected category %s, got %s", tt.expectedCategory, result.Category)
			}

			if result.Confidence < tt.expectedMinConfidence {
				t.Errorf("expected confidence >= %f, got %f", tt.expectedMinConfidence, result.Confidence)
			}

			if result.Method != tt.expectedMethod {
				t.Errorf("expected method %s, got %s", tt.expectedMethod, result.Method)
			}

			if tt.expectLLMProvider && result.LLMProvider == "" {
				t.Errorf("expected LLM provider to be set, but provider is empty")
			}
			if !tt.expectLLMProvider && result.LLMProvider != "" {
				t.Errorf("expected no LLM usage, but got provider %s", result.LLMProvider)
			}
		})
	}
}

func TestBatchCategorizationWithLLM(t *testing.T) {
	mockProvider := providers.NewMockProvider("mock")
	merchantDict := categorization.NewMerchantDictionary()
	confidencer := categorization.NewConfidenceScorer()

	service := categorization.NewCategorizationService(merchantDict, confidencer, nil).
		WithLLMProvider(mockProvider)

	merchantDict.Insert("Swiggy", "Food")
	merchantDict.Insert("Amazon", "Shopping")

	txns := []categorization.TransactionInput{
		{ID: "1", Merchant: "Swiggy", Amount: 300.0, Timestamp: 1234567890},
		{ID: "2", Merchant: "Unknown Restaurant", Amount: 250.0, Timestamp: 1234567891},
		{ID: "3", Merchant: "Amazon", Amount: 1500.0, Timestamp: 1234567892},
		{ID: "4", Merchant: "Unknown Merchant", Amount: 100.0, Timestamp: 1234567893},
	}

	results := service.CategorizeTransactions(context.Background(), txns)

	if len(results) != 4 {
		t.Errorf("expected 4 results, got %d", len(results))
	}

	if results[0].Category != "Food" || results[0].Method != "rule_based" {
		t.Errorf("transaction 1: expected Food/rule_based, got %s/%s", results[0].Category, results[0].Method)
	}

	if results[1].Category != "Uncategorized" || results[1].Method != "llm" {
		t.Errorf("transaction 2: expected Uncategorized/llm, got %s/%s", results[1].Category, results[1].Method)
	}

	if results[2].Category != "Shopping" || results[2].Method != "rule_based" {
		t.Errorf("transaction 3: expected Shopping/rule_based, got %s/%s", results[2].Category, results[2].Method)
	}

	ruleBasedCount := 0
	llmCount := 0

	for _, r := range results {
		switch r.Method {
		case "rule_based":
			ruleBasedCount++
		case "llm":
			llmCount++
		}
	}

	if ruleBasedCount < 2 {
		t.Errorf("expected at least 2 rule-based categorizations, got %d", ruleBasedCount)
	}
	if llmCount < 2 {
		t.Errorf("expected at least 2 LLM categorizations, got %d", llmCount)
	}
}

func TestLLMConfidenceScores(t *testing.T) {
	merchantDict := categorization.NewMerchantDictionary()
	confidencer := categorization.NewConfidenceScorer()
	mockProvider := providers.NewMockProvider("mock")

	service := categorization.NewCategorizationService(merchantDict, confidencer, nil).
		WithLLMProvider(mockProvider)

	merchantDict.Insert("Swiggy", "Food")

	tests := []struct {
		name     string
		merchant string
		minScore float64
		maxScore float64
	}{
		{
			name:     "Exact match has high confidence",
			merchant: "Swiggy",
			minScore: 0.95,
			maxScore: 1.0,
		},
		{
			name:     "Partial merchant name uses fuzzy match confidence",
			merchant: "SWIGGY FD",
			minScore: 0.85,
			maxScore: 0.99,
		},
		{
			name:     "Unknown merchant via LLM has zero confidence",
			merchant: "Unknown Merchant",
			minScore: 0.0,
			maxScore: 0.1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.CategorizeTransaction(context.Background(), tt.merchant, 100.0)

			if result.Confidence < tt.minScore || result.Confidence > tt.maxScore {
				t.Errorf("expected confidence between %f-%f, got %f",
					tt.minScore, tt.maxScore, result.Confidence)
			}
		})
	}
}
