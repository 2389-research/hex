// ABOUTME: MCP tool loader that integrates MCP servers with Hex's tool registry
// ABOUTME: Loads .mcp.json config and initializes mux MCP clients with their tools

package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/2389-research/hex/internal/tools"
	muxmcp "github.com/2389-research/mux/mcp"
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

	// Convert hex config to mux config
	muxConfig := muxmcp.ServerConfig{
		Name:    serverConfig.Name,
		Command: serverConfig.Command,
		Args:    serverConfig.Args,
		Env:     serverConfig.Env,
	}

	// Create mux MCP client
	client, err := muxmcp.NewClient(muxConfig)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	// Start the client (handles initialization handshake)
	if startErr := client.Start(ctx); startErr != nil {
		_ = client.Close()
		return fmt.Errorf("failed to start: %w", startErr)
	}

	// List available tools
	mcpTools, err := client.ListTools(ctx)
	if err != nil {
		_ = client.Close()
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

	// Note: We don't close the client here - it needs to stay alive
	// to handle tool executions. In a production system, we'd need
	// proper lifecycle management to close clients on exit.

	return nil
}
