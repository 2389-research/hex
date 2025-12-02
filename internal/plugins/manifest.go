// ABOUTME: Plugin manifest schema and validation for Clem plugins
// ABOUTME: Defines plugin.json structure with validation and parsing logic

package plugins

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Manifest represents a plugin's metadata and configuration
type Manifest struct {
	// Required fields
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`

	// Optional metadata
	Author     string            `json:"author,omitempty"`
	License    string            `json:"license,omitempty"`
	Homepage   string            `json:"homepage,omitempty"`
	Repository *RepositoryInfo   `json:"repository,omitempty"`
	Keywords   []string          `json:"keywords,omitempty"`
	Engines    map[string]string `json:"engines,omitempty"`

	// Plugin dependencies
	Dependencies map[string]string `json:"dependencies,omitempty"`

	// What the plugin contributes
	Skills     []string               `json:"skills,omitempty"`
	Commands   []string               `json:"commands,omitempty"`
	Hooks      map[string]interface{} `json:"hooks,omitempty"`
	MCPServers map[string]MCPConfig   `json:"mcpServers,omitempty"`
	Agents     []string               `json:"agents,omitempty"`
	Templates  []string               `json:"templates,omitempty"`
	Scripts    map[string]string      `json:"scripts,omitempty"`
	Activation *ActivationConfig      `json:"activation,omitempty"`
	Config     *ConfigSchema          `json:"configuration,omitempty"`
}

// RepositoryInfo contains source repository metadata
type RepositoryInfo struct {
	Type string `json:"type"` // git, svn, etc.
	URL  string `json:"url"`
}

// MCPConfig defines an MCP server configuration
type MCPConfig struct {
	Command string            `json:"command"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
}

// ActivationConfig defines when a plugin should activate
type ActivationConfig struct {
	OnStartup bool     `json:"onStartup,omitempty"`
	Languages []string `json:"languages,omitempty"`
	Files     []string `json:"files,omitempty"`
	Projects  []string `json:"projects,omitempty"`
}

// ConfigSchema defines plugin configuration options
type ConfigSchema struct {
	Defaults map[string]interface{} `json:"defaults,omitempty"`
	Schema   map[string]interface{} `json:"schema,omitempty"`
}

// LoadManifest loads and validates a plugin manifest from a file
func LoadManifest(path string) (*Manifest, error) {
	//nolint:gosec // G304 - reading plugin manifest from validated plugin directory
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read manifest: %w", err)
	}

	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("parse manifest: %w", err)
	}

	if err := manifest.Validate(); err != nil {
		return nil, fmt.Errorf("validate manifest: %w", err)
	}

	return &manifest, nil
}

// Validate checks if the manifest has all required fields and valid values
func (m *Manifest) Validate() error {
	// Required fields
	if m.Name == "" {
		return fmt.Errorf("name is required")
	}
	if !isValidPluginName(m.Name) {
		return fmt.Errorf("invalid plugin name %q: must contain only lowercase letters, numbers, and hyphens", m.Name)
	}
	if m.Version == "" {
		return fmt.Errorf("version is required")
	}
	if !isValidSemver(m.Version) {
		return fmt.Errorf("invalid version %q: must be valid semver (e.g., 1.0.0)", m.Version)
	}
	if m.Description == "" {
		return fmt.Errorf("description is required")
	}

	// Validate file paths exist (relative to manifest directory)
	// Note: We can't validate these at parse time since we don't know the plugin root
	// This validation should happen during installation/loading

	return nil
}

// isValidPluginName checks if a plugin name follows naming conventions
func isValidPluginName(name string) bool {
	// Plugin names must be lowercase alphanumeric with hyphens
	// Must start with a letter
	match, _ := regexp.MatchString(`^[a-z][a-z0-9-]*$`, name)
	return match
}

// isValidSemver checks if a version string is valid semantic versioning
func isValidSemver(version string) bool {
	// Simple semver validation: major.minor.patch with optional pre-release
	match, _ := regexp.MatchString(`^v?\d+\.\d+\.\d+(-[a-zA-Z0-9.-]+)?$`, version)
	return match
}

// GetSkillPaths returns full paths to skill files
func (m *Manifest) GetSkillPaths(pluginDir string) []string {
	paths := make([]string, 0, len(m.Skills))
	for _, skill := range m.Skills {
		paths = append(paths, filepath.Join(pluginDir, skill))
	}
	return paths
}

// GetCommandPaths returns full paths to command files
func (m *Manifest) GetCommandPaths(pluginDir string) []string {
	paths := make([]string, 0, len(m.Commands))
	for _, cmd := range m.Commands {
		paths = append(paths, filepath.Join(pluginDir, cmd))
	}
	return paths
}

// GetAgentPaths returns full paths to agent files
func (m *Manifest) GetAgentPaths(pluginDir string) []string {
	paths := make([]string, 0, len(m.Agents))
	for _, agent := range m.Agents {
		paths = append(paths, filepath.Join(pluginDir, agent))
	}
	return paths
}

// GetTemplatePaths returns full paths to template files
func (m *Manifest) GetTemplatePaths(pluginDir string) []string {
	paths := make([]string, 0, len(m.Templates))
	for _, tmpl := range m.Templates {
		paths = append(paths, filepath.Join(pluginDir, tmpl))
	}
	return paths
}

// FullID returns the full plugin identifier (name@version)
func (m *Manifest) FullID() string {
	return fmt.Sprintf("%s@%s", m.Name, m.Version)
}

// ShouldActivate checks if the plugin should activate in the current context
func (m *Manifest) ShouldActivate(context *ActivationContext) bool {
	if m.Activation == nil {
		// No activation rules = always active
		return true
	}

	// OnStartup plugins always activate
	if m.Activation.OnStartup {
		return true
	}

	// Check language match
	if len(m.Activation.Languages) > 0 && context != nil {
		for _, lang := range m.Activation.Languages {
			if contains(context.Languages, lang) {
				return true
			}
		}
	}

	// Check file patterns
	if len(m.Activation.Files) > 0 && context != nil {
		for _, pattern := range m.Activation.Files {
			for _, file := range context.Files {
				matched, _ := filepath.Match(pattern, filepath.Base(file))
				if matched {
					return true
				}
			}
		}
	}

	// Check project types
	if len(m.Activation.Projects) > 0 && context != nil {
		for _, proj := range m.Activation.Projects {
			if contains(context.ProjectTypes, proj) {
				return true
			}
		}
	}

	// If we have activation rules but none matched, don't activate
	if len(m.Activation.Languages) > 0 || len(m.Activation.Files) > 0 || len(m.Activation.Projects) > 0 {
		return false
	}

	// No specific rules = activate
	return true
}

// ActivationContext provides context for plugin activation decisions
type ActivationContext struct {
	Languages    []string // Detected languages in project
	Files        []string // Important files in project
	ProjectTypes []string // Detected project types (e.g., "react", "django")
}

// contains checks if a slice contains a string (case-insensitive)
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if strings.EqualFold(s, item) {
			return true
		}
	}
	return false
}

// Save writes a manifest to disk
func (m *Manifest) Save(path string) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal manifest: %w", err)
	}

	//nolint:gosec // G306 - 0644 is correct for plugin manifest files
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write manifest: %w", err)
	}

	return nil
}
