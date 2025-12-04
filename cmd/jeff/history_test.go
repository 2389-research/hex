// ABOUTME: Tests for history command functionality
// ABOUTME: Tests command parsing, output formatting, and integration with storage

package main

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/harper/pagent/internal/storage"
)

func TestHistoryCommand_NoHistory(t *testing.T) {
	// Create temporary database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Open database
	db, err := storage.OpenDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Set up command with output capture
	rootCmd.SetArgs([]string{"history", "--db-path", dbPath})
	var outBuf bytes.Buffer
	rootCmd.SetOut(&outBuf)
	rootCmd.SetErr(&outBuf)

	// Execute command
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Failed to execute command: %v", err)
	}

	// Check output
	output := outBuf.String()
	if !strings.Contains(output, "No history found") {
		t.Errorf("Expected 'No history found' message, got: %s", output)
	}
}

func TestHistoryCommand_WithHistory(t *testing.T) {
	// Create temporary database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Open database
	db, err := storage.OpenDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Create test conversation (required for foreign key)
	conv := &storage.Conversation{
		ID:    "conv-1",
		Title: "Test Conversation",
		Model: "claude-test",
	}
	if err := storage.CreateConversation(db, conv); err != nil {
		t.Fatalf("Failed to create conversation: %v", err)
	}

	// Add test history entries
	entries := []*storage.HistoryEntry{
		{
			ID:                "hist-1",
			ConversationID:    "conv-1",
			UserMessage:       "How do I use docker?",
			AssistantResponse: "Docker is a containerization platform...",
			CreatedAt:         time.Now().Add(-2 * time.Hour),
		},
		{
			ID:                "hist-2",
			ConversationID:    "conv-1",
			UserMessage:       "What about docker-compose?",
			AssistantResponse: "Docker Compose is a tool for defining...",
			CreatedAt:         time.Now().Add(-1 * time.Hour),
		},
	}

	for _, entry := range entries {
		if err := storage.AddHistoryEntry(db, entry); err != nil {
			t.Fatalf("Failed to add history entry: %v", err)
		}
	}

	// Set up command with output capture
	rootCmd.SetArgs([]string{"history", "--db-path", dbPath})
	var outBuf bytes.Buffer
	rootCmd.SetOut(&outBuf)
	rootCmd.SetErr(&outBuf)

	// Execute command
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Failed to execute command: %v", err)
	}

	// Check output
	output := outBuf.String()
	if !strings.Contains(output, "Recent history") {
		t.Errorf("Expected 'Recent history' header, got: %s", output)
	}
	if !strings.Contains(output, "How do I use docker?") {
		t.Errorf("Expected first message in output, got: %s", output)
	}
	if !strings.Contains(output, "What about docker-compose?") {
		t.Errorf("Expected second message in output, got: %s", output)
	}
	if !strings.Contains(output, "conv-1") {
		t.Errorf("Expected conversation ID in output, got: %s", output)
	}
}

func TestHistoryCommand_CustomLimit(t *testing.T) {
	// Create temporary database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Open database
	db, err := storage.OpenDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Create test conversation
	conv := &storage.Conversation{
		ID:    "conv-1",
		Title: "Test Conversation",
		Model: "claude-test",
	}
	if err := storage.CreateConversation(db, conv); err != nil {
		t.Fatalf("Failed to create conversation: %v", err)
	}

	// Add many test entries
	for i := 0; i < 50; i++ {
		entry := &storage.HistoryEntry{
			ID:                fmt.Sprintf("hist-%d", i),
			ConversationID:    "conv-1",
			UserMessage:       fmt.Sprintf("Test message %d", i),
			AssistantResponse: "Response",
			CreatedAt:         time.Now().Add(-time.Duration(i) * time.Hour),
		}
		if err := storage.AddHistoryEntry(db, entry); err != nil {
			t.Fatalf("Failed to add history entry: %v", err)
		}
	}

	// Test default limit (20)
	rootCmd.SetArgs([]string{"history", "--db-path", dbPath})
	var outBuf bytes.Buffer
	rootCmd.SetOut(&outBuf)
	rootCmd.SetErr(&outBuf)

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Failed to execute command: %v", err)
	}

	output := outBuf.String()
	if !strings.Contains(output, "showing 20") {
		t.Errorf("Expected default limit of 20, got: %s", output)
	}

	// Test custom limit
	rootCmd.SetArgs([]string{"history", "--limit", "5", "--db-path", dbPath})
	outBuf.Reset()
	rootCmd.SetOut(&outBuf)
	rootCmd.SetErr(&outBuf)

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Failed to execute command: %v", err)
	}

	output = outBuf.String()
	if !strings.Contains(output, "showing 5") {
		t.Errorf("Expected limit of 5, got: %s", output)
	}
}

func TestHistorySearchCommand_NoResults(t *testing.T) {
	// Create temporary database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Open database
	db, err := storage.OpenDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Create test conversation
	conv := &storage.Conversation{
		ID:    "conv-1",
		Title: "Test Conversation",
		Model: "claude-test",
	}
	if err := storage.CreateConversation(db, conv); err != nil {
		t.Fatalf("Failed to create conversation: %v", err)
	}

	// Add test history entry
	entry := &storage.HistoryEntry{
		ID:                "hist-1",
		ConversationID:    "conv-1",
		UserMessage:       "How do I use docker?",
		AssistantResponse: "Docker is a containerization platform...",
		CreatedAt:         time.Now(),
	}
	if err := storage.AddHistoryEntry(db, entry); err != nil {
		t.Fatalf("Failed to add history entry: %v", err)
	}

	// Search for non-existent term
	rootCmd.SetArgs([]string{"history", "search", "kubernetes", "--db-path", dbPath})
	var outBuf bytes.Buffer
	rootCmd.SetOut(&outBuf)
	rootCmd.SetErr(&outBuf)

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Failed to execute command: %v", err)
	}

	output := outBuf.String()
	if !strings.Contains(output, "No results found") {
		t.Errorf("Expected 'No results found' message, got: %s", output)
	}
}

