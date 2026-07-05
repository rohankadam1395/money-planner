package db

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

type Database struct {
	conn *sql.DB
}

func New(connectionString string) (*Database, error) {
	conn, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	err = conn.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Database{conn: conn}, nil
}

func (db *Database) Close() error {
	return db.conn.Close()
}

func (db *Database) GetConnection() *sql.DB {
	return db.conn
}
