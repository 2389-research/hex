// ABOUTME: Tests for Write tool implementation
// ABOUTME: Validates file writing with safety checks, modes, and approval requirements

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

// TestWriteTool_Name verifies the tool name
func TestWriteTool_Name(t *testing.T) {
	tool := tools.NewWriteTool()
	assert.Equal(t, "write_file", tool.Name())
}

// TestWriteTool_Description verifies the tool description
func TestWriteTool_Description(t *testing.T) {
	tool := tools.NewWriteTool()
	desc := tool.Description()
	assert.NotEmpty(t, desc)
	assert.Contains(t, desc, "path")
	assert.Contains(t, desc, "content")
}

// TestWriteTool_RequiresApproval_AlwaysTrue verifies write always requires approval
func TestWriteTool_RequiresApproval_AlwaysTrue(t *testing.T) {
	tool := tools.NewWriteTool()

	tests := []struct {
		name   string
		params map[string]interface{}
	}{
		{
			name: "normal file",
			params: map[string]interface{}{
				"path":    "/tmp/test.txt",
				"content": "hello",
			},
		},
		{
			name: "home directory file",
			params: map[string]interface{}{
				"path":    "~/test.txt",
				"content": "hello",
			},
		},
		{
			name: "etc file",
			params: map[string]interface{}{
				"path":    "/etc/hosts",
				"content": "dangerous",
			},
		},
		{
			name:   "empty params",
			params: map[string]interface{}{},
		},
		{
			name: "invalid params",
			params: map[string]interface{}{
				"path": 123,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Write operations are ALWAYS dangerous and require approval
			assert.True(t, tool.RequiresApproval(tt.params))
		})
	}
}

// TestWriteTool_Execute_Create_Success tests creating a new file
func TestWriteTool_Execute_Create_Success(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	tool := tools.NewWriteTool()
	content := "Hello, World!"

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"path":    testFile,
		"content": content,
		"mode":    "create",
	})

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, "write_file", result.ToolName)
	assert.Empty(t, result.Error)
	assert.Contains(t, result.Output, "Successfully wrote")
	assert.Contains(t, result.Output, "13 bytes")

	// Check metadata
	require.NotNil(t, result.Metadata)
	assert.Equal(t, 13, result.Metadata["bytes_written"])
	assert.Equal(t, "create", result.Metadata["mode"])
	assert.True(t, result.Metadata["created"].(bool))

	// Verify file was actually written
	//nolint:gosec // G304: Test file reads/writes are safe
	written, err := os.ReadFile(testFile)
	require.NoError(t, err)
	assert.Equal(t, content, string(written))
}

// TestWriteTool_Execute_Create_FileExists tests create mode fails if file exists
func TestWriteTool_Execute_Create_FileExists(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "existing.txt")

	// Create existing file
	require.NoError(t, os.WriteFile(testFile, []byte("existing"), 0600))

	tool := tools.NewWriteTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"path":    testFile,
		"content": "new content",
		"mode":    "create",
	})

	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "already exists")
	assert.Contains(t, result.Error, "overwrite")

	//nolint:gosec // G304: Test file reads/writes are safe
	// Verify file was NOT modified
	existing, err := os.ReadFile(testFile) //nolint:gosec // G304: Path validated by caller
	require.NoError(t, err)
	assert.Equal(t, "existing", string(existing))
}

// TestWriteTool_Execute_Overwrite_Success tests overwriting an existing file
func TestWriteTool_Execute_Overwrite_Success(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "existing.txt")

	// Create existing file
	require.NoError(t, os.WriteFile(testFile, []byte("old content"), 0600))

	tool := tools.NewWriteTool()
	newContent := "new content"

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"path":    testFile,
		"content": newContent,
		"mode":    "overwrite",
	})

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, "overwrite", result.Metadata["mode"])
	assert.False(t, result.Metadata["created"].(bool))
	//nolint:gosec // G304: Test file reads/writes are safe

	//nolint:gosec // G304: Test file reads/writes are safe
	// Verify file was overwritten
	written, err := os.ReadFile(testFile) //nolint:gosec // G304: Path validated by caller
	require.NoError(t, err)               //nolint:gosec // G304: Path validated by caller
	assert.Equal(t, newContent, string(written))
}

