// ABOUTME: Tests for the Tool interface and basic tool implementations
// ABOUTME: Validates tool interface contract using mock implementations

package tools_test

import (
	"context"
	"testing"

	"github.com/harper/jeff/internal/tools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockTool_Interface(_ *testing.T) {
	// Verify MockTool implements Tool interface
	var _ tools.Tool = (*tools.MockTool)(nil)
}

func TestMockTool_Name(t *testing.T) {
	mock := &tools.MockTool{
		NameValue: "test_tool",
	}

	assert.Equal(t, "test_tool", mock.Name())
}

func TestMockTool_Description(t *testing.T) {
	mock := &tools.MockTool{
		DescriptionValue: "A test tool",
	}

	assert.Equal(t, "A test tool", mock.Description())
}

func TestMockTool_RequiresApproval(t *testing.T) {
	tests := []struct {
		name     string
		tool     *tools.MockTool
		params   map[string]interface{}
		expected bool
	}{
		{
			name: "requires approval",
			tool: &tools.MockTool{
				RequiresApprovalValue: true,
			},
			params:   map[string]interface{}{"path": "/etc/passwd"},
			expected: true,
		},
		{
			name: "does not require approval",
			tool: &tools.MockTool{
				RequiresApprovalValue: false,
			},
			params:   map[string]interface{}{"path": "/tmp/test.txt"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.tool.RequiresApproval(tt.params)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMockTool_Execute_Success(t *testing.T) {
	mock := &tools.MockTool{
		NameValue: "test_tool",
		ExecuteFunc: func(_ context.Context, _ map[string]interface{}) (*tools.Result, error) {
			return &tools.Result{
				ToolName: "test_tool",
				Success:  true,
				Output:   "test output",
			}, nil
		},
	}

	ctx := context.Background()
	params := map[string]interface{}{"foo": "bar"}

	result, err := mock.Execute(ctx, params)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, "test output", result.Output)
	assert.Equal(t, "test_tool", result.ToolName)
}

func TestMockTool_Execute_Error(t *testing.T) {
	mock := &tools.MockTool{
		NameValue: "test_tool",
		ExecuteFunc: func(_ context.Context, _ map[string]interface{}) (*tools.Result, error) {
			return &tools.Result{
				ToolName: "test_tool",
				Success:  false,
				Error:    "execution failed",
			}, nil
		},
	}

	ctx := context.Background()
	params := map[string]interface{}{"foo": "bar"}

	result, err := mock.Execute(ctx, params)
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Equal(t, "execution failed", result.Error)
}

func TestMockTool_Execute_DefaultBehavior(t *testing.T) {
	// When ExecuteFunc is nil, should return default success result
	mock := &tools.MockTool{
		NameValue: "test_tool",
	}

	ctx := context.Background()
	params := map[string]interface{}{}

	result, err := mock.Execute(ctx, params)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, "mock output", result.Output)
	assert.Equal(t, "test_tool", result.ToolName)
}
