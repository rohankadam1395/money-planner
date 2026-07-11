package contract

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"money-planner/backend/internal/api"
)

// TestMultiBankStatementQuery tests querying transactions across multiple banks
// Contract: GET /api/transactions?bank_code=HDFC,ICICI&date_from=2024-01-01&date_to=2024-12-31
// Expected: Unified transaction list from both banks, chronologically sorted
func TestMultiBankStatementQuery(t *testing.T) {
	// Setup router and handlers
	router := api.SetupRouter()

	// Mock request: Get transactions across HDFC and ICICI
	req, err := http.NewRequest("GET",
		"/api/transactions?bank_code=HDFC,ICICI&date_from=2024-01-01&date_to=2024-12-31",
		nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Add authorization header
	req.Header.Set("Authorization", "Bearer test-token")

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify response contract
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
		t.Logf("Response body: %s", w.Body.String())
		return
	}

	// Parse response
	var response struct {
		Data []struct {
			ID          string  `json:"id"`
			Date        string  `json:"date"`
			Merchant    string  `json:"merchant"`
			Amount      float64 `json:"amount"`
			Type        string  `json:"type"`
			Bank        string  `json:"bank"`
			Currency    string  `json:"currency"`
			Description string  `json:"description"`
		} `json:"data"`
		Pagination struct {
			Total  int `json:"total"`
			Limit  int `json:"limit"`
			Offset int `json:"offset"`
		} `json:"pagination"`
	}

	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to parse response JSON: %v", err)
		return
	}

	// Verify contract: response contains transaction array with bank field
	if len(response.Data) == 0 {
		t.Log("Note: No transactions returned (expected if database empty)")
		return
	}

	// Verify bank codes are included
	for _, txn := range response.Data {
		if txn.Bank == "" {
			t.Errorf("Transaction missing bank field: %+v", txn)
		}
		if txn.Bank != "HDFC" && txn.Bank != "ICICI" {
			t.Errorf("Unexpected bank code: %s", txn.Bank)
		}
	}

	// Verify pagination contract
	if response.Pagination.Total < 0 {
		t.Errorf("Invalid pagination total: %d", response.Pagination.Total)
	}

	t.Logf("✓ Multi-bank query contract verified: %d transactions from multiple banks", response.Pagination.Total)
}

// TestMultiBankFilterByBank tests filtering transactions by specific bank code
// Contract: GET /api/transactions?bank_code=HDFC
// Expected: Only transactions from HDFC bank
func TestMultiBankFilterByBank(t *testing.T) {
	router := api.SetupRouter()

	req, err := http.NewRequest("GET",
		"/api/transactions?bank_code=HDFC",
		nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
		return
	}

	var response struct {
		Data []struct {
			Bank string `json:"bank"`
		} `json:"data"`
	}

	json.Unmarshal(w.Body.Bytes(), &response)

	// Verify all transactions are from specified bank
	for _, txn := range response.Data {
		if txn.Bank != "HDFC" {
			t.Errorf("Filter failed: expected HDFC, got %s", txn.Bank)
		}
	}

	t.Logf("✓ Bank filter contract verified: All %d transactions are from HDFC", len(response.Data))
}

// TestMultiBankDateRangeFilter tests date range filtering across banks
// Contract: GET /api/transactions?date_from=2024-01-01&date_to=2024-06-30
// Expected: Only transactions within date range, regardless of bank
func TestMultiBankDateRangeFilter(t *testing.T) {
	router := api.SetupRouter()

	req, err := http.NewRequest("GET",
		"/api/transactions?date_from=2024-01-01&date_to=2024-06-30",
		nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
		return
	}

	var response struct {
		Data []struct {
			Date string `json:"date"`
		} `json:"data"`
	}

	json.Unmarshal(w.Body.Bytes(), &response)

	// Verify all dates are within range
	for _, txn := range response.Data {
		if txn.Date < "2024-01-01" || txn.Date > "2024-06-30" {
			t.Errorf("Date filter failed: %s outside range", txn.Date)
		}
	}

	t.Logf("✓ Date range filter contract verified: All %d transactions within date range", len(response.Data))
}

// TestMultiBankPagination tests pagination contract
// Contract: GET /api/transactions?limit=10&offset=0
// Expected: Array of max 10 items with pagination metadata
func TestMultiBankPagination(t *testing.T) {
	router := api.SetupRouter()

	req, err := http.NewRequest("GET",
		"/api/transactions?limit=10&offset=0",
		nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
		return
	}

	var response struct {
		Data       []interface{} `json:"data"`
		Pagination struct {
			Total  int `json:"total"`
			Limit  int `json:"limit"`
			Offset int `json:"offset"`
		} `json:"pagination"`
	}

	json.Unmarshal(w.Body.Bytes(), &response)

	// Verify pagination contract
	if response.Pagination.Limit != 10 {
		t.Errorf("Expected limit 10, got %d", response.Pagination.Limit)
	}

	if response.Pagination.Offset != 0 {
		t.Errorf("Expected offset 0, got %d", response.Pagination.Offset)
	}

	if len(response.Data) > 10 {
		t.Errorf("Returned more items than limit: %d > 10", len(response.Data))
	}

	t.Logf("✓ Pagination contract verified: limit=%d, offset=%d, returned=%d",
		response.Pagination.Limit, response.Pagination.Offset, len(response.Data))
}
