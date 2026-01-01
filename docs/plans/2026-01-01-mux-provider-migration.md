# Migrate to mux LLM Providers and MCP Client

**Date:** 2026-01-01
**Status:** In Progress

## Overview

Migrate hex from custom LLM provider implementations to mux's battle-tested clients, and replace hex's custom MCP client with mux's implementation. This reduces maintenance burden and gives hex automatic improvements when mux is upgraded.

## Goals

1. Replace 4 custom LLM providers (~800 lines) with mux adapters
2. Replace custom MCP client (~350 lines) with mux's client
3. Add Ollama as new provider
4. Maintain backward compatibility for users

## Current State

### LLM Providers (custom implementations)
- `internal/providers/gemini/provider.go` - Raw HTTP, ~290 lines
- `internal/providers/openai/provider.go` - Custom impl, ~250 lines
- `internal/providers/openrouter/provider.go` - Custom impl, ~200 lines
- `internal/providers/anthropic_adapter.go` - Wraps core.Client

### MCP Client
- `internal/mcp/client.go` - JSON-RPC 2.0 over stdio, ~350 lines

## Target State

### LLM Providers (mux adapters)
- `internal/providers/mux_adapter.go` - Single adapter wrapping any `llm.Client`
- Factory creates appropriate mux client and wraps in adapter

### MCP Client
- Use `github.com/2389-research/mux/mcp.Client` directly
- Update `internal/mcp/loader.go` to use mux types

## Design

### MuxAdapter

```go
// MuxAdapter wraps a mux llm.Client to implement hex's Provider interface
type MuxAdapter struct {
    client llm.Client
    name   string
}

func (a *MuxAdapter) CreateMessage(ctx context.Context, req core.MessageRequest) (*core.MessageResponse, error) {
    // Translate core.MessageRequest → llm.Request
    // Call a.client.CreateMessage()
    // Translate llm.Response → core.MessageResponse
}

func (a *MuxAdapter) CreateMessageStream(ctx context.Context, req core.MessageRequest) (<-chan *core.StreamChunk, error) {
    // Translate core.MessageRequest → llm.Request
    // Call a.client.CreateMessageStream()
    // Goroutine to translate llm.StreamEvent → core.StreamChunk
}
```

### Provider Factory

```go
func createProvider(cfg *core.Config, providerName string) (providers.Provider, error) {
    switch providerName {
    case "anthropic":
        client := llm.NewAnthropicClient(apiKey, model)
        return providers.NewMuxAdapter("anthropic", client), nil
    case "openai":
        client := llm.NewOpenAIClient(apiKey, model)
        return providers.NewMuxAdapter("openai", client), nil
    case "gemini":
        client, _ := llm.NewGeminiClient(ctx, apiKey, model)
        return providers.NewMuxAdapter("gemini", client), nil
    case "openrouter":
        client := llm.NewOpenRouterClient(apiKey, model)
        return providers.NewMuxAdapter("openrouter", client), nil
    case "ollama":
        client := llm.NewOllamaClient(baseURL, model)
        return providers.NewMuxAdapter("ollama", client), nil
    }
}
```

### MCP Loader Changes

```go
// Before
client, _ := mcp.NewClient(command, args...)
client.Initialize(ctx, "hex", "0.1.0", "2024-11-05")

// After
client := muxmcp.NewClient(muxmcp.ServerConfig{
    Name:    name,
    Command: command,
    Args:    args,
})
client.Start(ctx)
```

## Files Changed

### Deleted
- `internal/providers/gemini/provider.go`
- `internal/providers/gemini/provider_test.go`
- `internal/providers/openai/provider.go`
- `internal/providers/openai/provider_test.go`
- `internal/providers/openrouter/provider.go`
- `internal/providers/openrouter/provider_test.go`
- `internal/mcp/client.go`
- `internal/mcp/client_test.go`

### Added
- `internal/providers/mux_adapter.go`
- `internal/providers/mux_adapter_test.go`

### Modified
- `internal/providers/factory.go`
- `internal/mcp/loader.go`
- `internal/mcp/tool_adapter.go`
- `cmd/hex/root.go` (add ollama)
- `go.mod` (mux v0.5.1)

## Implementation Order

1. Upgrade mux to v0.5.1
2. Create MuxAdapter
3. Migrate providers one by one (anthropic → openai → gemini → openrouter)
4. Add Ollama
5. Migrate MCP client
6. Delete old code
7. Test everything

## Rollback Plan

If issues arise, revert to previous commit. The old provider code will still be in git history.
