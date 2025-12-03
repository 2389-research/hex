// ABOUTME: Tests for TodoWrite tool functionality
// ABOUTME: Validates todo list creation, formatting, status handling, and validation

package tools_test

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/harper/hex/internal/storage"
	"github.com/harper/hex/internal/tools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestTodoWriteTool_Name(t *testing.T) {
	tool := tools.NewTodoWriteTool()
	assert.Equal(t, "todo_write", tool.Name())
}

func TestTodoWriteTool_Description(t *testing.T) {
	tool := tools.NewTodoWriteTool()
	desc := tool.Description()
	assert.NotEmpty(t, desc)
	assert.Contains(t, strings.ToLower(desc), "todo")
}

func TestTodoWriteTool_RequiresApproval_Never(t *testing.T) {
	tool := tools.NewTodoWriteTool()

	tests := []struct {
		name   string
		params map[string]interface{}
	}{
		{
			name:   "empty params",
			params: map[string]interface{}{},
		},
		{
			name: "single todo",
			params: map[string]interface{}{
				"todos": []interface{}{
					map[string]interface{}{
						"content":    "Run tests",
						"activeForm": "Running tests",
						"status":     "pending",
					},
				},
			},
		},
		{
			name: "multiple todos",
			params: map[string]interface{}{
				"todos": []interface{}{
					map[string]interface{}{
						"content":    "Write code",
						"activeForm": "Writing code",
						"status":     "in_progress",
					},
					map[string]interface{}{
						"content":    "Deploy",
						"activeForm": "Deploying",
						"status":     "completed",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TodoWrite never requires approval (read-only display)
			assert.False(t, tool.RequiresApproval(tt.params))
		})
	}
}

func TestTodoWriteTool_Execute_MissingTodos(t *testing.T) {
	tool := tools.NewTodoWriteTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{})
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "todos")
}

func TestTodoWriteTool_Execute_InvalidTodosType(t *testing.T) {
	tool := tools.NewTodoWriteTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"todos": "not an array",
	})
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "todos")
	assert.Contains(t, result.Error, "array")
}

func TestTodoWriteTool_Execute_EmptyTodosArray(t *testing.T) {
	tool := tools.NewTodoWriteTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"todos": []interface{}{},
	})
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "empty")
}

func TestTodoWriteTool_Execute_InvalidTodoType(t *testing.T) {
	tool := tools.NewTodoWriteTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"todos": []interface{}{
			"not a map",
		},
	})
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "object")
}

func TestTodoWriteTool_Execute_MissingContent(t *testing.T) {
	tool := tools.NewTodoWriteTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"todos": []interface{}{
			map[string]interface{}{
				"activeForm": "Running tests",
				"status":     "pending",
			},
		},
	})
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "content")
}

func TestTodoWriteTool_Execute_MissingActiveForm(t *testing.T) {
	tool := tools.NewTodoWriteTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"todos": []interface{}{
			map[string]interface{}{
				"content": "Run tests",
				"status":  "pending",
			},
		},
	})
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "activeForm")
}

func TestTodoWriteTool_Execute_MissingStatus(t *testing.T) {
	tool := tools.NewTodoWriteTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"todos": []interface{}{
			map[string]interface{}{
				"content":    "Run tests",
				"activeForm": "Running tests",
			},
		},
	})
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "status")
}

func TestTodoWriteTool_Execute_InvalidContentType(t *testing.T) {
	tool := tools.NewTodoWriteTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"todos": []interface{}{
			map[string]interface{}{
				"content":    123, // Not a string
				"activeForm": "Running tests",
				"status":     "pending",
			},
		},
	})
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "content")
	assert.Contains(t, result.Error, "string")
}

func TestTodoWriteTool_Execute_InvalidActiveFormType(t *testing.T) {
	tool := tools.NewTodoWriteTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"todos": []interface{}{
			map[string]interface{}{
				"content":    "Run tests",
				"activeForm": 456, // Not a string
				"status":     "pending",
			},
		},
	})
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "activeForm")
	assert.Contains(t, result.Error, "string")
}

func TestTodoWriteTool_Execute_InvalidStatusType(t *testing.T) {
	tool := tools.NewTodoWriteTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"todos": []interface{}{
			map[string]interface{}{
				"content":    "Run tests",
				"activeForm": "Running tests",
				"status":     true, // Not a string
			},
		},
	})
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "status")
	assert.Contains(t, result.Error, "string")
}

