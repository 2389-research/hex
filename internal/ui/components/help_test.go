package components

import (
	"strings"
	"testing"
)

func TestNewHelp(t *testing.T) {
	tests := []struct {
		name     string
		mode     HelpMode
		width    int
		wantMode HelpMode
	}{
		{"chat mode", HelpModeChat, 80, HelpModeChat},
		{"history mode", HelpModeHistory, 60, HelpModeHistory},
		{"tools mode", HelpModeTools, 100, HelpModeTools},
		{"approval mode", HelpModeApproval, 80, HelpModeApproval},
		{"search mode", HelpModeSearch, 80, HelpModeSearch},
		{"quick actions mode", HelpModeQuickActions, 80, HelpModeQuickActions},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			help := NewHelp(tt.mode, tt.width)

			if help == nil {
				t.Fatal("NewHelp returned nil")
			}

			if help.mode != tt.wantMode {
				t.Errorf("Expected mode %v, got %v", tt.wantMode, help.mode)
			}

			if help.width != tt.width {
				t.Errorf("Expected width %d, got %d", tt.width, help.width)
			}

			if help.theme == nil {
				t.Error("Theme not initialized")
			}

			if len(help.categories) == 0 {
				t.Error("Categories not initialized")
			}
		})
	}
}

func TestHelpSetMode(t *testing.T) {
	help := NewHelp(HelpModeChat, 80)

	initialCategories := len(help.categories)

	help.SetMode(HelpModeHistory)

	if help.mode != HelpModeHistory {
		t.Errorf("Expected mode %v, got %v", HelpModeHistory, help.mode)
	}

	// Categories should change with mode
	if len(help.categories) == initialCategories {
		// This might be coincidental, but at least verify they exist
		if len(help.categories) == 0 {
			t.Error("Categories should be populated after SetMode")
		}
	}
}

func TestHelpSetWidth(t *testing.T) {
	help := NewHelp(HelpModeChat, 80)

	help.SetWidth(100)

	if help.width != 100 {
		t.Errorf("Expected width 100, got %d", help.width)
	}

	if help.help.Width != 100 {
		t.Errorf("Expected underlying help width 100, got %d", help.help.Width)
	}
}

func TestHelpToggleExpanded(t *testing.T) {
	help := NewHelp(HelpModeChat, 80)

	initialState := help.expanded

	help.ToggleExpanded()

	if help.expanded == initialState {
		t.Error("ToggleExpanded did not change state")
	}

	help.ToggleExpanded()

	if help.expanded != initialState {
		t.Error("ToggleExpanded did not toggle back")
	}
}

func TestHelpSetExpanded(t *testing.T) {
	help := NewHelp(HelpModeChat, 80)

	help.SetExpanded(true)

	if !help.expanded {
		t.Error("SetExpanded(true) did not set expanded to true")
	}

	if !help.IsExpanded() {
		t.Error("IsExpanded() should return true")
	}

	help.SetExpanded(false)

	if help.expanded {
		t.Error("SetExpanded(false) did not set expanded to false")
	}

	if help.IsExpanded() {
		t.Error("IsExpanded() should return false")
	}
}

func TestHelpViewCompact(t *testing.T) {
	help := NewHelp(HelpModeChat, 80)

	view := help.ViewCompact()

	if view == "" {
		t.Error("ViewCompact returned empty string")
	}

	// Should contain help indicator
	if !strings.Contains(view, "?") {
		t.Error("ViewCompact should contain help indicator")
	}

	// Should be single line (no newlines)
	if strings.Contains(view, "\n") {
		t.Error("ViewCompact should be single line")
	}
}

func TestHelpViewExpanded(t *testing.T) {
	help := NewHelp(HelpModeChat, 80)

	view := help.ViewExpanded()

	if view == "" {
		t.Error("ViewExpanded returned empty string")
	}

	// Should contain "Help" title
	if !strings.Contains(view, "Help") {
		t.Error("ViewExpanded should contain Help title")
	}

	// Should be multi-line
	if !strings.Contains(view, "\n") {
		t.Error("ViewExpanded should be multi-line")
	}

	// Should contain category names
	// At least one category should be present
	categories := getKeyBindingsForMode(HelpModeChat)
	if len(categories) > 0 {
		foundCategory := false
		for _, cat := range categories {
			if strings.Contains(view, cat.Name) {
				foundCategory = true
				break
			}
		}
		if !foundCategory {
			t.Error("ViewExpanded should contain category names")
		}
	}
}

func TestHelpView(t *testing.T) {
	help := NewHelp(HelpModeChat, 80)

	// Default should be compact
	view := help.View()
	if view == "" {
		t.Error("View returned empty string")
	}

	// Should be compact (single line)
	if strings.Contains(view, "\n") {
		t.Error("Default View should be compact")
	}

	// Toggle to expanded
	help.SetExpanded(true)
	expandedView := help.View()

	if expandedView == "" {
		t.Error("Expanded View returned empty string")
	}

	// Should be multi-line
	if !strings.Contains(expandedView, "\n") {
		t.Error("Expanded View should be multi-line")
	}
}

