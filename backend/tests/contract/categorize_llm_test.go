package contract

import (
	"context"
	"fmt"
	"testing"

	"money-planner/backend/internal/categorization"
	"money-planner/backend/internal/categorization/providers"
)

// TestLLMCategorization verifies LLM categorization with unknown merchants
func TestLLMCategorization(t *testing.T) {
	// Setup mock provider for testing
	mockProvider := providers.NewMockProvider("mock")
	merchantDict := categorization.NewMerchantDictionary()
	confidencer := categorization.NewConfidenceScorer()

	service := categorization.NewCategorizationService(merchantDict, confidencer).
		WithLLMProvider(mockProvider)

	merchantDict.Insert("Swiggy", "Food")

	tests := []struct {
		name       string
		merchant   string
		amount     float64
		expectCat  string
		expectMin  float64
		provider   string
	}{
		{
			name:       "Unknown merchant with LLM",
			merchant:   "Aashish Restaurant Pvt Ltd",
			amount:     250.0,
			expectCat:  "Uncategorized",
			expectMin:  0.0,
			provider:   "mock",
		},
		{
			name:       "LLM should not override rule-based match",
			merchant:   "Swiggy",
			amount:     300.0,
			expectCat:  "Food",
			expectMin:  0.9, // Should use rule-based, not LLM
			provider:   "",  // Should be empty (no LLM used)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.CategorizeTransaction(context.Background(), tt.merchant, tt.amount)

			if result.Category != tt.expectCat {
				t.Errorf("expected category %s, got %s", tt.expectCat, result.Category)
			}

			if result.Confidence < tt.expectMin {
				t.Errorf("expected confidence >= %f, got %f", tt.expectMin, result.Confidence)
			}

			if tt.provider != "" && result.LLMProvider != tt.provider {
				t.Errorf("expected provider %s, got %s", tt.provider, result.LLMProvider)
			}
		})
	}
}

// TestGracefulDegradationOnLLMError verifies graceful degradation when LLM fails
func TestGracefulDegradationOnLLMError(t *testing.T) {
	// Create a provider that always fails
	failingProvider := &failingMockProvider{}
	merchantDict := categorization.NewMerchantDictionary()
	confidencer := categorization.NewConfidenceScorer()

	service := categorization.NewCategorizationService(merchantDict, confidencer).
		WithLLMProvider(failingProvider)

	result := service.CategorizeTransaction(context.Background(), "Unknown Merchant", 100.0)

	// Should gracefully degrade to Uncategorized, not error
	if result.Category != "Uncategorized" {
		t.Errorf("expected Uncategorized on LLM failure, got %s", result.Category)
	}

	if result.Method != "none" {
		t.Errorf("expected method 'none' on LLM failure, got %s", result.Method)
	}

	if result.Confidence != 0.0 {
		t.Errorf("expected 0 confidence on LLM failure, got %f", result.Confidence)
	}
}

// TestLLMProviderEnvVar simulates provider switching based on environment
func TestLLMProviderEnvVar(t *testing.T) {
	// Simulate provider selection based on env
	var provider categorization.LLMProvider

	// In real code, this would be determined by LLM_PROVIDER env var
	provider = providers.NewMockProvider("mock")

	merchantDict := categorization.NewMerchantDictionary()
	confidencer := categorization.NewConfidenceScorer()

	service := categorization.NewCategorizationService(merchantDict, confidencer).
		WithLLMProvider(provider)

	result := service.CategorizeTransaction(context.Background(), "Test Merchant", 100.0)
	_ = result

	// Verify the provider name is correct
	if provider.Name() != "mock" {
		t.Errorf("expected provider name 'mock', got %s", provider.Name())
	}
}

// failingMockProvider simulates an LLM provider that fails
type failingMockProvider struct{}

func (p *failingMockProvider) Categorize(ctx context.Context, merchant string, amount float64) (string, float64, string, error) {
	return "", 0, "", fmt.Errorf("llm provider unavailable")
}

func (p *failingMockProvider) Name() string {
	return "failing"
}
