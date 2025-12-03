// Package components provides reusable Bubbles components with Dracula theme styling.
// ABOUTME: List component for browsing and selecting items with fuzzy search
// ABOUTME: Wraps bubbles.List with Dracula theme for history, skills, commands, and plugins
package components

import (
	"fmt"
	"io"
	"strings"

	"github.com/2389-research/hex/internal/ui/theme"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"
)

// ListType represents different types of lists
type ListType int

const (
	// ListTypeHistory for conversation history browser
	ListTypeHistory ListType = iota
	// ListTypeSkills for skills selector
	ListTypeSkills
	// ListTypeCommands for commands selector
	ListTypeCommands
	// ListTypePlugins for plugin browser
	ListTypePlugins
	// ListTypeMCPServers for MCP server browser
	ListTypeMCPServers
)

// ListItem is an item in the list
type ListItem interface {
	list.Item
	// GetSearchText returns the text to search against
	GetSearchText() string
	// GetTitle returns the item title
	GetTitle() string
	// GetDescription returns the item description
	GetDescription() string
	// GetMetadata returns additional metadata
	GetMetadata() string
}

// DefaultListItem is a default implementation of ListItem
type DefaultListItem struct {
	title       string
	description string
	metadata    string
}

// FilterValue implements list.Item
func (i DefaultListItem) FilterValue() string {
	return i.title
}

// GetSearchText implements ListItem
func (i DefaultListItem) GetSearchText() string {
	return i.title + " " + i.description
}

// GetTitle implements ListItem
func (i DefaultListItem) GetTitle() string {
	return i.title
}

// GetDescription implements ListItem
func (i DefaultListItem) GetDescription() string {
	return i.description
}

// GetMetadata implements ListItem
func (i DefaultListItem) GetMetadata() string {
	return i.metadata
}

// NewDefaultListItem creates a new default list item
func NewDefaultListItem(title, description, metadata string) DefaultListItem {
	return DefaultListItem{
		title:       title,
		description: description,
		metadata:    metadata,
	}
}

// List wraps bubbles.List with Dracula styling and fuzzy search
type List struct {
	list     list.Model
	theme    *theme.Theme
	listType ListType
	title    string
	items    []ListItem
	filtered []ListItem
}

// itemDelegate implements list.ItemDelegate for Dracula styling
type itemDelegate struct {
	theme *theme.Theme
}

func (d itemDelegate) Height() int {
	return 2
}

func (d itemDelegate) Spacing() int {
	return 1
}

func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd {
	return nil
}

func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(ListItem)
	if !ok {
		return
	}

	title := item.GetTitle()
	description := item.GetDescription()
	metadata := item.GetMetadata()

	// Style based on selection
	var titleStyle, descStyle lipgloss.Style
	if index == m.Index() {
		titleStyle = d.theme.ListItemSelected
		descStyle = d.theme.Muted.Background(d.theme.Colors.CurrentLine)
	} else {
		titleStyle = d.theme.ListItem
		descStyle = d.theme.Muted
	}

	// Render title
	_, _ = fmt.Fprint(w, titleStyle.Render(title))

	// Render metadata if present
	if metadata != "" {
		metadataStyle := d.theme.Muted.Foreground(d.theme.Colors.Comment)
		_, _ = fmt.Fprint(w, " ")
		_, _ = fmt.Fprint(w, metadataStyle.Render(metadata))
	}

	_, _ = fmt.Fprint(w, "\n")

	// Render description
	if description != "" {
		_, _ = fmt.Fprint(w, descStyle.Render(description))
	}
}

// NewList creates a new list with Dracula styling
func NewList(items []ListItem, width, height int, listType ListType) *List {
	draculaTheme := theme.DraculaTheme()

	// Convert ListItem to list.Item
	listItems := make([]list.Item, len(items))
	for i, item := range items {
		listItems[i] = item
	}

	delegate := itemDelegate{theme: draculaTheme}
	l := list.New(listItems, delegate, width, height)

	// Apply Dracula styling
	l.Styles.Title = draculaTheme.Title
	l.Styles.FilterPrompt = draculaTheme.SearchPrompt
	l.Styles.FilterCursor = lipgloss.NewStyle().Foreground(draculaTheme.Colors.Pink)

	// Set list type-specific title
	var title string
	switch listType {
	case ListTypeHistory:
		title = "Conversation History"
	case ListTypeSkills:
		title = "Available Skills"
	case ListTypeCommands:
		title = "Commands"
	case ListTypePlugins:
		title = "Plugins"
	case ListTypeMCPServers:
		title = "MCP Servers"
	}
	l.Title = title

	return &List{
		list:     l,
		theme:    draculaTheme,
		listType: listType,
		title:    title,
		items:    items,
		filtered: items,
	}
}

// SetTitle sets the list title
func (l *List) SetTitle(title string) {
	l.title = title
	l.list.Title = title
}

// SetItems updates the list items
func (l *List) SetItems(items []ListItem) {
	l.items = items
	l.filtered = items
	l.updateListItems()
}

// updateListItems converts filtered items to list.Items and updates the list
func (l *List) updateListItems() {
	listItems := make([]list.Item, len(l.filtered))
	for i, item := range l.filtered {
		listItems[i] = item
	}
	l.list.SetItems(listItems)
}

// FuzzySearch filters items using fuzzy search
func (l *List) FuzzySearch(query string) {
	if query == "" {
		l.filtered = l.items
		l.updateListItems()
		return
	}

	// Build search data
	searchData := make([]string, len(l.items))
	for i, item := range l.items {
		searchData[i] = item.GetSearchText()
	}

	// Perform fuzzy search
	matches := fuzzy.Find(query, searchData)

	// Update filtered items
	l.filtered = make([]ListItem, len(matches))
	for i, match := range matches {
		l.filtered[i] = l.items[match.Index]
	}

	l.updateListItems()
}

