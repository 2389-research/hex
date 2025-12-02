// ABOUTME: Integration tests for Write tool with Registry and Executor
// ABOUTME: Validates Write tool integration with the tool system

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

// TestWriteTool_Integration_WithRegistry tests Write tool registration
func TestWriteTool_Integration_WithRegistry(t *testing.T) {
	registry := tools.NewRegistry()
	writeTool := tools.NewWriteTool()

	// Register the tool
	err := registry.Register(writeTool)
	require.NoError(t, err)

	// Retrieve it
	retrieved, err := registry.Get("write_file")
	require.NoError(t, err)
	assert.Equal(t, writeTool, retrieved)

	// Verify it's in the list
	allTools := registry.List()
	assert.Contains(t, allTools, "write_file")
}

// TestWriteTool_Integration_WithExecutor tests Write tool execution through executor
func TestWriteTool_Integration_WithExecutor(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	registry := tools.NewRegistry()
	writeTool := tools.NewWriteTool()
	require.NoError(t, registry.Register(writeTool))

	// Create executor with auto-approve (for testing)
	approvalFunc := func(_ string, _ map[string]interface{}) bool {
		return true // Auto-approve for testing
	}
	executor := tools.NewExecutor(registry, approvalFunc)

	// Execute write
	result, err := executor.Execute(context.Background(), "write_file", map[string]interface{}{
		"path":    testFile,
		"content": "Hello from executor!",
	})

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Empty(t, result.Error)

	// Verify file was written
	//nolint:gosec // G304: Test file reads/writes are safe
	written, err := os.ReadFile(testFile)
	require.NoError(t, err)
	assert.Equal(t, "Hello from executor!", string(written))
}

// TestWriteTool_Integration_WithExecutor_ApprovalRequired tests approval requirement
func TestWriteTool_Integration_WithExecutor_ApprovalRequired(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	registry := tools.NewRegistry()
	writeTool := tools.NewWriteTool()
	require.NoError(t, registry.Register(writeTool))

	// Create executor with approval tracking
	approvalRequested := false
	approvalFunc := func(toolName string, params map[string]interface{}) bool {
		approvalRequested = true
		assert.Equal(t, "write_file", toolName)
		assert.Equal(t, testFile, params["path"])
		return true
	}
	executor := tools.NewExecutor(registry, approvalFunc)

	// Execute write
	result, err := executor.Execute(context.Background(), "write_file", map[string]interface{}{
		"path":    testFile,
		"content": "content",
	})

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.True(t, approvalRequested, "Approval should have been requested for write operation")
}

// TestWriteTool_Integration_WithExecutor_ApprovalDenied tests approval denial
func TestWriteTool_Integration_WithExecutor_ApprovalDenied(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	registry := tools.NewRegistry()
	writeTool := tools.NewWriteTool()
	require.NoError(t, registry.Register(writeTool))

	// Create executor that denies approval
	approvalFunc := func(_ string, _ map[string]interface{}) bool {
		return false // Deny approval
	}
	executor := tools.NewExecutor(registry, approvalFunc)

	// Execute write
	result, err := executor.Execute(context.Background(), "write_file", map[string]interface{}{
		"path":    testFile,
		"content": "content",
	})

	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "denied")

	// Verify file was NOT written
	_, err = os.Stat(testFile)
	assert.True(t, os.IsNotExist(err), "File should not exist when approval is denied")
}

// TestWriteTool_Integration_MultipleOperations tests multiple write operations
func TestWriteTool_Integration_MultipleOperations(t *testing.T) {
	tmpDir := t.TempDir()

	registry := tools.NewRegistry()
	writeTool := tools.NewWriteTool()
	require.NoError(t, registry.Register(writeTool))

	approvalFunc := func(_ string, _ map[string]interface{}) bool {
		return true
	}
	executor := tools.NewExecutor(registry, approvalFunc)

	// Create new file
	file1 := filepath.Join(tmpDir, "file1.txt")
	result, err := executor.Execute(context.Background(), "write_file", map[string]interface{}{
		"path":    file1,
		"content": "content1",
		"mode":    "create",
	})
	require.NoError(t, err)
	assert.True(t, result.Success)

	// Overwrite it
	result, err = executor.Execute(context.Background(), "write_file", map[string]interface{}{
		"path":    file1,
		"content": "new content",
		"mode":    "overwrite",
	})
	require.NoError(t, err)
	assert.True(t, result.Success)

	//nolint:gosec // G304: Test file reads/writes are safe
	// Verify overwrite
	content, err := os.ReadFile(file1) //nolint:gosec // G304: Path validated by caller
	require.NoError(t, err)
	assert.Equal(t, "new content", string(content))

	// Append to it
	result, err = executor.Execute(context.Background(), "write_file", map[string]interface{}{
		"path":    file1,
		"content": " appended",
		"mode":    "append",
	})
	require.NoError(t, err)
	assert.True(t, result.Success)
	//nolint:gosec // G304: Test file reads/writes are safe

	//nolint:gosec // G304: Test file reads/writes are safe
	// Verify append
	content, err = os.ReadFile(file1) //nolint:gosec // G304: Path validated by caller
	require.NoError(t, err)           //nolint:gosec // G304: Path validated by caller
	assert.Equal(t, "new content appended", string(content))
}

// TestWriteTool_Integration_WithNestedDirectories tests creating nested directories
func TestWriteTool_Integration_WithNestedDirectories(t *testing.T) {
	tmpDir := t.TempDir()

	registry := tools.NewRegistry()
	writeTool := tools.NewWriteTool()
	require.NoError(t, registry.Register(writeTool))

	approvalFunc := func(_ string, _ map[string]interface{}) bool {
		return true
	}
	executor := tools.NewExecutor(registry, approvalFunc)

	// Write to nested path that doesn't exist
	nestedFile := filepath.Join(tmpDir, "a", "b", "c", "file.txt")
	result, err := executor.Execute(context.Background(), "write_file", map[string]interface{}{
		"path":    nestedFile,
		"content": "nested content",
	})

	require.NoError(t, err)
	//nolint:gosec // G304: Test file reads/writes are safe
	assert.True(t, result.Success)

	// Verify file and directories were created
	//nolint:gosec // G304: Test file reads/writes are safe
	content, err := os.ReadFile(nestedFile)
	require.NoError(t, err)
	assert.Equal(t, "nested content", string(content))

	// Verify parent directories exist
	assert.DirExists(t, filepath.Join(tmpDir, "a"))
	assert.DirExists(t, filepath.Join(tmpDir, "a", "b"))
	assert.DirExists(t, filepath.Join(tmpDir, "a", "b", "c"))
}

// TestWriteTool_Integration_Interface verifies Write tool implements Tool interface
func TestWriteTool_Integration_Interface(t *testing.T) {
	var _ tools.Tool = (*tools.WriteTool)(nil) // Compile-time check

	writeTool := tools.NewWriteTool()
	assert.NotNil(t, writeTool)
	assert.Implements(t, (*tools.Tool)(nil), writeTool)
}
