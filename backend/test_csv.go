package main

import (
	"fmt"
	"strings"

	"money-planner/backend/internal/statement"
)

func main() {
	csv := `Date,Narration,Debit,Credit,Balance
01/07/2026,Opening Balance,,50000.00,50000.00
02/07/2026,Salary Deposit,,75000.00,125000.00
03/07/2026,Electricity Bill,2500.00,,122500.00
04/07/2026,Grocery,1200.50,,121299.50`
	
	parser := statement.NewCSVParser()
	txns, err := parser.ParseCSV(strings.NewReader(csv))
	
	if err != nil {
		fmt.Printf("Parser Error: %v\n", err)
		return
	}
	
	fmt.Printf("Parsed %d transactions\n", len(txns))
	for i, txn := range txns {
		fmt.Printf("%d: Date=%v, Merchant=%s, Amount=%.2f, Type=%s\n", i, txn.Date, txn.Merchant, txn.Amount, txn.Type)
	}
}
