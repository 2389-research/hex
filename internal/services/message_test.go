// ABOUTME: Test suite for MessageService implementation
// ABOUTME: Tests message storage, retrieval, filtering, and event publishing

package services

import (
	"context"
	"testing"
	"time"

	"github.com/2389-research/hex/internal/pubsub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMessageService_Add(t *testing.T) {
	db := setupTestDB(t)
	convSvc := NewConversationService(db)
	msgSvc := NewMessageService(db)

	// Create a conversation first
	conv, err := convSvc.Create(context.Background(), "Test Conversation")
	require.NoError(t, err)

	// Add a message
	msg := &Message{
		ConversationID: conv.ID,
		Role:           "user",
		Content:        "Hello, world!",
		Provider:       "anthropic",
		Model:          "claude-3-5-sonnet-20241022",
		IsSummary:      false,
	}

	err = msgSvc.Add(context.Background(), msg)
	require.NoError(t, err)
	assert.NotEmpty(t, msg.ID, "ID should be generated")
	assert.NotZero(t, msg.CreatedAt, "CreatedAt should be set")
}

func TestMessageService_AddPublishesEvent(t *testing.T) {
	db := setupTestDB(t)
	convSvc := NewConversationService(db)
	msgSvc := NewMessageService(db)

	// Create a conversation
	conv, err := convSvc.Create(context.Background(), "Test Conversation")
	require.NoError(t, err)

	// Subscribe to message events
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	events := msgSvc.Subscribe(ctx)

	// Add message in goroutine
	go func() {
		msg := &Message{
			ConversationID: conv.ID,
			Role:           "assistant",
			Content:        "Hi there!",
			Provider:       "anthropic",
			Model:          "claude-3-5-sonnet-20241022",
			IsSummary:      false,
		}
		_ = msgSvc.Add(context.Background(), msg)
	}()

	// Wait for event
	select {
	case event := <-events:
		assert.Equal(t, pubsub.Created, event.Type)
		assert.Equal(t, "Hi there!", event.Payload.Content)
		assert.Equal(t, "assistant", event.Payload.Role)
	case <-ctx.Done():
		t.Fatal("timeout waiting for event")
	}
}

func TestMessageService_GetByConversation(t *testing.T) {
	db := setupTestDB(t)
	convSvc := NewConversationService(db)
	msgSvc := NewMessageService(db)

	// Create a conversation
	conv, err := convSvc.Create(context.Background(), "Test Conversation")
	require.NoError(t, err)

	// Add multiple messages with delays to ensure ordering
	msg1 := &Message{
		ConversationID: conv.ID,
		Role:           "user",
		Content:        "First message",
		Provider:       "anthropic",
		Model:          "claude-3-5-sonnet-20241022",
		IsSummary:      false,
	}
	err = msgSvc.Add(context.Background(), msg1)
	require.NoError(t, err)

	time.Sleep(10 * time.Millisecond)

	msg2 := &Message{
		ConversationID: conv.ID,
		Role:           "assistant",
		Content:        "Second message",
		Provider:       "anthropic",
		Model:          "claude-3-5-sonnet-20241022",
		IsSummary:      false,
	}
	err = msgSvc.Add(context.Background(), msg2)
	require.NoError(t, err)

	time.Sleep(10 * time.Millisecond)

	msg3 := &Message{
		ConversationID: conv.ID,
		Role:           "user",
		Content:        "Third message",
		Provider:       "anthropic",
		Model:          "claude-3-5-sonnet-20241022",
		IsSummary:      false,
	}
	err = msgSvc.Add(context.Background(), msg3)
	require.NoError(t, err)

	// Get all messages
	messages, err := msgSvc.GetByConversation(context.Background(), conv.ID)
	require.NoError(t, err)
	assert.Len(t, messages, 3)

	// Should be ordered by created_at ASC (oldest first)
	assert.Equal(t, "First message", messages[0].Content)
	assert.Equal(t, "Second message", messages[1].Content)
	assert.Equal(t, "Third message", messages[2].Content)
}

