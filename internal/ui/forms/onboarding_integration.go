// Package forms provides beautiful huh-based forms for the hex TUI.
// ABOUTME: Integration layer for onboarding flow with Model
// ABOUTME: Provides async command for running first-time onboarding
package forms

import (
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
)

// OnboardingResultMsg is sent when the onboarding flow completes
type OnboardingResultMsg struct {
	Result *OnboardingFormResult
	Error  error
}

// RunOnboardingFormAsync runs the onboarding wizard in a goroutine
// and returns a tea.Cmd that will send the result when complete
func RunOnboardingFormAsync(configPath string) tea.Cmd {
	return func() tea.Msg {
		form := NewOnboardingForm(configPath)
		result, err := form.Run()

		if err != nil {
			return &OnboardingResultMsg{
				Error: err,
			}
		}

		// If completed (not skipped), save configuration
		if result.Completed && !result.Skipped {
			if saveErr := saveOnboardingConfig(configPath, result); saveErr != nil {
				return &OnboardingResultMsg{
					Result: result,
					Error:  saveErr,
				}
			}
		}

		return &OnboardingResultMsg{
			Result: result,
			Error:  nil,
		}
	}
}

// saveOnboardingConfig saves the onboarding results to the config file
func saveOnboardingConfig(configPath string, result *OnboardingFormResult) error {
	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return err
	}

	// Build config content
	content := "# Hex Configuration\n"
	content += "# Created by onboarding wizard\n\n"
	content += "model = \"" + result.Model + "\"\n"
	content += "api_key = \"" + result.APIKey + "\"\n"
	content += "temperature = 1.0\n"
	content += "max_tokens = 4096\n"

	// Write to file with restricted permissions
	return os.WriteFile(configPath, []byte(content), 0600)
}

// IsFirstRun checks if this is the first time running Hex
// by checking if the config file exists
func IsFirstRun(configPath string) bool {
	_, err := os.Stat(configPath)
	return os.IsNotExist(err)
}
