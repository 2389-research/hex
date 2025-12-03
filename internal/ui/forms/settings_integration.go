// Package forms provides beautiful huh-based forms for the hex TUI.
// ABOUTME: Integration layer for settings wizard with Model
// ABOUTME: Provides async command for running settings wizard
package forms

import (
	tea "github.com/charmbracelet/bubbletea"
)

// SettingsResultMsg is sent when the settings wizard completes
type SettingsResultMsg struct {
	Result *SettingsFormResult
	Error  error
}

// RunSettingsFormAsync runs the settings wizard in a goroutine
// and returns a tea.Cmd that will send the result when complete
func RunSettingsFormAsync(currentModel, currentAPIKey, configPath string) tea.Cmd {
	return func() tea.Msg {
		form := NewSettingsForm(currentModel, currentAPIKey, configPath)
		result, err := form.Run()

		if err != nil {
			return &SettingsResultMsg{
				Error: err,
			}
		}

		// If not cancelled, save to file
		if !result.Cancelled {
			if saveErr := form.SaveToFile(); saveErr != nil {
				return &SettingsResultMsg{
					Result: result,
					Error:  saveErr,
				}
			}
		}

		return &SettingsResultMsg{
			Result: result,
			Error:  nil,
		}
	}
}
