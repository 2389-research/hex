// ABOUTME: End-to-end scenario tests for multi-tool batch execution
// ABOUTME: Tests the complete flow of collecting, approving, and executing multiple tools

package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/harper/clem/internal/tools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestScenario_BatchToolExecutionWithThreeWrites tests executing 3 write_file tools in sequence
func TestScenario_BatchToolExecutionWithThreeWrites(t *testing.T) {
	tmpDir := t.TempDir()

	// Setup tool system
	registry := tools.NewRegistry()
	require.NoError(t, registry.Register(tools.NewWriteTool()))

	// Track which tools were approved
	approvedTools := make([]string, 0)
	approvalFunc := func(toolName string, params map[string]interface{}) bool {
		approvedTools = append(approvedTools, toolName)
		return true // Auto-approve
	}

	executor := tools.NewExecutor(registry, approvalFunc)
	ctx := context.Background()

	// Simulate batch execution: 3 file writes in sequence
	toolConfigs := []struct {
		path    string
		content string
	}{
		{filepath.Join(tmpDir, "file1.txt"), "Content 1"},
		{filepath.Join(tmpDir, "file2.txt"), "Content 2"},
		{filepath.Join(tmpDir, "file3.txt"), "Content 3"},
	}

	results := make([]*tools.Result, 0, len(toolConfigs))

	for _, tool := range toolConfigs {
		params := map[string]interface{}{
			"path":    tool.path,
			"content": tool.content,
		}

		result, err := executor.Execute(ctx, "write_file", params)
		require.NoError(t, err)
		results = append(results, result)
	}

	// Verify all tools were approved
	assert.Len(t, approvedTools, 3)
	for _, toolName := range approvedTools {
		assert.Equal(t, "write_file", toolName)
	}

	// Verify all results are successful
	for i, result := range results {
		assert.True(t, result.Success, "Tool %d should succeed", i+1)
		assert.Empty(t, result.Error, "Tool %d should have no error", i+1)
	}

	// Verify all files were created
	for i, tool := range toolConfigs {
		content, err := os.ReadFile(tool.path)
		require.NoError(t, err, "File %d should exist", i+1)
		assert.Equal(t, tool.content, string(content), "File %d content should match", i+1)
	}
}

// TestScenario_MixedToolBatchExecution tests executing different tool types in one batch
func TestScenario_MixedToolBatchExecution(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "input.txt")

	// Create input file for read
	require.NoError(t, os.WriteFile(testFile, []byte("Input content"), 0600))

	// Setup tools
	registry := tools.NewRegistry()
	require.NoError(t, registry.Register(tools.NewReadTool()))
	require.NoError(t, registry.Register(tools.NewWriteTool()))

	executor := tools.NewExecutor(registry, func(toolName string, params map[string]interface{}) bool {
		return true
	})

	ctx := context.Background()

	// Execute sequence: read, write, read
	// 1. Read input file
	result1, err := executor.Execute(ctx, "read_file", map[string]interface{}{
		"path": testFile,
	})
	require.NoError(t, err)
	assert.True(t, result1.Success)
	assert.Contains(t, result1.Output, "Input content")

	// 2. Write output file
	outputFile := filepath.Join(tmpDir, "output.txt")
	result2, err := executor.Execute(ctx, "write_file", map[string]interface{}{
		"path":    outputFile,
		"content": "Output content",
	})
	require.NoError(t, err)
	assert.True(t, result2.Success)

	// 3. Read output file to verify
	result3, err := executor.Execute(ctx, "read_file", map[string]interface{}{
		"path": outputFile,
	})
	require.NoError(t, err)
	assert.True(t, result3.Success)
	assert.Contains(t, result3.Output, "Output content")
}

// TestScenario_BatchWithPartialFailure tests batch execution where one tool fails
func TestScenario_BatchWithPartialFailure(t *testing.T) {
	tmpDir := t.TempDir()

	registry := tools.NewRegistry()
	require.NoError(t, registry.Register(tools.NewWriteTool()))
	require.NoError(t, registry.Register(tools.NewReadTool()))

	executor := tools.NewExecutor(registry, func(toolName string, params map[string]interface{}) bool {
		return true
	})

	ctx := context.Background()

	// Tool 1: Valid write
	result1, err := executor.Execute(ctx, "write_file", map[string]interface{}{
		"path":    filepath.Join(tmpDir, "success1.txt"),
		"content": "Success 1",
	})
	require.NoError(t, err)
	assert.True(t, result1.Success)

	// Tool 2: Invalid read (nonexistent file) - should fail gracefully
	result2, err := executor.Execute(ctx, "read_file", map[string]interface{}{
		"path": "/nonexistent/path/file.txt",
	})
	require.NoError(t, err, "Should not return error, but failed result")
	assert.False(t, result2.Success, "Should have failed result")
	assert.NotEmpty(t, result2.Error, "Should have error message")

	// Tool 3: Valid write - should still succeed
	result3, err := executor.Execute(ctx, "write_file", map[string]interface{}{
		"path":    filepath.Join(tmpDir, "success2.txt"),
		"content": "Success 2",
	})
	require.NoError(t, err)
	assert.True(t, result3.Success)

	// Verify successful tools created their files
	//nolint:gosec // G304: Test file reads/writes are safe
	content1, err := os.ReadFile(filepath.Join(tmpDir, "success1.txt"))
	require.NoError(t, err)
	assert.Equal(t, "Success 1", string(content1))
	//nolint:gosec // G304: Test file reads/writes are safe

	content3, err := os.ReadFile(filepath.Join(tmpDir, "success2.txt")) //nolint:gosec // G304: Path validated by caller
	require.NoError(t, err)
	assert.Equal(t, "Success 2", string(content3))
}

