// Package components provides reusable UI components for the TUI.
// ABOUTME: Interactive table component using bubbles table
// ABOUTME: Renders tabular data with theme colors and keyboard navigation
package components

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/harper/pagent/internal/ui/themes"
)

// Table is a themed table component
type Table struct {
	theme       themes.Theme
	table       table.Model
	columns     []string
	rows        [][]string
	selectedRow int
}

// NewTable creates a new table component
func NewTable(theme themes.Theme, columns []string, rows [][]string) *Table {
	// Create table columns
	tableCols := make([]table.Column, len(columns))
	for i, col := range columns {
		tableCols[i] = table.Column{
			Title: col,
			Width: 20,
		}
	}

	// Create table rows
	tableRows := make([]table.Row, len(rows))
	for i, row := range rows {
		tableRows[i] = table.Row(row)
	}

	// Create styled table
	t := table.New(
		table.WithColumns(tableCols),
		table.WithRows(tableRows),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	// Apply theme styles with enhanced visual feedback
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(theme.BorderFocus()).
		BorderBottom(true).
		Bold(true).
		Foreground(theme.Primary())

	// Enhanced selected style with better visual feedback
	s.Selected = s.Selected.
		Foreground(theme.Background()).
		Background(theme.Primary()).
		Bold(true).
		Underline(true).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Primary()).
		Padding(0, 1)

	t.SetStyles(s)

	return &Table{
		theme:       theme,
		table:       t,
		columns:     columns,
		rows:        rows,
		selectedRow: 0,
	}
}

// Init implements tea.Model
func (t *Table) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (t *Table) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	t.table, cmd = t.table.Update(msg)
	t.selectedRow = t.table.Cursor()
	return t, cmd
}

// View implements tea.Model
func (t *Table) View() string {
	return t.table.View()
}

// GetSelectedRow returns the currently selected row index
func (t *Table) GetSelectedRow() int {
	return t.selectedRow
}

// MoveDown moves selection down
func (t *Table) MoveDown() {
	t.table.MoveDown(1)
	t.selectedRow = t.table.Cursor()
}

// MoveUp moves selection up
func (t *Table) MoveUp() {
	t.table.MoveUp(1)
	t.selectedRow = t.table.Cursor()
}

// SetRows updates the table rows
func (t *Table) SetRows(rows [][]string) {
	tableRows := make([]table.Row, len(rows))
	for i, row := range rows {
		tableRows[i] = table.Row(row)
	}
	t.table.SetRows(tableRows)
	t.rows = rows
}
