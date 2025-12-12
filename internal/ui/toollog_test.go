package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestToolLogInitialization(t *testing.T) {
	model := NewModel("test-conv", "test-model")

	// Tool log should start empty
	assert.Empty(t, model.toolLogLines)
	// Tool log overlay should not be active
	assert.NotEqual(t, model.toolLogOverlay, model.overlayManager.GetActive())
}

func TestToolLogAppendLines(t *testing.T) {
	model := NewModel("test-conv", "test-model")

	// Append some lines
	model.appendToolLogLine("Building...")
	model.appendToolLogLine("Compiling main.go")
	model.appendToolLogLine("Done!")

	assert.Len(t, model.toolLogLines, 3)
	assert.Equal(t, "Building...", model.toolLogLines[0])
	assert.Equal(t, "Compiling main.go", model.toolLogLines[1])
	assert.Equal(t, "Done!", model.toolLogLines[2])
}

func TestToolLogGetLastNLines(t *testing.T) {
	model := NewModel("test-conv", "test-model")

	// Add 5 lines
	model.appendToolLogLine("Line 1")
	model.appendToolLogLine("Line 2")
	model.appendToolLogLine("Line 3")
	model.appendToolLogLine("Line 4")
	model.appendToolLogLine("Line 5")

	// Get last 3
	last3 := model.getToolLogLastN(3)
	assert.Len(t, last3, 3)
	assert.Equal(t, "Line 3", last3[0])
	assert.Equal(t, "Line 4", last3[1])
	assert.Equal(t, "Line 5", last3[2])

	// Get last 10 when only 5 exist
	last10 := model.getToolLogLastN(10)
	assert.Len(t, last10, 5)
}

func TestToolLogClearChunk(t *testing.T) {
	model := NewModel("test-conv", "test-model")

	model.appendToolLogLine("Line 1")
	model.appendToolLogLine("Line 2")
	assert.Len(t, model.toolLogLines, 2)

	model.clearToolLogChunk()
	assert.Empty(t, model.toolLogLines)
}

func TestToolLogOverlayToggle(t *testing.T) {
	model := NewModel("test-conv", "test-model")
	model.Ready = true
	model.Width = 80
	model.Height = 24

	// Tool log overlay should not be active initially
	assert.NotEqual(t, model.toolLogOverlay, model.overlayManager.GetActive())

	// Ctrl+O should toggle overlay on
	ctrlO := tea.KeyMsg{Type: tea.KeyCtrlO}
	newModel, _ := model.Update(ctrlO)
	m := newModel.(*Model)
	assert.Equal(t, m.toolLogOverlay, m.overlayManager.GetActive())

	// Ctrl+O again should toggle off
	newModel, _ = m.Update(ctrlO)
	m = newModel.(*Model)
	assert.NotEqual(t, m.toolLogOverlay, m.overlayManager.GetActive())
}

func TestToolLogOverlayEscapeCloses(t *testing.T) {
	model := NewModel("test-conv", "test-model")
	model.Ready = true
	model.Width = 80
	model.Height = 24

	// Open the tool log overlay
	model.overlayManager.Push(model.toolLogOverlay, model.Width, model.Height)
	assert.Equal(t, model.toolLogOverlay, model.overlayManager.GetActive())

	// Escape should close overlay
	esc := tea.KeyMsg{Type: tea.KeyEsc}
	newModel, _ := model.Update(esc)
	m := newModel.(*Model)
	assert.NotEqual(t, m.toolLogOverlay, m.overlayManager.GetActive())
}

func TestRenderCollapsedToolLog(t *testing.T) {
	model := NewModel("test-conv", "test-model")
	model.Ready = true

	// Add some lines
	model.appendToolLogLine("go build -o bin/hex")
	model.appendToolLogLine("✅ Built bin/hexviz")
	model.appendToolLogLine("✅ Built bin/hex")

	rendered, hiddenLines := model.renderCollapsedToolLog()

	// Should have prefix on each line
	assert.Contains(t, rendered, "│")
	// Should contain the output
	assert.Contains(t, rendered, "go build")
	assert.Contains(t, rendered, "Built bin/hex")
	// Only 3 lines, no hidden
	assert.Equal(t, 0, hiddenLines)
}

func TestRenderCollapsedToolLogEmpty(t *testing.T) {
	model := NewModel("test-conv", "test-model")
	model.Ready = true

	// No tool output = empty render
	rendered, hiddenLines := model.renderCollapsedToolLog()
	assert.Empty(t, rendered)
	assert.Equal(t, 0, hiddenLines)
}

func TestRenderToolLogOverlay(t *testing.T) {
	model := NewModel("test-conv", "test-model")
	model.Ready = true
	model.Width = 80
	model.Height = 24

	model.appendToolLogLine("Line 1")
	model.appendToolLogLine("Line 2")
	model.appendToolLogLine("Line 3")

	// Initialize the overlay
	model.toolLogOverlay.OnPush(model.Width, model.Height)
	rendered := model.toolLogOverlay.Render(model.Width, model.Height)

	// Should have header
	assert.Contains(t, rendered, "Tool Output Log")
	// Should have close hint
	assert.Contains(t, rendered, "Ctrl+O")
	// Should contain all lines
	assert.Contains(t, rendered, "Line 1")
	assert.Contains(t, rendered, "Line 2")
	assert.Contains(t, rendered, "Line 3")
}

func TestRenderToolLogOverlayEmpty(t *testing.T) {
	model := NewModel("test-conv", "test-model")
	model.Ready = true
	model.Width = 80
	model.Height = 24

	// Initialize the overlay
	model.toolLogOverlay.OnPush(model.Width, model.Height)
	rendered := model.toolLogOverlay.Render(model.Width, model.Height)

	// Should show empty message
	assert.Contains(t, rendered, "No tool output")
}

func TestToolLogMultilineOutput(t *testing.T) {
	model := NewModel("test-conv", "test-model")

	// Simulate tool output with multiple lines
	output := "Line 1\nLine 2\nLine 3"
	model.appendToolLogOutput(output)

	assert.Len(t, model.toolLogLines, 3)
	assert.Equal(t, "Line 1", model.toolLogLines[0])
	assert.Equal(t, "Line 2", model.toolLogLines[1])
	assert.Equal(t, "Line 3", model.toolLogLines[2])
}

func TestToolLogTracksCurrentTool(t *testing.T) {
	model := NewModel("test-conv", "test-model")

	// Start a tool
	model.startToolLogEntry("bash", `"make build"`)
	assert.Equal(t, "bash", model.currentToolLogName)
	assert.Equal(t, `"make build"`, model.currentToolLogParam)

	// Should have header in lines
	assert.True(t, strings.Contains(model.toolLogLines[0], "bash"))
}