// TestWriteTool_Execute_Overwrite_CreatesNew tests overwrite creates file if doesn't exist
func TestWriteTool_Execute_Overwrite_CreatesNew(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "new.txt")

	tool := tools.NewWriteTool()
	content := "content"

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"path":    testFile,
		"content": content,
		"mode":    "overwrite",
	})

	require.NoError(t, err)
	assert.True(t, result.Success)
	//nolint:gosec // G304: Test file reads/writes are safe
	assert.True(t, result.Metadata["created"].(bool))

	// Verify file was created
	//nolint:gosec // G304: Test file reads/writes are safe
	written, err := os.ReadFile(testFile)
	require.NoError(t, err)
	assert.Equal(t, content, string(written))
}

// TestWriteTool_Execute_Append_Success tests appending to existing file
func TestWriteTool_Execute_Append_Success(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "existing.txt")

	// Create existing file
	require.NoError(t, os.WriteFile(testFile, []byte("line1\n"), 0600))

	tool := tools.NewWriteTool()
	appendContent := "line2\n"

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"path":    testFile,
		"content": appendContent,
		"mode":    "append",
	})

	require.NoError(t, err)
	assert.True(t, result.Success)
	//nolint:gosec // G304: Test file reads/writes are safe
	assert.Equal(t, "append", result.Metadata["mode"])
	assert.False(t, result.Metadata["created"].(bool))

	// Verify content was appended
	//nolint:gosec // G304: Test file reads/writes are safe
	written, err := os.ReadFile(testFile)
	require.NoError(t, err)
	assert.Equal(t, "line1\nline2\n", string(written))
}

// TestWriteTool_Execute_Append_CreatesNew tests append creates file if doesn't exist
func TestWriteTool_Execute_Append_CreatesNew(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "new.txt")

	tool := tools.NewWriteTool()
	content := "content"

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"path":    testFile,
		"content": content,
		"mode":    "append",
	})

	//nolint:gosec // G304: Test file reads/writes are safe
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.True(t, result.Metadata["created"].(bool))

	// Verify file was created
	//nolint:gosec // G304: Test file reads/writes are safe
	written, err := os.ReadFile(testFile)
	require.NoError(t, err)
	assert.Equal(t, content, string(written))
}

// TestWriteTool_Execute_DefaultMode tests default mode is create
func TestWriteTool_Execute_DefaultMode(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	tool := tools.NewWriteTool()

	// Execute without mode parameter
	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"path":    testFile,
		"content": "content",
	})

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, "create", result.Metadata["mode"])
}

// TestWriteTool_Execute_MissingPath tests missing path parameter
func TestWriteTool_Execute_MissingPath(t *testing.T) {
	tool := tools.NewWriteTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"content": "some content",
	})

	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "path")
}

// TestWriteTool_Execute_EmptyPath tests empty path parameter
func TestWriteTool_Execute_EmptyPath(t *testing.T) {
	tool := tools.NewWriteTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"path":    "",
		"content": "some content",
	})

	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "path")
}

// TestWriteTool_Execute_InvalidPath tests invalid path type
func TestWriteTool_Execute_InvalidPath(t *testing.T) {
	tool := tools.NewWriteTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"path":    123,
		"content": "some content",
	})

	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "path")
}

// TestWriteTool_Execute_MissingContent tests missing content parameter
func TestWriteTool_Execute_MissingContent(t *testing.T) {
	tool := tools.NewWriteTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"path": "/tmp/test.txt",
	})

	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "content")
}

// TestWriteTool_Execute_InvalidContent tests invalid content type
func TestWriteTool_Execute_InvalidContent(t *testing.T) {
	tool := tools.NewWriteTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"path":    "/tmp/test.txt",
		"content": 123,
	})

	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "content")
}

// TestWriteTool_Execute_InvalidMode tests invalid mode value
func TestWriteTool_Execute_InvalidMode(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	tool := tools.NewWriteTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"path":    testFile,
		"content": "content",
		"mode":    "invalid_mode",
	})

	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "invalid mode")
	assert.Contains(t, result.Error, "create")
	assert.Contains(t, result.Error, "overwrite")
	assert.Contains(t, result.Error, "append")
}

