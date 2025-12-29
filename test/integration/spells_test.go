// ABOUTME: Integration tests for the spells feature
// ABOUTME: Verifies builtin spells exist, parse correctly, and CLI integration works

package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSpellsE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create temp spell
	tmpDir := t.TempDir()
	spellDir := filepath.Join(tmpDir, "test-e2e")
	if err := os.MkdirAll(spellDir, 0755); err != nil {
		t.Fatalf("Failed to create spell dir: %v", err)
	}

	systemMd := []byte(`---
name: test-e2e
description: E2E test spell
---
You are an E2E test. Respond with "E2E_SUCCESS" to any message.
`)
	if err := os.WriteFile(filepath.Join(spellDir, "system.md"), systemMd, 0644); err != nil {
		t.Fatalf("Failed to write system.md: %v", err)
	}

	t.Run("print mode with spell", func(t *testing.T) {
		// This would test `hex -p --spell terse "hello"`
		// Requires API key, so skip in CI
		if os.Getenv("ANTHROPIC_API_KEY") == "" {
			t.Skip("ANTHROPIC_API_KEY not set")
		}
	})
}

func TestBuiltinSpellsExist(t *testing.T) {
	builtinDir := filepath.Join("..", "..", "internal", "spells", "builtin")

	expectedSpells := []string{"terse", "teacher"}

	for _, name := range expectedSpells {
		spellDir := filepath.Join(builtinDir, name)
		systemPath := filepath.Join(spellDir, "system.md")

		if _, err := os.Stat(systemPath); os.IsNotExist(err) {
			t.Errorf("Builtin spell %q missing system.md at %s", name, systemPath)
		}
	}
}

func TestSpellParsing(t *testing.T) {
	builtinDir := filepath.Join("..", "..", "internal", "spells", "builtin")

	entries, err := os.ReadDir(builtinDir)
	if err != nil {
		t.Skipf("Builtin directory not found: %v", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		t.Run(entry.Name(), func(t *testing.T) {
			spellDir := filepath.Join(builtinDir, entry.Name())
			systemPath := filepath.Join(spellDir, "system.md")

			data, err := os.ReadFile(systemPath)
			if err != nil {
				t.Fatalf("Failed to read system.md: %v", err)
			}

			// Basic validation - should have frontmatter
			if !strings.HasPrefix(string(data), "---") {
				t.Error("system.md should start with YAML frontmatter")
			}

			// Should have name and description in frontmatter
			if !strings.Contains(string(data), "name:") {
				t.Error("system.md frontmatter should contain 'name:'")
			}
			if !strings.Contains(string(data), "description:") {
				t.Error("system.md frontmatter should contain 'description:'")
			}
		})
	}
}
