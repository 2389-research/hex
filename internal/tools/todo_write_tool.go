// ABOUTME: TodoWrite tool for creating and managing structured task lists
// ABOUTME: Formats todo items with status indicators and tracks progress metrics

package tools

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/2389-research/hex/internal/storage"
)

// TodoWriteTool implements the todo_write tool for managing todo lists
type TodoWriteTool struct {
	db *sql.DB
}

// NewTodoWriteTool creates a new TodoWrite tool instance
func NewTodoWriteTool() Tool {
	return &TodoWriteTool{db: nil}
}

// NewTodoWriteToolWithDB creates a new TodoWrite tool with database persistence
func NewTodoWriteToolWithDB(db *sql.DB) Tool {
	return &TodoWriteTool{db: db}
}

// Name returns the tool identifier
func (t *TodoWriteTool) Name() string {
	return "todo_write"
}

// Description returns a human-readable description
func (t *TodoWriteTool) Description() string {
	return "Create and manage structured todo lists for tracking progress"
}

// RequiresApproval returns false (TodoWrite is read-only display)
func (t *TodoWriteTool) RequiresApproval(_ map[string]interface{}) bool {
	return false
}

// Execute creates and formats a todo list
func (t *TodoWriteTool) Execute(_ context.Context, params map[string]interface{}) (*Result, error) {
	// Check if we should load from database
	loadFromDB := false
	if loadParam, ok := params["load_from_db"]; ok {
		if loadBool, ok := loadParam.(bool); ok {
			loadFromDB = loadBool
		}
	}

	// Extract optional conversation_id for scoping
	var conversationID *string
	if convIDParam, ok := params["conversation_id"]; ok {
		if convIDStr, ok := convIDParam.(string); ok && convIDStr != "" {
			conversationID = &convIDStr
		}
	}

	// If load_from_db is true and we have a DB, load existing todos
	if loadFromDB && t.db != nil {
		loadedTodos, err := storage.LoadTodos(t.db, conversationID)
		if err != nil {
			return &Result{
				ToolName: t.Name(),
				Success:  false,
				Error:    fmt.Sprintf("failed to load todos from database: %v", err),
			}, nil
		}

		// If we have loaded todos, format and return them
		if len(loadedTodos) > 0 {
			return t.formatTodos(loadedTodos), nil
		}
		// If no todos found, fall through to normal processing
	}

	// Validate todos parameter exists (unless we're loading from DB)
	todosParam, ok := params["todos"]
	if !ok {
		return &Result{
			ToolName: t.Name(),
			Success:  false,
			Error:    "missing required parameter: todos",
		}, nil
	}

	// Validate todos is an array
	todosArray, ok := todosParam.([]interface{})
	if !ok {
		return &Result{
			ToolName: t.Name(),
			Success:  false,
			Error:    "todos must be an array",
		}, nil
	}

	// Validate array is not empty
	if len(todosArray) == 0 {
		return &Result{
			ToolName: t.Name(),
			Success:  false,
			Error:    "todos array cannot be empty",
		}, nil
	}

	// Parse and validate each todo
	todos := make([]storage.Todo, 0, len(todosArray))
	for i, todoParam := range todosArray {
		// Validate todo is a map
		todoMap, ok := todoParam.(map[string]interface{})
		if !ok {
			return &Result{
				ToolName: t.Name(),
				Success:  false,
				Error:    fmt.Sprintf("todo at index %d must be an object", i),
			}, nil
		}

		// Extract and validate content
		contentParam, ok := todoMap["content"]
		if !ok {
			return &Result{
				ToolName: t.Name(),
				Success:  false,
				Error:    fmt.Sprintf("todo at index %d is missing required field: content", i),
			}, nil
		}
		content, ok := contentParam.(string)
		if !ok {
			return &Result{
				ToolName: t.Name(),
				Success:  false,
				Error:    fmt.Sprintf("todo at index %d: content must be a string", i),
			}, nil
		}
		if strings.TrimSpace(content) == "" {
			return &Result{
				ToolName: t.Name(),
				Success:  false,
				Error:    fmt.Sprintf("todo at index %d: content cannot be empty", i),
			}, nil
		}

		// Extract and validate activeForm
		activeFormParam, ok := todoMap["activeForm"]
		if !ok {
			return &Result{
				ToolName: t.Name(),
				Success:  false,
				Error:    fmt.Sprintf("todo at index %d is missing required field: activeForm", i),
			}, nil
		}
		activeForm, ok := activeFormParam.(string)
		if !ok {
			return &Result{
				ToolName: t.Name(),
				Success:  false,
				Error:    fmt.Sprintf("todo at index %d: activeForm must be a string", i),
			}, nil
		}
		if strings.TrimSpace(activeForm) == "" {
			return &Result{
				ToolName: t.Name(),
				Success:  false,
				Error:    fmt.Sprintf("todo at index %d: activeForm cannot be empty", i),
			}, nil
		}

		// Extract and validate status
		statusParam, ok := todoMap["status"]
		if !ok {
			return &Result{
				ToolName: t.Name(),
				Success:  false,
				Error:    fmt.Sprintf("todo at index %d is missing required field: status", i),
			}, nil
		}
		status, ok := statusParam.(string)
		if !ok {
			return &Result{
				ToolName: t.Name(),
				Success:  false,
				Error:    fmt.Sprintf("todo at index %d: status must be a string", i),
			}, nil
		}

		// Validate status value
		if status != "pending" && status != "in_progress" && status != "completed" {
			return &Result{
				ToolName: t.Name(),
				Success:  false,
				Error:    fmt.Sprintf("todo at index %d: status must be one of: pending, in_progress, completed", i),
			}, nil
		}

		todos = append(todos, storage.Todo{
			Content:    content,
			ActiveForm: activeForm,
			Status:     status,
		})
	}

	// Auto-save todos to database if available
	if t.db != nil {
		if err := storage.SaveTodos(t.db, todos, conversationID); err != nil {
			return &Result{
				ToolName: t.Name(),
				Success:  false,
				Error:    fmt.Sprintf("failed to save todos to database: %v", err),
			}, nil
		}
	}

	// Format and return the todos
	return t.formatTodos(todos), nil
}

// formatTodos formats a list of todos for display
func (t *TodoWriteTool) formatTodos(todos []storage.Todo) *Result {
	var output strings.Builder
	pendingCount := 0
	inProgressCount := 0
	completedCount := 0

	for _, todo := range todos {
		var statusIcon string
		var displayText string

		switch todo.Status {
		case "pending":
			statusIcon = "☐"
			displayText = todo.Content
			pendingCount++
		case "in_progress":
			statusIcon = "⏳"
			displayText = todo.ActiveForm
			inProgressCount++
		case "completed":
			statusIcon = "✅"
			displayText = todo.Content
			completedCount++
		}

		output.WriteString(fmt.Sprintf("%s %s\n", statusIcon, displayText))
	}

	return &Result{
		ToolName: t.Name(),
		Success:  true,
		Output:   strings.TrimSpace(output.String()),
		Metadata: map[string]interface{}{
			"total_count":       len(todos),
			"pending_count":     pendingCount,
			"in_progress_count": inProgressCount,
			"completed_count":   completedCount,
		},
	}
}