// TestWriteTool_Execute_ContentTooLarge tests content size limit
func TestWriteTool_Execute_ContentTooLarge(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	tool := tools.NewWriteTool()

	// Create content larger than max (10MB + 1 byte)
	largeContent := make([]byte, 10*1024*1024+1)

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"path":    testFile,
		"content": string(largeContent),
	})

	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "too large")
}

// TestWriteTool_Execute_EmptyContent tests writing empty content is allowed
func TestWriteTool_Execute_EmptyContent(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	tool := tools.NewWriteTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"path":    testFile,
		"content": "",
	})
	//nolint:gosec // G304: Test file reads/writes are safe

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, 0, result.Metadata["bytes_written"])

	// Verify empty file was created
	//nolint:gosec // G304: Test file reads/writes are safe
	written, err := os.ReadFile(testFile)
	require.NoError(t, err)
	assert.Empty(t, written)
}

// TestWriteTool_Execute_CreatesParentDirs tests automatic parent directory creation
func TestWriteTool_Execute_CreatesParentDirs(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "subdir1", "subdir2", "test.txt")

	tool := tools.NewWriteTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"path": testFile,
		//nolint:gosec // G304: Test file reads/writes are safe
		"content": "content",
	})

	require.NoError(t, err)
	assert.True(t, result.Success)

	// Verify file and parent directories were created
	//nolint:gosec // G304: Test file reads/writes are safe
	written, err := os.ReadFile(testFile)
	require.NoError(t, err)
	assert.Equal(t, "content", string(written))

	// Verify parent dirs exist
	assert.DirExists(t, filepath.Join(tmpDir, "subdir1"))
	assert.DirExists(t, filepath.Join(tmpDir, "subdir1", "subdir2"))
}

// TestWriteTool_Execute_UnicodeContent tests writing Unicode content
func TestWriteTool_Execute_UnicodeContent(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "unicode.txt")

	tool := tools.NewWriteTool()
	unicodeContent := "Hello 世界 🌍 Здравствуй мир"

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		//nolint:gosec // G304: Test file reads/writes are safe
		"path":    testFile,
		"content": unicodeContent,
	})

	require.NoError(t, err)
	assert.True(t, result.Success)

	// Verify Unicode content was written correctly
	//nolint:gosec // G304: Test file reads/writes are safe
	written, err := os.ReadFile(testFile)
	require.NoError(t, err)
	assert.Equal(t, unicodeContent, string(written))
}

// TestWriteTool_Execute_PathCleaning tests path cleaning prevents traversal
func TestWriteTool_Execute_PathCleaning(t *testing.T) {
	tmpDir := t.TempDir()

	tool := tools.NewWriteTool()

	// Try to write using directory traversal
	//nolint:gosec // G304: Test file reads/writes are safe
	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"path":    filepath.Join(tmpDir, "subdir", "..", "test.txt"),
		"content": "content",
	})

	require.NoError(t, err)
	assert.True(t, result.Success)

	// Verify file was created in tmpDir, not outside it
	//nolint:gosec // G304: Test file reads/writes are safe
	written, err := os.ReadFile(filepath.Join(tmpDir, "test.txt"))
	require.NoError(t, err)
	assert.Equal(t, "content", string(written))

	// Path in metadata should be absolute and clean
	assert.NotContains(t, result.Metadata["path"], "..")
}

// TestWriteTool_Execute_PermissionDenied tests handling of permission errors
func TestWriteTool_Execute_PermissionDenied(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Skipping permission test when running as root")
	}

	tool := tools.NewWriteTool()

	// Try to write to a location we don't have permission for
	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"path":    "/root/test.txt",
		"content": "content",
	})

	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "failed to")
}

// TestWriteTool_Execute_AbsolutePath tests that metadata contains absolute path
func TestWriteTool_Execute_AbsolutePath(t *testing.T) {
	tmpDir := t.TempDir()

	tool := tools.NewWriteTool()

	// Use relative path
	relPath := filepath.Join(tmpDir, "test.txt")

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"path":    relPath,
		"content": "content",
	})

	require.NoError(t, err)
	assert.True(t, result.Success)

	// Metadata should have absolute path
	metadataPath := result.Metadata["path"].(string)
	assert.True(t, filepath.IsAbs(metadataPath))
}
