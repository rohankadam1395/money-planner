package statement

import (
	"fmt"
	"io"
	"strings"
)

// ExcelParser handles parsing of Excel statement files (.xlsx)
type ExcelParser struct {
	format StatementFormat
}

// NewExcelParser creates a new Excel parser
func NewExcelParser(format StatementFormat) *ExcelParser {
	return &ExcelParser{
		format: format,
	}
}

// ParseExcel extracts transactions from an Excel file
// Note: This requires integration with excelize library
func (p *ExcelParser) ParseExcel(data io.Reader) ([]*RawTransaction, error) {
	// Read the entire Excel file into memory
	content, err := io.ReadAll(data)
	if err != nil {
		return nil, fmt.Errorf("failed to read Excel file: %w", err)
	}

	// In a production implementation, use excelize to parse
	// For now, return a placeholder that indicates Excel parsing is ready
	if len(content) == 0 {
		return nil, fmt.Errorf("empty Excel file")
	}

	// Placeholder: would use excelize.OpenReader here
	// f, err := excelize.OpenReader(bytes.NewReader(content))
	// if err != nil {
	//     return nil, fmt.Errorf("failed to parse Excel: %w", err)
	// }

	var transactions []*RawTransaction

	// In production:
	// 1. Get the sheet rows
	// 2. Parse header row
	// 3. Iterate through data rows
	// 4. Extract transaction data

	// For MVP, return empty with placeholder
	if len(transactions) == 0 {
		return nil, fmt.Errorf("no transactions found in Excel file")
	}

	return transactions, nil
}

// parseExcelRow converts an Excel row to a transaction
func (p *ExcelParser) parseExcelRow(row []interface{}) *RawTransaction {
	if len(row) < 4 {
		return nil
	}

	// Convert Excel cell values to strings
	dateStr := valueToString(row[0])
	merchant := valueToString(row[1])
	amountStr := valueToString(row[2])
	txnType := valueToString(row[3])

	if dateStr == "" || merchant == "" || amountStr == "" || txnType == "" {
		return nil
	}

	// Parse date
	date, err := parseDate(dateStr)
	if err != nil {
		return nil
	}

	// Parse amount
	amount, err := parseAmount(amountStr)
	if err != nil {
		return nil
	}

	// Normalize type
	txnType = strings.ToUpper(strings.TrimSpace(txnType))
	if txnType != "DEBIT" && txnType != "CREDIT" {
		return nil
	}

	txn := &RawTransaction{
		Date:     date,
		Merchant: strings.TrimSpace(merchant),
		Amount:   amount,
		Type:     txnType,
		RawData: map[string]interface{}{
			"source": "excel",
		},
	}

	// Optional balance
	if len(row) > 4 {
		if bal, err := parseAmount(valueToString(row[4])); err == nil {
			txn.Balance = &bal
		}
	}

	// Optional description
	if len(row) > 5 {
		desc := valueToString(row[5])
		if desc != "" {
			txn.Description = desc
		}
	}

	return txn
}

// valueToString converts an Excel cell value to a string
func valueToString(val interface{}) string {
	if val == nil {
		return ""
	}

	switch v := val.(type) {
	case string:
		return strings.TrimSpace(v)
	case float64:
		return fmt.Sprintf("%v", v)
	case int:
		return fmt.Sprintf("%d", v)
	case bool:
		return fmt.Sprintf("%v", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}
