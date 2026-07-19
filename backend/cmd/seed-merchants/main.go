package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

func main() {
	// Get database URL from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable not set")
	}

	// Connect to database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("✓ Connected to database")

	// First, seed the categories table
	categories := []struct {
		name string
		desc string
		color string
		icon string
	}{
		{"Food & Dining", "Restaurants, food delivery, groceries", "#FF6B6B", "🍔"},
		{"Shopping", "Retail, clothing, online marketplaces", "#4ECDC4", "🛍️"},
		{"Transport", "Ride-sharing, fuel, transport", "#45B7D1", "🚗"},
		{"Housing", "Rent, property, home maintenance", "#F7B731", "🏠"},
		{"Utilities", "Electricity, water, internet, phone", "#5F27CD", "💡"},
		{"Entertainment", "Movies, streaming, games, events", "#EE5A6F", "🎬"},
		{"Income", "Salary, freelance, refunds", "#2ECC71", "💰"},
		{"Healthcare", "Medical, pharmacy, gym, insurance", "#FF4757", "🏥"},
		{"Education", "Tuition, courses, books", "#1E90FF", "📚"},
		{"Miscellaneous", "Gifts, charity, other", "#95A5A6", "📌"},
	}

	for _, cat := range categories {
		_, err := db.Exec(
			`INSERT INTO categories (id, name, description, color, icon, is_predefined, created_at, updated_at)
			 VALUES (gen_random_uuid(), $1, $2, $3, $4, true, NOW(), NOW())
			 ON CONFLICT (name) DO NOTHING`,
			cat.name, cat.desc, cat.color, cat.icon)
		if err != nil {
			log.Printf("Warning: Failed to insert category %s: %v", cat.name, err)
		}
	}
	fmt.Println("✓ Categories seeded")

	// Read seed file
	seedSQL, err := os.ReadFile("db/seeds/merchant_dictionary_seed.sql")
	if err != nil {
		log.Fatalf("Failed to read seed file: %v", err)
	}

	// Execute entire seed file as one transaction
	txn, err := db.Begin()
	if err != nil {
		log.Fatalf("Failed to begin transaction: %v", err)
	}
	defer txn.Rollback()

	// Remove comments and execute
	lines := strings.Split(string(seedSQL), "\n")
	var currentStmt strings.Builder
	count := 0

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "--") {
			continue
		}

		currentStmt.WriteString(line)
		currentStmt.WriteString("\n")

		// If line ends with semicolon, execute the statement
		if strings.HasSuffix(line, ";") {
			stmt := currentStmt.String()
			if _, err := txn.Exec(stmt); err != nil {
				log.Printf("Warning: Failed to execute: %v\nStatement: %s", err, stmt)
			} else {
				count++
			}
			currentStmt.Reset()
		}
	}

	if err := txn.Commit(); err != nil {
		log.Fatalf("Failed to commit transaction: %v", err)
	}

	fmt.Printf("✓ Seeded %d merchant entries\n", count)
	fmt.Println("✓ Merchant dictionary seeding complete!")
}
