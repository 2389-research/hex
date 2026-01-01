// ABOUTME: Tests for the setup wizard first-run detection and API key validation
// ABOUTME: Ensures proper triggering of wizard and provider-specific key format checks

package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsFirstRun_NoConfigNoEnv(t *testing.T) {
	// Clear all provider env vars
	envVars := []string{
		"ANTHROPIC_API_KEY",
		"OPENAI_API_KEY",
		"GEMINI_API_KEY",
		"OPENROUTER_API_KEY",
		"HEX_API_KEY",
	}

	oldValues := make(map[string]string)
	for _, env := range envVars {
		oldValues[env] = os.Getenv(env)
		os.Unsetenv(env)
	}
	defer func() {
		for env, val := range oldValues {
			if val != "" {
				os.Setenv(env, val)
			}
		}
	}()

	// Create temp home dir without config
	tmpHome := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", oldHome)

	if !IsFirstRun() {
		t.Error("IsFirstRun() should return true when no config and no env vars")
	}
}

func TestIsFirstRun_WithEnvVar(t *testing.T) {
	// Set one env var
	oldVal := os.Getenv("ANTHROPIC_API_KEY")
	os.Setenv("ANTHROPIC_API_KEY", "sk-ant-test-key")
	defer func() {
		if oldVal != "" {
			os.Setenv("ANTHROPIC_API_KEY", oldVal)
		} else {
			os.Unsetenv("ANTHROPIC_API_KEY")
		}
	}()

	// Clear other env vars
	for _, env := range []string{"OPENAI_API_KEY", "GEMINI_API_KEY", "OPENROUTER_API_KEY", "HEX_API_KEY"} {
		os.Unsetenv(env)
	}

	if IsFirstRun() {
		t.Error("IsFirstRun() should return false when ANTHROPIC_API_KEY is set")
	}
}

func TestIsFirstRun_WithConfigFile(t *testing.T) {
	// Clear all provider env vars
	envVars := []string{
		"ANTHROPIC_API_KEY",
		"OPENAI_API_KEY",
		"GEMINI_API_KEY",
		"OPENROUTER_API_KEY",
		"HEX_API_KEY",
	}
	for _, env := range envVars {
		os.Unsetenv(env)
	}

	// Create temp home with config
	tmpHome := t.TempDir()
	hexDir := filepath.Join(tmpHome, ".hex")
	if err := os.MkdirAll(hexDir, 0700); err != nil {
		t.Fatalf("Failed to create .hex dir: %v", err)
	}

	configPath := filepath.Join(hexDir, "config.toml")
	if err := os.WriteFile(configPath, []byte("provider = \"anthropic\"\n"), 0600); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", oldHome)

	if IsFirstRun() {
		t.Error("IsFirstRun() should return false when config.toml exists")
	}
}

func TestValidateAPIKey_Anthropic(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		wantErr bool
	}{
		{"valid prefix", "sk-ant-api-12345", false},
		{"invalid prefix", "invalid-key", true},
		{"empty key", "", true},
		{"whitespace only", "   ", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAPIKey("anthropic", tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateAPIKey() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateAPIKey_OpenAI(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		wantErr bool
	}{
		{"valid prefix", "sk-proj-12345", false},
		{"valid simple", "sk-12345", false},
		{"invalid prefix", "invalid-key", true},
		{"empty key", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAPIKey("openai", tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateAPIKey() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateAPIKey_OpenRouter(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		wantErr bool
	}{
		{"valid prefix", "sk-or-v1-12345", false},
		{"invalid prefix", "sk-12345", true},
		{"empty key", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAPIKey("openrouter", tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateAPIKey() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateAPIKey_Gemini(t *testing.T) {
	// Gemini doesn't have prefix validation, just non-empty
	tests := []struct {
		name    string
		key     string
		wantErr bool
	}{
		{"any valid key", "AIzaSy12345", false},
		{"simple key", "any-key-works", false},
		{"empty key", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAPIKey("gemini", tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateAPIKey() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateAPIKey_Ollama(t *testing.T) {
	// Ollama doesn't need API key validation
	// This case shouldn't be called in practice since Ollama skips API key input
	err := validateAPIKey("ollama", "any-key")
	if err != nil {
		t.Errorf("validateAPIKey() for ollama should not error, got: %v", err)
	}
}

func TestNewWizardModel(t *testing.T) {
	model := NewWizardModel()

	if model.state != stateWelcome {
		t.Errorf("Initial state should be stateWelcome, got %v", model.state)
	}

	if model.provider != "" {
		t.Errorf("Initial provider should be empty, got %q", model.provider)
	}

	if model.apiKey != "" {
		t.Errorf("Initial apiKey should be empty, got %q", model.apiKey)
	}
}
