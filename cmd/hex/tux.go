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

	// Create autocomplete with command provider
	autocomplete := tux.NewAutocomplete()
	commands := []tux.Completion{
		{Value: "/help", Display: "/help", Description: "Show help", Score: 100},
		{Value: "/new", Display: "/new", Description: "Start new session", Score: 90},
		{Value: "/history", Display: "/history", Description: "Browse session history", Score: 80},
		{Value: "/clear", Display: "/clear", Description: "Clear chat display", Score: 70},
		{Value: "/model", Display: "/model", Description: "Show current model", Score: 60},
	}
	autocomplete.RegisterProvider("command", tux.NewCommandProvider(commands))

	// Define help categories for the ? overlay
	helpCategories := []tux.HelpCategory{
		{
			Title: "General",
			Bindings: []tux.HelpBinding{
				{Key: "?", Description: "Toggle help overlay"},
				{Key: "Tab", Description: "Autocomplete commands"},
				{Key: "Ctrl+C", Description: "Quit"},
				{Key: "Ctrl+E", Description: "Show errors"},
				{Key: "Esc", Description: "Toggle input/content focus"},
			},
		},
		{
			Title: "Tabs",
			Bindings: []tux.HelpBinding{
				{Key: "Alt+1", Description: "Switch to Chat tab"},
				{Key: "Alt+2", Description: "Switch to Tools tab"},
				{Key: "Alt+3", Description: "Switch to History tab"},
				{Key: "Ctrl+H", Description: "Switch to History tab"},
				{Key: "Ctrl+O", Description: "Switch to Tools tab"},
			},
		},
		{
			Title: "Chat (when focused)",
			Bindings: []tux.HelpBinding{
				{Key: "j/k", Description: "Scroll up/down"},
				{Key: "g", Description: "Jump to top"},
				{Key: "G", Description: "Jump to bottom"},
				{Key: "Ctrl+U/D", Description: "Page up/down"},
			},
		},
		{
			Title: "History",
			Bindings: []tux.HelpBinding{
				{Key: "j/k", Description: "Navigate sessions"},
				{Key: "Enter", Description: "Resume session"},
				{Key: "n", Description: "New session"},
				{Key: "f", Description: "Toggle favorite"},
				{Key: "d", Description: "Delete (press twice)"},
				{Key: "/", Description: "Search sessions"},
				{Key: "r", Description: "Refresh list"},
			},
		},
	}

	// Create tux app with Dracula theme, History tab, help, and autocomplete
	app = tux.New(agent,
		tux.WithTheme(th),
		tux.WithTab(tux.TabDef{
			ID:       "history",
			Label:    "History",
			Shortcut: "ctrl+h",
			Content:  historyContent,
		}),
		tux.WithHelpCategories(helpCategories...),
		tux.WithAutocomplete(autocomplete),
	)

	// Run the app
	if err := app.Run(); err != nil {
		return fmt.Errorf("tux app error: %w", err)
	}

	return nil
}
