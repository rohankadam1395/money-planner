package formats

import (
	"money-planner/backend/internal/statement"
)

// SBIBankFormat defines column mappings for State Bank of India (SBI) statements
// SBI statement format: Transaction Date, Value Date, Reference No., Narration, Withdrawal, Deposit, Closing Balance
type SBIBankFormat struct{}

// GetColumnMapping returns the column indices for SBI CSV format
func (f *SBIBankFormat) GetColumnMapping() statement.ColumnMapping {
	return statement.ColumnMapping{
		DateColumn:        0,  // Transaction Date
		MerchantColumn:    3,  // Narration
		AmountColumn:      4,  // Withdrawal or Deposit
		CreditColumn:      5,  // Deposit
		DebitColumn:       4,  // Withdrawal
		BalanceColumn:     6,  // Closing Balance
		DescriptionColumn: 3,  // Narration (same as merchant)
	}
}

// GetBankCode returns the bank identifier
func (f *SBIBankFormat) GetBankCode() string {
	return "SBI"
}

// GetBankName returns the bank full name
func (f *SBIBankFormat) GetBankName() string {
	return "State Bank of India"
}

// SkipRows returns number of header rows to skip
func (f *SBIBankFormat) SkipRows() int {
	return 1 // SBI has 1 header row
}

// ValidateFormat checks if the CSV appears to be SBI format
// Returns true if header contains expected SBI column names
func (f *SBIBankFormat) ValidateFormat(headers []string) bool {
	sbiHeaders := map[string]bool{
		"Transaction Date": true,
		"Value Date":       true,
		"Narration":        true,
		"Withdrawal":       true,
		"Deposit":          true,
		"Closing Balance":  true,
	}

	for _, header := range headers {
		if sbiHeaders[header] {
			return true
		}
	}
	return false
}
