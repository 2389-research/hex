// ABOUTME: Tests for todo repository operations
// ABOUTME: Validates todo persistence, retrieval, and cleanup operations
package storage_test

import (
	"testing"
	"time"

	"github.com/harper/jeff/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSaveTodos_CreatesNewTodos(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	todos := []storage.Todo{
		{
			Content:    "First task",
			ActiveForm: "Working on first task",
			Status:     "pending",
		},
		{
			Content:    "Second task",
			ActiveForm: "Working on second task",
			Status:     "in_progress",
		},
	}

	err := storage.SaveTodos(db, todos, nil)
	require.NoError(t, err)

	// Verify todos were saved
	loaded, err := storage.LoadTodos(db, nil)
	require.NoError(t, err)
	require.Len(t, loaded, 2)

	assert.Equal(t, "First task", loaded[0].Content)
	assert.Equal(t, "pending", loaded[0].Status)
	assert.NotEmpty(t, loaded[0].ID)
	assert.False(t, loaded[0].CreatedAt.IsZero())
	assert.False(t, loaded[0].UpdatedAt.IsZero())

	assert.Equal(t, "Second task", loaded[1].Content)
	assert.Equal(t, "in_progress", loaded[1].Status)
}

func TestSaveTodos_UpdatesExistingTodos(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	// Create initial todos
	initial := []storage.Todo{
		{Content: "Task 1", ActiveForm: "Working on task 1", Status: "pending"},
		{Content: "Task 2", ActiveForm: "Working on task 2", Status: "pending"},
	}
	require.NoError(t, storage.SaveTodos(db, initial, nil))

	// Load and modify
	loaded, err := storage.LoadTodos(db, nil)
	require.NoError(t, err)
	require.Len(t, loaded, 2)

	// Update status
	loaded[0].Status = "completed"
	loaded[1].Status = "in_progress"

	// Save updates
	time.Sleep(10 * time.Millisecond) // Ensure updated_at changes
	err = storage.SaveTodos(db, loaded, nil)
	require.NoError(t, err)

	// Verify updates
	updated, err := storage.LoadTodos(db, nil)
	require.NoError(t, err)
	require.Len(t, updated, 2)

	assert.Equal(t, "completed", updated[0].Status)
	assert.Equal(t, "in_progress", updated[1].Status)
	assert.True(t, updated[0].UpdatedAt.After(updated[0].CreatedAt), "updated_at should be after created_at")
}

func TestSaveTodos_ReplacesAllTodos(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	// Create 3 todos
	initial := []storage.Todo{
		{Content: "Task 1", ActiveForm: "Working 1", Status: "pending"},
		{Content: "Task 2", ActiveForm: "Working 2", Status: "pending"},
		{Content: "Task 3", ActiveForm: "Working 3", Status: "pending"},
	}
	require.NoError(t, storage.SaveTodos(db, initial, nil))

	// Replace with only 2 todos
	replacement := []storage.Todo{
		{Content: "New Task A", ActiveForm: "Working A", Status: "pending"},
		{Content: "New Task B", ActiveForm: "Working B", Status: "completed"},
	}
	err := storage.SaveTodos(db, replacement, nil)
	require.NoError(t, err)

	// Verify old todos are gone, new ones exist
	loaded, err := storage.LoadTodos(db, nil)
	require.NoError(t, err)
	require.Len(t, loaded, 2)
	assert.Equal(t, "New Task A", loaded[0].Content)
	assert.Equal(t, "New Task B", loaded[1].Content)
}

func TestLoadTodos_EmptyDatabase(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	todos, err := storage.LoadTodos(db, nil)
	require.NoError(t, err)
	assert.Empty(t, todos)
}

