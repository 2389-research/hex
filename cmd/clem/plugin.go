package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/harper/clem/internal/logging"
	"github.com/harper/clem/internal/plugins"
	"github.com/spf13/cobra"
)

var pluginCmd = &cobra.Command{
	Use:   "plugin",
	Short: "Manage Clem plugins",
	Long: `Install, uninstall, enable, and disable plugins.

Plugins extend Clem with additional skills, commands, hooks, and MCP servers.`,
}

var pluginInstallCmd = &cobra.Command{
	Use:   "install <source>",
	Short: "Install a plugin",
	Long: `Install a plugin from a source.

Sources can be:
  - Git repository URL: https://github.com/user/plugin.git
  - Local directory path: ./my-plugin or ~/plugins/my-plugin

Examples:
  clem plugin install https://github.com/user/clem-go-plugin.git
  clem plugin install ./my-local-plugin
  clem plugin install ~/plugins/my-plugin`,
	Args: cobra.ExactArgs(1),
	RunE: runPluginInstall,
}

var pluginUninstallCmd = &cobra.Command{
	Use:     "uninstall <name>",
	Short:   "Uninstall a plugin",
	Aliases: []string{"remove", "rm"},
	Args:    cobra.ExactArgs(1),
	RunE:    runPluginUninstall,
}

var pluginListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List installed plugins",
	Aliases: []string{"ls"},
	RunE:    runPluginList,
}

var pluginEnableCmd = &cobra.Command{
	Use:   "enable <name>",
	Short: "Enable a plugin",
	Args:  cobra.ExactArgs(1),
	RunE:  runPluginEnable,
}

var pluginDisableCmd = &cobra.Command{
	Use:   "disable <name>",
	Short: "Disable a plugin",
	Args:  cobra.ExactArgs(1),
	RunE:  runPluginDisable,
}

var pluginUpdateCmd = &cobra.Command{
	Use:   "update <name>",
	Short: "Update a plugin to the latest version",
	Args:  cobra.ExactArgs(1),
	RunE:  runPluginUpdate,
}

var pluginShowCmd = &cobra.Command{
	Use:     "show <name>",
	Short:   "Show detailed information about a plugin",
	Aliases: []string{"info"},
	Args:    cobra.ExactArgs(1),
	RunE:    runPluginShow,
}

func init() {
	pluginCmd.AddCommand(pluginInstallCmd)
	pluginCmd.AddCommand(pluginUninstallCmd)
	pluginCmd.AddCommand(pluginListCmd)
	pluginCmd.AddCommand(pluginEnableCmd)
	pluginCmd.AddCommand(pluginDisableCmd)
	pluginCmd.AddCommand(pluginUpdateCmd)
	pluginCmd.AddCommand(pluginShowCmd)
	rootCmd.AddCommand(pluginCmd)
}

func runPluginInstall(_ *cobra.Command, args []string) error {
	source := args[0]

	registry, err := plugins.DefaultRegistry()
	if err != nil {
		return fmt.Errorf("create registry: %w", err)
	}

	fmt.Printf("Installing plugin from %s...\n", source)
	if err := registry.Install(source); err != nil {
		return fmt.Errorf("install plugin: %w", err)
	}

	fmt.Println("Plugin installed successfully!")
	return nil
}

func runPluginUninstall(_ *cobra.Command, args []string) error {
	name := args[0]

	registry, err := plugins.DefaultRegistry()
	if err != nil {
		return fmt.Errorf("create registry: %w", err)
	}

	fmt.Printf("Uninstalling plugin %s...\n", name)
	if err := registry.Uninstall(name); err != nil {
		return fmt.Errorf("uninstall plugin: %w", err)
	}

	fmt.Println("Plugin uninstalled successfully!")
	return nil
}

func runPluginList(_ *cobra.Command, _ []string) error {
	registry, err := plugins.DefaultRegistry()
	if err != nil {
		return fmt.Errorf("create registry: %w", err)
	}

	allPlugins, err := registry.ListInstalled()
	if err != nil {
		logging.WarnWith("Error discovering plugins", "error", err.Error())
	}

	if len(allPlugins) == 0 {
		fmt.Println("No plugins installed.")
		fmt.Println("\nInstall a plugin with:")
		fmt.Println("  clem plugin install <source>")
		return nil
	}

	// Create tabwriter for aligned output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	//nolint:errcheck // Error writing to stdout is not actionable
	fmt.Fprintln(w, "NAME\tVERSION\tSTATUS\tDESCRIPTION")
	//nolint:errcheck // Error writing to stdout is not actionable
	fmt.Fprintln(w, "----\t-------\t------\t-----------")

	for _, plugin := range allPlugins {
		status := "enabled"
		if !plugin.Enabled {
			status = "disabled"
		}

		description := plugin.Manifest.Description
		if len(description) > 50 {
			description = description[:47] + "..."
		}

		//nolint:errcheck // Error writing to stdout is not actionable
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			plugin.Name,
			plugin.Version,
			status,
			description,
		)
	}

	_ = w.Flush() // Best effort flush, error writing to stdout is not actionable
	return nil
}

