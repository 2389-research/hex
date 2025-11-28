// ABOUTME: Integration tests for quick actions with Model
// ABOUTME: Validates quick actions work correctly with the Bubbletea model
package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModelQuickActionsMode(t *testing.T) {
	model := NewModel("test-conv", "test-model")

	// Initially not in quick actions mode
	assert.False(t, model.quickActionsMode)

	// Enter quick actions mode
	model.EnterQuickActionsMode()
	assert.True(t, model.quickActionsMode)
	assert.Equal(t, "", model.quickActionsInput)
	assert.Greater(t, len(model.quickActionsFiltered), 0, "Should have default actions")

	// Exit quick actions mode
	model.ExitQuickActionsMode()
	assert.False(t, model.quickActionsMode)
	assert.Equal(t, "", model.quickActionsInput)
	assert.Len(t, model.quickActionsFiltered, 0)
}

func TestModelQuickActionsSearch(t *testing.T) {
	model := NewModel("test-conv", "test-model")
	model.EnterQuickActionsMode()

	// Update input to search for "read"
	model.UpdateQuickActionsInput("read")
	assert.Equal(t, "read", model.quickActionsInput)

	// Should filter to read action
	require.Greater(t, len(model.quickActionsFiltered), 0)
	assert.Equal(t, "read", model.quickActionsFiltered[0].Name)
}

func TestModelQuickActionsKeyHandler(t *testing.T) {
	model := NewModel("test-conv", "test-model")
	model.EnterQuickActionsMode()

	// Test typing a character
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}
	updatedModel, _ := model.handleQuickActionsKey(msg)
	m := updatedModel.(*Model)
	assert.Equal(t, "r", m.quickActionsInput)

	// Test backspace
	msg = tea.KeyMsg{Type: tea.KeyBackspace}
	updatedModel, _ = m.handleQuickActionsKey(msg)
	m = updatedModel.(*Model)
	assert.Equal(t, "", m.quickActionsInput)

	// Test Esc to exit
	msg = tea.KeyMsg{Type: tea.KeyEsc}
	updatedModel, _ = m.handleQuickActionsKey(msg)
	m = updatedModel.(*Model)
	assert.False(t, m.quickActionsMode)
}

func TestModelQuickActionsColonKey(t *testing.T) {
	model := NewModel("test-conv", "test-model")

	// Simulate pressing ':' when textarea is not focused
	model.Input.Blur()

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(*Model)

	// Should enter quick actions mode
	assert.True(t, m.quickActionsMode)
}

func TestModelRenderQuickActionsModal(t *testing.T) {
	model := NewModel("test-conv", "test-model")
	model.EnterQuickActionsMode()

	// Render the modal
	rendered := model.renderQuickActionsModal()

	// Should contain key elements
	assert.Contains(t, rendered, "Quick Actions")
	assert.Contains(t, rendered, ":")
	assert.Contains(t, rendered, "Enter: execute")
	assert.Contains(t, rendered, "Esc: cancel")
}

func TestModelQuickActionsExecute(t *testing.T) {
	model := NewModel("test-conv", "test-model")
	model.EnterQuickActionsMode()

	// Type "save"
	model.UpdateQuickActionsInput("save")

	// Execute should work (even though handler isn't connected yet)
	err := model.ExecuteQuickAction()
	// Expected to fail because handlers aren't connected
	assert.Error(t, err)

	// But mode should be exited
	assert.False(t, model.quickActionsMode)
}

func TestModelQuickActionsWithArguments(t *testing.T) {
	model := NewModel("test-conv", "test-model")
	model.EnterQuickActionsMode()

	// Type "read /path/to/file"
	model.UpdateQuickActionsInput("read /path/to/file")

	// The fuzzy search searches for the full string, so it won't match
	// This is expected behavior - the command parser will extract "read" later
	// But the filtered list will be empty since no action name contains the full string
	// This is OK - ExecuteQuickAction will parse the command correctly

	// Parse command to verify it works
	command, args := ParseActionCommand("read /path/to/file")
	assert.Equal(t, "read", command)
	assert.Equal(t, "/path/to/file", args)
}
