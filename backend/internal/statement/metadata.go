package statement

import (
	"regexp"
	"strings"
	"time"
)

// MetadataExtractor extracts statement period and other metadata from statement files
type MetadataExtractor struct{}

// NewMetadataExtractor creates a new metadata extractor
func NewMetadataExtractor() *MetadataExtractor {
	return &MetadataExtractor{}
}

// StatementMetadata holds extracted metadata from a statement
type StatementMetadata struct {
	PeriodStart time.Time
	PeriodEnd   time.Time
	StatementID string
}

// ExtractFromText extracts metadata from statement text content
func (m *MetadataExtractor) ExtractFromText(content string, bankCode string) *StatementMetadata {
	metadata := &StatementMetadata{
		PeriodStart: time.Now().AddDate(0, -1, 0),
		PeriodEnd:   time.Now(),
	}

	// Try to extract period dates using bank-specific patterns
	switch bankCode {
	case "HDFC":
		m.extractHDFCPeriod(content, metadata)
	case "ICICI":
		m.extractICICIPeriod(content, metadata)
	case "AXIS":
		m.extractAxisPeriod(content, metadata)
	case "SBI":
		m.extractSBIPeriod(content, metadata)
	default:
		m.extractGenericPeriod(content, metadata)
	}

	return metadata
}

// extractHDFCPeriod extracts period from HDFC statement
func (m *MetadataExtractor) extractHDFCPeriod(content string, metadata *StatementMetadata) {
	patterns := []string{
		`Statement Period.*?(\d{2}[/-]\d{2}[/-]\d{4}).*?(\d{2}[/-]\d{2}[/-]\d{4})`,
		`Period.*?(\d{2}[/-]\d{2}[/-]\d{4})\s+to\s+(\d{2}[/-]\d{2}[/-]\d{4})`,
		`From\s+(\d{2}[/-]\d{2}[/-]\d{4}).*?To\s+(\d{2}[/-]\d{2}[/-]\d{4})`,
	}
	m.tryExtractPeriod(content, patterns, metadata)
}

// extractICICIPeriod extracts period from ICICI statement
func (m *MetadataExtractor) extractICICIPeriod(content string, metadata *StatementMetadata) {
	patterns := []string{
		`Statement For Period.*?(\d{2}[/-]\d{2}[/-]\d{4}).*?(\d{2}[/-]\d{2}[/-]\d{4})`,
		`For Period.*?(\d{2}[/-]\d{2}[/-]\d{4}).*?(\d{2}[/-]\d{2}[/-]\d{4})`,
		`Period.*?(\d{2}[/-]\d{2}[/-]\d{4})\s+to\s+(\d{2}[/-]\d{2}[/-]\d{4})`,
	}
	m.tryExtractPeriod(content, patterns, metadata)
}

// extractAxisPeriod extracts period from Axis statement
func (m *MetadataExtractor) extractAxisPeriod(content string, metadata *StatementMetadata) {
	patterns := []string{
		`Statement Period.*?(\d{2}[/-]\d{2}[/-]\d{4}).*?(\d{2}[/-]\d{2}[/-]\d{4})`,
		`Period from\s+(\d{2}[/-]\d{2}[/-]\d{4})\s+to\s+(\d{2}[/-]\d{2}[/-]\d{4})`,
	}
	m.tryExtractPeriod(content, patterns, metadata)
}

// extractSBIPeriod extracts period from SBI statement
func (m *MetadataExtractor) extractSBIPeriod(content string, metadata *StatementMetadata) {
	patterns := []string{
		`Statement Period.*?(\d{2}[/-]\d{2}[/-]\d{4}).*?(\d{2}[/-]\d{2}[/-]\d{4})`,
		`Period.*?(\d{2}[/-]\d{2}[/-]\d{4}).*?(\d{2}[/-]\d{2}[/-]\d{4})`,
	}
	m.tryExtractPeriod(content, patterns, metadata)
}

// extractGenericPeriod extracts period using generic patterns
func (m *MetadataExtractor) extractGenericPeriod(content string, metadata *StatementMetadata) {
	patterns := []string{
		`(\d{2}[/-]\d{2}[/-]\d{4}).*?to.*?(\d{2}[/-]\d{2}[/-]\d{4})`,
		`from\s+(\d{2}[/-]\d{2}[/-]\d{4}).*?(\d{2}[/-]\d{2}[/-]\d{4})`,
		`Period:\s*(\d{2}[/-]\d{2}[/-]\d{4})\s+-\s+(\d{2}[/-]\d{2}[/-]\d{4})`,
	}
	m.tryExtractPeriod(content, patterns, metadata)
}

// tryExtractPeriod attempts to extract period dates using a list of regex patterns
func (m *MetadataExtractor) tryExtractPeriod(content string, patterns []string, metadata *StatementMetadata) {
	// Normalize content for easier pattern matching
	normalizedContent := strings.ToLower(content)

	for _, pattern := range patterns {
		re := regexp.MustCompile(`(?i)` + pattern)
		matches := re.FindStringSubmatch(normalizedContent)

		if len(matches) >= 3 {
			startDate := m.parseDate(matches[1])
			endDate := m.parseDate(matches[2])

			if !startDate.IsZero() && !endDate.IsZero() {
				metadata.PeriodStart = startDate
				metadata.PeriodEnd = endDate
				return
			}
		}
	}
}

// parseDate attempts to parse a date string in multiple formats
func (m *MetadataExtractor) parseDate(dateStr string) time.Time {
	formats := []string{
		"02/01/2006",  // DD/MM/YYYY
		"02-01-2006",  // DD-MM-YYYY
		"01/02/2006",  // MM/DD/YYYY
		"01-02-2006",  // MM-DD-YYYY
		"2006-01-02",  // YYYY-MM-DD
		"02 Jan 2006", // DD Mon YYYY
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t
		}
	}

	return time.Time{}
}

// ExtractFromTransactions infers period from transaction dates
func (m *MetadataExtractor) ExtractFromTransactions(txns []*RawTransaction) *StatementMetadata {
	if len(txns) == 0 {
		return &StatementMetadata{
			PeriodStart: time.Now().AddDate(0, -1, 0),
			PeriodEnd:   time.Now(),
		}
	}

	var earliest, latest time.Time

	for _, txn := range txns {
		if txn.Date.IsZero() {
			continue
		}

		if earliest.IsZero() || txn.Date.Before(earliest) {
			earliest = txn.Date
		}
		if latest.IsZero() || txn.Date.After(latest) {
			latest = txn.Date
		}
	}

	return &StatementMetadata{
		PeriodStart: earliest,
		PeriodEnd:   latest,
	}
}
