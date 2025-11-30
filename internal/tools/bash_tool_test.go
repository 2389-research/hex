// ABOUTME: Tests for Bash tool execution
// ABOUTME: Validates command execution, timeout, output capture, error handling

package tools_test

import (
	"context"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/harper/clem/internal/tools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBashTool_Name(t *testing.T) {
	tool := tools.NewBashTool()
	assert.Equal(t, "bash", tool.Name())
}

func TestBashTool_Description(t *testing.T) {
	tool := tools.NewBashTool()
	desc := tool.Description()
	assert.NotEmpty(t, desc)
	assert.Contains(t, desc, "command")
}

func TestBashTool_RequiresApproval_Always(t *testing.T) {
	tool := tools.NewBashTool()

	tests := []struct {
		name   string
		params map[string]interface{}
	}{
		{
			name:   "simple echo command",
			params: map[string]interface{}{"command": "echo hello"},
		},
		{
			name:   "dangerous command",
			params: map[string]interface{}{"command": "rm -rf /"},
		},
		{
			name:   "empty command",
			params: map[string]interface{}{"command": ""},
		},
		{
			name:   "no command",
			params: map[string]interface{}{},
		},
		{
			name:   "background command",
			params: map[string]interface{}{"command": "echo test", "run_in_background": true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ALWAYS require approval for command execution
			assert.True(t, tool.RequiresApproval(tt.params))
		})
	}
}

func TestBashTool_Execute_MissingCommand(t *testing.T) {
	tool := tools.NewBashTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{})
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "command")
}

func TestBashTool_Execute_EmptyCommand(t *testing.T) {
	tool := tools.NewBashTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"command": "",
	})
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "command")
}

func TestBashTool_Execute_InvalidCommandType(t *testing.T) {
	tool := tools.NewBashTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"command": 123, // Invalid type
	})
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "command")
}

func TestBashTool_Execute_SimpleCommand(t *testing.T) {
	tool := tools.NewBashTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"command": "echo 'Hello, World!'",
	})

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Contains(t, result.Output, "Hello, World!")
	assert.Equal(t, "bash", result.ToolName)
	assert.Empty(t, result.Error)
	assert.Equal(t, 0, result.Metadata["exit_code"])
	assert.NotNil(t, result.Metadata["duration"])
	assert.Equal(t, "echo 'Hello, World!'", result.Metadata["command"])
}

func TestBashTool_Execute_StdoutOnly(t *testing.T) {
	tool := tools.NewBashTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"command": "echo 'stdout test'",
	})

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Contains(t, result.Output, "STDOUT:")
	assert.Contains(t, result.Output, "stdout test")
	assert.NotContains(t, result.Output, "STDERR:")
}

func TestBashTool_Execute_StderrOnly(t *testing.T) {
	tool := tools.NewBashTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"command": "echo 'stderr test' >&2",
	})

	require.NoError(t, err)
	// Command succeeds (exit 0) but output goes to stderr
	assert.True(t, result.Success)
	assert.Contains(t, result.Output, "STDERR:")
	assert.Contains(t, result.Output, "stderr test")
}

func TestBashTool_Execute_BothStdoutAndStderr(t *testing.T) {
	tool := tools.NewBashTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"command": "echo 'to stdout'; echo 'to stderr' >&2",
	})

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Contains(t, result.Output, "STDOUT:")
	assert.Contains(t, result.Output, "to stdout")
	assert.Contains(t, result.Output, "STDERR:")
	assert.Contains(t, result.Output, "to stderr")
}

func TestBashTool_Execute_NoOutput(t *testing.T) {
	tool := tools.NewBashTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"command": "true", // Command that produces no output
	})

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Contains(t, result.Output, "(no output)")
}

func TestBashTool_Execute_MultilineOutput(t *testing.T) {
	tool := tools.NewBashTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"command": "echo 'line1'; echo 'line2'; echo 'line3'",
	})

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Contains(t, result.Output, "line1")
	assert.Contains(t, result.Output, "line2")
	assert.Contains(t, result.Output, "line3")
	assert.Equal(t, 3, result.Metadata["stdout_lines"])
}

