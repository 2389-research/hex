package skills

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
	// ProjectDir may be empty if not in a project
}

func TestLoaderLoadAll_EmptyDirectories(t *testing.T) {
	loader := &Loader{
		UserDir:    "/nonexistent/user",
		ProjectDir: "/nonexistent/project",
		BuiltinDir: "",
	}

	skills, err := loader.LoadAll()
	if err != nil {
		t.Fatalf("LoadAll failed: %v", err)
	}

	if len(skills) != 0 {
		t.Errorf("Expected 0 skills from nonexistent directories, got %d", len(skills))
	}
}

func TestLoaderLoadAll_WithSkills(t *testing.T) {
	// Create temporary directories
	tmpDir := t.TempDir()
	userDir := filepath.Join(tmpDir, "user")
	projectDir := filepath.Join(tmpDir, "project")

	if err := os.MkdirAll(userDir, 0755); err != nil { //nolint:gosec // G301 - test dir
		t.Fatalf("Failed to create user dir: %v", err)
	}
	if err := os.MkdirAll(projectDir, 0755); err != nil { //nolint:gosec // G301 - test dir
		t.Fatalf("Failed to create project dir: %v", err)
	}

	// Create skill in user dir
	userSkill := []byte(`---
name: user-skill
description: User skill
priority: 5
---
User content
`)
	if err := os.WriteFile(filepath.Join(userDir, "user-skill.md"), userSkill, 0644); err != nil { //nolint:gosec // G306 - test file
		t.Fatalf("Failed to write user skill: %v", err)
	}

	// Create skill in project dir
	projectSkill := []byte(`---
name: project-skill
description: Project skill
priority: 7
---
Project content
`)
	if err := os.WriteFile(filepath.Join(projectDir, "project-skill.md"), projectSkill, 0644); err != nil { //nolint:gosec // G306 - test file
		t.Fatalf("Failed to write project skill: %v", err)
	}

	loader := &Loader{
		UserDir:    userDir,
		ProjectDir: projectDir,
	}

	skills, err := loader.LoadAll()
	if err != nil {
		t.Fatalf("LoadAll failed: %v", err)
	}

	if len(skills) != 2 {
		t.Fatalf("Expected 2 skills, got %d", len(skills))
	}

	// Verify skills are sorted by priority (project-skill should be first)
	if skills[0].Name != "project-skill" {
		t.Errorf("First skill name = %q; want %q", skills[0].Name, "project-skill")
	}
	if skills[1].Name != "user-skill" {
		t.Errorf("Second skill name = %q; want %q", skills[1].Name, "user-skill")
	}
}

func TestLoaderLoadAll_ProjectOverridesUser(t *testing.T) {
	tmpDir := t.TempDir()
	userDir := filepath.Join(tmpDir, "user")
	projectDir := filepath.Join(tmpDir, "project")

	if err := os.MkdirAll(userDir, 0755); err != nil { //nolint:gosec // G301 - test dir
		t.Fatalf("Failed to create user dir: %v", err)
	}
	if err := os.MkdirAll(projectDir, 0755); err != nil { //nolint:gosec // G301 - test dir
		t.Fatalf("Failed to create project dir: %v", err)
	}

	// Same skill name in both directories
	userSkill := []byte(`---
name: shared-skill
description: User version
priority: 5
---
User content
`)
	projectSkill := []byte(`---
name: shared-skill
description: Project version
priority: 8
---
Project content
`)

	if err := os.WriteFile(filepath.Join(userDir, "shared-skill.md"), userSkill, 0644); err != nil { //nolint:gosec // G306 - test file
		t.Fatalf("Failed to write user skill: %v", err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, "shared-skill.md"), projectSkill, 0644); err != nil { //nolint:gosec // G306 - test file
		t.Fatalf("Failed to write project skill: %v", err)
	}

	loader := &Loader{
		UserDir:    userDir,
		ProjectDir: projectDir,
	}

	skills, err := loader.LoadAll()
	if err != nil {
		t.Fatalf("LoadAll failed: %v", err)
	}

	if len(skills) != 1 {
		t.Fatalf("Expected 1 skill (project overrides user), got %d", len(skills))
	}

	skill := skills[0]
	if skill.Description != "Project version" {
		t.Errorf("Description = %q; want %q (project should override user)", skill.Description, "Project version")
	}
	if skill.Source != "project" {
		t.Errorf("Source = %q; want %q", skill.Source, "project")
	}
}

