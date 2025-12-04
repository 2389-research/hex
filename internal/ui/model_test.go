// ABOUTME: Tests for Bubbletea UI model
// ABOUTME: Validates model initialization, state transitions, message handling
package ui_test

import (
	"testing"

	"github.com/harper/jeff/internal/core"
	"github.com/harper/jeff/internal/ui"
	"github.com/stretchr/testify/assert"
)

func TestNewModel(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929", "dracula")

	assert.Equal(t, "conv-123", model.ConversationID)
	assert.Equal(t, "claude-sonnet-4-5-20250929", model.Model)
	assert.NotNil(t, model.Input)
	assert.NotNil(t, model.Viewport)
}

func TestModelAddMessage(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929", "dracula")

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
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929", "dracula")

	// Should start in chat view mode
	assert.Equal(t, ui.ViewModeChat, model.CurrentView)
}

func TestViewModeSwitching(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929", "dracula")

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
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929", "dracula")

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
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929", "dracula")

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
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929", "dracula")

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
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929", "dracula")

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
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929", "dracula")

	model.AppendStreamingText("Hello")
	assert.Equal(t, "Hello", model.StreamingText)

	model.AppendStreamingText(" world")
	assert.Equal(t, "Hello world", model.StreamingText)
}

func TestCommitStreamingText(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929", "dracula")
	model.StreamingText = "Streamed content"

	model.CommitStreamingText()

	assert.Equal(t, "", model.StreamingText)
	assert.Equal(t, 1, len(model.Messages))
	assert.Equal(t, "assistant", model.Messages[0].Role)
	assert.Equal(t, "Streamed content", model.Messages[0].Content)
}

func TestClearStreamingText(t *testing.T) {
	model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929", "dracula")
	model.StreamingText = "Partial content"

	model.ClearStreamingText()

	assert.Equal(t, "", model.StreamingText)
	assert.Equal(t, 0, len(model.Messages))
}

func TestThemeIntegration(t *testing.T) {
	t.Run("initializes with dracula theme", func(t *testing.T) {
		model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929", "dracula")

		assert.NotNil(t, model.GetTheme())
		assert.Equal(t, "Dracula", model.GetTheme().Name())
	})

	t.Run("initializes with gruvbox theme", func(t *testing.T) {
		model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929", "gruvbox")

		assert.NotNil(t, model.GetTheme())
		assert.Equal(t, "Gruvbox Dark", model.GetTheme().Name())
	})

	t.Run("initializes with nord theme", func(t *testing.T) {
		model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929", "nord")

		assert.NotNil(t, model.GetTheme())
		assert.Equal(t, "Nord", model.GetTheme().Name())
	})

	t.Run("defaults to dracula for unknown theme", func(t *testing.T) {
		model := ui.NewModel("conv-123", "claude-sonnet-4-5-20250929", "unknown-theme")

		assert.NotNil(t, model.GetTheme())
		assert.Equal(t, "Dracula", model.GetTheme().Name())
	})
}

// Phase 2: Huh Integration Tests

func TestModelHuhApprovalIntegration(t *testing.T) {
	model := ui.NewModel("test-conv", "claude-sonnet-4", "dracula")

	// Initially no approval in progress
	assert.False(t, model.IsToolApprovalMode())
	assert.Nil(t, model.GetHuhApproval())

	// Add a pending tool
	toolUse := &core.ToolUse{
		ID:    "tool-123",
		Name:  "bash",
		Input: map[string]interface{}{"command": "echo test"},
	}
	model.AddPendingToolUse(toolUse)

	// Enter approval mode
	model.EnterHuhApprovalMode()

	assert.True(t, model.IsToolApprovalMode())
	assert.NotNil(t, model.GetHuhApproval())
}

func TestModelExitHuhApprovalMode(t *testing.T) {
	model := ui.NewModel("test-conv", "claude-sonnet-4", "dracula")

	// Add a pending tool
	toolUse := &core.ToolUse{
		ID:    "tool-123",
		Name:  "bash",
		Input: map[string]interface{}{"command": "echo test"},
	}
	model.AddPendingToolUse(toolUse)

	// Enter and exit approval mode
	model.EnterHuhApprovalMode()
	assert.True(t, model.IsToolApprovalMode())

	model.ExitHuhApprovalMode()
	assert.False(t, model.IsToolApprovalMode())
	assert.Nil(t, model.GetHuhApproval())
}

// Phase 1: Message Formatting Tests

func TestMessageBulletFormatting(t *testing.T) {
	model := ui.NewModel("test-conv", "claude-sonnet-4", "dracula")
	model.Ready = true // Ensure model is ready for viewport rendering

	// Add user and assistant messages
	model.AddMessage("user", "Hello")
	model.AddMessage("assistant", "Hi there")

	// Update viewport to render messages
	model.UpdateViewport()

	// Get rendered content
	content := model.Viewport.View()

	// User messages should NOT have bullet
	assert.Contains(t, content, "You: Hello")
	assert.NotContains(t, content, "● You:")

	// Assistant messages SHOULD have ● bullet
	assert.Contains(t, content, "● Assistant:")
}

func TestStreamingMessageBulletFormatting(t *testing.T) {
	model := ui.NewModel("test-conv", "claude-sonnet-4", "dracula")
	model.Ready = true

	// Add user message first
	model.AddMessage("user", "Hello")

	// Set streaming text
	model.StreamingText = "This is streaming..."

	// Update viewport
	model.UpdateViewport()

	// Get rendered content
	content := model.Viewport.View()

	// Streaming assistant messages should have ● bullet
	assert.Contains(t, content, "● Assistant:")
}

func TestMultipleAssistantMessagesBullets(t *testing.T) {
	model := ui.NewModel("test-conv", "claude-sonnet-4", "dracula")
	model.Ready = true

	// Add multiple messages
	model.AddMessage("user", "First question")
	model.AddMessage("assistant", "First answer")
	model.AddMessage("user", "Second question")
	model.AddMessage("assistant", "Second answer")

	// Update viewport
	model.UpdateViewport()

	// Get rendered content
	content := model.Viewport.View()

	// Count occurrences of assistant bullet
	bulletCount := 0
	for i := 0; i < len(content)-len("● Assistant:"); i++ {
		if content[i:i+len("● Assistant:")] == "● Assistant:" {
			bulletCount++
		}
	}

	// Should have exactly 2 assistant message bullets
	assert.Equal(t, 2, bulletCount, "Expected exactly 2 assistant message bullets")

	// User messages should not have bullets
	assert.Contains(t, content, "You: First question")
	assert.Contains(t, content, "You: Second question")
}
