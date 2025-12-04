// Package ui provides the Bubble Tea terminal user interface components.
// ABOUTME: Tests for the StatusBar component
// ABOUTME: Validates rendering, token tracking, and connection state management
package ui

import (
	"strings"
	"testing"

	"github.com/harper/jeff/internal/ui/themes"
)

func TestNewStatusBar(t *testing.T) {
	theme := themes.GetTheme("dracula")
	model := "claude-sonnet-4"
	width := 100

	sb := NewStatusBar(model, width, theme)

	if sb == nil {
		t.Fatal("NewStatusBar returned nil")
	}

	if sb.model != model {
		t.Errorf("expected model %q, got %q", model, sb.model)
	}

	if sb.width != width {
		t.Errorf("expected width %d, got %d", width, sb.width)
	}

	if sb.connection != ConnectionDisconnected {
		t.Errorf("expected initial connection status ConnectionDisconnected, got %v", sb.connection)
	}

	if sb.currentMode != "chat" {
		t.Errorf("expected default mode %q, got %q", "chat", sb.currentMode)
	}

	if sb.contextSize != 200000 {
		t.Errorf("expected default context size 200000, got %d", sb.contextSize)
	}
}

func TestSetWidth(t *testing.T) {
	theme := themes.GetTheme("dracula")
	sb := NewStatusBar("claude-sonnet-4", 80, theme)

	newWidth := 120
	sb.SetWidth(newWidth)

	if sb.width != newWidth {
		t.Errorf("expected width %d, got %d", newWidth, sb.width)
	}
}

func TestSetModel(t *testing.T) {
	theme := themes.GetTheme("dracula")
	sb := NewStatusBar("claude-sonnet-4", 80, theme)

	newModel := "claude-opus-4"
	sb.SetModel(newModel)

	if sb.model != newModel {
		t.Errorf("expected model %q, got %q", newModel, sb.model)
	}
}

func TestUpdateTokens(t *testing.T) {
	theme := themes.GetTheme("dracula")
	sb := NewStatusBar("claude-sonnet-4", 80, theme)

	// Initial state
	if sb.tokensInput != 0 || sb.tokensOutput != 0 || sb.tokensTotal != 0 {
		t.Error("expected initial tokens to be zero")
	}

	// First update
	sb.UpdateTokens(100, 200)
	if sb.tokensInput != 100 || sb.tokensOutput != 200 || sb.tokensTotal != 300 {
		t.Errorf("after first update: expected 100/200/300, got %d/%d/%d",
			sb.tokensInput, sb.tokensOutput, sb.tokensTotal)
	}

	// Second update (cumulative)
	sb.UpdateTokens(50, 75)
	if sb.tokensInput != 150 || sb.tokensOutput != 275 || sb.tokensTotal != 425 {
		t.Errorf("after second update: expected 150/275/425, got %d/%d/%d",
			sb.tokensInput, sb.tokensOutput, sb.tokensTotal)
	}
}

func TestSetTokens(t *testing.T) {
	theme := themes.GetTheme("dracula")
	sb := NewStatusBar("claude-sonnet-4", 80, theme)

	// Set some initial values
	sb.UpdateTokens(100, 200)

	// SetTokens should replace, not add
	sb.SetTokens(500, 600)
	if sb.tokensInput != 500 || sb.tokensOutput != 600 || sb.tokensTotal != 1100 {
		t.Errorf("after SetTokens: expected 500/600/1100, got %d/%d/%d",
			sb.tokensInput, sb.tokensOutput, sb.tokensTotal)
	}
}

func TestSetContextSize(t *testing.T) {
	theme := themes.GetTheme("dracula")
	sb := NewStatusBar("claude-sonnet-4", 80, theme)

	newSize := 150000
	sb.SetContextSize(newSize)

	if sb.contextSize != newSize {
		t.Errorf("expected context size %d, got %d", newSize, sb.contextSize)
	}
}

func TestSetConnectionStatus(t *testing.T) {
	theme := themes.GetTheme("dracula")
	sb := NewStatusBar("claude-sonnet-4", 80, theme)

	tests := []struct {
		name   string
		status ConnectionStatus
	}{
		{"disconnected", ConnectionDisconnected},
		{"connected", ConnectionConnected},
		{"streaming", ConnectionStreaming},
		{"error", ConnectionError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sb.SetConnectionStatus(tt.status)
			if sb.connection != tt.status {
				t.Errorf("expected connection status %v, got %v", tt.status, sb.connection)
			}
		})
	}
}

func TestSetConnection(t *testing.T) {
	theme := themes.GetTheme("dracula")
	sb := NewStatusBar("claude-sonnet-4", 80, theme)

	// Test alias method
	sb.SetConnection(ConnectionStreaming)
	if sb.connection != ConnectionStreaming {
		t.Errorf("SetConnection alias failed: expected %v, got %v", ConnectionStreaming, sb.connection)
	}
}

func TestSetMode(t *testing.T) {
	theme := themes.GetTheme("dracula")
	sb := NewStatusBar("claude-sonnet-4", 80, theme)

	modes := []string{"chat", "history", "tools"}
	for _, mode := range modes {
		sb.SetMode(mode)
		if sb.currentMode != mode {
			t.Errorf("expected mode %q, got %q", mode, sb.currentMode)
		}
	}
}

