package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// OllamaProvider implements LLM categorization using Ollama
type OllamaProvider struct {
	baseURL    string
	model      string
	httpClient *http.Client
}

// NewOllamaProvider creates a new Ollama provider
func NewOllamaProvider(baseURL string, model string) *OllamaProvider {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	if model == "" {
		model = "mistral"
	}
	return &OllamaProvider{
		baseURL: baseURL,
		model:   model,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Categorize categorizes a transaction using Ollama
func (p *OllamaProvider) Categorize(ctx context.Context, merchant string, amount float64) (category string, confidence float64, explanation string, err error) {
	prompt := fmt.Sprintf(`You are a transaction categorization expert. Categorize this transaction into EXACTLY ONE of these categories:
- Food & Dining
- Shopping
- Transport
- Housing
- Utilities
- Entertainment
- Income
- Healthcare
- Education
- Miscellaneous

Rules:
1. You MUST choose from the list above. Do NOT invent categories.
2. If uncertain, choose "Miscellaneous".
3. Be consistent with the exact category names listed above.

Transaction:
Merchant: %s
Amount: ₹%.2f

Respond in this exact format:
Category: [category name from the list above]
Confidence: [0.0-1.0]
Reason: [brief reason]`, merchant, amount)

	reqBody := map[string]interface{}{
		"model":  p.model,
		"prompt": prompt,
		"stream": false,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", 0, "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/api/generate", bytes.NewReader(bodyBytes))
	if err != nil {
		return "", 0, "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", 0, "", fmt.Errorf("ollama request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", 0, "", fmt.Errorf("ollama returned status %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", 0, "", fmt.Errorf("failed to decode response: %w", err)
	}

	responseText, ok := result["response"].(string)
	if !ok {
		return "", 0, "", fmt.Errorf("invalid response format from ollama")
	}

	// Parse the response
	category, confidence, reason := parseOllamaResponse(responseText)
	if category == "" {
		return "Uncategorized", 0.0, "Failed to parse Ollama response", nil
	}

	// Validate and normalize the category
	validCategory := normalizeCategory(category)
	if validCategory == "" {
		// LLM returned invalid category, default to Miscellaneous
		return "Miscellaneous", confidence * 0.5, fmt.Sprintf("LLM returned invalid category '%s', defaulting to Miscellaneous", category), nil
	}

	return validCategory, confidence, reason, nil
}

// Name returns the provider name
func (p *OllamaProvider) Name() string {
	return "ollama"
}

// parseOllamaResponse parses the Ollama response to extract category, confidence, and reason
func parseOllamaResponse(response string) (category string, confidence float64, reason string) {
	lines := strings.Split(response, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(strings.ToLower(line), "category:") {
			category = strings.TrimSpace(strings.TrimPrefix(strings.ToLower(line), "category:"))
			category = capitalize(category)
		}

		if strings.HasPrefix(strings.ToLower(line), "confidence:") {
			confStr := strings.TrimSpace(strings.TrimPrefix(strings.ToLower(line), "confidence:"))
			if conf, err := strconv.ParseFloat(confStr, 64); err == nil {
				confidence = conf
			}
		}

		if strings.HasPrefix(strings.ToLower(line), "reason:") {
			reason = strings.TrimSpace(strings.TrimPrefix(strings.ToLower(line), "reason:"))
		}
	}
	return
}

// capitalize capitalizes the first letter of each word
func capitalize(s string) string {
	parts := strings.Fields(s)
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(string(p[0])) + strings.ToLower(p[1:])
		}
	}
	return strings.Join(parts, " ")
}

// normalizeCategory validates and normalizes category names to match database categories
func normalizeCategory(category string) string {
	validCategories := []string{
		"Food & Dining",
		"Shopping",
		"Transport",
		"Housing",
		"Utilities",
		"Entertainment",
		"Income",
		"Healthcare",
		"Education",
		"Miscellaneous",
	}

	// Exact match
	for _, valid := range validCategories {
		if strings.EqualFold(category, valid) {
			return valid
		}
	}

	// Fuzzy mapping for common LLM mistakes
	lowerCat := strings.ToLower(category)
	mappings := map[string]string{
		"food":           "Food & Dining",
		"dining":         "Food & Dining",
		"restaurant":     "Food & Dining",
		"grocery":        "Shopping",
		"retail":         "Shopping",
		"taxi":           "Transport",
		"uber":           "Transport",
		"travel":         "Transport",
		"electricity":    "Utilities",
		"water":          "Utilities",
		"bills":          "Utilities",
		"movie":          "Entertainment",
		"games":          "Entertainment",
		"salary":         "Income",
		"wages":          "Income",
		"hospital":       "Healthcare",
		"medicine":       "Healthcare",
		"doctor":         "Healthcare",
		"school":         "Education",
		"course":         "Education",
		"other":          "Miscellaneous",
		"misc":           "Miscellaneous",
	}

	if normalized, ok := mappings[lowerCat]; ok {
		return normalized
	}

	// If no match, return empty to indicate invalid
	return ""
}
