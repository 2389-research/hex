// ABOUTME: Tests for command file discovery and loading
// ABOUTME: Validates directory scanning and command file parsing

package commands

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewLoader(t *testing.T) {
	loader := NewLoader()

	if loader.UserDir == "" {
		t.Error("UserDir should not be empty")
	}

	// UserDir should contain .hex/commands
	if !filepath.IsAbs(loader.UserDir) {
		t.Error("UserDir should be absolute path")
	}
}

func TestLoadFromDir(t *testing.T) {
	// Create temporary directory for test commands
	tempDir := t.TempDir()

	// Create test command files
	testCommands := map[string]string{
		"test1.md": `---
name: test1
description: First test command
---
Content 1`,
		"test2.md": `---
name: test2
description: Second test command
args:
  file: File to test
---
Content 2`,
		"invalid.md": `---
name: invalid
---
Missing description`,
		"notmarkdown.txt": "This should be ignored",
	}

	for filename, content := range testCommands {
		path := filepath.Join(tempDir, filename)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil { //nolint:gosec // G306 - test file
			t.Fatalf("Failed to write test file: %v", err)
		}
	}

	loader := &Loader{}
	commands, err := loader.loadFromDir(tempDir, "test")
	if err != nil {
		t.Fatalf("loadFromDir() error = %v", err)
	}

	// Should load 2 valid commands (test1 and test2), skip invalid and .txt
	if len(commands) != 2 {
		t.Errorf("loadFromDir() loaded %d commands, want 2", len(commands))
	}

	// Verify source is set correctly
	for _, cmd := range commands {
		if cmd.Source != "test" {
			t.Errorf("Command %s has source %q, want %q", cmd.Name, cmd.Source, "test")
		}
	}

	// Verify command names
	names := make(map[string]bool)
	for _, cmd := range commands {
		names[cmd.Name] = true
	}
	if !names["test1"] || !names["test2"] {
		t.Error("Expected to find test1 and test2 commands")
	}
}

func TestLoadFromDirNonexistent(t *testing.T) {
	loader := &Loader{}
	_, err := loader.loadFromDir("/nonexistent/directory", "test")
	if err == nil {
		t.Error("Expected error for nonexistent directory")
	}
	if !os.IsNotExist(err) {
		t.Errorf("Expected os.IsNotExist error, got %v", err)
	}
}

func TestLoadAll(t *testing.T) {
	// Create temporary directories
	builtinDir := t.TempDir()
	userDir := t.TempDir()
	projectDir := t.TempDir()

	// Create builtin command
	builtinCmd := `---
name: builtin-cmd
description: Built-in command
---
Builtin content`
	_ = os.WriteFile(filepath.Join(builtinDir, "builtin.md"), []byte(builtinCmd), 0644) //nolint:errcheck,gosec // test setup

	// Create user command (overrides builtin)
	userCmd := `---
name: builtin-cmd
description: User override
---
User content`
	_ = os.WriteFile(filepath.Join(userDir, "builtin.md"), []byte(userCmd), 0644) //nolint:errcheck,gosec // test setup

	// Create project command
	projectCmd := `---
name: project-cmd
description: Project command
---
Project content`
	_ = os.WriteFile(filepath.Join(projectDir, "project.md"), []byte(projectCmd), 0644) //nolint:errcheck,gosec // test setup

	// Create another project command that overrides user
	projectOverride := `---
name: builtin-cmd
description: Project override
---
Project override content`
	_ = os.WriteFile(filepath.Join(projectDir, "override.md"), []byte(projectOverride), 0644) //nolint:errcheck,gosec // test setup

	loader := &Loader{
		BuiltinDir: builtinDir,
		UserDir:    userDir,
		ProjectDir: projectDir,
	}

	commands, err := loader.LoadAll()
	if err != nil {
		t.Fatalf("LoadAll() error = %v", err)
	}

	// Should have 2 commands: project-cmd and builtin-cmd (project override wins)
	if len(commands) != 2 {
		t.Errorf("LoadAll() loaded %d commands, want 2", len(commands))
	}

	// Find builtin-cmd and verify it's the project override
	var builtinCmdFound *Command
	for _, cmd := range commands {
		if cmd.Name == "builtin-cmd" {
			builtinCmdFound = cmd
			break
		}
	}

	if builtinCmdFound == nil {
		t.Fatal("builtin-cmd not found")
	}

	if builtinCmdFound.Source != "project" {
		t.Errorf("builtin-cmd source = %q, want %q", builtinCmdFound.Source, "project")
	}

	if builtinCmdFound.Description != "Project override" {
		t.Errorf("builtin-cmd description = %q, want %q", builtinCmdFound.Description, "Project override")
	}
}

func TestLoadByName(t *testing.T) {
	// Create temporary directories
	userDir := t.TempDir()
	projectDir := t.TempDir()

	// Create user command
	userCmd := `---
name: user-cmd
description: User command
---
User content`
	_ = os.WriteFile(filepath.Join(userDir, "user-cmd.md"), []byte(userCmd), 0644) //nolint:errcheck,gosec // test setup

	// Create project command with same name
	projectCmd := `---
name: user-cmd
description: Project command
---
Project content`
	_ = os.WriteFile(filepath.Join(projectDir, "user-cmd.md"), []byte(projectCmd), 0644) //nolint:errcheck,gosec // test setup

	loader := &Loader{
		UserDir:    userDir,
		ProjectDir: projectDir,
	}

	// Should load project command (higher priority)
	cmd, err := loader.LoadByName("user-cmd")
	if err != nil {
		t.Fatalf("LoadByName() error = %v", err)
	}

	if cmd.Source != "project" {
		t.Errorf("Source = %q, want %q", cmd.Source, "project")
	}

	if cmd.Description != "Project command" {
		t.Errorf("Description = %q, want %q", cmd.Description, "Project command")
	}
}

func TestLoadByNameNotFound(t *testing.T) {
	loader := &Loader{
		UserDir:    t.TempDir(),
		ProjectDir: t.TempDir(),
	}

	_, err := loader.LoadByName("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent command")
	}
}

func TestLoadAllEmptyDirs(t *testing.T) {
	loader := &Loader{
		UserDir:    t.TempDir(),
		ProjectDir: t.TempDir(),
		BuiltinDir: t.TempDir(),
	}

	commands, err := loader.LoadAll()
	if err != nil {
		t.Fatalf("LoadAll() error = %v", err)
	}

	if len(commands) != 0 {
		t.Errorf("LoadAll() loaded %d commands, want 0", len(commands))
	}
}

func TestLoadAllSorting(t *testing.T) {
	tempDir := t.TempDir()

	// Create commands in non-alphabetical order
	commands := []string{"zebra", "alpha", "beta"}
	for _, name := range commands {
		content := `---
name: ` + name + `
description: Command ` + name + `
---
Content`
		_ = os.WriteFile(filepath.Join(tempDir, name+".md"), []byte(content), 0644) //nolint:errcheck,gosec // test setup
	}

	loader := &Loader{
		ProjectDir: tempDir,
	}

	loaded, err := loader.LoadAll()
	if err != nil {
		t.Fatalf("LoadAll() error = %v", err)
	}

	// Verify alphabetical ordering
	if len(loaded) != 3 {
		t.Fatalf("LoadAll() loaded %d commands, want 3", len(loaded))
	}

	expectedOrder := []string{"alpha", "beta", "zebra"}
	for i, cmd := range loaded {
		if cmd.Name != expectedOrder[i] {
			t.Errorf("Command %d name = %q, want %q", i, cmd.Name, expectedOrder[i])
		}
	}
}
