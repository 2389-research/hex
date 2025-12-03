// ABOUTME: Tests for Bubbletea UI update function
// ABOUTME: Validates keyboard navigation, view switching, and event handling
package ui_test

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/harper/hex/internal/ui"
	"github.com/stretchr/testify/assert"
)

func TestUpdateTabKeySwitchesViews(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929")
	assert.Equal(t, ui.ViewModeIntro, model.CurrentView)

	// Press Tab to switch to Chat
	msg := tea.KeyMsg{Type: tea.KeyTab}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(*ui.Model)
	assert.Equal(t, ui.ViewModeChat, m.CurrentView)

	// Press Tab again to switch to History
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(*ui.Model)
	assert.Equal(t, ui.ViewModeHistory, m.CurrentView)

	// Press Tab again to switch to Tools
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(*ui.Model)
	assert.Equal(t, ui.ViewModeTools, m.CurrentView)

	// Press Tab again to cycle back to Intro
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(*ui.Model)
	assert.Equal(t, ui.ViewModeIntro, m.CurrentView)
}

func TestUpdateVimNavigation(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929")
	model.Ready = true

	// Add some messages to scroll
	for i := 0; i < 10; i++ {
		model.AddMessage("user", "Test message")
	}

	// Test 'j' key for scrolling down
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(*ui.Model)
	assert.NotNil(t, m.Viewport)

	// Test 'k' key for scrolling up
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(*ui.Model)
	assert.NotNil(t, m.Viewport)
}

func TestUpdateSlashEntersSearchMode(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929")
	model.CurrentView = ui.ViewModeChat // Skip intro for this test
	assert.False(t, model.SearchMode)

	// Blur the textarea first (slash only works when textarea not focused)
	model.Input.Blur()

	// Press '/' to enter search mode
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(*ui.Model)
	assert.True(t, m.SearchMode)
}

func TestUpdateEscExitsSearchMode(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929")
	model.CurrentView = ui.ViewModeChat // Skip intro for this test
	model.EnterSearchMode()
	assert.True(t, model.SearchMode)

	// Press Esc to exit search mode
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(*ui.Model)
	assert.False(t, m.SearchMode)
}

func TestUpdateGGGoesToTop(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929")
	model.Ready = true

	// Add messages and scroll down
	for i := 0; i < 20; i++ {
		model.AddMessage("user", "Test message")
	}

	// Press 'g' twice to go to top (vim-style)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(*ui.Model)

	// Second 'g' should go to top
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(*ui.Model)
	assert.Equal(t, 0, m.Viewport.YOffset)
}

func TestUpdateShiftGGoesToBottom(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929")
	model.Ready = true

	// Add messages
	for i := 0; i < 20; i++ {
		model.AddMessage("user", "Test message")
	}

	// Press 'G' (shift+g) to go to bottom
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(*ui.Model)
	assert.True(t, m.Viewport.AtBottom())
}

func TestSearchModeBackspace(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929")
	model.CurrentView = ui.ViewModeChat // Skip intro for this test
	model.EnterSearchMode()
	model.UpdateSearchQuery("test")

	// Simulate backspace
	msg := tea.KeyMsg{Type: tea.KeyBackspace}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(*ui.Model)

	assert.Equal(t, "tes", m.SearchQuery)

	// Backspace on empty should not panic
	m.SearchQuery = ""
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(*ui.Model)
	assert.Equal(t, "", m.SearchQuery)
}

func TestGGSequenceResetOnOtherKeys(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929")
	model.Ready = true
	model.AddMessage("user", "Test message")

	// Press 'g' once
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(*ui.Model)

	// Press a non-rune key (e.g., arrow key)
	msg2 := tea.KeyMsg{Type: tea.KeyUp}
	updatedModel, _ = m.Update(msg2)
	m = updatedModel.(*ui.Model)

	// Now pressing 'g' again should not trigger gg (viewport should not jump to top)
	initialYOffset := m.Viewport.YOffset
	msg3 := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}}
	updatedModel, _ = m.Update(msg3)
	m = updatedModel.(*ui.Model)
	assert.Equal(t, initialYOffset, m.Viewport.YOffset) // Should not jump to top
}

// Task 6: Streaming Integration Tests

