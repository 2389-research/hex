package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ToolLogOverlay displays tool output in a fullscreen scrollable view
type ToolLogOverlay struct {
	lines    *[]string      // Reference to Model's tool log lines
	viewport viewport.Model // Embedded viewport for scrolling
	width    int
	height   int
}

// NewToolLogOverlay creates a new tool log overlay
func NewToolLogOverlay(lines *[]string) *ToolLogOverlay {
	return &ToolLogOverlay{
		lines:    lines,
		viewport: viewport.New(0, 0), // Initialized in OnPush
	}
}

// IsFullscreen returns true (this is a fullscreen overlay)
func (o *ToolLogOverlay) IsFullscreen() bool {
	return true
}

// GetDesiredHeight returns -1 (fullscreen wants all available height)
func (o *ToolLogOverlay) GetDesiredHeight() int {
	return -1
}

// GetHeader returns the overlay header
func (o *ToolLogOverlay) GetHeader() string {
	return "Tool Output Log"
}

const toolLogMaxLines = 10000

// GetContent returns the current tool log lines
func (o *ToolLogOverlay) GetContent() string {
	if o.lines == nil || len(*o.lines) == 0 {
		return "No tool output in current chunk"
	}

	// Apply 10k line limit
	lines := *o.lines
	if len(lines) > toolLogMaxLines {
		lines = lines[len(lines)-toolLogMaxLines:]
	}

	return strings.Join(lines, "\n")
}

// GetFooter returns the overlay footer with line count
func (o *ToolLogOverlay) GetFooter() string {
	totalLines := 0
	if o.lines != nil {
		totalLines = len(*o.lines)
	}
	if totalLines > toolLogMaxLines {
		return fmt.Sprintf("Showing last 10,000 of %d lines • Esc to close", totalLines)
	}
	return fmt.Sprintf("%d lines • Esc to close", totalLines)
}

// OnPush initializes the viewport with dimensions
func (o *ToolLogOverlay) OnPush(width, height int) {
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

	// Initialize viewport (leave space for header and footer)
	o.viewport = viewport.New(vw, vh)
	o.viewport.SetContent(o.GetContent())

	// Auto-scroll to bottom
	o.viewport.GotoBottom()
}

// OnPop cleans up
func (o *ToolLogOverlay) OnPop() {
	// Nothing to clean up
}

// SetHeight updates the viewport height
func (o *ToolLogOverlay) SetHeight(height int) {
	o.height = height
	vh := height - 6
	if vh < 1 {
		vh = 1
	}
	o.viewport.Height = vh
	o.viewport.SetContent(o.GetContent())
}

// Update handles viewport messages
func (o *ToolLogOverlay) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	o.viewport, cmd = o.viewport.Update(msg)

	// Update viewport dimensions and content on window size changes
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
		o.viewport.SetContent(o.GetContent())
	}

	return cmd
}

// HandleKey processes keyboard input
func (o *ToolLogOverlay) HandleKey(msg tea.KeyMsg) (bool, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc, tea.KeyCtrlC:
		// Pop will be handled by caller
		return true, nil

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

// Render returns the complete fullscreen view
func (o *ToolLogOverlay) Render(width, height int) string {
	var b strings.Builder

	// Update content if lines changed
	o.viewport.SetContent(o.GetContent())

	// Header with border
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("cyan"))

	closeHint := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render("Ctrl+O or Esc to close")

	header := headerStyle.Render(o.GetHeader())
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

	// Viewport content
	b.WriteString(o.viewport.View())
	b.WriteString("\n\n")

	// Footer with border
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	b.WriteString(footerStyle.Render(o.GetFooter()))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("┗%s┛", strings.Repeat("━", width-2)))

	return b.String()
}

// Cancel dismisses the tool log overlay (no cleanup needed)
func (o *ToolLogOverlay) Cancel() tea.Cmd {
	return nil
}
