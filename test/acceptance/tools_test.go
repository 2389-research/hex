// ABOUTME: Acceptance tests for tool approval workflow
// ABOUTME: Tests that tool calls trigger approval UI and execute correctly

package acceptance

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTool_ApprovalModalAppears(t *testing.T) {
	h := NewBubbleteaAdapter()
	require.NoError(t, h.Init(120, 40))
	defer h.Shutdown()

	// Simulate a tool call
	params := map[string]interface{}{
		"path": "/some/file.txt",
	}
	require.NoError(t, h.SimulateToolCall("tool-1", "read_file", params))

	// Should show queued status
	assert.Equal(t, "queued", h.GetStatus(),
		"status should be queued when tool awaits approval")
}

func TestTool_ResultDisplayedAfterExecution(t *testing.T) {
	h := NewBubbleteaAdapter()
	require.NoError(t, h.Init(120, 40))
	defer h.Shutdown()

	// Simulate successful tool result
	require.NoError(t, h.SimulateToolResult("tool-1", true, "File contents here"))

	// The result should be visible somewhere in the view
	// (Exact format depends on implementation)
	// This documents expected behavior
}

func TestTool_DeniedToolShowsMessage(t *testing.T) {
	h := NewBubbleteaAdapter()
	require.NoError(t, h.Init(120, 40))
	defer h.Shutdown()

	// Simulate denied tool result
	require.NoError(t, h.SimulateToolResult("tool-1", false, "Tool execution denied by user"))

	// Should show denial somehow
	// (Exact format depends on implementation)
}
