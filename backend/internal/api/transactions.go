package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"money-planner/backend/internal/statement"
)

// TransactionResponse represents a single transaction in API responses
type TransactionResponse struct {
	ID          string    `json:"id"`
	Date        string    `json:"date"`
	Merchant    string    `json:"merchant"`
	Amount      float64   `json:"amount"`
	Type        string    `json:"type"`
	Bank        string    `json:"bank"`
	Currency    string    `json:"currency"`
	Balance     float64   `json:"balance"`
	Description string    `json:"description"`
	StatementID string    `json:"statement_id"`
	ImportTime  time.Time `json:"import_time"`
}

// PaginationMeta holds pagination information
type PaginationMeta struct {
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// TransactionsListResponse is the API response for listing transactions
type TransactionsListResponse struct {
	Data       []TransactionResponse `json:"data"`
	Pagination PaginationMeta        `json:"pagination"`
	Message    string                `json:"message,omitempty"`
}

// ListTransactions handles GET /api/transactions
// Query parameters:
// - bank_code: Filter by bank code (HDFC,ICICI,AXIS,SBI) - comma-separated for multiple
// - date_from: Filter from date (YYYY-MM-DD)
// - date_to: Filter to date (YYYY-MM-DD)
// - limit: Pagination limit (default 10, max 100)
// - offset: Pagination offset (default 0)
// - merchant: Search by merchant name
// - min_amount: Filter by minimum amount
// - max_amount: Filter by maximum amount
func ListTransactions(w http.ResponseWriter, r *http.Request, queryService *statement.QueryService) {
	// Extract query parameters
	bankCodesStr := r.URL.Query().Get("bank_code")
	dateFromStr := r.URL.Query().Get("date_from")
	dateToStr := r.URL.Query().Get("date_to")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	merchant := r.URL.Query().Get("merchant")
	minAmountStr := r.URL.Query().Get("min_amount")
	maxAmountStr := r.URL.Query().Get("max_amount")

	// Parse pagination parameters
	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	offset := 0
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Parse bank codes
	var bankCodes []string
	if bankCodesStr != "" {
		bankCodes = strings.Split(strings.ToUpper(bankCodesStr), ",")
		// Trim whitespace
		for i, code := range bankCodes {
			bankCodes[i] = strings.TrimSpace(code)
		}
	}

	// Parse date range
	dateFrom := time.Now().AddDate(-1, 0, 0) // Default: 1 year ago
	dateTo := time.Now()

	if dateFromStr != "" {
		if df, err := time.Parse("2006-01-02", dateFromStr); err == nil {
			dateFrom = df
		}
	}

	if dateToStr != "" {
		if dt, err := time.Parse("2006-01-02", dateToStr); err == nil {
			dateTo = dt
		}
	}

	// Parse amount filters
	var minAmount, maxAmount float64
	if minAmountStr != "" {
		if ma, err := strconv.ParseFloat(minAmountStr, 64); err == nil && ma > 0 {
			minAmount = ma
		}
	}
	if maxAmountStr != "" {
		if ma, err := strconv.ParseFloat(maxAmountStr, 64); err == nil && ma > 0 {
			maxAmount = ma
		}
	}

	// Build filter (would need user_id from auth context in real implementation)
	filter := statement.TransactionFilter{
		BankCodes:  bankCodes,
		DateFrom:   dateFrom,
		DateTo:     dateTo,
		Limit:      limit,
		Offset:     offset,
		SearchText: merchant,
		MinAmount:  minAmount,
		MaxAmount:  maxAmount,
		// UserID would come from request context (auth middleware)
	}

	// Query transactions
	result, err := queryService.ListTransactionsAcrossBanks(filter)
	if err != nil {
		http.Error(w, "Failed to query transactions", http.StatusInternalServerError)
		return
	}

	// Convert to response format
	transactions := make([]TransactionResponse, len(result.Transactions))
	for i, txn := range result.Transactions {
		balance := 0.0
		if txn.Balance != nil {
			balance = *txn.Balance
		}
		transactions[i] = TransactionResponse{
			ID:          txn.TransactionID,
			Date:        txn.TransactionDate.Format("2006-01-02"),
			Merchant:    txn.Merchant,
			Amount:      txn.Amount,
			Type:        txn.Type,
			Bank:        txn.BankCode,
			Currency:    txn.Currency,
			Balance:     balance,
			Description: txn.Description,
			StatementID: txn.StatementID,
			ImportTime:  txn.ImportedAt,
		}
	}

	// Build response
	response := TransactionsListResponse{
		Data: transactions,
		Pagination: PaginationMeta{
			Total:  result.Total,
			Limit:  result.Limit,
			Offset: result.Offset,
		},
	}

	if result.Total == 0 {
		response.Message = "No transactions found matching criteria"
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// SetupRouter creates a test router with transaction routes registered.
func SetupRouter() http.Handler {
	mux := http.NewServeMux()
	queryService := statement.NewQueryService(nil)
	RegisterTransactionRoutes(mux, queryService)
	return mux
}

// GetBankSummary handles GET /api/transactions/summary/by-bank
// Returns total amount per bank
func GetBankSummary(w http.ResponseWriter, r *http.Request, queryService *statement.QueryService) {
	// Would call queryService.GetBankSummary(userID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	// Would return: {"HDFC": 150000.00, "ICICI": 250000.00, ...}
}

// GetMerchantSummary handles GET /api/transactions/summary/by-merchant
// Returns total amount per merchant across all banks
func GetMerchantSummary(w http.ResponseWriter, r *http.Request, queryService *statement.QueryService) {
	// Would call queryService.GetMerchantSummary(userID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	// Would return: {"Whole Foods": 5000.00, "Amazon": 12000.00, ...}
}

// ExportTransactions handles GET /api/transactions/export
// Exports transactions as CSV
func ExportTransactions(w http.ResponseWriter, r *http.Request, queryService *statement.QueryService) {
	// Extract filters from query parameters (same as ListTransactions)
	filter := statement.TransactionFilter{
		Limit:  10000, // Large limit for export
		Offset: 0,
	}

	csv, err := queryService.ExportTransactionsAsCSV(filter)
	if err != nil {
		http.Error(w, "Failed to export transactions", http.StatusInternalServerError)
		return
	}

	// Send CSV file
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=transactions.csv")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(csv))
}

// RegisterTransactionRoutes registers all transaction-related routes
func RegisterTransactionRoutes(mux *http.ServeMux, queryService *statement.QueryService) {
	mux.HandleFunc("/api/transactions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			ListTransactions(w, r, queryService)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/transactions/summary/by-bank", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			GetBankSummary(w, r, queryService)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/transactions/summary/by-merchant", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			GetMerchantSummary(w, r, queryService)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/transactions/export", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			ExportTransactions(w, r, queryService)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
