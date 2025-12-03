// Package forms provides beautiful huh-based forms for the hex TUI.
// ABOUTME: Huh-based settings wizard for interactive configuration
// ABOUTME: Provides multi-step settings form with model selection and API configuration
package forms

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/2389-research/hex/internal/ui/theme"
	"github.com/charmbracelet/huh"
)

// SettingsFormResult contains the result of the settings wizard
type SettingsFormResult struct {
	Model       string
	APIKey      string
	Temperature float64
	MaxTokens   int
	Preferences map[string]interface{}
	Cancelled   bool
}

// Internal string fields for huh form binding
type settingsFormFields struct {
	temperature string
	maxTokens   string
}

// SettingsForm creates a multi-step settings wizard using huh
type SettingsForm struct {
	result *SettingsFormResult
	theme  *theme.Theme

	// Available models for selection
	availableModels []string

	// Current config file path
	configPath string

	// Internal string fields for form binding
	fields settingsFormFields
}

// NewSettingsForm creates a new settings wizard form
func NewSettingsForm(currentModel, currentAPIKey string, configPath string) *SettingsForm {
	// Default available models (can be customized)
	availableModels := []string{
		"claude-3-5-sonnet-20241022",
		"claude-3-5-haiku-20241022",
		"claude-3-opus-20240229",
		"claude-3-sonnet-20240229",
		"claude-3-haiku-20240307",
	}

	result := &SettingsFormResult{
		Model:       currentModel,
		APIKey:      currentAPIKey,
		Temperature: 1.0,
		MaxTokens:   4096,
		Preferences: make(map[string]interface{}),
	}

	return &SettingsForm{
		result:          result,
		theme:           theme.DraculaTheme(),
		availableModels: availableModels,
		configPath:      configPath,
		fields: settingsFormFields{
			temperature: "1.0",
			maxTokens:   "4096",
		},
	}
}

// Run displays the multi-step settings wizard and returns the result
func (f *SettingsForm) Run() (*SettingsFormResult, error) {
	// Create multi-step form
	form := huh.NewForm(
		// Step 1: Model Selection
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("🤖 Select AI Model").
				Description("Choose which Claude model to use for conversations").
				Options(f.buildModelOptions()...).
				Value(&f.result.Model),
		),

		// Step 2: API Configuration
		huh.NewGroup(
			huh.NewInput().
				Title("🔑 Anthropic API Key").
				Description("Enter your API key (will be masked)").
				Value(&f.result.APIKey).
				EchoMode(huh.EchoModePassword).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("API key is required")
					}
					if !strings.HasPrefix(s, "sk-ant-") {
						return fmt.Errorf("API key should start with 'sk-ant-'")
					}
					return nil
				}),
		),

		// Step 3: Model Parameters
		huh.NewGroup(
			huh.NewInput().
				Title("🌡 Temperature").
				Description("Randomness (0.0 - 1.0, default: 1.0)").
				Value(&f.fields.temperature).
				Validate(func(s string) error {
					return f.validateTemperature(s)
				}),

			huh.NewInput().
				Title("📏 Max Tokens").
				Description("Maximum response length (256 - 8192, default: 4096)").
				Value(&f.fields.maxTokens).
				Validate(func(s string) error {
					return f.validateMaxTokens(s)
				}),
		),

		// Step 4: Confirmation
		huh.NewGroup(
			huh.NewConfirm().
				Title("💾 Save Settings?").
				Description(f.buildConfirmationText()).
				Affirmative("Save").
				Negative("Cancel").
				Value(&f.result.Cancelled), // Note: huh.Confirm uses inverted logic
		),
	).WithTheme(f.getDraculaTheme())

	// Run the form
	err := form.Run()
	if err != nil {
		return nil, err
	}

	// Invert the cancelled flag (huh uses inverted logic)
	f.result.Cancelled = !f.result.Cancelled

	// Parse string fields to result struct
	if err := f.parseFields(); err != nil {
		return nil, fmt.Errorf("parse settings: %w", err)
	}

	return f.result, nil
}

// validateTemperature validates the temperature input
func (f *SettingsForm) validateTemperature(s string) error {
	var temp float64
	_, err := fmt.Sscanf(s, "%f", &temp)
	if err != nil {
		return fmt.Errorf("temperature must be a number")
	}
	if temp < 0.0 || temp > 1.0 {
		return fmt.Errorf("temperature must be between 0.0 and 1.0")
	}
	return nil
}

// validateMaxTokens validates the max tokens input
func (f *SettingsForm) validateMaxTokens(s string) error {
	var tokens int
	_, err := fmt.Sscanf(s, "%d", &tokens)
	if err != nil {
		return fmt.Errorf("max tokens must be an integer")
	}
	if tokens < 256 || tokens > 8192 {
		return fmt.Errorf("max tokens must be between 256 and 8192")
	}
	return nil
}

// parseFields parses string fields into the result struct
func (f *SettingsForm) parseFields() error {
	// Parse temperature
	var temp float64
	_, err := fmt.Sscanf(f.fields.temperature, "%f", &temp)
	if err != nil {
		return fmt.Errorf("invalid temperature: %w", err)
	}
	f.result.Temperature = temp

	// Parse max tokens
	var tokens int
	_, err = fmt.Sscanf(f.fields.maxTokens, "%d", &tokens)
	if err != nil {
		return fmt.Errorf("invalid max tokens: %w", err)
	}
	f.result.MaxTokens = tokens

	return nil
}