func TestTodoWriteTool_Execute_InvalidStatusValue(t *testing.T) {
	tool := tools.NewTodoWriteTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"todos": []interface{}{
			map[string]interface{}{
				"content":    "Run tests",
				"activeForm": "Running tests",
				"status":     "invalid_status",
			},
		},
	})
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "status")
	assert.Contains(t, strings.ToLower(result.Error), "pending")
	assert.Contains(t, strings.ToLower(result.Error), "in_progress")
	assert.Contains(t, strings.ToLower(result.Error), "completed")
}

func TestTodoWriteTool_Execute_SinglePendingTodo(t *testing.T) {
	tool := tools.NewTodoWriteTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"todos": []interface{}{
			map[string]interface{}{
				"content":    "Run tests",
				"activeForm": "Running tests",
				"status":     "pending",
			},
		},
	})

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, "todo_write", result.ToolName)
	assert.Empty(t, result.Error)

	// Check output format
	assert.Contains(t, result.Output, "☐")
	assert.Contains(t, result.Output, "Run tests")
	assert.NotContains(t, result.Output, "Running tests") // activeForm only shown for in_progress

	// Check metadata
	assert.Equal(t, 1, result.Metadata["total_count"])
	assert.Equal(t, 1, result.Metadata["pending_count"])
	assert.Equal(t, 0, result.Metadata["in_progress_count"])
	assert.Equal(t, 0, result.Metadata["completed_count"])
}

func TestTodoWriteTool_Execute_SingleInProgressTodo(t *testing.T) {
	tool := tools.NewTodoWriteTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"todos": []interface{}{
			map[string]interface{}{
				"content":    "Write code",
				"activeForm": "Writing code",
				"status":     "in_progress",
			},
		},
	})

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, "todo_write", result.ToolName)
	assert.Empty(t, result.Error)

	// Check output format
	assert.Contains(t, result.Output, "⏳")
	assert.Contains(t, result.Output, "Writing code")  // activeForm shown for in_progress
	assert.NotContains(t, result.Output, "Write code") // content not shown when in_progress

	// Check metadata
	assert.Equal(t, 1, result.Metadata["total_count"])
	assert.Equal(t, 0, result.Metadata["pending_count"])
	assert.Equal(t, 1, result.Metadata["in_progress_count"])
	assert.Equal(t, 0, result.Metadata["completed_count"])
}

func TestTodoWriteTool_Execute_SingleCompletedTodo(t *testing.T) {
	tool := tools.NewTodoWriteTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"todos": []interface{}{
			map[string]interface{}{
				"content":    "Deploy to production",
				"activeForm": "Deploying to production",
				"status":     "completed",
			},
		},
	})

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, "todo_write", result.ToolName)
	assert.Empty(t, result.Error)

	// Check output format
	assert.Contains(t, result.Output, "✅")
	assert.Contains(t, result.Output, "Deploy to production")
	assert.NotContains(t, result.Output, "Deploying to production") // activeForm only shown for in_progress

	// Check metadata
	assert.Equal(t, 1, result.Metadata["total_count"])
	assert.Equal(t, 0, result.Metadata["pending_count"])
	assert.Equal(t, 0, result.Metadata["in_progress_count"])
	assert.Equal(t, 1, result.Metadata["completed_count"])
}

func TestTodoWriteTool_Execute_MultipleTodos_MixedStatuses(t *testing.T) {
	tool := tools.NewTodoWriteTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"todos": []interface{}{
			map[string]interface{}{
				"content":    "Write tests",
				"activeForm": "Writing tests",
				"status":     "completed",
			},
			map[string]interface{}{
				"content":    "Write implementation",
				"activeForm": "Writing implementation",
				"status":     "in_progress",
			},
			map[string]interface{}{
				"content":    "Write documentation",
				"activeForm": "Writing documentation",
				"status":     "pending",
			},
			map[string]interface{}{
				"content":    "Review code",
				"activeForm": "Reviewing code",
				"status":     "pending",
			},
		},
	})

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, "todo_write", result.ToolName)
	assert.Empty(t, result.Error)

	// Check output format contains all todos
	assert.Contains(t, result.Output, "✅")
	assert.Contains(t, result.Output, "Write tests")
	assert.Contains(t, result.Output, "⏳")
	assert.Contains(t, result.Output, "Writing implementation")
	assert.Contains(t, result.Output, "☐")
	assert.Contains(t, result.Output, "Write documentation")
	assert.Contains(t, result.Output, "Review code")

	// Check metadata counts
	assert.Equal(t, 4, result.Metadata["total_count"])
	assert.Equal(t, 2, result.Metadata["pending_count"])
	assert.Equal(t, 1, result.Metadata["in_progress_count"])
	assert.Equal(t, 1, result.Metadata["completed_count"])
}

