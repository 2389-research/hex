// ABOUTME: Test suite for interactive table component using bubbles
// ABOUTME: Validates table rendering, theming, and interaction
package components

import (
	"testing"

	"github.com/harper/pagent/internal/ui/themes"
	"github.com/stretchr/testify/assert"
)

func TestNewTable(t *testing.T) {
	theme := themes.NewDracula()
	columns := []string{"From", "Subject", "Date"}
	rows := [][]string{
		{"alice@example.com", "Meeting", "2h ago"},
		{"bob@example.com", "Report", "5h ago"},
	}

	table := NewTable(theme, columns, rows)

	assert.NotNil(t, table)
	assert.Equal(t, theme, table.theme)
	assert.Equal(t, 2, len(rows))
}

func TestTableView(t *testing.T) {
	theme := themes.NewDracula()
	columns := []string{"Name", "Value"}
	rows := [][]string{{"Key1", "Value1"}}

	table := NewTable(theme, columns, rows)
	view := table.View()

	assert.Contains(t, view, "Name")
	assert.Contains(t, view, "Value")
	assert.Contains(t, view, "Key1")
}

func TestTableSelection(t *testing.T) {
	theme := themes.NewDracula()
	columns := []string{"Name"}
	rows := [][]string{{"Row1"}, {"Row2"}, {"Row3"}}

	table := NewTable(theme, columns, rows)

	// Initially first row selected
	assert.Equal(t, 0, table.GetSelectedRow())

	// Move down
	table.MoveDown()
	assert.Equal(t, 1, table.GetSelectedRow())

	// Move up
	table.MoveUp()
	assert.Equal(t, 0, table.GetSelectedRow())
}
