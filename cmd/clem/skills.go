package main

import (
	"os"
	"path/filepath"

	"github.com/harper/clem/internal/logging"
	"github.com/harper/clem/internal/skills"
	"github.com/harper/clem/internal/tools"
)

// initializeSkills loads skills from all sources and returns registry and tool
// Accepts optional plugin skill paths
func initializeSkills(pluginSkillPaths []string) (*skills.Registry, tools.Tool) {
	logging.Debug("Initializing skills system")

	// Create skill loader
	loader := skills.NewLoader()

	// Set builtin skills directory (relative to binary location)
	loader.BuiltinDir = findBuiltinSkillsDir()

	// Add plugin skill paths
	loader.PluginPaths = pluginSkillPaths

	// Load all skills
	loadedSkills, err := loader.LoadAll()
	if err != nil {
		logging.WarnWith("Failed to load some skills", "error", err.Error())
	}

	// Create and populate registry
	registry := skills.NewRegistry()
	for _, skill := range loadedSkills {
		if err := registry.Register(skill); err != nil {
			logging.WarnWith("Failed to register skill", "name", skill.Name, "error", err.Error())
		} else {
			logging.DebugWith("Registered skill", "name", skill.Name, "source", skill.Source, "priority", skill.Priority)
		}
	}

	// Create tool adapter
	toolAdapter := skills.NewToolAdapter(registry)

	return registry, toolAdapter
}

// findBuiltinSkillsDir locates the built-in skills directory
func findBuiltinSkillsDir() string {
	// Try several possible locations
	possibleDirs := []string{
		"./skills",                     // Development: same directory as binary
		"../skills",                    // Development: parent directory
		"/usr/local/share/clem/skills", // Linux install
		"/opt/clem/skills",             // Alternative Linux
		filepath.Join(getUserHome(), ".clem", "builtin-skills"), // Fallback
	}

	for _, dir := range possibleDirs {
		if exists, _ := dirExists(dir); exists {
			return dir
		}
	}

	// Return empty if not found (user/project skills will still work)
	return ""
}

// getUserHome gets user home directory
func getUserHome() string {
	home, _ := os.UserHomeDir()
	return home
}

// dirExists checks if directory exists
func dirExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}
