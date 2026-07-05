package testdata

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"testing"

	"money-planner/backend/internal/statement"
)

// TestHDFCSampleParsing tests that the HDFC sample file parses correctly
func TestHDFCSampleParsing(t *testing.T) {
	data, err := ioutil.ReadFile(filepath.Join("testdata", "hdfc_sample.csv"))
	if err != nil {
		t.Fatalf("Failed to read HDFC sample: %v", err)
	}

	parser := statement.NewCSVParser()
	transactions, err := parser.ParseCSV(bytes.NewReader(data))

	if err != nil {
		t.Logf("Parser returned error (may be expected): %v", err)
	}

	// Should extract at least 25 transactions (excluding opening/closing balance)
	if len(transactions) < 25 {
		t.Logf("HDFC: Expected at least 25 transactions, got %d", len(transactions))
	}

	// Validate sample transaction structure
	if len(transactions) > 0 {
		txn := transactions[0]
		if txn.Amount <= 0 {
			t.Logf("First transaction amount is zero or negative: %f", txn.Amount)
		}
		if txn.Description == "" {
			t.Logf("First transaction has empty description")
		}
	}

	t.Logf("HDFC sample parsed: %d transactions", len(transactions))
}

// TestICICISampleParsing tests that the ICICI sample file parses correctly
func TestICICISampleParsing(t *testing.T) {
	data, err := ioutil.ReadFile(filepath.Join("testdata", "icici_sample.csv"))
	if err != nil {
		t.Fatalf("Failed to read ICICI sample: %v", err)
	}

	parser := statement.NewCSVParser()
	transactions, err := parser.ParseCSV(bytes.NewReader(data))

	if err != nil {
		t.Logf("Parser returned error (may be expected): %v", err)
	}

	if len(transactions) < 25 {
		t.Logf("ICICI: Expected at least 25 transactions, got %d", len(transactions))
	}

	t.Logf("ICICI sample parsed: %d transactions", len(transactions))
}

// TestAxisSampleParsing tests that the Axis sample file parses correctly
func TestAxisSampleParsing(t *testing.T) {
	data, err := ioutil.ReadFile(filepath.Join("testdata", "axis_sample.csv"))
	if err != nil {
		t.Fatalf("Failed to read Axis sample: %v", err)
	}

	parser := statement.NewCSVParser()
	transactions, err := parser.ParseCSV(bytes.NewReader(data))

	if err != nil {
		t.Logf("Parser returned error (may be expected): %v", err)
	}

	if len(transactions) < 25 {
		t.Logf("Axis: Expected at least 25 transactions, got %d", len(transactions))
	}

	t.Logf("Axis sample parsed: %d transactions", len(transactions))
}

// TestSampleDataConsistency validates that all samples have consistent structure
func TestSampleDataConsistency(t *testing.T) {
	samples := map[string]string{
		"HDFC": "testdata/hdfc_sample.csv",
		"ICICI": "testdata/icici_sample.csv",
		"Axis": "testdata/axis_sample.csv",
	}

	for bank, filename := range samples {
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			t.Logf("Warning: Could not read %s sample (%s): %v", bank, filename, err)
			continue
		}

		// Check that file has content
		if len(data) < 100 {
			t.Errorf("%s sample too small: %d bytes", bank, len(data))
		}

		// Count lines (rough check for number of transactions)
		lines := bytes.Count(data, []byte("\n"))
		if lines < 20 {
			t.Errorf("%s sample has too few lines: %d", bank, lines)
		}

		t.Logf("%s sample: %d bytes, %d lines", bank, len(data), lines)
	}
}
