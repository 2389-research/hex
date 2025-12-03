// ABOUTME: Tests for MCP tool loading and integration with tool registry
// ABOUTME: Verifies loading from .mcp.json and graceful error handling

package mcp

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/harper/hex/internal/tools"
)

func TestLoadMCPTools_NoConfigFile(t *testing.T) {
	// Create temp directory without .mcp.json
	tmpDir := t.TempDir()

	// Create tool registry
	registry := tools.NewRegistry()

	// Should succeed with no tools added
	err := LoadMCPTools(tmpDir, registry)
	if err != nil {
		t.Fatalf("LoadMCPTools failed with no config: %v", err)
	}

	// Registry should be empty (or have only pre-registered tools)
	toolList := registry.List()
	if len(toolList) != 0 {
		t.Errorf("Expected 0 tools, got %d: %v", len(toolList), toolList)
	}
}

func TestLoadMCPTools_EmptyConfig(t *testing.T) {
	tmpDir := t.TempDir()

	// Create empty .mcp.json
	config := MCPConfig{
		Version: "1.0",
		Servers: make(map[string]ServerConfig),
	}
	configData, _ := json.Marshal(config)
	err := os.WriteFile(filepath.Join(tmpDir, ".mcp.json"), configData, 0600)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Create tool registry
	registry := tools.NewRegistry()

	// Should succeed with no tools added
	err = LoadMCPTools(tmpDir, registry)
	if err != nil {
		t.Fatalf("LoadMCPTools failed with empty config: %v", err)
	}

	// Registry should be empty
	toolList := registry.List()
	if len(toolList) != 0 {
		t.Errorf("Expected 0 tools, got %d: %v", len(toolList), toolList)
	}
}

func TestLoadMCPTools_ValidServerConfig(t *testing.T) {
	t.Skip("Skipping integration test - requires real MCP server")
	// This test would require a real MCP server process
	// Integration tests should be in mcp_integration_test.go
}

func TestLoadMCPTools_InvalidConfig(t *testing.T) {
	tmpDir := t.TempDir()

	// Write invalid JSON to .mcp.json
	err := os.WriteFile(filepath.Join(tmpDir, ".mcp.json"), []byte("not valid json"), 0600)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Create tool registry
	registry := tools.NewRegistry()

	// Should fail with parse error
	err = LoadMCPTools(tmpDir, registry)
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestLoadMCPTools_ServerInitializationFailure(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .mcp.json with non-existent command
	config := MCPConfig{
		Version: "1.0",
		Servers: map[string]ServerConfig{
			"broken-server": {
				Name:      "broken-server",
				Transport: "stdio",
				Command:   "/nonexistent/command",
				Args:      []string{},
			},
		},
	}
	configData, _ := json.MarshalIndent(config, "", "  ")
	err := os.WriteFile(filepath.Join(tmpDir, ".mcp.json"), configData, 0600)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Create tool registry
	registry := tools.NewRegistry()

	// Should fail when trying to create client
	err = LoadMCPTools(tmpDir, registry)
	if err == nil {
		t.Error("Expected error for non-existent command, got nil")
	}
}

func TestLoadMCPTools_MultipleServers(t *testing.T) {
	t.Skip("Skipping integration test - requires real MCP servers")
	// This test would require real MCP server processes
	// Integration tests should be in mcp_integration_test.go
}

func TestLoadMCPTools_ContextTimeout(t *testing.T) {
	t.Skip("Skipping integration test - requires real MCP server that times out")
	// This test would require a real MCP server that doesn't respond
	// Integration tests should be in mcp_integration_test.go
}
