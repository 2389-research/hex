// ABOUTME: Tools integration tests for tool registry, executor, and all tools
// ABOUTME: Tests Read, Write, Bash tools with approval/denial flows

package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/2389-research/hex/internal/tools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestToolRegistryIntegration tests that all tools are registered correctly
func TestToolRegistryIntegration(t *testing.T) {
	registry := tools.NewRegistry()

	// Register all three tools
	readTool := tools.NewReadTool()
	writeTool := tools.NewWriteTool()
	bashTool := tools.NewBashTool()

	err := registry.Register(readTool)
	require.NoError(t, err)
	err = registry.Register(writeTool)
	require.NoError(t, err)
	err = registry.Register(bashTool)
	require.NoError(t, err)

	// Verify all tools registered
	toolNames := registry.List()
	assert.Len(t, toolNames, 3, "should have 3 tools registered")
	assert.Contains(t, toolNames, "read_file", "should have Read tool")
	assert.Contains(t, toolNames, "write_file", "should have Write tool")
	assert.Contains(t, toolNames, "bash", "should have Bash tool")

	// Verify we can get each tool
	rt, err := registry.Get("read_file")
	require.NoError(t, err)
	assert.NotNil(t, rt)

	wt, err := registry.Get("write_file")
	require.NoError(t, err)
	assert.NotNil(t, wt)

	bt, err := registry.Get("bash")
	require.NoError(t, err)
	assert.NotNil(t, bt)

	// Test getting non-existent tool
	_, err = registry.Get("NonExistent")
	assert.Error(t, err)
}

// TestReadToolExecution tests Read tool full execution
func TestReadToolExecution(t *testing.T) {
	// Create test file
	testFile := CreateTestFile(t, "Hello, world!\nThis is a test file.")

	// Create tool
	readTool := tools.NewReadTool()

	// Execute
	ctx := context.Background()
	params := map[string]interface{}{
		"path": testFile,
	}

	result, err := readTool.Execute(ctx, params)

	// Verify success
	require.NoError(t, err)
	require.True(t, result.Success, "read should succeed")
	assert.Contains(t, result.Output, "Hello, world!")
	assert.Contains(t, result.Output, "This is a test file.")
}

// TestWriteToolExecution tests Write tool full execution
func TestWriteToolExecution(t *testing.T) {
	tmpDir := CreateTestDir(t)
	testFile := filepath.Join(tmpDir, "output.txt")

	// Create tool
	writeTool := tools.NewWriteTool()

	// Execute
	ctx := context.Background()
	params := map[string]interface{}{
		"path":    testFile,
		"content": "Test content written by tool",
	}

	result, err := writeTool.Execute(ctx, params)

	// Verify success
	require.NoError(t, err)
	if !result.Success {
		t.Logf("Write tool failed: %s", result.Error)
	}
	require.True(t, result.Success, "write should succeed: %s", result.Error)

	// Verify file was created
	AssertFileExists(t, testFile)
	AssertFileContains(t, testFile, "Test content written by tool")
}

// TestBashToolExecution tests Bash tool full execution
func TestBashToolExecution(t *testing.T) {
	// Create tool
	bashTool := tools.NewBashTool()

	// Execute simple command
	ctx := context.Background()
	params := map[string]interface{}{
		"command": "echo 'Hello from bash'",
	}

	result, err := bashTool.Execute(ctx, params)

	// Verify success
	require.NoError(t, err)
	require.True(t, result.Success, "bash should succeed")
	assert.Contains(t, result.Output, "Hello from bash")
}

// TestBashToolTimeout tests that bash commands can timeout
func TestBashToolTimeout(t *testing.T) {
	bashTool := tools.NewBashTool()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	params := map[string]interface{}{
		"command": "sleep 10", // This will timeout
	}

	result, err := bashTool.Execute(ctx, params)

	// Should fail due to timeout (could be error or result with error)
	if err != nil {
		assert.Contains(t, err.Error(), "context deadline exceeded")
	} else {
		assert.False(t, result.Success, "long-running command should timeout")
	}
}

// TestToolExecutorWithApproval tests executor with approval flow
func TestToolExecutorWithApproval(t *testing.T) {
	registry := tools.NewRegistry()
	_ = registry.Register(tools.NewReadTool())

	// Create executor with auto-approve
	executor := tools.NewExecutor(registry, func(_ string, _ map[string]interface{}) bool {
		return true // Always approve
	})

	// Create test file
	testFile := CreateTestFile(t, "Executor test content")

	// Execute
	ctx := context.Background()
	params := map[string]interface{}{
		"path": testFile,
	}

	result, err := executor.Execute(ctx, "read_file", params)

	// Should succeed
	require.NoError(t, err)
	require.True(t, result.Success, "tool execution should succeed with auto-approve")
	assert.Contains(t, result.Output, "Executor test content")
}

