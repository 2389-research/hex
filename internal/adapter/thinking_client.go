// ABOUTME: LLM client wrapper that injects extended thinking config into requests
// ABOUTME: Enables deep reasoning for complex tasks without modifying the mux orchestrator
package adapter

import (
	"context"

	"github.com/2389-research/mux/llm"
)

// ThinkingClient wraps an LLM client and injects ThinkingConfig into every request.
type ThinkingClient struct {
	inner  llm.Client
	budget int
}

// NewThinkingClient creates a client wrapper that enables extended thinking.
func NewThinkingClient(inner llm.Client, budgetTokens int) *ThinkingClient {
	return &ThinkingClient{
		inner:  inner,
		budget: budgetTokens,
	}
}

// CreateMessage injects thinking config and delegates to the inner client.
func (c *ThinkingClient) CreateMessage(ctx context.Context, req *llm.Request) (*llm.Response, error) {
	req.Thinking = &llm.ThinkingConfig{
		Enabled: true,
		Budget:  c.budget,
	}
	return c.inner.CreateMessage(ctx, req)
}

// CreateMessageStream injects thinking config and delegates to the inner client.
func (c *ThinkingClient) CreateMessageStream(ctx context.Context, req *llm.Request) (<-chan llm.StreamEvent, error) {
	req.Thinking = &llm.ThinkingConfig{
		Enabled: true,
		Budget:  c.budget,
	}
	return c.inner.CreateMessageStream(ctx, req)
}
