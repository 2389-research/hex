// ABOUTME: Plugin discovery and loading system for Clem
// ABOUTME: Scans directories for plugins and loads their manifests and components

package plugins

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// Loader discovers and loads plugins from directories
type Loader struct {
	pluginsDir string
	state      *State
	stateFile  string
}

// NewLoader creates a new plugin loader
func NewLoader(pluginsDir, stateFile string) (*Loader, error) {
	// Ensure plugins directory exists
	//nolint:gosec // G301 - 0755 is correct for plugin directories
	if err := os.MkdirAll(pluginsDir, 0755); err != nil {
		return nil, fmt.Errorf("create plugins directory: %w", err)
	}

	// Load state
	state, err := LoadState(stateFile)
	if err != nil {
		return nil, fmt.Errorf("load state: %w", err)
	}

	return &Loader{
		pluginsDir: pluginsDir,
		state:      state,
		stateFile:  stateFile,
	}, nil
}

// Plugin represents a loaded plugin with its manifest and metadata
type Plugin struct {
	Name      string
	Version   string
	Dir       string
	Manifest  *Manifest
	Enabled   bool
	Installed bool
}

// DiscoverAll finds all plugins in the plugins directory
func (l *Loader) DiscoverAll() ([]*Plugin, error) {
	entries, err := os.ReadDir(l.pluginsDir)
	if err != nil {
		return nil, fmt.Errorf("read plugins directory: %w", err)
	}

	//nolint:prealloc // Size unknown until we filter directories with manifests
	var plugins []*Plugin
	var errors []error

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		pluginDir := filepath.Join(l.pluginsDir, entry.Name())
		manifestPath := filepath.Join(pluginDir, "plugin.json")

		// Skip if no manifest
		if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
			continue
		}

		// Load manifest
		manifest, err := LoadManifest(manifestPath)
		if err != nil {
			errors = append(errors, fmt.Errorf("load plugin %s: %w", entry.Name(), err))
			continue
		}

		// Check if enabled in state
		enabled := l.state.IsEnabled(manifest.Name)
		installed := l.state.IsInstalled(manifest.Name)

		plugin := &Plugin{
			Name:      manifest.Name,
			Version:   manifest.Version,
			Dir:       pluginDir,
			Manifest:  manifest,
			Enabled:   enabled,
			Installed: installed,
		}

		plugins = append(plugins, plugin)
	}

	// Sort by name
	sort.Slice(plugins, func(i, j int) bool {
		return plugins[i].Name < plugins[j].Name
	})

	// Return plugins even if some had errors
	if len(errors) > 0 {
		// Log errors but don't fail completely
		return plugins, fmt.Errorf("encountered %d errors during discovery (loaded %d plugins)", len(errors), len(plugins))
	}

	return plugins, nil
}

// LoadEnabled loads all enabled plugins
func (l *Loader) LoadEnabled(context *ActivationContext) ([]*Plugin, error) {
	allPlugins, err := l.DiscoverAll()
	if err != nil {
		// Still return partial results
		return filterEnabled(allPlugins, context), err
	}

	return filterEnabled(allPlugins, context), nil
}

// filterEnabled filters plugins that are enabled and should activate
func filterEnabled(plugins []*Plugin, context *ActivationContext) []*Plugin {
	var enabled []*Plugin
	for _, plugin := range plugins {
		if plugin.Enabled && plugin.Manifest.ShouldActivate(context) {
			enabled = append(enabled, plugin)
		}
	}
	return enabled
}

// GetPlugin loads a specific plugin by name
func (l *Loader) GetPlugin(name string) (*Plugin, error) {
	pluginState, err := l.state.GetPlugin(name)
	if err != nil {
		return nil, err
	}

	manifestPath := filepath.Join(pluginState.Path, "plugin.json")
	manifest, err := LoadManifest(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("load manifest: %w", err)
	}

	return &Plugin{
		Name:      manifest.Name,
		Version:   manifest.Version,
		Dir:       pluginState.Path,
		Manifest:  manifest,
		Enabled:   pluginState.Enabled,
		Installed: true,
	}, nil
}

// ValidatePlugin checks if a plugin directory is valid
func (l *Loader) ValidatePlugin(pluginDir string) error {
	// Check manifest exists
	manifestPath := filepath.Join(pluginDir, "plugin.json")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		return fmt.Errorf("plugin.json not found in %s", pluginDir)
	}

	// Load and validate manifest
	manifest, err := LoadManifest(manifestPath)
	if err != nil {
		return err
	}

	// Check referenced files exist
	errors := make([]error, 0)

	// Validate skills
	for _, skillPath := range manifest.GetSkillPaths(pluginDir) {
		if _, err := os.Stat(skillPath); os.IsNotExist(err) {
			errors = append(errors, fmt.Errorf("skill file not found: %s", skillPath))
		}
	}

	// Validate commands
	for _, cmdPath := range manifest.GetCommandPaths(pluginDir) {
		if _, err := os.Stat(cmdPath); os.IsNotExist(err) {
			errors = append(errors, fmt.Errorf("command file not found: %s", cmdPath))
		}
	}

	// Validate agents
	for _, agentPath := range manifest.GetAgentPaths(pluginDir) {
		if _, err := os.Stat(agentPath); os.IsNotExist(err) {
			errors = append(errors, fmt.Errorf("agent file not found: %s", agentPath))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("plugin validation failed: %v", errors)
	}

	return nil
}

// State returns the current plugin state
func (l *Loader) State() *State {
	return l.state
}

// SaveState persists the current state to disk
func (l *Loader) SaveState() error {
	return l.state.Save(l.stateFile)
}

// GetPluginsDir returns the plugins directory path
func (l *Loader) GetPluginsDir() string {
	return l.pluginsDir
}
