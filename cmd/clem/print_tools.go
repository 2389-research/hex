// ABOUTME: Tool setup helpers for print mode
package main

import (
	"fmt"
	"strings"

	"github.com/harper/clem/internal/logging"
	"github.com/harper/clem/internal/tools"
)

// setupPrintModeTools creates and configures tool registry and executor for print mode
func setupPrintModeTools() (*tools.Registry, *tools.Executor, error) {
	logging.Debug("Setting up tool registry and executor for print mode")

	// Create registry
	registry := tools.NewRegistry()

	// Determine which tools to register
	toolsToRegister := enabledTools
	if len(toolsToRegister) == 0 {
		// If no tools specified, enable all tools by default
		toolsToRegister = []string{"write_file", "read_file", "edit_file", "bash", "grep", "glob"}
	}

	// Register each tool
	for _, toolName := range toolsToRegister {
		toolName = strings.TrimSpace(strings.ToLower(toolName))
		var tool tools.Tool

		switch toolName {
		case "write_file", "write":
			tool = tools.NewWriteTool()
		case "read_file", "read":
			tool = tools.NewReadTool()
		case "edit_file", "edit":
			tool = tools.NewEditTool()
		case "bash":
			tool = tools.NewBashTool()
		case "grep":
			tool = tools.NewGrepTool()
		case "glob":
			tool = tools.NewGlobTool()
		default:
			logging.Warn("Unknown tool: " + toolName)
			continue
		}

		if err := registry.Register(tool); err != nil {
			return nil, nil, fmt.Errorf("register tool %s: %w", toolName, err)
		}
	}

	// Create policy-based approval function for print mode
	// Three approval strategies:
	// 1. --dangerously-skip-permissions: approve everything
	// 2. --tools specified: approve only those tools
	// 3. Default: approve only safe read-only tools
	approvalFunc := func(toolName string, _ map[string]interface{}) bool {
		// Strategy 1: Skip all permissions
		if dangerouslySkipPermissions {
			logging.DebugWith("Approving via --dangerously-skip-permissions", "tool", toolName)
			return true
		}

		// Strategy 2: Explicit allowlist via --tools
		if len(enabledTools) > 0 {
			// User explicitly specified tools, use that as allowlist
			for _, allowed := range enabledTools {
				if strings.EqualFold(allowed, toolName) ||
					strings.EqualFold(allowed+"_file", toolName) {
					logging.DebugWith("Approving via --tools allowlist", "tool", toolName)
					return true
				}
			}
			logging.WarnWith("Tool not in --tools allowlist", "tool", toolName, "allowed", enabledTools)
			return false
		}

		// Strategy 3: Default safe-only policy
		safeTool := toolName == "read_file" || toolName == "grep" || toolName == "glob"
		if safeTool {
			logging.DebugWith("Auto-approving safe tool", "tool", toolName)
			return true
		}

		// Block dangerous operations without explicit permission
		logging.WarnWith("Blocking dangerous tool - use --dangerously-skip-permissions or --tools to allow", "tool", toolName)
		return false
	}

	// Create executor
	executor := tools.NewExecutor(registry, approvalFunc)

	toolDefs := registry.GetDefinitions()
	logging.InfoWith("Tools enabled for print mode", "count", len(toolDefs), "auto_approve", dangerouslySkipPermissions)

	return registry, executor, nil
}
