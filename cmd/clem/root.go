package main

import (
	"database/sql"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	ctxmgr "github.com/harper/clem/internal/context"
	"github.com/harper/clem/internal/core"
	"github.com/harper/clem/internal/logging"
	"github.com/harper/clem/internal/mcp"
	"github.com/harper/clem/internal/permissions"
	"github.com/harper/clem/internal/storage"
	"github.com/harper/clem/internal/templates"
	"github.com/harper/clem/internal/tools"
	"github.com/harper/clem/internal/ui"
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

	// Print mode tool flags (like Claude Code)
	dangerouslySkipPermissions bool
	enabledTools               []string
	systemPrompt               string

	// Phase 3: Permission system flags
	permissionMode  string
	allowedTools    []string
	disallowedTools []string
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
	rootCmd.PersistentFlags().StringVar(&templateName, "template", "", "Use a session template (see 'clem templates list')")

	// Print mode tool support flags (like Claude Code)
	rootCmd.PersistentFlags().BoolVar(&dangerouslySkipPermissions, "dangerously-skip-permissions", false, "Auto-approve all tool executions (use with caution)")
	rootCmd.PersistentFlags().StringSliceVar(&enabledTools, "tools", []string{}, "Tools to enable in print mode (comma-separated, e.g. 'write_file,read_file')")
	rootCmd.PersistentFlags().StringVar(&systemPrompt, "system-prompt", "", "System prompt to use for the session")

	// Phase 3: Permission system flags
	rootCmd.PersistentFlags().StringVar(&permissionMode, "permission-mode", "ask", "Permission mode: auto (approve all), ask (prompt for each), deny (block all)")
	rootCmd.PersistentFlags().StringSliceVar(&allowedTools, "allowed-tools", []string{}, "Whitelist of allowed tools (comma-separated). If set, only these tools are allowed.")
	rootCmd.PersistentFlags().StringSliceVar(&disallowedTools, "disallowed-tools", []string{}, "Blacklist of disallowed tools (comma-separated). These tools are blocked.")
}

func runRoot(_ *cobra.Command, args []string) error {
	// Initialize logging
	if err := initializeLogging(); err != nil {
		return fmt.Errorf("initialize logging: %w", err)
	}
	defer closeLogger()

	logging.InfoWith("Clem starting", "version", version)

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

	// Use default model if not specified, or from template
	modelName := model
	if modelName == "" {
		if template != nil && template.Model != "" {
			modelName = template.Model
		} else {
			modelName = "claude-sonnet-4-5-20250929"
		}
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
		uiModel = ui.NewModel(conversationID, modelName)
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
	// Load config to get API key
	cfg, err := core.LoadConfig()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Prioritize environment variable, fall back to config file
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		apiKey = cfg.APIKey
	}

	if apiKey != "" {
		client := core.NewClient(apiKey)
		uiModel.SetAPIClient(client)
	} else {
		return fmt.Errorf("API key not configured. Run 'clem setup-token <key>' or set ANTHROPIC_API_KEY environment variable")
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

	// Phase 6: Initialize plugin system
	pluginRegistry, err := initializePlugins()
	if err != nil {
		// Log warning but continue - plugins are optional
		logging.WarnWith("Failed to initialize plugins", "error", err.Error())
	}

	// Get plugin-provided skill and command paths
	var pluginSkillPaths, pluginCommandPaths []string
	if pluginRegistry != nil {
		pluginSkillPaths = getPluginSkillPaths(pluginRegistry)
		pluginCommandPaths = getPluginCommandPaths(pluginRegistry)
		logging.DebugWith("Plugin paths", "skills", len(pluginSkillPaths), "commands", len(pluginCommandPaths))
	}

	// Phase 2: Register Skills system (with plugin skills)
	skillRegistry, skillTool := initializeSkills(pluginSkillPaths)
	if err := registry.Register(skillTool); err != nil {
		return fmt.Errorf("register skill tool: %w", err)
	}
	logging.InfoWith("Loaded skills", "count", skillRegistry.Count())

	// Phase 4: Register Slash Commands system (with plugin commands)
	commandRegistry, slashCommandTool := initializeCommands(pluginCommandPaths)
	if err := registry.Register(slashCommandTool); err != nil {
		return fmt.Errorf("register slash command tool: %w", err)
	}
	logging.InfoWith("Loaded slash commands", "count", commandRegistry.Count())

	// Phase 5B: Load MCP tools from .mcp.json if present
	logging.Debug("Loading MCP tools")
	if err := mcp.LoadMCPTools(".", registry); err != nil {
		// Log error but don't fail - continue with built-in tools
		logging.WarnWith("Failed to load MCP tools", "error", err.Error())
	} else {
		logging.Info("MCP tools loaded successfully")
	}

	// Phase 3: Create permission checker from flags
	permChecker, err := createPermissionChecker()
	if err != nil {
		return fmt.Errorf("create permission checker: %w", err)
	}

	// Create executor with approval function
	// The actual approval is handled by the UI, so we return true here
	approvalFunc := func(_ string, _ map[string]interface{}) bool {
		return true
	}
	executor := tools.NewExecutor(registry, approvalFunc)

	// Phase 3: Attach permission checker to executor
	if permChecker != nil {
		executor.SetPermissionChecker(permChecker)
		logging.InfoWith("Permission system enabled", "mode", permChecker.GetMode())
	}

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

	// Suppress debug logs to stderr during TUI operation to avoid display corruption
	// Save original stderr and redirect to /dev/null
	origStderr := os.Stderr
	devNull, err := os.Open(os.DevNull)
	if err == nil {
		os.Stderr = devNull
		defer func() {
			os.Stderr = origStderr
			_ = devNull.Close() // Best effort close
		}()
	}

	// Start Bubbletea program
	p := tea.NewProgram(uiModel, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		os.Stderr = origStderr // Restore stderr for error reporting
		logging.ErrorWithErr("Failed to run UI", err)
		return fmt.Errorf("run UI: %w", err)
	}

	logging.Info("Clem shutting down")
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
		// Discard logs when no log file specified to avoid disrupting TUI
		// Users running the TUI don't see stderr logs anyway due to alt screen
		config.Writer = io.Discard
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

// createPermissionChecker creates a permission checker from CLI flags
func createPermissionChecker() (*permissions.Checker, error) {
	// Handle legacy --dangerously-skip-permissions flag for backward compatibility
	mode := permissionMode
	if dangerouslySkipPermissions {
		mode = "auto"
		logging.Warn("Using --dangerously-skip-permissions (deprecated). Consider using --permission-mode=auto instead")
	}

	// Parse permission mode
	parsedMode, err := permissions.ParseMode(mode)
	if err != nil {
		return nil, fmt.Errorf("invalid permission mode: %w", err)
	}

	// Create permission config
	config := &permissions.Config{
		Mode:            parsedMode,
		AllowedTools:    allowedTools,
		DisallowedTools: disallowedTools,
	}

	// Validate and convert to checker
	checker, err := config.ToChecker()
	if err != nil {
		return nil, fmt.Errorf("create permission checker: %w", err)
	}

	logging.DebugWith("Permission checker created", "config", config.String())
	return checker, nil
}
