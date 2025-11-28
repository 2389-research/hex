// ABOUTME: TodoWrite tool for creating and managing structured task lists
// ABOUTME: Formats todo items with status indicators and tracks progress metrics

package tools

import (
	"context"
	"fmt"
	"strings"
)

// TodoWriteTool implements the todo_write tool for managing todo lists
type TodoWriteTool struct{}

// NewTodoWriteTool creates a new TodoWrite tool instance
func NewTodoWriteTool() *TodoWriteTool {
	return &TodoWriteTool{}
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
func (t *TodoWriteTool) RequiresApproval(params map[string]interface{}) bool {
	return false
}

// Execute creates and formats a todo list
func (t *TodoWriteTool) Execute(ctx context.Context, params map[string]interface{}) (*Result, error) {
	// Validate todos parameter exists
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
	type todo struct {
		content    string
		activeForm string
		status     string
	}

	var todos []todo
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

		todos = append(todos, todo{
			content:    content,
			activeForm: activeForm,
			status:     status,
		})
	}

	// Format the todo list
	var output strings.Builder
	pendingCount := 0
	inProgressCount := 0
	completedCount := 0

	for _, t := range todos {
		var statusIcon string
		var displayText string

		switch t.status {
		case "pending":
			statusIcon = "☐"
			displayText = t.content
			pendingCount++
		case "in_progress":
			statusIcon = "⏳"
			displayText = t.activeForm
			inProgressCount++
		case "completed":
			statusIcon = "✅"
			displayText = t.content
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
	}, nil
}
