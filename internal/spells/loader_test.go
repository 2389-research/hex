// ABOUTME: Tests for spell loader with directory discovery and priority cascade
// ABOUTME: Verifies loading from user, project, and builtin directories

package spells

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
}

func TestLoaderLoadAll_EmptyDirectories(t *testing.T) {
	loader := &Loader{
		UserDir:    "/nonexistent/user",
		ProjectDir: "/nonexistent/project",
		BuiltinDir: "",
	}

	spells, err := loader.LoadAll()
	if err != nil {
		t.Fatalf("LoadAll failed: %v", err)
	}

	if len(spells) != 0 {
		t.Errorf("Expected 0 spells from nonexistent directories, got %d", len(spells))
	}
}

func TestLoaderLoadAll_WithSpells(t *testing.T) {
	tmpDir := t.TempDir()
	userDir := filepath.Join(tmpDir, "user")
	projectDir := filepath.Join(tmpDir, "project")

	// Create user spell
	userSpellDir := filepath.Join(userDir, "user-spell")
	if err := os.MkdirAll(userSpellDir, 0755); err != nil {
		t.Fatalf("Failed to create user spell dir: %v", err)
	}
	userSystem := []byte(`---
name: user-spell
description: User spell
---
User prompt.
`)
	if err := os.WriteFile(filepath.Join(userSpellDir, "system.md"), userSystem, 0644); err != nil {
		t.Fatalf("Failed to write user system.md: %v", err)
	}

	// Create project spell
	projectSpellDir := filepath.Join(projectDir, "project-spell")
	if err := os.MkdirAll(projectSpellDir, 0755); err != nil {
		t.Fatalf("Failed to create project spell dir: %v", err)
	}
	projectSystem := []byte(`---
name: project-spell
description: Project spell
---
Project prompt.
`)
	if err := os.WriteFile(filepath.Join(projectSpellDir, "system.md"), projectSystem, 0644); err != nil {
		t.Fatalf("Failed to write project system.md: %v", err)
	}

	loader := &Loader{
		UserDir:    userDir,
		ProjectDir: projectDir,
	}

	spells, err := loader.LoadAll()
	if err != nil {
		t.Fatalf("LoadAll failed: %v", err)
	}

	if len(spells) != 2 {
		t.Fatalf("Expected 2 spells, got %d", len(spells))
	}
}

