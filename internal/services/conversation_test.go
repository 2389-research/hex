// ABOUTME: Test suite for ConversationService implementation
// ABOUTME: Tests CRUD operations, event publishing, and token tracking

package services

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/2389-research/hex/internal/pubsub"
	"github.com/2389-research/hex/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)

	// Enable foreign keys
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)

	// Run migrations
	err = storage.RunMigrations(db)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = db.Close()
	})

	return db
}

func TestConversationService_Create(t *testing.T) {
	db := setupTestDB(t)

	svc := NewConversationService(db)

	conv, err := svc.Create(context.Background(), "Test Conversation")
	require.NoError(t, err)
	assert.NotEmpty(t, conv.ID)
	assert.Equal(t, "Test Conversation", conv.Title)
	assert.NotZero(t, conv.CreatedAt)
	assert.NotZero(t, conv.UpdatedAt)
	assert.Equal(t, int64(0), conv.PromptTokens)
	assert.Equal(t, int64(0), conv.CompletionTokens)
	assert.Equal(t, 0.0, conv.TotalCost)
}

func TestConversationService_CreatePublishesEvent(t *testing.T) {
	db := setupTestDB(t)

	svc := NewConversationService(db)

	// Subscribe to events
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	events := svc.Subscribe(ctx)

	// Create conversation in goroutine to avoid blocking
	go func() {
		_, _ = svc.Create(context.Background(), "Test Conversation")
	}()

	// Wait for event
	select {
	case event := <-events:
		assert.Equal(t, pubsub.Created, event.Type)
		assert.Equal(t, "Test Conversation", event.Payload.Title)
	case <-ctx.Done():
		t.Fatal("timeout waiting for event")
	}
}

func TestConversationService_Get(t *testing.T) {
	db := setupTestDB(t)

	svc := NewConversationService(db)

	// Create a conversation
	created, err := svc.Create(context.Background(), "Test Conversation")
	require.NoError(t, err)

	// Get it back
	conv, err := svc.Get(context.Background(), created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, conv.ID)
	assert.Equal(t, "Test Conversation", conv.Title)
}

func TestConversationService_List(t *testing.T) {
	db := setupTestDB(t)

	svc := NewConversationService(db)

	// Create multiple conversations
	_, err := svc.Create(context.Background(), "First")
	require.NoError(t, err)
	time.Sleep(10 * time.Millisecond) // Ensure different timestamps

	_, err = svc.Create(context.Background(), "Second")
	require.NoError(t, err)

	// List all
	convs, err := svc.List(context.Background())
	require.NoError(t, err)
	assert.Len(t, convs, 2)

	// Should be ordered by updated_at DESC (newest first)
	assert.Equal(t, "Second", convs[0].Title)
	assert.Equal(t, "First", convs[1].Title)
}

