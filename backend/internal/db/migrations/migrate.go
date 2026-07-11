package migrations

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"sort"
	"strings"
)

//go:embed *.sql
var migrationFiles embed.FS

// Migration represents a single migration
type Migration struct {
	Name    string
	Content string
}

// RunMigrations executes all pending migrations
// This is idempotent and safe to call multiple times
func RunMigrations(db *sql.DB) error {
	log.Println("Starting database migrations...")

	// Create migrations table if it doesn't exist
	if err := createMigrationsTable(db); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Read all migration files
	migrations, err := readMigrations()
	if err != nil {
		return fmt.Errorf("failed to read migrations: %w", err)
	}

	// Sort by name to ensure order
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Name < migrations[j].Name
	})

	// Run each migration that hasn't been run yet
	for _, migration := range migrations {
		applied, err := isMigrationApplied(db, migration.Name)
		if err != nil {
			return fmt.Errorf("failed to check migration status: %w", err)
		}

		if applied {
			log.Printf("✓ Migration already applied: %s", migration.Name)
			continue
		}

		log.Printf("Running migration: %s", migration.Name)

		// Execute migration
		if _, err := db.Exec(migration.Content); err != nil {
			return fmt.Errorf("migration %s failed: %w", migration.Name, err)
		}

		// Record migration as applied
		if err := recordMigration(db, migration.Name); err != nil {
			return fmt.Errorf("failed to record migration: %w", err)
		}

		log.Printf("✓ Migration applied: %s", migration.Name)
	}

	log.Println("✓ All migrations completed successfully")
	return nil
}

func createMigrationsTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) UNIQUE NOT NULL,
		applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err := db.Exec(query)
	return err
}

func readMigrations() ([]Migration, error) {
	entries, err := migrationFiles.ReadDir(".")
	if err != nil {
		return nil, err
	}

	var migrations []Migration
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		// Skip migrate.go itself
		if entry.Name() == "migrate.go" {
			continue
		}

		content, err := migrationFiles.ReadFile(entry.Name())
		if err != nil {
			return nil, fmt.Errorf("failed to read migration file %s: %w", entry.Name(), err)
		}

		migrations = append(migrations, Migration{
			Name:    entry.Name(),
			Content: string(content),
		})
	}

	return migrations, nil
}

func isMigrationApplied(db *sql.DB, name string) (bool, error) {
	var count int
	err := db.QueryRow(
		"SELECT COUNT(*) FROM schema_migrations WHERE name = $1",
		name,
	).Scan(&count)

	if err != nil && err != sql.ErrNoRows {
		return false, err
	}

	return count > 0, nil
}

func recordMigration(db *sql.DB, name string) error {
	_, err := db.Exec(
		"INSERT INTO schema_migrations (name) VALUES ($1)",
		name,
	)
	return err
}
