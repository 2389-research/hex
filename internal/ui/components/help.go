// Package components provides reusable Bubbles components with Dracula theme styling.
// ABOUTME: Help system component for displaying key bindings and context-aware help
// ABOUTME: Wraps bubbles.Help with Dracula theme and mode-based key binding display
package components

import (
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/harper/clem/internal/ui/theme"
)

// HelpMode represents different help contexts
type HelpMode int

const (
	// HelpModeChat shows help for chat interface
	HelpModeChat HelpMode = iota
	// HelpModeHistory shows help for history browser
	HelpModeHistory
	// HelpModeTools shows help for tools view
	HelpModeTools
	// HelpModeApproval shows help for tool approval
	HelpModeApproval
	// HelpModeSearch shows help for search mode
	HelpModeSearch
	// HelpModeQuickActions shows help for quick actions
	HelpModeQuickActions
)

// KeyBinding represents a key binding with description
type KeyBinding struct {
	Key         string
	Description string
}

// KeyBindingCategory represents a category of key bindings
type KeyBindingCategory struct {
	Name     string
	Bindings []KeyBinding
}

// Help wraps bubbles.Help with Dracula styling and context awareness
type Help struct {
	help       help.Model
	theme      *theme.Theme
	mode       HelpMode
	categories []KeyBindingCategory
	expanded   bool
	width      int
}

// NewHelp creates a new help component with Dracula styling
func NewHelp(mode HelpMode, width int) *Help {
	h := help.New()
	draculaTheme := theme.DraculaTheme()

	// Apply Dracula styles
	h.Styles.ShortKey = draculaTheme.HelpKey
	h.Styles.ShortDesc = draculaTheme.HelpDesc
	h.Styles.FullKey = draculaTheme.HelpKey
	h.Styles.FullDesc = draculaTheme.HelpDesc
	h.Styles.ShortSeparator = draculaTheme.Muted
	h.Styles.FullSeparator = draculaTheme.Muted
	h.Styles.Ellipsis = draculaTheme.Muted

	h.Width = width

	categories := getKeyBindingsForMode(mode)

	return &Help{
		help:       h,
		theme:      draculaTheme,
		mode:       mode,
		categories: categories,
		expanded:   false,
		width:      width,
	}
}

// getKeyBindingsForMode returns key bindings for a specific mode
func getKeyBindingsForMode(mode HelpMode) []KeyBindingCategory {
	switch mode {
	case HelpModeChat:
		return []KeyBindingCategory{
			{
				Name: "Navigation",
				Bindings: []KeyBinding{
					{"↑/↓", "scroll up/down"},
					{"ctrl+u/d", "page up/down"},
					{"gg/G", "jump to top/bottom"},
					{"tab", "cycle views"},
				},
			},
			{
				Name: "Actions",
				Bindings: []KeyBinding{
					{"enter", "send message"},
					{"ctrl+c", "clear input"},
					{"ctrl+k", "quick actions"},
					{"ctrl+l", "clear screen"},
					{"ctrl+s", "save conversation"},
					{"ctrl+f", "toggle favorite"},
				},
			},
			{
				Name: "Tools",
				Bindings: []KeyBinding{
					{"y/n", "approve/deny tool"},
					{"a", "always allow tool"},
					{"N", "never allow tool"},
				},
			},
			{
				Name: "Help",
				Bindings: []KeyBinding{
					{"?", "toggle help"},
					{"q/esc", "quit"},
				},
			},
		}

	case HelpModeHistory:
		return []KeyBindingCategory{
			{
				Name: "Navigation",
				Bindings: []KeyBinding{
					{"↑/↓/j/k", "move up/down"},
					{"enter", "load conversation"},
					{"/", "search"},
					{"esc", "back to chat"},
				},
			},
			{
				Name: "Actions",
				Bindings: []KeyBinding{
					{"d", "delete conversation"},
					{"f", "toggle favorite"},
					{"e", "export conversation"},
				},
			},
		}

	case HelpModeTools:
		return []KeyBindingCategory{
			{
				Name: "Navigation",
				Bindings: []KeyBinding{
					{"↑/↓/j/k", "move up/down"},
					{"enter", "view tool details"},
					{"esc", "back to chat"},
				},
			},
			{
				Name: "Actions",
				Bindings: []KeyBinding{
					{"e", "enable/disable tool"},
					{"c", "configure tool"},
					{"r", "refresh tools"},
				},
			},
		}

	case HelpModeApproval:
		return []KeyBindingCategory{
			{
				Name: "Tool Approval",
				Bindings: []KeyBinding{
					{"y", "approve tool"},
					{"n", "deny tool"},
					{"a", "always allow"},
					{"N", "never allow"},
					{"i", "inspect tool details"},
				},
			},
		}

	case HelpModeSearch:
		return []KeyBindingCategory{
			{
				Name: "Search",
				Bindings: []KeyBinding{
					{"type", "enter search query"},
					{"enter", "execute search"},
					{"esc", "exit search"},
					{"↑/↓", "navigate results"},
				},
			},
		}

	case HelpModeQuickActions:
		return []KeyBindingCategory{
			{
				Name: "Quick Actions",
				Bindings: []KeyBinding{
					{"type", "filter actions"},
					{"↑/↓", "navigate actions"},
					{"enter", "execute action"},
					{"esc", "cancel"},
				},
			},
		}

	default:
		return []KeyBindingCategory{}
	}
}

