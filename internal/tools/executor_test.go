// ABOUTME: Tests for tool executor with permission management
// ABOUTME: Validates tool execution, approval flow, and error handling

package tools_test

import (
	"context"
	"errors"
	"testing"

	"github.com/harper/pagent/internal/tools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewExecutor(t *testing.T) {
	registry := tools.NewRegistry()
	executor := tools.NewExecutor(registry, nil)

	assert.NotNil(t, executor)
}

func TestExecutor_Execute_NoApprovalNeeded(t *testing.T) {
	registry := tools.NewRegistry()

	tool := &tools.MockTool{
		NameValue:             "safe_tool",
		RequiresApprovalValue: false,
		ExecuteFunc: func(_ context.Context, _ map[string]interface{}) (*tools.Result, error) {
			return &tools.Result{
				ToolName: "safe_tool",
				Success:  true,
				Output:   "executed successfully",
			}, nil
		},
	}

	require.NoError(t, registry.Register(tool))

	executor := tools.NewExecutor(registry, nil)
	ctx := context.Background()
	params := map[string]interface{}{"foo": "bar"}

	result, err := executor.Execute(ctx, "safe_tool", params)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, "executed successfully", result.Output)
}

func TestExecutor_Execute_ApprovalNeeded_Approved(t *testing.T) {
	registry := tools.NewRegistry()

	tool := &tools.MockTool{
		NameValue:             "dangerous_tool",
		RequiresApprovalValue: true,
		ExecuteFunc: func(_ context.Context, _ map[string]interface{}) (*tools.Result, error) {
			return &tools.Result{
				ToolName: "dangerous_tool",
				Success:  true,
				Output:   "executed with approval",
			}, nil
		},
	}

	require.NoError(t, registry.Register(tool))

	// Approval function that always approves
	approvalFunc := func(_ string, _ map[string]interface{}) bool {
		return true
	}

	executor := tools.NewExecutor(registry, approvalFunc)
	ctx := context.Background()
	params := map[string]interface{}{"path": "/etc/passwd"}

	result, err := executor.Execute(ctx, "dangerous_tool", params)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, "executed with approval", result.Output)
}

func TestExecutor_Execute_ApprovalNeeded_Denied(t *testing.T) {
	registry := tools.NewRegistry()

	tool := &tools.MockTool{
		NameValue:             "dangerous_tool",
		RequiresApprovalValue: true,
		ExecuteFunc: func(_ context.Context, _ map[string]interface{}) (*tools.Result, error) {
			// Should not be called
			t.Fatal("Execute should not be called when approval is denied")
			return nil, nil
		},
	}

	require.NoError(t, registry.Register(tool))

	// Approval function that always denies
	approvalFunc := func(_ string, _ map[string]interface{}) bool {
		return false
	}

	executor := tools.NewExecutor(registry, approvalFunc)
	ctx := context.Background()
	params := map[string]interface{}{"path": "/etc/passwd"}

	result, err := executor.Execute(ctx, "dangerous_tool", params)
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "user denied permission")
}

func TestExecutor_Execute_ApprovalNeeded_NoApprovalFunc(t *testing.T) {
	registry := tools.NewRegistry()

	tool := &tools.MockTool{
		NameValue:             "dangerous_tool",
		RequiresApprovalValue: true,
		ExecuteFunc: func(_ context.Context, _ map[string]interface{}) (*tools.Result, error) {
			return &tools.Result{
				ToolName: "dangerous_tool",
				Success:  true,
				Output:   "executed without approval check",
			}, nil
		},
	}

	require.NoError(t, registry.Register(tool))

	// No approval function provided (nil)
	executor := tools.NewExecutor(registry, nil)
	ctx := context.Background()
	params := map[string]interface{}{"path": "/etc/passwd"}

	// Should execute without checking approval
	result, err := executor.Execute(ctx, "dangerous_tool", params)
	require.NoError(t, err)
	assert.True(t, result.Success)
}

