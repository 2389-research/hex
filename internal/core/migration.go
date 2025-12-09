// internal/core/migration.go
// ABOUTME: Migrates existing config.yaml to config.toml format
// ABOUTME: Preserves all existing settings and adds provider structure
package core

import (
	"fmt"
	"os"

	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

// migrateYAMLToTOML converts config.yaml to config.toml
func migrateYAMLToTOML(yamlPath, tomlPath string) error {
	// Read YAML
	yamlData, err := os.ReadFile(yamlPath)
	if err != nil {
		return fmt.Errorf("read yaml: %w", err)
	}

	var yamlCfg map[string]interface{}
	if unmarshalErr := yaml.Unmarshal(yamlData, &yamlCfg); unmarshalErr != nil {
		return fmt.Errorf("unmarshal yaml: %w", unmarshalErr)
	}

	// Convert to new structure
	tomlCfg := make(map[string]interface{})

	// Copy simple fields
	if v, ok := yamlCfg["model"]; ok {
		tomlCfg["model"] = v
	}
	if v, ok := yamlCfg["permission_mode"]; ok {
		tomlCfg["permission_mode"] = v
	}
	if v, ok := yamlCfg["default_tools"]; ok {
		tomlCfg["default_tools"] = v
	}

	// Set default provider
	tomlCfg["provider"] = "anthropic"

	// Migrate API key to provider structure
	providers := make(map[string]map[string]string)
	if apiKey, ok := yamlCfg["api_key"].(string); ok && apiKey != "" {
		providers["anthropic"] = map[string]string{"api_key": apiKey}
	}
	tomlCfg["providers"] = providers

	// Write TOML
	tomlData, err := toml.Marshal(tomlCfg)
	if err != nil {
		return fmt.Errorf("marshal toml: %w", err)
	}

	if err := os.WriteFile(tomlPath, tomlData, 0644); err != nil {
		return fmt.Errorf("write toml: %w", err)
	}

	fmt.Printf("Migrated %s to %s\n", yamlPath, tomlPath)
	return nil
}
