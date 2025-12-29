// ABOUTME: Spell directory discovery and loading from multiple locations
// ABOUTME: Scans user, project, and builtin directories for spell subdirectories

package spells

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/2389-research/hex/internal/project"
)

// Loader discovers and loads spells from multiple directories
type Loader struct {
	UserDir    string // User-global spells directory (~/.hex/spells/)
	ProjectDir string // Project-local spells directory (.hex/spells/)
	BuiltinDir string // Built-in spells directory
}

// NewLoader creates a loader with default directories
func NewLoader() *Loader {
	homeDir, _ := os.UserHomeDir()
	userDir := filepath.Join(homeDir, ".hex", "spells")

	// Find project directory by looking for .hex directory
	projectDir := project.FindDir("spells")

	return &Loader{
		UserDir:    userDir,
		ProjectDir: projectDir,
		BuiltinDir: "", // Will be set by caller if builtin spells exist
	}
}

// LoadAll discovers and loads all spells from all directories
// Later directories override earlier ones if spell names conflict
func (l *Loader) LoadAll() ([]*Spell, error) {
	spellsByName := make(map[string]*Spell)
	var loadOrder []string

	// Load in priority order: builtin -> user -> project
	sources := []struct {
		dir    string
		source string
	}{
		{l.BuiltinDir, "builtin"},
		{l.UserDir, "user"},
		{l.ProjectDir, "project"},
	}

	for _, src := range sources {
		if src.dir == "" {
			continue
		}

		spells, err := l.loadFromDir(src.dir, src.source)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("load spells from %s: %w", src.dir, err)
		}

		for _, spell := range spells {
			if _, exists := spellsByName[spell.Name]; !exists {
				loadOrder = append(loadOrder, spell.Name)
			}
			spellsByName[spell.Name] = spell
		}
	}

	// Convert to slice in load order
	result := make([]*Spell, 0, len(spellsByName))
	for _, name := range loadOrder {
		result = append(result, spellsByName[name])
	}

	// Sort alphabetically by name
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result, nil
}

// loadFromDir loads all spell directories from a parent directory
func (l *Loader) loadFromDir(dir, source string) ([]*Spell, error) {
	info, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", dir)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read directory: %w", err)
	}

	var spells []*Spell
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		spellDir := filepath.Join(dir, entry.Name())

		// Check if it has a system.md file
		if _, err := os.Stat(filepath.Join(spellDir, "system.md")); os.IsNotExist(err) {
			continue
		}

		spell, err := ParseSpellDirectory(spellDir)
		if err != nil {
			// Log warning but continue loading other spells
			fmt.Fprintf(os.Stderr, "Warning: failed to parse spell %s: %v\n", entry.Name(), err)
			continue
		}

		spell.Source = source
		spells = append(spells, spell)
	}

	return spells, nil
}

// LoadByName loads a specific spell by name from any directory
func (l *Loader) LoadByName(name string) (*Spell, error) {
	// Try project first, then user, then builtin
	sources := []struct {
		dir    string
		source string
	}{
		{l.ProjectDir, "project"},
		{l.UserDir, "user"},
		{l.BuiltinDir, "builtin"},
	}

	for _, src := range sources {
		if src.dir == "" {
			continue
		}

		spellDir := filepath.Join(src.dir, name)
		if _, err := os.Stat(filepath.Join(spellDir, "system.md")); err != nil {
			continue
		}

		spell, err := ParseSpellDirectory(spellDir)
		if err != nil {
			return nil, fmt.Errorf("parse spell: %w", err)
		}
		spell.Source = src.source
		return spell, nil
	}

	return nil, fmt.Errorf("spell not found: %s", name)
}
