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

// Config holds application configuration
type Config struct {
	APIKey         string   `mapstructure:"api_key"`
	Model          string   `mapstructure:"model"`
	DefaultTools   []string `mapstructure:"default_tools"`
	PermissionMode string   `mapstructure:"permission_mode"`
}

// LoadConfig loads configuration from multiple sources
// Priority (highest to lowest):
// 1. Environment variables (CLEM_*)
// 2. .env file (current directory)
// 3. ~/.clem/config.yaml
// 4. Defaults
func LoadConfig() (*Config, error) {
	// Load .env file if it exists (don't error if missing)
	_ = godotenv.Load()

	v := viper.New()

	// Set defaults
	v.SetDefault("model", DefaultModel)
	v.SetDefault("permission_mode", "ask")
	v.SetDefault("default_tools", []string{"Bash", "Read", "Write", "Edit", "Grep"})

	// Environment variables
	v.SetEnvPrefix("CLEM")
	v.AutomaticEnv()
	// Bind specific keys to handle underscore conversion
	_ = v.BindEnv("api_key")
	_ = v.BindEnv("model")
	_ = v.BindEnv("permission_mode")
	_ = v.BindEnv("default_tools")

	// Config file
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	// Check for custom config path
	if configPath := os.Getenv("CLEM_CONFIG_PATH"); configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		// Add search paths
		v.AddConfigPath(".") // Current directory
		home, err := os.UserHomeDir()
		if err == nil {
			clemDir := filepath.Join(home, ".clem")
			v.AddConfigPath(clemDir)
		}
	}

	// Read config file (ignore error if file doesn't exist)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("read config: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return &cfg, nil
}

// GetAPIKey returns the API key from config or environment
func (c *Config) GetAPIKey() (string, error) {
	if c.APIKey == "" {
		return "", fmt.Errorf("API key not configured. Set CLEM_API_KEY or run 'clem setup-token'")
	}
	return c.APIKey, nil
}
