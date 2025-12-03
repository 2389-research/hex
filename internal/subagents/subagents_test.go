// ABOUTME: Comprehensive tests for subagent framework
// ABOUTME: Tests types, context isolation, execution engine, and parallel dispatch

package subagents_test

import (
	"context"
	"testing"
	"time"

	"github.com/harper/hex/internal/subagents"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ========== Type Tests ==========

func TestSubagentTypes_ValidTypes(t *testing.T) {
	validTypes := subagents.ValidSubagentTypes()
	assert.Len(t, validTypes, 4, "Should have 4 subagent types")

	expectedTypes := []string{
		"general-purpose",
		"Explore",
		"Plan",
		"code-reviewer",
	}

	for _, expected := range expectedTypes {
		assert.Contains(t, validTypes, expected)
	}
}

func TestSubagentTypes_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		typeStr  string
		expected bool
	}{
		{"general-purpose valid", "general-purpose", true},
		{"Explore valid", "Explore", true},
		{"Plan valid", "Plan", true},
		{"code-reviewer valid", "code-reviewer", true},
		{"invalid type", "invalid-type", false},
		{"empty string", "", false},
		{"wrong case", "General-Purpose", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := subagents.IsValid(tt.typeStr)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDefaultConfig_GeneralPurpose(t *testing.T) {
	config := subagents.DefaultConfig(subagents.TypeGeneralPurpose)

	assert.Equal(t, subagents.TypeGeneralPurpose, config.Type)
	assert.Equal(t, 5*time.Minute, config.Timeout)
	assert.Equal(t, 4096, config.MaxTokens)
	assert.Equal(t, 1.0, config.Temperature)
	assert.Nil(t, config.AllowedTools, "General purpose should have all tools")
}

func TestDefaultConfig_Explore(t *testing.T) {
	config := subagents.DefaultConfig(subagents.TypeExplore)

	assert.Equal(t, subagents.TypeExplore, config.Type)
	assert.Equal(t, 0.7, config.Temperature)
	assert.Equal(t, 8192, config.MaxTokens, "Explorer should have larger max tokens")

	// Check allowed tools
	assert.NotNil(t, config.AllowedTools)
	assert.Contains(t, config.AllowedTools, "Read")
	assert.Contains(t, config.AllowedTools, "Grep")
	assert.Contains(t, config.AllowedTools, "Glob")
	assert.Contains(t, config.AllowedTools, "Bash")
}

func TestDefaultConfig_Plan(t *testing.T) {
	config := subagents.DefaultConfig(subagents.TypePlan)

	assert.Equal(t, subagents.TypePlan, config.Type)
	assert.Equal(t, 0.6, config.Temperature)
	assert.Equal(t, 6144, config.MaxTokens)

	// Should be read-only (no Bash)
	assert.NotNil(t, config.AllowedTools)
	assert.Contains(t, config.AllowedTools, "Read")
	assert.Contains(t, config.AllowedTools, "Grep")
	assert.Contains(t, config.AllowedTools, "Glob")
	assert.NotContains(t, config.AllowedTools, "Bash")
}

func TestDefaultConfig_CodeReviewer(t *testing.T) {
	config := subagents.DefaultConfig(subagents.TypeCodeReviewer)

	assert.Equal(t, subagents.TypeCodeReviewer, config.Type)
	assert.Equal(t, 0.3, config.Temperature, "Reviewer should have low temperature")
	assert.Equal(t, 6144, config.MaxTokens)

	// Should be read-only
	assert.NotNil(t, config.AllowedTools)
	assert.Contains(t, config.AllowedTools, "Read")
	assert.NotContains(t, config.AllowedTools, "Edit")
	assert.NotContains(t, config.AllowedTools, "Write")
}

func TestResult_Duration(t *testing.T) {
	result := &subagents.Result{
		StartTime: time.Now(),
	}

	// Sleep a bit
	time.Sleep(10 * time.Millisecond)
	result.EndTime = time.Now()

	duration := result.Duration()
	assert.Greater(t, duration, 5*time.Millisecond)
	assert.Less(t, duration, 100*time.Millisecond)
}

func TestResult_Duration_ZeroTimes(t *testing.T) {
	result := &subagents.Result{}
	duration := result.Duration()
	assert.Equal(t, time.Duration(0), duration)
}

func TestExecutionRequest_Validate_Valid(t *testing.T) {
	req := &subagents.ExecutionRequest{
		Type:        subagents.TypeExplore,
		Prompt:      "Find all authentication code",
		Description: "Explore auth",
	}

	err := req.Validate()
	assert.NoError(t, err)
}

func TestExecutionRequest_Validate_MissingType(t *testing.T) {
	req := &subagents.ExecutionRequest{
		Prompt:      "Test",
		Description: "Test",
	}

	err := req.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "type")
}

func TestExecutionRequest_Validate_InvalidType(t *testing.T) {
	req := &subagents.ExecutionRequest{
		Type:        "invalid",
		Prompt:      "Test",
		Description: "Test",
	}

	err := req.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}

func TestExecutionRequest_Validate_MissingPrompt(t *testing.T) {
	req := &subagents.ExecutionRequest{
		Type:        subagents.TypeExplore,
		Description: "Test",
	}

	err := req.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "prompt")
}

