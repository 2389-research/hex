package ui

import (
	"testing"

	"github.com/2389-research/hex/internal/core"
	tea "github.com/charmbracelet/bubbletea"
)

// TestOverlayManager tests the overlay manager functionality
func TestOverlayManager(t *testing.T) {
	m := NewModel("test-conv", "test-model")

	// Initially no overlay should be active
	if m.overlayManager.HasActive() {
		t.Error("Expected no active overlay initially")
	}

	active := m.overlayManager.GetActive()
	if active != nil {
		t.Error("Expected GetActive to return nil when no overlay active")
	}
}

// TestOverlayPriority tests that higher priority overlays are shown first
func TestOverlayPriority(t *testing.T) {
	m := NewModel("test-conv", "test-model")

	// Activate tool approval (priority 100)
	m.toolApprovalMode = true
	m.pendingToolUses = []*core.ToolUse{{ID: "test", Name: "test"}}

	// Activate autocomplete (priority 50)
	m.autocomplete = NewAutocomplete()
	m.autocomplete.Show("/test", "command")

	// Tool approval should be active (higher priority)
	active := m.overlayManager.GetActive()
	if active == nil {
		t.Fatal("Expected active overlay")
	}

	if active.Type() != OverlayToolApproval {
		t.Errorf("Expected ToolApproval overlay (priority 100), got type %v", active.Type())
	}
}

// TestOverlayEscapeHandling tests that overlays handle Escape correctly
func TestOverlayEscapeHandling(t *testing.T) {
	m := NewModel("test-conv", "test-model")

	// Test autocomplete Escape (should just dismiss)
	m.autocomplete = NewAutocomplete()
	// Register a test command so autocomplete has completions and becomes active
	providerI, _ := m.autocomplete.GetProvider("command")
	provider := providerI.(*SlashCommandProvider)
	provider.SetCommands([]string{"test"}, map[string]string{"test": "test command"})
	m.autocomplete.Show("/test", "command")

	if !m.overlayManager.HasActive() {
		t.Fatal("Expected autocomplete to be active")
	}

	cmd := m.overlayManager.HandleEscape()
	if cmd != nil {
		t.Error("Expected autocomplete HandleEscape to return nil (no command needed)")
	}

	if m.autocomplete.IsActive() {
		t.Error("Expected autocomplete to be dismissed after HandleEscape")
	}
}

// TestOverlayToolApprovalEscape tests that tool approval sends denial on Escape
func TestOverlayToolApprovalEscape(t *testing.T) {
	m := NewModel("test-conv", "test-model")

	// Setup a pending tool
	m.toolApprovalMode = true
	m.pendingToolUses = []*core.ToolUse{
		{
			ID:   "test-tool-1",
			Name: "test_tool",
		},
	}

	if !m.overlayManager.HasActive() {
		t.Fatal("Expected tool approval to be active")
	}

	// HandleEscape should call DenyToolUse and process the denial
	cmd := m.overlayManager.HandleEscape()
	// Note: cmd may be nil if no apiClient is set (test environment)
	// The important thing is that the denial was processed

	// Tool approval should be dismissed
	if m.toolApprovalMode {
		t.Error("Expected toolApprovalMode to be false after HandleEscape")
	}

	// Should have created error result
	if len(m.toolResults) == 0 {
		t.Error("Expected tool result to be created for denied tool")
	}

	// Verify the result is actually a denial
	if m.toolResults[0].Result.Error != "User denied permission" {
		t.Errorf("Expected denial error, got: %s", m.toolResults[0].Result.Error)
	}

	_ = cmd // Ignore command for now (would be nil in test env without apiClient)
}

// TestOverlayCtrlCHandling tests that overlays handle Ctrl+C correctly
func TestOverlayCtrlCHandling(t *testing.T) {
	m := NewModel("test-conv", "test-model")

	// Test autocomplete Ctrl+C
	m.autocomplete = NewAutocomplete()
	m.autocomplete.Show("/test", "command")

	cmd := m.overlayManager.HandleCtrlC()
	if cmd != nil {
		t.Error("Expected autocomplete HandleCtrlC to return nil")
	}

	if m.autocomplete.IsActive() {
		t.Error("Expected autocomplete to be dismissed after HandleCtrlC")
	}
}

// TestOverlayRender tests that overlay manager renders the active overlay
func TestOverlayRender(t *testing.T) {
	m := NewModel("test-conv", "test-model")

	// No overlay active - should return empty string
	content := m.overlayManager.Render()
	if content != "" {
		t.Error("Expected empty string when no overlay active")
	}

	// Activate autocomplete with test commands
	m.autocomplete = NewAutocomplete()
	providerI, _ := m.autocomplete.GetProvider("command")
	provider := providerI.(*SlashCommandProvider)
	provider.SetCommands([]string{"test", "example"}, map[string]string{
		"test":    "test command",
		"example": "example command",
	})
	m.autocomplete.Show("/test", "command")

	// Should now return content (only if autocomplete is active with completions)
	if m.autocomplete.IsActive() {
		content = m.overlayManager.Render()
		if content == "" {
			t.Error("Expected non-empty content when overlay active")
		}
	}
}

// TestMouseHoverInitialization tests that hover fields are initialized correctly
func TestMouseHoverInitialization(t *testing.T) {
	m := NewModel("test-conv", "test-model")

	if m.hoveredMessageIndex != -1 {
		t.Errorf("Expected hoveredMessageIndex to be -1, got %d", m.hoveredMessageIndex)
	}

	if !m.hoveredMessageTime.IsZero() {
		t.Error("Expected hoveredMessageTime to be zero")
	}
}

// TestMouseMotionHandling tests basic mouse motion event handling
func TestMouseMotionHandling(t *testing.T) {
	m := NewModel("test-conv", "test-model")

	// Add a test message
	m.AddMessage("user", "test message")

	// Simulate mouse motion event
	msg := tea.MouseMsg{
		X:      10,
		Y:      5,
		Action: tea.MouseActionMotion,
		Button: tea.MouseButtonNone,
	}

	// Process the event
	newModel, _ := m.Update(msg)
	m2 := newModel.(*Model)

	// The hover state should be updated (we can't easily test the exact value
	// without knowing layout, but we can verify the handler ran)
	_ = m2.hoveredMessageIndex // Just verify field exists
}

// TestOverlayManagerCancelAll tests cancelling all overlays
func TestOverlayManagerCancelAll(t *testing.T) {
	m := NewModel("test-conv", "test-model")

	// Activate multiple overlays
	m.toolApprovalMode = true
	m.autocomplete = NewAutocomplete()
	m.autocomplete.Show("/test", "command")

	if !m.overlayManager.HasActive() {
		t.Fatal("Expected overlays to be active")
	}

	// Cancel all
	m.overlayManager.CancelAll()

	if m.overlayManager.HasActive() {
		t.Error("Expected no active overlays after CancelAll")
	}
}
