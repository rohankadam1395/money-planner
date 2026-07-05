package statement

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	MaxMerchantLength   = 256
	MaxDescriptionLength = 512
	MaxAccountHashLen   = 64
)

type ValidationError struct {
	Field   string
	Message string
}

type TransactionValidationResult struct {
	Valid  bool
	Errors []ValidationError
}

func ValidateTransaction(t *Transaction, periodStart, periodEnd time.Time) *TransactionValidationResult {
	result := &TransactionValidationResult{
		Valid:  true,
		Errors: []ValidationError{},
	}

	if err := ValidateTransactionDate(t.TransactionDate, periodStart, periodEnd); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "transaction_date",
			Message: err.Error(),
		})
	}

	if err := ValidateMerchant(t.Merchant); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "merchant",
			Message: err.Error(),
		})
	}

	if err := ValidateAmount(t.Amount); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "amount",
			Message: err.Error(),
		})
	}

	if err := ValidateType(t.Type); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "type",
			Message: err.Error(),
		})
	}

	if err := ValidateCurrency(t.Currency); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "currency",
			Message: err.Error(),
		})
	}

	if err := ValidateAccountHash(t.AccountNumberHash); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "account_number_hash",
			Message: err.Error(),
		})
	}

	return result
}

func ValidateTransactionDate(date time.Time, periodStart, periodEnd time.Time) error {
	if date.Before(periodStart) || date.After(periodEnd) {
		return fmt.Errorf("transaction date %s is outside statement period [%s, %s]",
			date.Format("2006-01-02"), periodStart.Format("2006-01-02"), periodEnd.Format("2006-01-02"))
	}
	return nil
}

func ValidateMerchant(merchant string) error {
	merchant = strings.TrimSpace(merchant)
	if merchant == "" {
		return fmt.Errorf("merchant cannot be empty")
	}

	if len(merchant) > MaxMerchantLength {
		return fmt.Errorf("merchant exceeds maximum length of %d characters", MaxMerchantLength)
	}

	if strings.Contains(merchant, "\x00") {
		return fmt.Errorf("merchant contains null bytes")
	}

	if isHTMLOrScript(merchant) {
		return fmt.Errorf("merchant contains HTML or script tags")
	}

	return nil
}

func ValidateAmount(amount float64) error {
	if amount <= 0 {
		return fmt.Errorf("amount must be positive, got %f", amount)
	}

	// Check for valid decimal representation (max 2 decimal places)
	amountStr := fmt.Sprintf("%.2f", amount)
	parsed, _ := strconv.ParseFloat(amountStr, 64)
	if parsed != amount {
		return fmt.Errorf("amount must have at most 2 decimal places")
	}

	return nil
}

func ValidateType(txnType string) error {
	txnType = strings.ToUpper(strings.TrimSpace(txnType))
	if txnType != "DEBIT" && txnType != "CREDIT" {
		return fmt.Errorf("type must be DEBIT or CREDIT, got %s", txnType)
	}
	return nil
}

func ValidateCurrency(currency string) error {
	currency = strings.ToUpper(strings.TrimSpace(currency))
	if len(currency) != 3 {
		return fmt.Errorf("currency must be a 3-character ISO code, got %s", currency)
	}

	// Basic check for alphabetic characters
	if !regexp.MustCompile(`^[A-Z]{3}$`).MatchString(currency) {
		return fmt.Errorf("currency must contain only letters, got %s", currency)
	}

	return nil
}

func ValidateAccountHash(hash string) error {
	if len(hash) != MaxAccountHashLen {
		return fmt.Errorf("account number hash must be %d characters (SHA-256 hex), got %d", MaxAccountHashLen, len(hash))
	}

	if !regexp.MustCompile(`^[a-f0-9]{64}$`).MatchString(strings.ToLower(hash)) {
		return fmt.Errorf("account number hash must be valid hex string")
	}

	return nil
}

func isHTMLOrScript(s string) bool {
	s = strings.ToLower(s)
	htmlPatterns := []string{"<script", "<iframe", "<object", "<embed", "onclick", "onerror", "javascript:"}
	for _, pattern := range htmlPatterns {
		if strings.Contains(s, pattern) {
			return true
		}
	}
	return false
}

type Transaction struct {
	TransactionID     string
	UserID            string
	StatementID       string
	TransactionDate   time.Time
	Merchant          string
	Amount            float64
	Type              string // DEBIT or CREDIT
	Balance           *float64
	Description       string
	Currency          string
	BankCode          string
	AccountNumberHash string
	RawData           map[string]interface{}
	ImportedAt        time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// RawTransaction represents a transaction extracted from a statement file before validation
type RawTransaction struct {
	Date        time.Time
	Merchant    string
	Amount      float64
	Type        string // DEBIT or CREDIT
	Balance     *float64
	Description string
	Currency    string
	RawData     map[string]interface{}
}