// TestToolExecutorWithDenial tests executor denial flow
func TestToolExecutorWithDenial(t *testing.T) {
	registry := tools.NewRegistry()
	_ = registry.Register(tools.NewWriteTool())

	// Create executor with denial callback
	executor := tools.NewExecutor(registry, func(_ string, _ map[string]interface{}) bool {
		return false // Always deny
	})

	tmpDir := CreateTestDir(t)
	testFile := filepath.Join(tmpDir, "denied.txt")

	params := map[string]interface{}{
		"path":    testFile,
		"content": "This should be denied",
	}

	ctx := context.Background()
	result, err := executor.Execute(ctx, "write_file", params)

	// Should get denial result (not an error, but a failed result)
	require.NoError(t, err)
	assert.False(t, result.Success, "tool execution should fail when denied")
	assert.Contains(t, result.Error, "denied", "error should indicate denial")

	// File should not exist
	_, statErr := os.Stat(testFile)
	assert.Error(t, statErr, "file should not exist when tool is denied")
}

// TestReadToolWithNonExistentFile tests error handling
func TestReadToolWithNonExistentFile(t *testing.T) {
	readTool := tools.NewReadTool()

	ctx := context.Background()
	params := map[string]interface{}{
		"path": "/nonexistent/path/to/file.txt",
	}

	result, err := readTool.Execute(ctx, params)

	// Should return a result (not panic), but with error
	require.NoError(t, err, "should not return error, but failed result")
	assert.False(t, result.Success)
	assert.NotEmpty(t, result.Error)
}

// TestWriteToolCreatesDirectories tests that Write tool creates parent directories
func TestWriteToolCreatesDirectories(t *testing.T) {
	tmpDir := CreateTestDir(t)
	nestedPath := filepath.Join(tmpDir, "a", "b", "c", "file.txt")

	writeTool := tools.NewWriteTool()

	ctx := context.Background()
	params := map[string]interface{}{
		"path":    nestedPath,
		"content": "Nested file content",
	}

	result, err := writeTool.Execute(ctx, params)

	// Should succeed and create directories
	require.NoError(t, err)
	require.True(t, result.Success)
	AssertFileExists(t, nestedPath)
	AssertFileContains(t, nestedPath, "Nested file content")
}

// TestConcurrentToolExecution tests that multiple tools can execute concurrently
func TestConcurrentToolExecution(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}

	registry := tools.NewRegistry()
	_ = registry.Register(tools.NewReadTool())

	executor := tools.NewExecutor(registry, func(_ string, _ map[string]interface{}) bool {
		return true
	})

	// Create test files
	file1 := CreateTestFile(t, "File 1")
	file2 := CreateTestFile(t, "File 2")
	file3 := CreateTestFile(t, "File 3")

	// Execute concurrently
	results := make(chan *tools.Result, 3)

	execute := func(filepath string) {
		params := map[string]interface{}{
			"path": filepath,
		}
		result, err := executor.Execute(context.Background(), "read_file", params)
		require.NoError(t, err)
		results <- result
	}

	go execute(file1)
	go execute(file2)
	go execute(file3)

	// Collect results
	for i := 0; i < 3; i++ {
		result := <-results
		assert.True(t, result.Success, "concurrent execution should succeed")
	}
}

// ===========================
// Phase 3: Edit/Grep/Glob Integration Tests
// ===========================

func TestEditToolIntegration(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	// Create test file
	content := "Hello World\nThis is a test\n"
	err := os.WriteFile(testFile, []byte(content), 0600)
	require.NoError(t, err)

	registry := tools.NewRegistry()
	editTool := tools.NewEditTool()
	err = registry.Register(editTool)
	require.NoError(t, err)

	approvalFunc := func(_ string, _ map[string]interface{}) bool {
		return true // Auto-approve for test
	}
	executor := tools.NewExecutor(registry, approvalFunc)

	params := map[string]interface{}{
		"file_path":  testFile,
		"old_string": "World",
		"new_string": "Universe",
	}

	result, err := executor.Execute(context.Background(), "edit", params)
	require.NoError(t, err)
	assert.True(t, result.Success, "Edit should succeed")

	// Verify file was modified
	//nolint:gosec // G304: Test file reads/writes are safe
	newContent, err := os.ReadFile(testFile)
	require.NoError(t, err)
	assert.Contains(t, string(newContent), "Universe")
	assert.NotContains(t, string(newContent), "World")
}

