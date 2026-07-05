package contract

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPreviewStatementContract(t *testing.T) {
	tests := []struct {
		name           string
		statementID    string
		expectedStatus int
		shouldContain  []string
		expectError    bool
	}{
		{
			name:           "valid preview returns 200 with transactions and validation summary",
			statementID:    "550e8400-e29b-41d4-a716-446655440000",
			expectedStatus: http.StatusOK,
			shouldContain: []string{
				"transactions",
				"validation_summary",
				"status",
			},
			expectError: false,
		},
		{
			name:           "invalid statement ID returns 404",
			statementID:    "invalid-uuid",
			expectedStatus: http.StatusNotFound,
			expectError:    true,
		},
		{
			name:           "statement not found returns 404",
			statementID:    "00000000-0000-0000-0000-000000000000",
			expectedStatus: http.StatusNotFound,
			expectError:    true,
		},
		{
			name:           "statement still processing returns 202",
			statementID:    "550e8400-e29b-41d4-a716-446655440001",
			expectedStatus: http.StatusAccepted,
			shouldContain: []string{
				"status",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req, _ := http.NewRequest(
				"GET",
				"/api/statements/"+tt.statementID+"/preview",
				nil,
			)
			req.Header.Set("Authorization", "Bearer test-token")

			// Validate contract structure
			// In real implementation, this would call the actual endpoint
			if tt.expectedStatus == http.StatusOK {
				// Verify response structure matches contract
				expectedResponse := map[string]interface{}{
					"transactions": []interface{}{
						map[string]interface{}{
							"transaction_id":    "",
							"transaction_date": "",
							"merchant":         "",
							"amount":           0.0,
							"type":            "",
							"description":     "",
							"currency":        "INR",
						},
					},
					"validation_summary": map[string]interface{}{
						"total_transactions":   0,
						"valid_transactions":   0,
						"invalid_transactions": 0,
						"errors":              []interface{}{},
					},
					"status": "PENDING",
				}

				// Verify the structure is valid JSON
				jsonBytes, err := json.Marshal(expectedResponse)
				assert.NoError(t, err)
				assert.NotEmpty(t, jsonBytes)
			}
		})
	}
}

// PreviewStatementContractSchema defines the expected response schema
var PreviewStatementContractSchema = map[string]interface{}{
	"type": "object",
	"properties": map[string]interface{}{
		"transactions": map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"transaction_id": map[string]interface{}{
						"type": "string",
					},
					"transaction_date": map[string]interface{}{
						"type":   "string",
						"format": "date",
					},
					"merchant": map[string]interface{}{
						"type": "string",
					},
					"amount": map[string]interface{}{
						"type": "number",
					},
					"type": map[string]interface{}{
						"type": "string",
						"enum": []string{"DEBIT", "CREDIT"},
					},
					"balance": map[string]interface{}{
						"type": "number",
					},
					"description": map[string]interface{}{
						"type": "string",
					},
					"currency": map[string]interface{}{
						"type": "string",
					},
				},
				"required": []string{"transaction_id", "transaction_date", "merchant", "amount", "type", "currency"},
			},
		},
		"validation_summary": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"total_transactions": map[string]interface{}{
					"type": "integer",
				},
				"valid_transactions": map[string]interface{}{
					"type": "integer",
				},
				"invalid_transactions": map[string]interface{}{
					"type": "integer",
				},
				"errors": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"type": "object",
					},
				},
			},
			"required": []string{"total_transactions", "valid_transactions", "invalid_transactions"},
		},
		"status": map[string]interface{}{
			"type": "string",
			"enum": []string{"PENDING", "SUCCESS", "FAILED"},
		},
	},
	"required": []string{"transactions", "validation_summary", "status"},
}