func TestBashTool_Execute_NonZeroExitCode(t *testing.T) {
	tool := tools.NewBashTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"command": "exit 42",
	})

	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "42")
	assert.Equal(t, 42, result.Metadata["exit_code"])
}

func TestBashTool_Execute_CommandNotFound(t *testing.T) {
	tool := tools.NewBashTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"command": "nonexistent_command_xyz",
	})

	require.NoError(t, err)
	assert.False(t, result.Success)
	// Exit code 127 is standard for "command not found"
	assert.Equal(t, 127, result.Metadata["exit_code"])
}

func TestBashTool_Execute_CommandFails(t *testing.T) {
	tool := tools.NewBashTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"command": "ls /nonexistent_dir_xyz",
	})

	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.NotEqual(t, 0, result.Metadata["exit_code"])
	// Error should be in stderr
	assert.Contains(t, result.Output, "STDERR:")
}

func TestBashTool_Execute_Timeout(t *testing.T) {
	tool := tools.NewBashTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"command": "sleep 10",
		"timeout": float64(1), // 1 second timeout
	})

	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "timed out")
	assert.Equal(t, float64(1), result.Metadata["timeout"])
}

func TestBashTool_Execute_DefaultTimeout(t *testing.T) {
	tool := tools.NewBashTool()

	// Command that completes quickly should succeed with default timeout
	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"command": "echo 'quick'",
	})

	require.NoError(t, err)
	assert.True(t, result.Success)

	// Duration should be much less than default timeout (30s)
	duration := result.Metadata["duration"].(float64)
	assert.Less(t, duration, 5.0)
}

func TestBashTool_Execute_CustomTimeout(t *testing.T) {
	tool := tools.NewBashTool()

	// Sleep for 2 seconds with 3 second timeout should succeed
	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"command": "sleep 2",
		"timeout": float64(3),
	})

	require.NoError(t, err)
	assert.True(t, result.Success)

	// Duration should be around 2 seconds
	duration := result.Metadata["duration"].(float64)
	assert.Greater(t, duration, 1.5)
	assert.Less(t, duration, 3.0)
}

func TestBashTool_Execute_MaxTimeoutCap(t *testing.T) {
	tool := tools.NewBashTool()

	// Request 10 minute timeout, should be capped at 5 minutes
	// We'll verify by checking that a quick command still succeeds
	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"command": "echo 'test'",
		"timeout": float64(600), // 10 minutes
	})

	require.NoError(t, err)
	assert.True(t, result.Success)
	// The timeout cap is internal, we just verify command executes normally
}

func TestBashTool_Execute_WorkingDirectory(t *testing.T) {
	tool := tools.NewBashTool()

	// Use a known directory that exists on all platforms
	var workingDir string
	if runtime.GOOS == "windows" {
		workingDir = "C:\\Windows"
	} else {
		workingDir = "/tmp"
	}

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"command":     "pwd",
		"working_dir": workingDir,
	})

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Contains(t, result.Output, workingDir)
	assert.Equal(t, workingDir, result.Metadata["working_dir"])
}

func TestBashTool_Execute_InvalidWorkingDirectory(t *testing.T) {
	tool := tools.NewBashTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"command":     "echo test",
		"working_dir": "/nonexistent/directory/xyz",
	})

	require.NoError(t, err)
	// The command execution will fail due to invalid directory
	assert.False(t, result.Success)
}

func TestBashTool_Execute_OutputTooLarge(t *testing.T) {
	tool := tools.NewBashTool()

	// Generate output larger than 1MB (default max)
	// Using printf to repeat a string many times
	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"command": "yes 'a' | head -c 2000000", // 2MB of output
	})

	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "too large")
}

func TestBashTool_Execute_ContextCancellation(t *testing.T) {
	tool := tools.NewBashTool()

	ctx, cancel := context.WithCancel(context.Background())

	// Start a long-running command
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	result, err := tool.Execute(ctx, map[string]interface{}{
		"command": "sleep 10",
	})

	require.NoError(t, err)
	assert.False(t, result.Success)
	// Should timeout or be cancelled
	assert.NotEmpty(t, result.Error)
}

