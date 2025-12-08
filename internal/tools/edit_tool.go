// ABOUTME: Edit tool implementation - exact string replacement in files
// ABOUTME: Supports single replacement (must be unique) or replace_all mode

package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/2389-research/hex/internal/filelock"
)

// EditTool performs exact string replacements in files
type EditTool struct{}

// NewEditTool creates a new Edit tool instance
func NewEditTool() *EditTool {
	return &EditTool{}
}

// Name returns the tool identifier
func (t *EditTool) Name() string {
	return "edit"
}

// Description returns the tool's purpose
func (t *EditTool) Description() string {
	return "Performs exact string replacements in files. Requires unique match unless replace_all is true."
}

// RequiresApproval always returns true - editing files is destructive
func (t *EditTool) RequiresApproval(_ map[string]interface{}) bool {
	// ALWAYS require approval for file edits
	return true
}

// Execute performs the string replacement
func (t *EditTool) Execute(_ context.Context, params map[string]interface{}) (*Result, error) {
	// Extract and validate parameters
	filePath, ok := params["file_path"].(string)
	if !ok || filePath == "" {
		return &Result{
			ToolName: t.Name(),
			Success:  false,
			Error:    "missing or invalid 'file_path' parameter",
		}, nil
	}

	oldString, ok := params["old_string"].(string)
	if !ok || oldString == "" {
		return &Result{
			ToolName: t.Name(),
			Success:  false,
			Error:    "missing or invalid 'old_string' parameter",
		}, nil
	}

	newString, ok := params["new_string"].(string)
	if !ok {
		return &Result{
			ToolName: t.Name(),
			Success:  false,
			Error:    "missing or invalid 'new_string' parameter",
		}, nil
	}

	// Check if old_string and new_string are identical
	if oldString == newString {
		return &Result{
			ToolName: t.Name(),
			Success:  false,
			Error:    "old_string and new_string must be different",
		}, nil
	}

	// Extract replace_all flag (default: false)
	replaceAll := false
	if replaceAllParam, ok := params["replace_all"].(bool); ok {
		replaceAll = replaceAllParam
	}

	// Get absolute path for locking
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return &Result{
			ToolName: t.Name(),
			Success:  false,
			Error:    fmt.Sprintf("failed to get absolute path: %v", err),
		}, nil
	}

	// Acquire file lock
	agentID := os.Getenv("HEX_AGENT_ID")
	if agentID == "" {
		agentID = "main" // Default for non-agent execution
	}

	lockManager := filelock.Global()
	if err := lockManager.Acquire(absPath, agentID, 30*time.Second); err != nil {
		return &Result{
			ToolName: t.Name(),
			Success:  false,
			Error:    fmt.Sprintf("failed to acquire file lock: %v", err),
		}, nil
	}
	defer func() {
		_ = lockManager.Release(absPath, agentID)
	}()

	// Read the file
	content, err := os.ReadFile(filePath) //nolint:gosec // G304: Tool accepts user file paths as intended functionality
	if err != nil {
		return &Result{
			ToolName: t.Name(),
			Success:  false,
			Error:    fmt.Sprintf("failed to read file: %v", err),
		}, nil
	}

	fileContent := string(content)

	// Count occurrences of old_string
	count := strings.Count(fileContent, oldString)

	if count == 0 {
		return &Result{
			ToolName: t.Name(),
			Success:  false,
			Error:    fmt.Sprintf("old_string not found in file: '%s'", oldString),
		}, nil
	}

	// If not replace_all and multiple matches, fail with ambiguous error
	if !replaceAll && count > 1 {
		return &Result{
			ToolName: t.Name(),
			Success:  false,
			Error:    fmt.Sprintf("old_string appears %d times in file. Use replace_all: true to replace all occurrences, or provide a more specific old_string to match exactly once.", count),
		}, nil
	}

	// Perform replacement
	var newContent string
	if replaceAll {
		newContent = strings.ReplaceAll(fileContent, oldString, newString)
	} else {
		// Replace only the first (and only) occurrence
		newContent = strings.Replace(fileContent, oldString, newString, 1)
	}

	// Write back to file
	if err := os.WriteFile(filePath, []byte(newContent), 0600); err != nil {
		return &Result{
			ToolName: t.Name(),
			Success:  false,
			Error:    fmt.Sprintf("failed to write file: %v", err),
		}, nil
	}

	// Success
	replacedCount := count
	if !replaceAll {
		replacedCount = 1
	}

	return &Result{
		ToolName: t.Name(),
		Success:  true,
		Output:   fmt.Sprintf("Successfully replaced %d occurrence(s) in %s", replacedCount, filePath),
	}, nil
}
