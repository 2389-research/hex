// ABOUTME: Helper functions and utilities for integration tests
// ABOUTME: Provides test database setup, fixtures, and mock clients

package integration

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/harper/clem/internal/core"
	"github.com/harper/clem/internal/storage"
	"github.com/stretchr/testify/require"
)

// SetupTestDB creates a temporary test database
func SetupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	// Create temporary directory
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Open database
	db, err := storage.OpenDatabase(dbPath)
	require.NoError(t, err, "failed to open test database")

	// Cleanup on test end
	t.Cleanup(func() {
		_ = db.Close()
	})

	return db
}

// CreateTestConversation creates a conversation with test data
func CreateTestConversation(t *testing.T, db *sql.DB, model string) string {
	t.Helper()

	conversationID := fmt.Sprintf("conv-%s", uuid.New().String()[:8])

	conv := &storage.Conversation{
		ID:           conversationID,
		Title:        "Test Conversation",
		Model:        model,
		SystemPrompt: "",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := storage.CreateConversation(db, conv)
	require.NoError(t, err, "failed to create test conversation")

	return conversationID
}

// CreateTestMessage creates a message in the database
func CreateTestMessage(t *testing.T, db *sql.DB, conversationID, role, content string) string {
	t.Helper()

	messageID := fmt.Sprintf("msg-%s", uuid.New().String()[:8])

	msg := &storage.Message{
		ID:             messageID,
		ConversationID: conversationID,
		Role:           role,
		Content:        content,
		CreatedAt:      time.Now(),
	}

	err := storage.CreateMessage(db, msg)
	require.NoError(t, err, "failed to create test message")

	return messageID
}

// CreateTestFile creates a temporary test file with content
func CreateTestFile(t *testing.T, content string) string {
	t.Helper()

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "testfile.txt")

	err := os.WriteFile(filePath, []byte(content), 0600)
	require.NoError(t, err, "failed to create test file")

	return filePath
}

// CreateTestDir creates a temporary test directory
func CreateTestDir(t *testing.T) string {
	t.Helper()
	return t.TempDir()
}

// MockAPIClient returns a mock API client for testing
// This is a placeholder - actual implementation depends on whether you have mocking infrastructure
func MockAPIClient(t *testing.T, _ []core.MessageResponse) *core.Client {
	t.Helper()

	// For now, just create a real client
	// In production tests, you'd use a mock HTTP server or dependency injection
	client := core.NewClient("test-api-key-mock")
	return client
}

// AssertFileExists checks that a file exists and has expected content
func AssertFileExists(t *testing.T, path string) {
	t.Helper()

	_, err := os.Stat(path)
	require.NoError(t, err, "file should exist at path: %s", path)
}

// AssertFileContains checks that a file contains expected content
func AssertFileContains(t *testing.T, path, expectedContent string) {
	t.Helper()

	content, err := os.ReadFile(path) //nolint:gosec // G304: Path validated by caller
	require.NoError(t, err, "failed to read file: %s", path)
	require.Contains(t, string(content), expectedContent, "file should contain expected content")
}

// WaitForCondition polls a condition function until it returns true or timeout
func WaitForCondition(t *testing.T, timeout time.Duration, interval time.Duration, condition func() bool, message string) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(interval)
	}

	t.Fatalf("Condition not met within timeout: %s", message)
}
