// ABOUTME: MCP server registry managing server configurations and persistence
// ABOUTME: Handles CRUD operations and .mcp.json configuration file management

package mcp

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// ServerConfig represents an MCP server configuration
type ServerConfig struct {
	Name      string   `json:"name"`
	Transport string   `json:"transport"`
	Command   string   `json:"command"`
	Args      []string `json:"args,omitempty"`
}

// MCPConfig represents the .mcp.json configuration file format
//
//nolint:revive // MCP prefix clarifies this is MCP protocol config vs generic config
type MCPConfig struct {
	Version string                  `json:"version"`
	Servers map[string]ServerConfig `json:"servers"`
}

// Registry manages MCP server configurations
type Registry struct {
	mu      sync.RWMutex
	servers map[string]ServerConfig
	baseDir string
}

// NewRegistry creates a new server registry
func NewRegistry(baseDir string) *Registry {
	return &Registry{
		servers: make(map[string]ServerConfig),
		baseDir: baseDir,
	}
}

// AddServer adds a new server configuration
func (r *Registry) AddServer(server ServerConfig) error {
	if err := r.validateServer(server); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.servers[server.Name]; exists {
		return fmt.Errorf("server '%s' already exists", server.Name)
	}

	r.servers[server.Name] = server
	return nil
}

// RemoveServer removes a server configuration
func (r *Registry) RemoveServer(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.servers[name]; !exists {
		return fmt.Errorf("server '%s' not found", name)
	}

	delete(r.servers, name)
	return nil
}

// GetServer retrieves a server configuration by name
func (r *Registry) GetServer(name string) (ServerConfig, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	server, exists := r.servers[name]
	if !exists {
		return ServerConfig{}, fmt.Errorf("server '%s' not found", name)
	}

	return server, nil
}

// UpdateServer updates an existing server configuration
func (r *Registry) UpdateServer(server ServerConfig) error {
	if err := r.validateServer(server); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.servers[server.Name]; !exists {
		return fmt.Errorf("server '%s' not found", server.Name)
	}

	r.servers[server.Name] = server
	return nil
}

// ListServers returns all registered servers
func (r *Registry) ListServers() []ServerConfig {
	r.mu.RLock()
	defer r.mu.RUnlock()

	servers := make([]ServerConfig, 0, len(r.servers))
	for _, server := range r.servers {
		servers = append(servers, server)
	}

	return servers
}

// Save persists the registry to .mcp.json
func (r *Registry) Save() error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	config := MCPConfig{
		Version: "1.0",
		Servers: r.servers,
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	configPath := filepath.Join(r.baseDir, ".mcp.json")
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Load reads the registry from .mcp.json
func (r *Registry) Load() error {
	configPath := filepath.Join(r.baseDir, ".mcp.json")

	// If file doesn't exist, start with empty registry
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(configPath) //nolint:gosec // G304: Loading config/template files from validated paths
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var config MCPConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.servers = config.Servers
	if r.servers == nil {
		r.servers = make(map[string]ServerConfig)
	}

	return nil
}

// validateServer checks if a server configuration is valid
func (r *Registry) validateServer(server ServerConfig) error {
	if server.Name == "" {
		return fmt.Errorf("server name cannot be empty")
	}

	if server.Command == "" {
		return fmt.Errorf("server command cannot be empty")
	}

	// Only stdio transport is supported in this phase
	if server.Transport != "stdio" {
		return fmt.Errorf("unsupported transport: %s (only 'stdio' supported)", server.Transport)
	}

	return nil
}

// ConfigPath returns the path to the .mcp.json file
func (r *Registry) ConfigPath() string {
	return filepath.Join(r.baseDir, ".mcp.json")
}

// ServerExists checks if a server with the given name exists
func (r *Registry) ServerExists(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.servers[name]
	return exists
}

// Count returns the number of registered servers
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.servers)
}
