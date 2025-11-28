// ABOUTME: Tests for Task tool subprocess execution
// ABOUTME: Validates sub-agent spawning, parameter handling, output capture, error handling

package tools_test

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/harper/clem/internal/tools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskTool_Name(t *testing.T) {
	tool := tools.NewTaskTool()
	assert.Equal(t, "task", tool.Name())
}

func TestTaskTool_Description(t *testing.T) {
	tool := tools.NewTaskTool()
	desc := tool.Description()
	assert.NotEmpty(t, desc)
	assert.Contains(t, desc, "sub-agent")
}

func TestTaskTool_RequiresApproval_Always(t *testing.T) {
	tool := tools.NewTaskTool()

	tests := []struct {
		name   string
		params map[string]interface{}
	}{
		{
			name: "simple task",
			params: map[string]interface{}{
				"prompt":        "Say hello",
				"description":   "greeting task",
				"subagent_type": "general-purpose",
			},
		},
		{
			name: "task with model",
			params: map[string]interface{}{
				"prompt":        "Complex analysis",
				"description":   "analysis task",
				"subagent_type": "general-purpose",
				"model":         "claude-sonnet-4",
			},
		},
		{
			name: "empty params",
			params: map[string]interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Task tool ALWAYS requires approval (spawns processes, uses API)
			assert.True(t, tool.RequiresApproval(tt.params))
		})
	}
}

func TestTaskTool_Execute_MissingPrompt(t *testing.T) {
	tool := tools.NewTaskTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"description":   "test task",
		"subagent_type": "general-purpose",
	})
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, strings.ToLower(result.Error), "prompt")
}

func TestTaskTool_Execute_EmptyPrompt(t *testing.T) {
	tool := tools.NewTaskTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"prompt":        "",
		"description":   "test task",
		"subagent_type": "general-purpose",
	})
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, strings.ToLower(result.Error), "prompt")
}

func TestTaskTool_Execute_InvalidPromptType(t *testing.T) {
	tool := tools.NewTaskTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"prompt":        123, // Invalid type
		"description":   "test task",
		"subagent_type": "general-purpose",
	})
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, strings.ToLower(result.Error), "prompt")
}

func TestTaskTool_Execute_MissingDescription(t *testing.T) {
	tool := tools.NewTaskTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"prompt":        "Say hello",
		"subagent_type": "general-purpose",
	})
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, strings.ToLower(result.Error), "description")
}

func TestTaskTool_Execute_EmptyDescription(t *testing.T) {
	tool := tools.NewTaskTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"prompt":        "Say hello",
		"description":   "",
		"subagent_type": "general-purpose",
	})
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, strings.ToLower(result.Error), "description")
}

func TestTaskTool_Execute_InvalidDescriptionType(t *testing.T) {
	tool := tools.NewTaskTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"prompt":        "Say hello",
		"description":   456, // Invalid type
		"subagent_type": "general-purpose",
	})
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, strings.ToLower(result.Error), "description")
}

func TestTaskTool_Execute_MissingSubagentType(t *testing.T) {
	tool := tools.NewTaskTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"prompt":      "Say hello",
		"description": "greeting task",
	})
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, strings.ToLower(result.Error), "subagent_type")
}

func TestTaskTool_Execute_EmptySubagentType(t *testing.T) {
	tool := tools.NewTaskTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"prompt":        "Say hello",
		"description":   "greeting task",
		"subagent_type": "",
	})
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, strings.ToLower(result.Error), "subagent_type")
}

func TestTaskTool_Execute_InvalidSubagentTypeType(t *testing.T) {
	tool := tools.NewTaskTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"prompt":        "Say hello",
		"description":   "greeting task",
		"subagent_type": 789, // Invalid type
	})
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, strings.ToLower(result.Error), "subagent_type")
}

func TestTaskTool_Execute_SimpleTask(t *testing.T) {
	// Skip if no API key available
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("ANTHROPIC_API_KEY not set")
	}

	tool := tools.NewTaskTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"prompt":        "Say exactly: TASK_COMPLETE",
		"description":   "simple echo test",
		"subagent_type": "general-purpose",
	})

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Contains(t, result.Output, "TASK_COMPLETE")
	assert.Equal(t, "task", result.ToolName)

	// Check metadata
	assert.Contains(t, result.Metadata, "duration")
	assert.Contains(t, result.Metadata, "prompt")
	assert.Contains(t, result.Metadata, "description")
	assert.Contains(t, result.Metadata, "subagent_type")
}

