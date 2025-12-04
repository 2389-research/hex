// Package browser provides conversation browsing and search interfaces for the TUI.
// ABOUTME: Conversation browser with fuzzy search and filtering capabilities
// ABOUTME: Browse past conversations, search content, and load into current session
package browser

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/2389-research/hex/internal/services"
	"github.com/2389-research/hex/internal/ui/theme"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"
)

// SortMode determines how conversations are sorted
type SortMode int

// Sort mode constants
const (
	SortByDate SortMode = iota
	SortByFavorite
	SortByTitle
)

// ConversationBrowser manages the conversation browsing interface
type ConversationBrowser struct {
	db             *sql.DB
	convSvc        services.ConversationService
	theme          *theme.Theme
	conversations  []*services.Conversation
	filteredItems  []list.Item
	list           list.Model
	searchQuery    string
	sortMode       SortMode
	width          int
	height         int
	selectedConv   *services.Conversation
	previewContent string
	err            error
}

// conversationItem wraps a conversation for list display
type conversationItem struct {
	conv  *services.Conversation
	theme *theme.Theme
}

// FilterValue implements list.Item interface
func (ci conversationItem) FilterValue() string {
	return ci.conv.Title
}

// Title returns the formatted title
func (ci conversationItem) Title() string {
	title := ci.conv.Title
	if title == "" {
		title = "Untitled Conversation"
	}

	// Add favorite indicator
	if ci.conv.IsFavorite {
		title = "⭐ " + title
	}

	style := lipgloss.NewStyle().Foreground(ci.theme.Colors.Purple).Bold(true)
	return style.Render(title)
}

// Description returns the formatted description
func (ci conversationItem) Description() string {
	// Format: Created: date | Updated: date
	created := ci.conv.CreatedAt.Format("Jan 2, 2006")
	updated := ci.conv.UpdatedAt.Format("Jan 2, 15:04")

	dateStyle := lipgloss.NewStyle().Foreground(ci.theme.Colors.Comment)

	return fmt.Sprintf("%s │ %s",
		dateStyle.Render("Created: "+created),
		dateStyle.Render("Updated: "+updated),
	)
}

// NewConversationBrowser creates a new conversation browser
func NewConversationBrowser(db *sql.DB, convSvc services.ConversationService, t *theme.Theme) *ConversationBrowser {
	// Create list delegate with Dracula styling
	delegate := list.NewDefaultDelegate()

	// Style the list with Dracula theme
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().
		Foreground(t.Colors.Background).
		Background(t.Colors.Purple).
		Bold(true)
	delegate.Styles.SelectedDesc = lipgloss.NewStyle().
		Foreground(t.Colors.Background).
		Background(t.Colors.Purple)
	delegate.Styles.NormalTitle = lipgloss.NewStyle().
		Foreground(t.Colors.Purple).
		Bold(true)
	delegate.Styles.NormalDesc = lipgloss.NewStyle().
		Foreground(t.Colors.Comment)

	l := list.New([]list.Item{}, delegate, 0, 0)
	l.Title = "Conversation Browser"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)

	// Style the list with Dracula colors
	l.Styles.Title = lipgloss.NewStyle().
		Foreground(t.Colors.Pink).
		Bold(true).
		Padding(0, 1)
	l.Styles.TitleBar = lipgloss.NewStyle().
		Background(t.Colors.CurrentLine)
	l.Styles.StatusBar = lipgloss.NewStyle().
		Foreground(t.Colors.Comment).
		Background(t.Colors.CurrentLine)

	return &ConversationBrowser{
		db:            db,
		convSvc:       convSvc,
		theme:         t,
		list:          l,
		sortMode:      SortByDate,
		conversations: []*services.Conversation{},
		filteredItems: []list.Item{},
	}
}

// Init initializes the browser
func (cb *ConversationBrowser) Init() tea.Cmd {
	return cb.loadConversations
}

