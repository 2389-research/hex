// ABOUTME: Entry point for tux-based TUI
// ABOUTME: Uses tux.New() instead of ui.NewModel()

package main

import (
	"fmt"

	"github.com/2389-research/hex/internal/core"
	"github.com/2389-research/hex/internal/tui"
	"github.com/2389-research/tux"
	"github.com/2389-research/tux/theme"
)

// runTuxMode starts the tux-based TUI.
// This is called when --tux flag is passed.
func runTuxMode(apiKey, model, systemPrompt string) error {
	// Create API client
	client := core.NewClient(apiKey)

	// Create HexAgent
	agent := tui.NewHexAgent(client, model, systemPrompt)

	// Create tux app with Dracula theme
	app := tux.New(agent,
		tux.WithTheme(theme.NewDraculaTheme()),
	)

	// Run the app
	if err := app.Run(); err != nil {
		return fmt.Errorf("tux app error: %w", err)
	}

	return nil
}