func TestExecutionRequest_Validate_MissingDescription(t *testing.T) {
	req := &subagents.ExecutionRequest{
		Type:   subagents.TypeExplore,
		Prompt: "Test",
	}

	err := req.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "description")
}

// ========== Context Tests ==========

func TestIsolatedContext_NewIsolatedContext(t *testing.T) {
	ctx := subagents.NewIsolatedContext("parent-123", subagents.TypeExplore)

	assert.NotEmpty(t, ctx.ID)
	assert.Equal(t, "parent-123", ctx.ParentID)
	assert.Equal(t, subagents.TypeExplore, ctx.Type)
	assert.NotZero(t, ctx.CreatedAt)
	assert.Empty(t, ctx.GetMessages())
}

func TestIsolatedContext_AddMessage(t *testing.T) {
	ctx := subagents.NewIsolatedContext("parent", subagents.TypePlan)

	ctx.AddMessage("user", "Hello")
	ctx.AddMessage("assistant", "Hi there")

	messages := ctx.GetMessages()
	assert.Len(t, messages, 2)
	assert.Equal(t, "user", messages[0].Role)
	assert.Equal(t, "Hello", messages[0].Content)
	assert.Equal(t, "assistant", messages[1].Role)
	assert.Equal(t, "Hi there", messages[1].Content)
}

func TestIsolatedContext_MessageIsolation(t *testing.T) {
	ctx := subagents.NewIsolatedContext("parent", subagents.TypePlan)
	ctx.AddMessage("user", "Test")

	// Get messages and try to modify
	messages := ctx.GetMessages()
	messages[0].Content = "Modified"

	// Original should be unchanged
	originalMessages := ctx.GetMessages()
	assert.Equal(t, "Test", originalMessages[0].Content)
}

func TestIsolatedContext_WorkingMemory(t *testing.T) {
	ctx := subagents.NewIsolatedContext("parent", subagents.TypeExplore)

	ctx.SetMemory("key1", "value1")
	ctx.SetMemory("key2", 42)

	val1, ok := ctx.GetMemory("key1")
	assert.True(t, ok)
	assert.Equal(t, "value1", val1)

	val2, ok := ctx.GetMemory("key2")
	assert.True(t, ok)
	assert.Equal(t, 42, val2)

	_, ok = ctx.GetMemory("nonexistent")
	assert.False(t, ok)
}

func TestIsolatedContext_Clear(t *testing.T) {
	ctx := subagents.NewIsolatedContext("parent", subagents.TypeExplore)

	ctx.AddMessage("user", "Test")
	ctx.SetMemory("key", "value")

	ctx.Clear()

	assert.Empty(t, ctx.GetMessages())
	_, ok := ctx.GetMemory("key")
	assert.False(t, ok)
}

func TestIsolatedContext_MessageCount(t *testing.T) {
	ctx := subagents.NewIsolatedContext("parent", subagents.TypePlan)

	assert.Equal(t, 0, ctx.MessageCount())

	ctx.AddMessage("user", "Test 1")
	assert.Equal(t, 1, ctx.MessageCount())

	ctx.AddMessage("assistant", "Test 2")
	assert.Equal(t, 2, ctx.MessageCount())
}

