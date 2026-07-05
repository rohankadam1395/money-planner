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
		"amount":   false,
		"type":     false,
	}

	for i, col := range header {
		normalized := strings.ToLower(strings.TrimSpace(col))

		// Try exact match
		if _, ok := requiredFields[normalized]; ok {
			p.columnMapping[normalized] = i
			requiredFields[normalized] = true
			continue
		}

		// Try fuzzy match for common variations
		switch {
		case strings.Contains(normalized, "date") || strings.Contains(normalized, "transaction"):
			if _, exists := p.columnMapping["date"]; !exists {
				p.columnMapping["date"] = i
				requiredFields["date"] = true
			}
		case strings.Contains(normalized, "merchant") || strings.Contains(normalized, "payee") || strings.Contains(normalized, "description"):
			if _, exists := p.columnMapping["merchant"]; !exists {
				p.columnMapping["merchant"] = i
				requiredFields["merchant"] = true
			}
		case strings.Contains(normalized, "amount") || strings.Contains(normalized, "value"):
			if _, exists := p.columnMapping["amount"]; !exists {
				p.columnMapping["amount"] = i
				requiredFields["amount"] = true
			}
		case strings.Contains(normalized, "type") || strings.Contains(normalized, "debit") || strings.Contains(normalized, "credit"):
			if _, exists := p.columnMapping["type"]; !exists {
				p.columnMapping["type"] = i
				requiredFields["type"] = true
			}
		}
	}

	// Check that we found all required fields
	for field, found := range requiredFields {
		if !found {
			return fmt.Errorf("required column '%s' not found in CSV header", field)
		}
	}

	return nil
}

// parseRow parses a single CSV row into a transaction
func (p *CSVParser) parseRow(record []string) (*RawTransaction, error) {
	if len(record) == 0 {
		return nil, fmt.Errorf("empty record")
	}

	// Extract fields based on column mapping
	dateStr := p.getField(record, "date")
	merchant := p.getField(record, "merchant")
	amountStr := p.getField(record, "amount")
	txnType := p.getField(record, "type")

	if dateStr == "" || merchant == "" || amountStr == "" || txnType == "" {
		return nil, fmt.Errorf("missing required fields")
	}

	// Parse date (try common formats)
	txnDate, err := parseDate(dateStr)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
	}

	// Parse amount
	amount, err := parseAmount(amountStr)
	if err != nil {
		return nil, fmt.Errorf("invalid amount: %w", err)
	}

	// Normalize transaction type
	txnType = strings.ToUpper(strings.TrimSpace(txnType))
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
		"01/02/2006",
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