func TestBashTool_Execute_MetadataPopulated(t *testing.T) {
	tool := tools.NewBashTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"command": "echo 'line1'; echo 'line2'",
	})

	require.NoError(t, err)
	assert.True(t, result.Success)

	// Check all expected metadata fields
	assert.Contains(t, result.Metadata, "exit_code")
	assert.Contains(t, result.Metadata, "duration")
	assert.Contains(t, result.Metadata, "command")
	assert.Contains(t, result.Metadata, "working_dir")
	assert.Contains(t, result.Metadata, "stdout_lines")
	assert.Contains(t, result.Metadata, "stderr_lines")

	assert.Equal(t, 0, result.Metadata["exit_code"])
	assert.Equal(t, "echo 'line1'; echo 'line2'", result.Metadata["command"])
	assert.Equal(t, 2, result.Metadata["stdout_lines"])
}

func TestBashTool_Execute_ComplexCommand(t *testing.T) {
	tool := tools.NewBashTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"command": "for i in 1 2 3; do echo \"Number: $i\"; done",
	})

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Contains(t, result.Output, "Number: 1")
	assert.Contains(t, result.Output, "Number: 2")
	assert.Contains(t, result.Output, "Number: 3")
}

func TestBashTool_Execute_PipesAndRedirection(t *testing.T) {
	tool := tools.NewBashTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"command": "echo 'hello world' | grep 'hello' | wc -w",
	})

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Contains(t, strings.TrimSpace(result.Output), "2") // "hello world" is 2 words
}

func TestBashTool_Execute_EnvironmentVariables(t *testing.T) {
	tool := tools.NewBashTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"command": "echo \"User: $USER, Home: $HOME\"",
	})

	require.NoError(t, err)
	assert.True(t, result.Success)
	// Should have environment variables available
	assert.Contains(t, result.Output, "User:")
	assert.Contains(t, result.Output, "Home:")
}

// ========================================
// NEW TESTS FOR run_in_background FEATURE
// ========================================

func TestBashTool_Execute_Background_InvalidRunInBackgroundType(t *testing.T) {
	tool := tools.NewBashTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"command":           "echo 'test'",
		"run_in_background": "not a boolean", // Invalid type
	})

	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "run_in_background")
	assert.Contains(t, result.Error, "boolean")
}

func TestBashTool_Execute_Background_RunInBackgroundFalse(t *testing.T) {
	tool := tools.NewBashTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"command":           "echo 'sync test'",
		"run_in_background": false,
	})

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Contains(t, result.Output, "sync test")
	// Should NOT have bash_id in metadata (synchronous execution)
	assert.NotContains(t, result.Metadata, "bash_id")
}

func TestBashTool_Execute_Background_LaunchProcess(t *testing.T) {
	tool := tools.NewBashTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"command":           "echo 'background test'; sleep 1; echo 'done'",
		"run_in_background": true,
	})

	require.NoError(t, err)
	assert.True(t, result.Success)

	// Should return a bash_id in metadata
	bashID, ok := result.Metadata["bash_id"].(string)
	assert.True(t, ok, "bash_id should be a string")
	assert.NotEmpty(t, bashID)

	// Should have status in metadata
	status, ok := result.Metadata["status"].(string)
	assert.True(t, ok)
	assert.Equal(t, "running", status)

	// Output should contain the bash_id
	assert.Contains(t, result.Output, bashID)
	assert.Contains(t, result.Output, "Background process started")

	// Verify process is registered
	proc, err := tools.GetBackgroundRegistry().Get(bashID)
	require.NoError(t, err)
	assert.Equal(t, "echo 'background test'; sleep 1; echo 'done'", proc.Command)

	// Clean up
	time.Sleep(2 * time.Second) // Let command finish
	_ = tools.GetBackgroundRegistry().Remove(bashID)
}

