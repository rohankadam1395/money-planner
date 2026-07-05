package unit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"money-planner/backend/internal/statement"
)

func TestValidateTransaction(t *testing.T) {
	periodStart := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	periodEnd := time.Date(2026, 1, 31, 23, 59, 59, 0, time.UTC)

	tests := []struct {
		name       string
		txn        *statement.Transaction
		periodStart time.Time
		periodEnd   time.Time
		wantValid  bool
		wantErrors int
	}{
		{
			name: "valid transaction passes validation",
			txn: &statement.Transaction{
				TransactionID:     "550e8400-e29b-41d4-a716-446655440000",
				UserID:            "550e8400-e29b-41d4-a716-446655440001",
				StatementID:       "550e8400-e29b-41d4-a716-446655440002",
				TransactionDate:   time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
				Merchant:          "Test Store",
				Amount:            100.50,
				Type:              "DEBIT",
				Currency:          "INR",
				BankCode:          "HDFC",
				AccountNumberHash: "a" + "0123456789abcdef0123456789abcdef0123456789abcdef012345678",
			},
			periodStart: periodStart,
			periodEnd:   periodEnd,
			wantValid:   true,
			wantErrors:  0,
		},
		{
			name: "transaction date before period returns error",
			txn: &statement.Transaction{
				TransactionID:     "550e8400-e29b-41d4-a716-446655440000",
				UserID:            "550e8400-e29b-41d4-a716-446655440001",
				StatementID:       "550e8400-e29b-41d4-a716-446655440002",
				TransactionDate:   time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC),
				Merchant:          "Test Store",
				Amount:            100.50,
				Type:              "DEBIT",
				Currency:          "INR",
				BankCode:          "HDFC",
				AccountNumberHash: "a" + "0123456789abcdef0123456789abcdef0123456789abcdef012345678",
			},
			periodStart: periodStart,
			periodEnd:   periodEnd,
			wantValid:   false,
			wantErrors:  1,
		},
		{
			name: "transaction date after period returns error",
			txn: &statement.Transaction{
				TransactionID:     "550e8400-e29b-41d4-a716-446655440000",
				UserID:            "550e8400-e29b-41d4-a716-446655440001",
				StatementID:       "550e8400-e29b-41d4-a716-446655440002",
				TransactionDate:   time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
				Merchant:          "Test Store",
				Amount:            100.50,
				Type:              "DEBIT",
				Currency:          "INR",
				BankCode:          "HDFC",
				AccountNumberHash: "a" + "0123456789abcdef0123456789abcdef0123456789abcdef012345678",
			},
			periodStart: periodStart,
			periodEnd:   periodEnd,
			wantValid:   false,
			wantErrors:  1,
		},
		{
			name: "empty merchant returns error",
			txn: &statement.Transaction{
				TransactionID:     "550e8400-e29b-41d4-a716-446655440000",
				UserID:            "550e8400-e29b-41d4-a716-446655440001",
				StatementID:       "550e8400-e29b-41d4-a716-446655440002",
				TransactionDate:   time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
				Merchant:          "",
				Amount:            100.50,
				Type:              "DEBIT",
				Currency:          "INR",
				BankCode:          "HDFC",
				AccountNumberHash: "a" + "0123456789abcdef0123456789abcdef0123456789abcdef012345678",
			},
			periodStart: periodStart,
			periodEnd:   periodEnd,
			wantValid:   false,
			wantErrors:  1,
		},
		{
			name: "negative amount returns error",
			txn: &statement.Transaction{
				TransactionID:     "550e8400-e29b-41d4-a716-446655440000",
				UserID:            "550e8400-e29b-41d4-a716-446655440001",
				StatementID:       "550e8400-e29b-41d4-a716-446655440002",
				TransactionDate:   time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
				Merchant:          "Test Store",
				Amount:            -100.50,
				Type:              "DEBIT",
				Currency:          "INR",
				BankCode:          "HDFC",
				AccountNumberHash: "a" + "0123456789abcdef0123456789abcdef0123456789abcdef012345678",
			},
			periodStart: periodStart,
			periodEnd:   periodEnd,
			wantValid:   false,
			wantErrors:  1,
		},
		{
			name: "zero amount returns error",
			txn: &statement.Transaction{
				TransactionID:     "550e8400-e29b-41d4-a716-446655440000",
				UserID:            "550e8400-e29b-41d4-a716-446655440001",
				StatementID:       "550e8400-e29b-41d4-a716-446655440002",
				TransactionDate:   time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
				Merchant:          "Test Store",
				Amount:            0,
				Type:              "DEBIT",
				Currency:          "INR",
				BankCode:          "HDFC",
				AccountNumberHash: "a" + "0123456789abcdef0123456789abcdef0123456789abcdef012345678",
			},
			periodStart: periodStart,
			periodEnd:   periodEnd,
			wantValid:   false,
			wantErrors:  1,
		},
		{
			name: "invalid transaction type returns error",
			txn: &statement.Transaction{
				TransactionID:     "550e8400-e29b-41d4-a716-446655440000",
				UserID:            "550e8400-e29b-41d4-a716-446655440001",
				StatementID:       "550e8400-e29b-41d4-a716-446655440002",
				TransactionDate:   time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
				Merchant:          "Test Store",
				Amount:            100.50,
				Type:              "INVALID",
				Currency:          "INR",
				BankCode:          "HDFC",
				AccountNumberHash: "a" + "0123456789abcdef0123456789abcdef0123456789abcdef012345678",
			},
			periodStart: periodStart,
			periodEnd:   periodEnd,
			wantValid:   false,
			wantErrors:  1,
		},
		{
			name: "invalid currency code returns error",
			txn: &statement.Transaction{
				TransactionID:     "550e8400-e29b-41d4-a716-446655440000",
				UserID:            "550e8400-e29b-41d4-a716-446655440001",
				StatementID:       "550e8400-e29b-41d4-a716-446655440002",
				TransactionDate:   time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
				Merchant:          "Test Store",
				Amount:            100.50,
				Type:              "DEBIT",
				Currency:          "INVALID",
				BankCode:          "HDFC",
				AccountNumberHash: "a" + "0123456789abcdef0123456789abcdef0123456789abcdef012345678",
			},
			periodStart: periodStart,
			periodEnd:   periodEnd,
			wantValid:   false,
			wantErrors:  1,
		},
		{
			name: "invalid account hash returns error",
			txn: &statement.Transaction{
				TransactionID:     "550e8400-e29b-41d4-a716-446655440000",
				UserID:            "550e8400-e29b-41d4-a716-446655440001",
				StatementID:       "550e8400-e29b-41d4-a716-446655440002",
				TransactionDate:   time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
				Merchant:          "Test Store",
				Amount:            100.50,
				Type:              "DEBIT",
				Currency:          "INR",
				BankCode:          "HDFC",
				AccountNumberHash: "invalid-hash",
			},
			periodStart: periodStart,
			periodEnd:   periodEnd,
			wantValid:   false,
			wantErrors:  1,
		},
		{
			name: "multiple validation errors",
			txn: &statement.Transaction{
				TransactionID:     "550e8400-e29b-41d4-a716-446655440000",
				UserID:            "550e8400-e29b-41d4-a716-446655440001",
				StatementID:       "550e8400-e29b-41d4-a716-446655440002",
				TransactionDate:   time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC),
				Merchant:          "",
				Amount:            -100.50,
				Type:              "INVALID",
				Currency:          "INVALID",
				BankCode:          "HDFC",
				AccountNumberHash: "invalid-hash",
			},
			periodStart: periodStart,
			periodEnd:   periodEnd,
			wantValid:   false,
			wantErrors:  5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := statement.ValidateTransaction(tt.txn, tt.periodStart, tt.periodEnd)

			assert.Equal(t, tt.wantValid, result.Valid, "validation result mismatch")
			assert.Equal(t, tt.wantErrors, len(result.Errors), "error count mismatch")

			if !result.Valid {
				for _, err := range result.Errors {
					assert.NotEmpty(t, err.Field, "error should have field")
					assert.NotEmpty(t, err.Message, "error should have message")
				}
			}
		})
	}
}

