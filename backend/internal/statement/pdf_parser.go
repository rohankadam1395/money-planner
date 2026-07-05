package statement

import (
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"
)

// PDFParser handles parsing of PDF statement files
type PDFParser struct {
	format StatementFormat
}

// NewPDFParser creates a new PDF parser
func NewPDFParser(format StatementFormat) *PDFParser {
	return &PDFParser{
		format: format,
	}
}

// ParsePDF extracts transactions from a PDF file
// Note: This is a simplified implementation using text extraction
// In production, use a proper PDF library like pdfcpu or unidoc
func (p *PDFParser) ParsePDF(data io.Reader) ([]*RawTransaction, error) {
	// Read the entire PDF into memory
	content, err := io.ReadAll(data)
	if err != nil {
		return nil, fmt.Errorf("failed to read PDF: %w", err)
	}

	// Extract text from PDF (simplified - in production use proper PDF library)
	text := p.extractTextFromPDF(content)

	// Parse the extracted text for transactions
	return p.parseTransactionTable(text)
}

// extractTextFromPDF performs basic text extraction from PDF
// This is a simplified implementation - production code should use a proper PDF library
func (p *PDFParser) extractTextFromPDF(content []byte) string {
	// Convert bytes to string and filter for readable content
	// In production, use a library like pdfcpu or unidoc

	// For now, return a placeholder indicating PDF needs proper parsing
	// This allows the framework to work while we integrate a real PDF library
	text := string(content)

	// Remove binary PDF markers but keep the structure
	text = strings.TrimSpace(text)
	if len(text) == 0 {
		return ""
	}

	return text
}

// parseTransactionTable parses a transaction table from PDF text
func (p *PDFParser) parseTransactionTable(text string) ([]*RawTransaction, error) {
	if text == "" {
		return nil, fmt.Errorf("no text content extracted from PDF")
	}

	var transactions []*RawTransaction

	// Look for lines that match transaction patterns
	// Format: Date | Merchant | Amount | Type | Balance (optional)
	lines := strings.Split(text, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Try to parse as a transaction line
		txn := p.parseTransactionLine(line)
		if txn != nil {
			transactions = append(transactions, txn)
		}
	}

	if len(transactions) == 0 {
		return nil, fmt.Errorf("no transactions found in PDF")
	}

	return transactions, nil
}

// parseTransactionLine attempts to parse a single line as a transaction
func (p *PDFParser) parseTransactionLine(line string) *RawTransaction {
	// Pattern to match transaction lines
	// Expects format like: YYYY-MM-DD | Merchant | Amount | DEBIT/CREDIT
	fields := strings.Split(line, "|")
	if len(fields) < 4 {
		return nil
	}

	for i := range fields {
		fields[i] = strings.TrimSpace(fields[i])
	}

	// Parse date
	date, err := parseDate(fields[0])
	if err != nil {
		return nil
	}

	// Parse merchant
	merchant := fields[1]
	if merchant == "" {
		return nil
	}

	// Parse amount
	amount, err := parseAmount(fields[2])
	if err != nil {
		return nil
	}

	// Parse type
	txnType := strings.ToUpper(fields[3])
	if txnType != "DEBIT" && txnType != "CREDIT" {
		return nil
	}

	txn := &RawTransaction{
		Date:     date,
		Merchant: merchant,
		Amount:   amount,
		Type:     txnType,
		RawData: map[string]interface{}{
			"source": "pdf",
		},
	}

	// Optional balance
	if len(fields) > 4 {
		if bal, err := parseAmount(fields[4]); err == nil {
			txn.Balance = &bal
		}
	}

	return txn
}

// StatementFormat represents a bank statement format configuration
type StatementFormat interface {
	GetName() string
	GetBankCode() string
	MapColumns(header []string) (map[string]int, error)
}

// HDFCFormat handles HDFC statement parsing
type HDFCFormat struct{}

func (f *HDFCFormat) GetName() string {
	return "HDFC"
}

func (f *HDFCFormat) GetBankCode() string {
	return "HDFC"
}

func (f *HDFCFormat) MapColumns(header []string) (map[string]int, error) {
	mapping := make(map[string]int)

	for i, col := range header {
		normalized := strings.ToLower(strings.TrimSpace(col))

		switch {
		case strings.Contains(normalized, "date") || strings.Contains(normalized, "transaction date"):
			mapping["date"] = i
		case strings.Contains(normalized, "description") || strings.Contains(normalized, "particulars"):
			mapping["merchant"] = i
		case strings.Contains(normalized, "withdrawal") || strings.Contains(normalized, "debit"):
			mapping["debit_amount"] = i
		case strings.Contains(normalized, "deposit") || strings.Contains(normalized, "credit"):
			mapping["credit_amount"] = i
		case strings.Contains(normalized, "balance"):
			mapping["balance"] = i
		}
	}

	return mapping, nil
}

// ICICIFormat handles ICICI statement parsing
type ICICIFormat struct{}

func (f *ICICIFormat) GetName() string {
	return "ICICI"
}

func (f *ICICIFormat) GetBankCode() string {
	return "ICIC"
}

func (f *ICICIFormat) MapColumns(header []string) (map[string]int, error) {
	mapping := make(map[string]int)

	for i, col := range header {
		normalized := strings.ToLower(strings.TrimSpace(col))

		switch {
		case strings.Contains(normalized, "date"):
			mapping["date"] = i
		case strings.Contains(normalized, "description"):
			mapping["merchant"] = i
		case strings.Contains(normalized, "amount"):
			mapping["amount"] = i
		case strings.Contains(normalized, "type"):
			mapping["type"] = i
		case strings.Contains(normalized, "balance"):
			mapping["balance"] = i
		}
	}

	return mapping, nil
}

// ParseStatementPeriod extracts statement period dates from text
func ParseStatementPeriod(text string) (time.Time, time.Time, error) {
	// Look for period dates in the text
	// Format: "Statement Period: DD-MM-YYYY to DD-MM-YYYY"
	periodRegex := regexp.MustCompile(`(?i)(?:period|from|date).*?(\d{1,2}[-/]\d{1,2}[-/]\d{4}).*?to.*?(\d{1,2}[-/]\d{1,2}[-/]\d{4})`)

	matches := periodRegex.FindStringSubmatch(text)
	if len(matches) < 3 {
		return time.Time{}, time.Time{}, fmt.Errorf("could not find statement period")
	}

	startDate, err := parseDate(matches[1])
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid start date: %w", err)
	}

	endDate, err := parseDate(matches[2])
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid end date: %w", err)
	}

	return startDate, endDate, nil
}