func TestGrepToolIntegration(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	file1 := filepath.Join(tmpDir, "test.go")
	file2 := filepath.Join(tmpDir, "main.go")
	file3 := filepath.Join(tmpDir, "readme.txt")

	err := os.WriteFile(file1, []byte("package main\nfunc test() {}\n"), 0600)
	require.NoError(t, err)
	err = os.WriteFile(file2, []byte("package main\nfunc main() {}\n"), 0600)
	require.NoError(t, err)
	err = os.WriteFile(file3, []byte("This is a readme\n"), 0600)
	require.NoError(t, err)

	registry := tools.NewRegistry()
	grepTool := tools.NewGrepTool()
	err = registry.Register(grepTool)
	require.NoError(t, err)

	approvalFunc := func(_ string, _ map[string]interface{}) bool {
		return true
	}
	executor := tools.NewExecutor(registry, approvalFunc)

	// Search for "package" in .go files
	params := map[string]interface{}{
		"pattern":     "package",
		"path":        tmpDir,
		"glob":        "*.go",
		"output_mode": "files_with_matches",
	}

	result, err := executor.Execute(context.Background(), "grep", params)
	require.NoError(t, err)
	assert.True(t, result.Success, "Grep should succeed")

	output := result.Output
	assert.Contains(t, output, "test.go")
	assert.Contains(t, output, "main.go")
	assert.NotContains(t, output, "readme.txt")
}

func TestGlobToolIntegration(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files with different extensions
	files := []string{"test.go", "main.go", "test.txt", "README.md"}
	for _, name := range files {
		path := filepath.Join(tmpDir, name)
		err := os.WriteFile(path, []byte("content"), 0600)
		require.NoError(t, err)
	}

	registry := tools.NewRegistry()
	globTool := tools.NewGlobTool()
	err := registry.Register(globTool)
	require.NoError(t, err)

	approvalFunc := func(_ string, _ map[string]interface{}) bool {
		return true
	}
	executor := tools.NewExecutor(registry, approvalFunc)

	// Find all .go files
	params := map[string]interface{}{
		"pattern": "*.go",
		"path":    tmpDir,
	}

	result, err := executor.Execute(context.Background(), "glob", params)
	require.NoError(t, err)
	assert.True(t, result.Success, "Glob should succeed")

	output := result.Output
	assert.Contains(t, output, "test.go")
	assert.Contains(t, output, "main.go")
	assert.NotContains(t, output, "test.txt")
	assert.NotContains(t, output, "README.md")
}

func TestAllToolsRegistered(t *testing.T) {
	registry := tools.NewRegistry()

	// Register all 13 tools (Phase 1-4)
	// Phase 1: Read, Write, Bash
	err := registry.Register(tools.NewReadTool())
	require.NoError(t, err)
	err = registry.Register(tools.NewWriteTool())
	require.NoError(t, err)
	err = registry.Register(tools.NewBashTool())
	require.NoError(t, err)

	// Phase 3: Edit, Grep, Glob
	err = registry.Register(tools.NewEditTool())
	require.NoError(t, err)
	err = registry.Register(tools.NewGrepTool())
	require.NoError(t, err)
	err = registry.Register(tools.NewGlobTool())
	require.NoError(t, err)

	// Phase 4A: AskUserQuestion, TodoWrite
	err = registry.Register(tools.NewAskUserQuestionTool())
	require.NoError(t, err)
	err = registry.Register(tools.NewTodoWriteTool())
	require.NoError(t, err)

	// Phase 4B: WebFetch, WebSearch
	err = registry.Register(tools.NewWebFetchTool())
	require.NoError(t, err)
	err = registry.Register(tools.NewWebSearchTool())
	require.NoError(t, err)

	// Phase 4C: Task, BashOutput, KillShell
	err = registry.Register(tools.NewTaskTool())
	require.NoError(t, err)
	err = registry.Register(tools.NewBashOutputTool())
	require.NoError(t, err)
	err = registry.Register(tools.NewKillShellTool())
	require.NoError(t, err)

	// Verify all 13 tools are registered
	toolNames := registry.List()
	assert.Len(t, toolNames, 13, "should have 13 tools registered")

	// Phase 1 tools
	assert.Contains(t, toolNames, "read_file")
	assert.Contains(t, toolNames, "write_file")
	assert.Contains(t, toolNames, "bash")

	// Phase 3 tools
	assert.Contains(t, toolNames, "edit")
	assert.Contains(t, toolNames, "grep")
	assert.Contains(t, toolNames, "glob")

	// Phase 4A tools
	assert.Contains(t, toolNames, "ask_user_question")
	assert.Contains(t, toolNames, "todo_write")

	// Phase 4B tools
	assert.Contains(t, toolNames, "web_fetch")
	assert.Contains(t, toolNames, "web_search")

	// Phase 4C tools
	assert.Contains(t, toolNames, "task")
	assert.Contains(t, toolNames, "bash_output")
	assert.Contains(t, toolNames, "kill_shell")
}
