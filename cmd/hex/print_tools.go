// ABOUTME: Tool setup helpers for print mode
package main

import (
	"fmt"

	"github.com/harper/hex/internal/logging"
	"github.com/harper/hex/internal/permissions"
	"github.com/harper/hex/internal/tools"
)

// setupPrintModeTools creates and configures tool registry and executor for print mode
func setupPrintModeTools() (*tools.Registry, *tools.Executor, error) {
	logging.Debug("Setting up tool registry and executor for print mode")

	// Create registry and register all available tools
	registry := tools.NewRegistry()

	// Register all core tools
	coreTools := []tools.Tool{
		tools.NewReadTool(),
		tools.NewWriteTool(),
		tools.NewEditTool(),
		tools.NewBashTool(),
		tools.NewGrepTool(),
		tools.NewGlobTool(),
	}

	for _, tool := range coreTools {
		if err := registry.Register(tool); err != nil {
			return nil, nil, fmt.Errorf("register tool: %w", err)
		}
	}

	// Create permission checker using the unified system
	permChecker, err := createPermissionChecker()
	if err != nil {
		return nil, nil, fmt.Errorf("create permission checker: %w", err)
	}

	// For print mode, if --tools flag is used (legacy), convert to allowed-tools
	// This maintains backward compatibility
	if len(enabledTools) > 0 && len(allowedTools) == 0 {
		logging.Info("Converting --tools flag to --allowed-tools for permission system")
		// Override the checker with tools-based rules
		rules := permissions.NewRules(enabledTools, nil)
		permChecker = permissions.NewChecker(permChecker.GetMode(), rules)
	}

	// Approval function - just returns true since permissions are handled by checker
	approvalFunc := func(_ string, _ map[string]interface{}) bool {
		return true
	}

	// Create executor and attach permission checker
	executor := tools.NewExecutor(registry, approvalFunc)
	executor.SetPermissionChecker(permChecker)

	toolDefs := registry.GetDefinitions()
	logging.InfoWith("Tools enabled for print mode",
		"count", len(toolDefs),
		"mode", permChecker.GetMode(),
		"allowed", allowedTools,
		"disallowed", disallowedTools,
	)

	return registry, executor, nil
}
