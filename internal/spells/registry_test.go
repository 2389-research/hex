// ABOUTME: Tests for spell registry thread-safe storage and retrieval
// ABOUTME: Covers Register, Get, List, All, Count, and Clear operations

package spells

import (
	"sync"
	"testing"
)

func TestRegistryRegister(t *testing.T) {
	r := NewRegistry()

	spell := &Spell{
		Name:        "test",
		Description: "Test spell",
	}

	err := r.Register(spell)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	if r.Count() != 1 {
		t.Errorf("Count = %d; want 1", r.Count())
	}
}

func TestRegistryRegister_NilSpell(t *testing.T) {
	r := NewRegistry()

	err := r.Register(nil)
	if err == nil {
		t.Fatal("Expected error for nil spell")
	}
}

func TestRegistryRegister_EmptyName(t *testing.T) {
	r := NewRegistry()

	spell := &Spell{
		Name:        "",
		Description: "No name spell",
	}

	err := r.Register(spell)
	if err == nil {
		t.Fatal("Expected error for empty name")
	}
}

func TestRegistryGet(t *testing.T) {
	r := NewRegistry()

	spell := &Spell{
		Name:        "test",
		Description: "Test spell",
	}
	_ = r.Register(spell)

	got, err := r.Get("test")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if got.Name != "test" {
		t.Errorf("Name = %q; want %q", got.Name, "test")
	}
}

func TestRegistryGet_NotFound(t *testing.T) {
	r := NewRegistry()

	_, err := r.Get("nonexistent")
	if err == nil {
		t.Fatal("Expected error for nonexistent spell")
	}
}

func TestRegistryList(t *testing.T) {
	r := NewRegistry()

	_ = r.Register(&Spell{Name: "zebra", Description: "Z"})
	_ = r.Register(&Spell{Name: "alpha", Description: "A"})

	names := r.List()

	if len(names) != 2 {
		t.Fatalf("List length = %d; want 2", len(names))
	}
	// Should be sorted
	if names[0] != "alpha" {
		t.Errorf("names[0] = %q; want %q", names[0], "alpha")
	}
	if names[1] != "zebra" {
		t.Errorf("names[1] = %q; want %q", names[1], "zebra")
	}
}

func TestRegistryAll(t *testing.T) {
	r := NewRegistry()

	_ = r.Register(&Spell{Name: "one", Description: "1"})
	_ = r.Register(&Spell{Name: "two", Description: "2"})

	all := r.All()

	if len(all) != 2 {
		t.Errorf("All length = %d; want 2", len(all))
	}

	// Should be sorted by name
	if all[0].Name != "one" {
		t.Errorf("all[0].Name = %q; want %q", all[0].Name, "one")
	}
	if all[1].Name != "two" {
		t.Errorf("all[1].Name = %q; want %q", all[1].Name, "two")
	}
}

func TestRegistryCount(t *testing.T) {
	r := NewRegistry()

	if r.Count() != 0 {
		t.Errorf("Count = %d; want 0 for empty registry", r.Count())
	}

	_ = r.Register(&Spell{Name: "one", Description: "1"})
	if r.Count() != 1 {
		t.Errorf("Count = %d; want 1", r.Count())
	}

	_ = r.Register(&Spell{Name: "two", Description: "2"})
	if r.Count() != 2 {
		t.Errorf("Count = %d; want 2", r.Count())
	}
}

func TestRegistryClear(t *testing.T) {
	r := NewRegistry()

	_ = r.Register(&Spell{Name: "one", Description: "1"})
	_ = r.Register(&Spell{Name: "two", Description: "2"})

	if r.Count() != 2 {
		t.Fatalf("Count = %d; want 2 before clear", r.Count())
	}

	r.Clear()

	if r.Count() != 0 {
		t.Errorf("Count = %d; want 0 after clear", r.Count())
	}
}

func TestRegistryOverwrite(t *testing.T) {
	r := NewRegistry()

	spell1 := &Spell{Name: "test", Description: "First version"}
	spell2 := &Spell{Name: "test", Description: "Second version"}

	_ = r.Register(spell1)
	_ = r.Register(spell2)

	// Should overwrite, not add
	if r.Count() != 1 {
		t.Errorf("Count = %d; want 1 (overwrite, not add)", r.Count())
	}

	got, _ := r.Get("test")
	if got.Description != "Second version" {
		t.Errorf("Description = %q; want %q", got.Description, "Second version")
	}
}

func TestRegistryConcurrentAccess(t *testing.T) {
	r := NewRegistry()

	// Concurrent writes
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			spell := &Spell{
				Name:        "spell" + string(rune('A'+n%26)),
				Description: "Concurrent spell",
			}
			_ = r.Register(spell)
		}(i)
	}
	wg.Wait()

	// Concurrent reads
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = r.List()
			_ = r.All()
			_ = r.Count()
		}()
	}
	wg.Wait()

	// If we get here without a race condition panic, the test passes
	if r.Count() == 0 {
		t.Error("Expected some spells to be registered")
	}
}

func TestRegistryListEmpty(t *testing.T) {
	r := NewRegistry()

	names := r.List()
	if names == nil {
		t.Error("List should return empty slice, not nil")
	}
	if len(names) != 0 {
		t.Errorf("List length = %d; want 0", len(names))
	}
}

func TestRegistryAllEmpty(t *testing.T) {
	r := NewRegistry()

	all := r.All()
	if all == nil {
		t.Error("All should return empty slice, not nil")
	}
	if len(all) != 0 {
		t.Errorf("All length = %d; want 0", len(all))
	}
}
