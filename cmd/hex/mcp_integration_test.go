// ABOUTME: Integration tests for MCP tool loading in the CLI
// ABOUTME: Verifies that MCP tools are properly registered and available

package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/2389-research/hex/internal/mcp"
	"github.com/2389-research/hex/internal/tools"
)

func TestMCPToolsRegistration_NoConfig(t *testing.T) {
	// Create temp directory without .mcp.json
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldDir) }()
	_ = os.Chdir(tmpDir)

	// Create tool registry
	registry := tools.NewRegistry()

	// Register some built-in tools
	if err := registry.Register(tools.NewReadTool()); err != nil {
		t.Fatalf("Failed to register read tool: %v", err)
	}

	// Load MCP tools (should gracefully skip with no config)
	err := mcp.LoadMCPTools(".", registry)
	if err != nil {
		t.Fatalf("LoadMCPTools failed: %v", err)
	}

	// Should still have the built-in tool
	toolList := registry.List()
	if len(toolList) != 1 {
		t.Errorf("Expected 1 tool (read), got %d: %v", len(toolList), toolList)
	}

	// Verify read tool is present
	readTool, err := registry.Get("read_file")
	if err != nil {
		t.Errorf("Failed to get read_file tool: %v", err)
	}
	if readTool == nil {
		t.Error("Read tool is nil")
	}
}

func TestMCPToolsRegistration_EmptyConfig(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldDir) }()
	_ = os.Chdir(tmpDir)

	// Create empty .mcp.json
	config := mcp.MCPConfig{
		Version: "1.0",
		Servers: make(map[string]mcp.ServerConfig),
	}
	configData, _ := json.Marshal(config)
	if err := os.WriteFile(".mcp.json", configData, 0600); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Create tool registry
	registry := tools.NewRegistry()

	// Register some built-in tools
	if err := registry.Register(tools.NewReadTool()); err != nil {
		t.Fatalf("Failed to register read tool: %v", err)
	}
	if err := registry.Register(tools.NewWriteTool()); err != nil {
		t.Fatalf("Failed to register write tool: %v", err)
	}

	// Load MCP tools (should succeed with empty config)
	err := mcp.LoadMCPTools(".", registry)
	if err != nil {
		t.Fatalf("LoadMCPTools failed: %v", err)
	}

	// Should still have the built-in tools
	toolList := registry.List()
	if len(toolList) != 2 {
		t.Errorf("Expected 2 tools (read, write), got %d: %v", len(toolList), toolList)
	}
}

func TestMCPToolsRegistration_InvalidConfig(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldDir) }()
	_ = os.Chdir(tmpDir)

	// Write invalid JSON
	if err := os.WriteFile(".mcp.json", []byte("invalid json"), 0600); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Create tool registry
	registry := tools.NewRegistry()

	// Register some built-in tools
	if err := registry.Register(tools.NewReadTool()); err != nil {
		t.Fatalf("Failed to register read tool: %v", err)
	}

	// Load MCP tools should fail
	err := mcp.LoadMCPTools(".", registry)
	if err == nil {
		t.Error("Expected error for invalid config, got nil")
	}

	// Built-in tools should still be registered
	toolList := registry.List()
	if len(toolList) != 1 {
		t.Errorf("Expected 1 tool (read), got %d: %v", len(toolList), toolList)
	}
}

func TestToolDiscovery(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldDir) }()
	_ = os.Chdir(tmpDir)

	// Create registry with multiple built-in tools
	registry := tools.NewRegistry()
	if err := registry.Register(tools.NewReadTool()); err != nil {
		t.Fatalf("Failed to register read tool: %v", err)
	}
	if err := registry.Register(tools.NewWriteTool()); err != nil {
		t.Fatalf("Failed to register write tool: %v", err)
	}
	if err := registry.Register(tools.NewBashTool()); err != nil {
		t.Fatalf("Failed to register bash tool: %v", err)
	}

	// Load MCP tools (no config, should skip gracefully)
	if err := mcp.LoadMCPTools(".", registry); err != nil {
		t.Fatalf("LoadMCPTools failed: %v", err)
	}

	// List all tools
	toolList := registry.List()
	if len(toolList) != 3 {
		t.Errorf("Expected 3 tools, got %d: %v", len(toolList), toolList)
	}

	// Verify specific tools
	expectedTools := []string{"bash", "read_file", "write_file"}
	for _, toolName := range expectedTools {
		tool, err := registry.Get(toolName)
		if err != nil {
			t.Errorf("Failed to get tool %s: %v", toolName, err)
		}
		if tool == nil {
			t.Errorf("Tool %s is nil", toolName)
		}
		if tool != nil && tool.Name() != toolName {
			t.Errorf("Tool name mismatch: expected %s, got %s", toolName, tool.Name())
		}
	}
}

func TestMCPToolsGracefulDegradation(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldDir) }()
	_ = os.Chdir(tmpDir)

	// Create config with non-existent server
	config := mcp.MCPConfig{
		Version: "1.0",
		Servers: map[string]mcp.ServerConfig{
			"broken": {
				Name:      "broken",
				Transport: "stdio",
				Command:   "/nonexistent/command",
			},
		},
	}
	configData, _ := json.MarshalIndent(config, "", "  ")
	if err := os.WriteFile(".mcp.json", configData, 0600); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Create tool registry with built-in tools
	registry := tools.NewRegistry()
	if err := registry.Register(tools.NewReadTool()); err != nil {
		t.Fatalf("Failed to register read tool: %v", err)
	}

	// Load MCP tools (should fail but not crash)
	err := mcp.LoadMCPTools(".", registry)
	if err == nil {
		t.Error("Expected error for broken server, got nil")
	}

	// Built-in tools should still work
	toolList := registry.List()
	if len(toolList) != 1 {
		t.Errorf("Expected 1 tool (read), got %d: %v", len(toolList), toolList)
	}

	// Verify read tool still works
	readTool, err := registry.Get("read_file")
	if err != nil {
		t.Errorf("Failed to get read_file tool after MCP failure: %v", err)
	}
	if readTool == nil {
		t.Error("Read tool is nil after MCP failure")
	}
}

func TestMCPConfigInSubdirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a subdirectory with .mcp.json
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0750); err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	// Create empty .mcp.json in subdirectory
	config := mcp.MCPConfig{
		Version: "1.0",
		Servers: make(map[string]mcp.ServerConfig),
	}
	configData, _ := json.Marshal(config)
	if err := os.WriteFile(filepath.Join(subDir, ".mcp.json"), configData, 0600); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Create tool registry
	registry := tools.NewRegistry()

	// Load from subdirectory
	err := mcp.LoadMCPTools(subDir, registry)
	if err != nil {
		t.Fatalf("LoadMCPTools failed: %v", err)
	}

	// Should succeed with no tools (empty config)
	toolList := registry.List()
	if len(toolList) != 0 {
		t.Errorf("Expected 0 tools, got %d: %v", len(toolList), toolList)
	}
}
