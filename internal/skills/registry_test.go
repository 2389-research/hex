package skills

import (
	"testing"
)

func TestNewRegistry(t *testing.T) {
	registry := NewRegistry()
	if registry == nil {
		t.Fatal("NewRegistry returned nil")
	}
	if registry.Count() != 0 {
		t.Errorf("New registry should be empty, got count %d", registry.Count())
	}
}

func TestRegistryRegister(t *testing.T) {
	registry := NewRegistry()

	skill := &Skill{
		Name:        "test-skill",
		Description: "Test description",
	}

	err := registry.Register(skill)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	if registry.Count() != 1 {
		t.Errorf("Count = %d; want 1", registry.Count())
	}
}

func TestRegistryRegister_NoName(t *testing.T) {
	registry := NewRegistry()

	skill := &Skill{
		Description: "No name",
	}

	err := registry.Register(skill)
	if err == nil {
		t.Fatal("Expected error for skill with no name, got nil")
	}
}

func TestRegistryRegister_Override(t *testing.T) {
	registry := NewRegistry()

	skill1 := &Skill{
		Name:        "test-skill",
		Description: "First version",
	}
	skill2 := &Skill{
		Name:        "test-skill",
		Description: "Second version",
	}

	_ = registry.Register(skill1)
	_ = registry.Register(skill2)

	// Should have only one skill (overridden)
	if registry.Count() != 1 {
		t.Errorf("Count = %d; want 1 (override)", registry.Count())
	}

	// Should be second version
	loaded, _ := registry.Get("test-skill")
	if loaded.Description != "Second version" {
		t.Errorf("Description = %q; want %q", loaded.Description, "Second version")
	}
}

func TestRegistryGet(t *testing.T) {
	registry := NewRegistry()

	skill := &Skill{
		Name:        "test-skill",
		Description: "Test",
		Priority:    7,
	}

	_ = registry.Register(skill)

	loaded, err := registry.Get("test-skill")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if loaded.Name != "test-skill" {
		t.Errorf("Name = %q; want %q", loaded.Name, "test-skill")
	}
	if loaded.Priority != 7 {
		t.Errorf("Priority = %d; want 7", loaded.Priority)
	}
}

func TestRegistryGet_NotFound(t *testing.T) {
	registry := NewRegistry()

	_, err := registry.Get("nonexistent")
	if err == nil {
		t.Fatal("Expected error for nonexistent skill, got nil")
	}
}

func TestRegistryList(t *testing.T) {
	registry := NewRegistry()

	skills := []*Skill{
		{Name: "zebra-skill", Description: "Z"},
		{Name: "alpha-skill", Description: "A"},
		{Name: "middle-skill", Description: "M"},
	}

	for _, s := range skills {
		_ = registry.Register(s)
	}

	names := registry.List()
	if len(names) != 3 {
		t.Fatalf("List length = %d; want 3", len(names))
	}

	// Should be sorted alphabetically
	expected := []string{"alpha-skill", "middle-skill", "zebra-skill"}
	for i, name := range names {
		if name != expected[i] {
			t.Errorf("names[%d] = %q; want %q", i, name, expected[i])
		}
	}
}

func TestRegistryAll(t *testing.T) {
	registry := NewRegistry()

	skills := []*Skill{
		{Name: "low-priority", Description: "Low", Priority: 3},
		{Name: "high-priority", Description: "High", Priority: 9},
		{Name: "medium-priority", Description: "Med", Priority: 5},
	}

	for _, s := range skills {
		_ = registry.Register(s)
	}

	all := registry.All()
	if len(all) != 3 {
		t.Fatalf("All length = %d; want 3", len(all))
	}

	// Should be sorted by priority (high to low)
	if all[0].Name != "high-priority" {
		t.Errorf("all[0].Name = %q; want %q", all[0].Name, "high-priority")
	}
	if all[1].Name != "medium-priority" {
		t.Errorf("all[1].Name = %q; want %q", all[1].Name, "medium-priority")
	}
	if all[2].Name != "low-priority" {
		t.Errorf("all[2].Name = %q; want %q", all[2].Name, "low-priority")
	}
}

func TestRegistryFindByTags(t *testing.T) {
	registry := NewRegistry()

	skills := []*Skill{
		{Name: "skill1", Description: "S1", Tags: []string{"testing", "go"}},
		{Name: "skill2", Description: "S2", Tags: []string{"debugging", "go"}},
		{Name: "skill3", Description: "S3", Tags: []string{"testing", "python"}},
	}

	for _, s := range skills {
		_ = registry.Register(s)
	}

	// Find by "testing" tag
	matches := registry.FindByTags("testing")
	if len(matches) != 2 {
		t.Errorf("FindByTags(testing) length = %d; want 2", len(matches))
	}

	// Find by "go" tag
	matches = registry.FindByTags("go")
	if len(matches) != 2 {
		t.Errorf("FindByTags(go) length = %d; want 2", len(matches))
	}

	// Find by "debugging" tag
	matches = registry.FindByTags("debugging")
	if len(matches) != 1 {
		t.Errorf("FindByTags(debugging) length = %d; want 1", len(matches))
	}

	// Find by nonexistent tag
	matches = registry.FindByTags("nonexistent")
	if len(matches) != 0 {
		t.Errorf("FindByTags(nonexistent) length = %d; want 0", len(matches))
	}
}

