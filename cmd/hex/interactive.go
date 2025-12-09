// ABOUTME: Shared interactive mode setup logic
// ABOUTME: Used by both main interactive command and resume command
package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/2389-research/hex/internal/agentsmd"
	ctxmgr "github.com/2389-research/hex/internal/convcontext"
	"github.com/2389-research/hex/internal/core"
	"github.com/2389-research/hex/internal/logging"
	"github.com/2389-research/hex/internal/mcp"
	"github.com/2389-research/hex/internal/services"
	"github.com/2389-research/hex/internal/tools"
	"github.com/2389-research/hex/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"
)

// continueInteractiveWithModel runs the interactive TUI with a pre-configured model
// This is used by both normal interactive mode and the resume command
func continueInteractiveWithModel(db *sql.DB, uiModel *ui.Model, initialPrompt string) error {
	// Load AGENTS.md context from directory hierarchy (repo root → CWD)
	agentsContext, err := agentsmd.LoadContext()
	if err != nil {
		logging.WarnWith("Failed to load AGENTS.md context", "error", err.Error())
	} else if agentsContext != "" {
		// Set AGENTS.md context as system prompt
		// Note: Resumed conversations may already have a system prompt from their original session
		// AGENTS.md context is loaded fresh each time to reflect current directory context
		uiModel.SetSystemPrompt(agentsContext)
		logging.InfoWith("Loaded AGENTS.md context", "length", len(agentsContext))
	}

	// Phase 4: Initialize service layer
	convSvc := services.NewConversationService(db)
	msgSvc := services.NewMessageService(db)

	// Load config to get API key
	cfg, err := core.LoadConfig()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Get API key from config (handles all providers and env vars)
	apiKey, _ := cfg.GetAPIKey() // Ignore error - we'll check if empty below

	if apiKey != "" {
		client := core.NewClient(apiKey)
		uiModel.SetAPIClient(client)

		// Phase 4: Initialize AgentService (requires client)
		agentSvc := services.NewAgentService(client, convSvc, msgSvc)

		// Phase 4: Set all services on UI model
		uiModel.SetServices(convSvc, msgSvc, agentSvc)
	} else {
		return fmt.Errorf("API key not configured. Run 'hex setup-token <key>' or set ANTHROPIC_API_KEY environment variable")
	}

	// Create tool registry and executor
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
	if regErr := registry.Register(tools.NewEditTool()); regErr != nil {
		return fmt.Errorf("register edit tool: %w", regErr)
	}
	if regErr := registry.Register(tools.NewGrepTool()); regErr != nil {
		return fmt.Errorf("register grep tool: %w", regErr)
	}
	if regErr := registry.Register(tools.NewGlobTool()); regErr != nil {
		return fmt.Errorf("register glob tool: %w", regErr)
	}
	if regErr := registry.Register(tools.NewAskUserQuestionTool()); regErr != nil {
		return fmt.Errorf("register ask_user_question tool: %w", regErr)
	}
	if regErr := registry.Register(tools.NewTodoWriteTool()); regErr != nil {
		return fmt.Errorf("register todo_write tool: %w", regErr)
	}
	if regErr := registry.Register(tools.NewWebFetchTool()); regErr != nil {
		return fmt.Errorf("register web_fetch tool: %w", regErr)
	}
	if regErr := registry.Register(tools.NewWebSearchTool()); regErr != nil {
		return fmt.Errorf("register web_search tool: %w", regErr)
	}
	if regErr := registry.Register(tools.NewTaskTool()); regErr != nil {
		return fmt.Errorf("register task tool: %w", regErr)
	}
	if regErr := registry.Register(tools.NewBashOutputTool()); regErr != nil {
		return fmt.Errorf("register bash_output tool: %w", regErr)
	}
	if regErr := registry.Register(tools.NewKillShellTool()); regErr != nil {
		return fmt.Errorf("register kill_shell tool: %w", regErr)
	}

	// Initialize plugin system
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

	// Register Skills system (with plugin skills)
	skillRegistry, skillTool := initializeSkills(pluginSkillPaths)
	if regErr := registry.Register(skillTool); regErr != nil {
		return fmt.Errorf("register skill tool: %w", regErr)
	}
	logging.InfoWith("Loaded skills", "count", skillRegistry.Count())

	// Register Slash Commands system (with plugin commands)
	commandRegistry, slashCommandTool := initializeCommands(pluginCommandPaths)
	if regErr := registry.Register(slashCommandTool); regErr != nil {
		return fmt.Errorf("register slash command tool: %w", regErr)
	}
	logging.InfoWith("Loaded slash commands", "count", commandRegistry.Count())

	// Load MCP tools from .mcp.json if present
	logging.Debug("Loading MCP tools")
	if mcpErr := mcp.LoadMCPTools(".", registry); mcpErr != nil {
		// Log error but don't fail - continue with built-in tools
		logging.WarnWith("Failed to load MCP tools", "error", mcpErr.Error())
	} else {
		logging.Info("MCP tools loaded successfully")
	}

	// Create permission checker from flags
	permChecker, err := createPermissionChecker()
	if err != nil {
		return fmt.Errorf("create permission checker: %w", err)
	}

	// Create executor with approval function
	approvalFunc := func(_ string, _ map[string]interface{}) bool {
		return true
	}
	executor := tools.NewExecutor(registry, approvalFunc)

	// Attach permission checker to executor
	if permChecker != nil {
		executor.SetPermissionChecker(permChecker)
		logging.InfoWith("Permission system enabled", "mode", permChecker.GetMode())
	}

	// Set tool system in model
	uiModel.SetToolSystem(registry, executor)

	// Set up context manager
	contextManager := ctxmgr.NewManager(maxContextTokens)
	uiModel.SetContextManager(contextManager)
	logging.DebugWith("Context manager initialized", "maxTokens", maxContextTokens, "strategy", contextStrategy)

	// Add initial prompt if provided
	if initialPrompt != "" {
		uiModel.AddMessage("user", initialPrompt)
	}

	// Suppress debug logs to stderr during TUI operation to avoid display corruption
	// Save original stderr and redirect to /dev/null
	origStderr := os.Stderr
	devNull, err := os.Open(os.DevNull)
	if err == nil {
		os.Stderr = devNull
		defer func() {
			os.Stderr = origStderr
			_ = devNull.Close()
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
		os.Stderr = origStderr // Restore stderr for error reporting
		logging.ErrorWithErr("Failed to run UI", err)
		return fmt.Errorf("run UI: %w", err)
	}

	logging.Info("Hex shutting down")
	return nil
}
