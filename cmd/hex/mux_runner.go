// ABOUTME: Mux agent runner for print mode.
// ABOUTME: Default agent implementation using the mux framework.
package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/2389-research/hex/internal/adapter"
	"github.com/2389-research/hex/internal/agentsmd"
	"github.com/2389-research/hex/internal/core"
	"github.com/2389-research/hex/internal/hooks"
	"github.com/2389-research/hex/internal/logging"
	"github.com/2389-research/hex/internal/permissions"
	"github.com/2389-research/hex/internal/tools"
	"github.com/2389-research/mux/llm"
	"github.com/2389-research/mux/orchestrator"
)

// runPrintModeWithMux runs print mode using the mux agent framework.
func runPrintModeWithMux(prompt string) error {
	if prompt == "" && len(imagePaths) == 0 {
		return fmt.Errorf("prompt or image required in print mode")
	}

	// Image support not yet implemented for mux mode
	if len(imagePaths) > 0 {
		return fmt.Errorf("image input not yet supported in mux mode (use --legacy flag for image support)")
	}

	ctx := context.Background()

	// Load config
	cfg, err := core.LoadConfig()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Determine provider (flag > config > default)
	providerName := provider
	if providerName == "" {
		providerName = cfg.Provider
	}
	if providerName == "" {
		providerName = "anthropic"
	}

	// Determine model
	modelToUse := model
	if modelToUse == "" {
		if cfg.Model != "" {
			modelToUse = cfg.Model
		} else if providerName == "anthropic" {
			modelToUse = core.DefaultModel
		} else {
			return fmt.Errorf("--model flag is required when using --provider=%s", providerName)
		}
	}

	// Create LLM client based on provider
	llmClient, apiKey, err := createMuxLLMClient(cfg, providerName, modelToUse)
	if err != nil {
		return fmt.Errorf("create LLM client: %w", err)
	}

	// Set up tools with mux-based subagent support
	hexTools, err := getHexToolsWithMuxSubagents(llmClient)
	if err != nil {
		return fmt.Errorf("setup tools: %w", err)
	}

	// Build system prompt
	sysPrompt := core.DefaultSystemPrompt
	if systemPrompt != "" {
		sysPrompt = sysPrompt + "\n\n" + systemPrompt
	}

	// Load AGENTS.md context
	agentsContext, err := agentsmd.LoadContext()
	if err != nil {
		logging.WarnWith("Failed to load AGENTS.md context", "error", err.Error())
	} else if agentsContext != "" {
		sysPrompt = agentsContext + "\n\n" + sysPrompt
	}

	// Load project memory context
	projContext := loadProjectContext()
	if projContext != "" {
		sysPrompt = sysPrompt + "\n\n" + projContext
	}

	// Create approval function based on permission mode
	approvalFunc, err := createMuxApprovalFunc()
	if err != nil {
		return fmt.Errorf("create approval function: %w", err)
	}

	// Create hook engine and mux bridge for lifecycle events
	cwd, _ := os.Getwd()
	hookEngine, err := hooks.NewEngine(cwd, modelToUse)
	if err != nil {
		logging.WarnWith("Failed to create hook engine", "error", err.Error())
		// Continue without hooks - they're optional
	}

	// Create mux bridge to forward mux events to hex hooks
	var muxHookManager *hooks.MuxBridge
	if hookEngine != nil {
		muxHookManager, _ = hooks.NewMuxBridge(hookEngine)
	}

	// Create agent config
	agentCfg := adapter.Config{
		APIKey:       apiKey,
		Model:        modelToUse,
		SystemPrompt: sysPrompt,
		HexTools:     hexTools,
		ApprovalFunc: approvalFunc,
	}

	// Wire up hooks if available
	if muxHookManager != nil {
		agentCfg.HookManager = muxHookManager.Manager()
	}

	// Create agent (root or subagent based on environment)
	var agent interface {
		Run(ctx context.Context, prompt string) error
		Subscribe() <-chan orchestrator.Event
	}

	if adapter.IsSubagent() {
		agent = adapter.NewSubagent(agentCfg)
		logging.InfoWith("Created mux subagent", "model", modelToUse)
	} else {
		agent = adapter.NewRootAgent(agentCfg)
		logging.InfoWith("Created mux root agent", "model", modelToUse)
	}

	// Subscribe to events before running
	events := agent.Subscribe()

	// Run agent in goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- agent.Run(ctx, prompt)
	}()

	// Process events
	var finalText string
	for event := range events {
		switch event.Type {
		case orchestrator.EventText:
			// Stream text to stdout
			fmt.Print(event.Text)

		case orchestrator.EventToolCall:
			logging.InfoWith("Tool call", "name", event.ToolName, "id", event.ToolID)

		case orchestrator.EventToolResult:
			if event.Result != nil {
				logging.DebugWith("Tool result", "output_len", len(event.Result.Output))
			}

		case orchestrator.EventComplete:
			finalText = event.FinalText

		case orchestrator.EventError:
			if event.Error != nil {
				return fmt.Errorf("agent error: %w", event.Error)
			}

		case orchestrator.EventStateChange:
			logging.DebugWith("State change", "from", event.FromState, "to", event.ToState)
		}
	}

	// Wait for agent to complete
	if err := <-errChan; err != nil {
		return fmt.Errorf("agent run failed: %w", err)
	}

	// Ensure newline after streamed output
	if finalText != "" {
		fmt.Println()
	}

	return nil
}

