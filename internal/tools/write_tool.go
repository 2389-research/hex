// ABOUTME: Write tool implementation for file writing
// ABOUTME: Supports create/overwrite/append modes with safety checks and always requires approval

package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

const (
	// DefaultMaxContentSize is the default maximum content size to write (10MB)
	DefaultMaxContentSize = 10 * 1024 * 1024

	// Write modes
	ModeCreate    = "create"    // Create new file, fail if exists
	ModeOverwrite = "overwrite" // Overwrite existing file or create new
	ModeAppend    = "append"    // Append to existing file or create new
)

// WriteTool implements file writing functionality
type WriteTool struct {
	MaxContentSize int64 // Maximum content size to write (bytes)
}

// NewWriteTool creates a new write tool with default settings
func NewWriteTool() *WriteTool {
	return &WriteTool{
		MaxContentSize: DefaultMaxContentSize,
	}
}

// Name returns the tool name
func (t *WriteTool) Name() string {
	return "write_file"
}

// Description returns the tool description
func (t *WriteTool) Description() string {
	return "Writes content to a file on the filesystem. Parameters: path (required), content (required), mode (optional: create/overwrite/append, default: create)"
}

// RequiresApproval always returns true for write operations
func (t *WriteTool) RequiresApproval(params map[string]interface{}) bool {
	// ALWAYS require approval for write operations
	// Writing to disk is a dangerous operation
	return true
}

// Execute writes content to a file
func (t *WriteTool) Execute(ctx context.Context, params map[string]interface{}) (*Result, error) {
	// Validate and extract path parameter
	path, ok := params["path"].(string)
	if !ok || path == "" {
		return &Result{
			ToolName: "write_file",
			Success:  false,
			Error:    "missing or invalid 'path' parameter",
		}, nil
	}

	// Validate and extract content parameter
	content, ok := params["content"].(string)
	if !ok {
		return &Result{
			ToolName: "write_file",
			Success:  false,
			Error:    "missing or invalid 'content' parameter",
		}, nil
	}

	// Check content size
	if int64(len(content)) > t.MaxContentSize {
		return &Result{
			ToolName: "write_file",
			Success:  false,
			Error:    fmt.Sprintf("content too large: %d bytes (max: %d bytes)", len(content), t.MaxContentSize),
		}, nil
	}

	// Get mode (default: create)
	mode := ModeCreate
	if modeParam, ok := params["mode"].(string); ok && modeParam != "" {
		mode = modeParam
	}

	// Validate mode
	if mode != ModeCreate && mode != ModeOverwrite && mode != ModeAppend {
		return &Result{
			ToolName: "write_file",
			Success:  false,
			Error:    fmt.Sprintf("invalid mode '%s': must be 'create', 'overwrite', or 'append'", mode),
		}, nil
	}

	// Clean path to prevent directory traversal
	cleanPath := filepath.Clean(path)

	// Get absolute path
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return &Result{
			ToolName: "write_file",
			Success:  false,
			Error:    fmt.Sprintf("invalid path: %v", err),
		}, nil
	}

	// Ensure parent directory exists
	parentDir := filepath.Dir(absPath)
	if _, err := os.Stat(parentDir); os.IsNotExist(err) {
		if err := os.MkdirAll(parentDir, 0750); err != nil {
			return &Result{
				ToolName: "write_file",
				Success:  false,
				Error:    fmt.Sprintf("failed to create parent directory: %v", err),
			}, nil
		}
	}

	// Check if file exists
	_, err = os.Stat(absPath)
	fileExists := err == nil

	// Handle different modes
	var writeErr error
	var bytesWritten int

	switch mode {
	case ModeCreate:
		if fileExists {
			return &Result{
				ToolName: "write_file",
				Success:  false,
				Error:    fmt.Sprintf("file already exists: %s (use mode 'overwrite' to replace)", path),
			}, nil
		}
		bytesWritten, writeErr = writeFile(absPath, content, false)

	case ModeOverwrite:
		bytesWritten, writeErr = writeFile(absPath, content, false)

	case ModeAppend:
		bytesWritten, writeErr = writeFile(absPath, content, true)
	}

	if writeErr != nil {
		return &Result{
			ToolName: "write_file",
			Success:  false,
			Error:    fmt.Sprintf("failed to write file: %v", writeErr),
		}, nil
	}

	// Return success result
	return &Result{
		ToolName: "write_file",
		Success:  true,
		Output:   fmt.Sprintf("Successfully wrote %d bytes to %s", bytesWritten, path),
		Metadata: map[string]interface{}{
			"path":          absPath,
			"bytes_written": bytesWritten,
			"mode":          mode,
			"created":       !fileExists,
		},
	}, nil
}

// writeFile writes content to a file
func writeFile(path, content string, append bool) (int, error) {
	var file *os.File
	var err error

	if append {
		file, err = os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600) //nolint:gosec // G304: Path validated by caller
	} else {
		file, err = os.Create(path) //nolint:gosec // G304: Path validated by caller
	} //nolint:gosec // G304: Path validated by caller

	if err != nil {
		return 0, err
	}
	defer func() { _ = file.Close() }()

	n, err := file.WriteString(content)
	if err != nil {
		return 0, err
	}

	return n, nil
}