func TestStreamChunkAppendsText(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929")
	model.Ready = true

	// Simulate receiving a stream chunk message
	chunk := &ui.StreamChunkMsg{
		Chunk: &ui.StreamChunk{
			Type: "content_block_delta",
			Delta: &ui.Delta{
				Type: "text_delta",
				Text: "Hello",
			},
		},
	}

	updatedModel, _ := model.Update(chunk)
	m := updatedModel.(*ui.Model)

	assert.Equal(t, "Hello", m.StreamingText)
	assert.Equal(t, ui.StatusStreaming, m.Status)
}

func TestStreamChunkAccumulation(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929")
	model.Ready = true

	// Receive multiple chunks
	chunk1 := &ui.StreamChunkMsg{
		Chunk: &ui.StreamChunk{
			Type: "content_block_delta",
			Delta: &ui.Delta{
				Type: "text_delta",
				Text: "Hello ",
			},
		},
	}

	chunk2 := &ui.StreamChunkMsg{
		Chunk: &ui.StreamChunk{
			Type: "content_block_delta",
			Delta: &ui.Delta{
				Type: "text_delta",
				Text: "world",
			},
		},
	}

	updatedModel, _ := model.Update(chunk1)
	m := updatedModel.(*ui.Model)
	updatedModel, _ = m.Update(chunk2)
	m = updatedModel.(*ui.Model)

	assert.Equal(t, "Hello world", m.StreamingText)
}

func TestStreamMessageStopCommitsText(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929")
	model.Ready = true
	model.StreamingText = "Completed response"

	// Simulate message_stop event
	chunk := &ui.StreamChunkMsg{
		Chunk: &ui.StreamChunk{
			Type: "message_stop",
			Done: true,
		},
	}

	updatedModel, _ := model.Update(chunk)
	m := updatedModel.(*ui.Model)

	// Streaming text should be committed to messages
	assert.Equal(t, "", m.StreamingText)
	assert.Len(t, m.Messages, 1)
	assert.Equal(t, "assistant", m.Messages[0].Role)
	assert.Equal(t, "Completed response", m.Messages[0].Content)
	assert.Equal(t, ui.StatusIdle, m.Status)
}

func TestStreamErrorClearsStreamingText(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929")
	model.Ready = true
	model.StreamingText = "Partial response"

	// Simulate error
	chunk := &ui.StreamChunkMsg{
		Error: fmt.Errorf("API error"),
	}

	updatedModel, _ := model.Update(chunk)
	m := updatedModel.(*ui.Model)

	// Streaming text should be cleared
	assert.Equal(t, "", m.StreamingText)
	assert.Equal(t, ui.StatusError, m.Status)
	assert.NotEmpty(t, m.ErrorMessage)
}

func TestStreamUsageUpdatesTokens(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929")
	model.Ready = true

	initialInput := model.TokensInput
	initialOutput := model.TokensOutput

	// Simulate usage metadata
	chunk := &ui.StreamChunkMsg{
		Chunk: &ui.StreamChunk{
			Type: "message_delta",
			Usage: &ui.Usage{
				InputTokens:  100,
				OutputTokens: 50,
			},
		},
	}

	updatedModel, _ := model.Update(chunk)
	m := updatedModel.(*ui.Model)

	assert.Equal(t, initialInput+100, m.TokensInput)
	assert.Equal(t, initialOutput+50, m.TokensOutput)
}

func TestEnterKeyTriggersStreaming(t *testing.T) {
	// This test will need a mock API client
	// For now, we'll just verify that the model state changes appropriately
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929")
	model.CurrentView = ui.ViewModeChat // Skip intro for this test
	model.Ready = true

	// Set some input text
	model.Input.SetValue("Hello AI")

	// Simulate Enter key press
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updatedModel, cmd := model.Update(msg)
	m := updatedModel.(*ui.Model)

	// User message should be added
	assert.Len(t, m.Messages, 1)
	assert.Equal(t, "user", m.Messages[0].Role)
	assert.Equal(t, "Hello AI", m.Messages[0].Content)

	// Input should be cleared (may contain newline from textarea)
	assert.Empty(t, strings.TrimSpace(m.Input.Value()))

	// Should return a command (for streaming)
	// Note: We can't fully test this without mocking the API
	// The command will be nil if no API client is set
	_ = cmd
}
