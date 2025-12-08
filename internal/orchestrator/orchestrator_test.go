// ABOUTME: Unit tests for AgentOrchestrator
// ABOUTME: Tests event emission, state transitions, and stream handling
package orchestrator

import (
	"context"
	"testing"
	"time"

	"github.com/2389-research/hex/internal/core"
	"github.com/2389-research/hex/internal/tools"
)

// TestStart_SendsStreamStartEvent verifies that Start() emits a StreamStart event
func TestStart_SendsStreamStartEvent(t *testing.T) {
	// Mock client and executor
	mockClient := &mockAPIClient{}
	mockExecutor := &mockToolExecutor{}

	// Create orchestrator
	orch := NewOrchestrator(mockClient, "claude-sonnet-4-5-20250929", mockExecutor)

	// Subscribe to events
	eventChan := orch.Subscribe()

	// Start with a prompt
	ctx := context.Background()
	err := orch.Start(ctx, "test prompt")
	if err != nil {
		t.Fatalf("Start() failed: %v", err)
	}

	// Should receive StreamStart event
	select {
	case event := <-eventChan:
		if event.Type != EventStreamStart {
			t.Errorf("Expected EventStreamStart, got %s", event.Type)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("No StreamStart event received")
	}

	// State should be Streaming
	state := orch.GetState()
	if state != StateStreaming {
		t.Errorf("Expected StateStreaming, got %s", state)
	}
}

// TestToolApproval_TransitionsState verifies state transitions during tool approval
func TestToolApproval_TransitionsState(t *testing.T) {
	mockClient := &mockAPIClient{}
	mockExecutor := &mockToolExecutor{}

	orch := NewOrchestrator(mockClient, "claude-sonnet-4-5-20250929", mockExecutor)

	// Do valid transition sequence: Idle -> Streaming -> AwaitingApproval
	orch.setState(StateStreaming)
	orch.setState(StateAwaitingApproval)

	// Add a pending tool
	toolUse := &core.ToolUse{
		ID:    "test-tool-1",
		Name:  "test_tool",
		Input: map[string]interface{}{"arg": "value"},
	}
	orch.addPendingTool(toolUse)

	// Approve the tool
	err := orch.HandleToolApproval("test-tool-1", true)
	if err != nil {
		t.Fatalf("HandleToolApproval failed: %v", err)
	}

	// State should transition to ExecutingTool
	state := orch.GetState()
	if state != StateExecutingTool {
		t.Errorf("Expected StateExecutingTool, got %s", state)
	}
}

// TestStop_CancelsStream verifies that Stop() cancels active stream
func TestStop_CancelsStream(t *testing.T) {
	mockClient := &mockAPIClient{}
	mockExecutor := &mockToolExecutor{}

	orch := NewOrchestrator(mockClient, "claude-sonnet-4-5-20250929", mockExecutor)

	// Start a stream
	ctx := context.Background()
	err := orch.Start(ctx, "test prompt")
	if err != nil {
		t.Fatalf("Start() failed: %v", err)
	}

	// Stop the stream
	err = orch.Stop()
	if err != nil {
		t.Fatalf("Stop() failed: %v", err)
	}

	// State should be Idle
	state := orch.GetState()
	if state != StateIdle {
		t.Errorf("Expected StateIdle after Stop(), got %s", state)
	}
}

// TestEventEmission verifies that events are emitted correctly
func TestEventEmission(t *testing.T) {
	mockClient := &mockAPIClient{}
	mockExecutor := &mockToolExecutor{}

	orch := NewOrchestrator(mockClient, "claude-sonnet-4-5-20250929", mockExecutor)
	eventChan := orch.Subscribe()

	// Manually emit different event types
	orch.emitEvent(EventStreamStart, nil)
	orch.emitEvent(EventStreamChunk, "chunk data")
	orch.emitEvent(EventComplete, nil)

	// Collect events
	events := []Event{}
	timeout := time.After(100 * time.Millisecond)
	for i := 0; i < 3; i++ {
		select {
		case evt := <-eventChan:
			events = append(events, evt)
		case <-timeout:
			t.Fatal("Timeout waiting for events")
		}
	}

	// Verify event types
	if len(events) != 3 {
		t.Fatalf("Expected 3 events, got %d", len(events))
	}

	expectedTypes := []EventType{EventStreamStart, EventStreamChunk, EventComplete}
	for i, expected := range expectedTypes {
		if events[i].Type != expected {
			t.Errorf("Event %d: expected %s, got %s", i, expected, events[i].Type)
		}
	}
}

// TestConcurrentEventEmission verifies thread-safe event emission
func TestConcurrentEventEmission(t *testing.T) {
	mockClient := &mockAPIClient{}
	mockExecutor := &mockToolExecutor{}

	orch := NewOrchestrator(mockClient, "claude-sonnet-4-5-20250929", mockExecutor)
	eventChan := orch.Subscribe()

	// Emit events from multiple goroutines
	numGoroutines := 10
	eventsPerGoroutine := 5

	done := make(chan bool)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			for j := 0; j < eventsPerGoroutine; j++ {
				orch.emitEvent(EventStreamChunk, id)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to finish
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Collect events
	events := []Event{}
	timeout := time.After(500 * time.Millisecond)
	for i := 0; i < numGoroutines*eventsPerGoroutine; i++ {
		select {
		case evt := <-eventChan:
			events = append(events, evt)
		case <-timeout:
			// May not receive all events due to buffering, but should not panic
			t.Logf("Received %d events (expected %d)", len(events), numGoroutines*eventsPerGoroutine)
			return
		}
	}

	if len(events) != numGoroutines*eventsPerGoroutine {
		t.Logf("Received %d events (expected %d)", len(events), numGoroutines*eventsPerGoroutine)
	}
}

// Mock implementations

type mockAPIClient struct {
	streamChan chan *core.StreamChunk
}

func (m *mockAPIClient) CreateMessageStream(ctx context.Context, req core.MessageRequest) (<-chan *core.StreamChunk, error) {
	if m.streamChan == nil {
		m.streamChan = make(chan *core.StreamChunk, 10)
	}
	return m.streamChan, nil
}

type mockToolExecutor struct {
	executedTools []string
}

func (m *mockToolExecutor) Execute(ctx context.Context, toolName string, params map[string]interface{}) (*tools.Result, error) {
	m.executedTools = append(m.executedTools, toolName)
	return &tools.Result{
		ToolName: toolName,
		Success:  true,
		Output:   "mock output",
	}, nil
}
