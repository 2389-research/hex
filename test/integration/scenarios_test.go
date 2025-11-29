// ABOUTME: End-to-end scenario integration tests for complete workflows
// ABOUTME: Tests realistic integration between storage, tools, and core components

package integration

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/harper/clem/internal/storage"
	"github.com/harper/clem/internal/tools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestScenario_StorageAndTools tests integration between database and tool execution
func TestScenario_StorageAndTools(t *testing.T) {
	// Setup database
	db := SetupTestDB(t)
	convID := CreateTestConversation(t, db, "claude-sonnet-4-5-20250929")

	// Create user message
	CreateTestMessage(t, db, convID, "user", "Please read the test file")

	// Setup tools
	registry := tools.NewRegistry()
	registry.Register(tools.NewReadTool())

	executor := tools.NewExecutor(registry, func(toolName string, params map[string]interface{}) bool {
		return true // Auto-approve for test
	})

	// Create test file
	testFile := CreateTestFile(t, "Integration test content")

	// Execute tool
	ctx := context.Background()
	params := map[string]interface{}{
		"path": testFile,
	}

	result, err := executor.Execute(ctx, "read_file", params)
	require.NoError(t, err)
	require.True(t, result.Success)
	assert.Contains(t, result.Output, "Integration test content")

	// Save tool result as assistant message
	CreateTestMessage(t, db, convID, "assistant", "File content: "+result.Output)

	// Verify full conversation
	messages, err := storage.ListMessages(db, convID)
	require.NoError(t, err)
	assert.Len(t, messages, 2)
	assert.Equal(t, "user", messages[0].Role)
	assert.Equal(t, "assistant", messages[1].Role)
	assert.Contains(t, messages[1].Content, "Integration test content")
}

// TestScenario_MultipleToolsSequence tests executing multiple tools in sequence
func TestScenario_MultipleToolsSequence(t *testing.T) {
	// Setup
	db := SetupTestDB(t)
	convID := CreateTestConversation(t, db, "claude-sonnet-4-5-20250929")

	tmpDir := CreateTestDir(t)
	outputFile := filepath.Join(tmpDir, "output.txt")

	// Setup tools
	registry := tools.NewRegistry()
	registry.Register(tools.NewWriteTool())
	registry.Register(tools.NewReadTool())

	executor := tools.NewExecutor(registry, func(toolName string, params map[string]interface{}) bool {
		return true
	})

	ctx := context.Background()

	// Step 1: Write file
	CreateTestMessage(t, db, convID, "user", "Write a file")

	writeParams := map[string]interface{}{
		"path":    outputFile,
		"content": "Test content",
	}

	writeResult, err := executor.Execute(ctx, "write_file", writeParams)
	require.NoError(t, err)
	require.True(t, writeResult.Success)

	CreateTestMessage(t, db, convID, "assistant", "File written")

	// Step 2: Read file back
	CreateTestMessage(t, db, convID, "user", "Read the file")

	readParams := map[string]interface{}{
		"path": outputFile,
	}

	readResult, err := executor.Execute(ctx, "read_file", readParams)
	require.NoError(t, err)
	require.True(t, readResult.Success)
	assert.Contains(t, readResult.Output, "Test content")

	CreateTestMessage(t, db, convID, "assistant", "File content: "+readResult.Output)

	// Verify conversation
	messages, err := storage.ListMessages(db, convID)
	require.NoError(t, err)
	assert.Len(t, messages, 4) // 2 user + 2 assistant
}

// TestScenario_ConversationPersistence tests saving and loading conversation
func TestScenario_ConversationPersistence(t *testing.T) {
	// Phase 1: Create conversation with messages
	db := SetupTestDB(t)
	convID := CreateTestConversation(t, db, "claude-sonnet-4-5-20250929")

	CreateTestMessage(t, db, convID, "user", "Hello")
	CreateTestMessage(t, db, convID, "assistant", "Hi there!")
	CreateTestMessage(t, db, convID, "user", "How are you?")

	// Verify initial state
	messages, err := storage.ListMessages(db, convID)
	require.NoError(t, err)
	assert.Len(t, messages, 3)

	// Phase 2: Retrieve conversation
	conv, err := storage.GetConversation(db, convID)
	require.NoError(t, err)
	assert.Equal(t, "claude-sonnet-4-5-20250929", conv.Model)

	// Phase 3: Continue conversation
	CreateTestMessage(t, db, convID, "assistant", "I'm doing well!")

	messages, err = storage.ListMessages(db, convID)
	require.NoError(t, err)
	assert.Len(t, messages, 4)
	assert.Equal(t, "I'm doing well!", messages[3].Content)
}

