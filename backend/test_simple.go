package main

import (
	"fmt"
	"strings"
	"money-planner/backend/internal/statement"
)

func main() {
	// Exact format that worked before
	csv := `Date,Narration,Debit,Credit,Balance
01/07/2026,Opening Balance,,50000.00,50000.00
02/07/2026,Salary,,75000.00,125000.00`
	
	parser := statement.NewCSVParser()
	txns, err := parser.ParseCSV(strings.NewReader(csv))
	fmt.Printf("Parse error: %v\n", err)
	fmt.Printf("Transactions parsed: %d\n", len(txns))
	if len(txns) > 0 {
		fmt.Printf("First: %+v\n", txns[0])
	}
}
