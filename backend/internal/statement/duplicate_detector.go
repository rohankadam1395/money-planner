package statement

import (
	"crypto/sha256"
	"fmt"
	"time"
)

// DuplicateDetector handles duplicate statement and transaction detection
type DuplicateDetector struct {
	repo *StatementRepository
}

// NewDuplicateDetector creates a new duplicate detector
func NewDuplicateDetector(repo *StatementRepository) *DuplicateDetector {
	return &DuplicateDetector{repo: repo}
}

// DuplicateCheckResult holds the result of duplicate detection
type DuplicateCheckResult struct {
	IsDuplicate        bool
	ExistingStatement  *Statement
	SimilarityScore    float64 // 0-1, higher = more similar
	DuplicateType      string  // "EXACT_FILE", "DATE_RANGE", "TRANSACTION_OVERLAP", "NONE"
	WarningMessage     string
}

// CheckForDuplicateFile checks if a statement file hash already exists
func (dd *DuplicateDetector) CheckForDuplicateFile(userID, bankCode, fileHash string) *DuplicateCheckResult {
	// Query: SELECT * FROM statements WHERE user_id = ? AND bank_code = ? AND file_hash = ?
	// In real implementation, would query database

	result := &DuplicateCheckResult{
		IsDuplicate:   false,
		DuplicateType: "NONE",
	}

	// Placeholder logic - would check DB
	return result
}

// CheckForDateRangeOverlap checks if a statement date range overlaps with existing statements from same bank
func (dd *DuplicateDetector) CheckForDateRangeOverlap(userID, bankCode string, periodStart, periodEnd time.Time) *DuplicateCheckResult {
	// Query: SELECT * FROM statements
	// WHERE user_id = ? AND bank_code = ?
	//   AND (period_start <= ? AND period_end >= ?)
	//   ORDER BY period_start DESC LIMIT 1

	result := &DuplicateCheckResult{
		IsDuplicate:   false,
		DuplicateType: "NONE",
	}

	// Placeholder logic - would check DB for overlapping date ranges
	// In real implementation:
	// for _, stmt := range existingStatements {
	//   if dd.hasDateOverlap(stmt.PeriodStart, stmt.PeriodEnd, periodStart, periodEnd) {
	//     result.IsDuplicate = true
	//     result.DuplicateType = "DATE_RANGE"
	//     result.ExistingStatement = stmt
	//     result.WarningMessage = fmt.Sprintf("Statement for %s overlaps with existing statement (%s to %s)",
	//       bankCode, stmt.PeriodStart.Format("2006-01-02"), stmt.PeriodEnd.Format("2006-01-02"))
	//   }
	// }

	return result
}

// hasDateOverlap checks if two date ranges overlap
func (dd *DuplicateDetector) hasDateOverlap(s1Start, s1End, s2Start, s2End time.Time) bool {
	return !(s1End.Before(s2Start) || s2End.Before(s1Start))
}

// CheckForTransactionOverlap checks for transaction-level duplicates
// Returns transactions that appear in both statements
func (dd *DuplicateDetector) CheckForTransactionOverlap(userID string, newTxns []Transaction, bankCode string) ([]Transaction, error) {
	// Query: SELECT * FROM transactions
	// WHERE user_id = ? AND source_bank = ?
	//   AND (
	//     (transaction_date, amount, merchant) IN (?)  // Exact match
	//     OR (transaction_date BETWEEN ? AND ? AND amount = ?)  // Same day, same amount
	//   )
	// ORDER BY transaction_date DESC

	duplicates := []Transaction{}

	// Placeholder logic - would check DB for duplicate transactions
	// In real implementation:
	// for _, newTxn := range newTxns {
	//   for _, existingTxn := range existingTxns {
	//     if dd.isTransactionDuplicate(newTxn, existingTxn) {
	//       duplicates = append(duplicates, newTxn)
	//     }
	//   }
	// }

	return duplicates, nil
}

// isTransactionDuplicate checks if two transactions are duplicates
// Uses: amount match + date match + merchant similarity
func (dd *DuplicateDetector) isTransactionDuplicate(txn1, txn2 Transaction) bool {
	// Same amount + same date + similar merchant = likely duplicate
	if txn1.Amount == txn2.Amount && txn1.TransactionDate.Equal(txn2.TransactionDate) {
		// Check merchant similarity (can use Levenshtein distance for fuzzy match)
		return txn1.Merchant == txn2.Merchant // Simplified
	}
	return false
}

// ComputeFileHash computes SHA-256 hash of file contents
func (dd *DuplicateDetector) ComputeFileHash(fileContent []byte) string {
	hash := sha256.Sum256(fileContent)
	return fmt.Sprintf("%x", hash)
}

// CheckFullDuplicate performs comprehensive duplicate detection
func (dd *DuplicateDetector) CheckFullDuplicate(userID, bankCode string, fileHash string, periodStart, periodEnd time.Time) *DuplicateCheckResult {
	// First: Check exact file duplicate
	if result := dd.CheckForDuplicateFile(userID, bankCode, fileHash); result.IsDuplicate {
		return result
	}

	// Second: Check date range overlap
	if result := dd.CheckForDateRangeOverlap(userID, bankCode, periodStart, periodEnd); result.IsDuplicate {
		result.WarningMessage = fmt.Sprintf(
			"Warning: This statement overlaps with a previously imported %s statement (%s to %s). "+
				"Duplicate transactions will be detected during import.",
			bankCode,
			periodStart.Format("2006-01-02"),
			periodEnd.Format("2006-01-02"),
		)
		return result
	}

	// No duplicates found
	return &DuplicateCheckResult{
		IsDuplicate:   false,
		DuplicateType: "NONE",
	}
}

// DetectOverlappingStatements finds all statements that overlap with a given date range
func (dd *DuplicateDetector) DetectOverlappingStatements(userID, bankCode string, periodStart, periodEnd time.Time) ([]*Statement, error) {
	// Query: SELECT * FROM statements
	// WHERE user_id = ? AND bank_code = ?
	//   AND (period_start <= ? AND period_end >= ?)
	//   ORDER BY period_start DESC

	overlapping := []*Statement{}

	// Placeholder logic
	// In real implementation:
	// for _, stmt := range statements {
	//   if dd.hasDateOverlap(stmt.PeriodStart, stmt.PeriodEnd, periodStart, periodEnd) {
	//     overlapping = append(overlapping, stmt)
	//   }
	// }

	return overlapping, nil
}

// MergeDuplicateTransactions merges duplicate transactions from overlapping statements
// Keeps the newer one (most recent import)
func (dd *DuplicateDetector) MergeDuplicateTransactions(txn1, txn2 Transaction) Transaction {
	// Keep the one with more recent import timestamp
	if txn1.ImportedAt.After(txn2.ImportedAt) {
		return txn1
	}
	return txn2
}
