// ABOUTME: Tests for MCP server registry including persistence to .mcp.json
// ABOUTME: Covers add/remove/list operations and configuration file handling

package mcp

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestRegistry_AddServer(t *testing.T) {
	tempDir := t.TempDir()
	registry := NewRegistry(tempDir)

	server := ServerConfig{
		Name:      "test-server",
		Transport: "stdio",
		Command:   "node",
		Args:      []string{"server.js"},
	}

	err := registry.AddServer(server)
	if err != nil {
		t.Fatalf("AddServer() error = %v", err)
	}

	// Verify server was added
	servers := registry.ListServers()
	if len(servers) != 1 {
		t.Errorf("Expected 1 server, got %d", len(servers))
	}

	if servers[0].Name != "test-server" {
		t.Errorf("Server name = %q, want 'test-server'", servers[0].Name)
	}
}

func TestRegistry_AddServer_Duplicate(t *testing.T) {
	tempDir := t.TempDir()
	registry := NewRegistry(tempDir)

	server := ServerConfig{
		Name:      "test-server",
		Transport: "stdio",
		Command:   "node",
		Args:      []string{"server.js"},
	}

	// Add first time - should succeed
	if err := registry.AddServer(server); err != nil {
		t.Fatalf("First AddServer() error = %v", err)
	}

	// Add second time - should fail
	err := registry.AddServer(server)
	if err == nil {
		t.Error("AddServer() should fail for duplicate name")
	}

	if err != nil && err.Error() != "server 'test-server' already exists" {
		t.Errorf("Expected duplicate error, got: %v", err)
	}
}

func TestRegistry_RemoveServer(t *testing.T) {
	tempDir := t.TempDir()
	registry := NewRegistry(tempDir)

	server := ServerConfig{
		Name:      "test-server",
		Transport: "stdio",
		Command:   "node",
		Args:      []string{"server.js"},
	}

	_ = registry.AddServer(server)

	// Remove server
	err := registry.RemoveServer("test-server")
	if err != nil {
		t.Fatalf("RemoveServer() error = %v", err)
	}

	// Verify removed
	servers := registry.ListServers()
	if len(servers) != 0 {
		t.Errorf("Expected 0 servers after removal, got %d", len(servers))
	}
}

func TestRegistry_RemoveServer_Nonexistent(t *testing.T) {
	tempDir := t.TempDir()
	registry := NewRegistry(tempDir)

	err := registry.RemoveServer("nonexistent")
	if err == nil {
		t.Error("RemoveServer() should fail for nonexistent server")
	}

	if err != nil && err.Error() != "server 'nonexistent' not found" {
		t.Errorf("Expected not found error, got: %v", err)
	}
}

func TestRegistry_ListServers(t *testing.T) {
	tempDir := t.TempDir()
	registry := NewRegistry(tempDir)

	// Initially empty
	servers := registry.ListServers()
	if len(servers) != 0 {
		t.Errorf("Expected 0 servers initially, got %d", len(servers))
	}

	// Add multiple servers
	servers1 := []ServerConfig{
		{Name: "server1", Transport: "stdio", Command: "cmd1"},
		{Name: "server2", Transport: "stdio", Command: "cmd2"},
		{Name: "server3", Transport: "stdio", Command: "cmd3"},
	}

	for _, s := range servers1 {
		if err := registry.AddServer(s); err != nil {
			t.Fatalf("AddServer(%q) error = %v", s.Name, err)
		}
	}

	servers = registry.ListServers()
	if len(servers) != 3 {
		t.Errorf("Expected 3 servers, got %d", len(servers))
	}
}

func TestRegistry_GetServer(t *testing.T) {
	tempDir := t.TempDir()
	registry := NewRegistry(tempDir)

	server := ServerConfig{
		Name:      "test-server",
		Transport: "stdio",
		Command:   "node",
		Args:      []string{"server.js", "--port", "8080"},
	}

	_ = registry.AddServer(server)

	// Get existing server
	retrieved, err := registry.GetServer("test-server")
	if err != nil {
		t.Fatalf("GetServer() error = %v", err)
	}

	if retrieved.Name != server.Name {
		t.Errorf("Name = %q, want %q", retrieved.Name, server.Name)
	}
	if retrieved.Command != server.Command {
		t.Errorf("Command = %q, want %q", retrieved.Command, server.Command)
	}
	if len(retrieved.Args) != len(server.Args) {
		t.Errorf("Args length = %d, want %d", len(retrieved.Args), len(server.Args))
	}
}

func TestRegistry_GetServer_Nonexistent(t *testing.T) {
	tempDir := t.TempDir()
	registry := NewRegistry(tempDir)

	_, err := registry.GetServer("nonexistent")
	if err == nil {
		t.Error("GetServer() should fail for nonexistent server")
	}
}

