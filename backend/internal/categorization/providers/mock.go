package providers

import (
	"context"

	"money-planner/backend/internal/categorization"
)

// MockProvider implements LLMProvider for testing
type MockProvider struct {
	name    string
	mapping map[string]struct {
		category   string
		confidence float64
	}
}

// NewMockProvider creates a new mock provider for testing
func NewMockProvider(name string) *MockProvider {
	return &MockProvider{
		name: name,
		mapping: map[string]struct {
			category   string
			confidence float64
		}{
			"swiggy":  {"Food", 0.9},
			"amazon":  {"Shopping", 0.95},
			"uber":    {"Transport", 0.92},
			"netflix": {"Entertainment", 0.88},
		},
	}
}

// Categorize returns mock categorization results
func (p *MockProvider) Categorize(ctx context.Context, merchant string, amount float64) (category string, confidence float64, explanation string, err error) {
	merchant = capitalize(merchant)

	if result, ok := p.mapping[merchant]; ok {
		return result.category, result.confidence, "Mock categorization", nil
	}

	return "Uncategorized", 0.0, "No mock mapping found", nil
}

// Name returns the provider name
func (p *MockProvider) Name() string {
	return p.name
}

// CategorizeBatch categorizes multiple transactions using the mock mapping
func (p *MockProvider) CategorizeBatch(ctx context.Context, items []categorization.BatchItem) ([]categorization.BatchResult, error) {
	results := make([]categorization.BatchResult, len(items))
	for i, item := range items {
		merchant := capitalize(item.Merchant)
		if result, ok := p.mapping[merchant]; ok {
			results[i] = categorization.BatchResult{
				Category:    result.category,
				Confidence:  result.confidence,
				Explanation: "Mock categorization",
				Err:         nil,
			}
		} else {
			results[i] = categorization.BatchResult{
				Category:    "Uncategorized",
				Confidence:  0.0,
				Explanation: "No mock mapping found",
				Err:         nil,
			}
		}
	}
	return results, nil
}