func runPluginEnable(_ *cobra.Command, args []string) error {
	name := args[0]

	registry, err := plugins.DefaultRegistry()
	if err != nil {
		return fmt.Errorf("create registry: %w", err)
	}

	if err := registry.Enable(name); err != nil {
		return fmt.Errorf("enable plugin: %w", err)
	}

	fmt.Printf("Plugin %s enabled!\n", name)
	return nil
}

func runPluginDisable(_ *cobra.Command, args []string) error {
	name := args[0]

	registry, err := plugins.DefaultRegistry()
	if err != nil {
		return fmt.Errorf("create registry: %w", err)
	}

	if err := registry.Disable(name); err != nil {
		return fmt.Errorf("disable plugin: %w", err)
	}

	fmt.Printf("Plugin %s disabled!\n", name)
	return nil
}

func runPluginUpdate(_ *cobra.Command, args []string) error {
	name := args[0]

	registry, err := plugins.DefaultRegistry()
	if err != nil {
		return fmt.Errorf("create registry: %w", err)
	}

	fmt.Printf("Updating plugin %s...\n", name)
	if err := registry.Update(name); err != nil {
		return fmt.Errorf("update plugin: %w", err)
	}

	fmt.Println("Plugin updated successfully!")
	return nil
}

func runPluginShow(_ *cobra.Command, args []string) error {
	name := args[0]

	registry, err := plugins.DefaultRegistry()
	if err != nil {
		return fmt.Errorf("create registry: %w", err)
	}

	allPlugins, err := registry.ListInstalled()
	if err != nil {
		return fmt.Errorf("list plugins: %w", err)
	}

	var targetPlugin *plugins.Plugin
	for _, p := range allPlugins {
		if p.Name == name {
			targetPlugin = p
			break
		}
	}

	if targetPlugin == nil {
		return fmt.Errorf("plugin not found: %s", name)
	}

	// Display plugin information
	fmt.Printf("Name:        %s\n", targetPlugin.Name)
	fmt.Printf("Version:     %s\n", targetPlugin.Version)
	fmt.Printf("Status:      %s\n", func() string {
		if targetPlugin.Enabled {
			return "enabled"
		}
		return "disabled"
	}())
	fmt.Printf("Description: %s\n", targetPlugin.Manifest.Description)

	if targetPlugin.Manifest.Author != "" {
		fmt.Printf("Author:      %s\n", targetPlugin.Manifest.Author)
	}

	if targetPlugin.Manifest.License != "" {
		fmt.Printf("License:     %s\n", targetPlugin.Manifest.License)
	}

	if targetPlugin.Manifest.Homepage != "" {
		fmt.Printf("Homepage:    %s\n", targetPlugin.Manifest.Homepage)
	}

	fmt.Printf("Path:        %s\n", targetPlugin.Dir)

	// Show what the plugin contributes
	fmt.Println("\nContributes:")
	if len(targetPlugin.Manifest.Skills) > 0 {
		fmt.Printf("  Skills:    %d\n", len(targetPlugin.Manifest.Skills))
	}
	if len(targetPlugin.Manifest.Commands) > 0 {
		fmt.Printf("  Commands:  %d\n", len(targetPlugin.Manifest.Commands))
	}
	if len(targetPlugin.Manifest.Agents) > 0 {
		fmt.Printf("  Agents:    %d\n", len(targetPlugin.Manifest.Agents))
	}
	if len(targetPlugin.Manifest.Hooks) > 0 {
		fmt.Printf("  Hooks:     %d event types\n", len(targetPlugin.Manifest.Hooks))
	}
	if len(targetPlugin.Manifest.MCPServers) > 0 {
		fmt.Printf("  MCP Servers: %d\n", len(targetPlugin.Manifest.MCPServers))
	}

	return nil
}