func TestLoaderLoadAll_ProjectOverridesUser(t *testing.T) {
	tmpDir := t.TempDir()
	userDir := filepath.Join(tmpDir, "user")
	projectDir := filepath.Join(tmpDir, "project")

	// Same spell name in both
	for _, dir := range []string{userDir, projectDir} {
		spellDir := filepath.Join(dir, "shared-spell")
		if err := os.MkdirAll(spellDir, 0755); err != nil {
			t.Fatalf("Failed to create spell dir: %v", err)
		}
	}

	userSystem := []byte(`---
name: shared-spell
description: User version
---
User.
`)
	projectSystem := []byte(`---
name: shared-spell
description: Project version
---
Project.
`)

	if err := os.WriteFile(filepath.Join(userDir, "shared-spell", "system.md"), userSystem, 0644); err != nil {
		t.Fatalf("Failed to write user system.md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, "shared-spell", "system.md"), projectSystem, 0644); err != nil {
		t.Fatalf("Failed to write project system.md: %v", err)
	}

	loader := &Loader{
		UserDir:    userDir,
		ProjectDir: projectDir,
	}

	spells, err := loader.LoadAll()
	if err != nil {
		t.Fatalf("LoadAll failed: %v", err)
	}

	if len(spells) != 1 {
		t.Fatalf("Expected 1 spell (project overrides user), got %d", len(spells))
	}

	if spells[0].Description != "Project version" {
		t.Errorf("Description = %q; want %q", spells[0].Description, "Project version")
	}
	if spells[0].Source != "project" {
		t.Errorf("Source = %q; want %q", spells[0].Source, "project")
	}
}

func TestLoaderLoadByName_Found(t *testing.T) {
	tmpDir := t.TempDir()
	userDir := filepath.Join(tmpDir, "user")

	spellDir := filepath.Join(userDir, "test-spell")
	if err := os.MkdirAll(spellDir, 0755); err != nil {
		t.Fatalf("Failed to create spell dir: %v", err)
	}

	systemMd := []byte(`---
name: test-spell
description: Test spell
---
Test prompt.
`)
	if err := os.WriteFile(filepath.Join(spellDir, "system.md"), systemMd, 0644); err != nil {
		t.Fatalf("Failed to write system.md: %v", err)
	}

	loader := &Loader{UserDir: userDir}

	spell, err := loader.LoadByName("test-spell")
	if err != nil {
		t.Fatalf("LoadByName failed: %v", err)
	}

	if spell.Name != "test-spell" {
		t.Errorf("Name = %q; want %q", spell.Name, "test-spell")
	}
	if spell.Source != "user" {
		t.Errorf("Source = %q; want %q", spell.Source, "user")
	}
}

func TestLoaderLoadByName_NotFound(t *testing.T) {
	loader := &Loader{
		UserDir:    "/nonexistent/user",
		ProjectDir: "/nonexistent/project",
	}

	_, err := loader.LoadByName("nonexistent")
	if err == nil {
		t.Fatal("Expected error for nonexistent spell, got nil")
	}
}

func TestLoaderLoadByName_ProjectOverridesUser(t *testing.T) {
	tmpDir := t.TempDir()
	userDir := filepath.Join(tmpDir, "user")
	projectDir := filepath.Join(tmpDir, "project")

	// Same spell name in both
	for _, dir := range []string{userDir, projectDir} {
		spellDir := filepath.Join(dir, "override-spell")
		if err := os.MkdirAll(spellDir, 0755); err != nil {
			t.Fatalf("Failed to create spell dir: %v", err)
		}
	}

	userSystem := []byte(`---
name: override-spell
description: User version
---
User.
`)
	projectSystem := []byte(`---
name: override-spell
description: Project version
---
Project.
`)

	if err := os.WriteFile(filepath.Join(userDir, "override-spell", "system.md"), userSystem, 0644); err != nil {
		t.Fatalf("Failed to write user system.md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, "override-spell", "system.md"), projectSystem, 0644); err != nil {
		t.Fatalf("Failed to write project system.md: %v", err)
	}

	loader := &Loader{
		UserDir:    userDir,
		ProjectDir: projectDir,
	}

	// LoadByName should find project version first
	spell, err := loader.LoadByName("override-spell")
	if err != nil {
		t.Fatalf("LoadByName failed: %v", err)
	}

	if spell.Description != "Project version" {
		t.Errorf("Description = %q; want %q", spell.Description, "Project version")
	}
	if spell.Source != "project" {
		t.Errorf("Source = %q; want %q", spell.Source, "project")
	}
}

func TestLoaderLoadAll_BuiltinLowestPriority(t *testing.T) {
	tmpDir := t.TempDir()
	builtinDir := filepath.Join(tmpDir, "builtin")
	userDir := filepath.Join(tmpDir, "user")

	// Create builtin spell
	builtinSpellDir := filepath.Join(builtinDir, "shared-spell")
	if err := os.MkdirAll(builtinSpellDir, 0755); err != nil {
		t.Fatalf("Failed to create builtin spell dir: %v", err)
	}
	builtinSystem := []byte(`---
name: shared-spell
description: Builtin version
---
Builtin.
`)
	if err := os.WriteFile(filepath.Join(builtinSpellDir, "system.md"), builtinSystem, 0644); err != nil {
		t.Fatalf("Failed to write builtin system.md: %v", err)
	}

	// Create user spell with same name
	userSpellDir := filepath.Join(userDir, "shared-spell")
	if err := os.MkdirAll(userSpellDir, 0755); err != nil {
		t.Fatalf("Failed to create user spell dir: %v", err)
	}
	userSystem := []byte(`---
name: shared-spell
description: User version
---
User.
`)
	if err := os.WriteFile(filepath.Join(userSpellDir, "system.md"), userSystem, 0644); err != nil {
		t.Fatalf("Failed to write user system.md: %v", err)
	}

	loader := &Loader{
		BuiltinDir: builtinDir,
		UserDir:    userDir,
	}

	spells, err := loader.LoadAll()
	if err != nil {
		t.Fatalf("LoadAll failed: %v", err)
	}

	if len(spells) != 1 {
		t.Fatalf("Expected 1 spell (user overrides builtin), got %d", len(spells))
	}

	if spells[0].Description != "User version" {
		t.Errorf("Description = %q; want %q", spells[0].Description, "User version")
	}
	if spells[0].Source != "user" {
		t.Errorf("Source = %q; want %q", spells[0].Source, "user")
	}
}

func TestLoaderLoadAll_SkipsInvalidSpells(t *testing.T) {
	tmpDir := t.TempDir()
	userDir := filepath.Join(tmpDir, "user")

	// Create valid spell
	validSpellDir := filepath.Join(userDir, "valid-spell")
	if err := os.MkdirAll(validSpellDir, 0755); err != nil {
		t.Fatalf("Failed to create valid spell dir: %v", err)
	}
	validSystem := []byte(`---
name: valid-spell
description: Valid spell
---
Valid.
`)
	if err := os.WriteFile(filepath.Join(validSpellDir, "system.md"), validSystem, 0644); err != nil {
		t.Fatalf("Failed to write valid system.md: %v", err)
	}

	// Create invalid spell (missing description)
	invalidSpellDir := filepath.Join(userDir, "invalid-spell")
	if err := os.MkdirAll(invalidSpellDir, 0755); err != nil {
		t.Fatalf("Failed to create invalid spell dir: %v", err)
	}
	invalidSystem := []byte(`---
name: invalid-spell
---
Missing description.
`)
	if err := os.WriteFile(filepath.Join(invalidSpellDir, "system.md"), invalidSystem, 0644); err != nil {
		t.Fatalf("Failed to write invalid system.md: %v", err)
	}

	loader := &Loader{UserDir: userDir}

	spells, err := loader.LoadAll()
	if err != nil {
		t.Fatalf("LoadAll failed: %v", err)
	}

	// Should only have the valid spell
	if len(spells) != 1 {
		t.Fatalf("Expected 1 valid spell, got %d", len(spells))
	}

	if spells[0].Name != "valid-spell" {
		t.Errorf("Name = %q; want %q", spells[0].Name, "valid-spell")
	}
}

func TestLoaderLoadAll_SkipsNonDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	userDir := filepath.Join(tmpDir, "user")
	if err := os.MkdirAll(userDir, 0755); err != nil {
		t.Fatalf("Failed to create user dir: %v", err)
	}

	// Create a file (not a directory) in the spells dir
	if err := os.WriteFile(filepath.Join(userDir, "not-a-spell.txt"), []byte("hello"), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	// Create valid spell directory
	spellDir := filepath.Join(userDir, "valid-spell")
	if err := os.MkdirAll(spellDir, 0755); err != nil {
		t.Fatalf("Failed to create spell dir: %v", err)
	}
	systemMd := []byte(`---
name: valid-spell
description: Valid spell
---
Valid.
`)
	if err := os.WriteFile(filepath.Join(spellDir, "system.md"), systemMd, 0644); err != nil {
		t.Fatalf("Failed to write system.md: %v", err)
	}

	loader := &Loader{UserDir: userDir}

	spells, err := loader.LoadAll()
	if err != nil {
		t.Fatalf("LoadAll failed: %v", err)
	}

	if len(spells) != 1 {
		t.Fatalf("Expected 1 spell, got %d", len(spells))
	}
}

func TestLoaderLoadAll_SkipsDirectoriesWithoutSystemMd(t *testing.T) {
	tmpDir := t.TempDir()
	userDir := filepath.Join(tmpDir, "user")

	// Create directory without system.md
	noSystemDir := filepath.Join(userDir, "no-system-spell")
	if err := os.MkdirAll(noSystemDir, 0755); err != nil {
		t.Fatalf("Failed to create no-system spell dir: %v", err)
	}
	// Write some other file
	if err := os.WriteFile(filepath.Join(noSystemDir, "readme.md"), []byte("hello"), 0644); err != nil {
		t.Fatalf("Failed to write readme: %v", err)
	}

	// Create valid spell directory
	spellDir := filepath.Join(userDir, "valid-spell")
	if err := os.MkdirAll(spellDir, 0755); err != nil {
		t.Fatalf("Failed to create spell dir: %v", err)
	}
	systemMd := []byte(`---
name: valid-spell
description: Valid spell
---
Valid.
`)
	if err := os.WriteFile(filepath.Join(spellDir, "system.md"), systemMd, 0644); err != nil {
		t.Fatalf("Failed to write system.md: %v", err)
	}

	loader := &Loader{UserDir: userDir}

	spells, err := loader.LoadAll()
	if err != nil {
		t.Fatalf("LoadAll failed: %v", err)
	}

	if len(spells) != 1 {
		t.Fatalf("Expected 1 spell, got %d", len(spells))
	}
}

func TestLoaderLoadAll_SortedAlphabetically(t *testing.T) {
	tmpDir := t.TempDir()
	userDir := filepath.Join(tmpDir, "user")

	// Create spells with names that should sort: alpha, beta, charlie
	names := []string{"charlie", "alpha", "beta"}
	for _, name := range names {
		spellDir := filepath.Join(userDir, name)
		if err := os.MkdirAll(spellDir, 0755); err != nil {
			t.Fatalf("Failed to create spell dir: %v", err)
		}
		systemMd := []byte(`---
name: ` + name + `
description: ` + name + ` spell
---
` + name + ` prompt.
`)
		if err := os.WriteFile(filepath.Join(spellDir, "system.md"), systemMd, 0644); err != nil {
			t.Fatalf("Failed to write system.md: %v", err)
		}
	}

	loader := &Loader{UserDir: userDir}

	spells, err := loader.LoadAll()
	if err != nil {
		t.Fatalf("LoadAll failed: %v", err)
	}

	if len(spells) != 3 {
		t.Fatalf("Expected 3 spells, got %d", len(spells))
	}

	// Check alphabetical order
	expected := []string{"alpha", "beta", "charlie"}
	for i, exp := range expected {
		if spells[i].Name != exp {
			t.Errorf("spells[%d].Name = %q; want %q", i, spells[i].Name, exp)
		}
	}
}
