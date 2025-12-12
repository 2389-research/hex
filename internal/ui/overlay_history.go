package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// HistoryOverlay displays conversation history
type HistoryOverlay struct {
	messages *[]Message
	viewport viewport.Model
	width    int
	height   int
}

// NewHistoryOverlay creates a new history overlay
func NewHistoryOverlay(messages *[]Message) *HistoryOverlay {
	return &HistoryOverlay{
		messages: messages,
		viewport: viewport.New(0, 0),
	}
}

// IsFullscreen returns true
func (o *HistoryOverlay) IsFullscreen() bool {
	return true
}

// GetDesiredHeight returns -1 (fullscreen)
func (o *HistoryOverlay) GetDesiredHeight() int {
	return -1
}

// GetHeader returns the header
func (o *HistoryOverlay) GetHeader() string {
	return fmt.Sprintf("Conversation History (%d messages)", len(*o.messages))
}

// GetContent returns formatted message history
func (o *HistoryOverlay) GetContent() string {
	if len(*o.messages) == 0 {
		return "No messages in conversation"
	}

	// Apply 1000 message limit
	messages := *o.messages
	if len(messages) > 1000 {
		messages = messages[len(messages)-1000:]
	}

	var b strings.Builder
	userStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("green"))
	assistantStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("cyan"))
	timeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	for i, msg := range messages {
		// Timestamp
		timestamp := msg.Timestamp.Format("15:04:05")
		b.WriteString(timeStyle.Render(timestamp))
		b.WriteString(" ")

		// Role
		if msg.Role == "user" {
			b.WriteString(userStyle.Render("[YOU]"))
		} else {
			b.WriteString(assistantStyle.Render("[ASSISTANT]"))
		}
		b.WriteString("\n")

		// Content (truncate long messages)
		content := msg.Content
		if len(content) > 500 {
			content = content[:497] + "..."
		}
		b.WriteString(content)
		b.WriteString("\n")

		// Separator between messages
		if i < len(messages)-1 {
			b.WriteString(strings.Repeat("─", 80))
			b.WriteString("\n")
		}
	}

	return b.String()
}

// GetFooter returns the footer
func (o *HistoryOverlay) GetFooter() string {
	totalMessages := len(*o.messages)
	if totalMessages > 1000 {
		return fmt.Sprintf("Showing last 1,000 of %d messages • Escape to close", totalMessages)
	}
	return fmt.Sprintf("%d messages • Escape to close", totalMessages)
}

// OnPush initializes viewport
func (o *HistoryOverlay) OnPush(width, height int) {
	o.width = width
	o.height = height
	o.viewport = viewport.New(width-4, height-6)
	o.viewport.SetContent(o.GetContent())
	o.viewport.GotoBottom() // Start at most recent
}

// OnPop cleans up
func (o *HistoryOverlay) OnPop() {}

// SetHeight updates viewport height
func (o *HistoryOverlay) SetHeight(height int) {
	o.height = height
	o.viewport.Height = height - 6
}

// Update handles messages
func (o *HistoryOverlay) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	o.viewport, cmd = o.viewport.Update(msg)

	// Update content on window size changes
	if _, ok := msg.(tea.WindowSizeMsg); ok {
		o.viewport.SetContent(o.GetContent())
	}

	return cmd
}

// HandleKey processes input
func (o *HistoryOverlay) HandleKey(msg tea.KeyMsg) (bool, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc, tea.KeyCtrlC, tea.KeyCtrlR:
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
func (o *HistoryOverlay) Render(width, height int) string {
	var b strings.Builder

	// Update content if messages changed
	o.viewport.SetContent(o.GetContent())

	// Header
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("cyan"))
	closeHint := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render("Ctrl+R or Esc to close")

	header := headerStyle.Render(o.GetHeader())
	headerLine := fmt.Sprintf("┏━━ %s %s %s ┓",
		header,
		strings.Repeat("━", max(0, width-len(o.GetHeader())-len("Ctrl+R or Esc to close")-12)),
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

// Cancel dismisses the history overlay (no cleanup needed)
func (o *HistoryOverlay) Cancel() tea.Cmd {
	return nil
}
