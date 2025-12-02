// ABOUTME: Tool executor with permission management
// ABOUTME: Executes tools, handles approval flow, and manages tool lifecycle

package tools

import (
	"context"
	"fmt"

	"github.com/harper/clem/internal/hooks"
	"github.com/harper/clem/internal/permissions"
)

// ApprovalFunc is called when a tool needs user approval
// Returns true if approved, false if denied
type ApprovalFunc func(toolName string, params map[string]interface{}) bool

// PermissionHook is called before permission check, allowing external systems to intercept
type PermissionHook func(toolName string, params map[string]interface{}, checkResult permissions.CheckResult)

// Executor runs tools with permission management
type Executor struct {
	registry          *Registry
	approvalFunc      ApprovalFunc
	permissionChecker *permissions.Checker
	permissionHook    PermissionHook
	hookEngine        *hooks.Engine
}

// NewExecutor creates a new tool executor
func NewExecutor(registry *Registry, approvalFunc ApprovalFunc) *Executor {
	return &Executor{
		registry:          registry,
		approvalFunc:      approvalFunc,
		permissionChecker: nil, // No permission checker by default (backward compatible)
		permissionHook:    nil,
		hookEngine:        nil, // No hook engine by default (backward compatible)
	}
}

// SetPermissionChecker sets the permission checker for this executor
func (e *Executor) SetPermissionChecker(checker *permissions.Checker) {
	e.permissionChecker = checker
}

// SetPermissionHook sets a hook that fires before permission checking
func (e *Executor) SetPermissionHook(hook PermissionHook) {
	e.permissionHook = hook
}

// GetPermissionChecker returns the current permission checker (may be nil)
func (e *Executor) GetPermissionChecker() *permissions.Checker {
	return e.permissionChecker
}

// SetHookEngine sets the hook engine for this executor
func (e *Executor) SetHookEngine(engine *hooks.Engine) {
	e.hookEngine = engine
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

	// If we have a permission checker, use it first
	if e.permissionChecker != nil {
		checkResult := e.permissionChecker.Check(toolName, params)

		// Fire permission hook if set
		if e.permissionHook != nil {
			e.permissionHook(toolName, params, checkResult)
		}

		// If tool is blocked by rules, deny immediately
		if !checkResult.Allowed && !checkResult.RequiresPrompt {
			return &Result{
				ToolName: toolName,
				Success:  false,
				Error:    fmt.Sprintf("permission denied: %s", checkResult.Reason),
			}, nil
		}

		// If mode is auto, allow immediately
		if checkResult.Allowed && !checkResult.RequiresPrompt {
			// Fire PreToolUse hook if engine is set
			filePath := extractFilePath(params)
			if e.hookEngine != nil {
				_ = e.hookEngine.FirePreToolUse(toolName, filePath, false)
			}

			// Execute tool without prompt
			result, err := tool.Execute(ctx, params)

			// Fire PostToolUse hook if engine is set
			if e.hookEngine != nil {
				success := err == nil && result != nil && result.Success
				errMsg := ""
				if err != nil {
					errMsg = err.Error()
				} else if result != nil && !result.Success {
					errMsg = result.Error
				}
				_ = e.hookEngine.FirePostToolUse(toolName, filePath, success, errMsg, false)
			}

			if err != nil {
				return nil, fmt.Errorf("execute tool: %w", err)
			}
			return result, nil
		}

		// If we get here, RequiresPrompt is true, fall through to approval check
	}

	// Check if approval needed (legacy path or when mode is "ask")
	if tool.RequiresApproval(params) {
		if e.approvalFunc != nil && !e.approvalFunc(toolName, params) {
			return &Result{
				ToolName: toolName,
				Success:  false,
				Error:    "user denied permission",
			}, nil
		}
	}

	// Fire PreToolUse hook if engine is set
	filePath := extractFilePath(params)
	if e.hookEngine != nil {
		_ = e.hookEngine.FirePreToolUse(toolName, filePath, false)
	}

	// Execute tool
	result, err := tool.Execute(ctx, params)

	// Fire PostToolUse hook if engine is set
	if e.hookEngine != nil {
		success := err == nil && result != nil && result.Success
		errMsg := ""
		if err != nil {
			errMsg = err.Error()
		} else if result != nil && !result.Success {
			errMsg = result.Error
		}
		_ = e.hookEngine.FirePostToolUse(toolName, filePath, success, errMsg, false)
	}

	if err != nil {
		return nil, fmt.Errorf("execute tool: %w", err)
	}

	return result, nil
}

// extractFilePath extracts file_path parameter if present
func extractFilePath(params map[string]interface{}) string {
	if filePath, ok := params["file_path"].(string); ok {
		return filePath
	}
	return ""
}
