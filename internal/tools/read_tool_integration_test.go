// ABOUTME: Integration tests for Read tool with Registry and Executor
// ABOUTME: Validates end-to-end tool execution workflow

package tools_test

import (
	"context"
	"os"
	"testing"

	"github.com/2389-research/hex/internal/tools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestReadTool_Integration_WithRegistry tests Read tool with Registry
func TestReadTool_Integration_WithRegistry(t *testing.T) {
	// Create test file
	tmpFile, err := os.CreateTemp("", "integration-*.txt")
	require.NoError(t, err)
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	content := "Integration test content"
	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)
	_ = tmpFile.Close()

	// Create registry and register Read tool
	registry := tools.NewRegistry()
	err = registry.Register(tools.NewReadTool())
	require.NoError(t, err)

	// Verify tool is registered
	allTools := registry.List()
	assert.Len(t, allTools, 1)
	assert.Equal(t, "read_file", allTools[0])

	// Get tool and execute
	readTool, err := registry.Get("read_file")
	require.NoError(t, err)

	result, err := readTool.Execute(context.Background(), map[string]interface{}{
		"path": tmpFile.Name(),
	})

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, content, result.Output)
}

// TestReadTool_Integration_WithExecutor tests Read tool with Executor
func TestReadTool_Integration_WithExecutor(t *testing.T) {
	// Create test file
	tmpFile, err := os.CreateTemp("", "integration-*.txt")
	require.NoError(t, err)
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	content := "Executor integration test"
	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)
	_ = tmpFile.Close()

	// Create registry and executor
	registry := tools.NewRegistry()
	_ = registry.Register(tools.NewReadTool())

	approvalCalled := false
	executor := tools.NewExecutor(registry, func(_ string, _ map[string]interface{}) bool {
		approvalCalled = true
		return true
	})

	// Execute through executor (no approval needed for /tmp)
	result, err := executor.Execute(context.Background(), "read_file", map[string]interface{}{
		"path": tmpFile.Name(),
	})

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, content, result.Output)
	assert.False(t, approvalCalled) // /tmp doesn't require approval
}

// TestReadTool_Integration_WithExecutor_Approval tests approval flow
func TestReadTool_Integration_WithExecutor_Approval(t *testing.T) {
	registry := tools.NewRegistry()
	_ = registry.Register(tools.NewReadTool())

	approvalParams := make(map[string]interface{})
	executor := tools.NewExecutor(registry, func(_ string, params map[string]interface{}) bool {
		approvalParams = params
		return true // Approve
	})

	// Try to read sensitive path
	_, err := executor.Execute(context.Background(), "read_file", map[string]interface{}{
		"path": "/etc/passwd",
	})

	require.NoError(t, err)
	// Should have requested approval
	assert.Equal(t, "/etc/passwd", approvalParams["path"])
	// Result may fail due to permissions or succeed if file exists
	// We just verify the approval was requested
}

// TestReadTool_Integration_WithExecutor_ApprovalDenied tests denial flow
func TestReadTool_Integration_WithExecutor_ApprovalDenied(t *testing.T) {
	registry := tools.NewRegistry()
	_ = registry.Register(tools.NewReadTool())

	executor := tools.NewExecutor(registry, func(_ string, _ map[string]interface{}) bool {
		return false // Deny
	})

	// Try to read sensitive path
	result, err := executor.Execute(context.Background(), "read_file", map[string]interface{}{
		"path": "/etc/passwd",
	})

	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "denied")
}

// TestReadTool_Integration_MultipleOperations tests multiple reads in sequence
func TestReadTool_Integration_MultipleOperations(t *testing.T) {
	// Create multiple test files
	files := make([]string, 3)
	for i := 0; i < 3; i++ {
		tmpFile, err := os.CreateTemp("", "multi-*.txt")
		require.NoError(t, err)
		defer func() { _ = os.Remove(tmpFile.Name()) }()
		_, _ = tmpFile.WriteString(string(rune('A' + i)))
		_ = tmpFile.Close()
		files[i] = tmpFile.Name()
	}

	// Create registry and executor
	registry := tools.NewRegistry()
	_ = registry.Register(tools.NewReadTool())
	executor := tools.NewExecutor(registry, nil)

	// Read all files
	for i, file := range files {
		result, err := executor.Execute(context.Background(), "read_file", map[string]interface{}{
			"path": file,
		})

		require.NoError(t, err)
		assert.True(t, result.Success)
		assert.Equal(t, string(rune('A'+i)), result.Output)
	}
}

// TestReadTool_Integration_WithOffsetLimit tests offset/limit through executor
func TestReadTool_Integration_WithOffsetLimit(t *testing.T) {
	// Create test file
	tmpFile, err := os.CreateTemp("", "offset-*.txt")
	require.NoError(t, err)
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	content := "0123456789ABCDEF"
	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)
	_ = tmpFile.Close()

	// Create registry and executor
	registry := tools.NewRegistry()
	_ = registry.Register(tools.NewReadTool())
	executor := tools.NewExecutor(registry, nil)

	// Read with offset and limit
	result, err := executor.Execute(context.Background(), "read_file", map[string]interface{}{
		"path":   tmpFile.Name(),
		"offset": float64(5),
		"limit":  float64(10),
	})

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, "56789ABCDE", result.Output)
	assert.Equal(t, 5, result.Metadata["offset"])
	assert.Equal(t, 10, result.Metadata["bytes_read"])
}
