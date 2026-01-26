// ABOUTME: History browser content component for tux UI
// Implements tux content.Content interface to display and navigate session history.
// Supports keyboard navigation, session selection, deletion, and favorites.

package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/2389-research/tux/content"
	"github.com/2389-research/tux/theme"
)

// HistoryAction represents actions that can be performed on sessions.
type HistoryAction int

const (
	// ActionNone indicates no action.
	ActionNone HistoryAction = iota
	// ActionNewSession requests creation of a new session.
	ActionNewSession
)

// SessionSelectedMsg is sent when a session is selected.
type SessionSelectedMsg struct {
	Session *Session
}

// NewSessionRequestedMsg is sent when a new session is requested.
type NewSessionRequestedMsg struct{}

// SessionDeletedMsg is sent when a session has been deleted.
type SessionDeletedMsg struct {
	SessionID string
}

// SessionFavoriteToggledMsg is sent when a session's favorite status changes.
type SessionFavoriteToggledMsg struct {
	Session *Session
}

// HistoryContent displays session list for selection.
// Implements tux content.Content interface.
type HistoryContent struct {
	storage  *SessionStorage
	sessions []*Session
	filtered []*Session // Filtered sessions when searching
	cursor   int
	width    int
	height   int
	onSelect func(session *Session)
	theme    theme.Theme

	// Delete confirmation state
	deleteConfirm bool
	deleteTarget  int

	// Search state
	searchMode  bool
	searchQuery string

	// Styles derived from theme
	titleStyle        lipgloss.Style
	dateStyle         lipgloss.Style
	selectedStyle     lipgloss.Style
	cursorStyle       lipgloss.Style
	favoriteStyle     lipgloss.Style
	emptyStyle        lipgloss.Style
	deletePromptStyle lipgloss.Style
	searchStyle       lipgloss.Style
}

// NewHistoryContent creates a new HistoryContent instance.
func NewHistoryContent(storage *SessionStorage, th theme.Theme, onSelect func(*Session)) *HistoryContent {
	h := &HistoryContent{
		storage:  storage,
		sessions: nil,
		cursor:   0,
		onSelect: onSelect,
		theme:    th,
	}
	h.initStyles()
	return h
}

// initStyles initializes lipgloss styles from the theme.
func (h *HistoryContent) initStyles() {
	styles := h.theme.Styles()

	h.titleStyle = styles.Body.
		Foreground(h.theme.Foreground())

	h.dateStyle = styles.Muted

	h.selectedStyle = styles.ListItemSelected.
		Bold(true)

	h.cursorStyle = lipgloss.NewStyle().
		Foreground(h.theme.Primary())

	h.favoriteStyle = lipgloss.NewStyle().
		Foreground(h.theme.Warning())

	h.emptyStyle = styles.Muted.
		Italic(true)

	h.deletePromptStyle = lipgloss.NewStyle().
		Foreground(h.theme.Error()).
		Bold(true)

	h.searchStyle = lipgloss.NewStyle().
		Foreground(h.theme.Primary()).
		Bold(true)
}

// Init implements content.Content. Loads sessions from storage.
func (h *HistoryContent) Init() tea.Cmd {
	return h.loadSessionsCmd()
}

// sessionsLoadedMsg is sent when sessions have been loaded.
type sessionsLoadedMsg struct {
	sessions []*Session
	err      error
}

// loadSessionsCmd returns a command that loads sessions from storage.
func (h *HistoryContent) loadSessionsCmd() tea.Cmd {
	return func() tea.Msg {
		sessions, err := h.storage.List()
		return sessionsLoadedMsg{sessions: sessions, err: err}
	}
}

// Update implements content.Content. Handles keyboard input.
func (h *HistoryContent) Update(msg tea.Msg) (content.Content, tea.Cmd) {
	switch msg := msg.(type) {
	case sessionsLoadedMsg:
		if msg.err == nil {
			h.sessions = msg.sessions
			// Reset cursor if out of bounds
			if h.cursor >= len(h.sessions) {
				h.cursor = len(h.sessions) - 1
			}
			if h.cursor < 0 {
				h.cursor = 0
			}
		}
		return h, nil

	case tea.KeyMsg:
		// Handle search mode input
		if h.searchMode {
			switch msg.Type {
			case tea.KeyEsc:
				h.searchMode = false
				h.searchQuery = ""
				h.filtered = nil
				h.cursor = 0
				return h, nil
			case tea.KeyBackspace:
				if len(h.searchQuery) > 0 {
					h.searchQuery = h.searchQuery[:len(h.searchQuery)-1]
					h.applySearch()
				} else {
					// Exit search mode when backspacing empty query
					h.searchMode = false
					h.filtered = nil
					h.cursor = 0
				}
				return h, nil
			case tea.KeyEnter:
				// Select from filtered results
				return h, h.selectCurrent()
			case tea.KeyUp:
				return h, h.moveUp()
			case tea.KeyDown:
				return h, h.moveDown()
			default:
				// Add character to search query
				if len(msg.Runes) > 0 {
					h.searchQuery += string(msg.Runes)
					h.applySearch()
				}
				return h, nil
			}
		}

		// Cancel delete confirmation on any key except 'd'
		if h.deleteConfirm && msg.String() != "d" {
			h.deleteConfirm = false
			h.deleteTarget = -1
		}

		switch msg.Type {
		case tea.KeyUp:
			return h, h.moveUp()
		case tea.KeyDown:
			return h, h.moveDown()
		case tea.KeyEnter:
			return h, h.selectCurrent()
		case tea.KeyEsc:
			// Exit search mode if active (already handled above)
			return h, nil
		}

		switch msg.String() {
		case "k":
			return h, h.moveUp()
		case "j":
			return h, h.moveDown()
		case "g":
			h.cursor = 0
			return h, nil
		case "G":
			list := h.displayList()
			if len(list) > 0 {
				h.cursor = len(list) - 1
			}
			return h, nil
		case "d":
			return h, h.handleDelete()
		case "f":
			return h, h.toggleFavorite()
		case "n":
			return h, func() tea.Msg { return NewSessionRequestedMsg{} }
		case "r":
			return h, h.loadSessionsCmd()
		case "/":
			h.searchMode = true
			h.searchQuery = ""
			h.filtered = nil
			h.cursor = 0
			return h, nil
		}
	}

	return h, nil
}

