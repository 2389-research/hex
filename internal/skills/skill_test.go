package skills

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/harper/hex/internal/frontmatter"
)

func TestParseBytes_ValidSkill(t *testing.T) {
	data := []byte(`---
name: test-skill
description: A test skill
tags:
  - testing
  - example
activationPatterns:
  - "test.*pattern"
priority: 7
version: 1.0.0
---

# Test Skill Content

This is the markdown content.
`)

	skill, err := ParseBytes("test.md", data)
	if err != nil {
		t.Fatalf("ParseBytes failed: %v", err)
	}

	if skill.Name != "test-skill" {
		t.Errorf("Name = %q; want %q", skill.Name, "test-skill")
	}
	if skill.Description != "A test skill" {
		t.Errorf("Description = %q; want %q", skill.Description, "A test skill")
	}
	if len(skill.Tags) != 2 {
		t.Errorf("Tags length = %d; want 2", len(skill.Tags))
	}
	if skill.Priority != 7 {
		t.Errorf("Priority = %d; want 7", skill.Priority)
	}
	if skill.Version != "1.0.0" {
		t.Errorf("Version = %q; want %q", skill.Version, "1.0.0")
	}
	if skill.Content == "" {
		t.Error("Content is empty, expected markdown content")
	}
}

func TestParseBytes_NoFrontmatter(t *testing.T) {
	data := []byte(`# Just Markdown

No YAML frontmatter here.
`)

	_, err := ParseBytes("test.md", data)
	if err == nil {
		t.Fatal("Expected error for missing required fields, got nil")
	}
}

func TestParseBytes_MissingName(t *testing.T) {
	data := []byte(`---
description: Missing name field
---

Content
`)

	_, err := ParseBytes("test.md", data)
	if err == nil {
		t.Fatal("Expected error for missing name, got nil")
	}
	if err.Error() != "skill missing required 'name' field" {
		t.Errorf("Error = %v; want 'skill missing required 'name' field'", err)
	}
}

func TestParseBytes_MissingDescription(t *testing.T) {
	data := []byte(`---
name: test-skill
---

Content
`)

	_, err := ParseBytes("test.md", data)
	if err == nil {
		t.Fatal("Expected error for missing description, got nil")
	}
	if err.Error() != "skill missing required 'description' field" {
		t.Errorf("Error = %v; want 'skill missing required 'description' field'", err)
	}
}

func TestParseBytes_DefaultPriority(t *testing.T) {
	data := []byte(`---
name: test-skill
description: Test description
---

Content
`)

	skill, err := ParseBytes("test.md", data)
	if err != nil {
		t.Fatalf("ParseBytes failed: %v", err)
	}

	if skill.Priority != 5 {
		t.Errorf("Priority = %d; want default 5", skill.Priority)
	}
}

func TestParseBytes_UnclosedFrontmatter(t *testing.T) {
	data := []byte(`---
name: test-skill
description: Test

Missing closing ---
`)

	_, err := ParseBytes("test.md", data)
	if err == nil {
		t.Fatal("Expected error for unclosed frontmatter, got nil")
	}
}

func TestParse_FileNotFound(t *testing.T) {
	_, err := Parse("/nonexistent/file.md")
	if err == nil {
		t.Fatal("Expected error for nonexistent file, got nil")
	}
}

func TestParse_ValidFile(t *testing.T) {
	// Create temporary file
	tmpDir := t.TempDir()
	skillFile := filepath.Join(tmpDir, "test-skill.md")

	content := []byte(`---
name: file-test-skill
description: Skill from file
---

File content
`)

	if err := os.WriteFile(skillFile, content, 0644); err != nil { //nolint:gosec // G306 - test file
		t.Fatalf("Failed to write test file: %v", err)
	}

	skill, err := Parse(skillFile)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if skill.Name != "file-test-skill" {
		t.Errorf("Name = %q; want %q", skill.Name, "file-test-skill")
	}
	if skill.FilePath != skillFile {
		t.Errorf("FilePath = %q; want %q", skill.FilePath, skillFile)
	}
}

func TestSkillMatchesPattern(t *testing.T) {
	skill := &Skill{
		Name:        "test-skill",
		Description: "Test",
		ActivationPatterns: []string{
			"write.*test",
			"debug",
			"fix.*bug",
		},
	}

	tests := []struct {
		message string
		matches bool
	}{
		{"write a test for this", true},
		{"Write Tests for feature", true},
		{"debug this issue", true},
		{"DEBUG MODE", true},
		{"fix the bug", true},
		{"Fix Bug #123", true},
		{"implement feature", false},
		{"random message", false},
	}

	for _, tt := range tests {
		t.Run(tt.message, func(t *testing.T) {
			result := skill.MatchesPattern(tt.message)
			if result != tt.matches {
				t.Errorf("MatchesPattern(%q) = %v; want %v", tt.message, result, tt.matches)
			}
		})
	}
}

func TestSkillMatchesPattern_NoPatterns(t *testing.T) {
	skill := &Skill{
		Name:               "test-skill",
		Description:        "Test",
		ActivationPatterns: []string{},
	}

	if skill.MatchesPattern("any message") {
		t.Error("Expected no match when no patterns defined")
	}
}

func TestSkillMatchesPattern_InvalidRegex(t *testing.T) {
	skill := &Skill{
		Name:        "test-skill",
		Description: "Test",
		ActivationPatterns: []string{
			"[invalid",
			"valid.*pattern",
		},
	}

	// Should not panic on invalid regex, just skip it
	if !skill.MatchesPattern("valid test pattern") {
		t.Error("Expected match on valid pattern despite invalid regex in list")
	}
}

func TestSkillString(t *testing.T) {
	skill := &Skill{
		Name:     "test-skill",
		Source:   "user",
		Priority: 8,
	}

	str := skill.String()
	expected := "Skill{name=test-skill, source=user, priority=8}"
	if str != expected {
		t.Errorf("String() = %q; want %q", str, expected)
	}
}

func TestSplitFrontmatter_WindowsLineEndings(t *testing.T) {
	data := []byte("---\r\nname: test\r\ndescription: Test\r\n---\r\n\r\nContent\r\n")

	fm, content, err := frontmatter.Split(data)
	if err != nil {
		t.Fatalf("frontmatter.Split failed: %v", err)
	}

	if len(fm) == 0 {
		t.Error("Frontmatter is empty")
	}
	if len(content) == 0 {
		t.Error("Content is empty")
	}
}
