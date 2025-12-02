package components

import (
	"strings"
	"testing"
)

func TestNewDefaultListItem(t *testing.T) {
	item := NewDefaultListItem("Title", "Description", "Metadata")

	if item.GetTitle() != "Title" {
		t.Errorf("Expected title 'Title', got '%s'", item.GetTitle())
	}

	if item.GetDescription() != "Description" {
		t.Errorf("Expected description 'Description', got '%s'", item.GetDescription())
	}

	if item.GetMetadata() != "Metadata" {
		t.Errorf("Expected metadata 'Metadata', got '%s'", item.GetMetadata())
	}

	if item.FilterValue() != "Title" {
		t.Errorf("Expected filter value 'Title', got '%s'", item.FilterValue())
	}

	searchText := item.GetSearchText()
	if !strings.Contains(searchText, "Title") || !strings.Contains(searchText, "Description") {
		t.Error("GetSearchText should contain title and description")
	}
}

func TestNewList(t *testing.T) {
	items := []ListItem{
		NewDefaultListItem("Item 1", "Desc 1", "Meta 1"),
		NewDefaultListItem("Item 2", "Desc 2", "Meta 2"),
	}

	tests := []struct {
		name     string
		listType ListType
		expTitle string
	}{
		{"history", ListTypeHistory, "Conversation History"},
		{"skills", ListTypeSkills, "Available Skills"},
		{"commands", ListTypeCommands, "Commands"},
		{"plugins", ListTypePlugins, "Plugins"},
		{"mcp", ListTypeMCPServers, "MCP Servers"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			list := NewList(items, 80, 20, tt.listType)

			if list == nil {
				t.Fatal("NewList returned nil")
			}

			if list.title != tt.expTitle {
				t.Errorf("Expected title '%s', got '%s'", tt.expTitle, list.title)
			}

			if list.listType != tt.listType {
				t.Errorf("Expected type %v, got %v", tt.listType, list.listType)
			}

			if len(list.items) != len(items) {
				t.Errorf("Expected %d items, got %d", len(items), len(list.items))
			}

			if list.theme == nil {
				t.Error("Theme not initialized")
			}
		})
	}
}

func TestListSetTitle(t *testing.T) {
	items := []ListItem{
		NewDefaultListItem("Item 1", "Desc 1", ""),
	}

	list := NewList(items, 80, 20, ListTypeHistory)
	newTitle := "Custom Title"

	list.SetTitle(newTitle)

	if list.title != newTitle {
		t.Errorf("Expected title '%s', got '%s'", newTitle, list.title)
	}

	if list.list.Title != newTitle {
		t.Errorf("Expected list.Title '%s', got '%s'", newTitle, list.list.Title)
	}
}

func TestListSetItems(t *testing.T) {
	items := []ListItem{
		NewDefaultListItem("Item 1", "Desc 1", ""),
	}

	list := NewList(items, 80, 20, ListTypeHistory)

	newItems := []ListItem{
		NewDefaultListItem("New Item 1", "New Desc 1", ""),
		NewDefaultListItem("New Item 2", "New Desc 2", ""),
	}

	list.SetItems(newItems)

	if len(list.items) != len(newItems) {
		t.Errorf("Expected %d items, got %d", len(newItems), len(list.items))
	}

	if len(list.filtered) != len(newItems) {
		t.Errorf("Expected %d filtered items, got %d", len(newItems), len(list.filtered))
	}
}

func TestListFuzzySearch(t *testing.T) {
	items := []ListItem{
		NewDefaultListItem("bash command", "Execute bash", ""),
		NewDefaultListItem("python script", "Run python", ""),
		NewDefaultListItem("grep search", "Search files", ""),
		NewDefaultListItem("git status", "Show git status", ""),
	}

	list := NewList(items, 80, 20, ListTypeCommands)

	// Test search that matches
	list.FuzzySearch("bash")
	if len(list.filtered) == 0 {
		t.Error("FuzzySearch should find 'bash'")
	}

	// Verify the match
	found := false
	for _, item := range list.filtered {
		if strings.Contains(item.GetTitle(), "bash") {
			found = true
			break
		}
	}
	if !found {
		t.Error("FuzzySearch did not return matching item")
	}

	// Test empty search resets filter
	list.FuzzySearch("")
	if len(list.filtered) != len(items) {
		t.Errorf("Empty search should show all items: expected %d, got %d", len(items), len(list.filtered))
	}

	// Test search with partial match
	list.FuzzySearch("py")
	if len(list.filtered) == 0 {
		t.Error("FuzzySearch should find items with 'py'")
	}
}

func TestListSelectedItem(t *testing.T) {
	items := []ListItem{
		NewDefaultListItem("Item 1", "Desc 1", ""),
		NewDefaultListItem("Item 2", "Desc 2", ""),
	}

	list := NewList(items, 80, 20, ListTypeHistory)

	selected := list.SelectedItem()
	if selected == nil {
		t.Fatal("SelectedItem returned nil")
	}

	// Should select first item by default
	if selected.GetTitle() != "Item 1" {
		t.Errorf("Expected first item selected, got '%s'", selected.GetTitle())
	}
}

func TestListIndex(t *testing.T) {
	items := []ListItem{
		NewDefaultListItem("Item 1", "Desc 1", ""),
		NewDefaultListItem("Item 2", "Desc 2", ""),
	}

	list := NewList(items, 80, 20, ListTypeHistory)

	index := list.Index()
	if index != 0 {
		t.Errorf("Expected initial index 0, got %d", index)
	}
}

