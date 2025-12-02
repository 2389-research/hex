// ABOUTME: SlashCommand tool implementation for executing slash commands
// ABOUTME: Loads and expands command templates to provide reusable prompts

package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/harper/clem/internal/tools"
)

// SlashCommandTool allows Claude to execute slash commands
type SlashCommandTool struct {
	registry *Registry
}

// NewSlashCommandTool creates a new slash command tool with the given registry
func NewSlashCommandTool(registry *Registry) *SlashCommandTool {
	return &SlashCommandTool{
		registry: registry,
	}
}

// Name returns the tool name
func (t *SlashCommandTool) Name() string {
	return "SlashCommand"
}

// Description returns the tool description
func (t *SlashCommandTool) Description() string {
	return `Execute a slash command to load a reusable prompt or workflow template.

Slash commands provide shortcuts for common workflows like planning, code review, debugging, and testing.

Parameters:
  - command (required): The slash command name (e.g., "plan", "review", "debug")
  - args (optional): Template arguments as a map (e.g., {"file": "main.go", "feature": "authentication"})

Available commands can be discovered with the 'list' parameter.

Example: {"command": "review", "args": {"file": "auth.go"}}`
}

// RequiresApproval returns false - slash commands are safe to execute
func (t *SlashCommandTool) RequiresApproval(_ map[string]interface{}) bool {
	// Slash commands are just prompt templates, safe to execute without approval
	return false
}

// Execute expands and returns a slash command's content
func (t *SlashCommandTool) Execute(_ context.Context, params map[string]interface{}) (*tools.Result, error) {
	// Check if user wants to list available commands
	if listParam, ok := params["list"].(bool); ok && listParam {
		return t.listCommands(), nil
	}

	// Validate command parameter
	command, ok := params["command"].(string)
	if !ok || command == "" {
		return &tools.Result{
			ToolName: "SlashCommand",
			Success:  false,
			Error:    "missing or invalid 'command' parameter (must be non-empty string with command name)",
		}, nil
	}

	// Remove leading slash if present
	command = strings.TrimPrefix(command, "/")

	// Get command from registry
	cmd, err := t.registry.Get(command)
	if err != nil {
		// Command not found - provide helpful error with suggestions
		suggestions := t.findSimilarCommands(command)
		errorMsg := fmt.Sprintf("Command '%s' not found", command)
		if len(suggestions) > 0 {
			errorMsg += fmt.Sprintf(". Did you mean: %s?", strings.Join(suggestions, ", "))
		}
		errorMsg += fmt.Sprintf("\n\nAvailable commands: %s", strings.Join(t.registry.List(), ", "))

		return &tools.Result{
			ToolName: "SlashCommand",
			Success:  false,
			Error:    errorMsg,
			Metadata: map[string]interface{}{
				"available_commands": t.registry.List(),
				"suggestions":        suggestions,
			},
		}, nil
	}

	// Extract args parameter if present
	// Ensure argsParam is not nil before type assertion to prevent panic
	var args map[string]interface{}
	if argsParam, ok := params["args"].(map[string]interface{}); ok && argsParam != nil {
		args = argsParam
	} else {
		args = make(map[string]interface{})
	}

	// Expand command template with args
	expanded, err := cmd.Expand(args)
	if err != nil {
		return &tools.Result{
			ToolName: "SlashCommand",
			Success:  false,
			Error:    fmt.Sprintf("failed to expand command template: %v", err),
			Metadata: map[string]interface{}{
				"command": cmd.Name,
				"args":    args,
			},
		}, nil
	}

	// Format output with command message indicator (matches Claude Code format)
	output := fmt.Sprintf("<command-message>%s is running…</command-message>\n\n%s", cmd.Name, expanded)

	return &tools.Result{
		ToolName: "SlashCommand",
		Success:  true,
		Output:   output,
		Metadata: map[string]interface{}{
			"command":     cmd.Name,
			"source":      cmd.Source,
			"has_args":    cmd.HasArgs(),
			"args_used":   args,
			"description": cmd.Description,
		},
	}, nil
}

// listCommands returns a formatted list of all available commands
func (t *SlashCommandTool) listCommands() *tools.Result {
	commands := t.registry.All()

	var sb strings.Builder
	sb.WriteString("# Available Slash Commands\n\n")

	if len(commands) == 0 {
		sb.WriteString("No commands available.\n")
	} else {
		for _, cmd := range commands {
			sb.WriteString(fmt.Sprintf("## /%s\n", cmd.Name))
			sb.WriteString(fmt.Sprintf("**Description**: %s\n", cmd.Description))
			sb.WriteString(fmt.Sprintf("**Source**: %s\n", cmd.Source))

			if cmd.HasArgs() {
				sb.WriteString("**Arguments**:\n")
				for argName, argDesc := range cmd.Args {
					sb.WriteString(fmt.Sprintf("  - `%s`: %s\n", argName, argDesc))
				}
			}

			sb.WriteString(fmt.Sprintf("\n**Usage**: `%s`\n\n", cmd.UsageString()))
			sb.WriteString("---\n\n")
		}
	}

	return &tools.Result{
		ToolName: "SlashCommand",
		Success:  true,
		Output:   sb.String(),
		Metadata: map[string]interface{}{
			"command_count": len(commands),
			"commands":      t.registry.List(),
		},
	}
}

// findSimilarCommands finds commands with similar names (simple fuzzy matching)
func (t *SlashCommandTool) findSimilarCommands(query string) []string {
	allCommands := t.registry.List()
	var similar []string

	lowerQuery := strings.ToLower(query)
	for _, name := range allCommands {
		lowerName := strings.ToLower(name)
		// Bidirectional substring matching
		if strings.Contains(lowerName, lowerQuery) || strings.Contains(lowerQuery, lowerName) {
			similar = append(similar, name)
		}
	}

	return similar
}
