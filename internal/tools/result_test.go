// ABOUTME: Tests for tool execution results
// ABOUTME: Validates result creation and metadata handling

package tools_test

import (
	"testing"

	"github.com/harper/pagent/internal/tools"
	"github.com/stretchr/testify/assert"
)

func TestResult_Success(t *testing.T) {
	result := &tools.Result{
		ToolName: "read_file",
		Success:  true,
		Output:   "file contents",
		Metadata: map[string]interface{}{
			"path": "/tmp/test.txt",
			"size": 1024,
		},
	}

	assert.True(t, result.Success)
	assert.Equal(t, "read_file", result.ToolName)
	assert.Equal(t, "file contents", result.Output)
	assert.Equal(t, "", result.Error)
	assert.Equal(t, "/tmp/test.txt", result.Metadata["path"])
	assert.Equal(t, 1024, result.Metadata["size"])
}

func TestResult_Error(t *testing.T) {
	result := &tools.Result{
		ToolName: "write_file",
		Success:  false,
		Error:    "permission denied",
		Metadata: map[string]interface{}{
			"path": "/etc/passwd",
		},
	}

	assert.False(t, result.Success)
	assert.Equal(t, "write_file", result.ToolName)
	assert.Equal(t, "permission denied", result.Error)
	assert.Equal(t, "", result.Output)
}

func TestResult_WithoutMetadata(t *testing.T) {
	result := &tools.Result{
		ToolName: "bash",
		Success:  true,
		Output:   "command output",
	}

	assert.Nil(t, result.Metadata)
	assert.True(t, result.Success)
}
