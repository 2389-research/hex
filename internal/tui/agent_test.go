// ABOUTME: Tests for HexAgent
// ABOUTME: Verifies tux.Agent interface implementation

package tui

import (
	"testing"
	"time"

	"github.com/2389-research/hex/internal/core"
	"github.com/2389-research/tux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestAgent creates a HexAgent for testing without requiring a real client.
// Only use this for tests that don't call Run().
func newTestAgent(model, systemPrompt string) *HexAgent {
	return &HexAgent{
		client:       nil,
		model:        model,
		systemPrompt: systemPrompt,
		messages:     make([]core.Message, 0),
		executor:     nil, // Tests that don't execute tools
		storage:      nil, // No session persistence in tests
	}
}

func TestHexAgent_ImplementsInterface(t *testing.T) {
	// Compile-time check that HexAgent implements tux.Agent
	var _ tux.Agent = (*HexAgent)(nil)
}

func TestHexAgent_Subscribe(t *testing.T) {
	agent := newTestAgent("test-model", "test system")

	ch := agent.Subscribe()
	require.NotNil(t, ch)
}

func TestHexAgent_Cancel(t *testing.T) {
	agent := newTestAgent("test-model", "test system")

	// Should not panic when no run is active
	agent.Cancel()
}

func TestHexAgent_ClearHistory(t *testing.T) {
	agent := newTestAgent("test-model", "test system")

	// Add some messages manually
	agent.mu.Lock()
	agent.messages = append(agent.messages, core.Message{
		Role:    "user",
		Content: "test",
	})
	agent.mu.Unlock()

	agent.ClearHistory()

	agent.mu.Lock()
	assert.Empty(t, agent.messages)
	agent.mu.Unlock()
}

func TestHexAgent_AddSystemContext(t *testing.T) {
	agent := newTestAgent("test-model", "initial system")

	agent.AddSystemContext("additional context")

	agent.mu.Lock()
	assert.Contains(t, agent.systemPrompt, "initial system")
	assert.Contains(t, agent.systemPrompt, "additional context")
	agent.mu.Unlock()
}

func TestHexAgent_EmitToSubscribers(t *testing.T) {
	agent := newTestAgent("test-model", "test system")

	ch := agent.Subscribe()

	// Emit an event
	agent.emit(tux.Event{Type: tux.EventText, Text: "hello"})

	// Should receive it
	select {
	case event := <-ch:
		assert.Equal(t, tux.EventText, event.Type)
		assert.Equal(t, "hello", event.Text)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for event")
	}
}

func TestHexAgent_MultipleSubscribes(t *testing.T) {
	agent := newTestAgent("test-model", "test system")

	// First subscription
	ch1 := agent.Subscribe()
	require.NotNil(t, ch1)

	// Second subscription should replace the first
	ch2 := agent.Subscribe()
	require.NotNil(t, ch2)

	// Emit to current subscriber
	agent.emit(tux.Event{Type: tux.EventText, Text: "test"})

	// Should receive on ch2
	select {
	case event := <-ch2:
		assert.Equal(t, tux.EventText, event.Type)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for event on ch2")
	}
}

func TestHexAgent_EmitWithNoSubscribers(t *testing.T) {
	agent := newTestAgent("test-model", "test system")

	// Should not panic when emitting with no subscribers
	agent.emit(tux.Event{Type: tux.EventText, Text: "hello"})
}

func TestHexAgent_CancelMultipleTimes(t *testing.T) {
	agent := newTestAgent("test-model", "test system")

	// Should not panic when called multiple times
	agent.Cancel()
	agent.Cancel()
	agent.Cancel()
}

func TestHexAgent_ResetToolState(t *testing.T) {
	agent := newTestAgent("test-model", "test system")

	// Set some state
	agent.assemblingTool = &core.ToolUse{ID: "test"}
	agent.toolInputJSONBuf.WriteString("test json")
	agent.pendingTools = []*core.ToolUse{{ID: "pending"}}

	// Reset
	agent.resetToolState()

	assert.Nil(t, agent.assemblingTool)
	assert.Equal(t, "", agent.toolInputJSONBuf.String())
	assert.Nil(t, agent.pendingTools)
}

func TestHexAgent_HandleContentBlockStart_ToolUse(t *testing.T) {
	agent := newTestAgent("test-model", "test system")
	ch := agent.Subscribe()

	chunk := &core.StreamChunk{
		Type: "content_block_start",
		ContentBlock: &core.Content{
			Type: "tool_use",
			ID:   "tool_123",
			Name: "read_file",
		},
	}

	agent.handleContentBlockStart(chunk)

	// Verify assembling tool is set
	require.NotNil(t, agent.assemblingTool)
	assert.Equal(t, "tool_123", agent.assemblingTool.ID)
	assert.Equal(t, "read_file", agent.assemblingTool.Name)

	// Verify event was emitted
	select {
	case event := <-ch:
		assert.Equal(t, tux.EventToolCall, event.Type)
		assert.Equal(t, "tool_123", event.ToolID)
		assert.Equal(t, "read_file", event.ToolName)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for event")
	}
}

func TestHexAgent_HandleContentBlockStop(t *testing.T) {
	agent := newTestAgent("test-model", "test system")

	// Set up assembling tool with JSON
	agent.assemblingTool = &core.ToolUse{
		ID:    "tool_123",
		Name:  "read_file",
		Input: make(map[string]interface{}),
	}
	agent.toolInputJSONBuf.WriteString(`{"path":"/test.txt"}`)

	agent.handleContentBlockStop()

	// Verify tool was added to pending
	require.Len(t, agent.pendingTools, 1)
	assert.Equal(t, "tool_123", agent.pendingTools[0].ID)
	assert.Equal(t, "/test.txt", agent.pendingTools[0].Input["path"])

	// Verify assembling state was cleared
	assert.Nil(t, agent.assemblingTool)
	assert.Equal(t, "", agent.toolInputJSONBuf.String())
}