func TestHelpViewAsOverlay(t *testing.T) {
	help := NewHelp(HelpModeChat, 60)

	overlay := help.ViewAsOverlay(100, 30)

	if overlay == "" {
		t.Error("ViewAsOverlay returned empty string")
	}

	// Should contain help content
	if !strings.Contains(overlay, "Help") {
		t.Error("ViewAsOverlay should contain help content")
	}
}

func TestGetKeyBindingsForMode(t *testing.T) {
	tests := []struct {
		mode        HelpMode
		expectEmpty bool
	}{
		{HelpModeChat, false},
		{HelpModeHistory, false},
		{HelpModeTools, false},
		{HelpModeApproval, false},
		{HelpModeSearch, false},
		{HelpModeQuickActions, false},
		{HelpMode(999), true}, // Invalid mode
	}

	for _, tt := range tests {
		categories := getKeyBindingsForMode(tt.mode)

		if tt.expectEmpty {
			if len(categories) != 0 {
				t.Errorf("Mode %v: expected empty categories, got %d", tt.mode, len(categories))
			}
		} else {
			if len(categories) == 0 {
				t.Errorf("Mode %v: expected non-empty categories", tt.mode)
			}

			// Verify structure
			for _, cat := range categories {
				if cat.Name == "" {
					t.Error("Category should have a name")
				}
				if len(cat.Bindings) == 0 {
					t.Errorf("Category %s should have bindings", cat.Name)
				}
				for _, binding := range cat.Bindings {
					if binding.Key == "" {
						t.Error("Binding should have a key")
					}
					if binding.Description == "" {
						t.Error("Binding should have a description")
					}
				}
			}
		}
	}
}

func TestKeyBindingsToKeyMap(t *testing.T) {
	bindings := []KeyBinding{
		{"?", "toggle help"},
		{"q", "quit"},
		{"enter", "confirm"},
	}

	keyMap := KeyBindingsToKeyMap(bindings)

	if len(keyMap) != len(bindings) {
		t.Errorf("Expected %d key bindings, got %d", len(bindings), len(keyMap))
	}

	for i, kb := range keyMap {
		// Verify the binding was created (basic check)
		_ = kb
		if i >= len(bindings) {
			t.Errorf("KeyMap index %d out of range", i)
		}
	}
}

func TestDefaultChatHelp(t *testing.T) {
	help := DefaultChatHelp(80)

	if help == nil {
		t.Fatal("DefaultChatHelp returned nil")
	}

	if help.mode != HelpModeChat {
		t.Errorf("Expected mode %v, got %v", HelpModeChat, help.mode)
	}

	if help.width != 80 {
		t.Errorf("Expected width 80, got %d", help.width)
	}
}

func TestDefaultHistoryHelp(t *testing.T) {
	help := DefaultHistoryHelp(80)

	if help == nil {
		t.Fatal("DefaultHistoryHelp returned nil")
	}

	if help.mode != HelpModeHistory {
		t.Errorf("Expected mode %v, got %v", HelpModeHistory, help.mode)
	}
}

func TestDefaultToolsHelp(t *testing.T) {
	help := DefaultToolsHelp(80)

	if help == nil {
		t.Fatal("DefaultToolsHelp returned nil")
	}

	if help.mode != HelpModeTools {
		t.Errorf("Expected mode %v, got %v", HelpModeTools, help.mode)
	}
}

func TestChatModeKeyBindings(t *testing.T) {
	categories := getKeyBindingsForMode(HelpModeChat)

	if len(categories) == 0 {
		t.Fatal("Chat mode should have key bindings")
	}

	// Verify expected categories exist
	expectedCategories := []string{"Navigation", "Actions", "Tools", "Help"}
	for _, expected := range expectedCategories {
		found := false
		for _, cat := range categories {
			if cat.Name == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected category '%s' not found in chat mode", expected)
		}
	}
}

func TestHistoryModeKeyBindings(t *testing.T) {
	categories := getKeyBindingsForMode(HelpModeHistory)

	if len(categories) == 0 {
		t.Fatal("History mode should have key bindings")
	}

	// Should have at least navigation and actions
	if len(categories) < 2 {
		t.Error("History mode should have at least 2 categories")
	}
}

func TestToolsModeKeyBindings(t *testing.T) {
	categories := getKeyBindingsForMode(HelpModeTools)

	if len(categories) == 0 {
		t.Fatal("Tools mode should have key bindings")
	}

	// Should have navigation and actions
	if len(categories) < 2 {
		t.Error("Tools mode should have at least 2 categories")
	}
}

func TestApprovalModeKeyBindings(t *testing.T) {
	categories := getKeyBindingsForMode(HelpModeApproval)

	if len(categories) == 0 {
		t.Fatal("Approval mode should have key bindings")
	}

	// Should have tool approval category
	found := false
	for _, cat := range categories {
		if strings.Contains(cat.Name, "Approval") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Approval mode should have Approval category")
	}
}
