// Package core provides the Anthropic API client and core conversation functionality.
// ABOUTME: Configuration loading with multi-source precedence support
// ABOUTME: Handles env vars, .env files, YAML config, and defaults via Viper
package core

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// ProviderConfig holds provider-specific configuration
type ProviderConfig struct {
	APIKey  string `mapstructure:"api_key"`
	BaseURL string `mapstructure:"base_url"`
}

// Config holds application configuration
type Config struct {
	Provider        string                    `mapstructure:"provider"`
	Model           string                    `mapstructure:"model"`
	ProviderConfigs map[string]ProviderConfig `mapstructure:"providers"`
	DefaultTools    []string                  `mapstructure:"default_tools"`
	PermissionMode  string                    `mapstructure:"permission_mode"`
}

// LoadConfig loads configuration from multiple sources
// Priority (highest to lowest):
// 1. Environment variables (HEX_*)
// 2. Provider-specific env vars (ANTHROPIC_API_KEY, OPENAI_API_KEY, etc.)
// 3. .env file (current directory)
// 4. ~/.hex/config.toml (migrated from config.yaml if needed)
// 5. Defaults
func LoadConfig() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	v := viper.New()

	// Set defaults
	v.SetDefault("provider", "anthropic")
	v.SetDefault("model", DefaultModel)
	v.SetDefault("permission_mode", "ask")
	v.SetDefault("default_tools", []string{"Bash", "Read", "Write", "Edit", "Grep"})

	// Environment variables
	v.SetEnvPrefix("HEX")
	v.AutomaticEnv()
	_ = v.BindEnv("provider")
	_ = v.BindEnv("model")
	_ = v.BindEnv("permission_mode")
	_ = v.BindEnv("default_tools")

	// Check for config file
	configPath := os.Getenv("HEX_CONFIG_PATH")
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		home, err := os.UserHomeDir()
		if err == nil {
			hexDir := filepath.Join(home, ".hex")

			// Check if config.yaml exists and migrate to config.toml
			yamlPath := filepath.Join(hexDir, "config.yaml")
			tomlPath := filepath.Join(hexDir, "config.toml")
			if _, err := os.Stat(yamlPath); err == nil {
				if _, err := os.Stat(tomlPath); os.IsNotExist(err) {
					if err := migrateYAMLToTOML(yamlPath, tomlPath); err != nil {
						fmt.Fprintf(os.Stderr, "Warning: failed to migrate config: %v\n", err)
					}
				}
			}

			v.SetConfigName("config")
			v.SetConfigType("toml")
			v.AddConfigPath(hexDir)
			v.AddConfigPath(".")
		}
	}

	// Read config file (ignore error if file doesn't exist)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file doesn't exist - create a default one
			home, err := os.UserHomeDir()
			if err == nil {
				hexDir := filepath.Join(home, ".hex")
				tomlPath := filepath.Join(hexDir, "config.toml")

				// Create directory if needed
				if err := os.MkdirAll(hexDir, 0755); err != nil {
					// Log the mkdir failure specifically
					fmt.Fprintf(os.Stderr, "Note: Could not create config directory: %v\n", err)
				} else if err := createDefaultConfig(tomlPath); err != nil {
					// Don't fail if we can't create default config
					fmt.Fprintf(os.Stderr, "Note: Could not create default config: %v\n", err)
				}
			}
		} else {
			return nil, fmt.Errorf("read config: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	// Initialize provider configs if nil
	if cfg.ProviderConfigs == nil {
		cfg.ProviderConfigs = make(map[string]ProviderConfig)
	}

	// Check standard provider env vars (these override config file)
	providers := map[string]string{
		"anthropic":  os.Getenv("ANTHROPIC_API_KEY"),
		"openai":     os.Getenv("OPENAI_API_KEY"),
		"gemini":     os.Getenv("GEMINI_API_KEY"),
		"openrouter": os.Getenv("OPENROUTER_API_KEY"),
	}

	for name, envKey := range providers {
		if envKey != "" {
			if pc, ok := cfg.ProviderConfigs[name]; ok {
				pc.APIKey = envKey // Override with env var
				cfg.ProviderConfigs[name] = pc
			} else {
				cfg.ProviderConfigs[name] = ProviderConfig{APIKey: envKey}
			}
		}
	}

	return &cfg, nil
}

// ValidateProvider ensures the selected provider is configured
func (c *Config) ValidateProvider() error {
	pc, ok := c.ProviderConfigs[c.Provider]
	if !ok {
		return fmt.Errorf("provider %s not configured", c.Provider)
	}
	if pc.APIKey == "" {
		return fmt.Errorf("API key not set for provider %s", c.Provider)
	}
	return nil
}

// GetProviderConfig returns the config for the selected provider
func (c *Config) GetProviderConfig() (ProviderConfig, error) {
	pc, ok := c.ProviderConfigs[c.Provider]
	if !ok {
		return ProviderConfig{}, fmt.Errorf("provider %s not configured", c.Provider)
	}
	return pc, nil
}

// GetAPIKey returns the API key for backward compatibility
// Deprecated: Use GetProviderConfig instead
func (c *Config) GetAPIKey() (string, error) {
	pc, err := c.GetProviderConfig()
	if err != nil {
		return "", fmt.Errorf("API key not configured. Set %s_API_KEY or run 'hex setup-token'", c.Provider)
	}
	return pc.APIKey, nil
}

// createDefaultConfig creates a default config.toml with helpful comments
func createDefaultConfig(path string) error {
	defaultConfig := `# Hex Configuration File
# This file is automatically created on first run
# Priority: env vars > .env file > this config > defaults

# Default provider (anthropic, openai, gemini, openrouter)
provider = "anthropic"

# Default model (optional - can be overridden with --model flag)
# Only Anthropic has a default; other providers require --model flag
# model = "claude-sonnet-4-5-20250929"

# Permission mode for tool execution (ask, auto)
permission_mode = "ask"

# Default tools available in print mode
default_tools = ["Bash", "Read", "Write", "Edit", "Grep", "Glob"]

# Provider configurations
# API keys can also be set via environment variables:
#   ANTHROPIC_API_KEY, OPENAI_API_KEY, GEMINI_API_KEY, OPENROUTER_API_KEY

[providers.anthropic]
# api_key = "sk-ant-..."
# base_url = "https://api.anthropic.com"  # Optional custom endpoint

[providers.openai]
# api_key = "sk-..."
# base_url = "https://api.openai.com/v1"  # Optional custom endpoint

[providers.gemini]
# api_key = "..."
# base_url = "https://generativelanguage.googleapis.com/v1beta"

[providers.openrouter]
# api_key = "sk-or-..."
# base_url = "https://openrouter.ai/api/v1"

# Current models (as of 2025-12-08):
# Anthropic: claude-sonnet-4-5-20250929, claude-opus-4-5-20251101, claude-haiku-4-5-20251001
# OpenAI: gpt-5.1, gpt-5.1-codex, gpt-5.1-codex-mini
# Gemini: gemini-2.5-pro, gemini-2.5-flash, gemini-pro-latest
# OpenRouter: anthropic/claude-sonnet-4-5, openai/gpt-5.1, google/gemini-2.5-pro
`

	// Use O_EXCL to atomically create file only if it doesn't exist
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		if os.IsExist(err) {
			return nil // File exists, someone else created it
		}
		return fmt.Errorf("create config: %w", err)
	}
	defer func() { _ = f.Close() }()

	if _, err := f.WriteString(defaultConfig); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	fmt.Fprintf(os.Stderr, "✅ Created default config at %s\n", path)
	return nil
}
