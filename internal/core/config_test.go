package core_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/harper/hex/internal/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	// Create temp config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configYAML := `api_key: test-key-123
model: claude-sonnet-4-5-20250929
`
	err := os.WriteFile(configPath, []byte(configYAML), 0600)
	require.NoError(t, err)

	// Set config path
	_ = os.Setenv("HEX_CONFIG_PATH", configPath)
	defer func() { _ = os.Unsetenv("HEX_CONFIG_PATH") }()

	// Load config
	cfg, err := core.LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, "test-key-123", cfg.APIKey)
	assert.Equal(t, "claude-sonnet-4-5-20250929", cfg.Model)
}

func TestConfigFromEnv(t *testing.T) {
	// Clean up any existing config path
	_ = os.Unsetenv("HEX_CONFIG_PATH")

	_ = os.Setenv("HEX_API_KEY", "env-key-456")
	_ = os.Setenv("HEX_MODEL", "claude-opus-4-5-20250929")
	defer func() { _ = os.Unsetenv("HEX_API_KEY") }()
	defer func() { _ = os.Unsetenv("HEX_MODEL") }()

	cfg, err := core.LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, "env-key-456", cfg.APIKey)
	assert.Equal(t, "claude-opus-4-5-20250929", cfg.Model)
}

func TestConfigPrecedence(t *testing.T) {
	// Env var should override config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configYAML := `api_key: file-key`
	err := os.WriteFile(configPath, []byte(configYAML), 0600)
	require.NoError(t, err)

	_ = os.Setenv("HEX_CONFIG_PATH", configPath)
	_ = os.Setenv("HEX_API_KEY", "env-key")
	defer func() { _ = os.Unsetenv("HEX_CONFIG_PATH") }()
	defer func() { _ = os.Unsetenv("HEX_API_KEY") }()

	cfg, err := core.LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, "env-key", cfg.APIKey, "env var should override file")
}

func TestConfigDefaults(t *testing.T) {
	// Clean all env vars that might affect the test
	_ = os.Unsetenv("HEX_CONFIG_PATH")
	_ = os.Unsetenv("HEX_API_KEY")
	_ = os.Unsetenv("HEX_MODEL")
	_ = os.Unsetenv("HEX_PERMISSION_MODE")
	_ = os.Unsetenv("HEX_DEFAULT_TOOLS")

	cfg, err := core.LoadConfig()
	require.NoError(t, err)

	// Should have defaults
	assert.Equal(t, "claude-sonnet-4-5-20250929", cfg.Model)
	assert.Equal(t, "ask", cfg.PermissionMode)
	assert.NotEmpty(t, cfg.DefaultTools)
}

func TestConfigGetAPIKey(t *testing.T) {
	t.Run("with API key", func(t *testing.T) {
		cfg := &core.Config{
			APIKey: "test-key",
		}

		key, err := cfg.GetAPIKey()
		require.NoError(t, err)
		assert.Equal(t, "test-key", key)
	})

	t.Run("without API key", func(t *testing.T) {
		cfg := &core.Config{
			APIKey: "",
		}

		_, err := cfg.GetAPIKey()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "API key not configured")
	})
}

func TestConfigFromDotEnv(t *testing.T) {
	// Create temp directory with .env file
	tmpDir := t.TempDir()
	dotEnvPath := filepath.Join(tmpDir, ".env")

	dotEnvContent := `HEX_API_KEY=dotenv-key-789
HEX_MODEL=claude-sonnet-4-5-20250929
`
	err := os.WriteFile(dotEnvPath, []byte(dotEnvContent), 0600)
	require.NoError(t, err)

	// Change to temp directory so LoadConfig finds the .env file
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tmpDir)

	// Clean env vars
	_ = os.Unsetenv("HEX_API_KEY")
	_ = os.Unsetenv("HEX_MODEL")
	_ = os.Unsetenv("HEX_CONFIG_PATH")

	cfg, err := core.LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, "dotenv-key-789", cfg.APIKey)
	assert.Equal(t, "claude-sonnet-4-5-20250929", cfg.Model)
}