func TestLoadTodos_OrderedByCreatedAt(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	// Create todos with slight delays to ensure ordering
	for i := 1; i <= 3; i++ {
		todos := []storage.Todo{
			{
				Content:    "Task " + string(rune('0'+i)),
				ActiveForm: "Working",
				Status:     "pending",
			},
		}
		require.NoError(t, storage.SaveTodos(db, todos, nil))
		if i < 3 {
			time.Sleep(10 * time.Millisecond)
		}
	}

	// Note: Each SaveTodos replaces all, so we should only have the last one
	// Let's modify the test to insert individually
	_, _ = db.Exec("DELETE FROM todos") // Clear first

	// Insert individually via SQL to test ordering
	now := time.Now()
	_, err := db.Exec(`INSERT INTO todos (id, content, active_form, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		"todo-1", "Task 1", "Working 1", "pending", now.Add(-2*time.Second), now.Add(-2*time.Second))
	require.NoError(t, err)

	_, err = db.Exec(`INSERT INTO todos (id, content, active_form, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		"todo-2", "Task 2", "Working 2", "in_progress", now.Add(-1*time.Second), now.Add(-1*time.Second))
	require.NoError(t, err)

	_, err = db.Exec(`INSERT INTO todos (id, content, active_form, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		"todo-3", "Task 3", "Working 3", "completed", now, now)
	require.NoError(t, err)

	// Load and verify order (should be oldest first)
	loaded, err := storage.LoadTodos(db, nil)
	require.NoError(t, err)
	require.Len(t, loaded, 3)

	assert.Equal(t, "Task 1", loaded[0].Content)
	assert.Equal(t, "Task 2", loaded[1].Content)
	assert.Equal(t, "Task 3", loaded[2].Content)
}

func TestClearCompleted_RemovesOnlyCompletedTodos(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	// Insert mixed todos directly
	now := time.Now()
	todos := []struct {
		id      string
		content string
		status  string
	}{
		{"todo-1", "Pending task", "pending"},
		{"todo-2", "In progress task", "in_progress"},
		{"todo-3", "Completed task 1", "completed"},
		{"todo-4", "Another pending", "pending"},
		{"todo-5", "Completed task 2", "completed"},
	}

	for _, td := range todos {
		_, err := db.Exec(`INSERT INTO todos (id, content, active_form, status, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?)`,
			td.id, td.content, "Working on "+td.content, td.status, now, now)
		require.NoError(t, err)
	}

	// Clear completed
	err := storage.ClearCompleted(db, nil)
	require.NoError(t, err)

	// Verify only non-completed remain
	loaded, err := storage.LoadTodos(db, nil)
	require.NoError(t, err)
	require.Len(t, loaded, 3, "should have 3 non-completed todos")

	for _, todo := range loaded {
		assert.NotEqual(t, "completed", todo.Status)
	}
}

func TestClearCompleted_EmptyDatabase(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	err := storage.ClearCompleted(db, nil)
	require.NoError(t, err)

	loaded, err := storage.LoadTodos(db, nil)
	require.NoError(t, err)
	assert.Empty(t, loaded)
}

func TestSaveTodos_WithConversationID(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	// Create a conversation first
	conv := &storage.Conversation{
		ID:    "conv-123",
		Title: "Test Conversation",
		Model: "claude-sonnet-4-5-20250929",
	}
	require.NoError(t, storage.CreateConversation(db, conv))

	// Save todos associated with conversation
	todos := []storage.Todo{
		{Content: "Task 1", ActiveForm: "Working 1", Status: "pending"},
		{Content: "Task 2", ActiveForm: "Working 2", Status: "completed"},
	}

	conversationID := "conv-123"
	err := storage.SaveTodos(db, todos, &conversationID)
	require.NoError(t, err)

	// Load and verify conversation_id is set
	loaded, err := storage.LoadTodos(db, &conversationID)
	require.NoError(t, err)
	require.Len(t, loaded, 2)

	for _, todo := range loaded {
		assert.NotNil(t, todo.ConversationID)
		assert.Equal(t, "conv-123", *todo.ConversationID)
	}
}

func TestLoadTodos_FiltersByConversationID(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	// Create two conversations
	conv1 := &storage.Conversation{ID: "conv-1", Title: "Conv 1", Model: "claude-sonnet-4-5-20250929"}
	conv2 := &storage.Conversation{ID: "conv-2", Title: "Conv 2", Model: "claude-sonnet-4-5-20250929"}
	require.NoError(t, storage.CreateConversation(db, conv1))
	require.NoError(t, storage.CreateConversation(db, conv2))

	// Save todos for conv-1
	conv1ID := "conv-1"
	todos1 := []storage.Todo{
		{Content: "Conv1 Task", ActiveForm: "Working", Status: "pending"},
	}
	require.NoError(t, storage.SaveTodos(db, todos1, &conv1ID))

	// Save todos for conv-2
	conv2ID := "conv-2"
	todos2 := []storage.Todo{
		{Content: "Conv2 Task", ActiveForm: "Working", Status: "pending"},
	}
	require.NoError(t, storage.SaveTodos(db, todos2, &conv2ID))

	// Load todos for conv-1 only
	loaded1, err := storage.LoadTodos(db, &conv1ID)
	require.NoError(t, err)
	require.Len(t, loaded1, 1)
	assert.Equal(t, "Conv1 Task", loaded1[0].Content)

	// Load todos for conv-2 only
	loaded2, err := storage.LoadTodos(db, &conv2ID)
	require.NoError(t, err)
	require.Len(t, loaded2, 1)
	assert.Equal(t, "Conv2 Task", loaded2[0].Content)
}

func TestClearCompleted_FiltersByConversationID(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	// Create conversations
	conv1 := &storage.Conversation{ID: "conv-1", Title: "Conv 1", Model: "claude-sonnet-4-5-20250929"}
	conv2 := &storage.Conversation{ID: "conv-2", Title: "Conv 2", Model: "claude-sonnet-4-5-20250929"}
	require.NoError(t, storage.CreateConversation(db, conv1))
	require.NoError(t, storage.CreateConversation(db, conv2))

	// Add completed todos to both conversations
	now := time.Now()
	_, err := db.Exec(`INSERT INTO todos (id, content, active_form, status, conversation_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		"todo-1", "Conv1 Completed", "Working", "completed", "conv-1", now, now)
	require.NoError(t, err)

	_, err = db.Exec(`INSERT INTO todos (id, content, active_form, status, conversation_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		"todo-2", "Conv2 Completed", "Working", "completed", "conv-2", now, now)
	require.NoError(t, err)

	// Clear completed only for conv-1
	conv1ID := "conv-1"
	err = storage.ClearCompleted(db, &conv1ID)
	require.NoError(t, err)

	// Verify conv-1 todos are gone
	loaded1, err := storage.LoadTodos(db, &conv1ID)
	require.NoError(t, err)
	assert.Empty(t, loaded1)

	// Verify conv-2 todos still exist
	conv2ID := "conv-2"
	loaded2, err := storage.LoadTodos(db, &conv2ID)
	require.NoError(t, err)
	assert.Len(t, loaded2, 1)
	assert.Equal(t, "Conv2 Completed", loaded2[0].Content)
}

func TestTodo_ValidationConstraints(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	now := time.Now()

	// Test empty content (should fail due to CHECK constraint)
	_, err := db.Exec(`INSERT INTO todos (id, content, active_form, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		"todo-bad", "   ", "Working", "pending", now, now)
	assert.Error(t, err, "empty content should violate CHECK constraint")

	// Test invalid status (should fail due to CHECK constraint)
	_, err = db.Exec(`INSERT INTO todos (id, content, active_form, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		"todo-bad2", "Task", "Working", "invalid_status", now, now)
	assert.Error(t, err, "invalid status should violate CHECK constraint")

	// Test empty active_form (should fail due to CHECK constraint)
	_, err = db.Exec(`INSERT INTO todos (id, content, active_form, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		"todo-bad3", "Task", "", "pending", now, now)
	assert.Error(t, err, "empty active_form should violate CHECK constraint")
}

func TestTodo_CascadeDelete(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	// Create conversation with todos
	conv := &storage.Conversation{ID: "conv-cascade", Title: "Test", Model: "claude-sonnet-4-5-20250929"}
	require.NoError(t, storage.CreateConversation(db, conv))

	convID := "conv-cascade"
	todos := []storage.Todo{
		{Content: "Task 1", ActiveForm: "Working 1", Status: "pending"},
		{Content: "Task 2", ActiveForm: "Working 2", Status: "completed"},
	}
	require.NoError(t, storage.SaveTodos(db, todos, &convID))

	// Verify todos exist
	loaded, err := storage.LoadTodos(db, &convID)
	require.NoError(t, err)
	assert.Len(t, loaded, 2)

	// Delete conversation
	err = storage.DeleteConversation(db, "conv-cascade")
	require.NoError(t, err)

	// Verify todos were cascade deleted
	loaded, err = storage.LoadTodos(db, &convID)
	require.NoError(t, err)
	assert.Empty(t, loaded, "todos should be deleted via CASCADE")
}
