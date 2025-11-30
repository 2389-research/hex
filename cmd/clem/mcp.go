// ABOUTME: MCP subcommands for managing MCP server configurations
// ABOUTME: Implements add, list, and remove commands for MCP servers

package main

import (
	"fmt"
	"strings"

	"github.com/harper/clem/internal/mcp"
	"github.com/spf13/cobra"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Manage MCP (Model Context Protocol) servers",
	Long: `Manage MCP servers that provide external tools and resources.

MCP servers extend Clem's capabilities by providing additional tools
that can be used during conversations with Claude.`,
}

var mcpAddCmd = &cobra.Command{
	Use:   "add <name> <command> [args...]",
	Short: "Add a new MCP server",
	Long: `Add a new MCP server configuration.

The server will be configured to use stdio transport. The command
and arguments specify how to launch the server process.

Examples:
  clem mcp add weather node weather-server.js
  clem mcp add database python -m database_server --port 8080
  clem mcp add files /usr/local/bin/file-server`,
	Args:               cobra.MinimumNArgs(2),
	DisableFlagParsing: true,
	RunE: func(_ *cobra.Command, args []string) error {
		return runMCPAdd(".", args)
	},
}

var mcpListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured MCP servers",
	Long: `List all configured MCP servers.

Shows the server name, transport type, command, and arguments
for each configured server.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		return runMCPList(cmd, ".")
	},
}

var mcpRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove an MCP server",
	Long: `Remove an MCP server configuration.

This removes the server from the configuration file. It does not
affect the actual server binary or scripts.`,
	Args: cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		return runMCPRemove(".", args[0])
	},
}

func init() {
	rootCmd.AddCommand(mcpCmd)
	mcpCmd.AddCommand(mcpAddCmd, mcpListCmd, mcpRemoveCmd)
}

func runMCPAdd(baseDir string, args []string) error {
	name := args[0]
	command := args[1]
	cmdArgs := []string{}
	if len(args) > 2 {
		cmdArgs = args[2:]
	}

	// Validate inputs
	if name == "" {
		return fmt.Errorf("server name cannot be empty")
	}
	if command == "" {
		return fmt.Errorf("server command cannot be empty")
	}

	// Load registry
	registry := mcp.NewRegistry(baseDir)
	if err := registry.Load(); err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	// Add server
	server := mcp.ServerConfig{
		Name:      name,
		Transport: "stdio",
		Command:   command,
		Args:      cmdArgs,
	}

	if err := registry.AddServer(server); err != nil {
		return fmt.Errorf("failed to add server: %w", err)
	}

	// Save
	if err := registry.Save(); err != nil {
		return fmt.Errorf("failed to save registry: %w", err)
	}

	fmt.Printf("Added MCP server '%s'\n", name)
	fmt.Printf("  Command: %s %s\n", command, strings.Join(cmdArgs, " "))
	fmt.Printf("  Config: %s\n", registry.ConfigPath())

	return nil
}

func runMCPList(cmd *cobra.Command, baseDir string) error {
	registry := mcp.NewRegistry(baseDir)
	if err := registry.Load(); err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	servers := registry.ListServers()
	if len(servers) == 0 {
		cmd.Println("No MCP servers configured.")
		cmd.Println()
		cmd.Println("Add a server with: clem mcp add <name> <command> [args...]")
		return nil
	}

	cmd.Println("Configured MCP servers:")
	cmd.Println()

	for _, server := range servers {
		cmd.Printf("  %s\n", server.Name)
		cmd.Printf("    Transport: %s\n", server.Transport)

		fullCmd := server.Command
		if len(server.Args) > 0 {
			fullCmd = fullCmd + " " + strings.Join(server.Args, " ")
		}
		cmd.Printf("    Command:   %s\n", fullCmd)
		cmd.Println()
	}

	cmd.Printf("Total: %d server(s)\n", len(servers))
	cmd.Printf("Config: %s\n", registry.ConfigPath())

	return nil
}

func runMCPRemove(baseDir string, name string) error {
	registry := mcp.NewRegistry(baseDir)
	if err := registry.Load(); err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	if err := registry.RemoveServer(name); err != nil {
		return fmt.Errorf("failed to remove server: %w", err)
	}

	if err := registry.Save(); err != nil {
		return fmt.Errorf("failed to save registry: %w", err)
	}

	fmt.Printf("Removed MCP server '%s'\n", name)
	return nil
}
