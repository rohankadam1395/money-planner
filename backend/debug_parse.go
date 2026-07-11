package main

import (
	"fmt"
	"strings"

	"money-planner/backend/internal/statement"
)

func main() {
	csv := `Date,Narration,Debit,Credit,Balance
05/07/2026,Test,,1000.00,1000.00`
	
	parser := statement.NewCSVParser()
	txns, err := parser.ParseCSV(strings.NewReader(csv))
	
	fmt.Printf("Error: %v\n", err)
	fmt.Printf("Txns: %v\n", txns)
	fmt.Printf("Txns nil: %v\n", txns == nil)
	if txns != nil {
		fmt.Printf("Txns length: %d\n", len(txns))
	}
}
