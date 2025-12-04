// ABOUTME: Resume command for continuing previous conversations
// ABOUTME: Provides interactive picker and --last flag for session management
package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/2389-research/hex/internal/logging"
	"github.com/2389-research/hex/internal/services"
	"github.com/2389-research/hex/internal/storage"
	"github.com/2389-research/hex/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var (
	resumeLast bool
)

var resumeCmd = &cobra.Command{
	Use:   "resume [conversation-id]",
	Short: "Resume a previous conversation",
	Long: `Resume an interactive session from history.

Examples:
  hex resume               # Show interactive picker
  hex resume --last        # Resume most recent conversation
  hex resume conv-123      # Resume specific conversation by ID

The picker shows recent conversations with:
  - Conversation title
  - Last updated time
  - Model used
  - Favorite status`,
	Args: cobra.MaximumNArgs(1),
	RunE: runResume,
}

func init() {
	resumeCmd.Flags().BoolVar(&resumeLast, "last", false, "Resume the most recent conversation")
	rootCmd.AddCommand(resumeCmd)
}

func runResume(_ *cobra.Command, args []string) error {
	// Initialize logging
	if err := initializeLogging(); err != nil {
		return fmt.Errorf("initialize logging: %w", err)
	}
	defer closeLogger()

	// Open database
	logging.DebugWith("Opening database for resume", "path", dbPath)
	db, err := openDatabase(dbPath)
	if err != nil {
		logging.ErrorWithErr("Failed to open database", err)
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer func() { _ = db.Close() }()

	var conversationID string

	// Handle different resume modes
	switch {
	case len(args) > 0:
		// Specific conversation ID provided
		conversationID = args[0]
		logging.InfoWith("Resuming specific conversation", "id", conversationID)

	case resumeLast:
		// Resume most recent conversation
		conv, err := storage.GetLatestConversation(db)
		if err == sql.ErrNoRows {
			return fmt.Errorf("no previous conversations found")
		} else if err != nil {
			return fmt.Errorf("failed to get latest conversation: %w", err)
		}
		conversationID = conv.ID
		logging.InfoWith("Resuming latest conversation", "id", conversationID, "title", conv.Title)

	default:
		// Show interactive picker
		logging.Debug("Showing conversation picker")
		selectedID, err := showConversationPicker(db)
		if err != nil {
			return fmt.Errorf("picker error: %w", err)
		}
		if selectedID == "" {
			// User cancelled
			return nil
		}
		conversationID = selectedID
		logging.InfoWith("Selected conversation from picker", "id", conversationID)
	}

	// Validate conversation exists
	conv, err := storage.GetConversation(db, conversationID)
	if err == sql.ErrNoRows {
		return fmt.Errorf("conversation %s not found", conversationID)
	} else if err != nil {
		return fmt.Errorf("failed to load conversation: %w", err)
	}

	// Load messages
	messages, err := storage.ListMessages(db, conversationID)
	if err != nil {
		return fmt.Errorf("failed to load messages: %w", err)
	}

	// Create UI model with loaded conversation
	modelName := conv.Model
	uiModel := ui.NewModel(conversationID, modelName)
	uiModel.SetDB(db)
	uiModel.IsFavorite = conv.IsFavorite

	// Load messages into UI
	for _, msg := range messages {
		uiModel.AddMessage(msg.Role, msg.Content)
	}

	// Show conversation info
	_, _ = fmt.Fprintf(os.Stderr, "Resuming: %s\n", conv.Title)
	_, _ = fmt.Fprintf(os.Stderr, "Model: %s\n", modelName)
	_, _ = fmt.Fprintf(os.Stderr, "Messages: %d\n", len(messages))
	_, _ = fmt.Fprintf(os.Stderr, "\n")

	// Continue with normal interactive setup
	return continueInteractiveWithModel(db, uiModel, "")
}

// showConversationPicker displays an interactive TUI for selecting a conversation
func showConversationPicker(db *sql.DB) (string, error) {
	// Create conversation service
	convSvc := services.NewConversationService(db)

	// Load recent conversations
	conversations, err := convSvc.List(context.Background())
	if err != nil {
		return "", fmt.Errorf("failed to list conversations: %w", err)
	}

	if len(conversations) == 0 {
		return "", fmt.Errorf("no conversations found")
	}

	// Limit to 20 most recent
	if len(conversations) > 20 {
		conversations = conversations[:20]
	}

	// Create picker model
	picker := ui.NewSessionPicker(conversations)

	// Run picker
	p := tea.NewProgram(picker, tea.WithAltScreen())
	result, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("picker failed: %w", err)
	}

	// Get selected conversation ID
	finalModel := result.(ui.SessionPicker)
	return finalModel.GetSelectedID(), nil
}
