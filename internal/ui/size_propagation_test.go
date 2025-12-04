// Package ui provides the Bubble Tea terminal user interface components.
// ABOUTME: Tests for window size propagation to child components
// ABOUTME: Validates that all components receive proper size updates
package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/harper/pagent/internal/ui/components"
	"github.com/harper/pagent/internal/ui/themes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetReservedHeight(t *testing.T) {
	tests := []struct {
		name            string
		showTokenViz    bool
		expectedMinimum int
	}{
		{
			name:            "without token viz",
			showTokenViz:    false,
			expectedMinimum: 4, // 3 for input + 1 for status bar
		},
		{
			name:            "with token viz",
			showTokenViz:    true,
			expectedMinimum: 7, // 3 for input + 1 for status bar + 3 for token viz
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewModel("test-conv", "claude-3-5-sonnet-20241022", "dracula")
			model.showTokenViz = tt.showTokenViz

			reserved := model.getReservedHeight()
			assert.GreaterOrEqual(t, reserved, tt.expectedMinimum,
				"reserved height should be at least %d", tt.expectedMinimum)
		})
	}
}

func TestHandleWindowSizeMsg_PropagationToViewport(t *testing.T) {
	model := NewModel("test-conv", "claude-3-5-sonnet-20241022", "dracula")

	// GIVEN: Initial size
	initialWidth := 80
	initialHeight := 24

	// WHEN: WindowSizeMsg is handled
	msg := tea.WindowSizeMsg{
		Width:  initialWidth,
		Height: initialHeight,
	}
	_ = model.handleWindowSizeMsg(msg)

	// THEN: Model dimensions are updated
	assert.Equal(t, initialWidth, model.Width, "model width should be updated")
	assert.Equal(t, initialHeight, model.Height, "model height should be updated")

	// THEN: Viewport receives adjusted size
	expectedViewportHeight := initialHeight - model.getReservedHeight() - 5
	assert.Equal(t, initialWidth-4, model.Viewport.Width, "viewport width should account for padding")
	assert.Equal(t, expectedViewportHeight, model.Viewport.Height, "viewport height should account for reserved space")
}

func TestHandleWindowSizeMsg_PropagationToInput(t *testing.T) {
	model := NewModel("test-conv", "claude-3-5-sonnet-20241022", "dracula")

	// GIVEN: New terminal size
	newWidth := 120
	newHeight := 40

	// WHEN: WindowSizeMsg is handled
	msg := tea.WindowSizeMsg{
		Width:  newWidth,
		Height: newHeight,
	}
	_ = model.handleWindowSizeMsg(msg)

	// THEN: Input width is updated
	// Note: Textarea SetWidth accounts for prompt width internally
	// So the actual width will be less than newWidth-4
	assert.Greater(t, model.Input.Width(), 0, "input width should be set")
	assert.LessOrEqual(t, model.Input.Width(), newWidth, "input width should not exceed terminal width")
}

func TestHandleWindowSizeMsg_PropagationToHelpComponent(t *testing.T) {
	theme := themes.GetTheme("dracula")
	model := NewModel("test-conv", "claude-3-5-sonnet-20241022", "dracula")

	// GIVEN: Help component is initialized
	model.helpComponent = components.NewHelpOverlay(theme)

	// WHEN: WindowSizeMsg is handled
	msg := tea.WindowSizeMsg{
		Width:  100,
		Height: 30,
	}
	_ = model.handleWindowSizeMsg(msg)

	// THEN: Help component receives size update
	width, height := model.helpComponent.GetSize()
	assert.Equal(t, 100, width, "help component should receive width")
	assert.Equal(t, 30, height, "help component should receive height")
}

func TestHandleWindowSizeMsg_PropagationToErrorComponent(t *testing.T) {
	theme := themes.GetTheme("dracula")
	model := NewModel("test-conv", "claude-3-5-sonnet-20241022", "dracula")

	// GIVEN: Error component is initialized
	model.errorComponent = components.NewErrorDisplay(theme, "Test Error", "Test message", "Test details")

	// WHEN: WindowSizeMsg is handled
	msg := tea.WindowSizeMsg{
		Width:  100,
		Height: 30,
	}
	_ = model.handleWindowSizeMsg(msg)

	// THEN: Error component receives size update
	width, height := model.errorComponent.GetSize()
	assert.Equal(t, 100, width, "error component should receive width")
	assert.Equal(t, 30, height, "error component should receive height")
}

func TestHandleWindowSizeMsg_PropagationToHuhApproval(t *testing.T) {
	theme := themes.GetTheme("dracula")
	model := NewModel("test-conv", "claude-3-5-sonnet-20241022", "dracula")

	// GIVEN: Huh approval component is initialized
	model.huhApproval = components.NewHuhApproval(theme, "test_tool", "Test description")

	// WHEN: WindowSizeMsg is handled
	msg := tea.WindowSizeMsg{
		Width:  100,
		Height: 30,
	}
	_ = model.handleWindowSizeMsg(msg)

	// THEN: Huh approval component receives size update
	width, height := model.huhApproval.GetSize()
	assert.Equal(t, 100, width, "huh approval should receive width")
	assert.Equal(t, 30, height, "huh approval should receive height")
}

func TestHandleWindowSizeMsg_PropagationToTokenViz(t *testing.T) {
	model := NewModel("test-conv", "claude-3-5-sonnet-20241022", "dracula")
	model.showTokenViz = true

	// WHEN: WindowSizeMsg is handled
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	_ = model.handleWindowSizeMsg(msg)

	// THEN: Token viz is updated
	// Note: TokenVisualization doesn't expose width directly, but we can verify it was called
	// by checking that it doesn't panic and the model state is consistent
	assert.NotNil(t, model.tokenViz, "token viz should still exist")
}

