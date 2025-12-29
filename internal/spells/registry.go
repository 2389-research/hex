// ABOUTME: Spell registry for storing and retrieving loaded spells
// ABOUTME: Thread-safe storage with lookup by name using RWMutex

package spells

import (
	"fmt"
	"sort"
	"sync"
)

// Registry stores and manages loaded spells
type Registry struct {
	spells map[string]*Spell
	mu     sync.RWMutex
}

// NewRegistry creates a new spell registry
func NewRegistry() *Registry {
	return &Registry{
		spells: make(map[string]*Spell),
	}
}

// Register adds a spell to the registry
func (r *Registry) Register(spell *Spell) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if spell == nil {
		return fmt.Errorf("cannot register nil spell")
	}

	if spell.Name == "" {
		return fmt.Errorf("spell has no name")
	}

	r.spells[spell.Name] = spell
	return nil
}

// Get retrieves a spell by name
func (r *Registry) Get(name string) (*Spell, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	spell, exists := r.spells[name]
	if !exists {
		return nil, fmt.Errorf("spell not found: %s", name)
	}

	return spell, nil
}

// List returns all registered spell names sorted alphabetically
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.spells))
	for name := range r.spells {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// All returns all registered spells sorted by name
func (r *Registry) All() []*Spell {
	r.mu.RLock()
	defer r.mu.RUnlock()

	spells := make([]*Spell, 0, len(r.spells))
	for _, spell := range r.spells {
		spells = append(spells, spell)
	}

	sort.Slice(spells, func(i, j int) bool {
		return spells[i].Name < spells[j].Name
	})

	return spells
}

// Count returns the number of registered spells
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.spells)
}

// Clear removes all spells from the registry
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.spells = make(map[string]*Spell)
}
