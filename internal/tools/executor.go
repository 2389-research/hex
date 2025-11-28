// ABOUTME: Tool executor with permission management
// ABOUTME: Executes tools, handles approval flow, and manages tool lifecycle

package tools

import (
	"context"
	"fmt"
)

// ApprovalFunc is called when a tool needs user approval
// Returns true if approved, false if denied
type ApprovalFunc func(toolName string, params map[string]interface{}) bool

// Executor runs tools with permission management
type Executor struct {
	registry     *Registry
	approvalFunc ApprovalFunc
}

// NewExecutor creates a new tool executor
func NewExecutor(registry *Registry, approvalFunc ApprovalFunc) *Executor {
	return &Executor{
		registry:     registry,
		approvalFunc: approvalFunc,
	}
}

// Execute runs a tool by name with given parameters
func (e *Executor) Execute(ctx context.Context, toolName string, params map[string]interface{}) (*Result, error) {
	// Handle nil parameters
	if params == nil {
		params = make(map[string]interface{})
	}

	// Get tool from registry
	tool, err := e.registry.Get(toolName)
	if err != nil {
		return nil, fmt.Errorf("get tool: %w", err)
	}

	// Check if approval needed
	if tool.RequiresApproval(params) {
		if e.approvalFunc != nil && !e.approvalFunc(toolName, params) {
			return &Result{
				ToolName: toolName,
				Success:  false,
				Error:    "user denied permission",
			}, nil
		}
	}

	// Execute tool
	result, err := tool.Execute(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("execute tool: %w", err)
	}

	return result, nil
}