// displayList returns the list to display (filtered if searching, otherwise all).
func (h *HistoryContent) displayList() []*Session {
	if h.searchMode && h.filtered != nil {
		return h.filtered
	}
	return h.sessions
}

// applySearch filters sessions by the search query.
func (h *HistoryContent) applySearch() {
	if h.searchQuery == "" {
		h.filtered = nil
		return
	}

	query := strings.ToLower(h.searchQuery)
	h.filtered = make([]*Session, 0)
	for _, session := range h.sessions {
		if strings.Contains(strings.ToLower(session.Title), query) {
			h.filtered = append(h.filtered, session)
		}
	}

	// Reset cursor if out of bounds
	if h.cursor >= len(h.filtered) {
		h.cursor = len(h.filtered) - 1
	}
	if h.cursor < 0 {
		h.cursor = 0
	}
}

// moveUp moves the cursor up.
func (h *HistoryContent) moveUp() tea.Cmd {
	if h.cursor > 0 {
		h.cursor--
	}
	return nil
}

// moveDown moves the cursor down.
func (h *HistoryContent) moveDown() tea.Cmd {
	list := h.displayList()
	if h.cursor < len(list)-1 {
		h.cursor++
	}
	return nil
}

// selectCurrent selects the currently highlighted session.
func (h *HistoryContent) selectCurrent() tea.Cmd {
	list := h.displayList()
	if len(list) == 0 || h.cursor < 0 || h.cursor >= len(list) {
		return nil
	}

	session := list[h.cursor]

	// Call the callback if provided
	if h.onSelect != nil {
		h.onSelect(session)
	}

	// Return a command that sends the selected message
	return func() tea.Msg {
		return SessionSelectedMsg{Session: session}
	}
}

// handleDelete handles the delete key press.
// First press sets confirmation, second press confirms.
func (h *HistoryContent) handleDelete() tea.Cmd {
	list := h.displayList()
	if len(list) == 0 || h.cursor < 0 || h.cursor >= len(list) {
		return nil
	}

	if h.deleteConfirm && h.deleteTarget == h.cursor {
		// Confirmed - perform deletion
		session := list[h.cursor]
		sessionID := session.ID

		// Delete from storage
		if err := h.storage.Delete(sessionID); err != nil {
			// Could return an error message, but for now just reset
			h.deleteConfirm = false
			h.deleteTarget = -1
			return nil
		}

		// Remove from main sessions list
		for i, s := range h.sessions {
			if s.ID == sessionID {
				h.sessions = append(h.sessions[:i], h.sessions[i+1:]...)
				break
			}
		}

		// Re-apply search filter if active
		if h.searchMode {
			h.applySearch()
		}

		// Adjust cursor if needed
		newList := h.displayList()
		if h.cursor >= len(newList) && h.cursor > 0 {
			h.cursor--
		}

		h.deleteConfirm = false
		h.deleteTarget = -1

		return func() tea.Msg {
			return SessionDeletedMsg{SessionID: sessionID}
		}
	}

	// First press - set confirmation
	h.deleteConfirm = true
	h.deleteTarget = h.cursor
	return nil
}

// toggleFavorite toggles the favorite status of the current session.
func (h *HistoryContent) toggleFavorite() tea.Cmd {
	list := h.displayList()
	if len(list) == 0 || h.cursor < 0 || h.cursor >= len(list) {
		return nil
	}

	session := list[h.cursor]
	session.Favorite = !session.Favorite

	// Save to storage
	if err := h.storage.Save(session); err != nil {
		// Revert on error
		session.Favorite = !session.Favorite
		return nil
	}

	return func() tea.Msg {
		return SessionFavoriteToggledMsg{Session: session}
	}
}