// Update handles messages
func (cb *ConversationBrowser) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		cb.width = msg.Width
		cb.height = msg.Height

		// Reserve space for preview pane (40% of width)
		listWidth := int(float64(msg.Width) * 0.6)
		cb.list.SetSize(listWidth, msg.Height-4)

		return cb, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return cb, tea.Quit

		case "enter":
			// Load selected conversation
			if i, ok := cb.list.SelectedItem().(conversationItem); ok {
				cb.selectedConv = i.conv
				return cb, cb.loadConversationContent
			}

		case "f":
			// Toggle favorite
			if i, ok := cb.list.SelectedItem().(conversationItem); ok {
				return cb, cb.toggleFavorite(i.conv.ID, !i.conv.IsFavorite)
			}

		case "d":
			// Delete conversation
			if i, ok := cb.list.SelectedItem().(conversationItem); ok {
				return cb, cb.deleteConversation(i.conv.ID)
			}

		case "s":
			// Cycle sort mode
			cb.sortMode = (cb.sortMode + 1) % 3
			return cb, cb.loadConversations

		case "r":
			// Refresh list
			return cb, cb.loadConversations
		}

	case conversationsLoadedMsg:
		cb.conversations = msg.conversations
		cb.err = msg.err
		cb.updateFilteredItems()
		return cb, nil

	case conversationContentMsg:
		cb.previewContent = msg.content
		return cb, nil

	case errorMsg:
		cb.err = msg.err
		return cb, nil
	}

	var cmd tea.Cmd
	cb.list, cmd = cb.list.Update(msg)

	// Update preview when selection changes
	if i, ok := cb.list.SelectedItem().(conversationItem); ok {
		if cb.selectedConv == nil || cb.selectedConv.ID != i.conv.ID {
			cb.selectedConv = i.conv
			return cb, tea.Batch(cmd, cb.loadConversationContent)
		}
	}

	return cb, cmd
}

// View renders the browser
func (cb *ConversationBrowser) View() string {
	if cb.width == 0 {
		return "Loading..."
	}

	// Calculate dimensions
	listWidth := int(float64(cb.width) * 0.6)
	previewWidth := cb.width - listWidth - 4

	// Render list
	listView := cb.list.View()

	// Render preview pane
	previewView := cb.renderPreview(previewWidth, cb.height-4)

	// Combine horizontally
	combined := lipgloss.JoinHorizontal(
		lipgloss.Top,
		listView,
		lipgloss.NewStyle().
			Width(2).
			Height(cb.height-4).
			Foreground(cb.theme.Colors.Comment).
			Render("│\n│"),
		previewView,
	)

	// Add help footer
	helpStyle := lipgloss.NewStyle().
		Foreground(cb.theme.Colors.Comment).
		Padding(1, 2)

	help := helpStyle.Render(
		"[↑/↓] Navigate • [Enter] Load • [f] Favorite • [d] Delete • [s] Sort • [r] Refresh • [q] Quit",
	)

	return lipgloss.JoinVertical(lipgloss.Left, combined, help)
}

// renderPreview renders the conversation preview pane
func (cb *ConversationBrowser) renderPreview(width, height int) string {
	if cb.selectedConv == nil {
		emptyStyle := lipgloss.NewStyle().
			Width(width).
			Height(height).
			Foreground(cb.theme.Colors.Comment).
			Align(lipgloss.Center).
			AlignVertical(lipgloss.Center)
		return emptyStyle.Render("Select a conversation to preview")
	}

	// Build preview content
	var preview strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().
		Foreground(cb.theme.Colors.Purple).
		Bold(true).
		Width(width)
	preview.WriteString(titleStyle.Render(cb.selectedConv.Title))
	preview.WriteString("\n\n")

	// Metadata
	metaStyle := lipgloss.NewStyle().
		Foreground(cb.theme.Colors.Comment).
		Width(width)

	preview.WriteString(metaStyle.Render(fmt.Sprintf("ID: %s", cb.selectedConv.ID)))
	preview.WriteString("\n")
	preview.WriteString(metaStyle.Render(fmt.Sprintf("Created: %s", cb.selectedConv.CreatedAt.Format(time.RFC1123))))
	preview.WriteString("\n")
	preview.WriteString(metaStyle.Render(fmt.Sprintf("Updated: %s", cb.selectedConv.UpdatedAt.Format(time.RFC1123))))
	preview.WriteString("\n")
	preview.WriteString(metaStyle.Render(fmt.Sprintf("Favorite: %v", cb.selectedConv.IsFavorite)))
	preview.WriteString("\n")
	preview.WriteString(metaStyle.Render(fmt.Sprintf("Tokens: %d prompt / %d completion", cb.selectedConv.PromptTokens, cb.selectedConv.CompletionTokens)))
	preview.WriteString("\n")
	preview.WriteString(metaStyle.Render(fmt.Sprintf("Cost: $%.4f", cb.selectedConv.TotalCost)))
	preview.WriteString("\n\n")

	// Content preview
	if cb.previewContent != "" {
		contentStyle := lipgloss.NewStyle().
			Foreground(cb.theme.Colors.Foreground).
			Width(width).
			MaxHeight(height - 15)

		preview.WriteString(titleStyle.Render("Recent Messages:"))
		preview.WriteString("\n")
		preview.WriteString(contentStyle.Render(cb.previewContent))
	}

	// Wrap in border
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(cb.theme.Colors.Purple).
		Width(width).
		Height(height).
		Padding(1)

	return borderStyle.Render(preview.String())
}