func TestListSetDimensions(_ *testing.T) {
	items := []ListItem{
		NewDefaultListItem("Item 1", "Desc 1", ""),
	}

	list := NewList(items, 80, 20, ListTypeHistory)

	list.SetWidth(100)
	list.SetHeight(30)

	// These should not panic
}

func TestListView(t *testing.T) {
	items := []ListItem{
		NewDefaultListItem("Item 1", "Desc 1", ""),
	}

	list := NewList(items, 80, 20, ListTypeHistory)

	view := list.View()
	if view == "" {
		t.Error("View returned empty string")
	}

	// Should contain the title
	if !strings.Contains(view, "Conversation History") {
		t.Error("View should contain list title")
	}
}

func TestConversationHistoryList(t *testing.T) {
	conversations := []ConversationItem{
		{
			ID:           "conv1",
			Title:        "Test Conversation",
			Created:      "2025-12-02",
			MessageCount: 10,
			TokenCount:   1500,
			Model:        "claude-3-opus",
			IsFavorite:   true,
		},
		{
			ID:           "conv2",
			Title:        "Another Chat",
			Created:      "2025-12-01",
			MessageCount: 5,
			TokenCount:   800,
			Model:        "claude-3-sonnet",
			IsFavorite:   false,
		},
	}

	list := ConversationHistoryList(conversations, 80, 20)

	if list == nil {
		t.Fatal("ConversationHistoryList returned nil")
	}

	if list.listType != ListTypeHistory {
		t.Error("List type should be History")
	}

	if len(list.items) != len(conversations) {
		t.Errorf("Expected %d items, got %d", len(conversations), len(list.items))
	}
}

func TestConversationItem(t *testing.T) {
	item := ConversationItem{
		ID:           "conv1",
		Title:        "Test",
		Created:      "2025-12-02",
		MessageCount: 10,
		TokenCount:   1500,
		Model:        "claude-3-opus",
		IsFavorite:   true,
	}

	title := item.GetTitle()
	if !strings.Contains(title, "★") {
		t.Error("Favorite conversation should have star in title")
	}

	desc := item.GetDescription()
	if !strings.Contains(desc, "10 messages") {
		t.Error("Description should contain message count")
	}
	if !strings.Contains(desc, "1500 tokens") {
		t.Error("Description should contain token count")
	}

	metadata := item.GetMetadata()
	if metadata != "2025-12-02" {
		t.Errorf("Expected metadata '2025-12-02', got '%s'", metadata)
	}
}

func TestSkillsList(t *testing.T) {
	skills := []SkillItem{
		{
			Name:        "test-skill",
			Description: "A test skill",
			Category:    "testing",
			Enabled:     true,
		},
	}

	list := SkillsList(skills, 80, 20)

	if list == nil {
		t.Fatal("SkillsList returned nil")
	}

	if list.listType != ListTypeSkills {
		t.Error("List type should be Skills")
	}
}

func TestSkillItem(t *testing.T) {
	item := SkillItem{
		Name:        "test-skill",
		Description: "A test skill",
		Category:    "testing",
		Enabled:     true,
	}

	title := item.GetTitle()
	if !strings.Contains(title, "●") {
		t.Error("Enabled skill should have filled circle")
	}

	metadata := item.GetMetadata()
	if !strings.Contains(metadata, "testing") {
		t.Error("Metadata should contain category")
	}

	// Test disabled skill
	item.Enabled = false
	title = item.GetTitle()
	if !strings.Contains(title, "○") {
		t.Error("Disabled skill should have empty circle")
	}
}

func TestCommandsList(t *testing.T) {
	commands := []CommandItem{
		{
			Name:        "help",
			Description: "Show help",
			KeyBinding:  "?",
		},
	}

	list := CommandsList(commands, 80, 20)

	if list == nil {
		t.Fatal("CommandsList returned nil")
	}

	if list.listType != ListTypeCommands {
		t.Error("List type should be Commands")
	}
}

func TestCommandItem(t *testing.T) {
	item := CommandItem{
		Name:        "help",
		Description: "Show help",
		KeyBinding:  "?",
	}

	title := item.GetTitle()
	if title != "help" {
		t.Errorf("Expected title 'help', got '%s'", title)
	}

	metadata := item.GetMetadata()
	if !strings.Contains(metadata, "?") {
		t.Error("Metadata should contain key binding")
	}

	searchText := item.GetSearchText()
	if !strings.Contains(searchText, "?") {
		t.Error("Search text should include key binding")
	}
}

func TestPluginsList(t *testing.T) {
	plugins := []PluginItem{
		{
			Name:        "test-plugin",
			Version:     "1.0.0",
			Description: "A test plugin",
			Status:      "Active",
			Enabled:     true,
		},
	}

	list := PluginsList(plugins, 80, 20)

	if list == nil {
		t.Fatal("PluginsList returned nil")
	}

	if list.listType != ListTypePlugins {
		t.Error("List type should be Plugins")
	}
}

func TestPluginItem(t *testing.T) {
	item := PluginItem{
		Name:        "test-plugin",
		Version:     "1.0.0",
		Description: "A test plugin",
		Status:      "Active",
		Enabled:     true,
	}

	title := item.GetTitle()
	if !strings.Contains(title, "●") {
		t.Error("Enabled plugin should have filled circle")
	}

	metadata := item.GetMetadata()
	if !strings.Contains(metadata, "1.0.0") {
		t.Error("Metadata should contain version")
	}
	if !strings.Contains(metadata, "active") {
		t.Error("Metadata should contain status (lowercase)")
	}

	// Test disabled plugin
	item.Enabled = false
	title = item.GetTitle()
	if !strings.Contains(title, "○") {
		t.Error("Disabled plugin should have empty circle")
	}
}
