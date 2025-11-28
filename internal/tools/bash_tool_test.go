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
