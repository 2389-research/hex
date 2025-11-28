// ABOUTME: Tests for tool integration in the UI
// ABOUTME: Validates tool approval flow, execution, and result handling

package ui_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/harper/clem/internal/core"
	"github.com/harper/clem/internal/tools"
	"github.com/harper/clem/internal/ui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetToolSystem(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929")

	registry := tools.NewRegistry()
	approvalFunc := func(toolName string, params map[string]interface{}) bool {
		return true
	}
	executor := tools.NewExecutor(registry, approvalFunc)

	model.SetToolSystem(registry, executor)

	// Model should now have tool system set (we can't directly check private fields,
	// but we can verify that HandleToolUse doesn't panic)
	toolUse := &core.ToolUse{
		Type:  "tool_use",
		ID:    "tool-123",
		Name:  "read",
		Input: map[string]interface{}{"path": "/test"},
	}

	cmd := model.HandleToolUse(toolUse)
	assert.Nil(t, cmd) // HandleToolUse should return nil (no command to execute yet)
}

func TestHandleToolUse(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929")

	// Initialize model
	model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	toolUse := &core.ToolUse{
		Type:  "tool_use",
		ID:    "tool-123",
		Name:  "read",
		Input: map[string]interface{}{"path": "/test"},
	}

	cmd := model.HandleToolUse(toolUse)
	assert.Nil(t, cmd)

	// Model should now be in tool approval mode (we verify through view rendering)
	view := model.View()
	assert.Contains(t, view, "Tool Approval Required")
	assert.Contains(t, view, "read")
}

func TestApproveToolUseWithoutToolSystem(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929")

	// Create a pending tool use
	toolUse := &core.ToolUse{
		Type:  "tool_use",
		ID:    "tool-123",
		Name:  "read",
		Input: map[string]interface{}{"path": "/test"},
	}
	model.HandleToolUse(toolUse)

	// Approve without setting up tool system should not panic
	cmd := model.ApproveToolUse()
	assert.Nil(t, cmd)
}

func TestApproveToolUseWithToolSystem(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929")

	// Setup tool system
	registry := tools.NewRegistry()
	require.NoError(t, registry.Register(tools.NewReadTool()))

	approvalFunc := func(toolName string, params map[string]interface{}) bool {
		return true
	}
	executor := tools.NewExecutor(registry, approvalFunc)
	model.SetToolSystem(registry, executor)

	// Create a pending tool use
	toolUse := &core.ToolUse{
		Type:  "tool_use",
		ID:    "tool-123",
		Name:  "read",
		Input: map[string]interface{}{"file_path": "/tmp/test.txt"},
	}
	model.HandleToolUse(toolUse)

	// Approve should return a command to execute the tool
	cmd := model.ApproveToolUse()
	assert.NotNil(t, cmd)
}

func TestDenyToolUse(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929")

	// Initialize model
	model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Create a pending tool use
	toolUse := &core.ToolUse{
		Type:  "tool_use",
		ID:    "tool-123",
		Name:  "read",
		Input: map[string]interface{}{"path": "/test"},
	}
	model.HandleToolUse(toolUse)

	// Deny will return nil if no API client is set (sendToolResults returns nil)
	// But it should still exit approval mode
	cmd := model.DenyToolUse()
	// cmd may be nil if no API client is set
	_ = cmd

	// View should no longer show approval prompt
	view := model.View()
	assert.NotContains(t, view, "Tool Approval Required")
}

func TestToolApprovalModeInView(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929")

	// Initialize model by sending a WindowSizeMsg to make it ready
	model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Initial view should not show tool approval
	view := model.View()
	assert.NotContains(t, view, "Tool Approval Required")

	// Create a pending tool use
	toolUse := &core.ToolUse{
		Type:  "tool_use",
		ID:    "tool-123",
		Name:  "read",
		Input: map[string]interface{}{"path": "/test.txt"},
	}
	model.HandleToolUse(toolUse)

	// View should now show tool approval prompt
	view = model.View()
	assert.Contains(t, view, "Tool Approval Required")
	assert.Contains(t, view, "read")
	assert.Contains(t, view, "path")
	assert.Contains(t, view, "/test.txt")
	assert.Contains(t, view, "Allow this tool to execute?")
}
