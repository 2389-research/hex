package main

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/2389-research/hex/internal/agentsmd"
	ctxmgr "github.com/2389-research/hex/internal/convcontext"
	"github.com/2389-research/hex/internal/coordinator"
	"github.com/2389-research/hex/internal/core"
	"github.com/2389-research/hex/internal/cost"
	"github.com/2389-research/hex/internal/events"
	"github.com/2389-research/hex/internal/logging"
	"github.com/2389-research/hex/internal/mcp"
	"github.com/2389-research/hex/internal/permissions"
	"github.com/2389-research/hex/internal/providers"
	"github.com/2389-research/hex/internal/services"
	"github.com/2389-research/hex/internal/shutdown"
	"github.com/2389-research/hex/internal/storage"
	"github.com/2389-research/hex/internal/templates"
	"github.com/2389-research/hex/internal/tools"
	"github.com/2389-research/hex/internal/ui"
	"github.com/2389-research/mux/llm"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// Context key type for storing values in context
type contextKey string

const (
	eventStoreKey contextKey = "event_store"
)

var (
	// Version information
	version = "1.0.0"
	_       = "dev"     // commit - populated by goreleaser ldflags, currently unused
	_       = "unknown" // date - populated by goreleaser ldflags, currently unused

	// Global flags
	printMode    bool
	outputFormat string
	provider     string
	model        string
	verbose      bool
	debug        bool

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

	// Legacy mode: use built-in orchestrator instead of mux (for backwards compatibility)
	useLegacyMode bool

	// Spell system flags
	spellName string
	spellMode string

	// Agent intelligence flags
	maxTurns      int
	planMode      bool
	refreshMemory bool
)

var rootCmd = &cobra.Command{
	Use:   "hex [prompt]",
	Short: "Hex - Powerful CLI for Claude AI",
	Long: `Hex is a powerful command-line interface for Claude AI, inspired by Claude Code, Crush, Codex, and MaKeR.

Start an interactive session or use --print for one-off queries.`,
	Version:       version,
	Args:          cobra.ArbitraryArgs,
	RunE:          runRoot,
	SilenceUsage:  true, // Don't print usage on errors - keeps error messages clean
	SilenceErrors: true, // Don't print errors - main.go handles it (prevents double printing)
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolVarP(&printMode, "print", "p", false, "Print mode (non-interactive)")
	rootCmd.PersistentFlags().StringVar(&outputFormat, "output-format", "text", "Output format: text, json, stream-json")
	rootCmd.PersistentFlags().StringVar(&provider, "provider", "", "LLM provider (anthropic) - other providers coming soon")
	rootCmd.PersistentFlags().StringVarP(&model, "model", "m", "", "Model to use")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "Verbose output")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug logging to /tmp/hex-debug.log")

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
	rootCmd.PersistentFlags().StringVar(&templateName, "template", "", "Use a session template (see 'hex templates list')")

	// Print mode tool support flags (like Claude Code)
	rootCmd.PersistentFlags().BoolVar(&dangerouslySkipPermissions, "dangerously-skip-permissions", false, "Auto-approve all tool executions (use with caution)")
	rootCmd.PersistentFlags().StringSliceVar(&enabledTools, "tools", []string{}, "Tools to enable in print mode (comma-separated, e.g. 'write_file,read_file')")
	rootCmd.PersistentFlags().StringVar(&systemPrompt, "system-prompt", "", "System prompt to use for the session")

	// Phase 3: Permission system flags
	rootCmd.PersistentFlags().StringVar(&permissionMode, "permission-mode", "ask", "Permission mode: auto (approve all), ask (prompt for each), deny (block all)")
	rootCmd.PersistentFlags().StringSliceVar(&allowedTools, "allowed-tools", []string{}, "Whitelist of allowed tools (comma-separated). If set, only these tools are allowed.")
	rootCmd.PersistentFlags().StringSliceVar(&disallowedTools, "disallowed-tools", []string{}, "Blacklist of disallowed tools (comma-separated). These tools are blocked.")

	// Legacy mode for backwards compatibility
	rootCmd.PersistentFlags().BoolVar(&useLegacyMode, "legacy", false, "Use legacy built-in orchestrator instead of mux (for backwards compatibility)")

	// Spell system flags
	rootCmd.PersistentFlags().StringVar(&spellName, "spell", "", "Use a spell (agent personality)")
	rootCmd.PersistentFlags().StringVar(&spellMode, "spell-mode", "", "Override spell mode: replace or layer")

	// Agent intelligence flags
	rootCmd.PersistentFlags().IntVar(&maxTurns, "max-turns", 20, "Maximum tool execution turns before stopping")
	rootCmd.PersistentFlags().BoolVar(&planMode, "plan", false, "Plan before executing: generate a plan first, then execute it step by step")
	rootCmd.PersistentFlags().BoolVar(&refreshMemory, "refresh-memory", false, "Force re-scan of project context (regenerate .hex/project.json)")

	// Register subcommands
	rootCmd.AddCommand(visualizeCmd)
	rootCmd.AddCommand(replayCmd)
}

