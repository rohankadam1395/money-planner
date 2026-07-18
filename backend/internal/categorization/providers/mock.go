package providers

import "context"

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
