// ABOUTME: Template loader for reading and parsing YAML session templates
// ABOUTME: Handles loading from ~/.clem/templates/ directory with validation

// Package templates provides template loading and rendering for conversations.
package templates

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// LoadTemplate loads a single template from a file path
func LoadTemplate(path string) (*Template, error) {
	// Read file
	data, err := os.ReadFile(path) //nolint:gosec // G304: Loading config/template files from validated paths
	if err != nil {
		return nil, fmt.Errorf("read template file: %w", err)
	}

	// Parse YAML
	var template Template
	if err := yaml.Unmarshal(data, &template); err != nil {
		return nil, fmt.Errorf("parse template YAML: %w", err)
	}

	// Validate
	if err := template.Validate(); err != nil {
		return nil, fmt.Errorf("invalid template: %w", err)
	}

	return &template, nil
}

// LoadTemplates loads all templates from a directory
func LoadTemplates(dir string) (map[string]*Template, error) {
	// Expand ~ if present
	if strings.HasPrefix(dir, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("get user home directory: %w", err)
		}
		dir = filepath.Join(home, dir[2:])
	}

	// Check if directory exists
	info, err := os.Stat(dir)
	if os.IsNotExist(err) {
		// Directory doesn't exist - return empty map (not an error)
		return make(map[string]*Template), nil
	}
	if err != nil {
		return nil, fmt.Errorf("stat templates directory: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("templates path is not a directory: %s", dir)
	}

	// Read directory
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read templates directory: %w", err)
	}

	// Load each YAML file
	templates := make(map[string]*Template)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Only load .yaml and .yml files
		name := entry.Name()
		if !strings.HasSuffix(name, ".yaml") && !strings.HasSuffix(name, ".yml") {
			continue
		}

		// Load template
		path := filepath.Join(dir, name)
		template, err := LoadTemplate(path)
		if err != nil {
			// Log error but continue loading other templates
			_, _ = fmt.Fprintf(os.Stderr, "Warning: Failed to load template %s: %v\n", name, err)
			continue
		}

		// Use filename (without extension) as key if name not set
		key := template.Name
		if key == "" {
			key = strings.TrimSuffix(name, filepath.Ext(name))
		}

		templates[key] = template
	}

	return templates, nil
}

// GetTemplatesDir returns the default templates directory path
func GetTemplatesDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get user home directory: %w", err)
	}
	return filepath.Join(home, ".clem", "templates"), nil
}

// EnsureTemplatesDir creates the templates directory if it doesn't exist
func EnsureTemplatesDir() (string, error) {
	dir, err := GetTemplatesDir()
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(dir, 0750); err != nil {
		return "", fmt.Errorf("create templates directory: %w", err)
	}

	return dir, nil
}
