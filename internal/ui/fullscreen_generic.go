package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FullscreenContentProvider defines the interface for providing content to a generic fullscreen overlay.
// Implementations should be minimal - just return strings. All viewport and lifecycle logic is handled by GenericFullscreenOverlay.
//
// Required methods: Header, Content, ToggleKeys (3 methods)
// Optional: implement FooterProvider interface for custom footer text
type FullscreenContentProvider interface {
	// Header returns the header text (e.g., "Help & Keyboard Shortcuts")
	Header() string

	// Content returns the scrollable content body
	Content() string

	// ToggleKeys returns the list of keys that should toggle this overlay closed (e.g., KeyCtrlH)
	// The close hint in the header AND default footer are auto-generated from these keys.
	// Escape and Ctrl+C always close the overlay regardless of this list.
	ToggleKeys() []tea.KeyType
}

// FooterProvider is an optional interface for custom footer text.
// If not implemented, the footer defaults to the close hint (e.g., "Ctrl+H or Esc to close").
type FooterProvider interface {
	Footer() string
}

// keyTypeName converts a tea.KeyType to a human-readable string
func keyTypeName(k tea.KeyType) string {
	switch k {
	case tea.KeyEsc:
		return "Esc"
	case tea.KeyCtrlC:
		return "Ctrl+C"
	case tea.KeyCtrlH:
		return "Ctrl+H"
	case tea.KeyCtrlO:
		return "Ctrl+O"
	case tea.KeyCtrlR:
		return "Ctrl+R"
	case tea.KeyCtrlL:
		return "Ctrl+L"
	case tea.KeyEnter:
		return "Enter"
	case tea.KeyTab:
		return "Tab"
	case tea.KeySpace:
		return "Space"
	default:
		return fmt.Sprintf("Key(%d)", k)
	}
}

// buildCloseHint generates a close hint string from toggle keys
func buildCloseHint(keys []tea.KeyType) string {
	if len(keys) == 0 {
		return "Esc to close"
	}

	names := make([]string, 0, len(keys)+1)
	hasEsc := false
	for _, k := range keys {
		if k == tea.KeyEsc {
			hasEsc = true
		}
		names = append(names, keyTypeName(k))
	}
	// Always include Esc if not already present
	if !hasEsc {
		names = append(names, "Esc")
	}

	return strings.Join(names, " or ") + " to close"
}

// GenericFullscreenOverlay is a reusable fullscreen overlay that delegates content generation
// to a FullscreenContentProvider. It handles ALL viewport logic, rendering, and input handling.
type GenericFullscreenOverlay struct {
	provider FullscreenContentProvider
	viewport viewport.Model
	width    int
	height   int
}

// NewGenericFullscreenOverlay creates a new generic fullscreen overlay with the given content provider
func NewGenericFullscreenOverlay(provider FullscreenContentProvider) *GenericFullscreenOverlay {
	return &GenericFullscreenOverlay{
		provider: provider,
		viewport: viewport.New(0, 0),
	}
}

// IsFullscreen returns true
func (o *GenericFullscreenOverlay) IsFullscreen() bool {
	return true
}

// GetDesiredHeight returns -1 (fullscreen)
func (o *GenericFullscreenOverlay) GetDesiredHeight() int {
	return -1
}

// GetHeader delegates to provider
func (o *GenericFullscreenOverlay) GetHeader() string {
	return o.provider.Header()
}

// GetContent delegates to provider
func (o *GenericFullscreenOverlay) GetContent() string {
	return o.provider.Content()
}

// GetFooter returns custom footer if provider implements FooterProvider,
// otherwise returns the auto-generated close hint
func (o *GenericFullscreenOverlay) GetFooter() string {
	if fp, ok := o.provider.(FooterProvider); ok {
		return fp.Footer()
	}
	return buildCloseHint(o.provider.ToggleKeys())
}

// OnPush initializes viewport with content
func (o *GenericFullscreenOverlay) OnPush(width, height int) {
	o.width = width
	o.height = height
	// Guard against negative dimensions on small terminals
	vw := width - 4
	vh := height - 6
	if vw < 1 {
		vw = 1
	}
	if vh < 1 {
		vh = 1
	}
	o.viewport = viewport.New(vw, vh)
	o.viewport.SetContent(o.provider.Content())
}

// OnPop cleans up
func (o *GenericFullscreenOverlay) OnPop() {}

// SetHeight updates viewport height
func (o *GenericFullscreenOverlay) SetHeight(height int) {
	o.height = height
	vh := height - 6
	if vh < 1 {
		vh = 1
	}
	o.viewport.Height = vh
}

// Update handles messages (viewport updates and window resize)
func (o *GenericFullscreenOverlay) Update(msg tea.Msg) tea.Cmd {
	// Handle window resize
	if wsm, ok := msg.(tea.WindowSizeMsg); ok {
		o.width = wsm.Width
		o.height = wsm.Height
		vw := wsm.Width - 4
		vh := wsm.Height - 6
		if vw < 1 {
			vw = 1
		}
		if vh < 1 {
			vh = 1
		}
		o.viewport.Width = vw
		o.viewport.Height = vh
	}

	var cmd tea.Cmd
	o.viewport, cmd = o.viewport.Update(msg)
	return cmd
}

// HandleKey processes input
func (o *GenericFullscreenOverlay) HandleKey(msg tea.KeyMsg) (bool, tea.Cmd) {
	// Check if this key should close the overlay
	toggleKeys := o.provider.ToggleKeys()
	for _, key := range toggleKeys {
		if msg.Type == key {
			return true, nil // Pop handled by caller
		}
	}

	// Always allow Escape and Ctrl+C to close
	if msg.Type == tea.KeyEsc || msg.Type == tea.KeyCtrlC {
		return true, nil
	}

	// For navigation keys, update viewport and capture the event
	switch msg.Type {
	case tea.KeyUp, tea.KeyDown:
		cmd := o.Update(msg)
		return true, cmd
	default:
		// Let viewport handle other keys (PageUp/PageDown, Home/End, etc.)
		// and capture them so they don't leak through
		cmd := o.Update(msg)
		return true, cmd
	}
}

// Render returns the complete view with header, viewport content, and footer
func (o *GenericFullscreenOverlay) Render(width, height int) string {
	var b strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("cyan"))
	closeHint := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render(buildCloseHint(o.provider.ToggleKeys()))

	header := headerStyle.Render(o.provider.Header())
	// Use lipgloss.Width for accurate visual width calculation
	headerWidth := lipgloss.Width(header)
	closeHintWidth := lipgloss.Width(closeHint)
	padding := max(0, width-headerWidth-closeHintWidth-8)
	headerLine := fmt.Sprintf("┏━━ %s %s %s ┓",
		header,
		strings.Repeat("━", padding),
		closeHint)
	b.WriteString(headerLine)
	b.WriteString("\n\n")

	// Content
	b.WriteString(o.viewport.View())
	b.WriteString("\n\n")

	// Footer
	footerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	b.WriteString(footerStyle.Render(o.GetFooter()))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("┗%s┛", strings.Repeat("━", max(0, width-2))))

	return b.String()
}

// Cancel dismisses the overlay (no cleanup needed by default)
func (o *GenericFullscreenOverlay) Cancel() tea.Cmd {
	return nil
}