func TestConversationService_Update(t *testing.T) {
	db := setupTestDB(t)

	svc := NewConversationService(db)

	// Create a conversation
	conv, err := svc.Create(context.Background(), "Original Title")
	require.NoError(t, err)

	// Update it
	conv.Title = "Updated Title"
	err = svc.Update(context.Background(), conv)
	require.NoError(t, err)

	// Verify the update
	updated, err := svc.Get(context.Background(), conv.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Title", updated.Title)
}

func TestConversationService_UpdatePublishesEvent(t *testing.T) {
	db := setupTestDB(t)

	svc := NewConversationService(db)

	// Create a conversation
	conv, err := svc.Create(context.Background(), "Original Title")
	require.NoError(t, err)

	// Subscribe to events after creation
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	events := svc.Subscribe(ctx)

	// Update in goroutine
	go func() {
		conv.Title = "Updated Title"
		_ = svc.Update(context.Background(), conv)
	}()

	// Wait for event
	select {
	case event := <-events:
		assert.Equal(t, pubsub.Updated, event.Type)
		assert.Equal(t, "Updated Title", event.Payload.Title)
	case <-ctx.Done():
		t.Fatal("timeout waiting for event")
	}
}

func TestConversationService_Delete(t *testing.T) {
	db := setupTestDB(t)

	svc := NewConversationService(db)

	// Create a conversation
	conv, err := svc.Create(context.Background(), "To Delete")
	require.NoError(t, err)

	// Delete it
	err = svc.Delete(context.Background(), conv.ID)
	require.NoError(t, err)

	// Verify it's gone
	_, err = svc.Get(context.Background(), conv.ID)
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
}

func TestConversationService_DeletePublishesEvent(t *testing.T) {
	db := setupTestDB(t)

	svc := NewConversationService(db)

	// Create a conversation
	conv, err := svc.Create(context.Background(), "To Delete")
	require.NoError(t, err)

	// Subscribe to events after creation
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	events := svc.Subscribe(ctx)

	// Delete in goroutine
	go func() {
		_ = svc.Delete(context.Background(), conv.ID)
	}()

	// Wait for event
	select {
	case event := <-events:
		assert.Equal(t, pubsub.Deleted, event.Type)
		assert.Equal(t, conv.ID, event.Payload.ID)
	case <-ctx.Done():
		t.Fatal("timeout waiting for event")
	}
}

func TestConversationService_UpdateTokenUsage(t *testing.T) {
	db := setupTestDB(t)

	svc := NewConversationService(db)

	// Create a conversation
	conv, err := svc.Create(context.Background(), "Token Test")
	require.NoError(t, err)

	// Update token usage
	promptTokens := int64(1000)
	completionTokens := int64(500)
	err = svc.UpdateTokenUsage(context.Background(), conv.ID, promptTokens, completionTokens)
	require.NoError(t, err)

	// Verify the update
	updated, err := svc.Get(context.Background(), conv.ID)
	require.NoError(t, err)
	assert.Equal(t, promptTokens, updated.PromptTokens)
	assert.Equal(t, completionTokens, updated.CompletionTokens)

	// Verify cost calculation: (1000/1M * $3) + (500/1M * $15)
	// = (0.001 * 3) + (0.0005 * 15) = 0.003 + 0.0075 = 0.0105
	expectedCost := 0.0105
	assert.InDelta(t, expectedCost, updated.TotalCost, 0.00001)
}

func TestConversationService_UpdateTokenUsageAccumulates(t *testing.T) {
	db := setupTestDB(t)

	svc := NewConversationService(db)

	// Create a conversation
	conv, err := svc.Create(context.Background(), "Token Accumulation Test")
	require.NoError(t, err)

	// First update
	err = svc.UpdateTokenUsage(context.Background(), conv.ID, 1000, 500)
	require.NoError(t, err)

	// Second update (should accumulate)
	err = svc.UpdateTokenUsage(context.Background(), conv.ID, 2000, 1000)
	require.NoError(t, err)

	// Verify accumulated values
	updated, err := svc.Get(context.Background(), conv.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(3000), updated.PromptTokens)
	assert.Equal(t, int64(1500), updated.CompletionTokens)

	// Cost: (3000/1M * $3) + (1500/1M * $15) = 0.009 + 0.0225 = 0.0315
	expectedCost := 0.0315
	assert.InDelta(t, expectedCost, updated.TotalCost, 0.00001)
}

func TestConversationService_UpdateTokenUsagePublishesEvent(t *testing.T) {
	db := setupTestDB(t)

	svc := NewConversationService(db)

	// Create a conversation
	conv, err := svc.Create(context.Background(), "Token Event Test")
	require.NoError(t, err)

	// Subscribe to events after creation
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	events := svc.Subscribe(ctx)

	// Update tokens in goroutine
	go func() {
		_ = svc.UpdateTokenUsage(context.Background(), conv.ID, 1000, 500)
	}()

	// Wait for event
	select {
	case event := <-events:
		assert.Equal(t, pubsub.Updated, event.Type)
		assert.Equal(t, int64(1000), event.Payload.PromptTokens)
		assert.Equal(t, int64(500), event.Payload.CompletionTokens)
		assert.Greater(t, event.Payload.TotalCost, 0.0)
	case <-ctx.Done():
		t.Fatal("timeout waiting for event")
	}
}
