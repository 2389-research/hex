package main

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	ctxmgr "github.com/harper/pagent/internal/context"
	"github.com/harper/pagent/internal/core"
	"github.com/harper/pagent/internal/hooks"
	"github.com/harper/pagent/internal/logging"
	"github.com/harper/pagent/internal/mcp"
	"github.com/harper/pagent/internal/providers"
	"github.com/harper/pagent/internal/storage"
	"github.com/harper/pagent/internal/templates"
	"github.com/harper/pagent/internal/tools"
	"github.com/harper/pagent/internal/ui"
	"github.com/spf13/cobra"
)

var (
	// Version information
	version = "1.0.0"
	commit  = "dev"     //nolint:unused // Populated by goreleaser ldflags
	date    = "unknown" //nolint:unused // Populated by goreleaser ldflags

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

	// Phase 6A: Logging flags
	logLevel  string
	logFile   string
	logFormat string

	// Phase 6B: Multimodal flags
	imagePaths []string

	// Phase 6B: Context management flags
	maxContextTokens int
	contextStrategy  string

	// Phase 6C: Template system flags
	templateName string

	// TUI theme flag
	theme string

	// Print mode tool flags (like Claude Code)
	dangerouslySkipPermissions bool
	enabledTools               []string
	systemPrompt               string
)

