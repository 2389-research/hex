// ABOUTME: Tests for input queuing during waitingForResponse state
// ABOUTME: Tests message queue array, UP arrow editing, queue processing

package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestWaitingForResponse_QueuesMessage(t *testing.T) {
	model := NewModel("conv-123", "claude-sonnet-4-5-20250929")
	model.waitingForResponse = true

	// Type a message and press enter
	model.Input.SetValue("queued message")
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := model.Update(enterMsg)
	m := newModel.(*Model)

	// Message should be queued, not added to Messages
	assert.Equal(t, 1, m.QueueCount())
	assert.Equal(t, "queued message", m.PeekQueue())
	assert.Empty(t, m.Messages, "Message should not be added to Messages while waiting")
	assert.Empty(t, m.Input.Value(), "Input should be cleared after queuing")
}

func TestWaitingForResponse_QueuesMultipleMessages(t *testing.T) {
	model := NewModel("conv-123", "claude-sonnet-4-5-20250929")
	model.waitingForResponse = true
	model.QueueMessage("first message")

	// Queue another message
	model.Input.SetValue("second message")
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := model.Update(enterMsg)
	m := newModel.(*Model)

	// Should have both messages queued
	assert.Equal(t, 2, m.QueueCount())
	assert.Equal(t, "first message", m.PeekQueue())
}

func TestUpArrow_EditsFirstQueuedMessage(t *testing.T) {
	model := NewModel("conv-123", "claude-sonnet-4-5-20250929")
	model.waitingForResponse = true
	model.QueueMessage("first message")
	model.QueueMessage("second message")
	model.Input.Focus()

	// Press UP with empty input
	upMsg := tea.KeyMsg{Type: tea.KeyUp}
	newModel, _ := model.Update(upMsg)
	m := newModel.(*Model)

	// First queued message should move to input, second remains in queue
	assert.Equal(t, 1, m.QueueCount())
	assert.Equal(t, "second message", m.PeekQueue())
	assert.Equal(t, "first message", m.Input.Value(), "Input should have first queued message")
}

func TestUpArrow_DoesNotEditIfInputNotEmpty(t *testing.T) {
	model := NewModel("conv-123", "claude-sonnet-4-5-20250929")
	model.waitingForResponse = true
	model.QueueMessage("queued")
	model.Input.Focus()
	model.Input.SetValue("typing something")

	// Press UP with non-empty input
	upMsg := tea.KeyMsg{Type: tea.KeyUp}
	newModel, _ := model.Update(upMsg)
	m := newModel.(*Model)

	// Should NOT pull queued message (input is not empty)
	assert.Equal(t, 1, m.QueueCount(), "Queued message should remain")
}

func TestTypingAllowed_WhenQueuedMessagesExist(t *testing.T) {
	model := NewModel("conv-123", "claude-sonnet-4-5-20250929")
	model.waitingForResponse = true
	model.QueueMessage("existing queued message")
	model.Input.Focus()

	// Should be able to type even with queued messages
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	newModel, _ := model.Update(keyMsg)
	m := newModel.(*Model)

	// Input should have the character (typing allowed)
	assert.Contains(t, m.Input.Value(), "a", "Typing should be allowed when queued messages exist")
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
	assert.Equal(t, 0, m.QueueCount(), "Should not be queued")
	assert.Len(t, m.Messages, 1, "Message should be added to Messages")
	assert.Equal(t, "immediate message", m.Messages[0].Content)
	assert.True(t, m.waitingForResponse, "Should now be waiting for response")
}

func TestProcessQueuedMessage_AddsToMessages(t *testing.T) {
	model := NewModel("conv-123", "claude-sonnet-4-5-20250929")
	model.QueueMessage("queued message")
	model.waitingForResponse = false // Simulate response complete

	// Process queued message
	cmd := model.processQueuedMessage()

	// Should be added to Messages now
	assert.Equal(t, 0, model.QueueCount(), "Queue should be empty")
	assert.Len(t, model.Messages, 1, "Message should be added")
	assert.Equal(t, "queued message", model.Messages[0].Content)
	assert.True(t, model.waitingForResponse, "Should be waiting again")
	// cmd would trigger streaming if apiClient was set
	_ = cmd
}

func TestProcessQueuedMessage_ProcessesFirstOnly(t *testing.T) {
	model := NewModel("conv-123", "claude-sonnet-4-5-20250929")
	model.QueueMessage("first")
	model.QueueMessage("second")
	model.QueueMessage("third")
	model.waitingForResponse = false

	// Process queued message
	_ = model.processQueuedMessage()

	// Only first should be processed, others remain
	assert.Equal(t, 2, model.QueueCount())
	assert.Equal(t, "second", model.PeekQueue())
	assert.Len(t, model.Messages, 1)
	assert.Equal(t, "first", model.Messages[0].Content)
}

func TestProcessQueuedMessage_EmptyQueue(t *testing.T) {
	model := NewModel("conv-123", "claude-sonnet-4-5-20250929")

	cmd := model.processQueuedMessage()

	// Should be nil command, no changes
	assert.Nil(t, cmd)
	assert.Empty(t, model.Messages)
}

func TestQueuedMessageDisplay(t *testing.T) {
	model := NewModel("conv-123", "claude-sonnet-4-5-20250929")
	model.QueueMessage("test message")
	model.CurrentView = ViewModeChat

	// Initialize model properly for rendering
	model.Width = 80
	model.Height = 24
	model.Ready = true
	model.ShowIntro = false

	view := model.View()

	// Should contain the queued message indicator with count
	assert.Contains(t, view, "test message")
	assert.Contains(t, view, "[Q:1]")
	assert.Contains(t, view, "↑ to edit")
}

func TestQueueMethods(t *testing.T) {
	model := NewModel("conv-123", "claude-sonnet-4-5-20250929")

	// Test empty queue
	assert.Equal(t, 0, model.QueueCount())
	assert.Equal(t, "", model.PeekQueue())
	assert.Equal(t, "", model.PopQueue())
	assert.Nil(t, model.DrainQueue())

	// Test adding messages
	model.QueueMessage("one")
	model.QueueMessage("two")
	assert.Equal(t, 2, model.QueueCount())
	assert.Equal(t, "one", model.PeekQueue())

	// Test PopQueue
	assert.Equal(t, "one", model.PopQueue())
	assert.Equal(t, 1, model.QueueCount())
	assert.Equal(t, "two", model.PeekQueue())

	// Test DrainQueue
	model.QueueMessage("three")
	queued := model.DrainQueue()
	assert.Equal(t, []string{"two", "three"}, queued)
	assert.Equal(t, 0, model.QueueCount())

	// Test Requeue (adds to front)
	model.QueueMessage("four")
	model.Requeue([]string{"one", "two"})
	assert.Equal(t, 3, model.QueueCount())
	assert.Equal(t, "one", model.PeekQueue())
}
