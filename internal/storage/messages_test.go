// ABOUTME: Tests for message CRUD operations
// ABOUTME: Validates message creation, retrieval, and conversation association
package storage_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/harper/clem/internal/core"
	"github.com/harper/clem/internal/storage"
)

func TestCreateMessage(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create conversation first
	conv := &storage.Conversation{ID: "conv-1", Title: "Test", Model: "claude-sonnet-4-5-20250929"}
	require.NoError(t, storage.CreateConversation(db, conv))

	// Create message
	toolCalls := []core.ToolUse{
		{Type: "tool_use", ID: "tool-1", Name: "read", Input: map[string]interface{}{"path": "/foo"}},
	}
	toolCallsJSON, _ := json.Marshal(toolCalls)

	msg := &storage.Message{
		ID:             "msg-1",
		ConversationID: "conv-1",
		Role:           "assistant",
		Content:        "Hello",
		ToolCalls:      string(toolCallsJSON),
	}

	err := storage.CreateMessage(db, msg)
	require.NoError(t, err)

	// Retrieve it
	retrieved, err := storage.GetMessage(db, "msg-1")
	require.NoError(t, err)
	assert.Equal(t, "Hello", retrieved.Content)
	assert.Equal(t, "assistant", retrieved.Role)
}

func TestListMessagesByConversation(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	conv := &storage.Conversation{ID: "conv-1", Title: "Test", Model: "claude-sonnet-4-5-20250929"}
	require.NoError(t, storage.CreateConversation(db, conv))

	// Create messages
	msg1 := &storage.Message{ID: "msg-1", ConversationID: "conv-1", Role: "user", Content: "Hi"}
	msg2 := &storage.Message{ID: "msg-2", ConversationID: "conv-1", Role: "assistant", Content: "Hello"}

	require.NoError(t, storage.CreateMessage(db, msg1))
	require.NoError(t, storage.CreateMessage(db, msg2))

	// List messages
	msgs, err := storage.ListMessages(db, "conv-1")
	require.NoError(t, err)
	assert.Len(t, msgs, 2)
	assert.Equal(t, "msg-1", msgs[0].ID)
	assert.Equal(t, "msg-2", msgs[1].ID)
}

func TestMessageWithMetadata(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	conv := &storage.Conversation{ID: "conv-1", Title: "Test", Model: "claude-sonnet-4-5-20250929"}
	require.NoError(t, storage.CreateConversation(db, conv))

	// Create message with metadata
	metadata := map[string]interface{}{
		"timestamp": "2025-11-26T12:00:00Z",
		"source":    "test",
	}
	metadataJSON, _ := json.Marshal(metadata)

	msg := &storage.Message{
		ID:             "msg-1",
		ConversationID: "conv-1",
		Role:           "user",
		Content:        "Test message",
		Metadata:       string(metadataJSON),
	}

	err := storage.CreateMessage(db, msg)
	require.NoError(t, err)

	// Retrieve and verify metadata
	retrieved, err := storage.GetMessage(db, "msg-1")
	require.NoError(t, err)
	assert.NotEmpty(t, retrieved.Metadata)

	var retrievedMetadata map[string]interface{}
	err = json.Unmarshal([]byte(retrieved.Metadata), &retrievedMetadata)
	require.NoError(t, err)
	assert.Equal(t, "test", retrievedMetadata["source"])
}
