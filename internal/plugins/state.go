// ABOUTME: Plugin state management for tracking installed plugins
// ABOUTME: Handles persistence of plugin installation state and configuration

package plugins

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// State tracks the installation state of all plugins
type State struct {
	Plugins map[string]*PluginState `json:"plugins"`
}

// PluginState tracks an individual plugin's state
type PluginState struct {
	Enabled   bool      `json:"enabled"`
	Version   string    `json:"version"`
	Installed time.Time `json:"installed"`
	Updated   time.Time `json:"updated,omitempty"`
	Path      string    `json:"path"` // Full path to plugin directory
}

// LoadState loads plugin state from disk
func LoadState(stateFile string) (*State, error) {
	// Create default state if file doesn't exist
	if _, err := os.Stat(stateFile); os.IsNotExist(err) {
		return &State{
			Plugins: make(map[string]*PluginState),
		}, nil
	}

	//nolint:gosec // G304 - reading state file from known config directory
	data, err := os.ReadFile(stateFile)
	if err != nil {
		return nil, fmt.Errorf("read state file: %w", err)
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("parse state file: %w", err)
	}

	// Initialize map if nil
	if state.Plugins == nil {
		state.Plugins = make(map[string]*PluginState)
	}

	return &state, nil
}

// Save writes the state to disk
func (s *State) Save(stateFile string) error {
	// Ensure parent directory exists
	//nolint:gosec // G301 - 0755 is correct for config directories
	if err := os.MkdirAll(filepath.Dir(stateFile), 0755); err != nil {
		return fmt.Errorf("create state directory: %w", err)
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}

	//nolint:gosec // G306 - 0644 is correct for state files
	if err := os.WriteFile(stateFile, data, 0644); err != nil {
		return fmt.Errorf("write state file: %w", err)
	}

	return nil
}

// AddPlugin records a newly installed plugin
func (s *State) AddPlugin(name, version, path string) {
	s.Plugins[name] = &PluginState{
		Enabled:   true,
		Version:   version,
		Installed: time.Now(),
		Path:      path,
	}
}

// RemovePlugin removes a plugin from state
func (s *State) RemovePlugin(name string) {
	delete(s.Plugins, name)
}

// UpdatePlugin updates a plugin's version and timestamp
func (s *State) UpdatePlugin(name, version string) error {
	plugin, exists := s.Plugins[name]
	if !exists {
		return fmt.Errorf("plugin not found: %s", name)
	}

	plugin.Version = version
	plugin.Updated = time.Now()
	return nil
}

// EnablePlugin enables a plugin
func (s *State) EnablePlugin(name string) error {
	plugin, exists := s.Plugins[name]
	if !exists {
		return fmt.Errorf("plugin not found: %s", name)
	}

	plugin.Enabled = true
	return nil
}

// DisablePlugin disables a plugin
func (s *State) DisablePlugin(name string) error {
	plugin, exists := s.Plugins[name]
	if !exists {
		return fmt.Errorf("plugin not found: %s", name)
	}

	plugin.Enabled = false
	return nil
}

// IsEnabled checks if a plugin is enabled
func (s *State) IsEnabled(name string) bool {
	plugin, exists := s.Plugins[name]
	if !exists {
		return false
	}
	return plugin.Enabled
}

// IsInstalled checks if a plugin is installed
func (s *State) IsInstalled(name string) bool {
	_, exists := s.Plugins[name]
	return exists
}

// GetPlugin retrieves a plugin's state
func (s *State) GetPlugin(name string) (*PluginState, error) {
	plugin, exists := s.Plugins[name]
	if !exists {
		return nil, fmt.Errorf("plugin not found: %s", name)
	}
	return plugin, nil
}

// ListEnabled returns all enabled plugin names
func (s *State) ListEnabled() []string {
	var enabled []string
	for name, plugin := range s.Plugins {
		if plugin.Enabled {
			enabled = append(enabled, name)
		}
	}
	return enabled
}

// ListAll returns all plugin names
func (s *State) ListAll() []string {
	//nolint:prealloc // Exact size is known but not worth pre-allocating for small maps
	var all []string
	for name := range s.Plugins {
		all = append(all, name)
	}
	return all
}