func TestTodoWriteTool_Execute_OutputStructure(t *testing.T) {
	tool := tools.NewTodoWriteTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"todos": []interface{}{
			map[string]interface{}{
				"content":    "First task",
				"activeForm": "Working on first task",
				"status":     "pending",
			},
			map[string]interface{}{
				"content":    "Second task",
				"activeForm": "Working on second task",
				"status":     "in_progress",
			},
		},
	})

	require.NoError(t, err)
	assert.True(t, result.Success)

	// Output should be structured with clear separation
	lines := strings.Split(strings.TrimSpace(result.Output), "\n")
	assert.GreaterOrEqual(t, len(lines), 2, "Should have at least 2 lines for 2 todos")

	// Each line should start with a status indicator
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			hasIndicator := strings.HasPrefix(trimmed, "☐") ||
				strings.HasPrefix(trimmed, "⏳") ||
				strings.HasPrefix(trimmed, "✅")
			assert.True(t, hasIndicator, "Line should start with status indicator: %s", line)
		}
	}
}

func TestTodoWriteTool_Execute_EmptyStrings(t *testing.T) {
	tool := tools.NewTodoWriteTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"todos": []interface{}{
			map[string]interface{}{
				"content":    "",
				"activeForm": "Running",
				"status":     "pending",
			},
		},
	})
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "content")
	assert.Contains(t, result.Error, "empty")
}

func TestTodoWriteTool_Execute_WhitespaceOnlyStrings(t *testing.T) {
	tool := tools.NewTodoWriteTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"todos": []interface{}{
			map[string]interface{}{
				"content":    "   ",
				"activeForm": "Running",
				"status":     "pending",
			},
		},
	})
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "content")
	assert.Contains(t, result.Error, "empty")
}

// Persistence tests

func setupTestDBForTools(t *testing.T) (*sql.DB, func()) {
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)

	// Enable foreign keys
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)

	err = storage.InitializeSchema(db)
	require.NoError(t, err)

	cleanup := func() {
		_ = db.Close()
	}

	return db, cleanup
}

func TestTodoWriteTool_Persistence_AutoSave(t *testing.T) {
	db, cleanup := setupTestDBForTools(t)
	defer cleanup()

	tool := tools.NewTodoWriteToolWithDB(db)

	// Create todos
	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"todos": []interface{}{
			map[string]interface{}{
				"content":    "Task 1",
				"activeForm": "Working on task 1",
				"status":     "pending",
			},
			map[string]interface{}{
				"content":    "Task 2",
				"activeForm": "Working on task 2",
				"status":     "completed",
			},
		},
	})

	require.NoError(t, err)
	assert.True(t, result.Success)

	// Verify they were saved to DB
	loaded, err := storage.LoadTodos(db, nil)
	require.NoError(t, err)
	assert.Len(t, loaded, 2)
	assert.Equal(t, "Task 1", loaded[0].Content)
	assert.Equal(t, "pending", loaded[0].Status)
	assert.Equal(t, "Task 2", loaded[1].Content)
	assert.Equal(t, "completed", loaded[1].Status)
}

func TestTodoWriteTool_Persistence_LoadFromDB(t *testing.T) {
	db, cleanup := setupTestDBForTools(t)
	defer cleanup()

	// Manually save todos to DB
	todos := []storage.Todo{
		{Content: "Loaded Task 1", ActiveForm: "Loading 1", Status: "pending"},
		{Content: "Loaded Task 2", ActiveForm: "Loading 2", Status: "in_progress"},
	}
	require.NoError(t, storage.SaveTodos(db, todos, nil))

	// Load them via the tool
	tool := tools.NewTodoWriteToolWithDB(db)
	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"load_from_db": true,
	})

	require.NoError(t, err)
	assert.True(t, result.Success)

	// Verify output contains loaded todos
	assert.Contains(t, result.Output, "Loaded Task 1")
	assert.Contains(t, result.Output, "Loading 2") // in_progress shows activeForm

	// Verify metadata
	assert.Equal(t, 2, result.Metadata["total_count"])
	assert.Equal(t, 1, result.Metadata["pending_count"])
	assert.Equal(t, 1, result.Metadata["in_progress_count"])
}

