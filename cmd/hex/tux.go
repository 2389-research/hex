// ABOUTME: Entry point for tux-based TUI
// ABOUTME: Uses tux.New() instead of ui.NewModel()

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/2389-research/hex/internal/core"
	"github.com/2389-research/hex/internal/tools"
	"github.com/2389-research/hex/internal/tui"
	"github.com/2389-research/tux"
	"github.com/2389-research/tux/theme"
)

// runTuxMode starts the tux-based TUI.
// This is called when --tux flag is passed.
func runTuxMode(apiKey, model, systemPrompt string, executor *tools.Executor) error {
	// Create API client
	client := core.NewClient(apiKey)

	// Create session storage
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("get home directory: %w", err)
	}
	sessionDir := filepath.Join(homeDir, ".hex", "sessions")
	storage, err := tui.NewSessionStorage(sessionDir)
	if err != nil {
		return fmt.Errorf("create session storage: %w", err)
	}

	// Create HexAgent
	agent := tui.NewHexAgent(client, model, systemPrompt, executor, storage)

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