// View implements content.Content. Renders the session list.
func (h *HistoryContent) View() string {
	var b strings.Builder

	// Show search bar if in search mode
	if h.searchMode {
		b.WriteString(h.searchStyle.Render("Search: "))
		b.WriteString(h.searchQuery)
		b.WriteString("_\n\n")
	}

	list := h.displayList()

	if len(list) == 0 {
		if h.searchMode && h.searchQuery != "" {
			return b.String() + h.emptyStyle.Render("No matching sessions.")
		}
		return h.emptyStyle.Render("No sessions yet. Press 'n' to start a new session.")
	}

	for i, session := range list {
		// Check if we've exceeded available height
		if h.height > 0 && b.Len() > 0 {
			lines := strings.Count(b.String(), "\n")
			if lines >= h.height-2 { // Leave room for last item
				break
			}
		}

		line := h.renderSession(i, session)
		b.WriteString(line)

		if i < len(list)-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}

// renderSession renders a single session row.
func (h *HistoryContent) renderSession(index int, session *Session) string {
	isSelected := index == h.cursor

	// Build the cursor/favorite prefix
	cursor := "  "
	if isSelected {
		cursor = h.cursorStyle.Render("> ")
	}

	favorite := "  "
	if session.Favorite {
		favorite = h.favoriteStyle.Render("* ")
	}

	// Format title - truncate if needed
	title := session.Title
	maxTitleWidth := h.width - 30 // Leave room for date and prefix
	if maxTitleWidth < 20 {
		maxTitleWidth = 20
	}
	if len(title) > maxTitleWidth {
		title = title[:maxTitleWidth-3] + "..."
	}

	// Format date
	relativeTime := formatRelativeTime(session.UpdatedAt)

	// Apply styles
	var titleStyled, dateStyled string
	if isSelected {
		titleStyled = h.selectedStyle.Render(title)
		dateStyled = h.selectedStyle.Copy().Foreground(h.theme.Muted()).Render(relativeTime)
	} else {
		titleStyled = h.titleStyle.Render(title)
		dateStyled = h.dateStyle.Render(relativeTime)
	}

	// Check for delete confirmation
	if h.deleteConfirm && h.deleteTarget == index {
		return cursor + favorite + h.deletePromptStyle.Render(title+" [press 'd' again to delete]")
	}

	return cursor + favorite + titleStyled + "  " + dateStyled
}

// Value implements content.Content. Returns the current sessions.
func (h *HistoryContent) Value() any {
	return h.sessions
}

// SetSize implements content.Content. Updates available dimensions.
func (h *HistoryContent) SetSize(width, height int) {
	h.width = width
	h.height = height
}

// Refresh reloads sessions from storage.
func (h *HistoryContent) Refresh() error {
	sessions, err := h.storage.List()
	if err != nil {
		return err
	}
	h.sessions = sessions

	// Reset cursor if out of bounds
	if h.cursor >= len(h.sessions) {
		h.cursor = len(h.sessions) - 1
	}
	if h.cursor < 0 {
		h.cursor = 0
	}

	return nil
}

// SelectedSession returns the currently selected session, or nil if none.
func (h *HistoryContent) SelectedSession() *Session {
	list := h.displayList()
	if len(list) == 0 || h.cursor < 0 || h.cursor >= len(list) {
		return nil
	}
	return list[h.cursor]
}

// Sessions returns all loaded sessions.
func (h *HistoryContent) Sessions() []*Session {
	return h.sessions
}

// OnActivate is called when the History tab becomes active.
// Implements shell.TabContent interface for lifecycle hooks.
// Returns a command to refresh the session list.
func (h *HistoryContent) OnActivate() tea.Cmd {
	return h.loadSessionsCmd()
}

// OnDeactivate is called when the History tab becomes inactive.
// Implements shell.TabContent interface for lifecycle hooks.
func (h *HistoryContent) OnDeactivate() {
	// Clear delete confirmation state when leaving tab
	h.deleteConfirm = false
	h.deleteTarget = -1
}

// formatRelativeTime formats a time as a human-readable relative string.
func formatRelativeTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	// Handle future dates (shouldn't happen, but be safe)
	if diff < 0 {
		return "just now"
	}

	seconds := int(diff.Seconds())
	minutes := int(diff.Minutes())
	hours := int(diff.Hours())
	days := int(diff.Hours() / 24)

	switch {
	case seconds < 60:
		return "just now"
	case minutes == 1:
		return "1 minute ago"
	case minutes < 60:
		return fmt.Sprintf("%d minutes ago", minutes)
	case hours == 1:
		return "1 hour ago"
	case hours < 24:
		return fmt.Sprintf("%d hours ago", hours)
	case days == 1:
		return "yesterday"
	case days < 7:
		return fmt.Sprintf("%d days ago", days)
	case days < 14:
		return "1 week ago"
	case days < 30:
		return fmt.Sprintf("%d weeks ago", days/7)
	case days < 60:
		return "1 month ago"
	case days < 365:
		return fmt.Sprintf("%d months ago", days/30)
	case days < 730:
		return "1 year ago"
	default:
		return fmt.Sprintf("%d years ago", days/365)
	}
}
