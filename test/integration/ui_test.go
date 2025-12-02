// ABOUTME: UI integration tests for Bubbletea model lifecycle and flows
// ABOUTME: Tests UI state transitions, message handling, basic UI operations

package integration

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/harper/clem/internal/core"
	"github.com/harper/clem/internal/ui"
	"github.com/stretchr/testify/assert"
)

// TestUIModelInitialization tests creating and initializing the UI model
func TestUIModelInitialization(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929")

	assert.Equal(t, "conv-123", model.ConversationID)
	assert.Equal(t, "claude-sonnet-4-5-20250929", model.Model)
	assert.NotNil(t, model.Input)
	assert.NotNil(t, model.Viewport)
	assert.Empty(t, model.Messages)
	assert.False(t, model.Streaming)
}

// TestUIAddMessage tests adding messages to the model
func TestUIAddMessage(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929")

	model.AddMessage("user", "Hello")
	model.AddMessage("assistant", "Hi there!")
	model.AddMessage("user", "How are you?")

	assert.Len(t, model.Messages, 3)
	assert.Equal(t, "user", model.Messages[0].Role)
	assert.Equal(t, "Hello", model.Messages[0].Content)
	assert.Equal(t, "assistant", model.Messages[1].Role)
	assert.Equal(t, "Hi there!", model.Messages[1].Content)
}

// TestUIStreamingFlow tests streaming message accumulation
func TestUIStreamingFlow(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929")

	// Initially not streaming
	assert.False(t, model.Streaming)
	assert.Empty(t, model.StreamingText)

	// Manually set streaming (in real app, this happens via Update)
	model.Streaming = true

	// Add streaming chunks
	model.AppendStreamingText("Hello ")
	assert.Equal(t, "Hello ", model.StreamingText)

	model.AppendStreamingText("world")
	assert.Equal(t, "Hello world", model.StreamingText)

	model.AppendStreamingText("!")
	assert.Equal(t, "Hello world!", model.StreamingText)

	// Commit streaming text
	model.CommitStreamingText()
	// Note: Streaming flag remains true, must be manually set to false
	assert.Empty(t, model.StreamingText)
	assert.Len(t, model.Messages, 1)
	assert.Equal(t, "assistant", model.Messages[0].Role)
	assert.Equal(t, "Hello world!", model.Messages[0].Content)

	// Manually stop streaming
	model.Streaming = false
	assert.False(t, model.Streaming)
}

// TestUIWindowResize tests that model responds to window size changes
func TestUIWindowResize(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929")

	// Send window size message
	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updatedModel, _ := model.Update(msg)

	m := updatedModel.(*ui.Model)
	assert.Equal(t, 120, m.Width)
	assert.Equal(t, 40, m.Height)
	assert.True(t, m.Ready, "model should be ready after window size")
}

// TestUISetAPIClient tests setting API client
func TestUISetAPIClient(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929")

	client := core.NewClient("test-api-key")
	model.SetAPIClient(client)

	// Can't directly test private field, but we can verify no panic
	assert.NotNil(t, model)
}

// TestUISetDatabase tests setting database
func TestUISetDatabase(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929")
	db := SetupTestDB(t)

	model.SetDB(db)

	// Verify no panic
	assert.NotNil(t, model)
}

// TestUIViewRendering tests that the view can be rendered
func TestUIViewRendering(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929")

	// Initially not ready
	view := model.View()
	assert.Contains(t, view, "Initializing", "should show initializing message")

	// After window size, should be ready
	msg := tea.WindowSizeMsg{Width: 80, Height: 24}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(*ui.Model)

	view = m.View()
	assert.NotContains(t, view, "Initializing", "should not show initializing after ready")
	assert.Contains(t, view, "Clem", "should show title")
}

// TestUIStateTransitions tests various state transitions
func TestUIStateTransitions(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929")

	// State: Not Ready -> Ready
	assert.False(t, model.Ready)
	msg := tea.WindowSizeMsg{Width: 80, Height: 24}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(*ui.Model)
	assert.True(t, m.Ready)

	// State: Streaming text accumulation
	m.Streaming = true
	m.AppendStreamingText("Test")
	assert.Equal(t, "Test", m.StreamingText)

	// Commit and verify clear
	m.CommitStreamingText()
	assert.Empty(t, m.StreamingText)
	assert.Len(t, m.Messages, 1)
}

// TestUIClearStreamingText tests clearing streaming buffer
func TestUIClearStreamingText(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929")

	model.AppendStreamingText("Some partial response")
	assert.NotEmpty(t, model.StreamingText)

	model.ClearStreamingText()
	assert.Empty(t, model.StreamingText)
}

// TestUIUpdateTokens tests token counter updates
func TestUIUpdateTokens(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929")

	assert.Equal(t, 0, model.TokensInput)
	assert.Equal(t, 0, model.TokensOutput)

	model.UpdateTokens(10, 20)
	assert.Equal(t, 10, model.TokensInput)
	assert.Equal(t, 20, model.TokensOutput)

	// Cumulative
	model.UpdateTokens(5, 10)
	assert.Equal(t, 15, model.TokensInput)
	assert.Equal(t, 30, model.TokensOutput)
}

// TestUISetStatus tests setting UI status
func TestUISetStatus(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929")

	assert.Equal(t, ui.StatusIdle, model.Status)

	model.SetStatus(ui.StatusStreaming)
	assert.Equal(t, ui.StatusStreaming, model.Status)

	model.SetStatus(ui.StatusError)
	assert.Equal(t, ui.StatusError, model.Status)
	assert.NotEmpty(t, model.ErrorMessage)
}

// TestUIViewModes tests switching between view modes
func TestUIViewModes(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929")

	assert.Equal(t, ui.ViewModeIntro, model.CurrentView)

	model.NextView()
	assert.Equal(t, ui.ViewModeChat, model.CurrentView)

	model.NextView()
	assert.Equal(t, ui.ViewModeHistory, model.CurrentView)

	model.NextView()
	assert.Equal(t, ui.ViewModeTools, model.CurrentView)

	model.NextView()
	assert.Equal(t, ui.ViewModeIntro, model.CurrentView)
}

// TestUISearchMode tests search mode functionality
func TestUISearchMode(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929")

	assert.False(t, model.SearchMode)

	model.EnterSearchMode()
	assert.True(t, model.SearchMode)
	assert.Empty(t, model.SearchQuery)

	model.UpdateSearchQuery("test query")
	assert.Equal(t, "test query", model.SearchQuery)

	model.ExitSearchMode()
	assert.False(t, model.SearchMode)
	assert.Empty(t, model.SearchQuery)
}
