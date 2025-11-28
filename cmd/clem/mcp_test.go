// ABOUTME: Tests for MCP subcommands (add, list, remove)
// ABOUTME: Verifies CLI command behavior and integration with registry

package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/harper/clem/internal/mcp"
	"github.com/spf13/cobra"
)

func TestMCPAddCommand(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name    string
		args    []string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid server",
			args:    []string{"add", "test-server", "node", "server.js"},
			wantErr: false,
		},
		{
			name:    "with args",
			args:    []string{"add", "server-with-args", "python", "server.py", "--port", "8080"},
			wantErr: false,
		},
		{
			name:    "missing name",
			args:    []string{"add"},
			wantErr: true,
			errMsg:  "requires at least 2 arg(s)",
		},
		{
			name:    "missing command",
			args:    []string{"add", "test-server"},
			wantErr: true,
			errMsg:  "requires at least 2 arg(s)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newMCPCommand(tempDir)
			cmd.SetArgs(tt.args)

			var stdout bytes.Buffer
			cmd.SetOut(&stdout)

			err := cmd.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && tt.errMsg != "" {
				if err == nil || !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error containing %q, got: %v", tt.errMsg, err)
				}
			}

			if !tt.wantErr {
				// Verify server was added
				registry := mcp.NewRegistry(tempDir)
				if err := registry.Load(); err != nil {
					t.Fatalf("Failed to load registry: %v", err)
				}

				serverName := tt.args[1]
				if !registry.ServerExists(serverName) {
					t.Errorf("Server %q was not added to registry", serverName)
				}
			}
		})
	}
}

func TestMCPAddCommand_Duplicate(t *testing.T) {
	tempDir := t.TempDir()

	cmd := newMCPCommand(tempDir)

	// Add first server
	cmd.SetArgs([]string{"add", "test-server", "node", "server.js"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("First add failed: %v", err)
	}

	// Try to add again with same name
	cmd = newMCPCommand(tempDir)
	cmd.SetArgs([]string{"add", "test-server", "python", "other.py"})

	var stderr bytes.Buffer
	cmd.SetErr(&stderr)

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error when adding duplicate server")
	}

	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("Expected 'already exists' error, got: %v", err)
	}
}

func TestMCPListCommand(t *testing.T) {
	tempDir := t.TempDir()

	// Add some servers first
	registry := mcp.NewRegistry(tempDir)
	servers := []mcp.ServerConfig{
		{Name: "server1", Transport: "stdio", Command: "node", Args: []string{"s1.js"}},
		{Name: "server2", Transport: "stdio", Command: "python", Args: []string{"s2.py"}},
	}

	for _, s := range servers {
		registry.AddServer(s)
	}
	registry.Save()

	// Test list command
	cmd := newMCPCommand(tempDir)
	cmd.SetArgs([]string{"list"})

	var stdout bytes.Buffer
	cmd.SetOut(&stdout)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("List command failed: %v", err)
	}

	output := stdout.String()

	// Verify both servers are listed
	if !strings.Contains(output, "server1") {
		t.Error("Output should contain 'server1'")
	}
	if !strings.Contains(output, "server2") {
		t.Error("Output should contain 'server2'")
	}
	if !strings.Contains(output, "node") {
		t.Error("Output should contain 'node'")
	}
	if !strings.Contains(output, "python") {
		t.Error("Output should contain 'python'")
	}
}

func TestMCPListCommand_Empty(t *testing.T) {
	tempDir := t.TempDir()

	cmd := newMCPCommand(tempDir)
	cmd.SetArgs([]string{"list"})

	var stdout bytes.Buffer
	cmd.SetOut(&stdout)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("List command failed: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "No MCP servers") {
		t.Error("Output should indicate no servers configured")
	}
}

func TestMCPRemoveCommand(t *testing.T) {
	tempDir := t.TempDir()

	// Add a server first
	registry := mcp.NewRegistry(tempDir)
	registry.AddServer(mcp.ServerConfig{
		Name:      "test-server",
		Transport: "stdio",
		Command:   "node",
	})
	registry.Save()

	// Remove it
	cmd := newMCPCommand(tempDir)
	cmd.SetArgs([]string{"remove", "test-server"})

	var stdout bytes.Buffer
	cmd.SetOut(&stdout)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Remove command failed: %v", err)
	}

	// Verify it's gone
	registry = mcp.NewRegistry(tempDir)
	registry.Load()

	if registry.ServerExists("test-server") {
		t.Error("Server should have been removed")
	}
}

