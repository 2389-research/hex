// ABOUTME: MCP tool loader that integrates MCP servers with Clem's tool registry
// ABOUTME: Loads .mcp.json config and initializes MCP clients with their tools

package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/harper/pagent/internal/tools"
)

// LoadMCPTools loads MCP server configurations from .mcp.json and registers their tools
// with the provided tool registry. Returns nil if .mcp.json doesn't exist (graceful skip).
// Returns error if config is invalid or if any server fails to initialize.
func LoadMCPTools(baseDir string, registry *tools.Registry) error {
	configPath := filepath.Join(baseDir, ".mcp.json")

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// No config file - this is OK, just skip loading
		return nil
	}

	// Read and parse config file
	data, err := os.ReadFile(configPath) //nolint:gosec // G304: Loading config/template files from validated paths
	if err != nil {
		return fmt.Errorf("failed to read .mcp.json: %w", err)
	}

	var config MCPConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse .mcp.json: %w", err)
	}

	// If no servers configured, return successfully
	if len(config.Servers) == 0 {
		return nil
	}

	// Initialize each server
	for _, serverConfig := range config.Servers {
		if err := loadServerTools(serverConfig, registry); err != nil {
			return fmt.Errorf("failed to load tools from server '%s': %w", serverConfig.Name, err)
		}
	}

	return nil
}

// loadServerTools initializes a single MCP server and registers its tools
func loadServerTools(serverConfig ServerConfig, registry *tools.Registry) error {
	// Create context with timeout for initialization
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create MCP client
	client, err := NewClient(serverConfig.Command, serverConfig.Args...)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	// Initialize the client
	if err := client.Initialize(ctx, "clem", "0.1.0", "2024-11-05"); err != nil {
		_ = client.Shutdown()
		return fmt.Errorf("failed to initialize: %w", err)
	}

	// List available tools
	mcpTools, err := client.ListTools(ctx)
	if err != nil {
		_ = client.Shutdown()
		return fmt.Errorf("failed to list tools: %w", err)
	}

	// Create adapters and register each tool
	for _, mcpTool := range mcpTools {
		adapter := NewMCPToolAdapter(client, mcpTool)
		if err := registry.Register(adapter); err != nil {
			// If tool already exists, skip it (don't fail)
			// This allows multiple servers to provide the same tool
			continue
		}
	}

	// Note: We don't shutdown the client here - it needs to stay alive
	// to handle tool executions. In a production system, we'd need
	// proper lifecycle management to shut down clients on exit.

	return nil
}
