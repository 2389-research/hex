// ABOUTME: Tests for API integration types
// ABOUTME: Validates ToolUse and ToolResult structures for API communication

package tools_test

import (
	"encoding/json"
	"testing"

	"github.com/harper/pagent/internal/tools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToolUse_JSON(t *testing.T) {
	toolUse := tools.ToolUse{
		Type: "tool_use",
		ID:   "toolu_123",
		Name: "read_file",
		Input: map[string]interface{}{
			"path": "/tmp/test.txt",
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(toolUse)
	require.NoError(t, err)

	// Unmarshal back
	var decoded tools.ToolUse
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "tool_use", decoded.Type)
	assert.Equal(t, "toolu_123", decoded.ID)
	assert.Equal(t, "read_file", decoded.Name)
	assert.Equal(t, "/tmp/test.txt", decoded.Input["path"])
}

func TestToolUse_FromJSON(t *testing.T) {
	jsonData := `{
		"type": "tool_use",
		"id": "toolu_456",
		"name": "write_file",
		"input": {
			"path": "/tmp/output.txt",
			"content": "Hello, World!"
		}
	}`

	var toolUse tools.ToolUse
	err := json.Unmarshal([]byte(jsonData), &toolUse)
	require.NoError(t, err)

	assert.Equal(t, "tool_use", toolUse.Type)
	assert.Equal(t, "toolu_456", toolUse.ID)
	assert.Equal(t, "write_file", toolUse.Name)
	assert.Equal(t, "/tmp/output.txt", toolUse.Input["path"])
	assert.Equal(t, "Hello, World!", toolUse.Input["content"])
}

func TestToolResult_JSON_Success(t *testing.T) {
	toolResult := tools.ToolResult{
		Type:      "tool_result",
		ToolUseID: "toolu_123",
		Content:   "file contents here",
		IsError:   false,
	}

	// Marshal to JSON
	data, err := json.Marshal(toolResult)
	require.NoError(t, err)

	// Unmarshal back
	var decoded tools.ToolResult
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "tool_result", decoded.Type)
	assert.Equal(t, "toolu_123", decoded.ToolUseID)
	assert.Equal(t, "file contents here", decoded.Content)
	assert.False(t, decoded.IsError)
}

func TestToolResult_JSON_Error(t *testing.T) {
	toolResult := tools.ToolResult{
		Type:      "tool_result",
		ToolUseID: "toolu_456",
		Content:   "Error: file not found",
		IsError:   true,
	}

	// Marshal to JSON
	data, err := json.Marshal(toolResult)
	require.NoError(t, err)

	// Unmarshal back
	var decoded tools.ToolResult
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "tool_result", decoded.Type)
	assert.Equal(t, "toolu_456", decoded.ToolUseID)
	assert.Equal(t, "Error: file not found", decoded.Content)
	assert.True(t, decoded.IsError)
}

func TestToolResult_FromResult_Success(t *testing.T) {
	result := &tools.Result{
		ToolName: "read_file",
		Success:  true,
		Output:   "file contents",
		Metadata: map[string]interface{}{
			"path": "/tmp/test.txt",
		},
	}

	toolResult := tools.ResultToToolResult(result, "toolu_123")

	assert.Equal(t, "tool_result", toolResult.Type)
	assert.Equal(t, "toolu_123", toolResult.ToolUseID)
	assert.Equal(t, "file contents", toolResult.Content)
	assert.False(t, toolResult.IsError)
}

func TestToolResult_FromResult_Error(t *testing.T) {
	result := &tools.Result{
		ToolName: "write_file",
		Success:  false,
		Error:    "permission denied",
	}

	toolResult := tools.ResultToToolResult(result, "toolu_456")

	assert.Equal(t, "tool_result", toolResult.Type)
	assert.Equal(t, "toolu_456", toolResult.ToolUseID)
	assert.Equal(t, "Error: permission denied", toolResult.Content)
	assert.True(t, toolResult.IsError)
}

func TestToolResult_FromResult_EmptyOutput(t *testing.T) {
	result := &tools.Result{
		ToolName: "bash",
		Success:  true,
		Output:   "",
	}

	toolResult := tools.ResultToToolResult(result, "toolu_789")

	assert.Equal(t, "tool_result", toolResult.Type)
	assert.Equal(t, "toolu_789", toolResult.ToolUseID)
	assert.Equal(t, "(no output)", toolResult.Content)
	assert.False(t, toolResult.IsError)
}