func TestMCPRemoveCommand_Nonexistent(t *testing.T) {
	tempDir := t.TempDir()

	cmd := newMCPCommand(tempDir)
	cmd.SetArgs([]string{"remove", "nonexistent"})

	var stderr bytes.Buffer
	cmd.SetErr(&stderr)

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error when removing nonexistent server")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected 'not found' error, got: %v", err)
	}
}

func TestMCPRemoveCommand_MissingName(t *testing.T) {
	tempDir := t.TempDir()

	cmd := newMCPCommand(tempDir)
	cmd.SetArgs([]string{"remove"})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error when server name is missing")
	}

	if !strings.Contains(err.Error(), "requires at least 1 arg") {
		t.Errorf("Expected argument error, got: %v", err)
	}
}

func TestMCPCommand_ConfigPath(t *testing.T) {
	tempDir := t.TempDir()

	// Add a server
	cmd := newMCPCommand(tempDir)
	cmd.SetArgs([]string{"add", "test", "node", "test.js"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// Verify .mcp.json was created in the right place
	configPath := filepath.Join(tempDir, ".mcp.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf(".mcp.json should exist at %s", configPath)
	}
}

func TestMCPAddCommand_Validation(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name    string
		args    []string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "empty server name",
			args:    []string{"add", "", "node", "server.js"},
			wantErr: true,
			errMsg:  "server name cannot be empty",
		},
		{
			name:    "empty command",
			args:    []string{"add", "test", "", ""},
			wantErr: true,
			errMsg:  "server command cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newMCPCommand(tempDir)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && tt.errMsg != "" {
				if err == nil || !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error containing %q, got: %v", tt.errMsg, err)
				}
			}
		})
	}
}

func TestMCPListCommand_Format(t *testing.T) {
	tempDir := t.TempDir()

	// Add servers with different configurations
	registry := mcp.NewRegistry(tempDir)
	registry.AddServer(mcp.ServerConfig{
		Name:      "simple-server",
		Transport: "stdio",
		Command:   "node",
		Args:      []string{"server.js"},
	})
	registry.AddServer(mcp.ServerConfig{
		Name:      "complex-server",
		Transport: "stdio",
		Command:   "python",
		Args:      []string{"-m", "server", "--verbose", "--port", "8080"},
	})
	registry.Save()

	cmd := newMCPCommand(tempDir)
	cmd.SetArgs([]string{"list"})

	var stdout bytes.Buffer
	cmd.SetOut(&stdout)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("List failed: %v", err)
	}

	output := stdout.String()

	// Verify formatting includes all details
	if !strings.Contains(output, "simple-server") {
		t.Error("Missing simple-server")
	}
	if !strings.Contains(output, "complex-server") {
		t.Error("Missing complex-server")
	}

	// Should show command with args
	if !strings.Contains(output, "node server.js") {
		t.Error("Should show full command for simple-server")
	}
	if !strings.Contains(output, "python") {
		t.Error("Should show command for complex-server")
	}
}

func TestMCPCommand_Persistence(t *testing.T) {
	tempDir := t.TempDir()

	// Add server
	cmd := newMCPCommand(tempDir)
	cmd.SetArgs([]string{"add", "persist-test", "node", "test.js", "--arg1"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// List servers (should load from disk)
	cmd = newMCPCommand(tempDir)
	cmd.SetArgs([]string{"list"})

	var stdout bytes.Buffer
	cmd.SetOut(&stdout)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("List failed: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "persist-test") {
		t.Error("Server should persist across command invocations")
	}
	if !strings.Contains(output, "--arg1") {
		t.Error("Server args should persist")
	}
}

// newMCPCommand creates a new mcp command for testing with a custom base directory
func newMCPCommand(baseDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Manage MCP servers",
	}

	addCmd := &cobra.Command{
		Use:   "add <name> <command> [args...]",
		Short: "Add an MCP server",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMCPAdd(baseDir, args)
		},
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List MCP servers",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMCPList(cmd, baseDir)
		},
	}

	removeCmd := &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove an MCP server",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMCPRemove(baseDir, args[0])
		},
	}

	cmd.AddCommand(addCmd, listCmd, removeCmd)
	return cmd
}
