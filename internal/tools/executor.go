// ABOUTME: Tool executor with permission management
// ABOUTME: Executes tools, handles approval flow, and manages tool lifecycle

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/2389-research/hex/internal/hooks"
	"github.com/2389-research/hex/internal/logging"
	"github.com/2389-research/hex/internal/permissions"
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
	resultCache       *ResultCache // LRU cache for read-only tool results
}

// NewExecutor creates a new tool executor
func NewExecutor(registry *Registry, approvalFunc ApprovalFunc) *Executor {
	return &Executor{
		registry:          registry,
		approvalFunc:      approvalFunc,
		permissionChecker: nil, // No permission checker by default (backward compatible)
		permissionHook:    nil,
		hookEngine:        nil,                                // No hook engine by default (backward compatible)
		resultCache:       NewResultCache(100, 5*time.Minute), // 100 entries, 5 minute TTL
	}
}

// EnableCache enables result caching with custom capacity and TTL
func (e *Executor) EnableCache(capacity int, ttl time.Duration) {
	e.resultCache = NewResultCache(capacity, ttl)
}

// DisableCache disables result caching
func (e *Executor) DisableCache() {
	e.resultCache = nil
}

// GetCacheStats returns cache statistics
func (e *Executor) GetCacheStats() *CacheStats {
	if e.resultCache == nil {
		return nil
	}
	stats := e.resultCache.Stats()
	return &stats
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

// Registry returns the tool registry associated with this executor.
// Returns nil if the executor is not initialized.
func (e *Executor) Registry() *Registry {
	return e.registry
}

// Execute runs a tool by name with given parameters
func (e *Executor) Execute(ctx context.Context, toolName string, params map[string]interface{}) (*Result, error) {
	// Handle nil parameters
	if params == nil {
		params = make(map[string]interface{})
	}

	// Debug logging: Log tool execution starting
	if os.Getenv("HEX_DEBUG") != "" {
		paramsJSON, _ := json.Marshal(params)
		logging.Debug("Tool execution starting", "tool", toolName, "params", string(paramsJSON))
	}

	// Check cache for cacheable tools
	if e.resultCache != nil && IsCacheable(toolName) {
		if cachedResult, found := e.resultCache.Get(toolName, params); found {
			if os.Getenv("HEX_DEBUG") != "" {
				logging.Debug("Tool result retrieved from cache", "tool", toolName)
			}
			return cachedResult, nil
		}
	}

	// Get tool from registry
	tool, err := e.registry.Get(toolName)
	if err != nil {
		if os.Getenv("HEX_DEBUG") != "" {
			logging.Debug("Tool not found in registry", "tool", toolName, "error", err)
		}
		return nil, fmt.Errorf("get tool: %w", err)
	}

	// If we have a permission checker, use it first
	if e.permissionChecker != nil {
		checkResult := e.permissionChecker.Check(toolName, params)

		// Debug logging: Log permission check result
		if os.Getenv("HEX_DEBUG") != "" {
			logging.Debug("Permission check result",
				"tool", toolName,
				"allowed", checkResult.Allowed,
				"requires_prompt", checkResult.RequiresPrompt,
				"reason", checkResult.Reason,
			)
		}

		// Fire permission hook if set
		if e.permissionHook != nil {
			e.permissionHook(toolName, params, checkResult)
		}

		// If tool is blocked by rules, deny immediately
		if !checkResult.Allowed && !checkResult.RequiresPrompt {
			if os.Getenv("HEX_DEBUG") != "" {
				logging.Debug("Tool execution denied by permission rules", "tool", toolName, "reason", checkResult.Reason)
			}
			return &Result{
				ToolName: toolName,
				Success:  false,
				Error:    fmt.Sprintf("permission denied: %s", checkResult.Reason),
			}, nil
		}

		// If mode is auto, allow immediately
		if checkResult.Allowed && !checkResult.RequiresPrompt {
			if os.Getenv("HEX_DEBUG") != "" {
				logging.Debug("Tool execution auto-approved", "tool", toolName)
			}
			// Fire PreToolUse hook if engine is set
			filePath := extractFilePath(params)
			if e.hookEngine != nil {
				_ = e.hookEngine.FirePreToolUse(toolName, filePath, false)
			}

			// Execute tool without prompt
			result, execErr := tool.Execute(ctx, params)

			// Debug logging: Log execution result
			if os.Getenv("HEX_DEBUG") != "" {
				if execErr != nil {
					logging.Debug("Tool execution failed", "tool", toolName, "error", execErr)
				} else if result != nil {
					resultJSON, _ := json.Marshal(result)
					logging.Debug("Tool execution completed", "tool", toolName, "success", result.Success, "result", string(resultJSON))
				}
			}

			// Fire PostToolUse hook if engine is set
			if e.hookEngine != nil {
				success := execErr == nil && result != nil && result.Success
				errMsg := ""
				if execErr != nil {
					errMsg = execErr.Error()
				} else if result != nil && !result.Success {
					errMsg = result.Error
				}
				_ = e.hookEngine.FirePostToolUse(toolName, filePath, success, errMsg, false)
			}

			if execErr != nil {
				return nil, fmt.Errorf("execute tool: %w", execErr)
			}
			if result == nil {
				return nil, fmt.Errorf("tool returned nil result")
			}

			// Cache successful results for cacheable tools
			if e.resultCache != nil && IsCacheable(toolName) && result.Success {
				e.resultCache.Set(toolName, params, result)
				if os.Getenv("HEX_DEBUG") != "" {
					logging.Debug("Tool result cached", "tool", toolName)
				}
			}

			return result, nil
		}

		// If we get here, RequiresPrompt is true, fall through to approval check
	}

	// Check if approval needed (legacy path or when mode is "ask")
	if tool.RequiresApproval(params) {
		if os.Getenv("HEX_DEBUG") != "" {
			logging.Debug("Tool requires user approval", "tool", toolName)
		}
		if e.approvalFunc != nil && !e.approvalFunc(toolName, params) {
			if os.Getenv("HEX_DEBUG") != "" {
				logging.Debug("User denied tool approval", "tool", toolName)
			}
			return &Result{
				ToolName: toolName,
				Success:  false,
				Error:    "user denied permission",
			}, nil
		}
		if os.Getenv("HEX_DEBUG") != "" {
			logging.Debug("User approved tool execution", "tool", toolName)
		}
	}

	// Fire PreToolUse hook if engine is set
	filePath := extractFilePath(params)
	if e.hookEngine != nil {
		_ = e.hookEngine.FirePreToolUse(toolName, filePath, false)
	}

	// Execute tool
	result, err := tool.Execute(ctx, params)

	// Debug logging: Log execution result
	if os.Getenv("HEX_DEBUG") != "" {
		if err != nil {
			logging.Debug("Tool execution failed", "tool", toolName, "error", err)
		} else if result != nil {
			resultJSON, _ := json.Marshal(result)
			logging.Debug("Tool execution completed", "tool", toolName, "success", result.Success, "result", string(resultJSON))
		}
	}

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
	if result == nil {
		return nil, fmt.Errorf("tool returned nil result")
	}

	// Cache successful results for cacheable tools
	if e.resultCache != nil && IsCacheable(toolName) && result.Success {
		e.resultCache.Set(toolName, params, result)
		if os.Getenv("HEX_DEBUG") != "" {
			logging.Debug("Tool result cached", "tool", toolName)
		}
	}

	return result, nil
}

// extractFilePath extracts file path from common parameter names
// Checks multiple common names: file_path, path, filepath, notebook_path
func extractFilePath(params map[string]interface{}) string {
	// Try common parameter names in order of priority
	pathKeys := []string{"file_path", "notebook_path", "path", "filepath"}

	for _, key := range pathKeys {
		if filePath, ok := params[key].(string); ok && filePath != "" {
			return filePath
		}
	}

	return ""
}