func TestHistorySearchCommand_WithResults(t *testing.T) {
	// Create temporary database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Open database
	db, err := storage.OpenDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Create test conversations
	for i := 1; i <= 2; i++ {
		conv := &storage.Conversation{
			ID:    fmt.Sprintf("conv-%d", i),
			Title: "Test Conversation",
			Model: "claude-test",
		}
		if err := storage.CreateConversation(db, conv); err != nil {
			t.Fatalf("Failed to create conversation: %v", err)
		}
	}

	// Add test history entries
	entries := []*storage.HistoryEntry{
		{
			ID:                "hist-1",
			ConversationID:    "conv-1",
			UserMessage:       "How do I use docker?",
			AssistantResponse: "Docker is a containerization platform...",
			CreatedAt:         time.Now().Add(-2 * time.Hour),
		},
		{
			ID:                "hist-2",
			ConversationID:    "conv-1",
			UserMessage:       "What about kubernetes?",
			AssistantResponse: "Kubernetes is an orchestration system...",
			CreatedAt:         time.Now().Add(-1 * time.Hour),
		},
		{
			ID:                "hist-3",
			ConversationID:    "conv-2",
			UserMessage:       "Explain docker-compose",
			AssistantResponse: "Docker Compose is a tool...",
			CreatedAt:         time.Now(),
		},
	}

	for _, entry := range entries {
		if err := storage.AddHistoryEntry(db, entry); err != nil {
			t.Fatalf("Failed to add history entry: %v", err)
		}
	}

	// Search for "docker"
	rootCmd.SetArgs([]string{"history", "search", "docker", "--db-path", dbPath})
	var outBuf bytes.Buffer
	rootCmd.SetOut(&outBuf)
	rootCmd.SetErr(&outBuf)

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Failed to execute command: %v", err)
	}

	output := outBuf.String()
	if !strings.Contains(output, "Search results for \"docker\"") {
		t.Errorf("Expected search results header, got: %s", output)
	}
	if !strings.Contains(output, "How do I use docker?") {
		t.Errorf("Expected first docker message in output, got: %s", output)
	}
	if !strings.Contains(output, "Explain docker-compose") {
		t.Errorf("Expected second docker message in output, got: %s", output)
	}
	if strings.Contains(output, "kubernetes") {
		t.Errorf("Expected kubernetes message to be excluded, got: %s", output)
	}
}

// Note: TestFormatRelativeTime is defined in favorites_test.go since it's shared

func TestTruncateString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "short string - no truncation",
			input:    "Hello world",
			maxLen:   60,
			expected: "Hello world",
		},
		{
			name:     "exact length - no truncation",
			input:    "This is exactly sixty characters long for testing purposes!!",
			maxLen:   60,
			expected: "This is exactly sixty characters long for testing purposes!!",
		},
		{
			name:     "long string - truncate at word",
			input:    "This is a very long message that needs to be truncated at a reasonable word boundary",
			maxLen:   60,
			expected: "This is a very long message that needs to be truncated...",
		},
		{
			name:     "long string - no word boundary",
			input:    "ThisIsAVeryLongWordWithNoSpacesThatNeedsToBeTruncatedSomewhere",
			maxLen:   20,
			expected: "ThisIsAVeryLongWo...",
		},
		{
			name:     "newlines normalized",
			input:    "First line\nSecond line\nThird line",
			maxLen:   60,
			expected: "First line Second line Third line",
		},
		{
			name:     "tabs normalized",
			input:    "First\tSecond\tThird",
			maxLen:   60,
			expected: "First Second Third",
		},
		{
			name:     "multiple spaces collapsed",
			input:    "Too    many     spaces",
			maxLen:   60,
			expected: "Too many spaces",
		},
		{
			name:     "mixed whitespace",
			input:    "Mixed\n\twhitespace   test",
			maxLen:   60,
			expected: "Mixed whitespace test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateString(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("truncateString(%q, %d) = %q, want %q", tt.input, tt.maxLen, result, tt.expected)
			}
		})
	}
}

func TestDisplayHistoryEntry(t *testing.T) {
	entry := &storage.HistoryEntry{
		ID:                "hist-1",
		ConversationID:    "conv-123",
		UserMessage:       "How do I use docker?",
		AssistantResponse: "Docker is a containerization platform...",
		CreatedAt:         time.Now().Add(-2 * time.Hour),
	}

	var outBuf bytes.Buffer
	cmd := rootCmd
	cmd.SetOut(&outBuf)

	displayHistoryEntry(cmd, entry)

	output := outBuf.String()
	if !strings.Contains(output, "How do I use docker?") {
		t.Errorf("Expected message in output, got: %s", output)
	}
	if !strings.Contains(output, "conv-123") {
		t.Errorf("Expected conversation ID in output, got: %s", output)
	}
	if !strings.Contains(output, "2 hours ago") {
		t.Errorf("Expected relative time in output, got: %s", output)
	}
}
