package components

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/table"
)

func TestNewTable(t *testing.T) {
	columns := []table.Column{
		{Title: "Name", Width: 20},
		{Title: "Status", Width: 10},
	}
	rows := []table.Row{
		{"test1", "active"},
		{"test2", "inactive"},
	}

	tbl := NewTable(columns, rows, 80, 10)

	if tbl == nil {
		t.Fatal("NewTable returned nil")
	}

	if tbl.width != 80 {
		t.Errorf("Expected width 80, got %d", tbl.width)
	}

	if tbl.height != 10 {
		t.Errorf("Expected height 10, got %d", tbl.height)
	}

	if tbl.theme == nil {
		t.Error("Theme not initialized")
	}
}

func TestTableSetTitle(t *testing.T) {
	columns := []table.Column{
		{Title: "Test", Width: 10},
	}
	rows := []table.Row{
		{"data"},
	}

	tbl := NewTable(columns, rows, 80, 10)
	tbl.SetTitle("Test Title")

	if tbl.title != "Test Title" {
		t.Errorf("Expected title 'Test Title', got '%s'", tbl.title)
	}

	view := tbl.View()
	if !strings.Contains(view, "Test Title") {
		t.Error("Title not rendered in view")
	}
}

func TestTableSetRows(t *testing.T) {
	columns := []table.Column{
		{Title: "Name", Width: 20},
	}
	initialRows := []table.Row{
		{"row1"},
	}

	tbl := NewTable(columns, initialRows, 80, 10)

	newRows := []table.Row{
		{"row2"},
		{"row3"},
	}
	tbl.SetRows(newRows)

	// Verify by checking selected row
	selected := tbl.SelectedRow()
	if len(selected) == 0 || selected[0] != "row2" {
		t.Error("SetRows did not update table rows correctly")
	}
}

func TestTableSetColumns(t *testing.T) {
	columns := []table.Column{
		{Title: "Name", Width: 20},
	}
	rows := []table.Row{
		{"data"},
	}

	tbl := NewTable(columns, rows, 80, 10)

	newColumns := []table.Column{
		{Title: "NewCol1", Width: 15},
		{Title: "NewCol2", Width: 15},
	}
	tbl.SetColumns(newColumns)

	// Table should accept new columns without error
	view := tbl.View()
	if view == "" {
		t.Error("View empty after SetColumns")
	}
}

func TestTableSetDimensions(t *testing.T) {
	columns := []table.Column{
		{Title: "Test", Width: 10},
	}
	rows := []table.Row{
		{"data"},
	}

	tbl := NewTable(columns, rows, 80, 10)

	tbl.SetWidth(100)
	if tbl.width != 100 {
		t.Errorf("Expected width 100, got %d", tbl.width)
	}

	tbl.SetHeight(20)
	if tbl.height != 20 {
		t.Errorf("Expected height 20, got %d", tbl.height)
	}
}

func TestTableFocusBlur(_ *testing.T) {
	columns := []table.Column{
		{Title: "Test", Width: 10},
	}
	rows := []table.Row{
		{"data"},
	}

	tbl := NewTable(columns, rows, 80, 10)

	// These should not panic
	tbl.Focus()
	tbl.Blur()
}

func TestTableCursor(t *testing.T) {
	columns := []table.Column{
		{Title: "Name", Width: 20},
	}
	rows := []table.Row{
		{"row1"},
		{"row2"},
		{"row3"},
	}

	tbl := NewTable(columns, rows, 80, 10)

	cursor := tbl.Cursor()
	if cursor != 0 {
		t.Errorf("Expected initial cursor at 0, got %d", cursor)
	}
}

func TestToolResultsTable(t *testing.T) {
	results := []ToolResult{
		{
			ToolName: "bash",
			Success:  true,
			Duration: "1.2s",
			Output:   "command output",
		},
		{
			ToolName: "grep",
			Success:  false,
			Duration: "0.5s",
			Output:   "error: not found",
		},
	}

	tbl := ToolResultsTable(results, 80, 10)

	if tbl == nil {
		t.Fatal("ToolResultsTable returned nil")
	}

	if tbl.title != "Tool Execution Results" {
		t.Errorf("Expected title 'Tool Execution Results', got '%s'", tbl.title)
	}

	view := tbl.View()
	if !strings.Contains(view, "bash") {
		t.Error("Tool name 'bash' not in view")
	}
	if !strings.Contains(view, "Success") {
		t.Error("Status 'Success' not in view")
	}
}

func TestConversationMetadataTable(t *testing.T) {
	metadata := []ConversationMetadata{
		{
			ID:           "conv-123",
			Created:      "2025-12-02",
			MessageCount: "10",
			TotalTokens:  "1500",
			Model:        "claude-3-opus",
		},
		{
			ID:           "conv-456",
			Created:      "2025-12-01",
			MessageCount: "5",
			TotalTokens:  "800",
			Model:        "claude-3-sonnet",
		},
	}

	tbl := ConversationMetadataTable(metadata, 80, 10)

	if tbl == nil {
		t.Fatal("ConversationMetadataTable returned nil")
	}

	if tbl.title != "Conversation History" {
		t.Errorf("Expected title 'Conversation History', got '%s'", tbl.title)
	}

	view := tbl.View()
	if !strings.Contains(view, "conv-123") {
		t.Error("Conversation ID not in view")
	}
}

func TestPluginTable(t *testing.T) {
	plugins := []PluginInfo{
		{
			Name:        "test-plugin",
			Status:      "Active",
			Version:     "1.0.0",
			Description: "Test plugin",
		},
		{
			Name:        "another-plugin",
			Status:      "Inactive",
			Version:     "2.1.0",
			Description: "Another test plugin",
		},
	}

	tbl := PluginTable(plugins, 80, 10)

	if tbl == nil {
		t.Fatal("PluginTable returned nil")
	}

	if tbl.title != "Plugins" {
		t.Errorf("Expected title 'Plugins', got '%s'", tbl.title)
	}

	view := tbl.View()
	if !strings.Contains(view, "test-plugin") {
		t.Error("Plugin name not in view")
	}
	if !strings.Contains(view, "Active") {
		t.Error("Plugin status not in view")
	}
}

func TestMCPServerTable(t *testing.T) {
	servers := []MCPServerInfo{
		{
			Name:      "server1",
			Status:    "Connected",
			ToolCount: "5",
			Endpoint:  "http://localhost:8080",
		},
		{
			Name:      "server2",
			Status:    "Disconnected",
			ToolCount: "3",
			Endpoint:  "http://localhost:8081",
		},
	}

	tbl := MCPServerTable(servers, 80, 10)

	if tbl == nil {
		t.Fatal("MCPServerTable returned nil")
	}

	if tbl.title != "MCP Servers" {
		t.Errorf("Expected title 'MCP Servers', got '%s'", tbl.title)
	}

	view := tbl.View()
	if !strings.Contains(view, "server1") {
		t.Error("Server name not in view")
	}
	if !strings.Contains(view, "Connected") {
		t.Error("Server status not in view")
	}
}