func TestValidateMerchant(t *testing.T) {
	tests := []struct {
		name     string
		merchant string
		wantErr  bool
	}{
		{"valid merchant", "Test Store", false},
		{"merchant with spaces", "  Test Store  ", false},
		{"empty merchant", "", true},
		{"merchant with HTML tag", "Store <script>", true},
		{"merchant with onclick", "Store onclick=alert()", true},
		{"very long merchant", string(make([]byte, 300)), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := statement.ValidateMerchant(tt.merchant)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateAmount(t *testing.T) {
	tests := []struct {
		name    string
		amount  float64
		wantErr bool
	}{
		{"valid positive amount", 100.50, false},
		{"valid integer amount", 100.0, false},
		{"zero amount", 0, true},
		{"negative amount", -100.50, true},
		{"very small amount", 0.01, false},
		{"large amount", 999999.99, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := statement.ValidateAmount(tt.amount)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateType(t *testing.T) {
	tests := []struct {
		name    string
		txnType string
		wantErr bool
	}{
		{"debit", "DEBIT", false},
		{"credit", "CREDIT", false},
		{"lowercase debit", "debit", false},
		{"lowercase credit", "credit", false},
		{"invalid type", "INVALID", true},
		{"empty type", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := statement.ValidateType(tt.txnType)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateCurrency(t *testing.T) {
	tests := []struct {
		name     string
		currency string
		wantErr  bool
	}{
		{"INR", "INR", false},
		{"USD", "USD", false},
		{"EUR", "EUR", false},
		{"lowercase inr", "inr", false},
		{"too short", "IN", true},
		{"too long", "INRR", true},
		{"with numbers", "1NR", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := statement.ValidateCurrency(tt.currency)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