func TestTodoWriteTool_Persistence_LoadFromDB_Empty(t *testing.T) {
	db, cleanup := setupTestDBForTools(t)
	defer cleanup()

	tool := tools.NewTodoWriteToolWithDB(db)

	// Try to load from empty DB - should fail gracefully
	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"load_from_db": true,
	})

	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "todos")
}

func TestTodoWriteTool_Persistence_WithConversationID(t *testing.T) {
	db, cleanup := setupTestDBForTools(t)
	defer cleanup()

	// Create a conversation
	conv := &storage.Conversation{
		ID:    "conv-123",
		Title: "Test Conversation",
		Model: "claude-sonnet-4-5-20250929",
	}
	require.NoError(t, storage.CreateConversation(db, conv))

	tool := tools.NewTodoWriteToolWithDB(db)

	// Save todos with conversation_id
	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"conversation_id": "conv-123",
		"todos": []interface{}{
			map[string]interface{}{
				"content":    "Conv Task",
				"activeForm": "Working",
				"status":     "pending",
			},
		},
	})

	require.NoError(t, err)
	assert.True(t, result.Success)

	// Verify saved with conversation_id
	convID := "conv-123"
	loaded, err := storage.LoadTodos(db, &convID)
	require.NoError(t, err)
	assert.Len(t, loaded, 1)
	assert.Equal(t, "Conv Task", loaded[0].Content)
	assert.NotNil(t, loaded[0].ConversationID)
	assert.Equal(t, "conv-123", *loaded[0].ConversationID)
}

func TestTodoWriteTool_Persistence_LoadFromDB_WithConversationID(t *testing.T) {
	db, cleanup := setupTestDBForTools(t)
	defer cleanup()

	// Create conversation and save todos
	conv := &storage.Conversation{ID: "conv-456", Title: "Test", Model: "claude-sonnet-4-5-20250929"}
	require.NoError(t, storage.CreateConversation(db, conv))

	convID := "conv-456"
	todos := []storage.Todo{
		{Content: "Conv Task", ActiveForm: "Working", Status: "pending"},
	}
	require.NoError(t, storage.SaveTodos(db, todos, &convID))

	// Load via tool with conversation_id
	tool := tools.NewTodoWriteToolWithDB(db)
	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"load_from_db":    true,
		"conversation_id": "conv-456",
	})

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Contains(t, result.Output, "Conv Task")
}

func TestTodoWriteTool_Persistence_UpdateExisting(t *testing.T) {
	db, cleanup := setupTestDBForTools(t)
	defer cleanup()

	tool := tools.NewTodoWriteToolWithDB(db)

	// Create initial todos
	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"todos": []interface{}{
			map[string]interface{}{
				"content":    "Task 1",
				"activeForm": "Working 1",
				"status":     "pending",
			},
		},
	})
	require.NoError(t, err)
	assert.True(t, result.Success)

	// Update todos (replace with new list)
	result, err = tool.Execute(context.Background(), map[string]interface{}{
		"todos": []interface{}{
			map[string]interface{}{
				"content":    "Task 1 Updated",
				"activeForm": "Working 1 Updated",
				"status":     "completed",
			},
			map[string]interface{}{
				"content":    "Task 2 New",
				"activeForm": "Working 2",
				"status":     "pending",
			},
		},
	})
	require.NoError(t, err)
	assert.True(t, result.Success)

	// Verify database has updated todos
	loaded, err := storage.LoadTodos(db, nil)
	require.NoError(t, err)
	assert.Len(t, loaded, 2)
	assert.Equal(t, "Task 1 Updated", loaded[0].Content)
	assert.Equal(t, "completed", loaded[0].Status)
	assert.Equal(t, "Task 2 New", loaded[1].Content)
}

func TestTodoWriteTool_NoPersistence_WithoutDB(t *testing.T) {
	// Tool without DB should work but not persist
	tool := tools.NewTodoWriteTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"todos": []interface{}{
			map[string]interface{}{
				"content":    "Task",
				"activeForm": "Working",
				"status":     "pending",
			},
		},
	})

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Contains(t, result.Output, "Task")
}

func TestTodoWriteTool_LoadFromDB_WithoutDB_Ignored(t *testing.T) {
	// load_from_db should be ignored if no DB
	tool := tools.NewTodoWriteTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"load_from_db": true,
	})

	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "todos") // Falls through to normal validation
}
