// ABOUTME: Tests for spell directory parser
// ABOUTME: Validates parsing of system.md, config.yaml, and tool overrides

package spells

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseSpellDirectory(t *testing.T) {
	// Create temp spell directory
	tmpDir := t.TempDir()
	spellDir := filepath.Join(tmpDir, "test-spell")
	if err := os.MkdirAll(spellDir, 0755); err != nil {
		t.Fatalf("Failed to create spell dir: %v", err)
	}

	// Create system.md
	systemMd := []byte(`---
name: test-spell
description: A test spell
author: test
version: 1.0.0
---

You are a test assistant. Be helpful and concise.
`)
	if err := os.WriteFile(filepath.Join(spellDir, "system.md"), systemMd, 0644); err != nil {
		t.Fatalf("Failed to write system.md: %v", err)
	}

	// Create config.yaml
	configYaml := []byte(`mode: replace

tools:
  enabled:
    - bash
    - read_file
  disabled:
    - web_search

reasoning:
  effort: medium
  show_thinking: false

response:
  max_tokens: 4096
  style: concise
`)
	if err := os.WriteFile(filepath.Join(spellDir, "config.yaml"), configYaml, 0644); err != nil {
		t.Fatalf("Failed to write config.yaml: %v", err)
	}

	// Parse
	spell, err := ParseSpellDirectory(spellDir)
	if err != nil {
		t.Fatalf("ParseSpellDirectory failed: %v", err)
	}

	// Verify metadata
	if spell.Name != "test-spell" {
		t.Errorf("Name = %q; want %q", spell.Name, "test-spell")
	}
	if spell.Description != "A test spell" {
		t.Errorf("Description = %q; want %q", spell.Description, "A test spell")
	}
	if spell.Author != "test" {
		t.Errorf("Author = %q; want %q", spell.Author, "test")
	}

	// Verify system prompt
	expectedPrompt := "\nYou are a test assistant. Be helpful and concise.\n"
	if spell.SystemPrompt != expectedPrompt {
		t.Errorf("SystemPrompt = %q; want %q", spell.SystemPrompt, expectedPrompt)
	}

	// Verify config
	if spell.Mode != LayerModeReplace {
		t.Errorf("Mode = %q; want %q", spell.Mode, LayerModeReplace)
	}
	if len(spell.Config.Tools.Enabled) != 2 {
		t.Errorf("Tools.Enabled length = %d; want 2", len(spell.Config.Tools.Enabled))
	}
	if spell.Config.Reasoning.Effort != ReasoningEffortMedium {
		t.Errorf("Reasoning.Effort = %q; want %q", spell.Config.Reasoning.Effort, ReasoningEffortMedium)
	}
}

func TestParseSpellDirectory_MinimalSpell(t *testing.T) {
	tmpDir := t.TempDir()
	spellDir := filepath.Join(tmpDir, "minimal")
	if err := os.MkdirAll(spellDir, 0755); err != nil {
		t.Fatalf("Failed to create spell dir: %v", err)
	}

	// Only system.md (config.yaml is optional)
	systemMd := []byte(`---
name: minimal
description: Minimal spell
---

Be minimal.
`)
	if err := os.WriteFile(filepath.Join(spellDir, "system.md"), systemMd, 0644); err != nil {
		t.Fatalf("Failed to write system.md: %v", err)
	}

	spell, err := ParseSpellDirectory(spellDir)
	if err != nil {
		t.Fatalf("ParseSpellDirectory failed: %v", err)
	}

	if spell.Name != "minimal" {
		t.Errorf("Name = %q; want %q", spell.Name, "minimal")
	}
	// Default mode should be layer
	if spell.Mode != LayerModeLayer {
		t.Errorf("Mode = %q; want %q (default)", spell.Mode, LayerModeLayer)
	}
}

func TestParseSpellDirectory_MissingSystemMd(t *testing.T) {
	tmpDir := t.TempDir()
	spellDir := filepath.Join(tmpDir, "no-system")
	if err := os.MkdirAll(spellDir, 0755); err != nil {
		t.Fatalf("Failed to create spell dir: %v", err)
	}

	_, err := ParseSpellDirectory(spellDir)
	if err == nil {
		t.Fatal("Expected error for missing system.md, got nil")
	}
}

func TestParseSpellDirectory_WithToolOverrides(t *testing.T) {
	tmpDir := t.TempDir()
	spellDir := filepath.Join(tmpDir, "with-tools")
	toolsDir := filepath.Join(spellDir, "tools")
	if err := os.MkdirAll(toolsDir, 0755); err != nil {
		t.Fatalf("Failed to create tools dir: %v", err)
	}

	systemMd := []byte(`---
name: with-tools
description: Spell with tool overrides
---

Content.
`)
	if err := os.WriteFile(filepath.Join(spellDir, "system.md"), systemMd, 0644); err != nil {
		t.Fatalf("Failed to write system.md: %v", err)
	}

	bashOverride := []byte(`defaults:
  timeout: 30000
restrictions:
  - no_sudo
  - no_rm_rf
`)
	if err := os.WriteFile(filepath.Join(toolsDir, "bash.yaml"), bashOverride, 0644); err != nil {
		t.Fatalf("Failed to write bash.yaml: %v", err)
	}

	spell, err := ParseSpellDirectory(spellDir)
	if err != nil {
		t.Fatalf("ParseSpellDirectory failed: %v", err)
	}

	if len(spell.ToolOverrides) != 1 {
		t.Fatalf("ToolOverrides length = %d; want 1", len(spell.ToolOverrides))
	}

	bashTool, ok := spell.ToolOverrides["bash"]
	if !ok {
		t.Fatal("Expected bash tool override")
	}
	if len(bashTool.Restrictions) != 2 {
		t.Errorf("Restrictions length = %d; want 2", len(bashTool.Restrictions))
	}
}

