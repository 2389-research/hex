// ABOUTME: Tests for SQLite storage schema and migrations
// ABOUTME: Validates database structure, indexes, and constraints
package storage_test

import (
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/2389-research/hex/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitializeSchema(t *testing.T) {
	// Use in-memory SQLite for tests
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

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
	defer func() { _ = db.Close() }()

	err = storage.InitializeSchema(db)
	require.NoError(t, err)

	// Verify index on messages(conversation_id)
	var indexName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='index' AND name='idx_messages_conversation'").Scan(&indexName)
	require.NoError(t, err)
	assert.Equal(t, "idx_messages_conversation", indexName)
}

func TestMigrations(t *testing.T) {
	// Create temporary database
	tmpDB := filepath.Join(t.TempDir(), "test.db")

	db, err := storage.OpenDatabase(tmpDB)
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	// Verify tables exist
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table'").Scan(&count)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, count, 2) // At least conversations and messages

	// Verify new columns exist in conversations
	rows, err := db.Query("PRAGMA table_info(conversations)")
	require.NoError(t, err)
	defer func() { _ = rows.Close() }()

	columns := make(map[string]bool)
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dfltValue sql.NullString
		require.NoError(t, rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk))
		columns[name] = true
	}

	assert.True(t, columns["prompt_tokens"], "missing prompt_tokens column")
	assert.True(t, columns["completion_tokens"], "missing completion_tokens column")
	assert.True(t, columns["total_cost"], "missing total_cost column")
}
