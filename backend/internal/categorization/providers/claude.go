package providers

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

// ClaudeProvider implements LLM categorization using Claude API
// Note: This is a stub that requires ANTHROPIC_API_KEY to be configured at runtime
type ClaudeProvider struct {
	apiKey string
	model  string
}

// NewClaudeProvider creates a new Claude provider
func NewClaudeProvider(apiKey string, model string) *ClaudeProvider {
	if model == "" {
		model = "claude-3-5-sonnet-20241022"
	}
	return &ClaudeProvider{
		apiKey: apiKey,
		model:  model,
	}
}

// Categorize categorizes a transaction using Claude API
func (p *ClaudeProvider) Categorize(ctx context.Context, merchant string, amount float64) (category string, confidence float64, explanation string, err error) {
	if p.apiKey == "" {
		return "", 0, "", fmt.Errorf("claude API key not configured")
	}

	// TODO: Implement actual Claude API calls using anthropic-sdk-go
	// For now, return a placeholder to enable provider switching
	prompt := fmt.Sprintf(`Categorize this transaction into ONE of: Food, Transport, Shopping, Entertainment, Bills, Healthcare, Education, Travel, Utilities, Other.

Merchant: %s
Amount: ₹%.2f

Respond in JSON format: {"category": "...", "confidence": 0.0-1.0, "reason": "..."}`, merchant, amount)

	_ = prompt // placeholder for future implementation

	return "Uncategorized", 0.75, "Claude provider not yet fully implemented", nil
}

// Name returns the provider name
func (p *ClaudeProvider) Name() string {
	return "claude"
}

// parseClaudeResponse parses Claude's JSON response
func parseClaudeResponse(response string) (category string, confidence float64, reason string) {
	// Look for JSON structure in response
	jsonStart := strings.Index(response, "{")
	jsonEnd := strings.LastIndex(response, "}")

	if jsonStart == -1 || jsonEnd == -1 || jsonStart > jsonEnd {
		return "", 0, ""
	}

	jsonStr := response[jsonStart : jsonEnd+1]

	// Simple JSON parsing (in production, use proper JSON decoder)
	if cat := extractJSONField(jsonStr, "category"); cat != "" {
		category = cat
	}
	if conf := extractJSONField(jsonStr, "confidence"); conf != "" {
		confidence, _ = strconv.ParseFloat(conf, 64)
	}
	if reason := extractJSONField(jsonStr, "reason"); reason != "" {
		reason = reason
	}

	return
}

// extractJSONField extracts a string field from simple JSON
func extractJSONField(json, field string) string {
	needle := fmt.Sprintf(`"%s":`, field)
	start := strings.Index(json, needle)
	if start == -1 {
		return ""
	}

	start += len(needle)
	// Skip whitespace and opening quote
	for start < len(json) && (json[start] == ' ' || json[start] == '"') {
		start++
	}

	end := start
	for end < len(json) && json[end] != '"' {
		end++
	}

	if start >= len(json) || end >= len(json) {
		return ""
	}

	return json[start:end]
}