func TestParseSpellDirectory_NotADirectory(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "not-a-dir.txt")
	if err := os.WriteFile(filePath, []byte("hello"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	_, err := ParseSpellDirectory(filePath)
	if err == nil {
		t.Fatal("Expected error for non-directory path, got nil")
	}
}

func TestParseSpellDirectory_InvalidFrontmatter(t *testing.T) {
	tmpDir := t.TempDir()
	spellDir := filepath.Join(tmpDir, "invalid-fm")
	if err := os.MkdirAll(spellDir, 0755); err != nil {
		t.Fatalf("Failed to create spell dir: %v", err)
	}

	// Invalid YAML in frontmatter
	systemMd := []byte(`---
name: [invalid yaml
description: broken
---

Content.
`)
	if err := os.WriteFile(filepath.Join(spellDir, "system.md"), systemMd, 0644); err != nil {
		t.Fatalf("Failed to write system.md: %v", err)
	}

	_, err := ParseSpellDirectory(spellDir)
	if err == nil {
		t.Fatal("Expected error for invalid YAML frontmatter, got nil")
	}
}

func TestParseSpellDirectory_InvalidConfigYaml(t *testing.T) {
	tmpDir := t.TempDir()
	spellDir := filepath.Join(tmpDir, "invalid-config")
	if err := os.MkdirAll(spellDir, 0755); err != nil {
		t.Fatalf("Failed to create spell dir: %v", err)
	}

	systemMd := []byte(`---
name: test
description: test spell
---

Content.
`)
	if err := os.WriteFile(filepath.Join(spellDir, "system.md"), systemMd, 0644); err != nil {
		t.Fatalf("Failed to write system.md: %v", err)
	}

	// Invalid YAML in config
	configYaml := []byte(`mode: [invalid yaml`)
	if err := os.WriteFile(filepath.Join(spellDir, "config.yaml"), configYaml, 0644); err != nil {
		t.Fatalf("Failed to write config.yaml: %v", err)
	}

	_, err := ParseSpellDirectory(spellDir)
	if err == nil {
		t.Fatal("Expected error for invalid config.yaml, got nil")
	}
}

func TestParseSpellDirectory_MissingName(t *testing.T) {
	tmpDir := t.TempDir()
	spellDir := filepath.Join(tmpDir, "no-name")
	if err := os.MkdirAll(spellDir, 0755); err != nil {
		t.Fatalf("Failed to create spell dir: %v", err)
	}

	// Missing name field
	systemMd := []byte(`---
description: spell without name
---

Content.
`)
	if err := os.WriteFile(filepath.Join(spellDir, "system.md"), systemMd, 0644); err != nil {
		t.Fatalf("Failed to write system.md: %v", err)
	}

	_, err := ParseSpellDirectory(spellDir)
	if err == nil {
		t.Fatal("Expected validation error for missing name, got nil")
	}
}

func TestParseSpellDirectory_YmlExtension(t *testing.T) {
	tmpDir := t.TempDir()
	spellDir := filepath.Join(tmpDir, "with-yml-tools")
	toolsDir := filepath.Join(spellDir, "tools")
	if err := os.MkdirAll(toolsDir, 0755); err != nil {
		t.Fatalf("Failed to create tools dir: %v", err)
	}

	systemMd := []byte(`---
name: yml-tools
description: Spell with .yml tool overrides
---

Content.
`)
	if err := os.WriteFile(filepath.Join(spellDir, "system.md"), systemMd, 0644); err != nil {
		t.Fatalf("Failed to write system.md: %v", err)
	}

	// Use .yml extension instead of .yaml
	readOverride := []byte(`defaults:
  max_lines: 1000
restrictions:
  - no_binary
`)
	if err := os.WriteFile(filepath.Join(toolsDir, "read_file.yml"), readOverride, 0644); err != nil {
		t.Fatalf("Failed to write read_file.yml: %v", err)
	}

	spell, err := ParseSpellDirectory(spellDir)
	if err != nil {
		t.Fatalf("ParseSpellDirectory failed: %v", err)
	}

	if len(spell.ToolOverrides) != 1 {
		t.Fatalf("ToolOverrides length = %d; want 1", len(spell.ToolOverrides))
	}

	readTool, ok := spell.ToolOverrides["read_file"]
	if !ok {
		t.Fatal("Expected read_file tool override")
	}
	if len(readTool.Restrictions) != 1 {
		t.Errorf("Restrictions length = %d; want 1", len(readTool.Restrictions))
	}
}
