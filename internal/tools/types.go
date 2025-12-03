// ABOUTME: API integration types for tool use and results
// ABOUTME: Maps between internal tool execution and Anthropic API formats

package tools

import "github.com/harper/hex/internal/core"

// ToolUse is an alias for core.ToolUse for convenience
// Use core.ToolUse directly in new code
type ToolUse = core.ToolUse

// ToolResult represents a tool execution result to send back to the Anthropic API.
// This is sent as a content block in a user message after tool execution completes.
// See: https://docs.anthropic.com/claude/docs/tool-use#returning-tool-results
type ToolResult struct {
	Type      string `json:"type"`        // Always "tool_result"
	ToolUseID string `json:"tool_use_id"` // ID from the ToolUse request
	Content   string `json:"content"`     // Tool output or error message
	IsError   bool   `json:"is_error"`    // true if this represents an error
}

// ResultToToolResult converts an internal Result to a ToolResult for the API
func ResultToToolResult(result *Result, toolUseID string) ToolResult {
	var content string
	var isError bool

	if result.Success {
		if result.Output == "" {
			content = "(no output)"
		} else {
			content = result.Output
		}
		isError = false
	} else {
		content = "Error: " + result.Error
		isError = true
	}

	return ToolResult{
		Type:      "tool_result",
		ToolUseID: toolUseID,
		Content:   content,
		IsError:   isError,
	}
}