func TestMessageService_GetByConversationEmpty(t *testing.T) {
	db := setupTestDB(t)
	convSvc := NewConversationService(db)
	msgSvc := NewMessageService(db)

	// Create a conversation
	conv, err := convSvc.Create(context.Background(), "Empty Conversation")
	require.NoError(t, err)

	// Get messages (should be empty)
	messages, err := msgSvc.GetByConversation(context.Background(), conv.ID)
	require.NoError(t, err)
	assert.Empty(t, messages)
}

func TestMessageService_GetSummaries(t *testing.T) {
	db := setupTestDB(t)
	convSvc := NewConversationService(db)
	msgSvc := NewMessageService(db)

	// Create a conversation
	conv, err := convSvc.Create(context.Background(), "Test Conversation")
	require.NoError(t, err)

	// Add regular message
	regularMsg := &Message{
		ConversationID: conv.ID,
		Role:           "user",
		Content:        "Regular message",
		Provider:       "anthropic",
		Model:          "claude-3-5-sonnet-20241022",
		IsSummary:      false,
	}
	err = msgSvc.Add(context.Background(), regularMsg)
	require.NoError(t, err)

	time.Sleep(10 * time.Millisecond)

	// Add summary message
	summaryMsg := &Message{
		ConversationID: conv.ID,
		Role:           "assistant",
		Content:        "This is a summary",
		Provider:       "anthropic",
		Model:          "claude-3-5-sonnet-20241022",
		IsSummary:      true,
	}
	err = msgSvc.Add(context.Background(), summaryMsg)
	require.NoError(t, err)

	time.Sleep(10 * time.Millisecond)

	// Add another regular message
	regularMsg2 := &Message{
		ConversationID: conv.ID,
		Role:           "user",
		Content:        "Another regular message",
		Provider:       "anthropic",
		Model:          "claude-3-5-sonnet-20241022",
		IsSummary:      false,
	}
	err = msgSvc.Add(context.Background(), regularMsg2)
	require.NoError(t, err)

	// Get only summaries
	summaries, err := msgSvc.GetSummaries(context.Background(), conv.ID)
	require.NoError(t, err)
	assert.Len(t, summaries, 1)
	assert.Equal(t, "This is a summary", summaries[0].Content)
	assert.True(t, summaries[0].IsSummary)
}

func TestMessageService_GetSummariesEmpty(t *testing.T) {
	db := setupTestDB(t)
	convSvc := NewConversationService(db)
	msgSvc := NewMessageService(db)

	// Create a conversation
	conv, err := convSvc.Create(context.Background(), "No Summaries")
	require.NoError(t, err)

	// Add only regular messages
	regularMsg := &Message{
		ConversationID: conv.ID,
		Role:           "user",
		Content:        "Regular message",
		Provider:       "anthropic",
		Model:          "claude-3-5-sonnet-20241022",
		IsSummary:      false,
	}
	err = msgSvc.Add(context.Background(), regularMsg)
	require.NoError(t, err)

	// Get summaries (should be empty)
	summaries, err := msgSvc.GetSummaries(context.Background(), conv.ID)
	require.NoError(t, err)
	assert.Empty(t, summaries)
}

func TestMessageService_NullableFields(t *testing.T) {
	db := setupTestDB(t)
	convSvc := NewConversationService(db)
	msgSvc := NewMessageService(db)

	// Create a conversation
	conv, err := convSvc.Create(context.Background(), "Test Conversation")
	require.NoError(t, err)

	// Add message with empty provider and model
	msg := &Message{
		ConversationID: conv.ID,
		Role:           "user",
		Content:        "Message with no provider/model",
		Provider:       "",
		Model:          "",
		IsSummary:      false,
	}
	err = msgSvc.Add(context.Background(), msg)
	require.NoError(t, err)

	// Retrieve and verify
	messages, err := msgSvc.GetByConversation(context.Background(), conv.ID)
	require.NoError(t, err)
	require.Len(t, messages, 1)
	assert.Equal(t, "", messages[0].Provider)
	assert.Equal(t, "", messages[0].Model)
}
