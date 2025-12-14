// ABOUTME: Bootstrap functions for creating mux agents in hex.
// ABOUTME: Handles root agent and subagent creation with proper tool filtering.
package adapter

import (
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

// Config holds configuration for creating an agent.
type Config struct {
	APIKey       string
	Model        string
	SystemPrompt string
	HexTools     []tools.Tool
}

// NewRootAgent creates a root agent with full tool access.
func NewRootAgent(cfg Config) *agent.Agent {
	llmClient := llm.NewAnthropicClient(cfg.APIKey, cfg.Model)

	registry := muxtool.NewRegistry()
	for _, hexTool := range cfg.HexTools {
		registry.Register(AdaptTool(hexTool))
	}

	return agent.New(agent.Config{
		Name:         "hex-root",
		Registry:     registry,
		LLMClient:    llmClient,
		SystemPrompt: cfg.SystemPrompt,
	})
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

	return agent.New(agent.Config{
		Name:         agentID,
		Registry:     registry,
		LLMClient:    llmClient,
		SystemPrompt: cfg.SystemPrompt,
		AllowedTools: allowed,
		DeniedTools:  denied,
	})
}
