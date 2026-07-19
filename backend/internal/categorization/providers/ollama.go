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

	"money-planner/backend/internal/categorization"
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

// CategorizeBatch categorizes multiple transactions in a single Ollama call
func (p *OllamaProvider) CategorizeBatch(ctx context.Context, items []categorization.BatchItem) ([]categorization.BatchResult, error) {
	if len(items) == 0 {
		return []categorization.BatchResult{}, nil
	}

	// Build batch prompt with all merchants listed by index
	merchantsList := ""
	for i, item := range items {
		merchantsList += fmt.Sprintf("%d. Merchant: %s, Amount: ₹%.2f\n", i, item.Merchant, item.Amount)
	}

	prompt := fmt.Sprintf(`You are a transaction categorization expert. Categorize each transaction into EXACTLY ONE of these categories:
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
4. Respond with ONLY a JSON array, one object per transaction, in the SAME ORDER as listed below, no other text.

Respond in this exact format:
[{"index":0,"category":"...","confidence":0.0-1.0,"reason":"..."},...,{"index":%d,"category":"...","confidence":0.0-1.0,"reason":"..."}]

Transactions:
%s`, len(items)-1, merchantsList)

	reqBody := map[string]interface{}{
		"model":  p.model,
		"prompt": prompt,
		"stream": false,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		// On request marshaling error, degrade all items
		return p.degradeAllItems(items, fmt.Sprintf("failed to marshal request: %v", err)), nil
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/api/generate", bytes.NewReader(bodyBytes))
	if err != nil {
		return p.degradeAllItems(items, fmt.Sprintf("failed to create request: %v", err)), nil
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return p.degradeAllItems(items, fmt.Sprintf("ollama request failed: %v", err)), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return p.degradeAllItems(items, fmt.Sprintf("ollama returned status %d: %s", resp.StatusCode, string(body))), nil
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return p.degradeAllItems(items, fmt.Sprintf("failed to decode response: %v", err)), nil
	}

	responseText, ok := result["response"].(string)
	if !ok {
		return p.degradeAllItems(items, "invalid response format from ollama"), nil
	}

	// Parse batch response
	batchResults, parseErr := p.parseOllamaBatchResponse(responseText, len(items))
	if parseErr != nil {
		return batchResults, nil // Already degraded by parser
	}

	return batchResults, nil
}

// degradeAllItems creates a full result array with all items set to Uncategorized
func (p *OllamaProvider) degradeAllItems(items []categorization.BatchItem, reason string) []categorization.BatchResult {
	results := make([]categorization.BatchResult, len(items))
	for i := range items {
		results[i] = categorization.BatchResult{
			Category:    "Uncategorized",
			Confidence:  0.0,
			Explanation: reason,
			Err:         fmt.Errorf("batch categorization failed"),
		}
	}
	return results
}

// parseOllamaBatchResponse parses Ollama's batch JSON response
func (p *OllamaProvider) parseOllamaBatchResponse(responseText string, expectedLength int) ([]categorization.BatchResult, error) {
	// Pre-fill with degraded results as fallback
	results := make([]categorization.BatchResult, expectedLength)
	for i := range results {
		results[i] = categorization.BatchResult{
			Category:    "Uncategorized",
			Confidence:  0.0,
			Explanation: "Failed to parse Ollama batch response",
			Err:         fmt.Errorf("parse error"),
		}
	}

	// Extract JSON array from response
	jsonStart := strings.Index(responseText, "[")
	jsonEnd := strings.LastIndex(responseText, "]")

	if jsonStart == -1 || jsonEnd == -1 || jsonStart >= jsonEnd {
		return results, fmt.Errorf("no JSON array found in response")
	}

	jsonStr := responseText[jsonStart : jsonEnd+1]

	// Parse JSON array
	var items []map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &items); err != nil {
		return results, fmt.Errorf("failed to unmarshal batch response: %w", err)
	}

	// Map results by index
	for _, item := range items {
		// Get index
		var index int
		if idx, ok := item["index"].(float64); ok {
			index = int(idx)
		} else {
			continue // Skip items without valid index
		}

		// Validate index range
		if index < 0 || index >= expectedLength {
			continue // Skip out-of-range indices
		}

		// Extract category
		var category string
		if cat, ok := item["category"].(string); ok {
			category = strings.TrimSpace(cat)
		}

		// Extract confidence
		var confidence float64
		if conf, ok := item["confidence"].(float64); ok {
			confidence = conf
			if confidence < 0.0 {
				confidence = 0.0
			}
			if confidence > 1.0 {
				confidence = 1.0
			}
		}

		// Extract reason
		var reason string
		if r, ok := item["reason"].(string); ok {
			reason = strings.TrimSpace(r)
		}

		// Validate and normalize category
		if category != "" {
			validCategory := normalizeCategory(category)
			if validCategory != "" {
				results[index] = categorization.BatchResult{
					Category:    validCategory,
					Confidence:  confidence,
					Explanation: reason,
					Err:         nil,
				}
			}
		}
	}

	return results, nil
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
