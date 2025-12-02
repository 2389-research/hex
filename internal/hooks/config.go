// ABOUTME: Configuration loading for .claude/settings.json
// ABOUTME: Loads hook definitions from user and project locations with proper merging

// Package hooks provides lifecycle hooks for Claude Code events.
package hooks

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

// HookConfig represents a single hook configuration
type HookConfig struct {
	Command       string            `json:"command"`
	Description   string            `json:"description,omitempty"`
	Timeout       int               `json:"timeout,omitempty"` // milliseconds, default 5000
	Env           map[string]string `json:"env,omitempty"`
	Match         *MatchConfig      `json:"match,omitempty"`
	IgnoreFailure bool              `json:"ignoreFailure,omitempty"`
	Async         bool              `json:"async,omitempty"`
}

// MatchConfig defines conditions for when a hook should execute
type MatchConfig struct {
	ToolName    string `json:"toolName,omitempty"`
	FilePattern string `json:"filePattern,omitempty"` // regex
	IsSubagent  *bool  `json:"isSubagent,omitempty"`  // pointer to distinguish between false and unset
	Level       string `json:"level,omitempty"`       // for notification events
}

// Settings represents the .claude/settings.json structure
type Settings struct {
	Hooks map[EventType]interface{} `json:"hooks,omitempty"`
}

// Config holds the complete hook configuration for all events
type Config struct {
	hooks map[EventType][]HookConfig
}

// NewConfig creates an empty hook configuration
func NewConfig() *Config {
	return &Config{
		hooks: make(map[EventType][]HookConfig),
	}
}

// LoadConfig loads hook configuration from user and project locations
// Project config overrides user config
func LoadConfig() (*Config, error) {
	config := NewConfig()

	// Load user config from ~/.clem/settings.json
	homeDir, err := os.UserHomeDir()
	if err == nil {
		userConfigPath := filepath.Join(homeDir, ".clem", "settings.json")
		if err := config.loadFromFile(userConfigPath); err != nil {
			// Ignore file not found, but report other errors
			if !os.IsNotExist(err) {
				return nil, fmt.Errorf("load user config: %w", err)
			}
		}
	}

	// Load project config from .claude/settings.json
	projectConfigPath := ".claude/settings.json"
	if err := config.loadFromFile(projectConfigPath); err != nil {
		// Ignore file not found, but report other errors
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("load project config: %w", err)
		}
	}

	return config, nil
}

// loadFromFile loads hook configuration from a JSON file
func (c *Config) loadFromFile(path string) error {
	data, err := os.ReadFile(path) //nolint:gosec // G304 - file paths from trusted config
	if err != nil {
		return err
	}

	var settings Settings
	if err := json.Unmarshal(data, &settings); err != nil {
		return fmt.Errorf("parse settings: %w", err)
	}

	// Process each event type
	for eventType, hookData := range settings.Hooks {
		hooks, err := parseHookData(hookData)
		if err != nil {
			return fmt.Errorf("parse hooks for %s: %w", eventType, err)
		}
		// Override existing hooks for this event type
		c.hooks[eventType] = hooks
	}

	return nil
}

// parseHookData converts the JSON hook data into HookConfig structs
// Handles both single hook objects and arrays of hooks
func parseHookData(data interface{}) ([]HookConfig, error) {
	// Marshal back to JSON then unmarshal to our struct
	// This handles the interface{} -> struct conversion
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	// Try to parse as single hook
	var singleHook HookConfig
	if err := json.Unmarshal(jsonData, &singleHook); err == nil {
		// Check if it's actually a single hook (has command field)
		if singleHook.Command != "" {
			return []HookConfig{singleHook}, nil
		}
	}

	// Try to parse as array of hooks
	var multipleHooks []HookConfig
	if err := json.Unmarshal(jsonData, &multipleHooks); err != nil {
		return nil, fmt.Errorf("invalid hook format: %w", err)
	}

	return multipleHooks, nil
}

// GetHooks returns all hooks configured for a given event type
func (c *Config) GetHooks(eventType EventType) []HookConfig {
	return c.hooks[eventType]
}

// ShouldExecute checks if a hook should execute based on its match configuration
func (hc *HookConfig) ShouldExecute(event *Event) bool {
	if hc.Match == nil {
		return true
	}

	match := hc.Match

	// Check tool name match
	if match.ToolName != "" {
		toolData, ok := event.Data.(ToolUseData)
		if !ok {
			return false
		}
		if toolData.ToolName != match.ToolName {
			return false
		}
	}

	// Check file pattern match
	if match.FilePattern != "" {
		toolData, ok := event.Data.(ToolUseData)
		if !ok {
			return false
		}
		if toolData.FilePath == "" {
			return false
		}
		matched, err := regexp.MatchString(match.FilePattern, toolData.FilePath)
		if err != nil || !matched {
			return false
		}
	}

	// Check subagent match
	if match.IsSubagent != nil {
		var isSubagent bool
		switch data := event.Data.(type) {
		case ToolUseData:
			isSubagent = data.IsSubagent
		case PermissionRequestData:
			isSubagent = data.IsSubagent
		case StopData:
			isSubagent = data.IsSubagent
		default:
			// Event type doesn't have IsSubagent field
			return false
		}
		if isSubagent != *match.IsSubagent {
			return false
		}
	}

	// Check notification level match
	if match.Level != "" {
		notifData, ok := event.Data.(NotificationData)
		if !ok {
			return false
		}
		if notifData.Level != match.Level {
			return false
		}
	}

	return true
}

// GetTimeout returns the timeout in milliseconds, defaulting to 5000 if not set
func (hc *HookConfig) GetTimeout() int {
	if hc.Timeout <= 0 {
		return 5000
	}
	return hc.Timeout
}