func TestContextManager_CreateAndGetContext(t *testing.T) {
	manager := subagents.NewContextManager()

	ctx := manager.CreateContext("parent", subagents.TypeExplore)
	assert.NotNil(t, ctx)
	assert.NotEmpty(t, ctx.ID)

	retrieved, err := manager.GetContext(ctx.ID)
	assert.NoError(t, err)
	assert.Equal(t, ctx.ID, retrieved.ID)
}

func TestContextManager_GetContext_NotFound(t *testing.T) {
	manager := subagents.NewContextManager()

	_, err := manager.GetContext("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestContextManager_DeleteContext(t *testing.T) {
	manager := subagents.NewContextManager()

	ctx := manager.CreateContext("parent", subagents.TypePlan)

	err := manager.DeleteContext(ctx.ID)
	assert.NoError(t, err)

	_, err = manager.GetContext(ctx.ID)
	assert.Error(t, err)
}

func TestContextManager_DeleteContext_NotFound(t *testing.T) {
	manager := subagents.NewContextManager()

	err := manager.DeleteContext("nonexistent")
	assert.Error(t, err)
}

func TestContextManager_ListContexts(t *testing.T) {
	manager := subagents.NewContextManager()

	ctx1 := manager.CreateContext("parent", subagents.TypeExplore)
	ctx2 := manager.CreateContext("parent", subagents.TypePlan)

	ids := manager.ListContexts()
	assert.Len(t, ids, 2)
	assert.Contains(t, ids, ctx1.ID)
	assert.Contains(t, ids, ctx2.ID)
}

func TestContextManager_ContextCount(t *testing.T) {
	manager := subagents.NewContextManager()

	assert.Equal(t, 0, manager.ContextCount())

	manager.CreateContext("parent", subagents.TypeExplore)
	assert.Equal(t, 1, manager.ContextCount())

	manager.CreateContext("parent", subagents.TypePlan)
	assert.Equal(t, 2, manager.ContextCount())
}

func TestContextManager_CleanupOldContexts(t *testing.T) {
	manager := subagents.NewContextManager()

	// Create contexts
	manager.CreateContext("parent", subagents.TypeExplore)
	manager.CreateContext("parent", subagents.TypePlan)

	// Wait longer to ensure contexts are old enough
	time.Sleep(50 * time.Millisecond)

	// Cleanup contexts older than 10ms
	removed := manager.CleanupOldContexts(10 * time.Millisecond)
	assert.Equal(t, 2, removed)
	assert.Equal(t, 0, manager.ContextCount())
}

func TestContextManager_CleanupOldContexts_PreservesNew(t *testing.T) {
	manager := subagents.NewContextManager()

	// Create old context
	manager.CreateContext("parent", subagents.TypeExplore)

	time.Sleep(10 * time.Millisecond)

	// Create new context
	manager.CreateContext("parent", subagents.TypePlan)

	// Cleanup contexts older than 5ms
	removed := manager.CleanupOldContexts(5 * time.Millisecond)
	assert.Equal(t, 1, removed)
	assert.Equal(t, 1, manager.ContextCount())
}

// ========== Executor Tests ==========

func TestExecutor_NewExecutor(t *testing.T) {
	executor := subagents.NewExecutor()

	assert.NotNil(t, executor)
	assert.NotNil(t, executor.ContextManager)
	assert.Equal(t, 5*time.Minute, executor.DefaultTimeout)
	assert.Equal(t, 30*time.Minute, executor.MaxTimeout)
}

func TestExecutor_Execute_ValidatesRequest(t *testing.T) {
	executor := subagents.NewExecutor()

	// Invalid request (missing type)
	req := &subagents.ExecutionRequest{
		Prompt:      "Test",
		Description: "Test",
	}

	_, err := executor.Execute(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}

// mockHookEngine implements the HookEngine interface for testing
type mockHookEngine struct {
	called bool
}

func (m *mockHookEngine) FireSubagentStop(_, _ string, _, _ int, _ bool, _ float64) error {
	m.called = true
	return nil
}

func TestExecutor_ExecuteWithHooks(t *testing.T) {
	executor := subagents.NewExecutor()

	req := &subagents.ExecutionRequest{
		Type:        subagents.TypeExplore,
		Prompt:      "Test",
		Description: "Test",
	}

	// Execute with nil hook engine
	result, err := executor.ExecuteWithHooks(context.Background(), req, nil)
	// Will fail due to missing hex binary, but that's ok
	assert.NotNil(t, result)
	_ = err

	// Execute with mock hook engine
	mockEngine := &mockHookEngine{}
	result2, _ := executor.ExecuteWithHooks(context.Background(), req, mockEngine)
	assert.NotNil(t, result2)
	assert.Contains(t, result2.Metadata, "hook_fired")
	assert.True(t, result2.Metadata["hook_fired"].(bool))
	assert.True(t, mockEngine.called, "Hook should have been fired")
}

// ========== Dispatcher Tests ==========

func TestDispatcher_NewDispatcher(t *testing.T) {
	executor := subagents.NewExecutor()
	dispatcher := subagents.NewDispatcher(executor)

	assert.NotNil(t, dispatcher)
	assert.Equal(t, executor, dispatcher.Executor)
	assert.Equal(t, 10, dispatcher.MaxConcurrent)
}

func TestDispatcher_DispatchParallel_EmptyRequests(t *testing.T) {
	executor := subagents.NewExecutor()
	dispatcher := subagents.NewDispatcher(executor)

	results := dispatcher.DispatchParallel(context.Background(), []*subagents.DispatchRequest{})
	assert.Empty(t, results)
}

func TestDispatcher_DispatchSequential_EmptyRequests(t *testing.T) {
	executor := subagents.NewExecutor()
	dispatcher := subagents.NewDispatcher(executor)

	results := dispatcher.DispatchSequential(context.Background(), []*subagents.DispatchRequest{})
	assert.Empty(t, results)
}

func TestDispatcher_WaitForAny_NoRequests(t *testing.T) {
	executor := subagents.NewExecutor()
	dispatcher := subagents.NewDispatcher(executor)

	_, err := dispatcher.WaitForAny(context.Background(), []*subagents.DispatchRequest{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no requests")
}

func TestCalculateStatistics_EmptyResults(t *testing.T) {
	stats := subagents.CalculateStatistics([]*subagents.DispatchResult{})

	assert.Equal(t, 0, stats.Total)
	assert.Equal(t, 0, stats.Successful)
	assert.Equal(t, 0, stats.Failed)
	assert.Empty(t, stats.Errors)
}

func TestCalculateStatistics_MixedResults(t *testing.T) {
	results := []*subagents.DispatchResult{
		{
			ID: "1",
			Result: &subagents.Result{
				Success: true,
			},
		},
		{
			ID: "2",
			Result: &subagents.Result{
				Success: false,
				Error:   "test error",
			},
		},
		{
			ID:    "3",
			Error: assert.AnError,
		},
	}

	stats := subagents.CalculateStatistics(results)

	assert.Equal(t, 3, stats.Total)
	assert.Equal(t, 1, stats.Successful)
	assert.Equal(t, 2, stats.Failed)
	assert.Len(t, stats.Errors, 2)
}

func TestCalculateStatistics_AllSuccessful(t *testing.T) {
	results := []*subagents.DispatchResult{
		{
			ID: "1",
			Result: &subagents.Result{
				Success: true,
			},
		},
		{
			ID: "2",
			Result: &subagents.Result{
				Success: true,
			},
		},
	}

	stats := subagents.CalculateStatistics(results)

	assert.Equal(t, 2, stats.Total)
	assert.Equal(t, 2, stats.Successful)
	assert.Equal(t, 0, stats.Failed)
	assert.Empty(t, stats.Errors)
}

func TestCalculateStatistics_AllFailed(t *testing.T) {
	results := []*subagents.DispatchResult{
		{
			ID:    "1",
			Error: assert.AnError,
		},
		{
			ID: "2",
			Result: &subagents.Result{
				Success: false,
				Error:   "failed",
			},
		},
	}

	stats := subagents.CalculateStatistics(results)

	assert.Equal(t, 2, stats.Total)
	assert.Equal(t, 0, stats.Successful)
	assert.Equal(t, 2, stats.Failed)
	assert.Len(t, stats.Errors, 2)
}

// ========== Validation Error Tests ==========

func TestValidationError_Error(t *testing.T) {
	err := &subagents.ValidationError{
		Field:   "Type",
		Message: "invalid type",
	}

	assert.Contains(t, err.Error(), "Type")
	assert.Contains(t, err.Error(), "invalid type")
}

func TestValidationError_ErrorWithDetails(t *testing.T) {
	err := &subagents.ValidationError{
		Field:   "Type",
		Message: "invalid type",
		Details: map[string]interface{}{"foo": "bar"},
	}

	// Should still format correctly with details
	assert.Contains(t, err.Error(), "Type")
	assert.Contains(t, err.Error(), "invalid type")
}

func TestConfig_CustomConfiguration(t *testing.T) {
	config := &subagents.Config{
		Type:         subagents.TypeExplore,
		Model:        "claude-opus-4",
		Timeout:      10 * time.Minute,
		MaxTokens:    16384,
		Temperature:  0.5,
		AllowedTools: []string{"Read", "Grep"},
		SystemPrompt: "Custom prompt",
	}

	assert.Equal(t, subagents.TypeExplore, config.Type)
	assert.Equal(t, "claude-opus-4", config.Model)
	assert.Equal(t, 10*time.Minute, config.Timeout)
	assert.Equal(t, 16384, config.MaxTokens)
	assert.Equal(t, 0.5, config.Temperature)
	assert.Len(t, config.AllowedTools, 2)
	assert.Equal(t, "Custom prompt", config.SystemPrompt)
}

func TestExecutionRequest_WithCustomContext(t *testing.T) {
	req := &subagents.ExecutionRequest{
		Type:        subagents.TypePlan,
		Prompt:      "Test",
		Description: "Test",
		Context: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
	}

	err := req.Validate()
	assert.NoError(t, err)
	assert.Len(t, req.Context, 2)
}

func TestNewExecutionContext(t *testing.T) {
	ctx := context.Background()
	isolated := subagents.NewIsolatedContext("parent", subagents.TypeExplore)

	execCtx := subagents.NewExecutionContext(ctx, isolated)

	assert.NotNil(t, execCtx)
	assert.Equal(t, isolated, execCtx.Isolated)
}

// ========== Integration Tests ==========

func TestExecutor_ContextIsolation(t *testing.T) {
	executor := subagents.NewExecutor()

	// Create context manager to verify isolation
	initialCount := executor.ContextManager.ContextCount()

	req := &subagents.ExecutionRequest{
		Type:        subagents.TypeExplore,
		Prompt:      "Test prompt",
		Description: "Test description",
	}

	// Execute (will fail since hex binary likely not available, but that's ok)
	_, _ = executor.Execute(context.Background(), req)

	// Context should be cleaned up after execution
	finalCount := executor.ContextManager.ContextCount()
	assert.Equal(t, initialCount, finalCount, "Context should be cleaned up after execution")
}

func TestDispatcher_DispatchBatch(t *testing.T) {
	executor := subagents.NewExecutor()
	dispatcher := subagents.NewDispatcher(executor)

	// Create 5 invalid requests (will fail validation but that's ok for this test)
	requests := make([]*subagents.DispatchRequest, 5)
	for i := 0; i < 5; i++ {
		requests[i] = &subagents.DispatchRequest{
			ID: string(rune(i + '0')),
			Request: &subagents.ExecutionRequest{
				Type:        subagents.TypeExplore,
				Prompt:      "Test",
				Description: "Test",
			},
		}
	}

	// Execute in batches of 2
	results := dispatcher.DispatchBatch(context.Background(), requests, 2)

	// Should get all results back
	require.Len(t, results, 5)
}

func TestDispatcher_ContextCancellation(t *testing.T) {
	executor := subagents.NewExecutor()
	dispatcher := subagents.NewDispatcher(executor)

	// Create context that we'll cancel
	ctx, cancel := context.WithCancel(context.Background())

	// Create requests
	requests := []*subagents.DispatchRequest{
		{
			ID: "1",
			Request: &subagents.ExecutionRequest{
				Type:        subagents.TypeExplore,
				Prompt:      "Test 1",
				Description: "Test 1",
			},
		},
	}

	// Cancel immediately
	cancel()

	// Execute should respect cancellation
	results := dispatcher.DispatchSequential(ctx, requests)

	// Should still return results (even if incomplete)
	assert.NotNil(t, results)
}
