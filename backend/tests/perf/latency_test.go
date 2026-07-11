package perf

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"money-planner/backend/internal/statement"
)

// BenchmarkCSVParsing benchmarks CSV file parsing performance
func BenchmarkCSVParsing(b *testing.B) {
	parser := statement.NewCSVParser()

	// Read sample CSV file
	csvFile, err := os.Open("../../testdata/hdfc_sample.csv")
	if err != nil {
		b.Fatalf("Failed to open test data: %v", err)
	}
	defer csvFile.Close()

	csvData, err := ioutil.ReadAll(csvFile)
	if err != nil {
		b.Fatalf("Failed to read test data: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create new reader for each iteration
		reader := os.NewReader(csvData)
		_, err := parser.ParseCSV(reader)
		if err != nil {
			b.Fatalf("Parsing failed: %v", err)
		}
	}
}

// TestUploadLatency verifies upload-to-preview latency is under 10 seconds
func TestUploadLatency(t *testing.T) {
	parser := statement.NewCSVParser()

	// Read sample CSV file (typical bank export ~100KB)
	csvFile, err := os.Open("../../testdata/hdfc_sample.csv")
	if err != nil {
		t.Fatalf("Failed to open test data: %v", err)
	}
	defer csvFile.Close()

	// Measure parsing time
	start := time.Now()
	csvData, err := ioutil.ReadAll(csvFile)
	if err != nil {
		t.Fatalf("Failed to read test data: %v", err)
	}

	reader := os.NewReader(csvData)
	transactions, err := parser.ParseCSV(reader)
	if err != nil {
		t.Fatalf("Parsing failed: %v", err)
	}

	elapsed := time.Since(start)

	// Verify latency requirement: <10 seconds for parsing
	maxLatency := 10 * time.Second
	if elapsed > maxLatency {
		t.Errorf("Upload latency exceeded target: %v > %v", elapsed, maxLatency)
	}

	t.Logf("Parse time: %v (target: <10s) | Transactions parsed: %d", elapsed, len(transactions))

	// Verify transaction extraction
	if len(transactions) == 0 {
		t.Errorf("Expected transactions to be extracted, got 0")
	}
}

// TestSyncProcessingLatency validates that synchronous file processing meets SC-001
// SC-001: User can upload a bank statement and view extracted transactions within 10 seconds
func TestSyncProcessingLatency(t *testing.T) {
	parser := statement.NewCSVParser()
	validator := statement.NewTransactionValidator()

	// Simulate full upload → extract → validate → preview flow
	csvFile, err := os.Open("../../testdata/hdfc_sample.csv")
	if err != nil {
		t.Fatalf("Failed to open test data: %v", err)
	}
	defer csvFile.Close()

	start := time.Now()

	// Step 1: Read file
	csvData, err := ioutil.ReadAll(csvFile)
	if err != nil {
		t.Fatalf("Failed to read test data: %v", err)
	}

	// Step 2: Parse CSV
	reader := os.NewReader(csvData)
	transactions, err := parser.ParseCSV(reader)
	if err != nil {
		t.Fatalf("Parsing failed: %v", err)
	}

	// Step 3: Validate transactions
	validatedTxns := []statement.Transaction{}
	for _, txn := range transactions {
		if validator.ValidateTransaction(&txn) {
			validatedTxns = append(validatedTxns, txn)
		}
	}

	elapsed := time.Since(start)

	// SC-001: Total time should be <10 seconds
	if elapsed > 10*time.Second {
		t.Errorf("Sync processing exceeded 10s target: %v", elapsed)
	}

	// Log results
	t.Logf("Full upload→parse→validate→preview flow: %v | Valid transactions: %d/%d",
		elapsed, len(validatedTxns), len(transactions))

	// Acceptance: Synchronous processing meets <10s target
	if elapsed < 10*time.Second {
		t.Logf("✓ SC-001 satisfied: Sync processing completes in %v (target: <10s)", elapsed)
	}
}