func TestSetCustomMessage(t *testing.T) {
	theme := themes.GetTheme("dracula")
	sb := NewStatusBar("claude-sonnet-4", 80, theme)

	msg := "Test warning message"
	sb.SetCustomMessage(msg)

	if sb.customMessage != msg {
		t.Errorf("expected custom message %q, got %q", msg, sb.customMessage)
	}

	// Test clear
	sb.ClearCustomMessage()
	if sb.customMessage != "" {
		t.Errorf("expected empty custom message after clear, got %q", sb.customMessage)
	}
}

func TestRenderOutput(t *testing.T) {
	theme := themes.GetTheme("dracula")
	sb := NewStatusBar("claude-sonnet-4", 100, theme)

	// Test basic rendering
	output := sb.View()
	if output == "" {
		t.Error("View() returned empty string")
	}

	// Should contain model name
	if !strings.Contains(output, "claude-sonnet-4") {
		t.Error("View() output should contain model name")
	}

	// Should contain mode
	if !strings.Contains(output, "chat") {
		t.Error("View() output should contain current mode")
	}
}

func TestRenderWithTokens(t *testing.T) {
	theme := themes.GetTheme("dracula")
	sb := NewStatusBar("claude-sonnet-4", 100, theme)

	sb.SetTokens(5000, 3000)
	output := sb.View()

	// Should display token information (5k and 3k)
	if !strings.Contains(output, "5k") || !strings.Contains(output, "3k") {
		t.Error("View() output should contain token counts when tokens are set")
	}
}

func TestRenderNarrowWidth(t *testing.T) {
	theme := themes.GetTheme("dracula")
	sb := NewStatusBar("claude-sonnet-4", 30, theme)

	output := sb.View()
	if output == "" {
		t.Error("View() should still render something with narrow width")
	}

	// Should show minimal info
	if !strings.Contains(output, "Jeff") {
		t.Error("Narrow width should show app name")
	}
}

func TestRenderConnectionStates(t *testing.T) {
	theme := themes.GetTheme("dracula")
	sb := NewStatusBar("claude-sonnet-4", 100, theme)

	tests := []struct {
		status   ConnectionStatus
		expected string
	}{
		{ConnectionDisconnected, "○"},
		{ConnectionConnected, "●"},
		{ConnectionStreaming, "◉"},
		{ConnectionError, "◉"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			sb.SetConnectionStatus(tt.status)
			output := sb.View()
			if !strings.Contains(output, tt.expected) {
				t.Errorf("expected connection indicator %q in output for status %v", tt.expected, tt.status)
			}
		})
	}
}

func TestRenderContextIndicator(t *testing.T) {
	theme := themes.GetTheme("dracula")
	sb := NewStatusBar("claude-sonnet-4", 100, theme)

	// Set context size
	sb.SetContextSize(100000)

	// Set tokens to more than half of context (should show indicator)
	sb.SetTokens(60000, 0)

	output := sb.View()

	// Should show context indicator
	if !strings.Contains(output, "[") {
		t.Error("View() should show context indicator when tokens exceed 50% of context size")
	}
}

func TestWidthAdaptation(t *testing.T) {
	theme := themes.GetTheme("dracula")
	sb := NewStatusBar("claude-sonnet-4", 80, theme)

	widths := []int{40, 80, 120, 160}
	for _, width := range widths {
		sb.SetWidth(width)
		output := sb.View()
		if output == "" {
			t.Errorf("View() returned empty string for width %d", width)
		}
		// The output should adapt to different widths
		// At minimum, we check it doesn't crash
	}
}

func TestGetFullHelp(t *testing.T) {
	theme := themes.GetTheme("dracula")
	sb := NewStatusBar("claude-sonnet-4", 80, theme)

	help := sb.GetFullHelp()
	if help == "" {
		t.Error("GetFullHelp() returned empty string")
	}

	// Should contain some keyboard shortcuts
	expectedShortcuts := []string{"Ctrl+C", "Enter", "Tab"}
	for _, shortcut := range expectedShortcuts {
		if !strings.Contains(help, shortcut) {
			t.Errorf("GetFullHelp() should contain %q", shortcut)
		}
	}
}

func TestApplyUpdate(t *testing.T) {
	theme := themes.GetTheme("dracula")
	sb := NewStatusBar("claude-sonnet-4", 80, theme)

	// Create an update with all fields
	connected := ConnectionConnected
	mode := "history"
	message := "Test message"

	update := StatusBarUpdate{
		Tokens: &TokenUpdate{
			Input:  1000,
			Output: 2000,
		},
		Connection: &connected,
		Mode:       &mode,
		Message:    &message,
	}

	sb.ApplyUpdate(update)

	// Verify all fields were updated
	if sb.tokensInput != 1000 || sb.tokensOutput != 2000 {
		t.Errorf("tokens not updated correctly: got %d/%d", sb.tokensInput, sb.tokensOutput)
	}
	if sb.connection != ConnectionConnected {
		t.Errorf("connection not updated correctly: got %v", sb.connection)
	}
	if sb.currentMode != "history" {
		t.Errorf("mode not updated correctly: got %q", sb.currentMode)
	}
	if sb.customMessage != "Test message" {
		t.Errorf("message not updated correctly: got %q", sb.customMessage)
	}
}