// updateFilteredItems updates the filtered conversation list
func (cb *ConversationBrowser) updateFilteredItems() {
	items := make([]list.Item, 0, len(cb.conversations))

	// Apply sorting
	sortedConvs := cb.sortConversations(cb.conversations)

	// Apply fuzzy search if query is active
	if cb.searchQuery != "" {
		matches := cb.fuzzySearch(sortedConvs, cb.searchQuery)
		for _, conv := range matches {
			items = append(items, conversationItem{conv: conv, theme: cb.theme})
		}
	} else {
		for _, conv := range sortedConvs {
			items = append(items, conversationItem{conv: conv, theme: cb.theme})
		}
	}

	cb.filteredItems = items
	cb.list.SetItems(items)
}

// sortConversations sorts conversations based on the current sort mode
func (cb *ConversationBrowser) sortConversations(convs []*services.Conversation) []*services.Conversation {
	// Make a copy to avoid modifying the original
	sorted := make([]*services.Conversation, len(convs))
	copy(sorted, convs)

	// Sorting is already handled by the database query (by date)
	// We can add custom sorting here if needed
	switch cb.sortMode {
	case SortByFavorite:
		// Move favorites to the top
		var favorites, others []*services.Conversation
		for _, c := range sorted {
			if c.IsFavorite {
				favorites = append(favorites, c)
			} else {
				others = append(others, c)
			}
		}
		sorted = append(favorites, others...)
	case SortByTitle:
		// Sort alphabetically by title (simple bubble sort for small lists)
		for i := 0; i < len(sorted)-1; i++ {
			for j := 0; j < len(sorted)-i-1; j++ {
				if strings.ToLower(sorted[j].Title) > strings.ToLower(sorted[j+1].Title) {
					sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
				}
			}
		}
	}

	return sorted
}

// fuzzySearch performs fuzzy search on conversations
func (cb *ConversationBrowser) fuzzySearch(convs []*services.Conversation, query string) []*services.Conversation {
	// Build searchable strings
	searchTargets := make([]string, len(convs))
	for i, conv := range convs {
		searchTargets[i] = conv.Title
	}

	// Perform fuzzy search
	matches := fuzzy.Find(query, searchTargets)

	// Extract matched conversations
	result := make([]*services.Conversation, 0, len(matches))
	for _, match := range matches {
		result = append(result, convs[match.Index])
	}

	return result
}

// Messages

type conversationsLoadedMsg struct {
	conversations []*services.Conversation
	err           error
}

type conversationContentMsg struct {
	content string
}

type errorMsg struct {
	err error
}

// Commands

func (cb *ConversationBrowser) loadConversations() tea.Msg {
	convs, err := cb.convSvc.List(context.Background())
	return conversationsLoadedMsg{conversations: convs, err: err}
}

func (cb *ConversationBrowser) loadConversationContent() tea.Msg {
	if cb.selectedConv == nil {
		return conversationContentMsg{content: ""}
	}

	// Load recent messages (simplified - you may want to implement full message loading)
	// For now, just return a placeholder
	return conversationContentMsg{
		content: fmt.Sprintf("Messages for conversation %s\n(Full message loading to be implemented)", cb.selectedConv.ID),
	}
}

func (cb *ConversationBrowser) toggleFavorite(id string, isFavorite bool) tea.Cmd {
	return func() tea.Msg {
		// Get conversation
		conv, err := cb.convSvc.Get(context.Background(), id)
		if err != nil {
			return errorMsg{err: err}
		}

		// Update favorite status
		conv.IsFavorite = isFavorite
		err = cb.convSvc.Update(context.Background(), conv)
		if err != nil {
			return errorMsg{err: err}
		}

		// Reload conversations
		convs, err := cb.convSvc.List(context.Background())
		return conversationsLoadedMsg{conversations: convs, err: err}
	}
}

func (cb *ConversationBrowser) deleteConversation(id string) tea.Cmd {
	return func() tea.Msg {
		err := cb.convSvc.Delete(context.Background(), id)
		if err != nil {
			return errorMsg{err: err}
		}
		// Reload conversations
		convs, err := cb.convSvc.List(context.Background())
		return conversationsLoadedMsg{conversations: convs, err: err}
	}
}

// GetSelectedConversation returns the currently selected conversation
func (cb *ConversationBrowser) GetSelectedConversation() *services.Conversation {
	return cb.selectedConv
}
