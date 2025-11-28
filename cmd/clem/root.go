package main

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/harper/clem/internal/core"
	"github.com/harper/clem/internal/mcp"
	"github.com/harper/clem/internal/storage"
	"github.com/harper/clem/internal/tools"
	"github.com/harper/clem/internal/ui"
	"github.com/spf13/cobra"
)

var (
	// Version information
	version = "0.1.0"

	// Global flags
	printMode    bool
	outputFormat string
	model        string
	verbose      bool
	debug        string

	// Task 7: Storage integration flags
	continueFlag bool
	resumeID     string
	dbPath       string
)

var rootCmd = &cobra.Command{
	Use:   "clem [prompt]",
	Short: "Clem - AI assistant CLI",
	Long: `Clem is an AI assistant for your terminal.

Start an interactive session or use --print for one-off queries.`,
	Version: version,
	Args:    cobra.ArbitraryArgs,
	RunE:    runRoot,
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolVarP(&printMode, "print", "p", false, "Print mode (non-interactive)")
	rootCmd.PersistentFlags().StringVar(&outputFormat, "output-format", "text", "Output format: text, json, stream-json")
	rootCmd.PersistentFlags().StringVarP(&model, "model", "m", "", "Model to use")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "Verbose output")
	rootCmd.PersistentFlags().StringVar(&debug, "debug", "", "Debug categories")

	// Task 7: Storage flags
	rootCmd.PersistentFlags().BoolVar(&continueFlag, "continue", false, "Continue the most recent conversation")
	rootCmd.PersistentFlags().StringVar(&resumeID, "resume", "", "Resume a specific conversation by ID")
	rootCmd.PersistentFlags().StringVar(&dbPath, "db-path", defaultDBPath(), "Path to database file")
}

func runRoot(cmd *cobra.Command, args []string) error {
	prompt := ""
	if len(args) > 0 {
		prompt = joinArgs(args)
	}

	if printMode {
		return runPrintMode(prompt)
	}

	return runInteractive(prompt)
}

func joinArgs(args []string) string {
	return strings.Join(args, " ")
}

