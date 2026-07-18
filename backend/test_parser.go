//go:build ignore

package main

import (
	"fmt"
	"os"
	"strings"

	"money-planner/backend/internal/statement"
)

func main() {
	// Read sample CSV
	content, _ := os.ReadFile("tests/testdata/hdfc_sample.csv")
	
	parser := statement.NewCSVParser()
	txns, err := parser.ParseCSV(strings.NewReader(string(content)))
	
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	
	fmt.Printf("Parsed %d transactions\n", len(txns))
	if len(txns) > 0 {
		fmt.Printf("First transaction: %+v\n", txns[0])
	}
}
