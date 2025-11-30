// ABOUTME: Read tool implementation for file reading
// ABOUTME: Supports offset/limit parameters with safety checks and approval for sensitive paths

package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	// DefaultMaxFileSize is the default maximum file size to read (1MB)
	DefaultMaxFileSize = 1024 * 1024
)

// ReadTool implements file reading functionality
type ReadTool struct {
	MaxFileSize int64 // Maximum file size to read (bytes)
}

// NewReadTool creates a new read tool with default settings
func NewReadTool() *ReadTool {
	return &ReadTool{
		MaxFileSize: DefaultMaxFileSize,
	}
}

// Name returns the tool name
func (t *ReadTool) Name() string {
	return "read_file"
}

// Description returns the tool description
func (t *ReadTool) Description() string {
	return "Reads the contents of a file from the filesystem. Parameters: path (required), offset (optional), limit (optional)"
}

// RequiresApproval returns true if the path is sensitive
func (t *ReadTool) RequiresApproval(params map[string]interface{}) bool {
	// Get path parameter
	path, ok := params["path"].(string)
	if !ok || path == "" {
		return false // Invalid params, will fail in Execute
	}

	// Clean and resolve path
	cleanPath := filepath.Clean(path)
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return true // Can't resolve path, require approval for safety
	}

	// Require approval for sensitive directories
	sensitivePaths := []string{
		"/etc",
		"/sys",
		"/proc",
		"/dev",
		"/boot",
		"/root",
		"/var/log",
	}

	for _, sensitive := range sensitivePaths {
		if strings.HasPrefix(absPath, sensitive) {
			return true
		}
	}

	return false
}

// Execute reads the file and returns its contents
func (t *ReadTool) Execute(ctx context.Context, params map[string]interface{}) (*Result, error) {
	// Validate and extract path parameter
	path, ok := params["path"].(string)
	if !ok || path == "" {
		return &Result{
			ToolName: "read_file",
			Success:  false,
			Error:    "missing or invalid 'path' parameter",
		}, nil
	}

	// Clean path to prevent directory traversal
	cleanPath := filepath.Clean(path)

	// Get absolute path
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return &Result{
			ToolName: "read_file",
			Success:  false,
			Error:    fmt.Sprintf("invalid path: %v", err),
		}, nil
	}

	// Check if file exists
	fileInfo, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &Result{
				ToolName: "read_file",
				Success:  false,
				Error:    fmt.Sprintf("file not found: %s", path),
			}, nil
		}
		return &Result{
			ToolName: "read_file",
			Success:  false,
			Error:    fmt.Sprintf("cannot access file: %v", err),
		}, nil
	}

	// Check if it's a directory
	if fileInfo.IsDir() {
		return &Result{
			ToolName: "read_file",
			Success:  false,
			Error:    fmt.Sprintf("path is a directory, not a file: %s", path),
		}, nil
	}

	// Check file size
	if fileInfo.Size() > t.MaxFileSize {
		return &Result{
			ToolName: "read_file",
			Success:  false,
			Error:    fmt.Sprintf("file too large: %d bytes (max: %d bytes)", fileInfo.Size(), t.MaxFileSize),
		}, nil
	}

	// Read file contents
	content, err := os.ReadFile(absPath) //nolint:gosec // G304: Tool accepts user file paths as intended functionality
	if err != nil {
		return &Result{
			ToolName: "read_file",
			Success:  false,
			Error:    fmt.Sprintf("failed to read file: %v", err),
		}, nil
	}

	// Handle offset and limit if provided
	offset := 0
	limit := len(content)

	if offsetParam, ok := params["offset"].(float64); ok {
		offset = int(offsetParam)
	}

	if limitParam, ok := params["limit"].(float64); ok {
		limit = int(limitParam)
	}

	// Apply offset and limit
	if offset >= len(content) {
		offset = len(content)
		limit = 0
	} else if offset+limit > len(content) {
		limit = len(content) - offset
	}

	content = content[offset : offset+limit]

	// Return success result
	return &Result{
		ToolName: "read_file",
		Success:  true,
		Output:   string(content),
		Metadata: map[string]interface{}{
			"path":       absPath,
			"size":       fileInfo.Size(),
			"bytes_read": len(content),
			"offset":     offset,
		},
	}, nil
}
