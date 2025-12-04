//go:build integration

package integration_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPhase1Integration(t *testing.T) {
	// Build binary
	buildCmd := exec.Command("go", "build", "-o", "clem-test", "./cmd/clem")
	buildCmd.Dir = "../.."
	err := buildCmd.Run()
	require.NoError(t, err, "failed to build test binary")
	defer func() { _ = os.Remove("../../clem-test")

	clemBin := "../../clem-test"

	t.Run("version flag", func(t *testing.T) {
		cmd := exec.Command(clemBin, "--version")
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "version command should succeed")
		assert.Contains(t, string(output), "0.1.0", "version should be 0.1.0")
	})

	t.Run("help flag", func(t *testing.T) {
		cmd := exec.Command(clemBin, "--help")
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "help command should succeed")
		assert.Contains(t, string(output), "Clem", "help should mention Clem")
		assert.Contains(t, string(output), "--print", "help should mention --print flag")
	})

	t.Run("setup-token command", func(t *testing.T) {
		tmpHome := t.TempDir()
		cmd := exec.Command(clemBin, "setup-token", "test-key-xyz")
		cmd.Env = append(os.Environ(), "HOME="+tmpHome)
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "setup-token should succeed")
		assert.Contains(t, string(output), "✓", "should show success checkmark")

		// Verify file was created
		configPath := filepath.Join(tmpHome, ".jeff", "config.yaml")
		assert.FileExists(t, configPath, "config.yaml should be created")
	})

	t.Run("doctor command", func(t *testing.T) {
		tmpHome := t.TempDir()

		// Setup first
		setupCmd := exec.Command(clemBin, "setup-token", "test-key")
		setupCmd.Env = append(os.Environ(), "HOME="+tmpHome)
		err := setupCmd.Run()
		require.NoError(t, err, "setup should succeed before doctor check")

		// Run doctor
		doctorCmd := exec.Command(clemBin, "doctor")
		doctorCmd.Env = append(os.Environ(), "HOME="+tmpHome)
		output, err := doctorCmd.CombinedOutput()
		require.NoError(t, err, "doctor command should succeed")
		assert.Contains(t, string(output), "✓ All checks passed", "doctor should report all checks passed")
	})

	t.Run("print mode error without key", func(t *testing.T) {
		tmpHome := t.TempDir()
		cmd := exec.Command(clemBin, "--print", "test")
		cmd.Env = append(os.Environ(), "HOME="+tmpHome)
		output, err := cmd.CombinedOutput()
		assert.Error(t, err, "print mode should error without API key")
		assert.Contains(t, string(output), "API key not configured", "should show API key error message")
	})
}

func TestPrintModeWithRealAPI(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping real API test in short mode")
	}

	apiKey := os.Getenv("PAGEN_API_KEY")
	if apiKey == "" {
		t.Skip("PAGEN_API_KEY not set, skipping real API test")
	}

	// Build binary
	buildCmd := exec.Command("go", "build", "-o", "clem-test", "./cmd/clem")
	buildCmd.Dir = "../.."
	err := buildCmd.Run()
	require.NoError(t, err, "failed to build test binary")
	defer func() { _ = os.Remove("../../clem-test")

	clemBin := "../../clem-test"

	// Test print mode with real API
	cmd := exec.Command(clemBin, "--print", "Say hello in exactly 3 words")
	cmd.Env = append(os.Environ(), "PAGEN_API_KEY="+apiKey)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "print mode with real API should succeed")

	result := strings.TrimSpace(string(output))
	assert.NotEmpty(t, result, "API response should not be empty")
	t.Logf("API Response: %s", result)
}
