// ABOUTME: Command file discovery and loading from multiple directories
// ABOUTME: Scans user, project, and builtin directories for .md command files

package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Loader discovers and loads commands from multiple directories
type Loader struct {
	UserDir    string // User-global commands directory (~/.clem/commands/)
	ProjectDir string // Project-local commands directory (.claude/commands/)
	BuiltinDir string // Built-in commands directory (embedded or distributed)
}

// NewLoader creates a loader with default directories
func NewLoader() *Loader {
	homeDir, _ := os.UserHomeDir()
	userDir := filepath.Join(homeDir, ".clem", "commands")

	// Find project directory by looking for .claude directory
	projectDir := findProjectCommandsDir()

	return &Loader{
		UserDir:    userDir,
		ProjectDir: projectDir,
		BuiltinDir: "", // Will be set by caller if builtin commands exist
	}
}

// findProjectCommandsDir searches for .claude/commands/ directory
func findProjectCommandsDir() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}

	// Search upwards for .claude directory
	searchDir := cwd
	for i := 0; i < 10; i++ {
		claudeDir := filepath.Join(searchDir, ".claude", "commands")
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

// LoadAll discovers and loads all commands from all directories
// Later directories override earlier ones if command names conflict
func (l *Loader) LoadAll() ([]*Command, error) {
	commandsByName := make(map[string]*Command)
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

		commands, err := l.loadFromDir(src.dir, src.source)
		if err != nil {
			// Don't fail if directory doesn't exist
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("load commands from %s: %w", src.dir, err)
		}

		for _, cmd := range commands {
			// Later sources override earlier ones
			if existing, exists := commandsByName[cmd.Name]; exists {
				// Track what got overridden
				_ = existing // Could log this
			} else {
				loadOrder = append(loadOrder, cmd.Name)
			}
			commandsByName[cmd.Name] = cmd
		}
	}

	// Convert map to slice in load order
	result := make([]*Command, 0, len(commandsByName))
	for _, name := range loadOrder {
		result = append(result, commandsByName[name])
	}

	// Sort by name alphabetically
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result, nil
}

// loadFromDir loads all .md files from a directory
func (l *Loader) loadFromDir(dir, source string) ([]*Command, error) {
	// Check if directory exists
	info, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", dir)
	}

	// Find all .md files
	var commandPaths []string
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".md") {
			commandPaths = append(commandPaths, path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk directory: %w", err)
	}

	// Parse each command file
	commands := make([]*Command, 0, len(commandPaths))
	for _, path := range commandPaths {
		cmd, err := Parse(path)
		if err != nil {
			// Log warning but continue loading other commands
			fmt.Fprintf(os.Stderr, "Warning: failed to parse command %s: %v\n", path, err)
			continue
		}
		cmd.Source = source
		commands = append(commands, cmd)
	}

	return commands, nil
}

// LoadByName loads a specific command by name from any directory
func (l *Loader) LoadByName(name string) (*Command, error) {
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

		cmd, err := Parse(path)
		if err != nil {
			return nil, fmt.Errorf("parse command: %w", err)
		}
		cmd.Source = src.source
		return cmd, nil
	}

	return nil, fmt.Errorf("command not found: %s", name)
}
