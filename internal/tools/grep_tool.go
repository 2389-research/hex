// ABOUTME: Grep tool implementation - ripgrep-based code search
// ABOUTME: Supports patterns, context lines, filters, output modes

package tools

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// GrepTool searches code using ripgrep patterns
type GrepTool struct{}

// NewGrepTool creates a new Grep tool instance
func NewGrepTool() *GrepTool {
	return &GrepTool{}
}

// Name returns the tool identifier
func (t *GrepTool) Name() string {
	return "grep"
}

// Description returns the tool's purpose
func (t *GrepTool) Description() string {
	return "Search code with ripgrep patterns. Supports context lines, filters, and multiple output modes."
}

// RequiresApproval returns false for normal paths (grep is read-only)
func (t *GrepTool) RequiresApproval(params map[string]interface{}) bool {
	// Grep is read-only, no approval needed
	return false
}

// Execute performs the grep search
func (t *GrepTool) Execute(ctx context.Context, params map[string]interface{}) (*Result, error) {
	// Extract pattern (required)
	pattern, ok := params["pattern"].(string)
	if !ok || pattern == "" {
		return &Result{
			ToolName: t.Name(),
			Success:  false,
			Error:    "missing or invalid 'pattern' parameter",
		}, nil
	}

	// Extract path (default to current directory)
	searchPath := "."
	if pathParam, ok := params["path"].(string); ok && pathParam != "" {
		searchPath = pathParam
	}

	// Verify path exists
	if _, err := os.Stat(searchPath); err != nil {
		return &Result{
			ToolName: t.Name(),
			Success:  false,
			Error:    fmt.Sprintf("path does not exist: %s", searchPath),
		}, nil
	}

	// Check if ripgrep is installed
	if _, err := exec.LookPath("rg"); err != nil {
		return &Result{
			ToolName: t.Name(),
			Success:  false,
			Error:    "ripgrep (rg) is not installed. Install via: brew install ripgrep",
		}, nil
	}

	// Build ripgrep arguments
	args := []string{}

	// Output mode
	outputMode := "files_with_matches" // default
	if mode, ok := params["output_mode"].(string); ok {
		outputMode = mode
	}

	switch outputMode {
	case "files_with_matches":
		args = append(args, "-l") // --files-with-matches
	case "count":
		args = append(args, "-c") // --count
	case "content":
		args = append(args, "-n") // --line-number
	default:
		return &Result{
			ToolName: t.Name(),
			Success:  false,
			Error:    fmt.Sprintf("invalid output_mode: %s (must be 'content', 'files_with_matches', or 'count')", outputMode),
		}, nil
	}

	// Context lines (only for content mode)
	if outputMode == "content" {
		// Handle -B (context before) - accept both int and float64
		if contextBefore, ok := params["-B"].(int); ok {
			args = append(args, "-B", strconv.Itoa(contextBefore))
		} else if contextBefore, ok := params["-B"].(float64); ok {
			args = append(args, "-B", strconv.Itoa(int(contextBefore)))
		}

		// Handle -A (context after) - accept both int and float64
		if contextAfter, ok := params["-A"].(int); ok {
			args = append(args, "-A", strconv.Itoa(contextAfter))
		} else if contextAfter, ok := params["-A"].(float64); ok {
			args = append(args, "-A", strconv.Itoa(int(contextAfter)))
		}

		// Handle -C (context around) - accept both int and float64
		if contextAround, ok := params["-C"].(int); ok {
			args = append(args, "-C", strconv.Itoa(contextAround))
		} else if contextAround, ok := params["-C"].(float64); ok {
			args = append(args, "-C", strconv.Itoa(int(contextAround)))
		}
	}

	// Case insensitive
	if caseInsensitive, ok := params["-i"].(bool); ok && caseInsensitive {
		args = append(args, "-i")
	}

	// Glob filter
	if glob, ok := params["glob"].(string); ok && glob != "" {
		args = append(args, "--glob", glob)
	}

	// Type filter
	if typeFilter, ok := params["type"].(string); ok && typeFilter != "" {
		args = append(args, "--type", typeFilter)
	}

	// Add pattern and path
	args = append(args, pattern, searchPath)

	// Execute ripgrep
	cmd := exec.CommandContext(ctx, "rg", args...)
	output, err := cmd.CombinedOutput()

	// ripgrep returns exit code 1 when no matches found (not an error)
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 1 {
				// No matches found - this is success with empty result
				return &Result{
					ToolName: t.Name(),
					Success:  true,
					Output:   "",
				}, nil
			}
			// Other exit codes are real errors
			return &Result{
				ToolName: t.Name(),
				Success:  false,
				Error:    fmt.Sprintf("ripgrep failed: %s", string(output)),
			}, nil
		}
		// Command execution error
		return &Result{
			ToolName: t.Name(),
			Success:  false,
			Error:    fmt.Sprintf("failed to execute ripgrep: %v", err),
		}, nil
	}

	// Success
	return &Result{
		ToolName: t.Name(),
		Success:  true,
		Output:   strings.TrimSpace(string(output)),
	}, nil
}
