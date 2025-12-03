// ABOUTME: Tests for Bubbletea UI update function
// ABOUTME: Validates keyboard navigation, view switching, and event handling
package ui_test

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/harper/pagent/internal/core"
	"github.com/harper/pagent/internal/ui"
	"github.com/stretchr/testify/assert"
)

func TestUpdateTabKeySwitchesViews(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929", "dracula")
	assert.Equal(t, ui.ViewModeChat, model.CurrentView)

	// Press Tab to switch to History
	msg := tea.KeyMsg{Type: tea.KeyTab}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(*ui.Model)
	assert.Equal(t, ui.ViewModeHistory, m.CurrentView)

	// Press Tab again to switch to Tools
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(*ui.Model)
	assert.Equal(t, ui.ViewModeTools, m.CurrentView)

	// Press Tab again to cycle back to Chat
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(*ui.Model)
	assert.Equal(t, ui.ViewModeChat, m.CurrentView)
}

func TestUpdateVimNavigation(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929", "dracula")
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
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929", "dracula")
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
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929", "dracula")
	model.EnterSearchMode()
	assert.True(t, model.SearchMode)

	// Press Esc to exit search mode
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(*ui.Model)
	assert.False(t, m.SearchMode)
}

func TestUpdateGGGoesToTop(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929", "dracula")
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
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929", "dracula")
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
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929", "dracula")
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
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929", "dracula")
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
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929", "dracula")
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
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929", "dracula")
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
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929", "dracula")
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
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929", "dracula")
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
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929", "dracula")
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
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929", "dracula")
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

// Phase 2: WindowSizeMsg Forwarding Tests

func TestWindowSizeMsgForwardedToApprovalForm(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929", "dracula")
	model.Ready = true

	// Add a pending tool use to trigger approval mode
	toolUse := &core.ToolUse{
		Type:  "tool_use",
		ID:    "test-tool-1",
		Name:  "bash",
		Input: map[string]interface{}{"command": "echo test"},
	}
	model.AddPendingToolUse(toolUse)

	// Enter Huh approval mode
	model.EnterHuhApprovalMode()

	// Verify approval form was created
	assert.NotNil(t, model.GetHuhApproval())
	assert.True(t, model.IsToolApprovalMode())

	// Send WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  100,
		Height: 50,
	}

	updatedModel, _ := model.Update(msg)
	m := updatedModel.(*ui.Model)

	// Verify model dimensions were updated
	assert.Equal(t, 100, m.Width)
	assert.Equal(t, 50, m.Height)

	// Verify approval form still exists (should have received the message)
	assert.NotNil(t, m.GetHuhApproval())
}

func TestInitialWindowSizeMsgSentOnApprovalMode(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929", "dracula")
	model.Ready = true

	// Set initial window size
	model.Width = 80
	model.Height = 24

	// Add a pending tool use
	toolUse := &core.ToolUse{
		Type:  "tool_use",
		ID:    "test-tool-1",
		Name:  "bash",
		Input: map[string]interface{}{"command": "echo test"},
	}
	model.AddPendingToolUse(toolUse)

	// Enter Huh approval mode - should send initial WindowSizeMsg
	model.EnterHuhApprovalMode()

	// Verify approval form was created
	approval := model.GetHuhApproval()
	assert.NotNil(t, approval)

	// The approval form should have received the initial WindowSizeMsg
	// We can verify this by checking that the form renders properly
	view := approval.View()
	assert.NotEmpty(t, view)
}

func TestWindowSizeResizeInApprovalMode(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929", "dracula")
	model.Ready = true

	// Add a pending tool use
	toolUse := &core.ToolUse{
		Type:  "tool_use",
		ID:    "test-tool-1",
		Name:  "bash",
		Input: map[string]interface{}{"command": "echo test"},
	}
	model.AddPendingToolUse(toolUse)

	// Enter approval mode
	model.EnterHuhApprovalMode()

	// Send first resize
	msg1 := tea.WindowSizeMsg{Width: 80, Height: 24}
	updatedModel, _ := model.Update(msg1)
	m := updatedModel.(*ui.Model)

	assert.Equal(t, 80, m.Width)
	assert.Equal(t, 24, m.Height)

	// Send second resize
	msg2 := tea.WindowSizeMsg{Width: 120, Height: 40}
	updatedModel, _ = m.Update(msg2)
	m = updatedModel.(*ui.Model)

	assert.Equal(t, 120, m.Width)
	assert.Equal(t, 40, m.Height)

	// Approval form should still be active
	assert.True(t, m.IsToolApprovalMode())
	assert.NotNil(t, m.GetHuhApproval())
}

func TestAllMessageTypesForwardedToApprovalForm(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929", "dracula")
	model.Ready = true

	// Add a pending tool use
	toolUse := &core.ToolUse{
		Type:  "tool_use",
		ID:    "test-tool-1",
		Name:  "bash",
		Input: map[string]interface{}{"command": "echo test"},
	}
	model.AddPendingToolUse(toolUse)

	// Enter approval mode
	model.EnterHuhApprovalMode()

	// Test KeyMsg forwarding (already tested above, but verify it doesn't break)
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}}
	updatedModel, _ := model.Update(keyMsg)
	m := updatedModel.(*ui.Model)
	assert.NotNil(t, m) // Should not panic

	// Test WindowSizeMsg forwarding
	winMsg := tea.WindowSizeMsg{Width: 100, Height: 50}
	updatedModel, _ = m.Update(winMsg)
	m = updatedModel.(*ui.Model)
	assert.NotNil(t, m) // Should not panic
}

func TestApprovalFormWorksInDifferentTerminalSizes(t *testing.T) {
	// Test that approval forms work across various terminal sizes
	sizes := []struct {
		width  int
		height int
	}{
		{80, 24},  // Standard terminal
		{120, 40}, // Large terminal
		{60, 20},  // Small terminal
		{200, 60}, // Very large terminal
	}

	for _, size := range sizes {
		t.Run(fmt.Sprintf("%dx%d", size.width, size.height), func(t *testing.T) {
			model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929", "dracula")
			model.Ready = true

			// Set initial size
			model.Width = size.width
			model.Height = size.height

			// Add pending tool
			toolUse := &core.ToolUse{
				Type:  "tool_use",
				ID:    "test-tool-1",
				Name:  "bash",
				Input: map[string]interface{}{"command": "echo test"},
			}
			model.AddPendingToolUse(toolUse)

			// Enter approval mode
			model.EnterHuhApprovalMode()

			// Verify approval form renders
			approval := model.GetHuhApproval()
			assert.NotNil(t, approval)

			view := approval.View()
			assert.NotEmpty(t, view)
		})
	}
}