// SelectedItem returns the currently selected item
func (l *List) SelectedItem() ListItem {
	item := l.list.SelectedItem()
	if item == nil {
		return nil
	}
	return item.(ListItem)
}

// Index returns the current cursor position
func (l *List) Index() int {
	return l.list.Index()
}

// SetWidth sets the list width
func (l *List) SetWidth(width int) {
	l.list.SetWidth(width)
}

// SetHeight sets the list height
func (l *List) SetHeight(height int) {
	l.list.SetHeight(height)
}

// Update handles list updates
func (l *List) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	l.list, cmd = l.list.Update(msg)
	return cmd
}

// View renders the list
func (l *List) View() string {
	return l.list.View()
}

// ConversationHistoryList creates a list for conversation history
func ConversationHistoryList(conversations []ConversationItem, width, height int) *List {
	items := make([]ListItem, len(conversations))
	for i, conv := range conversations {
		items[i] = conv
	}
	return NewList(items, width, height, ListTypeHistory)
}

// ConversationItem represents a conversation in the history list
type ConversationItem struct {
	ID           string
	Title        string
	Created      string
	MessageCount int
	TokenCount   int
	Model        string
	IsFavorite   bool
}

// FilterValue implements list.Item
func (c ConversationItem) FilterValue() string {
	return c.Title
}

// GetSearchText implements ListItem
func (c ConversationItem) GetSearchText() string {
	return c.Title + " " + c.Model
}

// GetTitle implements ListItem
func (c ConversationItem) GetTitle() string {
	title := c.Title
	if c.IsFavorite {
		title = "★ " + title
	}
	return title
}

// GetDescription implements ListItem
func (c ConversationItem) GetDescription() string {
	return fmt.Sprintf("%d messages • %d tokens • %s", c.MessageCount, c.TokenCount, c.Model)
}

// GetMetadata implements ListItem
func (c ConversationItem) GetMetadata() string {
	return c.Created
}

// SkillsList creates a list for skills selector
func SkillsList(skills []SkillItem, width, height int) *List {
	items := make([]ListItem, len(skills))
	for i, skill := range skills {
		items[i] = skill
	}
	return NewList(items, width, height, ListTypeSkills)
}

// SkillItem represents a skill in the skills list
type SkillItem struct {
	Name        string
	Description string
	Category    string
	Enabled     bool
}

// FilterValue implements list.Item
func (s SkillItem) FilterValue() string {
	return s.Name
}

// GetSearchText implements ListItem
func (s SkillItem) GetSearchText() string {
	return s.Name + " " + s.Description + " " + s.Category
}

// GetTitle implements ListItem
func (s SkillItem) GetTitle() string {
	prefix := "○"
	if s.Enabled {
		prefix = "●"
	}
	return fmt.Sprintf("%s %s", prefix, s.Name)
}

// GetDescription implements ListItem
func (s SkillItem) GetDescription() string {
	return s.Description
}

// GetMetadata implements ListItem
func (s SkillItem) GetMetadata() string {
	return fmt.Sprintf("[%s]", s.Category)
}

// CommandsList creates a list for commands selector
func CommandsList(commands []CommandItem, width, height int) *List {
	items := make([]ListItem, len(commands))
	for i, cmd := range commands {
		items[i] = cmd
	}
	return NewList(items, width, height, ListTypeCommands)
}

// CommandItem represents a command in the commands list
type CommandItem struct {
	Name        string
	Description string
	KeyBinding  string
}

// FilterValue implements list.Item
func (c CommandItem) FilterValue() string {
	return c.Name
}

// GetSearchText implements ListItem
func (c CommandItem) GetSearchText() string {
	return c.Name + " " + c.Description + " " + c.KeyBinding
}

// GetTitle implements ListItem
func (c CommandItem) GetTitle() string {
	return c.Name
}

// GetDescription implements ListItem
func (c CommandItem) GetDescription() string {
	return c.Description
}

// GetMetadata implements ListItem
func (c CommandItem) GetMetadata() string {
	if c.KeyBinding != "" {
		return fmt.Sprintf("<%s>", c.KeyBinding)
	}
	return ""
}

// PluginsList creates a list for plugin browser
func PluginsList(plugins []PluginItem, width, height int) *List {
	items := make([]ListItem, len(plugins))
	for i, plugin := range plugins {
		items[i] = plugin
	}
	return NewList(items, width, height, ListTypePlugins)
}

// PluginItem represents a plugin in the plugins list
type PluginItem struct {
	Name        string
	Version     string
	Description string
	Status      string
	Enabled     bool
}

// FilterValue implements list.Item
func (p PluginItem) FilterValue() string {
	return p.Name
}

// GetSearchText implements ListItem
func (p PluginItem) GetSearchText() string {
	return p.Name + " " + p.Description + " " + p.Status
}

// GetTitle implements ListItem
func (p PluginItem) GetTitle() string {
	status := "○"
	if p.Enabled {
		status = "●"
	}
	return fmt.Sprintf("%s %s", status, p.Name)
}

// GetDescription implements ListItem
func (p PluginItem) GetDescription() string {
	return p.Description
}

// GetMetadata implements ListItem
func (p PluginItem) GetMetadata() string {
	return fmt.Sprintf("v%s [%s]", p.Version, strings.ToLower(p.Status))
}