func TestLoaderLoadFromDir_InvalidSkill(t *testing.T) {
	tmpDir := t.TempDir()

	// Create skill with missing required field
	invalidSkill := []byte(`---
name: invalid-skill
---
Missing description
`)

	if err := os.WriteFile(filepath.Join(tmpDir, "invalid.md"), invalidSkill, 0644); err != nil { //nolint:gosec // G306 - test file
		t.Fatalf("Failed to write invalid skill: %v", err)
	}

	// Create valid skill
	validSkill := []byte(`---
name: valid-skill
description: Valid skill
---
Content
`)

	if err := os.WriteFile(filepath.Join(tmpDir, "valid.md"), validSkill, 0644); err != nil { //nolint:gosec // G306 - test file
		t.Fatalf("Failed to write valid skill: %v", err)
	}

	loader := &Loader{}
	skills, err := loader.loadFromDir(tmpDir, "test")
	if err != nil {
		t.Fatalf("loadFromDir failed: %v", err)
	}

	// Should load valid skill and skip invalid one
	if len(skills) != 1 {
		t.Errorf("Expected 1 skill (invalid should be skipped), got %d", len(skills))
	}
	if skills[0].Name != "valid-skill" {
		t.Errorf("Loaded skill name = %q; want %q", skills[0].Name, "valid-skill")
	}
}

func TestLoaderLoadFromDir_NonMarkdownFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create non-markdown files
	if err := os.WriteFile(filepath.Join(tmpDir, "readme.txt"), []byte("text"), 0644); err != nil { //nolint:gosec // G306 - test file
		t.Fatalf("Failed to write txt file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "config.json"), []byte("{}"), 0644); err != nil { //nolint:gosec // G306 - test file
		t.Fatalf("Failed to write json file: %v", err)
	}

	// Create markdown skill
	skill := []byte(`---
name: test-skill
description: Test
---
Content
`)
	if err := os.WriteFile(filepath.Join(tmpDir, "test.md"), skill, 0644); err != nil { //nolint:gosec // G306 - test file
		t.Fatalf("Failed to write md file: %v", err)
	}

	loader := &Loader{}
	skills, err := loader.loadFromDir(tmpDir, "test")
	if err != nil {
		t.Fatalf("loadFromDir failed: %v", err)
	}

	// Should only load .md file
	if len(skills) != 1 {
		t.Errorf("Expected 1 skill (.md file only), got %d", len(skills))
	}
}

func TestLoaderLoadFromDir_Subdirectories(t *testing.T) {
	tmpDir := t.TempDir()
	subdir := filepath.Join(tmpDir, "subdir")

	if err := os.MkdirAll(subdir, 0755); err != nil { //nolint:gosec // G301 - test dir
		t.Fatalf("Failed to create subdir: %v", err)
	}

	// Create skill in root
	rootSkill := []byte(`---
name: root-skill
description: Root skill
---
Content
`)
	if err := os.WriteFile(filepath.Join(tmpDir, "root.md"), rootSkill, 0644); err != nil { //nolint:gosec // G306 - test file
		t.Fatalf("Failed to write root skill: %v", err)
	}

	// Create skill in subdir
	subSkill := []byte(`---
name: sub-skill
description: Sub skill
---
Content
`)
	if err := os.WriteFile(filepath.Join(subdir, "sub.md"), subSkill, 0644); err != nil { //nolint:gosec // G306 - test file
		t.Fatalf("Failed to write sub skill: %v", err)
	}

	loader := &Loader{}
	skills, err := loader.loadFromDir(tmpDir, "test")
	if err != nil {
		t.Fatalf("loadFromDir failed: %v", err)
	}

	// Should find skills in subdirectories too
	if len(skills) != 2 {
		t.Errorf("Expected 2 skills (root + subdir), got %d", len(skills))
	}
}

