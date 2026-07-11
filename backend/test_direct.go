package main

import (
	"fmt"
	"strings"
	"money-planner/backend/internal/statement"
)

func main() {
	csv := `Date,Narration,Debit,Credit,Balance
22/07/2026,FixedCode1,,2000.00,2000.00
23/07/2026,FixedCode2,1000.00,,1000.00`
	
	parser := statement.NewCSVParser()
	txns, err := parser.ParseCSV(strings.NewReader(csv))
	fmt.Printf("Error: %v\n", err)
	fmt.Printf("Transactions: %d\n", len(txns))
	for i, t := range txns {
		fmt.Printf("%d: %s - %.2f (%s)\n", i, t.Merchant, t.Amount, t.Type)
	}
}