// SetMode updates the help mode and key bindings
func (h *Help) SetMode(mode HelpMode) {
	h.mode = mode
	h.categories = getKeyBindingsForMode(mode)
}

// SetWidth sets the help width
func (h *Help) SetWidth(width int) {
	h.width = width
	h.help.Width = width
}

// ToggleExpanded toggles between short and full help
func (h *Help) ToggleExpanded() {
	h.expanded = !h.expanded
}

// SetExpanded sets the expanded state
func (h *Help) SetExpanded(expanded bool) {
	h.expanded = expanded
}

// IsExpanded returns whether help is expanded
func (h *Help) IsExpanded() bool {
	return h.expanded
}

// Update handles help updates
func (h *Help) Update(_ tea.Msg) tea.Cmd {
	return nil
}

// View renders the help
func (h *Help) View() string {
	if h.expanded {
		return h.ViewExpanded()
	}
	return h.ViewCompact()
}

// ViewCompact renders a compact single-line help
func (h *Help) ViewCompact() string {
	// Get first few bindings from each category
	var bindings []string
	bindingCount := 0
	maxBindings := 4

	for _, category := range h.categories {
		for _, binding := range category.Bindings {
			if bindingCount >= maxBindings {
				break
			}
			keyStyle := h.theme.HelpKey.Render(binding.Key)
			descStyle := h.theme.HelpDesc.Render(binding.Description)
			bindings = append(bindings, keyStyle+" "+descStyle)
			bindingCount++
		}
		if bindingCount >= maxBindings {
			break
		}
	}

	// Add help toggle hint
	if bindingCount > 0 {
		moreStyle := h.theme.Muted.Render("• ? for more")
		bindings = append(bindings, moreStyle)
	}

	return strings.Join(bindings, " • ")
}

// ViewExpanded renders full categorized help
func (h *Help) ViewExpanded() string {
	// Pre-allocate sections slice with estimated capacity
	sections := make([]string, 0, len(h.categories)*5+3)

	// Render title
	titleStyle := h.theme.Title.Width(h.width).Align(lipgloss.Center)
	sections = append(sections, titleStyle.Render("Help"))
	sections = append(sections, "")

	// Render each category
	for _, category := range h.categories {
		// Category name
		categoryStyle := h.theme.Subtitle.Render(category.Name)
		sections = append(sections, categoryStyle)

		// Bindings in category
		for _, binding := range category.Bindings {
			keyStyle := h.theme.HelpKey.Render("  " + binding.Key)
			descStyle := h.theme.HelpDesc.Render(" • " + binding.Description)
			sections = append(sections, keyStyle+descStyle)
		}

		sections = append(sections, "")
	}

	// Add toggle hint
	hintStyle := h.theme.Muted.Render("Press ? to hide help")
	sections = append(sections, hintStyle)

	// Apply border
	content := strings.Join(sections, "\n")
	bordered := h.theme.HelpPanel.Width(h.width - 4).Render(content)

	return bordered
}

// ViewAsOverlay renders help as a centered overlay
func (h *Help) ViewAsOverlay(screenWidth, screenHeight int) string {
	content := h.ViewExpanded()

	// Center the help panel
	contentWidth := h.width
	contentLines := strings.Count(content, "\n") + 1

	horizontalPadding := (screenWidth - contentWidth) / 2
	verticalPadding := (screenHeight - contentLines) / 2

	if horizontalPadding < 0 {
		horizontalPadding = 0
	}
	if verticalPadding < 0 {
		verticalPadding = 0
	}

	// Add padding
	paddedContent := lipgloss.NewStyle().
		Width(screenWidth).
		Height(screenHeight).
		Padding(verticalPadding, 0, 0, horizontalPadding).
		Render(content)

	return paddedContent
}

// KeyBindingsToKeyMap converts KeyBinding slice to bubbles key.Binding slice
// for use with bubbles.Help.View()
func KeyBindingsToKeyMap(bindings []KeyBinding) []key.Binding {
	keyMap := make([]key.Binding, len(bindings))
	for i, binding := range bindings {
		keyMap[i] = key.NewBinding(
			key.WithKeys(binding.Key),
			key.WithHelp(binding.Key, binding.Description),
		)
	}
	return keyMap
}

// DefaultChatHelp returns default help for chat mode
func DefaultChatHelp(width int) *Help {
	return NewHelp(HelpModeChat, width)
}

// DefaultHistoryHelp returns default help for history mode
func DefaultHistoryHelp(width int) *Help {
	return NewHelp(HelpModeHistory, width)
}

// DefaultToolsHelp returns default help for tools mode
func DefaultToolsHelp(width int) *Help {
	return NewHelp(HelpModeTools, width)
}