func TestRegistry_Persist(t *testing.T) {
	tempDir := t.TempDir()
	registry := NewRegistry(tempDir)

	servers := []ServerConfig{
		{Name: "server1", Transport: "stdio", Command: "cmd1", Args: []string{"arg1"}},
		{Name: "server2", Transport: "stdio", Command: "cmd2"},
	}

	for _, s := range servers {
		_ = registry.AddServer(s)
	}

	// Persist to disk
	err := registry.Save()
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify file exists
	configPath := filepath.Join(tempDir, ".mcp.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal(".mcp.json file was not created")
	}

	// Verify file contents
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read .mcp.json: %v", err)
	}

	var config MCPConfig
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatalf("Failed to parse .mcp.json: %v", err)
	}

	if len(config.Servers) != 2 {
		t.Errorf("Expected 2 servers in config, got %d", len(config.Servers))
	}
}

func TestRegistry_Load(t *testing.T) {
	tempDir := t.TempDir()

	// Create a .mcp.json file
	config := MCPConfig{
		Version: "1.0",
		Servers: map[string]ServerConfig{
			"server1": {
				Name:      "server1",
				Transport: "stdio",
				Command:   "python",
				Args:      []string{"server.py"},
			},
			"server2": {
				Name:      "server2",
				Transport: "stdio",
				Command:   "node",
				Args:      []string{"server.js"},
			},
		},
	}

	configPath := filepath.Join(tempDir, ".mcp.json")
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Load registry
	registry := NewRegistry(tempDir)
	if err := registry.Load(); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify servers were loaded
	servers := registry.ListServers()
	if len(servers) != 2 {
		t.Errorf("Expected 2 servers after load, got %d", len(servers))
	}

	// Verify specific server
	server1, err := registry.GetServer("server1")
	if err != nil {
		t.Fatalf("GetServer('server1') error = %v", err)
	}

	if server1.Command != "python" {
		t.Errorf("server1 command = %q, want 'python'", server1.Command)
	}
}

func TestRegistry_Load_NoFile(t *testing.T) {
	tempDir := t.TempDir()
	registry := NewRegistry(tempDir)

	// Load when no file exists should succeed (empty registry)
	err := registry.Load()
	if err != nil {
		t.Errorf("Load() with no file should succeed, got error: %v", err)
	}

	servers := registry.ListServers()
	if len(servers) != 0 {
		t.Errorf("Expected 0 servers when no config file, got %d", len(servers))
	}
}

func TestRegistry_AutoSave(t *testing.T) {
	tempDir := t.TempDir()
	registry := NewRegistry(tempDir)

	server := ServerConfig{
		Name:      "test-server",
		Transport: "stdio",
		Command:   "test",
	}

	// Add server (should auto-save if enabled)
	_ = registry.AddServer(server)
	_ = registry.Save()

	// Create new registry and load
	registry2 := NewRegistry(tempDir)
	if err := registry2.Load(); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify server persisted
	servers := registry2.ListServers()
	if len(servers) != 1 {
		t.Errorf("Expected 1 server after reload, got %d", len(servers))
	}
}

func TestRegistry_ValidateServer(t *testing.T) {
	tests := []struct {
		name    string
		server  ServerConfig
		wantErr bool
	}{
		{
			name: "valid server",
			server: ServerConfig{
				Name:      "valid",
				Transport: "stdio",
				Command:   "node",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			server: ServerConfig{
				Transport: "stdio",
				Command:   "node",
			},
			wantErr: true,
		},
		{
			name: "missing command",
			server: ServerConfig{
				Name:      "test",
				Transport: "stdio",
			},
			wantErr: true,
		},
		{
			name: "invalid transport",
			server: ServerConfig{
				Name:      "test",
				Transport: "http", // Not supported yet
				Command:   "node",
			},
			wantErr: true,
		},
	}

	tempDir := t.TempDir()
	registry := NewRegistry(tempDir)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := registry.AddServer(tt.server)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddServer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRegistry_UpdateServer(t *testing.T) {
	tempDir := t.TempDir()
	registry := NewRegistry(tempDir)

	original := ServerConfig{
		Name:      "test-server",
		Transport: "stdio",
		Command:   "node",
		Args:      []string{"old.js"},
	}

	_ = registry.AddServer(original)

	updated := ServerConfig{
		Name:      "test-server",
		Transport: "stdio",
		Command:   "node",
		Args:      []string{"new.js"},
	}

	err := registry.UpdateServer(updated)
	if err != nil {
		t.Fatalf("UpdateServer() error = %v", err)
	}

	// Verify update
	server, err := registry.GetServer("test-server")
	if err != nil {
		t.Fatalf("GetServer() error = %v", err)
	}

	if len(server.Args) != 1 || server.Args[0] != "new.js" {
		t.Errorf("Server args = %v, want [new.js]", server.Args)
	}
}

func TestRegistry_UpdateServer_Nonexistent(t *testing.T) {
	tempDir := t.TempDir()
	registry := NewRegistry(tempDir)

	server := ServerConfig{
		Name:      "nonexistent",
		Transport: "stdio",
		Command:   "test",
	}

	err := registry.UpdateServer(server)
	if err == nil {
		t.Error("UpdateServer() should fail for nonexistent server")
	}
}
