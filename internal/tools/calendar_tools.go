// ABOUTME: Calendar productivity tools that route to provider implementations
// ABOUTME: Create, update, and manage calendar events through active provider

package tools

import (
	"context"

	"github.com/harper/clem/internal/providers"
)

// CreateEventTool creates calendar events
//
//nolint:revive // Tool methods follow standard Tool interface pattern
type CreateEventTool struct {
	registry *providers.Registry
}

// NewCreateEventTool creates a new create event tool
func NewCreateEventTool(registry *providers.Registry) Tool {
	return &CreateEventTool{registry: registry}
}

func (t *CreateEventTool) Name() string {
	return "create_event"
}

func (t *CreateEventTool) Description() string {
	return "Create a new calendar event with title, start/end time, attendees, and location"
}

func (t *CreateEventTool) RequiresApproval(_ map[string]interface{}) bool {
	// Creating events requires approval
	return true
}

func (t *CreateEventTool) Execute(_ context.Context, params map[string]interface{}) (*Result, error) {
	result, err := t.registry.ExecuteTool("create_event", params)
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

// ListEventsTool lists calendar events
//
//nolint:revive // Tool methods follow standard Tool interface pattern
type ListEventsTool struct {
	registry *providers.Registry
}

// NewListEventsTool creates a new list events tool
func NewListEventsTool(registry *providers.Registry) Tool {
	return &ListEventsTool{registry: registry}
}

func (t *ListEventsTool) Name() string {
	return "list_events"
}

func (t *ListEventsTool) Description() string {
	return "List calendar events within a date range"
}

func (t *ListEventsTool) RequiresApproval(_ map[string]interface{}) bool {
	return true
}

func (t *ListEventsTool) Execute(_ context.Context, params map[string]interface{}) (*Result, error) {
	result, err := t.registry.ExecuteTool("list_events", params)
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
