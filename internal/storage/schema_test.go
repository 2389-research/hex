// ABOUTME: Tests for SQLite storage schema and migrations
// ABOUTME: Validates database structure, indexes, and constraints
package storage_test

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/harper/clem/internal/storage"
)

func TestInitializeSchema(t *testing.T) {
	// Use in-memory SQLite for tests
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	// Initialize schema
	err = storage.InitializeSchema(db)
	require.NoError(t, err)

	// Verify conversations table exists
	var tableName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='conversations'").Scan(&tableName)
	require.NoError(t, err)
	assert.Equal(t, "conversations", tableName)

	// Verify messages table exists
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='messages'").Scan(&tableName)
	require.NoError(t, err)
	assert.Equal(t, "messages", tableName)
}

func TestSchemaIndexes(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	err = storage.InitializeSchema(db)
	require.NoError(t, err)

	// Verify index on messages(conversation_id)
	var indexName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='index' AND name='idx_messages_conversation'").Scan(&indexName)
	require.NoError(t, err)
	assert.Equal(t, "idx_messages_conversation", indexName)
}
