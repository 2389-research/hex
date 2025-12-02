package main

import (
	"github.com/harper/clem/internal/logging"
	"github.com/harper/clem/internal/plugins"
)

// initializePlugins loads all enabled plugins and returns the registry
func initializePlugins() (*plugins.Registry, error) {
	logging.Debug("Initializing plugin system")

	// Create plugin registry with default paths
	registry, err := plugins.DefaultRegistry()
	if err != nil {
		logging.WarnWith("Failed to create plugin registry", "error", err.Error())
		return nil, err
	}

	// Create activation context (could be enhanced to detect project type, languages, etc.)
	context := &plugins.ActivationContext{
		Languages:    []string{}, // TODO: Detect from project
		Files:        []string{}, // TODO: Scan project for marker files
		ProjectTypes: []string{}, // TODO: Detect project type
	}

	// Load all enabled plugins
	if err := registry.LoadAll(context); err != nil {
		// Log warning but don't fail - continue with loaded plugins
		logging.WarnWith("Failed to load some plugins", "error", err.Error())
	}

	count := registry.Count()
	if count > 0 {
		logging.InfoWith("Loaded plugins", "count", count)
	} else {
		logging.Debug("No plugins loaded")
	}

	return registry, nil
}

// getPluginSkillPaths returns all skill file paths from enabled plugins
func getPluginSkillPaths(registry *plugins.Registry) []string {
	if registry == nil {
		return nil
	}
	return registry.GetSkillPaths()
}

// getPluginCommandPaths returns all command file paths from enabled plugins
func getPluginCommandPaths(registry *plugins.Registry) []string {
	if registry == nil {
		return nil
	}
	return registry.GetCommandPaths()
}
