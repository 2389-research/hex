// ABOUTME: Task productivity tools that route to provider implementations
// ABOUTME: Create, complete, and manage tasks through active provider

package tools

import (
	"context"

	"github.com/harper/pagent/internal/providers"
)

// CreateTaskTool creates tasks
type CreateTaskTool struct {
	registry *providers.Registry
}

// NewCreateTaskTool creates a new create task tool
func NewCreateTaskTool(registry *providers.Registry) Tool {
	return &CreateTaskTool{registry: registry}
}

// Name returns the tool's identifier
func (t *CreateTaskTool) Name() string {
	return "create_task"
}

// Description returns what this tool does
func (t *CreateTaskTool) Description() string {
	return "Create a new task with title, due date, notes, and priority"
}

// RequiresApproval returns whether this tool needs user confirmation
func (t *CreateTaskTool) RequiresApproval(_ map[string]interface{}) bool {
	// Creating tasks requires approval
	return true
}

// Execute runs the create task operation
func (t *CreateTaskTool) Execute(_ context.Context, params map[string]interface{}) (*Result, error) {
	result, err := t.registry.ExecuteTool("create_task", params)
	if err != nil {
		return &Result{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &Result{
		Success: result.Success,
		Output:  formatProviderResult(result),
		Error:   result.Error,
	}, nil
}

// ListTasksTool lists tasks
type ListTasksTool struct {
	registry *providers.Registry
}

// NewListTasksTool creates a new list tasks tool
func NewListTasksTool(registry *providers.Registry) Tool {
	return &ListTasksTool{registry: registry}
}

// Name returns the tool's identifier
func (t *ListTasksTool) Name() string {
	return "list_tasks"
}

// Description returns what this tool does
func (t *ListTasksTool) Description() string {
	return "List tasks with optional filters (status, due date, priority)"
}

// RequiresApproval returns whether this tool needs user confirmation
func (t *ListTasksTool) RequiresApproval(_ map[string]interface{}) bool {
	return true
}

// Execute runs the list tasks operation
func (t *ListTasksTool) Execute(_ context.Context, params map[string]interface{}) (*Result, error) {
	result, err := t.registry.ExecuteTool("list_tasks", params)
	if err != nil {
		return &Result{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &Result{
		Success: result.Success,
		Output:  formatProviderResult(result),
		Error:   result.Error,
	}, nil
}

// CompleteTaskTool marks tasks as complete
type CompleteTaskTool struct {
	registry *providers.Registry
}

// NewCompleteTaskTool creates a new complete task tool
func NewCompleteTaskTool(registry *providers.Registry) Tool {
	return &CompleteTaskTool{registry: registry}
}

// Name returns the tool's identifier
func (t *CompleteTaskTool) Name() string {
	return "complete_task"
}

// Description returns what this tool does
func (t *CompleteTaskTool) Description() string {
	return "Mark a task as completed"
}

// RequiresApproval returns whether this tool needs user confirmation
func (t *CompleteTaskTool) RequiresApproval(_ map[string]interface{}) bool {
	return true
}

// Execute runs the complete task operation
func (t *CompleteTaskTool) Execute(_ context.Context, params map[string]interface{}) (*Result, error) {
	result, err := t.registry.ExecuteTool("complete_task", params)
	if err != nil {
		return &Result{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &Result{
		Success: result.Success,
		Output:  formatProviderResult(result),
		Error:   result.Error,
	}, nil
}