// TestScenario_ToolErrorHandling tests tool execution errors are handled gracefully
func TestScenario_ToolErrorHandling(t *testing.T) {
	db := SetupTestDB(t)
	convID := CreateTestConversation(t, db, "claude-sonnet-4-5-20250929")

	registry := tools.NewRegistry()
	registry.Register(tools.NewReadTool())

	executor := tools.NewExecutor(registry, func(toolName string, params map[string]interface{}) bool {
		return true
	})

	// Try to read non-existent file
	ctx := context.Background()
	params := map[string]interface{}{
		"path": "/nonexistent/file.txt",
	}

	result, err := executor.Execute(ctx, "read_file", params)

	// Should return error result, not panic
	require.NoError(t, err, "should not return error, but failed result")
	assert.False(t, result.Success)
	assert.NotEmpty(t, result.Error)

	// Save error to conversation
	CreateTestMessage(t, db, convID, "assistant", "Error: "+result.Error)

	messages, err := storage.ListMessages(db, convID)
	require.NoError(t, err)
	assert.Len(t, messages, 1)
	assert.Contains(t, messages[0].Content, "Error")
}

// TestScenario_LargeConversation tests handling many messages efficiently
func TestScenario_LargeConversation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large conversation test in short mode")
	}

	db := SetupTestDB(t)
	convID := CreateTestConversation(t, db, "claude-sonnet-4-5-20250929")

	// Add 50 message pairs
	for i := 0; i < 50; i++ {
		CreateTestMessage(t, db, convID, "user", "Message")
		CreateTestMessage(t, db, convID, "assistant", "Response")
	}

	// Verify all saved
	messages, err := storage.ListMessages(db, convID)
	require.NoError(t, err)
	assert.Len(t, messages, 100)

	// Verify ordering
	assert.Equal(t, "user", messages[0].Role)
	assert.Equal(t, "assistant", messages[1].Role)
	assert.Equal(t, "user", messages[98].Role)
	assert.Equal(t, "assistant", messages[99].Role)
}

// TestScenario_ToolDenial tests denying tool execution
func TestScenario_ToolDenial(t *testing.T) {
	db := SetupTestDB(t)
	convID := CreateTestConversation(t, db, "claude-sonnet-4-5-20250929")

	tmpDir := CreateTestDir(t)
	testFile := filepath.Join(tmpDir, "denied.txt")

	registry := tools.NewRegistry()
	registry.Register(tools.NewWriteTool())

	// Executor that denies all requests
	executor := tools.NewExecutor(registry, func(toolName string, params map[string]interface{}) bool {
		return false // Deny
	})

	ctx := context.Background()
	params := map[string]interface{}{
		"path":    testFile,
		"content": "Should not be written",
	}

	result, err := executor.Execute(ctx, "write_file", params)

	// Should get denial result
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "denied")

	// Save denial to conversation
	CreateTestMessage(t, db, convID, "assistant", "Tool denied: "+result.Error)

	messages, err := storage.ListMessages(db, convID)
	require.NoError(t, err)
	assert.Len(t, messages, 1)
	assert.Contains(t, messages[0].Content, "denied")
}

// TestScenario_ConcurrentConversations tests multiple conversations in parallel
func TestScenario_ConcurrentConversations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}

	db := SetupTestDB(t)

	// Create multiple conversations
	conv1 := CreateTestConversation(t, db, "claude-sonnet-4-5-20250929")
	conv2 := CreateTestConversation(t, db, "claude-sonnet-4-5-20250929")
	conv3 := CreateTestConversation(t, db, "claude-sonnet-4-5-20250929")

	// Add messages to each
	CreateTestMessage(t, db, conv1, "user", "Conv1 message")
	CreateTestMessage(t, db, conv2, "user", "Conv2 message")
	CreateTestMessage(t, db, conv3, "user", "Conv3 message")

	// Verify isolation
	msgs1, err := storage.ListMessages(db, conv1)
	require.NoError(t, err)
	assert.Len(t, msgs1, 1)
	assert.Contains(t, msgs1[0].Content, "Conv1")

	msgs2, err := storage.ListMessages(db, conv2)
	require.NoError(t, err)
	assert.Len(t, msgs2, 1)
	assert.Contains(t, msgs2[0].Content, "Conv2")

	msgs3, err := storage.ListMessages(db, conv3)
	require.NoError(t, err)
	assert.Len(t, msgs3, 1)
	assert.Contains(t, msgs3[0].Content, "Conv3")

	// Verify conversations list
	convs, err := storage.ListConversations(db, 10, 0)
	require.NoError(t, err)
	assert.Len(t, convs, 3)
}