var rootCmd = &cobra.Command{
	Use:   "pagent [prompt]",
	Short: "Pagen - Productivity AI Agent",
	Long: `Pagen is a productivity AI agent for your terminal.

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

	// Phase 6A: Logging flags
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "Log level: debug, info, warn, error")
	rootCmd.PersistentFlags().StringVar(&logFile, "log-file", "", "Log to file (optional)")
	rootCmd.PersistentFlags().StringVar(&logFormat, "log-format", "text", "Log format: text, json")

	// Phase 6B: Multimodal flags
	rootCmd.PersistentFlags().StringSliceVar(&imagePaths, "image", []string{}, "Image file path(s) to include in message (repeatable)")

	// Phase 6B: Context management flags
	rootCmd.PersistentFlags().IntVar(&maxContextTokens, "max-context-tokens", 180000, "Maximum context window size in tokens")
	rootCmd.PersistentFlags().StringVar(&contextStrategy, "context-strategy", "prune", "Context management strategy: keep-all, prune, summarize")

	// Phase 6C: Template system flags
	rootCmd.PersistentFlags().StringVar(&templateName, "template", "", "Use a session template (see 'pagent templates list')")

	// TUI theme flag
	rootCmd.PersistentFlags().StringVar(&theme, "theme", "", "TUI theme: dracula, gruvbox, nord (default: dracula)")

	// Print mode tool support flags (like Claude Code)
	rootCmd.PersistentFlags().BoolVar(&dangerouslySkipPermissions, "dangerously-skip-permissions", false, "Auto-approve all tool executions (use with caution)")
	rootCmd.PersistentFlags().StringSliceVar(&enabledTools, "tools", []string{}, "Tools to enable in print mode (comma-separated, e.g. 'write_file,read_file')")
	rootCmd.PersistentFlags().StringVar(&systemPrompt, "system-prompt", "", "System prompt to use for the session")
}

func runRoot(_ *cobra.Command, args []string) error {
	// Initialize logging
	if err := initializeLogging(); err != nil {
		return fmt.Errorf("initialize logging: %w", err)
	}
	defer closeLogger()

	logging.InfoWith("Pagen starting", "version", version)

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
	logging.Debug("Starting interactive mode")

	// Validate flag conflicts
	if continueFlag && resumeID != "" {
		logging.Error("Flag conflict: both --continue and --resume specified")
		return fmt.Errorf("cannot use both --continue and --resume flags together. Use either --continue to resume the latest conversation or --resume <ID> for a specific one")
	}

	// Task 7: Open database
	logging.DebugWith("Opening database", "path", dbPath)
	db, err := openDatabase(dbPath)
	if err != nil {
		logging.ErrorWithErr("Failed to open database", err, "path", dbPath)
		return fmt.Errorf("failed to open database at %s: %w. Try:\n  - Check if parent directory exists\n  - Check write permissions\n  - Try different path with --db-path", dbPath, err)
	}
	defer func() { _ = db.Close() }()
	logging.InfoWith("Database opened successfully", "path", dbPath)

	// Phase 6C: Load template if specified
	var template *templates.Template
	var systemPrompt string
	if templateName != "" {
		var err error
		template, err = loadTemplateByName(templateName)
		if err != nil {
			return fmt.Errorf("load template: %w", err)
		}
		systemPrompt = template.SystemPrompt
		logging.InfoWith("Loaded template", "name", template.Name)
	}

	// Load config early to get theme preference
	cfg, err := core.LoadConfig()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Use default model if not specified, or from template
	modelName := model
	if modelName == "" {
		if template != nil && template.Model != "" {
			modelName = template.Model
		} else {
			modelName = "claude-sonnet-4-5-20250929"
		}
	}

	// Determine theme: prioritize --theme flag, then config file, then default
	themeName := theme
	if themeName == "" {
		themeName = cfg.Theme
	}
	if themeName == "" {
		themeName = "dracula"
	}

	// Validate theme and warn if unknown
	validThemes := map[string]bool{"dracula": true, "gruvbox": true, "nord": true}
	if themeName != "" && !validThemes[themeName] {
		logging.Warn("Unknown theme '%s', falling back to dracula. Available themes: dracula, gruvbox, nord", themeName)
		themeName = "dracula"
	}

	var conversationID string
	var uiModel *ui.Model

	// Task 7: Handle --continue or --resume flags
	if continueFlag {
		// Load latest conversation
		conv, err := storage.GetLatestConversation(db)
		if err == sql.ErrNoRows {
			// No conversations found, start new one (this is OK)
			_, _ = fmt.Fprintf(os.Stderr, "No previous conversations found, starting new session\n")
		} else if err != nil {
			// Real database error (corrupt DB, connection issue, etc.)
			return fmt.Errorf("failed to load latest conversation: %w", err)
		} else {
			// Success - load the conversation
			conversationID = conv.ID
			modelName = conv.Model
			uiModel = ui.NewModel(conversationID, modelName, themeName)
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
		uiModel = ui.NewModel(conversationID, modelName, themeName)
		uiModel.SetDB(db)

		// Load favorite status
		uiModel.IsFavorite = conv.IsFavorite

		// Load messages into UI
		for _, msg := range messages {
			uiModel.AddMessage(msg.Role, msg.Content)
		}
	}

	// Create new conversation if not resuming
	if uiModel == nil {
		conversationID = fmt.Sprintf("conv-%d", time.Now().Unix())
		uiModel = ui.NewModel(conversationID, modelName, themeName)
		uiModel.SetDB(db)

		// Phase 6C: Set system prompt from template if available
		if systemPrompt != "" {
			uiModel.SetSystemPrompt(systemPrompt)
		}

		// Create conversation in database
		title := "New Conversation"
		if template != nil && template.Name != "" {
			title = template.Name + " Session"
		}
		conv := &storage.Conversation{
			ID:    conversationID,
			Title: title,
			Model: modelName,
		}
		if err := storage.CreateConversation(db, conv); err != nil {
			return fmt.Errorf("create conversation: %w", err)
		}

		// Phase 6C: Add initial messages from template
		if template != nil && len(template.InitialMessages) > 0 {
			for _, msg := range template.InitialMessages {
				uiModel.AddMessage(msg.Role, msg.Content)
			}
		}
	}

	// Task 6: Create and set API client
	// Config already loaded above for theme

	// Initialize hook manager
	hookManager := hooks.NewManager(cfg.Hooks)

	// Trigger SessionStart
	hookManager.TriggerAsync(hooks.SessionStart, hooks.EventData{
		"timestamp": time.Now().Format(time.RFC3339),
		"model":     modelName,
	})

	// Ensure SessionEnd runs on exit
	defer func() {
		hookManager.TriggerAsync(hooks.SessionEnd, hooks.EventData{
			"timestamp": time.Now().Format(time.RFC3339),
		})
		hookManager.Wait() // Wait for all async hooks to complete
	}()

	// Prioritize environment variable, fall back to config file
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		apiKey = cfg.APIKey
	}

	if apiKey != "" {
		client := core.NewClient(apiKey)
		uiModel.SetAPIClient(client)
	} else {
		return fmt.Errorf("API key not configured. Run 'pagent setup-token <key>' or set ANTHROPIC_API_KEY environment variable")
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

	// Pagen: Register productivity tools (email, calendar, tasks)
	// TODO: Initialize provider registry and load configured providers
	providerRegistry := providers.NewRegistry()
	// For now, tools will fail with "no active provider" until providers are configured

	// Email tools
	if err := registry.Register(tools.NewSendEmailTool(providerRegistry)); err != nil {
		return fmt.Errorf("register send_email tool: %w", err)
	}
	if err := registry.Register(tools.NewSearchEmailsTool(providerRegistry)); err != nil {
		return fmt.Errorf("register search_emails tool: %w", err)
	}
	if err := registry.Register(tools.NewReadEmailTool(providerRegistry)); err != nil {
		return fmt.Errorf("register read_email tool: %w", err)
	}

	// Calendar tools
	if err := registry.Register(tools.NewCreateEventTool(providerRegistry)); err != nil {
		return fmt.Errorf("register create_event tool: %w", err)
	}
	if err := registry.Register(tools.NewListEventsTool(providerRegistry)); err != nil {
		return fmt.Errorf("register list_events tool: %w", err)
	}

	// Task tools
	if err := registry.Register(tools.NewCreateTaskTool(providerRegistry)); err != nil {
		return fmt.Errorf("register create_task tool: %w", err)
	}
	if err := registry.Register(tools.NewListTasksTool(providerRegistry)); err != nil {
		return fmt.Errorf("register list_tasks tool: %w", err)
	}
	if err := registry.Register(tools.NewCompleteTaskTool(providerRegistry)); err != nil {
		return fmt.Errorf("register complete_task tool: %w", err)
	}

	// Phase 5B: Load MCP tools from .mcp.json if present
	logging.Debug("Loading MCP tools")
	if err := mcp.LoadMCPTools(".", registry); err != nil {
		// Log error but don't fail - continue with built-in tools
		logging.WarnWith("Failed to load MCP tools", "error", err.Error())
	} else {
		logging.Info("MCP tools loaded successfully")
	}

	// Create executor with approval function
	// The actual approval is handled by the UI, so we return true here
	approvalFunc := func(_ string, _ map[string]interface{}) bool {
		return true
	}
	executor := tools.NewExecutor(registry, approvalFunc)

	// Set tool system in model
	uiModel.SetToolSystem(registry, executor)

	// Phase 6B: Set up context manager
	contextManager := ctxmgr.NewManager(maxContextTokens)
	uiModel.SetContextManager(contextManager)
	logging.DebugWith("Context manager initialized", "maxTokens", maxContextTokens, "strategy", contextStrategy)

	// Add initial prompt if provided
	if prompt != "" {
		uiModel.AddMessage("user", prompt)
		// TODO: Send to API and stream response
	}

	// Start Bubbletea program
	p := tea.NewProgram(uiModel, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		logging.ErrorWithErr("Failed to run UI", err)
		return fmt.Errorf("run UI: %w", err)
	}

	logging.Info("Pagen shutting down")
	return nil
}

var globalLogger *logging.Logger

func initializeLogging() error {
	level := logging.LevelFromString(logLevel)

	var format logging.Format
	switch logFormat {
	case "json":
		format = logging.FormatJSON
	default:
		format = logging.FormatText
	}

	config := logging.Config{
		Level:  level,
		Format: format,
	}

	var logger *logging.Logger
	var err error

	if logFile != "" {
		// Log to file (and optionally stderr in debug mode)
		config.LogFile = logFile
		if level == logging.LevelDebug {
			config.Writer = os.Stderr
		}
		logger, err = logging.NewLoggerWithFile(config)
		if err != nil {
			return fmt.Errorf("failed to create logger with file %s: %w. Try:\n  - Check parent directory exists\n  - Check write permissions\n  - Use absolute path", logFile, err)
		}
	} else {
		// Log to stderr only
		config.Writer = os.Stderr
		logger = logging.NewLogger(config)
	}

	globalLogger = logger
	logging.SetGlobalLogger(logger)
	return nil
}

func closeLogger() {
	if globalLogger != nil {
		if err := globalLogger.Close(); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Warning: Failed to close logger: %v\n", err)
		}
	}
}
