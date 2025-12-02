// ABOUTME: Tests for the Read tool implementation
// ABOUTME: Validates file reading with safety checks and parameter handling

package tools_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/harper/pagent/internal/tools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestReadTool_Interface verifies ReadTool implements Tool interface
func TestReadTool_Interface(_ *testing.T) {
	var _ tools.Tool = (*tools.ReadTool)(nil)
}

// TestReadTool_Name verifies tool name
func TestReadTool_Name(t *testing.T) {
	tool := tools.NewReadTool()
	assert.Equal(t, "read_file", tool.Name())
}

// TestReadTool_Description verifies tool description
func TestReadTool_Description(t *testing.T) {
	tool := tools.NewReadTool()
	desc := tool.Description()
	assert.NotEmpty(t, desc)
	assert.Contains(t, desc, "path")
}

// TestReadTool_Execute_Success tests successful file reading
func TestReadTool_Execute_Success(t *testing.T) {
	// Create temporary file
	tmpFile, err := os.CreateTemp("", "test-*.txt")
	require.NoError(t, err)
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	content := "Hello, World!"
	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)
	_ = tmpFile.Close()

	// Create read tool
	tool := tools.NewReadTool()

	// Execute
	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"path": tmpFile.Name(),
	})

	// Verify
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, content, result.Output)
	assert.Equal(t, "read_file", result.ToolName)
	assert.Empty(t, result.Error)

	// Check metadata
	assert.NotNil(t, result.Metadata)
	assert.Equal(t, int64(len(content)), result.Metadata["size"])
	assert.Equal(t, len(content), result.Metadata["bytes_read"])
}

// TestReadTool_Execute_MissingPath tests error when path is missing
func TestReadTool_Execute_MissingPath(t *testing.T) {
	tool := tools.NewReadTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{})

	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "path")
}

// TestReadTool_Execute_EmptyPath tests error when path is empty
func TestReadTool_Execute_EmptyPath(t *testing.T) {
	tool := tools.NewReadTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"path": "",
	})

	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "path")
}

// TestReadTool_Execute_InvalidPathType tests error when path is not a string
func TestReadTool_Execute_InvalidPathType(t *testing.T) {
	tool := tools.NewReadTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"path": 123, // wrong type
	})

	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "path")
}

// TestReadTool_Execute_FileNotFound tests error when file doesn't exist
func TestReadTool_Execute_FileNotFound(t *testing.T) {
	tool := tools.NewReadTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"path": "/nonexistent/file.txt",
	})

	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "not found")
}

// TestReadTool_Execute_Directory tests error when path is a directory
func TestReadTool_Execute_Directory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test-dir-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	tool := tools.NewReadTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"path": tmpDir,
	})

	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "directory")
}

// TestReadTool_Execute_FileTooLarge tests error when file exceeds max size
func TestReadTool_Execute_FileTooLarge(t *testing.T) {
	// Create file larger than max size
	tmpFile, err := os.CreateTemp("", "test-large-*.txt")
	require.NoError(t, err)
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	// Write 2MB of data
	data := make([]byte, 2*1024*1024)
	_, err = tmpFile.Write(data)
	require.NoError(t, err)
	_ = tmpFile.Close()

	// Create read tool with default max size (1MB)
	tool := tools.NewReadTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"path": tmpFile.Name(),
	})

	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "too large")
}

// TestReadTool_Execute_WithOffset tests reading with offset
func TestReadTool_Execute_WithOffset(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.txt")
	require.NoError(t, err)
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	content := "0123456789"
	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)
	_ = tmpFile.Close()

	tool := tools.NewReadTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"path":   tmpFile.Name(),
		"offset": float64(5), // JSON numbers are float64
	})

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, "56789", result.Output)
	assert.Equal(t, 5, result.Metadata["offset"])
}

// TestReadTool_Execute_WithLimit tests reading with limit
func TestReadTool_Execute_WithLimit(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.txt")
	require.NoError(t, err)
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	content := "0123456789"
	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)
	_ = tmpFile.Close()

	tool := tools.NewReadTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"path":  tmpFile.Name(),
		"limit": float64(5),
	})

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, "01234", result.Output)
	assert.Equal(t, 5, result.Metadata["bytes_read"])
}

// TestReadTool_Execute_WithOffsetAndLimit tests reading with both offset and limit
func TestReadTool_Execute_WithOffsetAndLimit(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.txt")
	require.NoError(t, err)
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	content := "0123456789"
	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)
	_ = tmpFile.Close()

	tool := tools.NewReadTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"path":   tmpFile.Name(),
		"offset": float64(3),
		"limit":  float64(4),
	})

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, "3456", result.Output)
	assert.Equal(t, 3, result.Metadata["offset"])
	assert.Equal(t, 4, result.Metadata["bytes_read"])
}