// TestScenario_ToolDenialInBatch tests denying tools in a batch
func TestScenario_ToolDenialInBatch(t *testing.T) {
	tmpDir := t.TempDir()

	registry := tools.NewRegistry()
	require.NoError(t, registry.Register(tools.NewWriteTool()))

	// Deny the second tool only
	callCount := 0
	executor := tools.NewExecutor(registry, func(toolName string, params map[string]interface{}) bool {
		callCount++
		return callCount != 2 // Deny second tool
	})

	ctx := context.Background()

	// Tool 1: Approved
	result1, err := executor.Execute(ctx, "write_file", map[string]interface{}{
		"path":    filepath.Join(tmpDir, "file1.txt"),
		"content": "Content 1",
	})
	require.NoError(t, err)
	assert.True(t, result1.Success)

	// Tool 2: Denied
	result2, err := executor.Execute(ctx, "write_file", map[string]interface{}{
		"path":    filepath.Join(tmpDir, "file2.txt"),
		"content": "Content 2",
	})
	require.NoError(t, err)
	assert.False(t, result2.Success)
	assert.Contains(t, result2.Error, "denied")

	// Tool 3: Approved
	result3, err := executor.Execute(ctx, "write_file", map[string]interface{}{
		"path":    filepath.Join(tmpDir, "file3.txt"),
		"content": "Content 3",
	})
	require.NoError(t, err)
	assert.True(t, result3.Success)

	// Verify only approved files exist
	_, err = os.Stat(filepath.Join(tmpDir, "file1.txt"))
	assert.NoError(t, err, "File 1 should exist")

	_, err = os.Stat(filepath.Join(tmpDir, "file2.txt"))
	assert.True(t, os.IsNotExist(err), "File 2 should not exist (denied)")

	_, err = os.Stat(filepath.Join(tmpDir, "file3.txt"))
	assert.NoError(t, err, "File 3 should exist")
}

// TestScenario_LargeBatchExecution tests executing many tools in sequence
func TestScenario_LargeBatchExecution(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large batch test in short mode")
	}

	tmpDir := t.TempDir()

	registry := tools.NewRegistry()
	require.NoError(t, registry.Register(tools.NewWriteTool()))

	executor := tools.NewExecutor(registry, func(toolName string, params map[string]interface{}) bool {
		return true
	})

	ctx := context.Background()

	// Execute 20 file writes
	const batchSize = 20
	results := make([]*tools.Result, 0, batchSize)

	for i := 0; i < batchSize; i++ {
		result, err := executor.Execute(ctx, "write_file", map[string]interface{}{
			"path":    filepath.Join(tmpDir, "file_"+string(rune('A'+i))+".txt"),
			"content": "Content for file " + string(rune('A'+i)),
		})
		require.NoError(t, err)
		results = append(results, result)
	}

	// Verify all succeeded
	for i, result := range results {
		assert.True(t, result.Success, "Tool %d should succeed", i)
	}

	// Verify all files exist
	//nolint:gosec // G304: Test file reads/writes are safe
	for i := 0; i < batchSize; i++ {
		filename := filepath.Join(tmpDir, "file_"+string(rune('A'+i))+".txt")
		content, err := os.ReadFile(filename)
		require.NoError(t, err)
		assert.Equal(t, "Content for file "+string(rune('A'+i)), string(content))
	}
}