func runRoot(_ *cobra.Command, args []string) error {
	// Create cancellable context for future tasks
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize shutdown handler for cascading process cleanup
	// This handles SIGINT/SIGTERM, releases file locks, stops child processes, and checks for orphans
	shutdown.InitShutdownHandler()

	// Initialize the global resource coordinator for multi-agent file locking
	// Uses lazy initialization via sync.Once, but we trigger it early here
	_ = coordinator.Global()

	// Initialize logging
	if err := initializeLogging(); err != nil {
		return fmt.Errorf("initialize logging: %w", err)
	}
	defer closeLogger()

	// Extract prompt once at the top level to avoid duplication
	prompt := ""
	if len(args) > 0 {
		prompt = joinArgs(args)
	}

	// Initialize event store
	eventFile := filepath.Join(os.TempDir(), fmt.Sprintf("hex_events_%s.jsonl", time.Now().Format("20060102_150405")))
	eventStore, err := events.NewEventStore(eventFile)
	if err != nil {
		logging.WarnWith("Failed to create event store", "error", err)
	} else {
		// Set global event store for backward compatibility
		events.SetGlobal(eventStore)

		// Store event store in context for future use
		ctx = context.WithValue(ctx, eventStoreKey, eventStore)

		defer func() { _ = eventStore.Close() }()

		// Set agent ID if not already set
		agentID := os.Getenv("HEX_AGENT_ID")
		if agentID == "" {
			agentID = "root"
			_ = os.Setenv("HEX_AGENT_ID", agentID)
		}

		// Record session start event
		_ = eventStore.Record(events.Event{
			ID:        uuid.New().String(),
			AgentID:   agentID,
			Type:      "SessionStart",
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"prompt": prompt,
				"tools":  enabledTools,
			},
		})

		// Inform user about event recording
		if os.Getenv("HEX_DEBUG") != "" {
			fmt.Fprintf(os.Stderr, "📊 Recording events to: %s\n", eventFile)
			fmt.Fprintf(os.Stderr, "   Replay: hex replay %s\n", eventFile)
			fmt.Fprintf(os.Stderr, "   Visualize: hex visualize %s\n", eventFile)
		}
	}

	// Initialize global cost tracker (singleton pattern used throughout codebase)
	// Cost tracking is recorded in client.go via cost.Global().RecordUsage()
	_ = cost.Global()

	defer func() {
		agentID := os.Getenv("HEX_AGENT_ID")
		if agentID != "" {
			// Use the global tracker which has been populated by client.go
			summary, err := cost.Global().GetAgentCost(agentID)
			if err != nil {
				logging.DebugIf("Failed to get cost summary", "error", err, "agent_id", agentID)
				return
			}
			if summary.TotalCost > 0 {
				fmt.Fprintf(os.Stderr, "\n💰 Session Cost: $%.4f\n", summary.TotalCost)
				fmt.Fprintf(os.Stderr, "   Input tokens: %d, Output tokens: %d\n",
					summary.InputTokens, summary.OutputTokens)
			}
		}
	}()

	logging.InfoWith("Hex starting", "version", version)

	// Suppress unused context warning - ctx will be used in future tasks
	_ = ctx

	// Check for legacy mode flag from environment (for subagent inheritance)
	if os.Getenv("HEX_LEGACY_MODE") == "1" {
		useLegacyMode = true
	}

	// Propagate legacy mode flag to environment for subagents
	if useLegacyMode {
		_ = os.Setenv("HEX_LEGACY_MODE", "1")
		logging.InfoWith("Using legacy orchestrator mode", "mode", "legacy")
	}

	// First-run check - launch setup wizard if no config exists
	// Skip in print mode since it may be used in scripts/automation
	if IsFirstRun() && !printMode {
		if err := RunWizard(); err != nil {
			if err == ErrSetupCancelled {
				return nil // User cancelled, exit gracefully
			}
			return fmt.Errorf("setup wizard: %w", err)
		}
		// Wizard completed successfully, continue to interactive mode
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

	// TUI mode uses the built-in orchestrator (mux integration for TUI is future work)
	logging.DebugWith("TUI mode uses built-in orchestrator", "legacy", useLegacyMode)

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

	// Task 9: Load config and validate provider early (fail fast)
	cfg, err := core.LoadConfig()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Task 9: Determine provider (flag > config > default)
	providerName := provider
	if providerName == "" {
		providerName = cfg.Provider
	}
	if providerName == "" {
		providerName = "anthropic" // Default to anthropic
	}

	// Task 11: All providers now supported!
	validProviders := map[string]bool{
		"anthropic":  true,
		"openai":     true,
		"gemini":     true,
		"openrouter": true,
		"ollama":     true,
	}
	if !validProviders[providerName] {
		return fmt.Errorf("unknown provider '%s'. Supported providers: anthropic, openai, gemini, openrouter, ollama", providerName)
	}
	logging.InfoWith("Using provider", "provider", providerName)

	// Phase 6C: Load template if specified
	var template *templates.Template
	var systemPrompt string
	if templateName != "" {
		var tmplErr error
		template, tmplErr = loadTemplateByName(templateName)
		if tmplErr != nil {
			return fmt.Errorf("load template: %w", tmplErr)
		}
		systemPrompt = template.SystemPrompt
		logging.InfoWith("Loaded template", "name", template.Name)
	}

	// Load AGENTS.md context from directory hierarchy (repo root → CWD)
	agentsContext, err := agentsmd.LoadContext()
	if err != nil {
		logging.WarnWith("Failed to load AGENTS.md context", "error", err.Error())
	} else if agentsContext != "" {
		// Combine AGENTS.md context with template system prompt
		if systemPrompt != "" {
			systemPrompt = agentsContext + "\n\n" + systemPrompt
		} else {
			systemPrompt = agentsContext
		}
		logging.InfoWith("Loaded AGENTS.md context", "length", len(agentsContext))
	}

	// Use default model if not specified, or from template
	modelName := model
	if modelName == "" {
		if template != nil && template.Model != "" {
			modelName = template.Model
		} else if providerName == "anthropic" {
			modelName = "claude-sonnet-4-5-20250929"
		} else {
			// Non-Anthropic providers require explicit --model flag
			return fmt.Errorf("--model flag is required when using --provider=%s\n\nExample models:\n  anthropic: claude-sonnet-4-5-20250929, claude-opus-4-5-20251101, claude-haiku-4-5-20251001\n  openai: gpt-5.1, gpt-5.1-codex, gpt-5.1-codex-mini\n  gemini: gemini-2.5-pro, gemini-2.5-flash, gemini-pro-latest\n  openrouter: anthropic/claude-sonnet-4-5, openai/gpt-5.1, google/gemini-2.5-pro\n  ollama: llama3.2, codellama, mistral, mixtral", providerName)
		}
	}

	var conversationID string
	var uiModel *ui.Model

	// Task 7: Handle --continue or --resume flags
	if continueFlag {
		// Load latest conversation
		conv, latestErr := storage.GetLatestConversation(db)
		if latestErr == sql.ErrNoRows {
			// No conversations found, start new one (this is OK)
			_, _ = fmt.Fprintf(os.Stderr, "No previous conversations found, starting new session\n")
		} else if latestErr != nil {
			// Real database error (corrupt DB, connection issue, etc.)
			return fmt.Errorf("failed to load latest conversation: %w", latestErr)
		} else {
			// Success - load the conversation
			conversationID = conv.ID
			modelName = conv.Model
			uiModel = ui.NewModel(conversationID, modelName)

			// Load messages into UI
			msgs, msgErr := storage.ListMessages(db, conversationID)
			if msgErr != nil {
				return fmt.Errorf("failed to load conversation messages: %w", msgErr)
			}
			for _, msg := range msgs {
				uiModel.AddMessage(msg.Role, msg.Content)
			}
		}
	} else if resumeID != "" {
		// Load specific conversation
		conv, messages, loadErr := loadConversationHistory(db, resumeID)
		if loadErr != nil {
			return fmt.Errorf("load conversation: %w", loadErr)
		}
		conversationID = conv.ID
		modelName = conv.Model
		uiModel = ui.NewModel(conversationID, modelName)

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

		// Phase 6C: Set system prompt from template if available
		if systemPrompt != "" {
			uiModel.SetSystemPrompt(systemPrompt)
		}

		// Create conversation in database
		title := "New Conversation"
		if template != nil && template.Name != "" {
			title = template.Name + " Session"
		}

		// Task 9: Use validated provider name from config/flag
		conv := &storage.Conversation{
			ID:       conversationID,
			Title:    title,
			Provider: providerName,
			Model:    modelName,
		}
		if createErr := storage.CreateConversation(db, conv); createErr != nil {
			return fmt.Errorf("create conversation: %w", createErr)
		}

		// Phase 6C: Add initial messages from template
		if template != nil && len(template.InitialMessages) > 0 {
			for _, msg := range template.InitialMessages {
				uiModel.AddMessage(msg.Role, msg.Content)
			}
		}
	}

	// Phase 4: Initialize service layer
	// Services provide an abstraction over direct storage calls
	convSvc := services.NewConversationService(db)
	msgSvc := services.NewMessageService(db)

	// Task 11: Create API client
	// For now, TUI mode only supports Anthropic (which uses core.Client with both CreateMessage and CreateMessageStream)
	// Other providers will be supported once we extend the Provider interface
	if providerName != "anthropic" {
		return fmt.Errorf("TUI mode currently only supports Anthropic provider. Use --provider=anthropic or print mode for other providers")
	}

	// Get provider config for Anthropic
	providerCfg, ok := cfg.ProviderConfigs[providerName]
	if !ok {
		return fmt.Errorf("no configuration found for provider '%s'", providerName)
	}

	// Check for environment variable
	if envKey := os.Getenv("ANTHROPIC_API_KEY"); envKey != "" {
		providerCfg.APIKey = envKey
	}

	// Validate API key
	if providerCfg.APIKey == "" {
		return fmt.Errorf("API key not configured for Anthropic. Set ANTHROPIC_API_KEY environment variable or add to config")
	}

	// Create core.Client for Anthropic (has both CreateMessage and CreateMessageStream)
	apiClient := core.NewClient(providerCfg.APIKey)

	// Set API client on UI model
	uiModel.SetAPIClient(apiClient)

	// Phase 4: Initialize AgentService (requires LLMClient with both methods)
	agentSvc := services.NewAgentService(apiClient, convSvc, msgSvc)

	// Phase 4: Set all services on UI model
	uiModel.SetServices(convSvc, msgSvc, agentSvc)

	// Task 12: Create tool registry and executor
	registry := tools.NewRegistry()
	if regErr := registry.Register(tools.NewReadTool()); regErr != nil {
		return fmt.Errorf("register read tool: %w", regErr)
	}
	if regErr := registry.Register(tools.NewWriteTool()); regErr != nil {
		return fmt.Errorf("register write tool: %w", regErr)
	}
	if regErr := registry.Register(tools.NewBashTool()); regErr != nil {
		return fmt.Errorf("register bash tool: %w", regErr)
	}
	// Phase 3: Register Edit, Grep, Glob tools
	if regErr := registry.Register(tools.NewEditTool()); regErr != nil {
		return fmt.Errorf("register edit tool: %w", regErr)
	}
	if regErr := registry.Register(tools.NewGrepTool()); regErr != nil {
		return fmt.Errorf("register grep tool: %w", regErr)
	}
	if regErr := registry.Register(tools.NewGlobTool()); regErr != nil {
		return fmt.Errorf("register glob tool: %w", regErr)
	}

	// Phase 4A: Register Interactive tools (AskUserQuestion, TodoWrite)
	if regErr := registry.Register(tools.NewAskUserQuestionTool()); regErr != nil {
		return fmt.Errorf("register ask_user_question tool: %w", regErr)
	}
	if regErr := registry.Register(tools.NewTodoWriteTool()); regErr != nil {
		return fmt.Errorf("register todo_write tool: %w", regErr)
	}

	// Phase 4B: Register Research tools (WebFetch, WebSearch)
	if regErr := registry.Register(tools.NewWebFetchTool()); regErr != nil {
		return fmt.Errorf("register web_fetch tool: %w", regErr)
	}
	if regErr := registry.Register(tools.NewWebSearchTool()); regErr != nil {
		return fmt.Errorf("register web_search tool: %w", regErr)
	}

	// Phase 4C: Register Advanced Execution tools (Task, BashOutput, KillShell)
	if regErr := registry.Register(tools.NewTaskTool()); regErr != nil {
		return fmt.Errorf("register task tool: %w", regErr)
	}
	if regErr := registry.Register(tools.NewBashOutputTool()); regErr != nil {
		return fmt.Errorf("register bash_output tool: %w", regErr)
	}
	if regErr := registry.Register(tools.NewKillShellTool()); regErr != nil {
		return fmt.Errorf("register kill_shell tool: %w", regErr)
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
	if regErr := registry.Register(skillTool); regErr != nil {
		return fmt.Errorf("register skill tool: %w", regErr)
	}
	logging.InfoWith("Loaded skills", "count", skillRegistry.Count())

	// Phase 4: Register Slash Commands system (with plugin commands)
	commandRegistry, slashCommandTool := initializeCommands(pluginCommandPaths)
	if regErr := registry.Register(slashCommandTool); regErr != nil {
		return fmt.Errorf("register slash command tool: %w", regErr)
	}
	logging.InfoWith("Loaded slash commands", "count", commandRegistry.Count())

	// Phase 5B: Load MCP tools from .mcp.json if present
	logging.Debug("Loading MCP tools")
	if mcpErr := mcp.LoadMCPTools(".", registry); mcpErr != nil {
		// Log error but don't fail - continue with built-in tools
		logging.WarnWith("Failed to load MCP tools", "error", mcpErr.Error())
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

	// Set slash commands for autocomplete
	cmdNames := commandRegistry.List()
	cmdDescriptions := make(map[string]string)
	for _, cmd := range commandRegistry.All() {
		cmdDescriptions[cmd.Name] = cmd.Description
	}
	uiModel.SetSlashCommands(cmdNames, cmdDescriptions)

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

	// Start Bubbletea program with appropriate options based on terminal availability
	var opts []tea.ProgramOption
	// Check if stdin is a terminal
	if term.IsTerminal(int(os.Stdin.Fd())) {
		// Use alt screen mode when running in a proper terminal
		opts = append(opts, tea.WithAltScreen())
	} else {
		// When not in a terminal, explicitly specify I/O
		opts = append(opts, tea.WithInput(os.Stdin), tea.WithOutput(os.Stdout))
	}

	p := tea.NewProgram(uiModel, opts...)
	if _, err := p.Run(); err != nil {
		// os.Stderr = origStderr // Restore stderr for error reporting (DEBUG)
		logging.ErrorWithErr("Failed to run UI", err)
		return fmt.Errorf("run UI: %w", err)
	}

	// Restore stderr before printing cost summary
	// os.Stderr = origStderr // DEBUG

	// Print cost summary if agent ID is set
	agentID := os.Getenv("HEX_AGENT_ID")
	if agentID != "" {
		cost.PrintCostSummary(agentID)
	}

	logging.Info("Hex shutting down")
	return nil
}

var globalLogger *logging.Logger

func initializeLogging() error {
	// If --debug is specified, force debug level and enable stderr output
	level := logging.LevelFromString(logLevel)
	if debug {
		level = logging.LevelDebug
		// Set environment variable so other packages know we're in debug mode
		_ = os.Setenv("HEX_DEBUG", "1") // Ignore error - not critical if env var fails to set
	}

	var format logging.Format
	switch logFormat {
	case "json":
		format = logging.FormatJSON
	default:
		format = logging.FormatText
	}

	config := logging.Config{
		Level:     level,
		Format:    format,
		AddSource: level == logging.LevelDebug, // Add source file/line in debug mode
	}

	var logger *logging.Logger
	var err error

	// In debug mode, write to file and optionally to stderr (only in print mode)
	if debug || level == logging.LevelDebug {
		debugFile := logFile
		if debugFile == "" {
			// Default debug log file
			debugFile = "/tmp/hex-debug.log"
		}
		config.LogFile = debugFile
		// Only write to stderr in print mode - in interactive mode it corrupts the TUI
		if printMode {
			config.Writer = os.Stderr
		}
		logger, err = logging.NewLoggerWithFile(config)
		if err != nil {
			return fmt.Errorf("failed to create debug logger with file %s: %w", debugFile, err)
		}
		// Only print debug message in print mode
		if printMode {
			fmt.Fprintf(os.Stderr, "Debug mode enabled. Logs writing to: %s\n", debugFile)
		}
	} else if logFile != "" {
		// Log to file (and optionally stderr in debug mode, but only in print mode)
		config.LogFile = logFile
		if level == logging.LevelDebug && printMode {
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

// createProvider creates the appropriate provider based on config and provider name
// Uses mux's battle-tested LLM clients wrapped in MuxAdapter
func createProvider(cfg *core.Config, providerName string) (providers.Provider, error) {
	// Get provider-specific config (may not exist for ollama which is configless)
	providerCfg, ok := cfg.ProviderConfigs[providerName]
	if !ok {
		providerCfg = core.ProviderConfig{} // Use empty config
	}

	// Check for provider-specific environment variables
	switch providerName {
	case "anthropic":
		if envKey := os.Getenv("ANTHROPIC_API_KEY"); envKey != "" {
			providerCfg.APIKey = envKey
		}
	case "openai":
		if envKey := os.Getenv("OPENAI_API_KEY"); envKey != "" {
			providerCfg.APIKey = envKey
		}
	case "gemini":
		if envKey := os.Getenv("GEMINI_API_KEY"); envKey != "" {
			providerCfg.APIKey = envKey
		}
	case "openrouter":
		if envKey := os.Getenv("OPENROUTER_API_KEY"); envKey != "" {
			providerCfg.APIKey = envKey
		}
	case "ollama":
		// Ollama is local, no API key needed
		// Use OLLAMA_HOST if set, otherwise default to localhost
		if host := os.Getenv("OLLAMA_HOST"); host != "" {
			providerCfg.BaseURL = host
		} else if providerCfg.BaseURL == "" {
			providerCfg.BaseURL = "http://localhost:11434/v1"
		}
	}

	// Validate we have an API key (except for ollama which is local)
	if providerName != "ollama" && providerCfg.APIKey == "" {
		return nil, fmt.Errorf("API key not configured for provider '%s'. Set %s_API_KEY environment variable or add to config",
			providerName, strings.ToUpper(providerName))
	}

	// Create appropriate mux LLM client and wrap in adapter
	// Note: model is passed per-request, not at client creation
	var client llm.Client
	switch providerName {
	case "anthropic":
		client = llm.NewAnthropicClient(providerCfg.APIKey, "")
	case "openai":
		client = llm.NewOpenAIClient(providerCfg.APIKey, "")
	case "gemini":
		ctx := context.Background()
		var err error
		client, err = llm.NewGeminiClient(ctx, providerCfg.APIKey, "")
		if err != nil {
			return nil, fmt.Errorf("create gemini client: %w", err)
		}
	case "openrouter":
		client = llm.NewOpenRouterClient(providerCfg.APIKey, "")
	case "ollama":
		client = llm.NewOllamaClient(providerCfg.BaseURL, "")
	default:
		return nil, fmt.Errorf("unknown provider: %s", providerName)
	}

	// Wrap mux client in adapter to implement hex's Provider interface
	return providers.NewMuxAdapter(providerName, client), nil
}