func TestRegistryFindByTags_MultipleTags(t *testing.T) {
	registry := NewRegistry()

	skills := []*Skill{
		{Name: "skill1", Description: "S1", Tags: []string{"testing"}},
		{Name: "skill2", Description: "S2", Tags: []string{"debugging"}},
		{Name: "skill3", Description: "S3", Tags: []string{"performance"}},
	}

	for _, s := range skills {
		_ = registry.Register(s)
	}

	// Find by multiple tags (OR operation)
	matches := registry.FindByTags("testing", "debugging")
	if len(matches) != 2 {
		t.Errorf("FindByTags(testing, debugging) length = %d; want 2", len(matches))
	}
}

func TestRegistryFindByPattern(t *testing.T) {
	registry := NewRegistry()

	skills := []*Skill{
		{
			Name:               "tdd-skill",
			Description:        "TDD",
			ActivationPatterns: []string{"write.*test", "test.*driven"},
		},
		{
			Name:               "debug-skill",
			Description:        "Debug",
			ActivationPatterns: []string{"debug", "fix.*bug"},
		},
		{
			Name:        "no-patterns",
			Description: "No patterns",
		},
	}

	for _, s := range skills {
		_ = registry.Register(s)
	}

	// Find by "write tests" message
	matches := registry.FindByPattern("write tests for feature")
	if len(matches) != 1 {
		t.Fatalf("FindByPattern length = %d; want 1", len(matches))
	}
	if matches[0].Name != "tdd-skill" {
		t.Errorf("Match name = %q; want %q", matches[0].Name, "tdd-skill")
	}

	// Find by "debug" message
	matches = registry.FindByPattern("debug this issue")
	if len(matches) != 1 {
		t.Fatalf("FindByPattern length = %d; want 1", len(matches))
	}
	if matches[0].Name != "debug-skill" {
		t.Errorf("Match name = %q; want %q", matches[0].Name, "debug-skill")
	}

	// No matches
	matches = registry.FindByPattern("random message")
	if len(matches) != 0 {
		t.Errorf("FindByPattern length = %d; want 0", len(matches))
	}
}

func TestRegistrySearch(t *testing.T) {
	registry := NewRegistry()

	skills := []*Skill{
		{Name: "test-skill", Description: "Testing methodology", Tags: []string{"testing"}},
		{Name: "debug-skill", Description: "Debugging guide", Tags: []string{"debugging"}},
		{Name: "code-review", Description: "Code review checklist", Tags: []string{"quality"}},
	}

	for _, s := range skills {
		_ = registry.Register(s)
	}

	tests := []struct {
		query     string
		wantCount int
		wantName  string
	}{
		{"test", 1, "test-skill"},        // Matches name only (not tag "testing" - "test" is not substring of "testing")
		{"testing", 1, "test-skill"},     // Matches tag
		{"debug", 1, "debug-skill"},      // Matches name and description
		{"review", 1, "code-review"},     // Matches name
		{"methodology", 1, "test-skill"}, // Matches description
		{"quality", 1, "code-review"},    // Matches tag
		{"nonexistent", 0, ""},           // No matches
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			matches := registry.Search(tt.query)
			if len(matches) != tt.wantCount {
				t.Errorf("Search(%q) length = %d; want %d", tt.query, len(matches), tt.wantCount)
			}
			if tt.wantCount > 0 && matches[0].Name != tt.wantName {
				t.Errorf("First match name = %q; want %q", matches[0].Name, tt.wantName)
			}
		})
	}
}

func TestRegistrySearch_CaseInsensitive(t *testing.T) {
	registry := NewRegistry()

	skill := &Skill{
		Name:        "test-skill",
		Description: "Testing Methodology",
		Tags:        []string{"Testing"},
	}
	_ = registry.Register(skill)

	// All should match regardless of case
	queries := []string{"test", "TEST", "Test", "testing", "TESTING"}
	for _, q := range queries {
		matches := registry.Search(q)
		if len(matches) != 1 {
			t.Errorf("Search(%q) length = %d; want 1 (case-insensitive)", q, len(matches))
		}
	}
}

func TestRegistryCount(t *testing.T) {
	registry := NewRegistry()

	if registry.Count() != 0 {
		t.Errorf("Initial count = %d; want 0", registry.Count())
	}

	_ = registry.Register(&Skill{Name: "skill1", Description: "S1"})
	if registry.Count() != 1 {
		t.Errorf("Count after 1 register = %d; want 1", registry.Count())
	}

	_ = registry.Register(&Skill{Name: "skill2", Description: "S2"})
	if registry.Count() != 2 {
		t.Errorf("Count after 2 registers = %d; want 2", registry.Count())
	}
}

func TestRegistryClear(t *testing.T) {
	registry := NewRegistry()

	_ = registry.Register(&Skill{Name: "skill1", Description: "S1"})
	_ = registry.Register(&Skill{Name: "skill2", Description: "S2"})

	if registry.Count() != 2 {
		t.Fatalf("Count before clear = %d; want 2", registry.Count())
	}

	registry.Clear()

	if registry.Count() != 0 {
		t.Errorf("Count after clear = %d; want 0", registry.Count())
	}

	// Verify actually cleared
	_, err := registry.Get("skill1")
	if err == nil {
		t.Error("Expected error for cleared skill, got nil")
	}
}

func TestRegistryConcurrency(t *testing.T) {
	registry := NewRegistry()

	// Concurrent writes
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(n int) {
			skill := &Skill{
				Name:        string(rune('a' + n)),
				Description: "Concurrent",
			}
			_ = registry.Register(skill)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify count
	if registry.Count() != 10 {
		t.Errorf("Count after concurrent writes = %d; want 10", registry.Count())
	}
}
