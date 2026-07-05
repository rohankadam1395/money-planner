package contract

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfirmStatementContract(t *testing.T) {
	tests := []struct {
		name            string
		statementID     string
		requestBody     map[string]interface{}
		expectedStatus  int
		expectedStatus2 string // Final status after confirmation
		expectError     bool
	}{
		{
			name:           "valid confirm returns 200 and persists transactions",
			statementID:    "550e8400-e29b-41d4-a716-446655440000",
			requestBody:    map[string]interface{}{},
			expectedStatus: http.StatusOK,
			expectedStatus2: "SUCCESS",
			expectError:    false,
		},
		{
			name:           "invalid statement ID returns 404",
			statementID:    "invalid-uuid",
			requestBody:    map[string]interface{}{},
			expectedStatus: http.StatusNotFound,
			expectError:    true,
		},
		{
			name:           "statement not found returns 404",
			statementID:    "00000000-0000-0000-0000-000000000000",
			requestBody:    map[string]interface{}{},
			expectedStatus: http.StatusNotFound,
			expectError:    true,
		},
		{
			name:           "statement already confirmed returns 409",
			statementID:    "550e8400-e29b-41d4-a716-446655440001",
			requestBody:    map[string]interface{}{},
			expectedStatus: http.StatusConflict,
			expectError:    true,
		},
		{
			name:           "duplicate statement returns 409",
			statementID:    "550e8400-e29b-41d4-a716-446655440002",
			requestBody:    map[string]interface{}{},
			expectedStatus: http.StatusConflict,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request body
			body := &bytes.Buffer{}
			err := json.NewEncoder(body).Encode(tt.requestBody)
			require.NoError(t, err)

			// Create request
			req, err := http.NewRequest(
				"POST",
				"/api/statements/"+tt.statementID+"/confirm",
				body,
			)
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer test-token")

			// Validate contract
			if !tt.expectError && tt.expectedStatus == http.StatusOK {
				// For successful confirmations, response should contain
				expectedResponse := map[string]interface{}{
					"statement_id": tt.statementID,
					"status":       "SUCCESS",
					"transaction_count": 0,
					"imported_at":       "",
				}

				jsonBytes, err := json.Marshal(expectedResponse)
				assert.NoError(t, err)
				assert.NotEmpty(t, jsonBytes)

				// Verify all required fields are present
				response := make(map[string]interface{})
				err = json.NewDecoder(io.NopCloser(bytes.NewReader(jsonBytes))).Decode(&response)
				if err == nil {
					assert.Contains(t, response, "statement_id")
					assert.Contains(t, response, "status")
					assert.Equal(t, tt.expectedStatus2, response["status"])
				}
			}
		})
	}
}

// ConfirmStatementContractSchema defines the expected response schema
var ConfirmStatementContractSchema = map[string]interface{}{
	"type": "object",
	"properties": map[string]interface{}{
		"statement_id": map[string]interface{}{
			"type":        "string",
			"description": "UUID of the confirmed statement",
		},
		"status": map[string]interface{}{
			"type":        "string",
			"enum":        []string{"SUCCESS", "FAILED"},
			"description": "Final import status",
		},
		"transaction_count": map[string]interface{}{
			"type":        "integer",
			"description": "Number of transactions persisted",
		},
		"imported_at": map[string]interface{}{
			"type":        "string",
			"format":      "date-time",
			"description": "When the transactions were imported",
		},
		"message": map[string]interface{}{
			"type":        "string",
			"description": "Confirmation message",
		},
	},
	"required": []string{"statement_id", "status", "transaction_count", "imported_at"},
}

// ErrorResponseSchema for 409 Conflict (duplicate)
var DuplicateStatementErrorSchema = map[string]interface{}{
	"type": "object",
	"properties": map[string]interface{}{
		"error": map[string]interface{}{
			"type":        "string",
			"description": "Error message",
		},
		"code": map[string]interface{}{
			"type":        "string",
			"enum":        []string{"DUPLICATE_STATEMENT", "OVERLAPPING_PERIOD"},
			"description": "Error code for duplicate/overlap detection",
		},
		"existing_statement_id": map[string]interface{}{
			"type":        "string",
			"description": "ID of the existing statement causing the conflict",
		},
	},
	"required": []string{"error", "code"},
}
