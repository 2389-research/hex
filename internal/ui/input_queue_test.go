// ABOUTME: Tests for input queuing during waitingForResponse state
// ABOUTME: Tests single queued message, UP arrow editing, input blocking

package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestWaitingForResponse_BlocksInput(t *testing.T) {
	model := NewModel("conv-123", "claude-sonnet-4-5-20250929")
	model.waitingForResponse = true

	// Type a message and press enter
	model.Input.SetValue("queued message")
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := model.Update(enterMsg)
	m := newModel.(*Model)

	// Message should be queued, not added to Messages
	assert.Equal(t, "queued message", m.queuedMessage)
	assert.Empty(t, m.Messages, "Message should not be added to Messages while waiting")
	assert.Empty(t, m.Input.Value(), "Input should be cleared after queuing")
}

func TestWaitingForResponse_OnlyOneQueuedMessage(t *testing.T) {
	model := NewModel("conv-123", "claude-sonnet-4-5-20250929")
	model.waitingForResponse = true
	model.queuedMessage = "first message"

	// Try to queue another message
	model.Input.SetValue("second message")
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := model.Update(enterMsg)
	m := newModel.(*Model)

	// Should still have first message, second should be ignored
	assert.Equal(t, "first message", m.queuedMessage)
}

func TestUpArrow_EditsQueuedMessage(t *testing.T) {
	model := NewModel("conv-123", "claude-sonnet-4-5-20250929")
	model.waitingForResponse = true
	model.queuedMessage = "my queued message"
	model.Input.Focus()

	// Press UP with empty input
	upMsg := tea.KeyMsg{Type: tea.KeyUp}
	newModel, _ := model.Update(upMsg)
	m := newModel.(*Model)

	// Queued message should move to input
	assert.Empty(t, m.queuedMessage, "Queued message should be cleared")
	assert.Equal(t, "my queued message", m.Input.Value(), "Input should have queued message")
}

func TestUpArrow_DoesNotEditIfInputNotEmpty(t *testing.T) {
	model := NewModel("conv-123", "claude-sonnet-4-5-20250929")
	model.waitingForResponse = true
	model.queuedMessage = "queued"
	model.Input.Focus()
	model.Input.SetValue("typing something")

	// Press UP with non-empty input
	upMsg := tea.KeyMsg{Type: tea.KeyUp}
	newModel, _ := model.Update(upMsg)
	m := newModel.(*Model)

	// Should NOT pull queued message (input is not empty)
	assert.Equal(t, "queued", m.queuedMessage, "Queued message should remain")
}

func TestInputBlocked_WhenQueuedMessageExists(t *testing.T) {
	model := NewModel("conv-123", "claude-sonnet-4-5-20250929")
	model.waitingForResponse = true
	model.queuedMessage = "existing queued message"
	model.Input.Focus()

	// Try to type - input should be blocked
	oldValue := model.Input.Value()
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	newModel, _ := model.Update(keyMsg)
	m := newModel.(*Model)

	// Input should not have changed (blocked)
	assert.Equal(t, oldValue, m.Input.Value(), "Input should be blocked when queued message exists")
}

func TestNotWaiting_ProcessesImmediately(t *testing.T) {
	model := NewModel("conv-123", "claude-sonnet-4-5-20250929")
	model.waitingForResponse = false

	// Type and send a message
	model.Input.SetValue("immediate message")
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := model.Update(enterMsg)
	m := newModel.(*Model)

	// Message should be added to Messages immediately
	assert.Empty(t, m.queuedMessage, "Should not be queued")
	assert.Len(t, m.Messages, 1, "Message should be added to Messages")
	assert.Equal(t, "immediate message", m.Messages[0].Content)
	assert.True(t, m.waitingForResponse, "Should now be waiting for response")
}

func TestProcessQueuedMessage_AddsToMessages(t *testing.T) {
	model := NewModel("conv-123", "claude-sonnet-4-5-20250929")
	model.queuedMessage = "queued message"
	model.waitingForResponse = false // Simulate response complete

	// Process queued message
	cmd := model.processQueuedMessage()

	// Should be added to Messages now
	assert.Empty(t, model.queuedMessage, "Queued message should be cleared")
	assert.Len(t, model.Messages, 1, "Message should be added")
	assert.Equal(t, "queued message", model.Messages[0].Content)
	assert.True(t, model.waitingForResponse, "Should be waiting again")
	// cmd would trigger streaming if apiClient was set
	_ = cmd
}

func TestProcessQueuedMessage_EmptyQueue(t *testing.T) {
	model := NewModel("conv-123", "claude-sonnet-4-5-20250929")
	model.queuedMessage = ""

	cmd := model.processQueuedMessage()

	// Should be nil command, no changes
	assert.Nil(t, cmd)
	assert.Empty(t, model.Messages)
}

func TestQueuedMessageDisplay(t *testing.T) {
	model := NewModel("conv-123", "claude-sonnet-4-5-20250929")
	model.queuedMessage = "test message"
	model.CurrentView = ViewModeChat

	// Initialize model properly for rendering
	model.Width = 80
	model.Height = 24
	model.Ready = true
	model.ShowIntro = false

	view := model.View()

	// Should contain the queued message indicator
	assert.Contains(t, view, "test message")
	assert.Contains(t, view, "queued")
	assert.Contains(t, view, "↑ to edit")
}