// getHexToolsWithMuxSubagents returns hex tools with mux-based subagent support.
// The TaskTool is configured to use mux agents instead of subprocesses.
func getHexToolsWithMuxSubagents(llmClient llm.Client) ([]tools.Tool, error) {
	// Create the TaskTool first without mux config
	taskTool := tools.NewTaskTool()

	// Create base tools
	baseTools := []tools.Tool{
		tools.NewReadTool(),
		tools.NewWriteTool(),
		tools.NewEditTool(),
		tools.NewBashTool(),
		tools.NewGrepTool(),
		tools.NewGlobTool(),
	}

	// Tool factory for subagents - creates fresh tools excluding TaskTool to avoid recursion issues
	toolFactory := func() []tools.Tool {
		return []tools.Tool{
			tools.NewReadTool(),
			tools.NewWriteTool(),
			tools.NewEditTool(),
			tools.NewBashTool(),
			tools.NewGrepTool(),
			tools.NewGlobTool(),
		}
	}

	// Create AgentRunner and wire it to TaskTool
	agentRunner := adapter.NewAgentRunner(llmClient, toolFactory)
	taskTool.SetMuxRunner(agentRunner)

	// Combine base tools with TaskTool
	hexTools := append(baseTools, taskTool)

	// Filter based on --tools flag if specified
	if len(enabledTools) > 0 {
		enabledSet := make(map[string]bool)
		for _, t := range enabledTools {
			enabledSet[t] = true
		}

		filtered := make([]tools.Tool, 0)
		for _, t := range hexTools {
			if enabledSet[t.Name()] {
				filtered = append(filtered, t)
			}
		}
		hexTools = filtered
		logging.InfoWith("Filtered tools for mux", "count", len(hexTools), "enabled", enabledTools)
	}

	return hexTools, nil
}

// createMuxLLMClient creates an LLM client for the specified provider.
// Returns the client, API key (for adapter.Config), and any error.
func createMuxLLMClient(cfg *core.Config, providerName, modelName string) (llm.Client, string, error) {
	// Get provider-specific config
	providerCfg, ok := cfg.ProviderConfigs[providerName]
	if !ok {
		providerCfg = core.ProviderConfig{}
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
		if host := os.Getenv("OLLAMA_HOST"); host != "" {
			providerCfg.BaseURL = host
		} else if providerCfg.BaseURL == "" {
			providerCfg.BaseURL = "http://localhost:11434/v1"
		}
	}

	// Validate API key (except for Ollama which is local)
	if providerName != "ollama" && providerCfg.APIKey == "" {
		return nil, "", fmt.Errorf("API key not configured for provider '%s'. Set %s_API_KEY environment variable or add to config",
			providerName, strings.ToUpper(providerName))
	}

	// Create client based on provider
	ctx := context.Background()
	switch providerName {
	case "anthropic":
		return llm.NewAnthropicClient(providerCfg.APIKey, modelName), providerCfg.APIKey, nil
	case "openai":
		return llm.NewOpenAIClient(providerCfg.APIKey, modelName), providerCfg.APIKey, nil
	case "gemini":
		client, err := llm.NewGeminiClient(ctx, providerCfg.APIKey, modelName)
		return client, providerCfg.APIKey, err
	case "openrouter":
		return llm.NewOpenRouterClient(providerCfg.APIKey, modelName), providerCfg.APIKey, nil
	case "ollama":
		return llm.NewOllamaClient(providerCfg.BaseURL, modelName), "", nil
	default:
		return nil, "", fmt.Errorf("unknown provider: %s", providerName)
	}
}

// createMuxApprovalFunc creates an approval function for mux based on permission mode.
func createMuxApprovalFunc() (adapter.ApprovalFunc, error) {
	// Parse permission mode from flag
	mode := permissionMode
	if dangerouslySkipPermissions {
		mode = "auto"
	}

	parsedMode, err := permissions.ParseMode(mode)
	if err != nil {
		return nil, fmt.Errorf("invalid permission mode: %w", err)
	}

	// Create permission rules from flags
	rules := permissions.NewRules(allowedTools, disallowedTools)
	checker := permissions.NewChecker(parsedMode, rules)

	logging.InfoWith("Mux approval function created", "mode", parsedMode)

	// Return approval function that uses the permission checker
	return func(_ context.Context, toolName string, params map[string]any) (bool, error) {
		// Convert params for checker (map[string]any to map[string]interface{})
		converted := make(map[string]interface{}, len(params))
		for k, v := range params {
			converted[k] = v
		}

		result := checker.Check(toolName, converted)

		// If requires prompt but we're in print mode, we can't ask interactively
		if result.RequiresPrompt {
			logging.WarnWith("Tool requires approval but running in non-interactive mode",
				"tool", toolName, "reason", result.Reason)
			return false, nil
		}

		return result.Allowed, nil
	}, nil
}
