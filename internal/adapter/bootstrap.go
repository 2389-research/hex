// ABOUTME: Bootstrap functions for creating mux agents in hex.
// ABOUTME: Handles root agent and subagent creation with proper tool filtering.
package adapter

import (
	"context"
	"os"
	"strings"

	"github.com/2389-research/hex/internal/tools"
	"github.com/2389-research/mux/agent"
	muxhooks "github.com/2389-research/mux/hooks"
	"github.com/2389-research/mux/llm"
	muxtool "github.com/2389-research/mux/tool"
)

// IsSubagent returns true if this process is running as a subagent.
func IsSubagent() bool {
	return os.Getenv("HEX_SUBAGENT_TYPE") != ""
}

// parseCSV parses a comma-separated string into a slice.
func parseCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// ApprovalFunc is called when a tool requires approval before execution.
// Returns true to approve, false to deny. Error indicates approval check failed.
type ApprovalFunc func(ctx context.Context, toolName string, params map[string]any) (bool, error)

// Config holds configuration for creating an agent.
type Config struct {
	APIKey       string
	Model        string
	SystemPrompt string
	HexTools     []tools.Tool
	ApprovalFunc ApprovalFunc      // Optional: if nil, tools requiring approval will fail
	HookManager  *muxhooks.Manager // Optional: mux hook manager for lifecycle events
}

// NewRootAgent creates a root agent with full tool access.
func NewRootAgent(cfg Config) *agent.Agent {
	llmClient := llm.NewAnthropicClient(cfg.APIKey, cfg.Model)

	registry := muxtool.NewRegistry()
	for _, hexTool := range cfg.HexTools {
		registry.Register(AdaptTool(hexTool))
	}

	agentCfg := agent.Config{
		Name:         "hex-root",
		Registry:     registry,
		LLMClient:    llmClient,
		SystemPrompt: cfg.SystemPrompt,
		HookManager:  cfg.HookManager, // Wire up hooks if provided
	}

	// Wire up approval function if provided
	if cfg.ApprovalFunc != nil {
		agentCfg.ApprovalFunc = func(ctx context.Context, t muxtool.Tool, params map[string]any) (bool, error) {
			return cfg.ApprovalFunc(ctx, t.Name(), params)
		}
	}

	return agent.New(agentCfg)
}

// NewSubagent creates a subagent with filtered tool access based on env vars.
func NewSubagent(cfg Config) *agent.Agent {
	llmClient := llm.NewAnthropicClient(cfg.APIKey, cfg.Model)

	registry := muxtool.NewRegistry()
	for _, hexTool := range cfg.HexTools {
		registry.Register(AdaptTool(hexTool))
	}

	allowed := parseCSV(os.Getenv("HEX_ALLOWED_TOOLS"))
	denied := parseCSV(os.Getenv("HEX_DENIED_TOOLS"))

	agentID := os.Getenv("HEX_AGENT_ID")
	if agentID == "" {
		agentID = "hex-subagent"
	}

	agentCfg := agent.Config{
		Name:         agentID,
		Registry:     registry,
		LLMClient:    llmClient,
		SystemPrompt: cfg.SystemPrompt,
		AllowedTools: allowed,
		DeniedTools:  denied,
		HookManager:  cfg.HookManager, // Wire up hooks if provided
	}

	// Wire up approval function if provided
	if cfg.ApprovalFunc != nil {
		agentCfg.ApprovalFunc = func(ctx context.Context, t muxtool.Tool, params map[string]any) (bool, error) {
			return cfg.ApprovalFunc(ctx, t.Name(), params)
		}
	}

	return agent.New(agentCfg)
}

// NewSubagentWithClient creates a subagent with an existing LLM client.
// This is useful for in-process subagent execution where we want to share/reuse the client.
func NewSubagentWithClient(cfg Config, llmClient llm.Client, agentID string) *agent.Agent {
	registry := muxtool.NewRegistry()
	for _, hexTool := range cfg.HexTools {
		registry.Register(AdaptTool(hexTool))
	}

	if agentID == "" {
		agentID = os.Getenv("HEX_AGENT_ID")
		if agentID == "" {
			agentID = "hex-subagent"
		}
	}

	agentCfg := agent.Config{
		Name:         agentID,
		Registry:     registry,
		LLMClient:    llmClient,
		SystemPrompt: cfg.SystemPrompt,
		HookManager:  cfg.HookManager, // Wire up hooks if provided
	}

	// Wire up approval function if provided
	if cfg.ApprovalFunc != nil {
		agentCfg.ApprovalFunc = func(ctx context.Context, t muxtool.Tool, params map[string]any) (bool, error) {
			return cfg.ApprovalFunc(ctx, t.Name(), params)
		}
	}

	return agent.New(agentCfg)
}

// AgentRunner implements the tools.MuxAgentRunner interface.
// It creates and runs mux agents for in-process subagent execution.
type AgentRunner struct {
	LLMClient    llm.Client          // LLM client for agent communication
	ToolFactory  func() []tools.Tool // Factory to create fresh tool instances
	ApprovalFunc func(ctx context.Context, toolName string, params map[string]any) (bool, error)
}

// NewAgentRunner creates a new AgentRunner for executing mux agents.
func NewAgentRunner(llmClient llm.Client, toolFactory func() []tools.Tool) *AgentRunner {
	return &AgentRunner{
		LLMClient:   llmClient,
		ToolFactory: toolFactory,
	}
}

// RunAgent executes a subagent with the given configuration and returns the output.
func (r *AgentRunner) RunAgent(ctx context.Context, agentID, prompt, systemPrompt string, allowedTools []string) (string, error) {
	// Create fresh tools
	hexTools := r.ToolFactory()

	// Filter tools if allowedTools is specified
	var filteredTools []tools.Tool
	if allowedTools == nil {
		filteredTools = hexTools
	} else {
		allowedSet := make(map[string]bool)
		for _, name := range allowedTools {
			allowedSet[strings.ToLower(name)] = true
		}
		for _, tool := range hexTools {
			if allowedSet[strings.ToLower(tool.Name())] {
				filteredTools = append(filteredTools, tool)
			}
		}
	}

	// Create mux registry with adapted tools
	registry := muxtool.NewRegistry()
	for _, hexTool := range filteredTools {
		registry.Register(AdaptTool(hexTool))
	}

	// Build agent config
	agentCfg := agent.Config{
		Name:         agentID,
		Registry:     registry,
		LLMClient:    r.LLMClient,
		SystemPrompt: systemPrompt,
	}

	// Wire up approval function if provided
	if r.ApprovalFunc != nil {
		agentCfg.ApprovalFunc = func(ctx context.Context, t muxtool.Tool, params map[string]any) (bool, error) {
			return r.ApprovalFunc(ctx, t.Name(), params)
		}
	}

	// Create and run agent
	a := agent.New(agentCfg)
	err := a.Run(ctx, prompt)

	// Extract output from last assistant message
	var output string
	messages := a.Messages()
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "assistant" {
			output = messages[i].Content
			break
		}
	}

	return output, err
}
