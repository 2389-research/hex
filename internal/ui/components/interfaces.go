// Package components provides reusable UI components for the Pagen TUI.
// ABOUTME: Core component interfaces defining standard behaviors for all UI components
// ABOUTME: Provides Sizeable, Focusable, Helpable, and Component interfaces for standardization
package components

import tea "github.com/charmbracelet/bubbletea"

// Sizeable components can be resized to fit different terminal dimensions.
// Components implementing this interface can respond to terminal resize events
// and adjust their rendering accordingly.
//
// Example usage:
//
//	type MyComponent struct {
//	    width  int
//	    height int
//	}
//
//	func (m *MyComponent) SetSize(width, height int) tea.Cmd {
//	    m.width = width
//	    m.height = height
//	    return nil
//	}
//
//	func (m *MyComponent) GetSize() (int, int) {
//	    return m.width, m.height
//	}
//
// When a WindowSizeMsg is received, parent components should propagate
// the size to their children:
//
//	case tea.WindowSizeMsg:
//	    if m.child != nil {
//	        cmd := m.child.SetSize(msg.Width, msg.Height)
//	        return m, cmd
//	    }
type Sizeable interface {
	// SetSize updates the component's dimensions and returns any command
	// needed to complete the resize operation. For example, a component
	// might need to invalidate cached content or trigger a re-layout.
	//
	// The width and height parameters represent the available space for
	// the component in terminal cells.
	SetSize(width, height int) tea.Cmd

	// GetSize returns the current dimensions of the component in terminal cells.
	// This allows parent components to query child dimensions for layout calculations.
	GetSize() (int, int)
}

// Focusable components can receive keyboard focus and respond to focus changes.
// When a component is focused, it typically becomes the primary recipient of
// keyboard input and may render differently to indicate its focused state.
//
// Example usage:
//
//	type InputField struct {
//	    value   string
//	    focused bool
//	}
//
//	func (i *InputField) Focus() tea.Cmd {
//	    i.focused = true
//	    return nil
//	}
//
//	func (i *InputField) Blur() tea.Cmd {
//	    i.focused = false
//	    return nil
//	}
//
//	func (i *InputField) IsFocused() bool {
//	    return i.focused
//	}
//
//	func (i *InputField) View() string {
//	    if i.focused {
//	        return focusedStyle.Render(i.value)
//	    }
//	    return normalStyle.Render(i.value)
//	}
//
// Focus management in parent components:
//
//	case key.Matches(msg, m.keymap.NextField):
//	    if m.currentField != nil {
//	        m.currentField.Blur()
//	    }
//	    m.currentField = m.nextField
//	    return m, m.currentField.Focus()
type Focusable interface {
	// Focus gives keyboard focus to the component. This typically triggers
	// visual changes to indicate the focused state and may return a command
	// to perform focus-related initialization (e.g., blinking cursor).
	Focus() tea.Cmd

	// Blur removes keyboard focus from the component. This typically triggers
	// visual changes to indicate the unfocused state and may return a command
	// to perform cleanup (e.g., stop cursor blinking).
	Blur() tea.Cmd

	// IsFocused returns true if the component currently has keyboard focus.
	// This allows parent components to route keyboard events appropriately.
	IsFocused() bool
}

// Helpable components can provide contextual help text to users.
// This interface allows components to document their keyboard shortcuts,
// functionality, and usage in a consistent way.
//
// Example usage:
//
//	type FileManager struct {
//	    // ... fields ...
//	}
//
//	func (f *FileManager) HelpView() string {
//	    return lipgloss.JoinVertical(lipgloss.Left,
//	        "File Manager Help",
//	        "",
//	        "↑/↓    Navigate files",
//	        "enter  Open file",
//	        "d      Delete file",
//	        "r      Rename file",
//	        "?      Toggle help",
//	    )
//	}
//
// Parent components can aggregate help from children:
//
//	func (m *Model) aggregateHelp() string {
//	    var helpSections []string
//	    for _, child := range m.helpableChildren {
//	        if help := child.HelpView(); help != "" {
//	            helpSections = append(helpSections, help)
//	        }
//	    }
//	    return lipgloss.JoinVertical(lipgloss.Left, helpSections...)
//	}
type Helpable interface {
	// HelpView returns a formatted string containing help text for the component.
	// The returned string should be ready to display, including any styling.
	// Return an empty string if the component has no help to display.
	HelpView() string
}

// Component is the base interface that all UI components should implement.
// It combines the standard Bubble Tea Model interface with size management,
// ensuring that components can both handle the Bubble Tea update cycle and
// respond to terminal resize events.
//
// By implementing Component, a type automatically satisfies tea.Model and
// Sizeable, making it compatible with the Bubble Tea framework while
// supporting our size propagation system.
//
// Example usage:
//
//	type ChatMessage struct {
//	    content string
//	    width   int
//	    height  int
//	}
//
//	// Implement tea.Model
//	func (c *ChatMessage) Init() tea.Cmd {
//	    return nil
//	}
//
//	func (c *ChatMessage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
//	    return c, nil
//	}
//
//	func (c *ChatMessage) View() string {
//	    return lipgloss.NewStyle().
//	        Width(c.width).
//	        Render(c.content)
//	}
//
//	// Implement Sizeable
//	func (c *ChatMessage) SetSize(width, height int) tea.Cmd {
//	    c.width = width
//	    c.height = height
//	    return nil
//	}
//
//	func (c *ChatMessage) GetSize() (int, int) {
//	    return c.width, c.height
//	}
//
//	// Now ChatMessage implements Component and can be used throughout the UI
//
// Interface composition pattern:
//
//	// A component that is both focusable and provides help
//	type FullFeaturedComponent interface {
//	    Component
//	    Focusable
//	    Helpable
//	}
//
//	// Check if a component implements optional interfaces
//	if focusable, ok := component.(Focusable); ok {
//	    cmd := focusable.Focus()
//	}
//
//	if helpable, ok := component.(Helpable); ok {
//	    helpText := helpable.HelpView()
//	}
type Component interface {
	tea.Model
	Sizeable
}
