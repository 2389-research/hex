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

	// Create theme (reused for history content)
	th := theme.NewDraculaTheme()

	// Create app pointer - will be set after app creation.
	// This allows the callback to access app methods.
	var app *tux.App

	// Create history content with callback for session selection.
	// Note: The callback handles session resumption. User should switch
	// to Chat tab (via Tab key or Ctrl+1) after selecting a session.
	historyContent := tui.NewHistoryContent(storage, th, func(session *tui.Session) {
		if session == nil || app == nil {
			return
		}

		// Clear chat display and restore messages from session
		app.ClearChat()
		for _, msg := range session.Messages {
			if msg.Role == "user" {
				app.AddChatUserMessage(msg.Content)
			} else if msg.Role == "assistant" {
				app.AddChatAssistantMessage(msg.Content)
			}
		}

		// Resume the selected session in agent
		if err := agent.ResumeSession(session.ID); err != nil {
			// Error handling - session resume failed
			// The user will see an error when they try to chat
			return
		}
		// Session resumed successfully
		// User should switch to Chat tab to continue the conversation
	})

	// Create tux app with Dracula theme and History tab
	app = tux.New(agent,
		tux.WithTheme(th),
		tux.WithTab(tux.TabDef{
			ID:       "history",
			Label:    "History",
			Shortcut: "ctrl+h",
			Content:  historyContent,
		}),
	)

	// Run the app
	if err := app.Run(); err != nil {
		return fmt.Errorf("tux app error: %w", err)
	}

	return nil
}
