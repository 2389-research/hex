// Package ui provides TUI components including the session picker for resuming conversations.
// ABOUTME: Shows list of recent conversations with title, time, and model info
package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/2389-research/hex/internal/storage"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SessionPicker is a TUI for selecting a conversation to resume
type SessionPicker struct {
	list     list.Model
	selected string
	quitting bool
}

// sessionItem wraps a storage.Conversation for the list
type sessionItem struct {
	conv *storage.Conversation
}

func (i sessionItem) FilterValue() string {
	return i.conv.Title
}

func (i sessionItem) Title() string {
	title := i.conv.Title
	if i.conv.IsFavorite {
		title = "★ " + title
	}
	return title
}

func (i sessionItem) Description() string {
	// Format: "Updated: 2h ago • Model: claude-sonnet-4"
	timeAgo := formatTimeAgo(i.conv.UpdatedAt)
	model := truncateModel(i.conv.Model)
	return fmt.Sprintf("Updated: %s • Model: %s • ID: %s", timeAgo, model, truncateID(i.conv.ID))
}

// NewSessionPicker creates a new session picker with conversations
func NewSessionPicker(conversations []*storage.Conversation) SessionPicker {
	// Convert conversations to list items
	items := make([]list.Item, len(conversations))
	for i, conv := range conversations {
		items[i] = sessionItem{conv: conv}
	}

	// Create list delegate
	delegate := list.NewDefaultDelegate()

	// Customize list styles
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("170")).
		Bold(true)

	delegate.Styles.SelectedDesc = lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	// Create list
	l := list.New(items, delegate, 0, 0)
	l.Title = "Resume Session"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(true)

	// Add custom key bindings
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(
				key.WithKeys("enter"),
				key.WithHelp("enter", "select"),
			),
			key.NewBinding(
				key.WithKeys("esc", "q"),
				key.WithHelp("esc/q", "cancel"),
			),
		}
	}

	return SessionPicker{
		list: l,
	}
}

// Init initializes the session picker (required by tea.Model interface)
func (m SessionPicker) Init() tea.Cmd {
	return nil
}

// Update handles messages for the session picker (required by tea.Model interface)
func (m SessionPicker) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			// Select current item
			if item, ok := m.list.SelectedItem().(sessionItem); ok {
				m.selected = item.conv.ID
				m.quitting = true
				return m, tea.Quit
			}

		case "q", "esc":
			// Cancel selection
			m.quitting = true
			return m, tea.Quit
		}
	}

	// Update list
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View renders the session picker (required by tea.Model interface)
func (m SessionPicker) View() string {
	if m.quitting {
		return ""
	}
	return "\n" + m.list.View()
}

// GetSelectedID returns the selected conversation ID (empty if cancelled)
func (m SessionPicker) GetSelectedID() string {
	return m.selected
}

// Helper functions

func formatTimeAgo(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		mins := int(diff.Minutes())
		if mins == 1 {
			return "1 min ago"
		}
		return fmt.Sprintf("%d mins ago", mins)
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case diff < 7*24*time.Hour:
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	default:
		return t.Format("Jan 2")
	}
}

func truncateModel(model string) string {
	// "claude-sonnet-4-5-20250929" -> "claude-sonnet-4-5"
	parts := strings.Split(model, "-")
	if len(parts) > 4 {
		return strings.Join(parts[:4], "-")
	}
	return model
}

func truncateID(id string) string {
	// "conv-1234567890" -> "conv-123..."
	// "550e8400-e29b-41d4-a716-446655440000" -> "550e8400..."
	if len(id) > 12 {
		return id[:8] + "..."
	}
	return id
}
