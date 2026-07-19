package perf

import (
	"context"
	"fmt"
	"testing"
	"time"

	"money-planner/backend/internal/categorization"
	"money-planner/backend/internal/categorization/providers"
)

// TestSingleVsBatchLLMLatency measures REAL latency with actual Ollama
// ⚠️ REQUIRES: Ollama running locally (ollama serve)
// RUN: go test -v -run TestSingleVsBatchLLMLatency ./tests/perf/... -timeout 5m
func TestSingleVsBatchLLMLatency(t *testing.T) {
	ctx := context.Background()

	// REAL Ollama provider - connects to http://localhost:11434
	ollamaProvider := providers.NewOllamaProvider("http://localhost:11434", "mistral")

	// Test data: 10 unknown merchants (names not in seed dictionary)
	unknowns := []struct {
		merchant string
		amount   float64
	}{
		{"Aashish Restaurant Pvt Ltd", 500},
		{"Local Shop XYZ Traders", 300},
		{"Unknown Cafe & Bakery", 200},
		{"Mystery Vendor Inc", 1500},
		{"Xyz Holdings Limited", 2000},
		{"ABC Corporation", 5000},
		{"Random Store 123", 800},
		{"New Business Startup", 1200},
		{"Unknown Merchant Co", 600},
		{"Test Company Private", 4000},
	}

	t.Run("SINGLE TRANSACTION: 10 sequential LLM calls (current approach)", func(t *testing.T) {
		if testing.Short() {
			t.Skip("Skipping real LLM test in short mode")
		}

		t.Logf("📊 Testing 10 individual merchant → LLM calls (sequential)")
		start := time.Now()
		callCount := 0
		for i, txn := range unknowns {
			callStart := time.Now()
			cat, conf, _, err := ollamaProvider.Categorize(ctx, txn.merchant, txn.amount)
			callElapsed := time.Since(callStart)
			callCount++

			if err != nil {
				t.Logf("  [%d] ❌ %s: ERROR %v", i+1, txn.merchant, err)
				continue
			}
			t.Logf("  [%d] ✓ %s → %s (conf:%.2f, latency:%v)", i+1, txn.merchant, cat, conf, callElapsed)
		}
		elapsed := time.Since(start)
		avgPerCall := elapsed / time.Duration(callCount)

		t.Logf("\n📈 SINGLE TRANSACTION RESULTS:")
		t.Logf("   Total time: %v", elapsed)
		t.Logf("   API calls: %d", callCount)
		t.Logf("   Avg per call: %v", avgPerCall)
		t.Logf("   Per-transaction avg: %v", elapsed/time.Duration(len(unknowns)))
	})

	t.Run("BATCH REAL: 1 combined LLM call via CategorizeBatch", func(t *testing.T) {
		if testing.Short() {
			t.Skip("Skipping real LLM test in short mode")
		}

		t.Logf("📊 Testing real batch categorization: CategorizeBatch() for all 10 merchants in 1 call")

		start := time.Now()

		// Build batch items
		items := make([]categorization.BatchItem, len(unknowns))
		for i, txn := range unknowns {
			items[i] = categorization.BatchItem{
				Merchant: txn.merchant,
				Amount:   txn.amount,
			}
		}

		// Call real batch method (ollamaProvider has CategorizeBatch method)
		results, err := ollamaProvider.CategorizeBatch(ctx, items)
		if err != nil {
			t.Logf("  ❌ Batch call error: %v", err)
			return
		}
		elapsed := time.Since(start)

		successCount := 0
		for i, res := range results {
			if res.Err == nil && res.Category != "Uncategorized" {
				t.Logf("  ✓ [%d] %s → %s (conf:%.2f)", i, unknowns[i].merchant, res.Category, res.Confidence)
				successCount++
			} else {
				t.Logf("  ✗ [%d] %s → %s (err:%v)", i, unknowns[i].merchant, res.Category, res.Err)
			}
		}

		t.Logf("\n📈 BATCH REAL RESULTS:")
		t.Logf("   Total time: %v (1 API call)", elapsed)
		t.Logf("   Successful: %d/%d", successCount, len(unknowns))
		t.Logf("   Per-transaction avg: %v", elapsed/time.Duration(len(unknowns)))
	})

	t.Run("CONCURRENT: 5 parallel streams (SC-105 compliance)", func(t *testing.T) {
		if testing.Short() {
			t.Skip("Skipping real LLM test in short mode")
		}

		t.Logf("📊 Testing 5 concurrent LLM streams (with per-call timeouts)")
		start := time.Now()
		results := make(chan string, len(unknowns))
		errors := make(chan error, len(unknowns))
		completed := make(chan bool, 5) // Track goroutine completion

		// Spawn 5 concurrent workers with timeout per call
		for stream := 0; stream < 5; stream++ {
			go func(streamID int) {
				defer func() { completed <- true }() // Signal completion

				for i := streamID; i < len(unknowns); i += 5 {
					// Each call has a 15s timeout (Ollama inference can be slow)
					callCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
					callStart := time.Now()
					cat, conf, _, callErr := ollamaProvider.Categorize(callCtx, unknowns[i].merchant, unknowns[i].amount)
					cancel()
					callElapsed := time.Since(callStart)

					if callErr != nil {
						errors <- fmt.Errorf("[stream %d] %s: %v", streamID, unknowns[i].merchant, callErr)
						continue
					}
					results <- fmt.Sprintf("[stream %d] %s → %s (%.2f, %v)", streamID, unknowns[i].merchant, cat, conf, callElapsed)
				}
			}(stream)
		}

		// Collect results with proper timeout
		successCount := 0
		errorCount := 0
		goroutinesCompleted := 0

		for goroutinesCompleted < 5 {
			select {
			case res := <-results:
				t.Logf("  ✓ %s", res)
				successCount++
			case err := <-errors:
				t.Logf("  ❌ %v", err)
				errorCount++
			case <-completed:
				goroutinesCompleted++
			case <-time.After(120 * time.Second):
				t.Fatalf("❌ Test timeout: goroutines hung (completed: %d/5, results: %d, errors: %d)", goroutinesCompleted, successCount, errorCount)
			}
		}

		// Drain any remaining results
		for i := 0; i < len(unknowns); i++ {
			select {
			case res := <-results:
				t.Logf("  ✓ %s", res)
				successCount++
			case err := <-errors:
				t.Logf("  ❌ %v", err)
				errorCount++
			default:
				break
			}
		}

		elapsed := time.Since(start)

		t.Logf("\n📈 CONCURRENT RESULTS (SC-105):")
		t.Logf("   Total time: %v", elapsed)
		t.Logf("   Successful: %d/%d | Errors: %d", successCount, len(unknowns), errorCount)
		if successCount > 0 {
			t.Logf("   Per-transaction avg: %v", elapsed/time.Duration(successCount))
		}
	})
}
