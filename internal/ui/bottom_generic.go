package ui

import (
	"strings"

	"github.com/2389-research/hex/internal/ui/theme"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// BottomContentProvider defines the interface for providing content to a generic bottom overlay.
// Implementations should be minimal - just return strings. All rendering and lifecycle logic is handled by GenericBottomOverlay.
//
// Required methods: Header, Content, Footer, DesiredHeight (4 methods)
type BottomContentProvider interface {
	// Header returns the header text (e.g., "Tool Approval Required")
	Header() string

	// Content returns the main content body
	Content() string

	// Footer returns the footer text (e.g., navigation hints)
	Footer() string

	// DesiredHeight returns the desired height for this overlay
	DesiredHeight() int
}

// BottomKeyHandler is an optional interface for custom key handling.
// If not implemented, default behavior (Escape and Ctrl+C close the overlay) is used.
type BottomKeyHandler interface {
	HandleKey(msg tea.KeyMsg) (handled bool, cmd tea.Cmd)
}

// BottomActivationHandler is an optional interface for activation/deactivation callbacks.
type BottomActivationHandler interface {
	OnActivate(width, height int)
	OnDeactivate()
}

// BottomCancelHandler is an optional interface for custom cancel behavior.
// If not implemented, Cancel() returns nil.
type BottomCancelHandler interface {
	Cancel() tea.Cmd
}

// GenericBottomOverlay is a reusable bottom overlay that delegates content generation
// to a BottomContentProvider. It handles ALL rendering logic and implements the Overlay interface.
type GenericBottomOverlay struct {
	provider BottomContentProvider
	theme    *theme.Theme
	width    int
	height   int
}

// NewGenericBottomOverlay creates a new generic bottom overlay with the given content provider
func NewGenericBottomOverlay(provider BottomContentProvider, thm *theme.Theme) *GenericBottomOverlay {
	return &GenericBottomOverlay{
		provider: provider,
		theme:    thm,
	}
}

// GetHeader delegates to provider
func (o *GenericBottomOverlay) GetHeader() string {
	return o.provider.Header()
}

// GetContent delegates to provider
func (o *GenericBottomOverlay) GetContent() string {
	return o.provider.Content()
}

// GetFooter delegates to provider
func (o *GenericBottomOverlay) GetFooter() string {
	return o.provider.Footer()
}

// GetDesiredHeight delegates to provider
func (o *GenericBottomOverlay) GetDesiredHeight() int {
	return o.provider.DesiredHeight()
}

// OnPush initializes the overlay
func (o *GenericBottomOverlay) OnPush(width, height int) {
	o.width = width
	o.height = height
	// Call provider's OnActivate if it implements the interface
	if handler, ok := o.provider.(BottomActivationHandler); ok {
		handler.OnActivate(width, height)
	}
}

// OnPop cleans up the overlay
func (o *GenericBottomOverlay) OnPop() {
	// Call provider's OnDeactivate if it implements the interface
	if handler, ok := o.provider.(BottomActivationHandler); ok {
		handler.OnDeactivate()
	}
}

// HandleKey processes input
func (o *GenericBottomOverlay) HandleKey(msg tea.KeyMsg) (bool, tea.Cmd) {
	// If provider implements custom key handling, use it
	if handler, ok := o.provider.(BottomKeyHandler); ok {
		return handler.HandleKey(msg)
	}

	// Default behavior: Escape and Ctrl+C close the overlay
	if msg.Type == tea.KeyEsc || msg.Type == tea.KeyCtrlC {
		return true, nil // Pop handled by caller
	}

	// Modal: capture all other input to prevent leakage
	return true, nil
}

// Render returns the complete overlay rendering with theme-aware styling
func (o *GenericBottomOverlay) Render(width, height int) string {
	var b strings.Builder

	// Use theme's AutocompleteDropdown style as base for bottom overlays
	// Calculate dropdown width with clamping
	dropdownWidth := width - 4
	if dropdownWidth < 1 {
		dropdownWidth = 1
	}
	// Only enforce minimum if terminal is wide enough
	const minDropdownWidth = 40
	if width >= minDropdownWidth+4 && dropdownWidth < minDropdownWidth {
		dropdownWidth = minDropdownWidth
	}
	boxStyle := o.theme.AutocompleteDropdown.Width(dropdownWidth)

	// Title style - use theme colors
	titleStyle := lipgloss.NewStyle().
		Foreground(o.theme.Colors.Purple).
		Bold(true)

	// Help style - use theme's AutocompleteHelp
	helpStyle := o.theme.AutocompleteHelp

	// Header
	if header := o.GetHeader(); header != "" {
		b.WriteString(titleStyle.Render(header))
		b.WriteString("\n")
	}

	// Content
	b.WriteString(o.GetContent())

	// Footer
	b.WriteString("\n")
	if footer := o.GetFooter(); footer != "" {
		b.WriteString(helpStyle.Render(footer))
	}

	return boxStyle.Render(b.String())
}

// Cancel dismisses the overlay
func (o *GenericBottomOverlay) Cancel() tea.Cmd {
	// Call provider's Cancel if it implements the interface
	if handler, ok := o.provider.(BottomCancelHandler); ok {
		return handler.Cancel()
	}
	return nil
}