func runInteractive(prompt string) error {
	// Validate flag conflicts
	if continueFlag && resumeID != "" {
		return fmt.Errorf("cannot use both --continue and --resume flags together")
	}

	// Task 7: Open database
	db, err := openDatabase(dbPath)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	// Use default model if not specified
	modelName := model
	if modelName == "" {
		modelName = "claude-sonnet-4-5-20250929"
	}

	var conversationID string
	var uiModel *ui.Model

	// Task 7: Handle --continue or --resume flags
	if continueFlag {
		// Load latest conversation
		conv, err := storage.GetLatestConversation(db)
		if err == sql.ErrNoRows {
			// No conversations found, start new one (this is OK)
			conversationID = ""
			fmt.Fprintf(os.Stderr, "No previous conversations found, starting new session\n")
		} else if err != nil {
			// Real database error (corrupt DB, connection issue, etc.)
			return fmt.Errorf("failed to load latest conversation: %w", err)
		} else {
			// Success - load the conversation
			conversationID = conv.ID
			modelName = conv.Model
			uiModel = ui.NewModel(conversationID, modelName)
			uiModel.SetDB(db)

			// Load messages into UI
			msgs, err := storage.ListMessages(db, conversationID)
			if err != nil {
				return fmt.Errorf("failed to load conversation messages: %w", err)
			}
			for _, msg := range msgs {
				uiModel.AddMessage(msg.Role, msg.Content)
			}
		}
	} else if resumeID != "" {
		// Load specific conversation
		conv, messages, err := loadConversationHistory(db, resumeID)
		if err != nil {
			return fmt.Errorf("load conversation: %w", err)
		}
		conversationID = conv.ID
		modelName = conv.Model
		uiModel = ui.NewModel(conversationID, modelName)
		uiModel.SetDB(db)

		// Load messages into UI
		for _, msg := range messages {
			uiModel.AddMessage(msg.Role, msg.Content)
		}
	}

	// Create new conversation if not resuming
	if uiModel == nil {
		conversationID = fmt.Sprintf("conv-%d", time.Now().Unix())
		uiModel = ui.NewModel(conversationID, modelName)
		uiModel.SetDB(db)

		// Create conversation in database
		conv := &storage.Conversation{
			ID:    conversationID,
			Title: "New Conversation",
			Model: modelName,
		}
		if err := storage.CreateConversation(db, conv); err != nil {
			return fmt.Errorf("create conversation: %w", err)
		}
	}

	// Task 6: Create and set API client
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey != "" {
		client := core.NewClient(apiKey)
		uiModel.SetAPIClient(client)
	}

	// Task 12: Create tool registry and executor
	registry := tools.NewRegistry()
	if err := registry.Register(tools.NewReadTool()); err != nil {
		return fmt.Errorf("register read tool: %w", err)
	}
	if err := registry.Register(tools.NewWriteTool()); err != nil {
		return fmt.Errorf("register write tool: %w", err)
	}
	if err := registry.Register(tools.NewBashTool()); err != nil {
		return fmt.Errorf("register bash tool: %w", err)
	}
	// Phase 3: Register Edit, Grep, Glob tools
	if err := registry.Register(tools.NewEditTool()); err != nil {
		return fmt.Errorf("register edit tool: %w", err)
	}
	if err := registry.Register(tools.NewGrepTool()); err != nil {
		return fmt.Errorf("register grep tool: %w", err)
	}
	if err := registry.Register(tools.NewGlobTool()); err != nil {
		return fmt.Errorf("register glob tool: %w", err)
	}

	// Phase 4A: Register Interactive tools (AskUserQuestion, TodoWrite)
	if err := registry.Register(tools.NewAskUserQuestionTool()); err != nil {
		return fmt.Errorf("register ask_user_question tool: %w", err)
	}
	if err := registry.Register(tools.NewTodoWriteTool()); err != nil {
		return fmt.Errorf("register todo_write tool: %w", err)
	}

	// Phase 4B: Register Research tools (WebFetch, WebSearch)
	if err := registry.Register(tools.NewWebFetchTool()); err != nil {
		return fmt.Errorf("register web_fetch tool: %w", err)
	}
	if err := registry.Register(tools.NewWebSearchTool()); err != nil {
		return fmt.Errorf("register web_search tool: %w", err)
	}

	// Phase 4C: Register Advanced Execution tools (Task, BashOutput, KillShell)
	if err := registry.Register(tools.NewTaskTool()); err != nil {
		return fmt.Errorf("register task tool: %w", err)
	}
	if err := registry.Register(tools.NewBashOutputTool()); err != nil {
		return fmt.Errorf("register bash_output tool: %w", err)
	}
	if err := registry.Register(tools.NewKillShellTool()); err != nil {
		return fmt.Errorf("register kill_shell tool: %w", err)
	}

	// Phase 5B: Load MCP tools from .mcp.json if present
	if err := mcp.LoadMCPTools(".", registry); err != nil {
		// Log error but don't fail - continue with built-in tools
		fmt.Fprintf(os.Stderr, "Warning: Failed to load MCP tools: %v\n", err)
	}

	// Create executor with approval function
	// The actual approval is handled by the UI, so we return true here
	approvalFunc := func(toolName string, params map[string]interface{}) bool {
		return true
	}
	executor := tools.NewExecutor(registry, approvalFunc)

	// Set tool system in model
	uiModel.SetToolSystem(registry, executor)

	// Add initial prompt if provided
	if prompt != "" {
		uiModel.AddMessage("user", prompt)
		// TODO: Send to API and stream response
	}

	// Start Bubbletea program
	p := tea.NewProgram(uiModel, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("run UI: %w", err)
	}

	return nil
}
