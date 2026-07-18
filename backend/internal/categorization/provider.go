package categorization

import "context"

// LLMProvider is the interface for LLM-based categorization
type LLMProvider interface {
	// Categorize uses the LLM to categorize a transaction
	// Returns category name, confidence score (0-1), explanation, and error
	Categorize(ctx context.Context, merchant string, amount float64) (category string, confidence float64, explanation string, err error)

	// Name returns the provider name (e.g., "ollama", "claude", "openai")
	Name() string
}