func TestLoaderLoadFromDir_NotDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	file := filepath.Join(tmpDir, "file.txt")

	if err := os.WriteFile(file, []byte("content"), 0644); err != nil { //nolint:gosec // G306 - test file
		t.Fatalf("Failed to write file: %v", err)
	}

	loader := &Loader{}
	_, err := loader.loadFromDir(file, "test")
	if err == nil {
		t.Fatal("Expected error when path is not a directory, got nil")
	}
}

func TestLoaderLoadByName_Found(t *testing.T) {
	tmpDir := t.TempDir()
	userDir := filepath.Join(tmpDir, "user")

	if err := os.MkdirAll(userDir, 0755); err != nil { //nolint:gosec // G301 - test dir
		t.Fatalf("Failed to create user dir: %v", err)
	}

	skill := []byte(`---
name: specific-skill
description: Specific skill
---
Content
`)
	if err := os.WriteFile(filepath.Join(userDir, "specific-skill.md"), skill, 0644); err != nil { //nolint:gosec // G306 - test file
		t.Fatalf("Failed to write skill: %v", err)
	}

	loader := &Loader{
		UserDir: userDir,
	}

	loaded, err := loader.LoadByName("specific-skill")
	if err != nil {
		t.Fatalf("LoadByName failed: %v", err)
	}

	if loaded.Name != "specific-skill" {
		t.Errorf("Name = %q; want %q", loaded.Name, "specific-skill")
	}
	if loaded.Source != "user" {
		t.Errorf("Source = %q; want %q", loaded.Source, "user")
	}
}

func TestLoaderLoadByName_NotFound(t *testing.T) {
	loader := &Loader{
		UserDir:    "/nonexistent/user",
		ProjectDir: "/nonexistent/project",
		BuiltinDir: "",
	}

	_, err := loader.LoadByName("nonexistent-skill")
	if err == nil {
		t.Fatal("Expected error for nonexistent skill, got nil")
	}
}

func TestLoaderLoadByName_ProjectOverridesUser(t *testing.T) {
	tmpDir := t.TempDir()
	userDir := filepath.Join(tmpDir, "user")
	projectDir := filepath.Join(tmpDir, "project")

	if err := os.MkdirAll(userDir, 0755); err != nil { //nolint:gosec // G301 - test dir
		t.Fatalf("Failed to create user dir: %v", err)
	}
	if err := os.MkdirAll(projectDir, 0755); err != nil { //nolint:gosec // G301 - test dir
		t.Fatalf("Failed to create project dir: %v", err)
	}

	userSkill := []byte(`---
name: shared
description: User version
---
User
`)
	projectSkill := []byte(`---
name: shared
description: Project version
---
Project
`)

	if err := os.WriteFile(filepath.Join(userDir, "shared.md"), userSkill, 0644); err != nil { //nolint:gosec // G306 - test file
		t.Fatalf("Failed to write user skill: %v", err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, "shared.md"), projectSkill, 0644); err != nil { //nolint:gosec // G306 - test file
		t.Fatalf("Failed to write project skill: %v", err)
	}

	loader := &Loader{
		UserDir:    userDir,
		ProjectDir: projectDir,
	}

	loaded, err := loader.LoadByName("shared")
	if err != nil {
		t.Fatalf("LoadByName failed: %v", err)
	}

	// Project should win
	if loaded.Description != "Project version" {
		t.Errorf("Description = %q; want %q (project should override)", loaded.Description, "Project version")
	}
	if loaded.Source != "project" {
		t.Errorf("Source = %q; want %q", loaded.Source, "project")
	}
}