func TestBashTool_Execute_Background_ProcessIDUnique(t *testing.T) {
	tool := tools.NewBashTool()

	// Launch two background processes
	result1, err := tool.Execute(context.Background(), map[string]interface{}{
		"command":           "sleep 2",
		"run_in_background": true,
	})
	require.NoError(t, err)
	assert.True(t, result1.Success)

	result2, err := tool.Execute(context.Background(), map[string]interface{}{
		"command":           "sleep 2",
		"run_in_background": true,
	})
	require.NoError(t, err)
	assert.True(t, result2.Success)

	// IDs should be different
	bashID1 := result1.Metadata["bash_id"].(string)
	bashID2 := result2.Metadata["bash_id"].(string)
	assert.NotEqual(t, bashID1, bashID2)

	// Clean up
	time.Sleep(3 * time.Second)
	_ = tools.GetBackgroundRegistry().Remove(bashID1)
	_ = tools.GetBackgroundRegistry().Remove(bashID2)
}

func TestBashTool_Execute_Background_RetrieveOutputWithBashOutput(t *testing.T) {
	bashTool := tools.NewBashTool()
	outputTool := tools.NewBashOutputTool()

	// Launch background process
	result, err := bashTool.Execute(context.Background(), map[string]interface{}{
		"command":           "echo 'line1'; sleep 0.5; echo 'line2'; sleep 0.5; echo 'line3'",
		"run_in_background": true,
	})
	require.NoError(t, err)
	assert.True(t, result.Success)

	bashID := result.Metadata["bash_id"].(string)

	// Wait for some output
	time.Sleep(1 * time.Second)

	// Retrieve output using BashOutput tool
	outputResult, err := outputTool.Execute(context.Background(), map[string]interface{}{
		"bash_id": bashID,
	})
	require.NoError(t, err)
	assert.True(t, outputResult.Success)
	assert.Contains(t, outputResult.Output, "line1")

	// Wait for process to complete
	time.Sleep(2 * time.Second)

	// Retrieve remaining output
	outputResult2, err := outputTool.Execute(context.Background(), map[string]interface{}{
		"bash_id": bashID,
	})
	require.NoError(t, err)
	assert.True(t, outputResult2.Success)

	// Should have done=true in metadata
	done, ok := outputResult2.Metadata["done"].(bool)
	assert.True(t, ok)
	assert.True(t, done)

	// Clean up
	_ = tools.GetBackgroundRegistry().Remove(bashID)
}

func TestBashTool_Execute_Background_ProcessCompletesSuccessfully(t *testing.T) {
	bashTool := tools.NewBashTool()
	outputTool := tools.NewBashOutputTool()

	// Launch quick background process
	result, err := bashTool.Execute(context.Background(), map[string]interface{}{
		"command":           "echo 'test' && exit 0",
		"run_in_background": true,
	})
	require.NoError(t, err)
	assert.True(t, result.Success)

	bashID := result.Metadata["bash_id"].(string)

	// Wait for completion
	time.Sleep(1 * time.Second)

	// Check process status
	outputResult, err := outputTool.Execute(context.Background(), map[string]interface{}{
		"bash_id": bashID,
	})
	require.NoError(t, err)
	assert.True(t, outputResult.Success)

	done, _ := outputResult.Metadata["done"].(bool)
	assert.True(t, done)

	exitCode, _ := outputResult.Metadata["exit_code"].(int)
	assert.Equal(t, 0, exitCode)

	// Clean up
	_ = tools.GetBackgroundRegistry().Remove(bashID)
}

func TestBashTool_Execute_Background_ProcessFailsWithNonZeroExit(t *testing.T) {
	bashTool := tools.NewBashTool()
	outputTool := tools.NewBashOutputTool()

	// Launch background process that fails
	result, err := bashTool.Execute(context.Background(), map[string]interface{}{
		"command":           "exit 42",
		"run_in_background": true,
	})
	require.NoError(t, err)
	assert.True(t, result.Success) // Launch succeeds

	bashID := result.Metadata["bash_id"].(string)

	// Wait for completion
	time.Sleep(1 * time.Second)

	// Check process exit code
	outputResult, err := outputTool.Execute(context.Background(), map[string]interface{}{
		"bash_id": bashID,
	})
	require.NoError(t, err)
	assert.True(t, outputResult.Success)

	done, _ := outputResult.Metadata["done"].(bool)
	assert.True(t, done)

	exitCode, _ := outputResult.Metadata["exit_code"].(int)
	assert.Equal(t, 42, exitCode)

	// Clean up
	_ = tools.GetBackgroundRegistry().Remove(bashID)
}

