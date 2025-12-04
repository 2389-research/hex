// Package storage provides database operations for conversations, messages, and metadata.
// ABOUTME: SQLite schema initialization and migration management
// ABOUTME: Handles database setup, table creation, and version tracking
package storage

import (
	"database/sql"
	"embed"
	"fmt"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	// Import SQLite driver for database/sql
	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrations embed.FS

// RunMigrations executes database migrations using golang-migrate
func RunMigrations(db *sql.DB) error {
	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	sourceDriver, err := iofs.New(migrations, "migrations")
	if err != nil {
		return fmt.Errorf("failed to create migration source: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", sourceDriver, "sqlite3", driver)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration failed: %w", err)
	}

	return nil
}

// InitializeSchema creates tables and indexes
func InitializeSchema(db *sql.DB) error {
	// List of migrations to run in order
	migrationFiles := []string{
		"migrations/001_initial.sql",
		"migrations/002_todos.sql",
		"migrations/003_history.sql",
		"migrations/004_favorites.sql",
	}

	// Execute each migration
	for _, filename := range migrationFiles {
		migrationSQL, err := migrations.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", filename, err)
		}

		if _, err := db.Exec(string(migrationSQL)); err != nil {
			// Ignore "duplicate column" errors (happens when migrations run multiple times)
			// This is a workaround until proper migration tracking is implemented
			errStr := err.Error()
			isDuplicateColumn := strings.Contains(errStr, "duplicate column")
			if !isDuplicateColumn {
				return fmt.Errorf("execute migration %s: %w", filename, err)
			}
			// Silently ignore duplicate column errors
		}
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

	// Run migrations
	if err := RunMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}
