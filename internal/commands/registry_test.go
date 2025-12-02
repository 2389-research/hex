// ABOUTME: Tests for command registry storage and retrieval
// ABOUTME: Validates thread-safe command management operations

package commands

import (
	"sync"
	"testing"
)

func TestNewRegistry(t *testing.T) {
	reg := NewRegistry()
	if reg == nil {
		t.Fatal("NewRegistry() returned nil")
	}

	if reg.Count() != 0 {
		t.Errorf("New registry count = %d, want 0", reg.Count())
	}
}

func TestRegister(t *testing.T) {
	reg := NewRegistry()

	cmd := &Command{
		Name:        "test",
		Description: "Test command",
		Content:     "Test content",
	}

	err := reg.Register(cmd)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	if reg.Count() != 1 {
		t.Errorf("Count() = %d, want 1", reg.Count())
	}
}

func TestRegisterEmptyName(t *testing.T) {
	reg := NewRegistry()

	cmd := &Command{
		Name:        "",
		Description: "Test command",
	}

	err := reg.Register(cmd)
	if err == nil {
		t.Error("Expected error for empty command name")
	}
}

func TestRegisterAll(t *testing.T) {
	reg := NewRegistry()

	commands := []*Command{
		{Name: "cmd1", Description: "Command 1"},
		{Name: "cmd2", Description: "Command 2"},
		{Name: "cmd3", Description: "Command 3"},
	}

	err := reg.RegisterAll(commands)
	if err != nil {
		t.Fatalf("RegisterAll() error = %v", err)
	}

	if reg.Count() != 3 {
		t.Errorf("Count() = %d, want 3", reg.Count())
	}
}

func TestGet(t *testing.T) {
	reg := NewRegistry()

	cmd := &Command{
		Name:        "test",
		Description: "Test command",
		Content:     "Test content",
	}

	_ = reg.Register(cmd) //nolint:errcheck // test setup

	retrieved, err := reg.Get("test")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if retrieved.Name != cmd.Name {
		t.Errorf("Retrieved command name = %q, want %q", retrieved.Name, cmd.Name)
	}
}

func TestGetNotFound(t *testing.T) {
	reg := NewRegistry()

	_, err := reg.Get("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent command")
	}
}

func TestHas(t *testing.T) {
	reg := NewRegistry()

	cmd := &Command{
		Name:        "test",
		Description: "Test command",
	}

	_ = reg.Register(cmd) //nolint:errcheck // test setup

	if !reg.Has("test") {
		t.Error("Has() = false, want true")
	}

	if reg.Has("nonexistent") {
		t.Error("Has() = true, want false")
	}
}

func TestList(t *testing.T) {
	reg := NewRegistry()

	commands := []*Command{
		{Name: "zebra", Description: "Last alphabetically"},
		{Name: "alpha", Description: "First alphabetically"},
		{Name: "beta", Description: "Middle alphabetically"},
	}

	_ = reg.RegisterAll(commands) //nolint:errcheck // test setup

	list := reg.List()

	// Should be sorted alphabetically
	expectedOrder := []string{"alpha", "beta", "zebra"}
	if len(list) != len(expectedOrder) {
		t.Fatalf("List() returned %d items, want %d", len(list), len(expectedOrder))
	}

	for i, name := range list {
		if name != expectedOrder[i] {
			t.Errorf("List()[%d] = %q, want %q", i, name, expectedOrder[i])
		}
	}
}

func TestAll(t *testing.T) {
	reg := NewRegistry()

	commands := []*Command{
		{Name: "cmd1", Description: "Command 1"},
		{Name: "cmd2", Description: "Command 2"},
	}

	_ = reg.RegisterAll(commands) //nolint:errcheck // test setup

	all := reg.All()

	if len(all) != 2 {
		t.Fatalf("All() returned %d commands, want 2", len(all))
	}

	// Verify sorting by name
	if all[0].Name > all[1].Name {
		t.Error("All() not sorted by name")
	}
}

func TestRemove(t *testing.T) {
	reg := NewRegistry()

	cmd := &Command{
		Name:        "test",
		Description: "Test command",
	}

	_ = reg.Register(cmd) //nolint:errcheck // test setup

	if !reg.Remove("test") {
		t.Error("Remove() = false, want true")
	}

	if reg.Has("test") {
		t.Error("Command still exists after Remove()")
	}

	if reg.Count() != 0 {
		t.Errorf("Count() = %d after Remove(), want 0", reg.Count())
	}
}

func TestRemoveNonexistent(t *testing.T) {
	reg := NewRegistry()

	if reg.Remove("nonexistent") {
		t.Error("Remove() = true for nonexistent command, want false")
	}
}

func TestClear(t *testing.T) {
	reg := NewRegistry()

	commands := []*Command{
		{Name: "cmd1", Description: "Command 1"},
		{Name: "cmd2", Description: "Command 2"},
		{Name: "cmd3", Description: "Command 3"},
	}

	_ = reg.RegisterAll(commands) //nolint:errcheck // test setup

	reg.Clear()

	if reg.Count() != 0 {
		t.Errorf("Count() = %d after Clear(), want 0", reg.Count())
	}

	if reg.Has("cmd1") {
		t.Error("Command still exists after Clear()")
	}
}

func TestConcurrentAccess(t *testing.T) {
	reg := NewRegistry()

	// Pre-populate with some commands
	for i := 0; i < 10; i++ {
		cmd := &Command{
			Name:        string(rune('a' + i)),
			Description: "Command",
		}
		_ = reg.Register(cmd) //nolint:errcheck // test setup
	}

	var wg sync.WaitGroup
	iterations := 100

	// Concurrent reads
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				name := string(rune('a' + idx))
				_, err := reg.Get(name)
				if err != nil {
					t.Errorf("Concurrent Get() error = %v", err)
				}
				_ = reg.Has(name)
				_ = reg.List()
			}
		}(i)
	}

	// Concurrent writes
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				cmd := &Command{
					Name:        string(rune('A' + idx)),
					Description: "Concurrent command",
				}
				_ = reg.Register(cmd) //nolint:errcheck // test setup
			}
		}(i)
	}

	wg.Wait()

	// Verify registry is still consistent
	if reg.Count() != 15 { // 10 initial + 5 concurrent
		t.Errorf("Final count = %d, want 15", reg.Count())
	}
}

func TestRegisterOverwrite(t *testing.T) {
	reg := NewRegistry()

	cmd1 := &Command{
		Name:        "test",
		Description: "First version",
		Content:     "Content 1",
	}

	cmd2 := &Command{
		Name:        "test",
		Description: "Second version",
		Content:     "Content 2",
	}

	_ = reg.Register(cmd1) //nolint:errcheck // test setup
	_ = reg.Register(cmd2) //nolint:errcheck // test setup (should overwrite)

	if reg.Count() != 1 {
		t.Errorf("Count() = %d after overwrite, want 1", reg.Count())
	}

	retrieved, _ := reg.Get("test")
	if retrieved.Description != "Second version" {
		t.Errorf("Description = %q, want %q (overwrite failed)", retrieved.Description, "Second version")
	}
}
