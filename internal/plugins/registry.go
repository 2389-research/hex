// ABOUTME: Plugin registry for managing loaded plugins and their contributions
// ABOUTME: Coordinates plugin loading and integration with Clem's subsystems

package plugins

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Registry manages loaded plugins and their contributions
type Registry struct {
	mu         sync.RWMutex
	plugins    map[string]*Plugin
	loader     *Loader
	installer  *Installer
	pluginsDir string
	stateFile  string
}

// NewRegistry creates a new plugin registry
func NewRegistry(pluginsDir, stateFile string) (*Registry, error) {
	loader, err := NewLoader(pluginsDir, stateFile)
	if err != nil {
		return nil, fmt.Errorf("create loader: %w", err)
	}

	installer := NewInstaller(loader)

	return &Registry{
		plugins:    make(map[string]*Plugin),
		loader:     loader,
		installer:  installer,
		pluginsDir: pluginsDir,
		stateFile:  stateFile,
	}, nil
}

// DefaultRegistry creates a registry with default paths
func DefaultRegistry() (*Registry, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("get home directory: %w", err)
	}

	pluginsDir := filepath.Join(home, ".clem", "plugins")
	stateFile := filepath.Join(home, ".clem", "plugins", "state.json")

	return NewRegistry(pluginsDir, stateFile)
}

// LoadAll discovers and loads all enabled plugins
func (r *Registry) LoadAll(context *ActivationContext) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	plugins, err := r.loader.LoadEnabled(context)
	if err != nil {
		// Still continue with partial results
		for _, plugin := range plugins {
			r.plugins[plugin.Name] = plugin
		}
		return fmt.Errorf("load plugins: %w (loaded %d plugins)", err, len(plugins))
	}

	for _, plugin := range plugins {
		r.plugins[plugin.Name] = plugin
	}

	return nil
}

// Get retrieves a loaded plugin by name
func (r *Registry) Get(name string) (*Plugin, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plugin, exists := r.plugins[name]
	return plugin, exists
}

// List returns all loaded plugin names
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.plugins))
	for name := range r.plugins {
		names = append(names, name)
	}
	return names
}

// GetAll returns all loaded plugins
func (r *Registry) GetAll() []*Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plugins := make([]*Plugin, 0, len(r.plugins))
	for _, plugin := range r.plugins {
		plugins = append(plugins, plugin)
	}
	return plugins
}

// Install installs a plugin from a source
func (r *Registry) Install(source string) error {
	return r.installer.Install(source)
}

// Uninstall removes a plugin
func (r *Registry) Uninstall(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Remove from loaded plugins
	delete(r.plugins, name)

	return r.installer.Uninstall(name)
}

// Update updates a plugin to the latest version
func (r *Registry) Update(name string) error {
	return r.installer.Update(name)
}

// Enable enables a plugin
func (r *Registry) Enable(name string) error {
	return r.installer.Enable(name)
}

// Disable disables a plugin
func (r *Registry) Disable(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Remove from loaded plugins
	delete(r.plugins, name)

	return r.installer.Disable(name)
}

// ListInstalled returns all installed plugins (enabled and disabled)
func (r *Registry) ListInstalled() ([]*Plugin, error) {
	return r.loader.DiscoverAll()
}

// GetSkillPaths returns all skill file paths from loaded plugins
func (r *Registry) GetSkillPaths() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var paths []string
	for _, plugin := range r.plugins {
		paths = append(paths, plugin.Manifest.GetSkillPaths(plugin.Dir)...)
	}
	return paths
}

// GetCommandPaths returns all command file paths from loaded plugins
func (r *Registry) GetCommandPaths() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var paths []string
	for _, plugin := range r.plugins {
		paths = append(paths, plugin.Manifest.GetCommandPaths(plugin.Dir)...)
	}
	return paths
}

// GetAgentPaths returns all agent file paths from loaded plugins
func (r *Registry) GetAgentPaths() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var paths []string
	for _, plugin := range r.plugins {
		paths = append(paths, plugin.Manifest.GetAgentPaths(plugin.Dir)...)
	}
	return paths
}

// GetTemplatePaths returns all template file paths from loaded plugins
func (r *Registry) GetTemplatePaths() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var paths []string
	for _, plugin := range r.plugins {
		paths = append(paths, plugin.Manifest.GetTemplatePaths(plugin.Dir)...)
	}
	return paths
}

// GetHooks returns all hooks from loaded plugins
func (r *Registry) GetHooks() map[string][]map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	allHooks := make(map[string][]map[string]interface{})

	for _, plugin := range r.plugins {
		if plugin.Manifest.Hooks == nil {
			continue
		}

		// Merge hooks from this plugin
		for eventType, hooks := range plugin.Manifest.Hooks {
			if _, exists := allHooks[eventType]; !exists {
				allHooks[eventType] = make([]map[string]interface{}, 0)
			}

			// Convert hooks to slice format
			switch h := hooks.(type) {
			case map[string]interface{}:
				allHooks[eventType] = append(allHooks[eventType], h)
			case []interface{}:
				for _, hook := range h {
					if hookMap, ok := hook.(map[string]interface{}); ok {
						allHooks[eventType] = append(allHooks[eventType], hookMap)
					}
				}
			}
		}
	}

	return allHooks
}

// GetMCPServers returns all MCP server configs from loaded plugins
func (r *Registry) GetMCPServers() map[string]MCPConfig {
	r.mu.RLock()
	defer r.mu.RUnlock()

	allServers := make(map[string]MCPConfig)

	for _, plugin := range r.plugins {
		if plugin.Manifest.MCPServers == nil {
			continue
		}

		// Merge MCP servers from this plugin
		for name, config := range plugin.Manifest.MCPServers {
			// Prefix server name with plugin name to avoid conflicts
			serverName := fmt.Sprintf("%s:%s", plugin.Name, name)
			allServers[serverName] = config
		}
	}

	return allServers
}

// Count returns the number of loaded plugins
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.plugins)
}

// GetPluginsDir returns the plugins directory path
func (r *Registry) GetPluginsDir() string {
	return r.pluginsDir
}

// GetStateFile returns the state file path
func (r *Registry) GetStateFile() string {
	return r.stateFile
}
