// ABOUTME: Skill file discovery and loading from multiple directories
// ABOUTME: Scans user, project, and builtin directories for .md skill files

// Package skills provides skill loading and tool integration.
package skills

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Loader discovers and loads skills from multiple directories
type Loader struct {
	UserDir    string // User-global skills directory (~/.clem/skills/)
	ProjectDir string // Project-local skills directory (.claude/skills/)
	BuiltinDir string // Built-in skills directory (embedded or distributed)
}

// NewLoader creates a loader with default directories
func NewLoader() *Loader {
	homeDir, _ := os.UserHomeDir()
	userDir := filepath.Join(homeDir, ".clem", "skills")

	// Find project directory by looking for .claude directory
	projectDir := findProjectSkillsDir()

	return &Loader{
		UserDir:    userDir,
		ProjectDir: projectDir,
		BuiltinDir: "", // Will be set by caller if builtin skills exist
	}
}

// findProjectSkillsDir searches for .claude/skills/ directory
func findProjectSkillsDir() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}

	// Search upwards for .claude directory
	searchDir := cwd
	for i := 0; i < 10; i++ {
		claudeDir := filepath.Join(searchDir, ".claude", "skills")
		if info, err := os.Stat(claudeDir); err == nil && info.IsDir() {
			return claudeDir
		}

		parent := filepath.Dir(searchDir)
		if parent == searchDir {
			break // Reached filesystem root
		}
		searchDir = parent
	}

	return ""
}

// LoadAll discovers and loads all skills from all directories
// Later directories override earlier ones if skill names conflict
func (l *Loader) LoadAll() ([]*Skill, error) {
	skillsByName := make(map[string]*Skill)
	var loadOrder []string

	// Load in priority order: builtin -> user -> project
	// (project wins conflicts)
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

		skills, err := l.loadFromDir(src.dir, src.source)
		if err != nil {
			// Don't fail if directory doesn't exist
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("load skills from %s: %w", src.dir, err)
		}

		for _, skill := range skills {
			// Later sources override earlier ones
			if existing, exists := skillsByName[skill.Name]; exists {
				// Track what got overridden
				_ = existing // Could log this
			} else {
				loadOrder = append(loadOrder, skill.Name)
			}
			skillsByName[skill.Name] = skill
		}
	}

	// Convert map to slice in load order
	result := make([]*Skill, 0, len(skillsByName))
	for _, name := range loadOrder {
		result = append(result, skillsByName[name])
	}

	// Sort by priority (higher first), then name
	sort.Slice(result, func(i, j int) bool {
		if result[i].Priority != result[j].Priority {
			return result[i].Priority > result[j].Priority
		}
		return result[i].Name < result[j].Name
	})

	return result, nil
}

// loadFromDir loads all .md files from a directory
func (l *Loader) loadFromDir(dir, source string) ([]*Skill, error) {
	// Check if directory exists
	info, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", dir)
	}

	// Find all .md files
	var skillPaths []string
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".md") {
			skillPaths = append(skillPaths, path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk directory: %w", err)
	}

	// Parse each skill file
	skills := make([]*Skill, 0, len(skillPaths))
	for _, path := range skillPaths {
		skill, err := Parse(path)
		if err != nil {
			// Log warning but continue loading other skills
			fmt.Fprintf(os.Stderr, "Warning: failed to parse skill %s: %v\n", path, err)
			continue
		}
		skill.Source = source
		skills = append(skills, skill)
	}

	return skills, nil
}

// LoadByName loads a specific skill by name from any directory
func (l *Loader) LoadByName(name string) (*Skill, error) {
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

		// Look for <name>.md in directory
		path := filepath.Join(src.dir, name+".md")
		if _, err := os.Stat(path); err != nil {
			continue
		}

		skill, err := Parse(path)
		if err != nil {
			return nil, fmt.Errorf("parse skill: %w", err)
		}
		skill.Source = src.source
		return skill, nil
	}

	return nil, fmt.Errorf("skill not found: %s", name)
}
