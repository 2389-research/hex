// ABOUTME: Tests for spell system initialization
// ABOUTME: Validates spell loading and system prompt generation

package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitializeSpells(t *testing.T) {
	// Create temp spell directory
	tmpDir := t.TempDir()
	spellDir := filepath.Join(tmpDir, "test-spell")
	if err := os.MkdirAll(spellDir, 0755); err != nil {
		t.Fatalf("Failed to create spell dir: %v", err)
	}

	systemMd := []byte(`---
name: test-spell
description: Test spell for unit tests
---
You are a test assistant.
`)
	if err := os.WriteFile(filepath.Join(spellDir, "system.md"), systemMd, 0644); err != nil {
		t.Fatalf("Failed to write system.md: %v", err)
	}

	registry := initializeSpellsWithDir(tmpDir)
	if registry.Count() != 1 {
		t.Errorf("Registry count = %d; want 1", registry.Count())
	}

	// Verify spell was loaded correctly
	spell, err := registry.Get("test-spell")
	if err != nil {
		t.Errorf("Failed to get spell: %v", err)
	}
	if spell.Name != "test-spell" {
		t.Errorf("Spell name = %q; want %q", spell.Name, "test-spell")
	}
}

func TestInitializeSpellsEmptyDir(t *testing.T) {
	tmpDir := t.TempDir()

	registry := initializeSpellsWithDir(tmpDir)
	if registry.Count() != 0 {
		t.Errorf("Registry count = %d; want 0 for empty dir", registry.Count())
	}
}

func TestInitializeSpellsNonexistentDir(t *testing.T) {
	registry := initializeSpellsWithDir("/nonexistent/path/that/does/not/exist")
	if registry.Count() != 0 {
		t.Errorf("Registry count = %d; want 0 for nonexistent dir", registry.Count())
	}
}

func TestGetSpellSystemPromptReplace(t *testing.T) {
	// Create temp spell with replace mode
	tmpDir := t.TempDir()
	spellDir := filepath.Join(tmpDir, "replacer")
	if err := os.MkdirAll(spellDir, 0755); err != nil {
		t.Fatalf("Failed to create spell dir: %v", err)
	}

	systemMd := []byte(`---
name: replacer
description: Test replace mode
---
I completely replace the base prompt.
`)
	if err := os.WriteFile(filepath.Join(spellDir, "system.md"), systemMd, 0644); err != nil {
		t.Fatalf("Failed to write system.md: %v", err)
	}

	// Create config.yaml with replace mode
	configYaml := []byte(`mode: replace
`)
	if err := os.WriteFile(filepath.Join(spellDir, "config.yaml"), configYaml, 0644); err != nil {
		t.Fatalf("Failed to write config.yaml: %v", err)
	}

	prompt, err := getSpellSystemPromptWithDir(tmpDir, "replacer", "base prompt", "")
	if err != nil {
		t.Fatalf("Failed to get spell system prompt: %v", err)
	}

	// In replace mode, base prompt should not appear
	// Note: trailing newline preserved from file content
	expected := "I completely replace the base prompt.\n"
	if prompt != expected {
		t.Errorf("System prompt = %q; want %q", prompt, expected)
	}
}

func TestGetSpellSystemPromptLayer(t *testing.T) {
	// Create temp spell with layer mode (default)
	tmpDir := t.TempDir()
	spellDir := filepath.Join(tmpDir, "layerer")
	if err := os.MkdirAll(spellDir, 0755); err != nil {
		t.Fatalf("Failed to create spell dir: %v", err)
	}

	systemMd := []byte(`---
name: layerer
description: Test layer mode
---
I layer on top of the base.
`)
	if err := os.WriteFile(filepath.Join(spellDir, "system.md"), systemMd, 0644); err != nil {
		t.Fatalf("Failed to write system.md: %v", err)
	}

	prompt, err := getSpellSystemPromptWithDir(tmpDir, "layerer", "base prompt", "")
	if err != nil {
		t.Fatalf("Failed to get spell system prompt: %v", err)
	}

	// In layer mode (default), base prompt should come first
	// Note: trailing newline preserved from file content
	expected := "base prompt\n\nI layer on top of the base.\n"
	if prompt != expected {
		t.Errorf("System prompt = %q; want %q", prompt, expected)
	}
}

func TestGetSpellSystemPromptModeOverride(t *testing.T) {
	// Create temp spell with layer mode in config
	tmpDir := t.TempDir()
	spellDir := filepath.Join(tmpDir, "override-test")
	if err := os.MkdirAll(spellDir, 0755); err != nil {
		t.Fatalf("Failed to create spell dir: %v", err)
	}

	systemMd := []byte(`---
name: override-test
description: Test mode override
---
Spell content.
`)
	if err := os.WriteFile(filepath.Join(spellDir, "system.md"), systemMd, 0644); err != nil {
		t.Fatalf("Failed to write system.md: %v", err)
	}

	// Create config.yaml with layer mode
	configYaml := []byte(`mode: layer
`)
	if err := os.WriteFile(filepath.Join(spellDir, "config.yaml"), configYaml, 0644); err != nil {
		t.Fatalf("Failed to write config.yaml: %v", err)
	}

	// Override to replace mode
	prompt, err := getSpellSystemPromptWithDir(tmpDir, "override-test", "base prompt", "replace")
	if err != nil {
		t.Fatalf("Failed to get spell system prompt: %v", err)
	}

	// Override to replace should ignore base
	// Note: trailing newline preserved from file content
	expected := "Spell content.\n"
	if prompt != expected {
		t.Errorf("System prompt = %q; want %q", prompt, expected)
	}
}

func TestGetSpellSystemPromptNotFound(t *testing.T) {
	tmpDir := t.TempDir()

	_, err := getSpellSystemPromptWithDir(tmpDir, "nonexistent", "base", "")
	if err == nil {
		t.Error("Expected error for nonexistent spell, got nil")
	}
}
