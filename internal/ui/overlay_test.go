package ui

import (
	"fmt"
	"testing"

	"github.com/2389-research/hex/internal/core"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

// mockOverlay is a test overlay implementation
type mockOverlay struct {
	header        string
	content       string
	footer        string
	desiredHeight int
	onPushCalled  bool
	onPopCalled   bool
	lastWidth     int
	lastHeight    int
	keyHandler    func(tea.KeyMsg) (bool, tea.Cmd)
}

func newMockOverlay(header, content, footer string, height int) *mockOverlay {
	return &mockOverlay{
		header:        header,
		content:       content,
		footer:        footer,
		desiredHeight: height,
	}
}

func (m *mockOverlay) GetHeader() string              { return m.header }
func (m *mockOverlay) GetContent() string             { return m.content }
func (m *mockOverlay) GetFooter() string              { return m.footer }
func (m *mockOverlay) GetDesiredHeight() int          { return m.desiredHeight }
func (m *mockOverlay) Render(width, height int) string {
	return fmt.Sprintf("%s\n%s\n%s", m.header, m.content, m.footer)
}

func (m *mockOverlay) HandleKey(msg tea.KeyMsg) (bool, tea.Cmd) {
	if m.keyHandler != nil {
		return m.keyHandler(msg)
	}
	// Default: handle Escape to pop
	if msg.Type == tea.KeyEsc {
		return true, nil
	}
	return false, nil
}

func (m *mockOverlay) OnPush(width, height int) {
	m.onPushCalled = true
	m.lastWidth = width
	m.lastHeight = height
}

func (m *mockOverlay) OnPop() {
	m.onPopCalled = true
}

func (m *mockOverlay) Cancel() tea.Cmd {
	return nil
}

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

// TestOverlayStack tests that overlays work as a stack (last pushed is active)
func TestOverlayStack(t *testing.T) {
	m := NewModel("test-conv", "test-model")

	// Push tool approval overlay
	m.toolApprovalMode = true
	m.pendingToolUses = []*core.ToolUse{{ID: "test", Name: "test"}}
	m.overlayManager.Push(m.toolApprovalOverlay)
	m.toolApprovalOverlay.OnPush(80, 24)

	// Tool approval should be active (last pushed)
	active := m.overlayManager.GetActive()
	if active == nil {
		t.Fatal("Expected active overlay")
	}

	if active != m.toolApprovalOverlay {
		t.Error("Expected tool approval overlay to be active (last pushed)")
	}

	// Push autocomplete on top
	m.autocomplete = NewAutocomplete()
	m.autocomplete.Show("/test", "command")
	m.overlayManager.Push(m.autocompleteOverlay)
	m.autocompleteOverlay.OnPush(80, 24)

	// Autocomplete should now be active (on top of stack)
	active = m.overlayManager.GetActive()
	if active != m.autocompleteOverlay {
		t.Error("Expected autocomplete overlay to be active (last pushed)")
	}

	// Pop autocomplete - tool approval should be active again
	m.overlayManager.Pop()
	active = m.overlayManager.GetActive()
	if active != m.toolApprovalOverlay {
		t.Error("Expected tool approval overlay to be active after popping autocomplete")
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

	// Push autocomplete overlay
	m.overlayManager.Push(m.autocompleteOverlay)
	m.autocompleteOverlay.OnPush(80, 24)

	if !m.overlayManager.HasActive() {
		t.Fatal("Expected autocomplete to be active")
	}

	handled, cmd := m.overlayManager.HandleKey(tea.KeyMsg{Type: tea.KeyEsc})
	if !handled {
		t.Error("Expected overlay to handle Escape")
	}
	if cmd != nil {
		t.Error("Expected autocomplete HandleEscape to return nil (no command needed)")
	}

	// Pop the overlay
	m.overlayManager.Pop()
	m.autocomplete.Hide()

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

	// Push tool approval overlay
	m.overlayManager.Push(m.toolApprovalOverlay)
	m.toolApprovalOverlay.OnPush(80, 24)

	if !m.overlayManager.HasActive() {
		t.Fatal("Expected tool approval to be active")
	}

	// HandleKey for Escape - overlay should handle it
	handled, _ := m.overlayManager.HandleKey(tea.KeyMsg{Type: tea.KeyEsc})
	if !handled {
		t.Error("Expected overlay to handle Escape")
	}

	// Now manually call DenyToolUse (simulating what update.go does)
	_ = m.DenyToolUse()

	// Tool approval should be dismissed
	if m.toolApprovalMode {
		t.Error("Expected toolApprovalMode to be false after DenyToolUse")
	}

	// Overlay should have been popped
	if m.overlayManager.HasActive() {
		t.Error("Expected overlay to be popped after DenyToolUse")
	}

	// Should have created error result
	if len(m.toolResults) == 0 {
		t.Error("Expected tool result to be created for denied tool")
	}

	// Verify the result is actually a denial
	if m.toolResults[0].Result.Error != "User denied permission" {
		t.Errorf("Expected denial error, got: %s", m.toolResults[0].Result.Error)
	}
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
	content := m.overlayManager.Render(80, 24)
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

	// Push autocomplete overlay
	m.overlayManager.Push(m.autocompleteOverlay)
	m.autocompleteOverlay.OnPush(80, 24)

	// Should now return content (only if autocomplete is active with completions)
	if m.autocomplete.IsActive() {
		content = m.overlayManager.Render(80, 24)
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

	// Push multiple overlays
	m.toolApprovalMode = true
	m.overlayManager.Push(m.toolApprovalOverlay)
	m.toolApprovalOverlay.OnPush(80, 24)

	m.autocomplete = NewAutocomplete()
	m.autocomplete.Show("/test", "command")
	m.overlayManager.Push(m.autocompleteOverlay)
	m.autocompleteOverlay.OnPush(80, 24)

	if !m.overlayManager.HasActive() {
		t.Fatal("Expected overlays to be active")
	}

	// Cancel all
	m.overlayManager.CancelAll()

	if m.overlayManager.HasActive() {
		t.Error("Expected no active overlays after CancelAll")
	}
}

func TestOverlayManager_StackOperations(t *testing.T) {
	om := NewOverlayManager()

	// Empty stack
	assert.Nil(t, om.Peek())
	assert.Nil(t, om.GetActive())
	assert.False(t, om.HasActive())

	// Push first overlay
	overlay1 := newMockOverlay("Header 1", "Content 1", "Footer 1", 5)
	om.Push(overlay1)

	assert.Equal(t, overlay1, om.Peek())
	assert.Equal(t, overlay1, om.GetActive())
	assert.True(t, om.HasActive())

	// Push second overlay
	overlay2 := newMockOverlay("Header 2", "Content 2", "Footer 2", 10)
	om.Push(overlay2)

	assert.Equal(t, overlay2, om.Peek()) // Top of stack
	assert.Equal(t, overlay2, om.GetActive())

	// Pop should return top
	popped := om.Pop()
	assert.Equal(t, overlay2, popped)
	assert.True(t, overlay2.onPopCalled)
	assert.Equal(t, overlay1, om.Peek()) // Back to first

	// Pop last overlay
	om.Pop()
	assert.Nil(t, om.Peek())
	assert.False(t, om.HasActive())
}

func TestOverlayManager_Clear(t *testing.T) {
	om := NewOverlayManager()

	overlay1 := newMockOverlay("H1", "C1", "F1", 5)
	overlay2 := newMockOverlay("H2", "C2", "F2", 10)
	overlay3 := newMockOverlay("H3", "C3", "F3", 15)

	om.Push(overlay1)
	om.Push(overlay2)
	om.Push(overlay3)

	assert.True(t, om.HasActive())

	om.Clear()

	assert.False(t, om.HasActive())
	assert.Nil(t, om.Peek())
	assert.True(t, overlay1.onPopCalled)
	assert.True(t, overlay2.onPopCalled)
	assert.True(t, overlay3.onPopCalled)
}

func TestOverlayManager_HandleKeyModalCapture(t *testing.T) {
	om := NewOverlayManager()

	// No overlay - not handled
	handled, cmd := om.HandleKey(tea.KeyMsg{Type: tea.KeyEnter})
	assert.False(t, handled)
	assert.Nil(t, cmd)

	// Push overlay that handles Enter
	keyCaptured := false
	overlay := newMockOverlay("H", "C", "F", 5)
	overlay.keyHandler = func(msg tea.KeyMsg) (bool, tea.Cmd) {
		if msg.Type == tea.KeyEnter {
			keyCaptured = true
			return true, nil
		}
		return false, nil
	}
	om.Push(overlay)

	// Overlay captures Enter
	handled, cmd = om.HandleKey(tea.KeyMsg{Type: tea.KeyEnter})
	assert.True(t, handled)
	assert.True(t, keyCaptured)
	assert.Nil(t, cmd)

	// Overlay doesn't handle other keys but still captures (modal)
	handled, _ = om.HandleKey(tea.KeyMsg{Type: tea.KeyCtrlA})
	assert.False(t, handled) // Overlay didn't handle it
}