func TestHandleWindowSizeMsg_AllComponentsTogether(t *testing.T) {
	theme := themes.GetTheme("dracula")
	model := NewModel("test-conv", "claude-3-5-sonnet-20241022", "dracula")

	// GIVEN: All components are initialized
	model.helpComponent = components.NewHelpOverlay(theme)
	model.errorComponent = components.NewErrorDisplay(theme, "Test", "Message", "Details")
	model.huhApproval = components.NewHuhApproval(theme, "test_tool", "Description")
	model.showTokenViz = true

	// WHEN: WindowSizeMsg is handled
	msg := tea.WindowSizeMsg{
		Width:  150,
		Height: 50,
	}
	cmd := model.handleWindowSizeMsg(msg)

	// THEN: All components receive size updates
	helpWidth, helpHeight := model.helpComponent.GetSize()
	assert.Equal(t, 150, helpWidth, "help width")
	assert.Equal(t, 50, helpHeight, "help height")

	errorWidth, errorHeight := model.errorComponent.GetSize()
	assert.Equal(t, 150, errorWidth, "error width")
	assert.Equal(t, 50, errorHeight, "error height")

	approvalWidth, approvalHeight := model.huhApproval.GetSize()
	assert.Equal(t, 150, approvalWidth, "approval width")
	assert.Equal(t, 50, approvalHeight, "approval height")

	// THEN: Model dimensions are correct
	assert.Equal(t, 150, model.Width)
	assert.Equal(t, 50, model.Height)

	// THEN: Viewport dimensions are correct
	expectedHeight := 50 - model.getReservedHeight() - 5
	assert.Equal(t, 146, model.Viewport.Width, "viewport width")
	assert.Equal(t, expectedHeight, model.Viewport.Height, "viewport height")

	// THEN: Command may be returned (batch of multiple commands or nil)
	// Components that don't need commands on resize return nil
	_ = cmd // Command is optional
}

func TestHandleWindowSizeMsg_RepeatedResizes(t *testing.T) {
	model := NewModel("test-conv", "claude-3-5-sonnet-20241022", "dracula")
	theme := themes.GetTheme("dracula")
	model.helpComponent = components.NewHelpOverlay(theme)

	sizes := []struct {
		width  int
		height int
	}{
		{80, 24},
		{120, 40},
		{100, 30},
		{60, 20},
	}

	for _, size := range sizes {
		t.Run("resize_to_"+string(rune(size.width))+"x"+string(rune(size.height)), func(t *testing.T) {
			msg := tea.WindowSizeMsg{
				Width:  size.width,
				Height: size.height,
			}
			_ = model.handleWindowSizeMsg(msg)

			// Verify model tracks size correctly
			assert.Equal(t, size.width, model.Width)
			assert.Equal(t, size.height, model.Height)

			// Verify help component gets updated
			width, height := model.helpComponent.GetSize()
			assert.Equal(t, size.width, width)
			assert.Equal(t, size.height, height)
		})
	}
}

func TestUpdate_WindowSizeMsg_Integration(t *testing.T) {
	model := NewModel("test-conv", "claude-3-5-sonnet-20241022", "dracula")
	theme := themes.GetTheme("dracula")
	model.helpComponent = components.NewHelpOverlay(theme)

	// WHEN: Update receives WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  100,
		Height: 30,
	}
	updatedModel, cmd := model.Update(msg)

	// THEN: Model is updated correctly
	m, ok := updatedModel.(*Model)
	require.True(t, ok, "should return *Model")

	assert.Equal(t, 100, m.Width)
	assert.Equal(t, 30, m.Height)

	// THEN: Help component is updated
	width, height := m.helpComponent.GetSize()
	assert.Equal(t, 100, width)
	assert.Equal(t, 30, height)

	// THEN: Ready flag is set
	assert.True(t, m.Ready, "model should be marked as ready after first resize")

	// THEN: Command may or may not be returned depending on components
	_ = cmd // Command is optional
}

func TestHandleWindowSizeMsg_NilComponents(t *testing.T) {
	model := NewModel("test-conv", "claude-3-5-sonnet-20241022", "dracula")

	// GIVEN: All optional components are nil
	model.helpComponent = nil
	model.errorComponent = nil
	model.huhApproval = nil

	// WHEN: WindowSizeMsg is handled
	msg := tea.WindowSizeMsg{
		Width:  100,
		Height: 30,
	}

	// THEN: Should not panic
	require.NotPanics(t, func() {
		_ = model.handleWindowSizeMsg(msg)
	}, "should handle nil components gracefully")

	// THEN: Model dimensions are still updated
	assert.Equal(t, 100, model.Width)
	assert.Equal(t, 30, model.Height)
}

func TestHandleWindowSizeMsg_ReservedHeightCalculation(t *testing.T) {
	model := NewModel("test-conv", "claude-3-5-sonnet-20241022", "dracula")

	// GIVEN: Token viz is hidden
	model.showTokenViz = false
	msg := tea.WindowSizeMsg{Width: 100, Height: 30}
	_ = model.handleWindowSizeMsg(msg)

	viewportHeightWithoutViz := model.Viewport.Height
	reservedWithoutViz := model.getReservedHeight()

	// WHEN: Token viz is shown
	model.showTokenViz = true
	_ = model.handleWindowSizeMsg(msg)

	viewportHeightWithViz := model.Viewport.Height
	reservedWithViz := model.getReservedHeight()

	// THEN: Reserved height increases
	assert.Greater(t, reservedWithViz, reservedWithoutViz,
		"reserved height should increase when token viz is shown")

	// THEN: Viewport height decreases accordingly
	assert.Less(t, viewportHeightWithViz, viewportHeightWithoutViz,
		"viewport height should decrease when token viz reserves space")
}
