package formats

import (
	"fmt"
	"money-planner/backend/internal/statement"
)

// BankFormat interface defines the contract for bank-specific format handlers
type BankFormat interface {
	GetColumnMapping() statement.ColumnMapping
	GetBankCode() string
	GetBankName() string
	SkipRows() int
	ValidateFormat(headers []string) bool
}

// BankFormatRegistry manages registered bank formats and provides auto-detection
type BankFormatRegistry struct {
	formats map[string]BankFormat
}

// NewBankFormatRegistry creates a new format registry with all supported banks
func NewBankFormatRegistry() *BankFormatRegistry {
	registry := &BankFormatRegistry{
		formats: make(map[string]BankFormat),
	}

	// Register all supported bank formats
	registry.Register("HDFC", &HDFCBankFormat{})
	registry.Register("ICICI", &ICICIBankFormat{})
	registry.Register("AXIS", &AxisBankFormat{})
	registry.Register("SBI", &SBIBankFormat{})

	return registry
}

// Register adds a new bank format to the registry
func (r *BankFormatRegistry) Register(bankCode string, format BankFormat) {
	r.formats[bankCode] = format
}

// GetFormat returns the format for a specific bank code
func (r *BankFormatRegistry) GetFormat(bankCode string) (BankFormat, error) {
	format, exists := r.formats[bankCode]
	if !exists {
		return nil, fmt.Errorf("unsupported bank code: %s", bankCode)
	}
	return format, nil
}

// DetectBankFormat attempts to detect the bank format from CSV headers
// Returns the detected BankFormat and bank code
func (r *BankFormatRegistry) DetectBankFormat(headers []string) (BankFormat, string, error) {
	for bankCode, format := range r.formats {
		if format.ValidateFormat(headers) {
			return format, bankCode, nil
		}
	}
	return nil, "", fmt.Errorf("unable to detect bank format from headers")
}

// ListSupportedBanks returns a slice of all supported bank codes
func (r *BankFormatRegistry) ListSupportedBanks() []string {
	codes := make([]string, 0, len(r.formats))
	for code := range r.formats {
		codes = append(codes, code)
	}
	return codes
}

// AutoDetectAndParse attempts to auto-detect the bank format and parse the CSV
// If bankCode is provided, uses that format; otherwise attempts auto-detection
func (r *BankFormatRegistry) AutoDetectAndParse(bankCode string, headers []string) (BankFormat, error) {
	if bankCode != "" {
		// Use provided bank code
		format, err := r.GetFormat(bankCode)
		if err != nil {
			return nil, err
		}
		return format, nil
	}

	// Attempt auto-detection
	format, _, err := r.DetectBankFormat(headers)
	if err != nil {
		return nil, err
	}
	return format, nil
}

// Placeholder format implementations for HDFC and ICICI
// (These are already implemented in hdfc.go and icici.go)

// HDFCBankFormat is a placeholder (full implementation in hdfc.go)
type HDFCBankFormat struct{}

func (f *HDFCBankFormat) GetColumnMapping() statement.ColumnMapping {
	return statement.ColumnMapping{
		DateColumn:        0,
		MerchantColumn:    1,
		AmountColumn:      2,
		DebitColumn:       2,
		CreditColumn:      3,
		BalanceColumn:     4,
		DescriptionColumn: 1,
	}
}

func (f *HDFCBankFormat) GetBankCode() string   { return "HDFC" }
func (f *HDFCBankFormat) GetBankName() string   { return "HDFC Bank" }
func (f *HDFCBankFormat) SkipRows() int         { return 1 }
func (f *HDFCBankFormat) ValidateFormat(headers []string) bool {
	for _, h := range headers {
		if h == "Credit" || h == "Debit" {
			return true
		}
	}
	return false
}

// ICICIBankFormat is a placeholder (full implementation in icici.go)
type ICICIBankFormat struct{}

func (f *ICICIBankFormat) GetColumnMapping() statement.ColumnMapping {
	return statement.ColumnMapping{
		DateColumn:        0,
		MerchantColumn:    1,
		AmountColumn:      2,
		DebitColumn:       2,
		CreditColumn:      3,
		BalanceColumn:     4,
		DescriptionColumn: 1,
	}
}

func (f *ICICIBankFormat) GetBankCode() string   { return "ICICI" }
func (f *ICICIBankFormat) GetBankName() string   { return "ICICI Bank" }
func (f *ICICIBankFormat) SkipRows() int         { return 1 }
func (f *ICICIBankFormat) ValidateFormat(headers []string) bool {
	for _, h := range headers {
		if h == "Amount" || h == "Description" {
			return true
		}
	}
	return false
}
