package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// Compile-time interface checks
var _ Overlay = (*GenericFullscreenOverlay)(nil)
var _ Scrollable = (*GenericFullscreenOverlay)(nil)
var _ FullscreenOverlay = (*GenericFullscreenOverlay)(nil)

func TestGenericFullscreenOverlay_InterfaceCompliance(t *testing.T) {
	overlay := NewHelpOverlay()

	// Test FullscreenOverlay interface
	if !overlay.IsFullscreen() {
		t.Error("Expected IsFullscreen() to return true")
	}

	if overlay.GetDesiredHeight() != -1 {
		t.Error("Expected GetDesiredHeight() to return -1 for fullscreen")
	}

	// Test Overlay interface methods exist
	_ = overlay.GetHeader()
	_ = overlay.GetContent()
	_ = overlay.GetFooter()

	overlay.OnPush(100, 40)
	overlay.SetHeight(50)
	overlay.OnPop()

	_ = overlay.Render(100, 40)
	_ = overlay.Cancel()
}

func TestGenericFullscreenOverlay_ViewportInitialization(t *testing.T) {
	overlay := NewHelpOverlay()
	overlay.OnPush(100, 40)

	// Verify viewport was initialized with correct dimensions
	// width = 100 - 4 = 96
	// height = 40 - 6 = 34
	if overlay.viewport.Width != 96 {
		t.Errorf("Expected viewport width 96, got %d", overlay.viewport.Width)
	}
	if overlay.viewport.Height != 34 {
		t.Errorf("Expected viewport height 34, got %d", overlay.viewport.Height)
	}
}

func TestGenericFullscreenOverlay_HandleKey(t *testing.T) {
	overlay := NewHelpOverlay()
	overlay.OnPush(100, 40)

	tests := []struct {
		name        string
		keyType     tea.KeyType
		shouldClose bool
	}{
		{"Escape closes", tea.KeyEsc, true},
		{"Ctrl+C closes", tea.KeyCtrlC, true},
		{"Ctrl+H closes (toggle key)", tea.KeyCtrlH, true},
		{"Up arrow navigates", tea.KeyUp, false},
		{"Down arrow navigates", tea.KeyDown, false},
		{"PageUp navigates", tea.KeyPgUp, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handled, cmd := overlay.HandleKey(tea.KeyMsg{Type: tt.keyType})
			if !handled {
				t.Error("Expected key to be handled")
			}

			// Close keys return nil cmd, navigation keys may return viewport cmd
			if tt.shouldClose && cmd != nil {
				t.Error("Expected nil command for close key")
			}
		})
	}
}

func TestGenericFullscreenOverlay_SetHeight(t *testing.T) {
	overlay := NewHelpOverlay()
	overlay.OnPush(100, 40)

	overlay.SetHeight(60)

	// height = 60 - 6 = 54
	if overlay.viewport.Height != 54 {
		t.Errorf("Expected viewport height 54 after SetHeight(60), got %d", overlay.viewport.Height)
	}
}

func TestGenericFullscreenOverlay_Update(t *testing.T) {
	overlay := NewHelpOverlay()
	overlay.OnPush(100, 40)

	// Test window resize
	msg := tea.WindowSizeMsg{Width: 120, Height: 50}
	_ = overlay.Update(msg)

	// width = 120 - 4 = 116
	// height = 50 - 6 = 44
	if overlay.viewport.Width != 116 {
		t.Errorf("Expected viewport width 116 after resize, got %d", overlay.viewport.Width)
	}
	if overlay.viewport.Height != 44 {
		t.Errorf("Expected viewport height 44 after resize, got %d", overlay.viewport.Height)
	}
}

func TestGenericFullscreenOverlay_SmallTerminalGuard(t *testing.T) {
	overlay := NewHelpOverlay()

	// Test with very small terminal dimensions
	overlay.OnPush(2, 2)

	// Should guard against negative dimensions (min 1)
	if overlay.viewport.Width < 1 {
		t.Errorf("Expected viewport width >= 1, got %d", overlay.viewport.Width)
	}
	if overlay.viewport.Height < 1 {
		t.Errorf("Expected viewport height >= 1, got %d", overlay.viewport.Height)
	}
}

func TestHelpContentProvider_Interface(t *testing.T) {
	provider := &HelpContentProvider{}

	// Test all required interface methods (3 methods - Footer is optional)
	if provider.Header() == "" {
		t.Error("Expected non-empty header")
	}
	if provider.Content() == "" {
		t.Error("Expected non-empty content")
	}

	toggleKeys := provider.ToggleKeys()
	if len(toggleKeys) == 0 {
		t.Error("Expected at least one toggle key")
	}
	if toggleKeys[0] != tea.KeyCtrlH {
		t.Error("Expected first toggle key to be Ctrl+H")
	}

	// HelpContentProvider should NOT implement FooterProvider (uses default)
	_, hasCustomFooter := interface{}(provider).(FooterProvider)
	if hasCustomFooter {
		t.Error("HelpContentProvider should not implement FooterProvider - should use default")
	}
}

func TestGenericFullscreenOverlay_DefaultFooter(t *testing.T) {
	overlay := NewHelpOverlay()
	overlay.OnPush(80, 24)

	// Footer should be auto-generated from ToggleKeys
	footer := overlay.GetFooter()
	expected := "Ctrl+H or Esc to close"
	if footer != expected {
		t.Errorf("Expected default footer %q, got %q", expected, footer)
	}
}

func TestBuildCloseHint(t *testing.T) {
	tests := []struct {
		name     string
		keys     []tea.KeyType
		expected string
	}{
		{"empty keys defaults to Esc", []tea.KeyType{}, "Esc to close"},
		{"single key adds Esc", []tea.KeyType{tea.KeyCtrlH}, "Ctrl+H or Esc to close"},
		{"Esc only", []tea.KeyType{tea.KeyEsc}, "Esc to close"},
		{"multiple keys", []tea.KeyType{tea.KeyCtrlO, tea.KeyCtrlH}, "Ctrl+O or Ctrl+H or Esc to close"},
		{"includes Esc already", []tea.KeyType{tea.KeyCtrlR, tea.KeyEsc}, "Ctrl+R or Esc to close"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildCloseHint(tt.keys)
			if result != tt.expected {
				t.Errorf("buildCloseHint(%v) = %q, want %q", tt.keys, result, tt.expected)
			}
		})
	}
}

func TestKeyTypeName(t *testing.T) {
	tests := []struct {
		key      tea.KeyType
		expected string
	}{
		{tea.KeyEsc, "Esc"},
		{tea.KeyCtrlC, "Ctrl+C"},
		{tea.KeyCtrlH, "Ctrl+H"},
		{tea.KeyCtrlO, "Ctrl+O"},
		{tea.KeyCtrlR, "Ctrl+R"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := keyTypeName(tt.key)
			if result != tt.expected {
				t.Errorf("keyTypeName(%v) = %q, want %q", tt.key, result, tt.expected)
			}
		})
	}
}

func TestNewHelpOverlay_Factory(t *testing.T) {
	overlay := NewHelpOverlay()

	if overlay == nil {
		t.Fatal("Expected non-nil overlay")
	}

	// Verify it's properly initialized
	if overlay.provider == nil {
		t.Error("Expected provider to be set")
	}
}
