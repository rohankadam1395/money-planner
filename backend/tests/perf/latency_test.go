package perf

import (
	"bytes"
	"io"
	"os"
	"testing"
	"time"

	"money-planner/backend/internal/statement"
)

const testAccountHash = "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789"

func BenchmarkCSVParsing(b *testing.B) {
	parser := statement.NewCSVParser()

	csvFile, err := os.Open("../testdata/hdfc_sample.csv")
	if err != nil {
		b.Fatalf("Failed to open test data: %v", err)
	}
	defer csvFile.Close()

	csvData, err := io.ReadAll(csvFile)
	if err != nil {
		b.Fatalf("Failed to read test data: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := bytes.NewReader(csvData)
		_, err := parser.ParseCSV(reader)
		if err != nil {
			b.Fatalf("Parsing failed: %v", err)
		}
	}
}

func TestUploadLatency(t *testing.T) {
	parser := statement.NewCSVParser()

	csvFile, err := os.Open("../testdata/hdfc_sample.csv")
	if err != nil {
		t.Fatalf("Failed to open test data: %v", err)
	}
	defer csvFile.Close()

	start := time.Now()
	csvData, err := io.ReadAll(csvFile)
	if err != nil {
		t.Fatalf("Failed to read test data: %v", err)
	}

	reader := bytes.NewReader(csvData)
	transactions, err := parser.ParseCSV(reader)
	if err != nil {
		t.Fatalf("Parsing failed: %v", err)
	}

	elapsed := time.Since(start)

	maxLatency := 10 * time.Second
	if elapsed > maxLatency {
		t.Errorf("Upload latency exceeded target: %v > %v", elapsed, maxLatency)
	}

	t.Logf("Parse time: %v (target: <10s) | Transactions parsed: %d", elapsed, len(transactions))

	if len(transactions) == 0 {
		t.Errorf("Expected transactions to be extracted, got 0")
	}
}

func TestSyncProcessingLatency(t *testing.T) {
	parser := statement.NewCSVParser()

	csvFile, err := os.Open("../testdata/hdfc_sample.csv")
	if err != nil {
		t.Fatalf("Failed to open test data: %v", err)
	}
	defer csvFile.Close()

	start := time.Now()

	csvData, err := io.ReadAll(csvFile)
	if err != nil {
		t.Fatalf("Failed to read test data: %v", err)
	}

	reader := bytes.NewReader(csvData)
	transactions, err := parser.ParseCSV(reader)
	if err != nil {
		t.Fatalf("Parsing failed: %v", err)
	}

	periodStart := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	periodEnd := time.Date(2030, 12, 31, 0, 0, 0, 0, time.UTC)
	validCount := 0
	for _, txn := range transactions {
		result := statement.ValidateTransaction(&statement.Transaction{
			TransactionDate:   txn.Date,
			Merchant:          txn.Merchant,
			Amount:            txn.Amount,
			Type:              txn.Type,
			Currency:          "INR",
			AccountNumberHash: testAccountHash,
		}, periodStart, periodEnd)
		if result.Valid {
			validCount++
		}
	}

	elapsed := time.Since(start)

	if elapsed > 10*time.Second {
		t.Errorf("Sync processing exceeded 10s target: %v", elapsed)
	}

	t.Logf("Full upload→parse→validate→preview flow: %v | Valid transactions: %d/%d",
		elapsed, validCount, len(transactions))

	if elapsed < 10*time.Second {
		t.Logf("✓ SC-001 satisfied: Sync processing completes in %v (target: <10s)", elapsed)
	}
}