func TestBashTool_Execute_Background_IncrementalOutputRetrieval(t *testing.T) {
	bashTool := tools.NewBashTool()
	outputTool := tools.NewBashOutputTool()

	// Launch background process that outputs over time
	result, err := bashTool.Execute(context.Background(), map[string]interface{}{
		"command":           "echo 'first'; sleep 0.5; echo 'second'; sleep 0.5; echo 'third'",
		"run_in_background": true,
	})
	require.NoError(t, err)
	assert.True(t, result.Success)

	bashID := result.Metadata["bash_id"].(string)

	// First read - should get early output
	time.Sleep(700 * time.Millisecond)
	output1, err := outputTool.Execute(context.Background(), map[string]interface{}{
		"bash_id": bashID,
	})
	require.NoError(t, err)
	assert.Contains(t, output1.Output, "first")

	// Second read - should get new output only
	time.Sleep(700 * time.Millisecond)
	output2, err := outputTool.Execute(context.Background(), map[string]interface{}{
		"bash_id": bashID,
	})
	require.NoError(t, err)
	// Should have new output (second/third)
	// Due to incremental reading, first should not appear again
	assert.NotContains(t, output2.Output, "first")

	// Clean up
	_ = tools.GetBackgroundRegistry().Remove(bashID)
}

func TestBashTool_Execute_Background_WithWorkingDirectory(t *testing.T) {
	bashTool := tools.NewBashTool()
	outputTool := tools.NewBashOutputTool()

	var workingDir string
	if runtime.GOOS == "windows" {
		workingDir = "C:\\Windows"
	} else {
		workingDir = "/tmp"
	}

	// Launch background process with working directory
	result, err := bashTool.Execute(context.Background(), map[string]interface{}{
		"command":           "pwd",
		"working_dir":       workingDir,
		"run_in_background": true,
	})
	require.NoError(t, err)
	assert.True(t, result.Success)

	bashID := result.Metadata["bash_id"].(string)

	// Wait and retrieve output
	time.Sleep(1 * time.Second)
	outputResult, err := outputTool.Execute(context.Background(), map[string]interface{}{
		"bash_id": bashID,
	})
	require.NoError(t, err)
	assert.True(t, outputResult.Success)
	assert.Contains(t, outputResult.Output, workingDir)

	// Clean up
	_ = tools.GetBackgroundRegistry().Remove(bashID)
}

func TestBashTool_Execute_Background_OutputCapturedCorrectly(t *testing.T) {
	bashTool := tools.NewBashTool()
	outputTool := tools.NewBashOutputTool()

	// Launch background process with both stdout and stderr
	result, err := bashTool.Execute(context.Background(), map[string]interface{}{
		"command":           "echo 'to stdout'; echo 'to stderr' >&2",
		"run_in_background": true,
	})
	require.NoError(t, err)
	assert.True(t, result.Success)

	bashID := result.Metadata["bash_id"].(string)

	// Wait and retrieve output
	time.Sleep(1 * time.Second)
	outputResult, err := outputTool.Execute(context.Background(), map[string]interface{}{
		"bash_id": bashID,
	})
	require.NoError(t, err)
	assert.True(t, outputResult.Success)
	assert.Contains(t, outputResult.Output, "to stdout")
	assert.Contains(t, outputResult.Output, "to stderr")
	assert.Contains(t, outputResult.Output, "STDOUT:")
	assert.Contains(t, outputResult.Output, "STDERR:")

	// Clean up
	_ = tools.GetBackgroundRegistry().Remove(bashID)
}