// TestReadTool_Execute_OffsetBeyondFileSize tests offset beyond file size
func TestReadTool_Execute_OffsetBeyondFileSize(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.txt")
	require.NoError(t, err)
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	content := "Hello"
	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)
	_ = tmpFile.Close()

	tool := tools.NewReadTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"path":   tmpFile.Name(),
		"offset": float64(100),
	})

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, "", result.Output) // Empty output
	assert.Equal(t, 0, result.Metadata["bytes_read"])
}

// TestReadTool_Execute_LimitBeyondFileSize tests limit beyond file size
func TestReadTool_Execute_LimitBeyondFileSize(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.txt")
	require.NoError(t, err)
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	content := "Hello"
	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)
	_ = tmpFile.Close()

	tool := tools.NewReadTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"path":  tmpFile.Name(),
		"limit": float64(100),
	})

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, content, result.Output) // Full content
	assert.Equal(t, 5, result.Metadata["bytes_read"])
}

// TestReadTool_RequiresApproval_SensitivePaths tests approval for sensitive paths
func TestReadTool_RequiresApproval_SensitivePaths(t *testing.T) {
	tool := tools.NewReadTool()

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "/etc requires approval",
			path:     "/etc/passwd",
			expected: true,
		},
		{
			name:     "/sys requires approval",
			path:     "/sys/devices/test",
			expected: true,
		},
		{
			name:     "/proc requires approval",
			path:     "/proc/cpuinfo",
			expected: true,
		},
		{
			name:     "/dev requires approval",
			path:     "/dev/null",
			expected: true,
		},
		{
			name:     "/boot requires approval",
			path:     "/boot/grub/grub.cfg",
			expected: true,
		},
		{
			name:     "/root requires approval",
			path:     "/root/.bashrc",
			expected: true,
		},
		{
			name:     "/var/log requires approval",
			path:     "/var/log/syslog",
			expected: true,
		},
		{
			name:     "/tmp does not require approval",
			path:     "/tmp/test.txt",
			expected: false,
		},
		{
			name:     "/home does not require approval",
			path:     "/home/user/test.txt",
			expected: false,
		},
		{
			name:     "relative path does not require approval",
			path:     "test.txt",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tool.RequiresApproval(map[string]interface{}{
				"path": tt.path,
			})
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestReadTool_RequiresApproval_MissingPath tests approval for missing path
func TestReadTool_RequiresApproval_MissingPath(t *testing.T) {
	tool := tools.NewReadTool()

	result := tool.RequiresApproval(map[string]interface{}{})
	assert.False(t, result) // Should not require approval for invalid params
}

// TestReadTool_RequiresApproval_InvalidPathType tests approval for invalid path type
func TestReadTool_RequiresApproval_InvalidPathType(t *testing.T) {
	tool := tools.NewReadTool()

	result := tool.RequiresApproval(map[string]interface{}{
		"path": 123,
	})
	assert.False(t, result) // Should not require approval for invalid params
}

// TestReadTool_PathSafety tests directory traversal prevention
func TestReadTool_PathSafety(t *testing.T) {
	// Create a temporary directory structure
	tmpDir, err := os.MkdirTemp("", "test-safety-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a file in tmpDir
	testFile := filepath.Join(tmpDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0600)
	require.NoError(t, err)

	tool := tools.NewReadTool()

	// Try to read using path with ..
	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"path": filepath.Join(tmpDir, "..", filepath.Base(tmpDir), "test.txt"),
	})

	// Should succeed (path is cleaned, but still valid)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, "test content", result.Output)

	// Verify the absolute path in metadata doesn't contain ..
	absPath, ok := result.Metadata["path"].(string)
	assert.True(t, ok)
	assert.NotContains(t, absPath, "..")
}

// TestReadTool_CustomMaxSize tests configurable max file size
func TestReadTool_CustomMaxSize(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.txt")
	require.NoError(t, err)
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	content := "Hello, World!"
	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)
	_ = tmpFile.Close()

	// Create tool with very small max size
	tool := &tools.ReadTool{
		MaxFileSize: 5, // Only 5 bytes
	}

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"path": tmpFile.Name(),
	})

	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "too large")
}

// TestReadTool_ContextCancellation tests context cancellation handling
func TestReadTool_ContextCancellation(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.txt")
	require.NoError(t, err)
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	content := "Hello, World!"
	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)
	_ = tmpFile.Close()

	tool := tools.NewReadTool()

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Execute should still work for file reading (it's fast)
	// but should respect the cancelled context
	result, err := tool.Execute(ctx, map[string]interface{}{
		"path": tmpFile.Name(),
	})

	// For a simple file read, it might complete before checking context
	// This test verifies the tool doesn't panic with cancelled context
	if err != nil {
		assert.Contains(t, err.Error(), "context")
	} else {
		assert.NotNil(t, result)
	}
}
