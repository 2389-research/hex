package forms

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewSettingsForm(t *testing.T) {
	form := NewSettingsForm("claude-3-5-sonnet-20241022", "sk-ant-test", "/tmp/config.toml")

	if form == nil {
		t.Fatal("expected form to be created, got nil")
	}

	if form.result.Model != "claude-3-5-sonnet-20241022" {
		t.Errorf("expected model to be set, got %s", form.result.Model)
	}

	if form.result.APIKey != "sk-ant-test" {
		t.Errorf("expected API key to be set, got %s", form.result.APIKey)
	}

	if form.theme == nil {
		t.Error("expected theme to be initialized, got nil")
	}

	if len(form.availableModels) == 0 {
		t.Error("expected available models to be populated")
	}
}

func TestBuildModelOptions(t *testing.T) {
	form := NewSettingsForm("", "", "")

	options := form.buildModelOptions()

	if len(options) == 0 {
		t.Error("expected model options to be built")
	}

	// Check that options have labels and values
	for i, opt := range options {
		if opt.Value == "" {
			t.Errorf("option %d has empty value", i)
		}
	}
}

func TestFormatModelLabel(t *testing.T) {
	form := NewSettingsForm("", "", "")

	tests := []struct {
		name     string
		modelID  string
		contains []string
	}{
		{
			name:     "claude 3.5 sonnet",
			modelID:  "claude-3-5-sonnet-20241022",
			contains: []string{"3.5", "Sonnet"},
		},
		{
			name:     "claude 3.5 haiku",
			modelID:  "claude-3-5-haiku-20241022",
			contains: []string{"3.5", "Haiku"},
		},
		{
			name:     "claude 3 opus",
			modelID:  "claude-3-opus-20240229",
			contains: []string{"Opus"},
		},
		{
			name:    "unknown model",
			modelID: "unknown-model-123",
			// Should return the model ID as-is
			contains: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			label := form.formatModelLabel(tt.modelID)

			// Just verify we got some output
			if label == "" {
				t.Error("expected non-empty label")
			}
		})
	}
}

func TestMaskAPIKey(t *testing.T) {
	form := NewSettingsForm("", "", "")

	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{
			name:     "short key",
			key:      "short",
			expected: "sk-ant-***",
		},
		{
			name:     "full key",
			key:      "sk-ant-api03-1234567890abcdefghijklmnopqrstuvwxyz",
			expected: "sk-ant-api03...wxyz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			masked := form.maskAPIKey(tt.key)

			if masked == "" {
				t.Error("expected masked key, got empty string")
			}

			// Short keys should be completely masked
			if len(tt.key) <= 12 && masked != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, masked)
			}

			// Long keys should have visible prefix and suffix
			if len(tt.key) > 12 {
				if !strings.Contains(masked, "...") {
					t.Error("expected ellipsis in masked key")
				}
			}
		})
	}
}

func TestGetModelDisplayName(t *testing.T) {
	form := NewSettingsForm("", "", "")

	tests := []struct {
		modelID     string
		displayName string
	}{
		{"claude-3-5-sonnet-20241022", "Claude 3.5 Sonnet"},
		{"claude-3-5-haiku-20241022", "Claude 3.5 Haiku"},
		{"claude-3-opus-20240229", "Claude 3 Opus"},
		{"claude-3-sonnet-20240229", "Claude 3 Sonnet"},
		{"claude-3-haiku-20240307", "Claude 3 Haiku"},
		{"unknown-model", "unknown-model"},
	}

	for _, tt := range tests {
		t.Run(tt.modelID, func(t *testing.T) {
			displayName := form.getModelDisplayName(tt.modelID)
			if displayName != tt.displayName {
				t.Errorf("expected %q, got %q", tt.displayName, displayName)
			}
		})
	}
}

func TestBuildConfirmationText(t *testing.T) {
	form := NewSettingsForm("claude-3-5-sonnet-20241022", "sk-ant-test", "/tmp/config.toml")
	form.result.Temperature = 0.8
	form.result.MaxTokens = 2048

	text := form.buildConfirmationText()

	// Should contain key information
	if !strings.Contains(text, "Model:") {
		t.Error("expected confirmation to contain Model")
	}

	if !strings.Contains(text, "API Key:") {
		t.Error("expected confirmation to contain API Key")
	}

	if !strings.Contains(text, "Temperature:") {
		t.Error("expected confirmation to contain Temperature")
	}

	if !strings.Contains(text, "Max Tokens:") {
		t.Error("expected confirmation to contain Max Tokens")
	}

	if !strings.Contains(text, "0.8") {
		t.Error("expected confirmation to show temperature value")
	}

	if !strings.Contains(text, "2048") {
		t.Error("expected confirmation to show max tokens value")
	}
}

func TestSaveToFile(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.toml")

	form := NewSettingsForm("claude-3-5-sonnet-20241022", "sk-ant-test123", configPath)
	form.result.Temperature = 0.9
	form.result.MaxTokens = 3000

	// Save to file
	err := form.SaveToFile()
	if err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	// Verify file exists
	if _, statErr := os.Stat(configPath); os.IsNotExist(statErr) {
		t.Error("config file was not created")
	}

	// Read and verify content
	// #nosec G304 - this is a test file reading from a temp directory
	content, readErr := os.ReadFile(configPath)
	if readErr != nil {
		t.Fatalf("failed to read config file: %v", readErr)
	}

	contentStr := string(content)

	// Verify key values are present
	if !strings.Contains(contentStr, "claude-3-5-sonnet-20241022") {
		t.Error("config should contain model ID")
	}

	if !strings.Contains(contentStr, "sk-ant-test123") {
		t.Error("config should contain API key")
	}

	if !strings.Contains(contentStr, "0.9") {
		t.Error("config should contain temperature")
	}

	if !strings.Contains(contentStr, "3000") {
		t.Error("config should contain max tokens")
	}
}

func TestSaveToFileCreatesDirectory(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "subdir", "nested", "config.toml")

	form := NewSettingsForm("claude-3-5-sonnet-20241022", "sk-ant-test", configPath)

	// Save to file - should create nested directories
	err := form.SaveToFile()
	if err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("config file was not created")
	}

	// Verify directories were created
	dir := filepath.Dir(configPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error("directories were not created")
	}
}

func TestSettingsGetDraculaTheme(t *testing.T) {
	form := NewSettingsForm("", "", "")
	theme := form.getDraculaTheme()

	if theme == nil {
		t.Fatal("expected theme to be created, got nil")
	}

	// Verify theme has required style groups
	_ = theme.Focused.Title.String()
	_ = theme.Focused.TextInput.Cursor.String()
}
