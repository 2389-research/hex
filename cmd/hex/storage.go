// ABOUTME: Database initialization and management for CLI
// ABOUTME: Handles database opening, schema setup, and conversation loading
package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/harper/hex/internal/storage"
)

// defaultDBPath returns the default database path (~/.hex/hex.db)
func defaultDBPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory if home not found
		return filepath.Join(".", ".hex", "hex.db")
	}
	return filepath.Join(home, ".hex", "hex.db")
}

// openDatabase opens the database at the given path, creating directories and schema as needed
func openDatabase(path string) (*sql.DB, error) {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return nil, fmt.Errorf("create db directory: %w", err)
	}

	// Open database using storage.OpenDatabase which handles schema initialization
	db, err := storage.OpenDatabase(path)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	return db, nil
}

// loadConversationHistory loads a conversation and its messages from the database
func loadConversationHistory(db *sql.DB, convID string) (*storage.Conversation, []*storage.Message, error) {
	// Get conversation
	conv, err := storage.GetConversation(db, convID)
	if err != nil {
		return nil, nil, fmt.Errorf("get conversation: %w", err)
	}

	// Get messages
	messages, err := storage.ListMessages(db, convID)
	if err != nil {
		return nil, nil, fmt.Errorf("get messages: %w", err)
	}

	return conv, messages, nil
}
