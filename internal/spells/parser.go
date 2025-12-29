// ABOUTME: Spell directory parser for loading spell configuration
// ABOUTME: Parses system.md, config.yaml, and tools/*.yaml files

package spells

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/2389-research/hex/internal/frontmatter"
	"gopkg.in/yaml.v3"
)

// ParseSpellDirectory loads a spell from a directory
func ParseSpellDirectory(dir string) (*Spell, error) {
	// Check directory exists
	info, err := os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("stat spell directory: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("spell path is not a directory: %s", dir)
	}

	// Parse system.md (required)
	systemPath := filepath.Join(dir, "system.md")
	spell, err := parseSystemMd(systemPath)
	if err != nil {
		return nil, fmt.Errorf("parse system.md: %w", err)
	}
	spell.FilePath = dir

	// Parse config.yaml (optional)
	configPath := filepath.Join(dir, "config.yaml")
	if _, err := os.Stat(configPath); err == nil {
		config, configErr := parseConfigYaml(configPath)
		if configErr != nil {
			return nil, fmt.Errorf("parse config.yaml: %w", configErr)
		}
		spell.Config = *config
		spell.Mode = config.Mode
	}

	// Set default mode if not specified
	if spell.Mode == "" {
		spell.Mode = LayerModeLayer
	}

	// Parse tool overrides (optional)
	toolsDir := filepath.Join(dir, "tools")
	if info, err := os.Stat(toolsDir); err == nil && info.IsDir() {
		overrides, toolErr := parseToolOverrides(toolsDir)
		if toolErr != nil {
			return nil, fmt.Errorf("parse tool overrides: %w", toolErr)
		}
		spell.ToolOverrides = overrides
	}

	// Validate
	if err := spell.Validate(); err != nil {
		return nil, fmt.Errorf("invalid spell: %w", err)
	}

	return spell, nil
}

// parseSystemMd parses the system.md file (YAML frontmatter + markdown body)
func parseSystemMd(path string) (*Spell, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read system.md: %w", err)
	}

	// Split frontmatter and content
	fm, content, err := frontmatter.Split(data)
	if err != nil {
		return nil, fmt.Errorf("split frontmatter: %w", err)
	}

	// Parse YAML frontmatter
	var spell Spell
	if len(fm) > 0 {
		if err := yaml.Unmarshal(fm, &spell); err != nil {
			return nil, fmt.Errorf("parse YAML frontmatter: %w", err)
		}
	}

	// Store system prompt content
	spell.SystemPrompt = string(content)

	return &spell, nil
}

// parseConfigYaml parses the config.yaml file
func parseConfigYaml(path string) (*SpellConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config.yaml: %w", err)
	}

	var config SpellConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parse config YAML: %w", err)
	}

	return &config, nil
}

// parseToolOverrides parses all .yaml files in the tools/ directory
func parseToolOverrides(dir string) (map[string]ToolOverride, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read tools directory: %w", err)
	}

	overrides := make(map[string]ToolOverride)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".yaml") && !strings.HasSuffix(name, ".yml") {
			continue
		}

		path := filepath.Join(dir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read tool override %s: %w", name, err)
		}

		var override ToolOverride
		if err := yaml.Unmarshal(data, &override); err != nil {
			return nil, fmt.Errorf("parse tool override %s: %w", name, err)
		}

		// Use filename (without extension) as tool name
		toolName := strings.TrimSuffix(name, filepath.Ext(name))
		overrides[toolName] = override
	}

	return overrides, nil
}

// ParseSpellFromFS loads a spell from an embedded filesystem
func ParseSpellFromFS(fsys fs.FS, spellName string) (*Spell, error) {
	// Parse system.md (required)
	systemPath := spellName + "/system.md"
	spell, err := parseSystemMdFromFS(fsys, systemPath)
	if err != nil {
		return nil, fmt.Errorf("parse system.md: %w", err)
	}
	spell.FilePath = spellName

	// Parse config.yaml (optional)
	configPath := spellName + "/config.yaml"
	if data, err := fs.ReadFile(fsys, configPath); err == nil {
		config, configErr := parseConfigYamlData(data)
		if configErr != nil {
			return nil, fmt.Errorf("parse config.yaml: %w", configErr)
		}
		spell.Config = *config
		spell.Mode = config.Mode
	}

	// Set default mode if not specified
	if spell.Mode == "" {
		spell.Mode = LayerModeLayer
	}

	// Parse tool overrides (optional)
	toolsDir := spellName + "/tools"
	if entries, err := fs.ReadDir(fsys, toolsDir); err == nil {
		overrides, toolErr := parseToolOverridesFromFS(fsys, toolsDir, entries)
		if toolErr != nil {
			return nil, fmt.Errorf("parse tool overrides: %w", toolErr)
		}
		spell.ToolOverrides = overrides
	}

	// Validate
	if err := spell.Validate(); err != nil {
		return nil, fmt.Errorf("invalid spell: %w", err)
	}

	return spell, nil
}

// parseSystemMdFromFS parses system.md from an embedded filesystem
func parseSystemMdFromFS(fsys fs.FS, path string) (*Spell, error) {
	data, err := fs.ReadFile(fsys, path)
	if err != nil {
		return nil, fmt.Errorf("read system.md: %w", err)
	}

	// Split frontmatter and content
	fm, content, err := frontmatter.Split(data)
	if err != nil {
		return nil, fmt.Errorf("split frontmatter: %w", err)
	}

	// Parse YAML frontmatter
	var spell Spell
	if len(fm) > 0 {
		if err := yaml.Unmarshal(fm, &spell); err != nil {
			return nil, fmt.Errorf("parse YAML frontmatter: %w", err)
		}
	}

	// Store system prompt content
	spell.SystemPrompt = string(content)

	return &spell, nil
}

// parseConfigYamlData parses config YAML from raw bytes
func parseConfigYamlData(data []byte) (*SpellConfig, error) {
	var config SpellConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parse config YAML: %w", err)
	}
	return &config, nil
}

// parseToolOverridesFromFS parses tool overrides from an embedded filesystem
func parseToolOverridesFromFS(fsys fs.FS, dir string, entries []fs.DirEntry) (map[string]ToolOverride, error) {
	overrides := make(map[string]ToolOverride)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".yaml") && !strings.HasSuffix(name, ".yml") {
			continue
		}

		path := dir + "/" + name
		data, err := fs.ReadFile(fsys, path)
		if err != nil {
			return nil, fmt.Errorf("read tool override %s: %w", name, err)
		}

		var override ToolOverride
		if err := yaml.Unmarshal(data, &override); err != nil {
			return nil, fmt.Errorf("parse tool override %s: %w", name, err)
		}

		// Use filename (without extension) as tool name
		toolName := strings.TrimSuffix(name, filepath.Ext(name))
		overrides[toolName] = override
	}

	return overrides, nil
}