func TestExecutor_Execute_NonExistentTool(t *testing.T) {
	registry := tools.NewRegistry()
	executor := tools.NewExecutor(registry, nil)

	ctx := context.Background()
	params := map[string]interface{}{}

	result, err := executor.Execute(ctx, "nonexistent", params)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "get tool")
}

func TestExecutor_Execute_ToolReturnsError(t *testing.T) {
	registry := tools.NewRegistry()

	expectedErr := errors.New("tool execution failed")
	tool := &tools.MockTool{
		NameValue: "error_tool",
		ExecuteFunc: func(_ context.Context, _ map[string]interface{}) (*tools.Result, error) {
			return nil, expectedErr
		},
	}

	require.NoError(t, registry.Register(tool))

	executor := tools.NewExecutor(registry, nil)
	ctx := context.Background()
	params := map[string]interface{}{}

	result, err := executor.Execute(ctx, "error_tool", params)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "execute tool")
}

func TestExecutor_Execute_ContextCancellation(t *testing.T) {
	registry := tools.NewRegistry()

	tool := &tools.MockTool{
		NameValue: "slow_tool",
		ExecuteFunc: func(ctx context.Context, _ map[string]interface{}) (*tools.Result, error) {
			// Check if context is cancelled
			select {
			case <-ctx.Done():
				return &tools.Result{
					ToolName: "slow_tool",
					Success:  false,
					Error:    ctx.Err().Error(),
				}, nil
			default:
				return &tools.Result{
					ToolName: "slow_tool",
					Success:  true,
					Output:   "completed",
				}, nil
			}
		},
	}

	require.NoError(t, registry.Register(tool))

	executor := tools.NewExecutor(registry, nil)

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	params := map[string]interface{}{}

	result, err := executor.Execute(ctx, "slow_tool", params)
	require.NoError(t, err)
	assert.False(t, result.Success)
}

func TestExecutor_Execute_ApprovalFuncReceivesCorrectParams(t *testing.T) {
	registry := tools.NewRegistry()

	tool := &tools.MockTool{
		NameValue:             "param_tool",
		RequiresApprovalValue: true,
		ExecuteFunc: func(_ context.Context, _ map[string]interface{}) (*tools.Result, error) {
			return &tools.Result{
				ToolName: "param_tool",
				Success:  true,
				Output:   "executed",
			}, nil
		},
	}

	require.NoError(t, registry.Register(tool))

	// Track what was passed to approval function
	var receivedToolName string
	var receivedParams map[string]interface{}

	approvalFunc := func(toolName string, params map[string]interface{}) bool {
		receivedToolName = toolName
		receivedParams = params
		return true
	}

	executor := tools.NewExecutor(registry, approvalFunc)
	ctx := context.Background()
	params := map[string]interface{}{
		"path":   "/tmp/test.txt",
		"action": "write",
	}

	_, err := executor.Execute(ctx, "param_tool", params)
	require.NoError(t, err)

	assert.Equal(t, "param_tool", receivedToolName)
	assert.Equal(t, "/tmp/test.txt", receivedParams["path"])
	assert.Equal(t, "write", receivedParams["action"])
}

func TestExecutor_Execute_NilParams(t *testing.T) {
	registry := tools.NewRegistry()

	tool := &tools.MockTool{
		NameValue:             "test_tool",
		RequiresApprovalValue: false,
		ExecuteFunc: func(_ context.Context, params map[string]interface{}) (*tools.Result, error) {
			// Verify params is not nil
			assert.NotNil(t, params)
			assert.Empty(t, params)
			return &tools.Result{
				ToolName: "test_tool",
				Success:  true,
				Output:   "executed with empty params",
			}, nil
		},
	}

	require.NoError(t, registry.Register(tool))

	executor := tools.NewExecutor(registry, nil)

	// Execute with nil params
	result, err := executor.Execute(context.Background(), "test_tool", nil)

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, "executed with empty params", result.Output)
}