// buildModelOptions creates huh options for available models
func (f *SettingsForm) buildModelOptions() []huh.Option[string] {
	options := make([]huh.Option[string], 0, len(f.availableModels))

	for _, model := range f.availableModels {
		label := f.formatModelLabel(model)
		options = append(options, huh.NewOption(label, model))
	}

	return options
}

// formatModelLabel creates a friendly label for a model ID
func (f *SettingsForm) formatModelLabel(modelID string) string {
	// Extract model name and add description
	switch {
	case strings.Contains(modelID, "sonnet") && strings.Contains(modelID, "3-5"):
		return f.theme.Emphasized.Render("Claude 3.5 Sonnet") + f.theme.Muted.Render(" - Most capable, balanced")
	case strings.Contains(modelID, "haiku") && strings.Contains(modelID, "3-5"):
		return f.theme.Emphasized.Render("Claude 3.5 Haiku") + f.theme.Muted.Render(" - Fast and efficient")
	case strings.Contains(modelID, "opus"):
		return f.theme.Emphasized.Render("Claude 3 Opus") + f.theme.Muted.Render(" - Powerful, top performance")
	case strings.Contains(modelID, "sonnet"):
		return f.theme.Emphasized.Render("Claude 3 Sonnet") + f.theme.Muted.Render(" - Balanced performance")
	case strings.Contains(modelID, "haiku"):
		return f.theme.Emphasized.Render("Claude 3 Haiku") + f.theme.Muted.Render(" - Fast and cost-effective")
	default:
		return modelID
	}
}

// buildConfirmationText creates the confirmation message
func (f *SettingsForm) buildConfirmationText() string {
	var b strings.Builder

	b.WriteString("Review your settings:\n\n")
	b.WriteString(fmt.Sprintf("Model: %s\n", f.getModelDisplayName(f.result.Model)))
	b.WriteString(fmt.Sprintf("API Key: %s\n", f.maskAPIKey(f.result.APIKey)))
	b.WriteString(fmt.Sprintf("Temperature: %.1f\n", f.result.Temperature))
	b.WriteString(fmt.Sprintf("Max Tokens: %d\n", f.result.MaxTokens))

	if f.configPath != "" {
		b.WriteString(fmt.Sprintf("\nSave to: %s", f.configPath))
	}

	return b.String()
}

// getModelDisplayName returns a short display name for a model
func (f *SettingsForm) getModelDisplayName(modelID string) string {
	switch {
	case strings.Contains(modelID, "sonnet") && strings.Contains(modelID, "3-5"):
		return "Claude 3.5 Sonnet"
	case strings.Contains(modelID, "haiku") && strings.Contains(modelID, "3-5"):
		return "Claude 3.5 Haiku"
	case strings.Contains(modelID, "opus"):
		return "Claude 3 Opus"
	case strings.Contains(modelID, "sonnet"):
		return "Claude 3 Sonnet"
	case strings.Contains(modelID, "haiku"):
		return "Claude 3 Haiku"
	default:
		return modelID
	}
}

// maskAPIKey masks an API key for display
func (f *SettingsForm) maskAPIKey(key string) string {
	if len(key) <= 12 {
		return "sk-ant-***"
	}
	return key[:12] + "..." + key[len(key)-4:]
}

// getDraculaTheme returns a huh theme configured with Dracula colors
func (f *SettingsForm) getDraculaTheme() *huh.Theme {
	t := huh.ThemeBase()

	colors := f.theme.Colors

	// Configure with Dracula colors
	t.Focused.Base = t.Focused.Base.
		BorderForeground(colors.Purple)

	t.Focused.Title = t.Focused.Title.
		Foreground(colors.Purple).
		Bold(true)

	t.Focused.Description = t.Focused.Description.
		Foreground(colors.Foreground)

	t.Focused.SelectSelector = t.Focused.SelectSelector.
		Foreground(colors.Pink)

	t.Focused.SelectedOption = t.Focused.SelectedOption.
		Foreground(colors.Cyan).
		Bold(true)

	t.Focused.UnselectedOption = t.Focused.UnselectedOption.
		Foreground(colors.Comment)

	t.Focused.TextInput.Cursor = t.Focused.TextInput.Cursor.
		Foreground(colors.Pink)

	t.Focused.TextInput.Placeholder = t.Focused.TextInput.Placeholder.
		Foreground(colors.Comment)

	t.Focused.TextInput.Prompt = t.Focused.TextInput.Prompt.
		Foreground(colors.Purple)

	t.Focused.FocusedButton = t.Focused.FocusedButton.
		Foreground(colors.Background).
		Background(colors.Green).
		Bold(true)

	t.Focused.BlurredButton = t.Focused.BlurredButton.
		Foreground(colors.Foreground).
		Background(colors.CurrentLine)

	return t
}

// SaveToFile saves the settings to a configuration file
func (f *SettingsForm) SaveToFile() error {
	if f.configPath == "" {
		return fmt.Errorf("no config path specified")
	}

	// Ensure directory exists
	dir := filepath.Dir(f.configPath)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	// Build config content
	var b strings.Builder
	b.WriteString("# Hex Configuration\n")
	b.WriteString("# Generated by settings wizard\n\n")
	b.WriteString(fmt.Sprintf("model = %q\n", f.result.Model))
	b.WriteString(fmt.Sprintf("api_key = %q\n", f.result.APIKey))
	b.WriteString(fmt.Sprintf("temperature = %.1f\n", f.result.Temperature))
	b.WriteString(fmt.Sprintf("max_tokens = %d\n", f.result.MaxTokens))

	// Write to file
	err := os.WriteFile(f.configPath, []byte(b.String()), 0600)
	if err != nil {
		return fmt.Errorf("write config file: %w", err)
	}

	return nil
}
