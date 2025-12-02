package main

import (
	"path/filepath"

	"github.com/harper/clem/internal/commands"
	"github.com/harper/clem/internal/logging"
	"github.com/harper/clem/internal/tools"
)

// initializeCommands loads commands from all sources and returns registry and tool
func initializeCommands() (*commands.Registry, tools.Tool) {
	logging.Debug("Initializing slash commands system")

	// Create command loader
	loader := commands.NewLoader()

	// Set builtin commands directory (relative to binary location)
	loader.BuiltinDir = findBuiltinCommandsDir()

	// Load all commands
	loadedCommands, err := loader.LoadAll()
	if err != nil {
		logging.WarnWith("Failed to load some commands", "error", err.Error())
	}

	// Create and populate registry
	registry := commands.NewRegistry()
	for _, cmd := range loadedCommands {
		if err := registry.Register(cmd); err != nil {
			logging.WarnWith("Failed to register command", "name", cmd.Name, "error", err.Error())
		} else {
			logging.DebugWith("Registered command", "name", cmd.Name, "source", cmd.Source)
		}
	}

	// Create tool
	tool := commands.NewSlashCommandTool(registry)

	return registry, tool
}

// findBuiltinCommandsDir locates the built-in commands directory
func findBuiltinCommandsDir() string {
	// Try several possible locations
	possibleDirs := []string{
		"./commands",                     // Development: same directory as binary
		"../commands",                    // Development: parent directory
		"/usr/local/share/clem/commands", // Linux install
		"/opt/clem/commands",             // Alternative Linux
		filepath.Join(getUserHome(), ".clem", "builtin-commands"), // Fallback
	}

	for _, dir := range possibleDirs {
		if exists, _ := dirExists(dir); exists {
			return dir
		}
	}

	// Return empty if not found (user/project commands will still work)
	return ""
}
