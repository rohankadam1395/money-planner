package formats

import (
	"money-planner/backend/internal/statement"
)

// AxisBankFormat defines column mappings for Axis Bank statements
// Axis statement format: Transaction Date, Reference No., Narration, Cheque No., Withdrawal, Deposit, Balance
type AxisBankFormat struct{}

// GetColumnMapping returns the column indices for Axis CSV format
func (f *AxisBankFormat) GetColumnMapping() statement.ColumnMapping {
	return statement.ColumnMapping{
		DateColumn:        0,  // Transaction Date
		MerchantColumn:    2,  // Narration
		AmountColumn:      4,  // Withdrawal or Deposit (combined)
		CreditColumn:      5,  // Deposit
		DebitColumn:       4,  // Withdrawal
		BalanceColumn:     6,  // Balance
		DescriptionColumn: 2,  // Narration (same as merchant)
	}
}

// GetBankCode returns the bank identifier
func (f *AxisBankFormat) GetBankCode() string {
	return "AXIS"
}

// GetBankName returns the bank full name
func (f *AxisBankFormat) GetBankName() string {
	return "Axis Bank"
}

// SkipRows returns number of header rows to skip
func (f *AxisBankFormat) SkipRows() int {
	return 1 // Axis has 1 header row
}

// ValidateFormat checks if the CSV appears to be Axis format
// Returns true if header contains expected Axis column names
func (f *AxisBankFormat) ValidateFormat(headers []string) bool {
	axisHeaders := map[string]bool{
		"Transaction Date": true,
		"Narration":        true,
		"Withdrawal":       true,
		"Deposit":          true,
		"Balance":          true,
	}

	for _, header := range headers {
		if axisHeaders[header] {
			return true
		}
	}
	return false
}