// TestIntegration_SequentialToolBatches tests multiple separate batch executions
func TestIntegration_SequentialToolBatches(t *testing.T) {
	tmpDir := t.TempDir()

	registry := tools.NewRegistry()
	require.NoError(t, registry.Register(tools.NewWriteTool()))
	require.NoError(t, registry.Register(tools.NewReadTool()))

	executor := tools.NewExecutor(registry, func(toolName string, params map[string]interface{}) bool {
		return true
	})

	ctx := context.Background()

	// Batch 1: Create 2 files
	result1, err := executor.Execute(ctx, "write_file", map[string]interface{}{
		"path":    filepath.Join(tmpDir, "batch1_file1.txt"),
		"content": "Batch 1 File 1",
	})
	require.NoError(t, err)
	assert.True(t, result1.Success)

	result2, err := executor.Execute(ctx, "write_file", map[string]interface{}{
		"path":    filepath.Join(tmpDir, "batch1_file2.txt"),
		"content": "Batch 1 File 2",
	})
	require.NoError(t, err)
	assert.True(t, result2.Success)

	// Batch 2: Read both files from batch 1
	result3, err := executor.Execute(ctx, "read_file", map[string]interface{}{
		"path": filepath.Join(tmpDir, "batch1_file1.txt"),
	})
	require.NoError(t, err)
	assert.True(t, result3.Success)
	assert.Contains(t, result3.Output, "Batch 1 File 1")

	result4, err := executor.Execute(ctx, "read_file", map[string]interface{}{
		"path": filepath.Join(tmpDir, "batch1_file2.txt"),
	})
	require.NoError(t, err)
	assert.True(t, result4.Success)
	assert.Contains(t, result4.Output, "Batch 1 File 2")

	// Batch 3: Create 2 more files
	result5, err := executor.Execute(ctx, "write_file", map[string]interface{}{
		"path":    filepath.Join(tmpDir, "batch3_file1.txt"),
		"content": "Batch 3 File 1",
	})
	require.NoError(t, err)
	assert.True(t, result5.Success)

	result6, err := executor.Execute(ctx, "write_file", map[string]interface{}{
		"path":    filepath.Join(tmpDir, "batch3_file2.txt"),
		"content": "Batch 3 File 2",
	})
	require.NoError(t, err)
	assert.True(t, result6.Success)

	// Verify all 4 files exist
	files := []string{
		"batch1_file1.txt",
		"batch1_file2.txt",
		"batch3_file1.txt",
		"batch3_file2.txt",
	}

	for _, filename := range files {
		_, err := os.Stat(filepath.Join(tmpDir, filename))
		assert.NoError(t, err, "File %s should exist", filename)
	}
}

// TestIntegration_ToolBatchErrorRecovery tests that execution continues after errors
func TestIntegration_ToolBatchErrorRecovery(t *testing.T) {
	tmpDir := t.TempDir()

	registry := tools.NewRegistry()
	require.NoError(t, registry.Register(tools.NewReadTool()))
	require.NoError(t, registry.Register(tools.NewWriteTool()))

	executor := tools.NewExecutor(registry, func(toolName string, params map[string]interface{}) bool {
		return true
	})

	ctx := context.Background()

	// Execute sequence with intentional failures interspersed
	results := make([]*tools.Result, 0, 5)

	// 1. Valid write
	r1, err := executor.Execute(ctx, "write_file", map[string]interface{}{
		"path":    filepath.Join(tmpDir, "file1.txt"),
		"content": "File 1",
	})
	require.NoError(t, err)
	results = append(results, r1)

	// 2. Invalid read
	r2, err := executor.Execute(ctx, "read_file", map[string]interface{}{
		"path": "/invalid/path1.txt",
	})
	require.NoError(t, err)
	results = append(results, r2)

	// 3. Valid write
	r3, err := executor.Execute(ctx, "write_file", map[string]interface{}{
		"path":    filepath.Join(tmpDir, "file3.txt"),
		"content": "File 3",
	})
	require.NoError(t, err)
	results = append(results, r3)

	// 4. Invalid read
	r4, err := executor.Execute(ctx, "read_file", map[string]interface{}{
		"path": "/invalid/path2.txt",
	})
	require.NoError(t, err)
	results = append(results, r4)

	// 5. Valid write
	r5, err := executor.Execute(ctx, "write_file", map[string]interface{}{
		"path":    filepath.Join(tmpDir, "file5.txt"),
		"content": "File 5",
	})
	require.NoError(t, err)
	results = append(results, r5)

	// Verify pattern: success, failure, success, failure, success
	assert.True(t, results[0].Success)
	assert.False(t, results[1].Success)
	assert.True(t, results[2].Success)
	assert.False(t, results[3].Success)
	assert.True(t, results[4].Success)

	// Verify successful files exist
	_, err = os.Stat(filepath.Join(tmpDir, "file1.txt"))
	assert.NoError(t, err)
	_, err = os.Stat(filepath.Join(tmpDir, "file3.txt"))
	assert.NoError(t, err)
	_, err = os.Stat(filepath.Join(tmpDir, "file5.txt"))
	assert.NoError(t, err)
}
