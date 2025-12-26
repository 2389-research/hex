// ABOUTME: Experimental mux agent runner for print and interactive modes.
// ABOUTME: Provides alternative agent implementation using the mux framework.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/2389-research/hex/internal/adapter"
	"github.com/2389-research/hex/internal/agentsmd"
	"github.com/2389-research/hex/internal/core"
	"github.com/2389-research/hex/internal/logging"
	"github.com/2389-research/hex/internal/tools"
	"github.com/2389-research/mux/orchestrator"
)

// runPrintModeWithMux runs print mode using the experimental mux agent framework.
func runPrintModeWithMux(prompt string) error {
	if prompt == "" && len(imagePaths) == 0 {
		return fmt.Errorf("prompt or image required in print mode")
	}

	// Image support not yet implemented for mux
	if len(imagePaths) > 0 {
		return fmt.Errorf("image input not yet supported with --experimental-mux")
	}

	ctx := context.Background()

	// Load config
	cfg, err := core.LoadConfig()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Get API key
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		if providerCfg, ok := cfg.ProviderConfigs["anthropic"]; ok {
			apiKey = providerCfg.APIKey
		}
	}
	if apiKey == "" {
		return fmt.Errorf("ANTHROPIC_API_KEY not set")
	}

	// Determine model
	modelToUse := model
	if modelToUse == "" {
		if cfg.Model != "" {
			modelToUse = cfg.Model
		} else {
			modelToUse = core.DefaultModel
		}
	}

	// Set up tools
	hexTools, err := getHexTools()
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

	// Create agent config
	agentCfg := adapter.Config{
		APIKey:       apiKey,
		Model:        modelToUse,
		SystemPrompt: sysPrompt,
		HexTools:     hexTools,
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

// getHexTools returns all hex tools as a slice for use with mux adapter.
func getHexTools() ([]tools.Tool, error) {
	// Create all core tools
	hexTools := []tools.Tool{
		tools.NewReadTool(),
		tools.NewWriteTool(),
		tools.NewEditTool(),
		tools.NewBashTool(),
		tools.NewGrepTool(),
		tools.NewGlobTool(),
	}

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
