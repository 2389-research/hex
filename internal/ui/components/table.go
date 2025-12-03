// Package components provides reusable Bubbles components with Dracula theme styling.
// ABOUTME: Table component for displaying structured data with sorting and navigation
// ABOUTME: Wraps bubbles.Table with Dracula theme and additional features
package components

import (
	"github.com/2389-research/hex/internal/ui/theme"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Table wraps bubbles.Table with Dracula styling and additional functionality
type Table struct {
	table  table.Model
	theme  *theme.Theme
	title  string
	width  int
	height int
}

// NewTable creates a new table with Dracula styling
func NewTable(columns []table.Column, rows []table.Row, width, height int) *Table {
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(height),
	)

	draculaTheme := theme.DraculaTheme()

	// Apply Dracula styles
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(draculaTheme.Colors.Purple).
		BorderBottom(true).
		Bold(true).
		Foreground(draculaTheme.Colors.Pink)

	s.Selected = s.Selected.
		Foreground(draculaTheme.Colors.Background).
		Background(draculaTheme.Colors.Purple).
		Bold(true)

	s.Cell = s.Cell.
		Foreground(draculaTheme.Colors.Foreground)

	t.SetStyles(s)

	return &Table{
		table:  t,
		theme:  draculaTheme,
		width:  width,
		height: height,
	}
}

// SetTitle sets the table title
func (t *Table) SetTitle(title string) {
	t.title = title
}

// SetRows updates the table rows
func (t *Table) SetRows(rows []table.Row) {
	t.table.SetRows(rows)
}

// SetColumns updates the table columns
func (t *Table) SetColumns(columns []table.Column) {
	t.table.SetColumns(columns)
}

// SetWidth sets the table width
func (t *Table) SetWidth(width int) {
	t.width = width
}

// SetHeight sets the table height
func (t *Table) SetHeight(height int) {
	t.height = height
	t.table.SetHeight(height)
}

// Focus sets focus on the table
func (t *Table) Focus() {
	t.table.Focus()
}

// Blur removes focus from the table
func (t *Table) Blur() {
	t.table.Blur()
}

// SelectedRow returns the currently selected row
func (t *Table) SelectedRow() table.Row {
	return t.table.SelectedRow()
}

// Cursor returns the current cursor position
func (t *Table) Cursor() int {
	return t.table.Cursor()
}

// Update handles table updates
func (t *Table) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	t.table, cmd = t.table.Update(msg)
	return cmd
}

// View renders the table
func (t *Table) View() string {
	if t.title != "" {
		titleStyle := t.theme.Title.
			Width(t.width).
			Align(lipgloss.Center)
		return titleStyle.Render(t.title) + "\n" + t.table.View()
	}
	return t.table.View()
}

// ToolResultsTable creates a table for displaying tool execution results
func ToolResultsTable(results []ToolResult, width, height int) *Table {
	columns := []table.Column{
		{Title: "Tool", Width: 20},
		{Title: "Status", Width: 10},
		{Title: "Duration", Width: 10},
		{Title: "Output", Width: width - 45},
	}

	rows := make([]table.Row, len(results))
	for i, result := range results {
		status := "Success"
		if !result.Success {
			status = "Failed"
		}
		rows[i] = table.Row{
			result.ToolName,
			status,
			result.Duration,
			result.Output,
		}
	}

	t := NewTable(columns, rows, width, height)
	t.SetTitle("Tool Execution Results")
	return t
}

// ToolResult represents a tool execution result for table display
type ToolResult struct {
	ToolName string
	Success  bool
	Duration string
	Output   string
}

// ConversationMetadataTable creates a table for displaying conversation metadata
func ConversationMetadataTable(metadata []ConversationMetadata, width, height int) *Table {
	columns := []table.Column{
		{Title: "ID", Width: 20},
		{Title: "Created", Width: 20},
		{Title: "Messages", Width: 10},
		{Title: "Tokens", Width: 10},
		{Title: "Model", Width: width - 65},
	}

	rows := make([]table.Row, len(metadata))
	for i, m := range metadata {
		rows[i] = table.Row{
			m.ID,
			m.Created,
			m.MessageCount,
			m.TotalTokens,
			m.Model,
		}
	}

	t := NewTable(columns, rows, width, height)
	t.SetTitle("Conversation History")
	return t
}

// ConversationMetadata represents conversation metadata for table display
type ConversationMetadata struct {
	ID           string
	Created      string
	MessageCount string
	TotalTokens  string
	Model        string
}

// PluginTable creates a table for displaying plugin status
func PluginTable(plugins []PluginInfo, width, height int) *Table {
	columns := []table.Column{
		{Title: "Name", Width: 20},
		{Title: "Status", Width: 15},
		{Title: "Version", Width: 10},
		{Title: "Description", Width: width - 50},
	}

	rows := make([]table.Row, len(plugins))
	for i, p := range plugins {
		rows[i] = table.Row{
			p.Name,
			p.Status,
			p.Version,
			p.Description,
		}
	}

	t := NewTable(columns, rows, width, height)
	t.SetTitle("Plugins")
	return t
}

// PluginInfo represents plugin information for table display
type PluginInfo struct {
	Name        string
	Status      string
	Version     string
	Description string
}

// MCPServerTable creates a table for displaying MCP server status
func MCPServerTable(servers []MCPServerInfo, width, height int) *Table {
	columns := []table.Column{
		{Title: "Name", Width: 20},
		{Title: "Status", Width: 15},
		{Title: "Tools", Width: 10},
		{Title: "Endpoint", Width: width - 50},
	}

	rows := make([]table.Row, len(servers))
	for i, s := range servers {
		rows[i] = table.Row{
			s.Name,
			s.Status,
			s.ToolCount,
			s.Endpoint,
		}
	}

	t := NewTable(columns, rows, width, height)
	t.SetTitle("MCP Servers")
	return t
}

// MCPServerInfo represents MCP server information for table display
type MCPServerInfo struct {
	Name      string
	Status    string
	ToolCount string
	Endpoint  string
}
