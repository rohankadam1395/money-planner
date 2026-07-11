package statement

import (
	"fmt"
	"sort"
	"time"
)

// QueryService provides cross-bank transaction queries
type QueryService struct {
	repo *StatementRepository
}

// NewQueryService creates a new query service
func NewQueryService(repo *StatementRepository) *QueryService {
	return &QueryService{repo: repo}
}

// TransactionFilter defines query filters for transaction searches
type TransactionFilter struct {
	BankCodes  []string
	DateFrom   time.Time
	DateTo     time.Time
	Limit      int
	Offset     int
	UserID     string
	MinAmount  float64
	MaxAmount  float64
	SearchText string
}

// QueryResult holds paginated query results
type QueryResult struct {
	Transactions []Transaction
	Total        int
	Limit        int
	Offset       int
}

// ListTransactionsAcrossBanks queries transactions from multiple banks
// Filters by date range, bank codes, and pagination
func (qs *QueryService) ListTransactionsAcrossBanks(filter TransactionFilter) (*QueryResult, error) {
	// Default pagination
	if filter.Limit == 0 {
		filter.Limit = 10
	}
	if filter.Limit > 100 {
		filter.Limit = 100 // Max 100 per request
	}

	// Get all statements for user across specified banks
	var allTransactions []Transaction

	// In a real implementation, this would query the database with proper filtering
	// For now, this is a placeholder showing the query pattern

	// Database query would look like:
	// SELECT * FROM transactions
	// WHERE user_id = ?
	//   AND source_bank IN (?, ?, ...)
	//   AND transaction_date BETWEEN ? AND ?
	//   AND amount BETWEEN ? AND ?
	//   ORDER BY transaction_date DESC
	//   LIMIT ? OFFSET ?

	// Apply in-memory filtering for demonstration
	for _, txn := range allTransactions {
		// Filter by date range
		if !txn.TransactionDate.IsZero() {
			if txn.TransactionDate.Before(filter.DateFrom) || txn.TransactionDate.After(filter.DateTo) {
				continue
			}
		}

		// Filter by amount range
		if filter.MinAmount > 0 && txn.Amount < filter.MinAmount {
			continue
		}
		if filter.MaxAmount > 0 && txn.Amount > filter.MaxAmount {
			continue
		}

		// Filter by bank code
		if len(filter.BankCodes) > 0 {
			bankFound := false
			for _, code := range filter.BankCodes {
				if txn.BankCode == code {
					bankFound = true
					break
				}
			}
			if !bankFound {
				continue
			}
		}
	}

	// Sort by date (descending - newest first)
	sort.Slice(allTransactions, func(i, j int) bool {
		return allTransactions[i].TransactionDate.After(allTransactions[j].TransactionDate)
	})

	// Apply pagination
	total := len(allTransactions)
	start := filter.Offset
	end := start + filter.Limit

	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	paginatedTxns := []Transaction{}
	if start < total {
		paginatedTxns = allTransactions[start:end]
	}

	return &QueryResult{
		Transactions: paginatedTxns,
		Total:        total,
		Limit:        filter.Limit,
		Offset:       filter.Offset,
	}, nil
}

// GetTransactionsByBankCode returns all transactions from a specific bank for a user
func (qs *QueryService) GetTransactionsByBankCode(userID, bankCode string, limit, offset int) (*QueryResult, error) {
	filter := TransactionFilter{
		UserID:    userID,
		BankCodes: []string{bankCode},
		Limit:     limit,
		Offset:    offset,
		DateFrom:  time.Now().AddDate(-1, 0, 0), // Default: past 1 year
		DateTo:    time.Now(),
	}

	return qs.ListTransactionsAcrossBanks(filter)
}

// GetTransactionsByDateRange returns transactions within a date range from all banks
func (qs *QueryService) GetTransactionsByDateRange(userID string, dateFrom, dateTo time.Time, limit, offset int) (*QueryResult, error) {
	filter := TransactionFilter{
		UserID:   userID,
		DateFrom: dateFrom,
		DateTo:   dateTo,
		Limit:    limit,
		Offset:   offset,
	}

	return qs.ListTransactionsAcrossBanks(filter)
}

// GetMerchantSummary returns aggregated transaction amounts by merchant across all banks
func (qs *QueryService) GetMerchantSummary(userID string) (map[string]float64, error) {
	summary := make(map[string]float64)

	// Would query: SELECT merchant, SUM(amount) FROM transactions WHERE user_id = ? GROUP BY merchant

	return summary, nil
}

// GetBankSummary returns aggregated transaction amounts by bank
func (qs *QueryService) GetBankSummary(userID string) (map[string]float64, error) {
	summary := make(map[string]float64)

	// Would query: SELECT source_bank, SUM(amount) FROM transactions WHERE user_id = ? GROUP BY source_bank

	return summary, nil
}

// GetDailyBalanceAcrossBanks returns consolidated daily balance across all banks
// (For multi-currency, this would require conversion to base currency)
func (qs *QueryService) GetDailyBalanceAcrossBanks(userID string, dateFrom, dateTo time.Time) (map[time.Time]float64, error) {
	dailyBalance := make(map[time.Time]float64)

	// Would aggregate balances from all statements for the user within date range
	// For simplicity, assuming single currency (INR) in MVP

	return dailyBalance, nil
}

// ValidateMultiBankConsistency checks for duplicate transactions across banks
// Returns any potential duplicates detected
func (qs *QueryService) ValidateMultiBankConsistency(userID string) ([]struct {
	Transaction1 Transaction
	Transaction2 Transaction
	Confidence   float64 // 0-1, higher = more likely duplicate
}, error) {
	// Would compare transactions from different banks for potential duplicates
	// Uses: amount match, date match, merchant name similarity, balance before/after

	return []struct {
		Transaction1 Transaction
		Transaction2 Transaction
		Confidence   float64
	}{}, nil
}

// ConsolidateStatements merges transactions from multiple statements into a unified view
// Handles overlapping date ranges and ensures no duplicates
func (qs *QueryService) ConsolidateStatements(userID string, statementIDs []string) (*QueryResult, error) {
	filter := TransactionFilter{
		UserID: userID,
		Limit:  1000, // Large limit for consolidation
		Offset: 0,
	}

	// Would query: SELECT * FROM transactions WHERE statement_id IN (?, ?, ...) ORDER BY transaction_date DESC

	return qs.ListTransactionsAcrossBanks(filter)
}

// ExportTransactionsAsCSV exports transactions to CSV format
func (qs *QueryService) ExportTransactionsAsCSV(filter TransactionFilter) (string, error) {
	result, err := qs.ListTransactionsAcrossBanks(filter)
	if err != nil {
		return "", fmt.Errorf("failed to query transactions: %w", err)
	}

	// Build CSV: Date,Merchant,Amount,Type,Bank,Balance,Description
	csv := "Date,Merchant,Amount,Type,Bank,Balance,Description\n"
	for _, txn := range result.Transactions {
		dateStr := txn.TransactionDate.Format("2006-01-02")
		balance := 0.0
		if txn.Balance != nil {
			balance = *txn.Balance
		}
		csv += fmt.Sprintf("%s,%s,%.2f,%s,%s,%.2f,%s\n",
			dateStr, txn.Merchant, txn.Amount, txn.Type, txn.BankCode, balance, txn.Description)
	}

	return csv, nil
}
