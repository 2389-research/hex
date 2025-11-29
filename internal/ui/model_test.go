// ABOUTME: Tests for Bubbletea UI model
// ABOUTME: Validates model initialization, state transitions, message handling
package ui_test

import (
	"testing"

	"github.com/harper/clem/internal/ui"
	"github.com/stretchr/testify/assert"
)

func TestNewModel(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929")

	assert.Equal(t, "conv-123", model.ConversationID)
	assert.Equal(t, "claude-sonnet-4-5-20250929", model.Model)
	assert.NotNil(t, model.Input)
	assert.NotNil(t, model.Viewport)
}

func TestModelAddMessage(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929")

	model.AddMessage("user", "Hello")
	model.AddMessage("assistant", "Hi there")

	assert.Len(t, model.Messages, 2)
	assert.Equal(t, "user", model.Messages[0].Role)
	assert.Equal(t, "Hello", model.Messages[0].Content)
	assert.Equal(t, "assistant", model.Messages[1].Role)
	assert.Equal(t, "Hi there", model.Messages[1].Content)
}

// Task 5: Advanced UI Features Tests

func TestViewModeInitialization(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929")

	// Should start in chat view mode
	assert.Equal(t, ui.ViewModeChat, model.CurrentView)
}

func TestViewModeSwitching(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929")

	// Switch from Chat to History
	model.NextView()
	assert.Equal(t, ui.ViewModeHistory, model.CurrentView)

	// Switch from History to Tools
	model.NextView()
	assert.Equal(t, ui.ViewModeTools, model.CurrentView)

	// Switch from Tools back to Chat (cycle)
	model.NextView()
	assert.Equal(t, ui.ViewModeChat, model.CurrentView)
}

func TestTokenCounterTracking(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929")

	// Initial state
	assert.Equal(t, 0, model.TokensInput)
	assert.Equal(t, 0, model.TokensOutput)

	// Update token counts
	model.UpdateTokens(100, 250)
	assert.Equal(t, 100, model.TokensInput)
	assert.Equal(t, 250, model.TokensOutput)

	// Cumulative tracking
	model.UpdateTokens(50, 75)
	assert.Equal(t, 150, model.TokensInput)
	assert.Equal(t, 325, model.TokensOutput)
}

func TestStatusIndicators(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929")

	// Initial status should be idle
	assert.Equal(t, ui.StatusIdle, model.Status)

	// Set streaming status
	model.SetStatus(ui.StatusStreaming)
	assert.Equal(t, ui.StatusStreaming, model.Status)

	// Set error status
	model.SetStatus(ui.StatusError)
	assert.Equal(t, ui.StatusError, model.Status)
	assert.NotEmpty(t, model.ErrorMessage)
}

func TestMarkdownRendering(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929")

	// Add message with markdown content
	model.AddMessage("assistant", "# Header\n\n**bold** and *italic*")

	// Render should use glamour for assistant messages
	rendered, err := model.RenderMessage(model.Messages[0])
	assert.NoError(t, err)
	assert.NotEmpty(t, rendered)
	// Rendered output should be different from raw markdown
	assert.NotEqual(t, "# Header\n\n**bold** and *italic*", rendered)
}

func TestSearchMode(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929")

	// Initial state
	assert.False(t, model.SearchMode)
	assert.Empty(t, model.SearchQuery)

	// Enter search mode
	model.EnterSearchMode()
	assert.True(t, model.SearchMode)

	// Update search query
	model.UpdateSearchQuery("test query")
	assert.Equal(t, "test query", model.SearchQuery)

	// Exit search mode
	model.ExitSearchMode()
	assert.False(t, model.SearchMode)
	assert.Empty(t, model.SearchQuery)
}

func TestAppendStreamingText(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929")

	model.AppendStreamingText("Hello")
	assert.Equal(t, "Hello", model.StreamingText)

	model.AppendStreamingText(" world")
	assert.Equal(t, "Hello world", model.StreamingText)
}

func TestCommitStreamingText(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929")
	model.StreamingText = "Streamed content"

	model.CommitStreamingText()

	assert.Equal(t, "", model.StreamingText)
	assert.Equal(t, 1, len(model.Messages))
	assert.Equal(t, "assistant", model.Messages[0].Role)
	assert.Equal(t, "Streamed content", model.Messages[0].Content)
}

func TestClearStreamingText(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929")
	model.StreamingText = "Partial content"

	model.ClearStreamingText()

	assert.Equal(t, "", model.StreamingText)
	assert.Equal(t, 0, len(model.Messages))
}
