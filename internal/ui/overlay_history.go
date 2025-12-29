package ui

import (
	"fmt"
	"strings"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// HistoryContentProvider provides content for the conversation history overlay
type HistoryContentProvider struct {
	messages *[]Message
}

// Header returns the history overlay header
func (p *HistoryContentProvider) Header() string {
	return fmt.Sprintf("Conversation History (%d messages)", len(*p.messages))
}

// Content returns formatted message history
func (p *HistoryContentProvider) Content() string {
	if len(*p.messages) == 0 {
		return "No messages in conversation"
	}

	// Apply 1000 message limit
	messages := *p.messages
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

		// Content (truncate long messages with rune-aware truncation)
		content := msg.Content
		if utf8.RuneCountInString(content) > 500 {
			runes := []rune(content)
			content = string(runes[:497]) + "..."
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

// ToggleKeys returns the keys that toggle the history overlay
func (p *HistoryContentProvider) ToggleKeys() []tea.KeyType {
	return []tea.KeyType{tea.KeyCtrlR}
}

// Footer returns the custom footer with message count
func (p *HistoryContentProvider) Footer() string {
	totalMessages := len(*p.messages)
	if totalMessages > 1000 {
		return fmt.Sprintf("Showing last 1,000 of %d messages • Ctrl+R or Esc to close", totalMessages)
	}
	return fmt.Sprintf("%d messages • Ctrl+R or Esc to close", totalMessages)
}

// NewHistoryOverlay creates a new history overlay using the generic fullscreen wrapper
func NewHistoryOverlay(messages *[]Message) *GenericFullscreenOverlay {
	provider := &HistoryContentProvider{messages: messages}
	return NewGenericFullscreenOverlay(provider)
}
