// ABOUTME: Bootstrap functions for creating mux agents in hex.
// ABOUTME: Handles root agent and subagent creation with proper tool filtering.
package adapter

import (
	"context"
	"os"
	"strings"

	"github.com/2389-research/hex/internal/tools"
	"github.com/2389-research/mux/agent"
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
	ApprovalFunc ApprovalFunc // Optional: if nil, tools requiring approval will fail
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
	}

	// Wire up approval function if provided
	if cfg.ApprovalFunc != nil {
		agentCfg.ApprovalFunc = func(ctx context.Context, t muxtool.Tool, params map[string]any) (bool, error) {
			return cfg.ApprovalFunc(ctx, t.Name(), params)
		}
	}

	return agent.New(agentCfg)
}
