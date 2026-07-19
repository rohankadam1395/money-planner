package statement

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"time"
)

// CSVParser handles parsing of standard bank CSV statement files
type CSVParser struct {
	columnMapping map[string]int // Maps column names to indices
}

// NewCSVParser creates a new CSV parser
func NewCSVParser() *CSVParser {
	return &CSVParser{
		columnMapping: make(map[string]int),
	}
}

// ParseCSV parses a CSV file and extracts transactions
func (p *CSVParser) ParseCSV(data io.Reader) ([]*RawTransaction, error) {
	p.columnMapping = make(map[string]int)

	reader := csv.NewReader(data)

	// Read header
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	// Map column headers to indices
	if err := p.mapHeaders(header); err != nil {
		return nil, fmt.Errorf("failed to map CSV headers: %w", err)
	}

	var transactions []*RawTransaction

	// Read data rows
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV row: %w", err)
		}

		txn, err := p.parseRow(record)
		if err != nil {
			// Log and skip invalid rows
			continue
		}

		if txn != nil {
			transactions = append(transactions, txn)
		}
	}

	return transactions, nil
}

// mapHeaders maps CSV column headers to expected field names
func (p *CSVParser) mapHeaders(header []string) error {
	requiredFields := map[string]bool{
		"date":     false,
		"merchant": false,
	}

	for i, col := range header {
		normalized := strings.ToLower(strings.TrimSpace(col))

		// Try exact match
		if normalized == "date" {
			p.columnMapping["date"] = i
			requiredFields["date"] = true
			continue
		}

		// Try fuzzy match for common variations
		if strings.Contains(normalized, "date") && !strings.Contains(normalized, "update") {
			if _, exists := p.columnMapping["date"]; !exists {
				p.columnMapping["date"] = i
				requiredFields["date"] = true
			}
			continue
		}

		if strings.Contains(normalized, "merchant") || strings.Contains(normalized, "narration") || strings.Contains(normalized, "payee") || strings.Contains(normalized, "description") {
			if _, exists := p.columnMapping["merchant"]; !exists {
				p.columnMapping["merchant"] = i
				requiredFields["merchant"] = true
			}
			continue
		}

		// Handle amount/debit/credit columns
		if normalized == "debit" {
			p.columnMapping["debit"] = i
		} else if normalized == "credit" {
			p.columnMapping["credit"] = i
		} else if normalized == "amount" {
			p.columnMapping["amount"] = i
		}

		// Handle balance column
		if normalized == "balance" {
			p.columnMapping["balance"] = i
		}

		if strings.Contains(normalized, "type") && !strings.Contains(normalized, "date") {
			if _, exists := p.columnMapping["type"]; !exists {
				p.columnMapping["type"] = i
			}
		}
	}

	// Check that we found required fields
	for field, found := range requiredFields {
		if !found {
			return fmt.Errorf("required column '%s' not found in CSV header", field)
		}
	}

	// Check for amount source (either single amount column or debit/credit pair)
	hasAmount := false
	if _, ok := p.columnMapping["amount"]; ok {
		hasAmount = true
	}
	if _, debitOk := p.columnMapping["debit"]; debitOk {
		if _, creditOk := p.columnMapping["credit"]; creditOk {
			hasAmount = true
		}
	}

	if !hasAmount {
		return fmt.Errorf("no amount column found (need 'amount' or both 'debit' and 'credit')")
	}

	return nil
}

// parseRow parses a single CSV row into a transaction
func (p *CSVParser) parseRow(record []string) (*RawTransaction, error) {
	if len(record) == 0 {
		return nil, fmt.Errorf("empty record")
	}

	// Extract date and merchant (always required)
	dateStr := p.getField(record, "date")
	merchant := p.getField(record, "merchant")

	if dateStr == "" || merchant == "" {
		return nil, fmt.Errorf("missing required fields (date or merchant)")
	}

	// Parse date (try common formats)
	txnDate, err := parseDate(dateStr)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
	}

	// Extract amount and type from either:
	// 1. Single "amount" and "type" columns, or
	// 2. Separate "debit" and "credit" columns
	var amount float64
	var txnType string

	// Try single amount column first
	amountStr := p.getField(record, "amount")
	if amountStr != "" {
		amount, err = parseAmount(amountStr)
		if err != nil {
			return nil, fmt.Errorf("invalid amount: %w", err)
		}
		if amount < 0 {
			amount = -amount
		}

		typeStr := p.getField(record, "type")
		if typeStr == "" {
			return nil, fmt.Errorf("missing type field")
		}
		txnType = strings.ToUpper(strings.TrimSpace(typeStr))
	} else {
		// Try debit/credit columns
		debitStr := p.getField(record, "debit")
		creditStr := p.getField(record, "credit")

		if debitStr != "" && debitStr != "0" && debitStr != "0.00" {
			amount, err = parseAmount(debitStr)
			if err != nil {
				return nil, fmt.Errorf("invalid debit amount: %w", err)
			}
			txnType = "DEBIT"
		} else if creditStr != "" && creditStr != "0" && creditStr != "0.00" {
			amount, err = parseAmount(creditStr)
			if err != nil {
				return nil, fmt.Errorf("invalid credit amount: %w", err)
			}
			txnType = "CREDIT"
		} else {
			return nil, fmt.Errorf("no amount found in debit or credit")
		}
	}

	// Validate transaction type
	if txnType != "DEBIT" && txnType != "CREDIT" {
		return nil, fmt.Errorf("invalid transaction type: %s", txnType)
	}

	rawTxn := &RawTransaction{
		Date:     txnDate,
		Merchant: strings.TrimSpace(merchant),
		Amount:   amount,
		Type:     txnType,
		RawData: map[string]interface{}{
			"source": "csv",
		},
	}

	// Optional fields
	if balance := p.getField(record, "balance"); balance != "" {
		if bal, err := parseAmount(balance); err == nil {
			rawTxn.Balance = &bal
		}
	}

	if desc := p.getField(record, "description"); desc != "" {
		rawTxn.Description = strings.TrimSpace(desc)
	}

	return rawTxn, nil
}

// getField retrieves a field value from a record by column mapping
func (p *CSVParser) getField(record []string, fieldName string) string {
	idx, ok := p.columnMapping[fieldName]
	if !ok || idx >= len(record) {
		return ""
	}
	return strings.TrimSpace(record[idx])
}

// parseDate attempts to parse a date string in common formats
func parseDate(dateStr string) (time.Time, error) {
	dateStr = strings.TrimSpace(dateStr)

	formats := []string{
		"2006-01-02",
		"02-01-2006",
		"02/01/2006", // DD/MM/YYYY (Indian format)
		"01/02/2006", // MM/DD/YYYY (US format)
		"2006/01/02",
		"January 2, 2006",
		"02 Jan 2006",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

// parseAmount parses an amount string, handling various formats
func parseAmount(amountStr string) (float64, error) {
	amountStr = strings.TrimSpace(amountStr)
	// Remove common currency symbols and spaces
	amountStr = strings.TrimFunc(amountStr, func(r rune) bool {
		return r == '$' || r == '€' || r == '₹' || r == ',' || r == ' '
	})

	var amount float64
	_, err := fmt.Sscanf(amountStr, "%f", &amount)
	if err != nil {
		return 0, fmt.Errorf("invalid amount: %s", amountStr)
	}

	return amount, nil
}