func TestTaskTool_Execute_TaskWithModel(t *testing.T) {
	// Skip if no API key available
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("ANTHROPIC_API_KEY not set")
	}

	tool := tools.NewTaskTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"prompt":        "Say exactly: MODEL_TEST",
		"description":   "model test",
		"subagent_type": "general-purpose",
		"model":         "claude-sonnet-4-5-20250929",
	})

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Contains(t, result.Output, "MODEL_TEST")
	assert.Contains(t, result.Metadata, "model")
	assert.Equal(t, "claude-sonnet-4-5-20250929", result.Metadata["model"])
}

func TestTaskTool_Execute_InvalidModelType(t *testing.T) {
	tool := tools.NewTaskTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"prompt":        "Say hello",
		"description":   "greeting task",
		"subagent_type": "general-purpose",
		"model":         123, // Invalid type
	})
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, strings.ToLower(result.Error), "model")
}

func TestTaskTool_Execute_InvalidResumeType(t *testing.T) {
	tool := tools.NewTaskTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"prompt":        "Say hello",
		"description":   "greeting task",
		"subagent_type": "general-purpose",
		"resume":        456, // Invalid type
	})
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, strings.ToLower(result.Error), "resume")
}

func TestTaskTool_Execute_SubprocessTimeout(t *testing.T) {
	// Skip if no API key available
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("ANTHROPIC_API_KEY not set")
	}

	tool := tools.NewTaskTool()

	// Create a context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	result, err := tool.Execute(ctx, map[string]interface{}{
		"prompt":        "Write a very long essay about everything",
		"description":   "timeout test",
		"subagent_type": "general-purpose",
	})

	require.NoError(t, err)
	assert.False(t, result.Success)
	// Should contain timeout or deadline or cancelled
	lowerError := strings.ToLower(result.Error)
	hasTimeoutMsg := strings.Contains(lowerError, "timeout") ||
		strings.Contains(lowerError, "deadline") ||
		strings.Contains(lowerError, "cancelled") ||
		strings.Contains(lowerError, "context")
	assert.True(t, hasTimeoutMsg, "Expected timeout/deadline/cancelled in error: %s", result.Error)
}

func TestTaskTool_Execute_SubprocessError(t *testing.T) {
	tool := tools.NewTaskTool()

	// Set an invalid binary path to force subprocess error
	originalPath := os.Getenv("PATH")
	os.Setenv("PATH", "/nonexistent")
	defer os.Setenv("PATH", originalPath)

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"prompt":        "This won't work",
		"description":   "error test",
		"subagent_type": "general-purpose",
	})

	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.NotEmpty(t, result.Error)
}

func TestTaskTool_Execute_MetadataPopulated(t *testing.T) {
	// Skip if no API key available
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("ANTHROPIC_API_KEY not set")
	}

	tool := tools.NewTaskTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"prompt":        "Say: METADATA_TEST",
		"description":   "metadata validation",
		"subagent_type": "general-purpose",
		"model":         "claude-sonnet-4-5-20250929",
	})

	require.NoError(t, err)
	assert.True(t, result.Success)

	// Check all expected metadata fields
	assert.Contains(t, result.Metadata, "prompt")
	assert.Contains(t, result.Metadata, "description")
	assert.Contains(t, result.Metadata, "subagent_type")
	assert.Contains(t, result.Metadata, "model")
	assert.Contains(t, result.Metadata, "duration")

	assert.Equal(t, "Say: METADATA_TEST", result.Metadata["prompt"])
	assert.Equal(t, "metadata validation", result.Metadata["description"])
	assert.Equal(t, "general-purpose", result.Metadata["subagent_type"])
	assert.Equal(t, "claude-sonnet-4-5-20250929", result.Metadata["model"])
}

func TestTaskTool_Execute_OutputCaptured(t *testing.T) {
	// Skip if no API key available
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("ANTHROPIC_API_KEY not set")
	}

	tool := tools.NewTaskTool()

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"prompt":        "Write exactly three lines: LINE1, LINE2, LINE3",
		"description":   "output test",
		"subagent_type": "general-purpose",
	})

	require.NoError(t, err)
	assert.True(t, result.Success)

	// Output should contain the response
	assert.NotEmpty(t, result.Output)

	// Should contain at least some of the requested output
	output := strings.ToUpper(result.Output)
	hasLines := strings.Contains(output, "LINE1") ||
		strings.Contains(output, "LINE2") ||
		strings.Contains(output, "LINE3")
	assert.True(t, hasLines, "Expected output to contain LINE1/LINE2/LINE3")
}

func TestTaskTool_Execute_ContextCancellation(t *testing.T) {
	// Skip if no API key available
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("ANTHROPIC_API_KEY not set")
	}

	tool := tools.NewTaskTool()

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel context immediately
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	result, err := tool.Execute(ctx, map[string]interface{}{
		"prompt":        "Write a very long detailed analysis",
		"description":   "cancellation test",
		"subagent_type": "general-purpose",
	})

	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.NotEmpty(t, result.Error)
}
