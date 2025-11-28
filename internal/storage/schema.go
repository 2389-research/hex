// ABOUTME: SQLite schema initialization and migration management
// ABOUTME: Handles database setup, table creation, and version tracking
package storage

import (
	"database/sql"
	"embed"
	"fmt"

	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrations embed.FS

// InitializeSchema creates tables and indexes
func InitializeSchema(db *sql.DB) error {
	// Read migration file
	migrationSQL, err := migrations.ReadFile("migrations/001_initial.sql")
	if err != nil {
		return fmt.Errorf("read migration: %w", err)
	}

	// Execute migration
	if _, err := db.Exec(string(migrationSQL)); err != nil {
		return fmt.Errorf("execute migration: %w", err)
	}

	return nil
}

// OpenDatabase opens SQLite database at given path
func OpenDatabase(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}

	// Enable WAL mode for better concurrency
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return nil, fmt.Errorf("enable WAL mode: %w", err)
	}

	// Initialize schema
	if err := InitializeSchema(db); err != nil {
		return nil, fmt.Errorf("initialize schema: %w", err)
	}

	return db, nil
}
