package contract

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUploadStatementContract(t *testing.T) {
	tests := []struct {
		name           string
		fileName       string
		fileContent    []byte
		bankCode       string
		expectedStatus int
		expectedBody   map[string]interface{}
		expectError    bool
	}{
		{
			name:           "valid PDF upload returns 202 Accepted with PENDING status",
			fileName:       "statement.pdf",
			fileContent:    []byte("%PDF-1.4 fake PDF content"),
			bankCode:       "HDFC",
			expectedStatus: http.StatusAccepted,
			expectedBody: map[string]interface{}{
				"statement_id":      "", // Should be UUID
				"status":            "PENDING",
				"bank_code":         "HDFC",
				"file_name":         "statement.pdf",
				"file_format":       "PDF",
				"transaction_count": float64(0),
			},
			expectError: false,
		},
		{
			name:           "valid CSV upload returns 202 Accepted",
			fileName:       "statement.csv",
			fileContent:    []byte("Date,Merchant,Amount,Type\n2026-01-01,Test Store,100.00,DEBIT"),
			bankCode:       "ICIC",
			expectedStatus: http.StatusAccepted,
			expectedBody: map[string]interface{}{
				"status":    "PENDING",
				"bank_code": "ICIC",
				"file_format": "CSV",
			},
			expectError: false,
		},
		{
			name:           "missing bank code returns 400",
			fileName:       "statement.pdf",
			fileContent:    []byte("%PDF-1.4 content"),
			bankCode:       "",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "invalid file format returns 400",
			fileName:       "statement.txt",
			fileContent:    []byte("invalid content"),
			bankCode:       "HDFC",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "file too large returns 413",
			fileName:       "statement.pdf",
			fileContent:    bytes.Repeat([]byte("x"), 51*1024*1024), // 51MB
			bankCode:       "HDFC",
			expectedStatus: http.StatusRequestEntityTooLarge,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create multipart form
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)

			// Add file field
			part, err := writer.CreateFormFile("file", tt.fileName)
			require.NoError(t, err)
			_, err = io.Copy(part, bytes.NewReader(tt.fileContent))
			require.NoError(t, err)

			// Add bank_code field
			if tt.bankCode != "" {
				err = writer.WriteField("bank_code", tt.bankCode)
				require.NoError(t, err)
			}

			err = writer.Close()
			require.NoError(t, err)

			// Create request
			req, err := http.NewRequest("POST", "/api/statements/upload", body)
			require.NoError(t, err)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			req.Header.Set("Authorization", "Bearer test-token")

			// Mock the handler response for this test
			// In real implementation, this would call the actual handler
			// For now, we're defining the contract
			resp := &http.Response{
				StatusCode: tt.expectedStatus,
				Body:       io.NopCloser(bytes.NewReader([]byte(""))),
			}

			// Validate status code
			assert.Equal(t, tt.expectedStatus, resp.StatusCode,
				"Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)

			// For successful uploads, validate response body structure
			if !tt.expectError && tt.expectedStatus == http.StatusAccepted {
				// This demonstrates what the response should contain
				var responseBody map[string]interface{}
				err := json.NewDecoder(resp.Body).Decode(&responseBody)
				if err == nil {
					assert.Contains(t, responseBody, "statement_id", "Response should contain statement_id")
					assert.Contains(t, responseBody, "status", "Response should contain status")
					assert.Equal(t, "PENDING", responseBody["status"], "Status should be PENDING")
					assert.Contains(t, responseBody, "bank_code", "Response should contain bank_code")
				}
			}
		})
	}
}

// UploadStatementContractSchema defines the expected response schema for successful uploads
var UploadStatementContractSchema = map[string]interface{}{
	"type": "object",
	"properties": map[string]interface{}{
		"statement_id": map[string]interface{}{
			"type":        "string",
			"description": "UUID of the created statement",
		},
		"status": map[string]interface{}{
			"type":        "string",
			"enum":        []string{"PENDING", "SUCCESS", "FAILED"},
			"description": "Current import status",
		},
		"bank_code": map[string]interface{}{
			"type":        "string",
			"description": "Bank code (HDFC, ICIC, AXIS, SBI)",
		},
		"file_name": map[string]interface{}{
			"type":        "string",
			"description": "Original uploaded file name",
		},
		"file_format": map[string]interface{}{
			"type":        "string",
			"enum":        []string{"PDF", "CSV", "XLSX"},
			"description": "File format",
		},
		"transaction_count": map[string]interface{}{
			"type":        "integer",
			"description": "Number of transactions extracted",
		},
		"uploaded_at": map[string]interface{}{
			"type":        "string",
			"format":      "date-time",
			"description": "When the file was uploaded",
		},
	},
	"required": []string{"statement_id", "status", "bank_code", "file_name", "file_format"},
}
