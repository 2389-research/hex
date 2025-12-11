package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// HelpOverlay displays keyboard shortcuts and features
type HelpOverlay struct {
	viewport viewport.Model
	width    int
	height   int
}

// NewHelpOverlay creates a new help overlay
func NewHelpOverlay() *HelpOverlay {
	return &HelpOverlay{
		viewport: viewport.New(0, 0),
	}
}

// IsFullscreen returns true
func (o *HelpOverlay) IsFullscreen() bool {
	return true
}

// GetDesiredHeight returns -1 (fullscreen)
func (o *HelpOverlay) GetDesiredHeight() int {
	return -1
}

// GetHeader returns the header
func (o *HelpOverlay) GetHeader() string {
	return "Help & Keyboard Shortcuts"
}

// GetContent returns the help text
func (o *HelpOverlay) GetContent() string {
	return `# Keyboard Shortcuts

## Navigation
- **↑/↓**: Scroll viewport
- **PageUp/PageDown**: Page up/down
- **Ctrl+D/U**: Half page down/up
- **Home/End**: Go to top/bottom

## Overlays
- **Ctrl+O**: Toggle tool output log
- **Ctrl+H**: Toggle this help screen
- **Ctrl+R**: Open conversation history
- **Escape**: Close active overlay

## Input
- **Enter**: Send message
- **Shift+Enter**: New line in message
- **Ctrl+C**: Cancel stream or close overlay

## Tools
- **Enter**: Approve tool (when prompted)
- **Escape**: Deny tool (when prompted)

## Other
- **Ctrl+L**: Clear screen
- **Ctrl+Q**: Quit application

# Tips

- Use overlays to view detailed information without losing context
- All overlays are scrollable with arrow keys and PageUp/PageDown
- Tool output log shows the last 10,000 lines of tool execution
- Conversation history is limited to the last 1,000 messages
`
}

// GetFooter returns the footer
func (o *HelpOverlay) GetFooter() string {
	return "Press Escape or Ctrl+H to close"
}

// OnPush initializes viewport
func (o *HelpOverlay) OnPush(width, height int) {
	o.width = width
	o.height = height
	o.viewport = viewport.New(width-4, height-6)
	o.viewport.SetContent(o.GetContent())
}

// OnPop cleans up
func (o *HelpOverlay) OnPop() {}

// SetHeight updates viewport height
func (o *HelpOverlay) SetHeight(height int) {
	o.height = height
	o.viewport.Height = height - 6
}

// Update handles messages
func (o *HelpOverlay) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	o.viewport, cmd = o.viewport.Update(msg)
	return cmd
}

// HandleKey processes input
func (o *HelpOverlay) HandleKey(msg tea.KeyMsg) (bool, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc, tea.KeyCtrlC, tea.KeyCtrlH:
		return true, nil // Pop handled by caller

	case tea.KeyUp, tea.KeyDown:
		// Viewport navigation
		cmd := o.Update(msg)
		return true, cmd

	default:
		// Let viewport handle other keys (like PageUp/PageDown, Home/End, etc.)
		// and capture them so they don't leak through
		cmd := o.Update(msg)
		return true, cmd
	}
}

// Render returns the complete view
func (o *HelpOverlay) Render(width, height int) string {
	var b strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("cyan"))
	closeHint := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render("Ctrl+H or Esc to close")

	header := headerStyle.Render(o.GetHeader())
	headerLine := fmt.Sprintf("┏━━ %s %s %s ┓",
		header,
		strings.Repeat("━", max(0, width-len(o.GetHeader())-len("Ctrl+H or Esc to close")-12)),
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

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
