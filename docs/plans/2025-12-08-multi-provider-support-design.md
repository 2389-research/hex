# Multi-Provider Support Design

**Date:** 2025-12-08
**Status:** Design Complete, Ready for Implementation
**Scope:** Add support for OpenAI, Gemini, and OpenRouter alongside existing Anthropic support

## Problem Statement

Hex currently hardcodes Anthropic's Claude API. Users cannot use other model providers like OpenAI (GPT-4o, o3), Google Gemini, or OpenRouter. This limits flexibility and prevents users from choosing models based on task requirements, cost, or availability.

## Goals

1. Support four providers: Anthropic, OpenAI, Gemini, OpenRouter
2. Maintain backward compatibility with existing Anthropic configurations
3. Allow explicit provider selection via CLI flag and config file
4. Preserve existing tool calling, streaming, and cost tracking features
5. Use provider abstraction to simplify future provider additions

## Non-Goals (Deferred)

- Automatic model tier selection (reasoning/high/low complexity)
- Multi-provider fallback mechanisms
- Provider-specific feature detection
- Model capability introspection

## Architecture

### Provider Abstraction Layer

Create a `Provider` interface that abstracts API-specific details:

```go
type Provider interface {
    CreateStream(ctx context.Context, req *MessageRequest) (Stream, error)
    Name() string
    ValidateConfig(cfg ProviderConfig) error
}

type Stream interface {
    Next() (*StreamChunk, error)
    Close() error
}
```

Hex uses a unified internal message format. Providers translate between this format and their API-specific formats.

### Directory Structure

```
internal/providers/
├── provider.go           # Provider interface
├── anthropic/
│   ├── client.go        # Refactored from internal/core/client.go
│   ├── stream.go        # Anthropic SSE handling
│   └── types.go         # Message format translation
├── openai/
│   ├── client.go        # OpenAI Chat Completions API
│   ├── stream.go        # OpenAI SSE handling
│   └── types.go         # Message format translation
├── gemini/
│   ├── client.go        # Gemini API
│   ├── stream.go        # Gemini streaming
│   └── types.go         # Message format translation
├── openrouter/
│   ├── client.go        # OpenRouter (wraps OpenAI format)
│   └── types.go         # Model ID handling
└── factory.go           # Provider registry and instantiation
```

### Configuration

**File Location:** `~/.hex/config.toml` (migrates from existing `config.yaml`)

**Structure:**
```toml
# Default provider
provider = "anthropic"

# Default model
model = "claude-sonnet-4-5-20250929"

# Existing settings
permission_mode = "ask"
default_tools = ["Bash", "Read", "Write", "Edit", "Grep"]

[providers.anthropic]
api_key = "sk-ant-..."

[providers.openai]
api_key = "sk-proj-..."

[providers.gemini]
api_key = "AIza..."

[providers.openrouter]
api_key = "sk-or-..."
```

**Environment Variables:**
- `HEX_PROVIDER` - Override default provider
- `HEX_MODEL` - Override default model
- `ANTHROPIC_API_KEY` - Anthropic API key
- `OPENAI_API_KEY` - OpenAI API key
- `GEMINI_API_KEY` - Gemini API key
- `OPENROUTER_API_KEY` - OpenRouter API key

**CLI Flags:**
- `--provider <name>` - Select provider (anthropic|openai|gemini|openrouter)
- `--model <id>` - Select model (existing flag)

**Precedence (highest to lowest):**
1. CLI flags
2. Environment variables (HEX_*)
3. Standard provider environment variables
4. Config file
5. Defaults

### Provider-Specific Details

**Anthropic** (refactor existing)
- Endpoint: `https://api.anthropic.com/v1/messages`
- Auth: `x-api-key` header
- Streaming: SSE with `message_start`, `content_block_delta` events
- Move existing implementation from `internal/core/` to `internal/providers/anthropic/`

**OpenAI** (new)
- Endpoint: `https://api.openai.com/v1/chat/completions`
- Auth: `Authorization: Bearer <token>`
- Message format: `{"role": "user", "content": "..."}`
- Streaming: SSE with `data: {"choices": [{"delta": {...}}]}`
- Tool calling: "functions" format (translate from hex tools)

**Gemini** (new)
- Endpoint: `https://generativelanguage.googleapis.com/v1/models/{model}:streamGenerateContent`
- Auth: `x-goog-api-key` header or `?key=<key>` query param
- Message format: `{"contents": [{"role": "user", "parts": [...]}]}`
- Streaming: JSON streaming format
- Tool calling: "function declarations" format

**OpenRouter** (new)
- Endpoint: `https://openrouter.ai/api/v1/chat/completions`
- Auth: `Authorization: Bearer <token>`
- Format: OpenAI-compatible (reuse OpenAI translation)
- Model IDs: `provider/model` format (e.g., `anthropic/claude-sonnet-4-5`)

## Integration Points

### Changes Required

**1. Orchestrator** (`internal/orchestrator/stream_handler.go`)
- Remove hardcoded model on line 24
- Accept Provider interface instead of direct client
- Inject provider at initialization

**2. Services** (`internal/services/agent_impl.go`)
- Rename `LLMClient` to `Provider`
- Expand interface for provider-specific methods
- No major changes needed (already uses dependency injection)

**3. Commands** (`cmd/hex/root.go`, `cmd/hex/print.go`)
- Add provider selection logic
- Use factory to instantiate provider based on config
- Pass provider to orchestrator/services

**4. Config** (`internal/core/config.go`)
- Switch from YAML to TOML
- Add `Provider` field
- Add `ProviderConfigs map[string]ProviderConfig`
- Validate that selected provider has API key configured

**5. Storage** (`internal/storage/`)
- Add `Provider` column to conversations table
- Store provider with each conversation
- Resume conversations with original provider

**6. Cost Tracking** (`internal/cost/pricing.go`)
- Add pricing for OpenAI models (GPT-4o, o3, etc.)
- Add pricing for Gemini models (2.5 Pro, 2.5 Flash, etc.)
- Provider-aware cost calculation

## Backward Compatibility

- Existing `ANTHROPIC_API_KEY` continues to work
- Existing `--model claude-*` defaults to Anthropic provider
- Existing conversations without provider field default to Anthropic
- If `~/.hex/config.yaml` exists, migrate to `.toml` on first run (preserve settings)

## Migration Strategy

1. Create Provider interface and factory
2. Refactor Anthropic code into `internal/providers/anthropic/`
3. Update orchestrator/services to use Provider interface
4. Implement OpenAI provider
5. Implement Gemini provider
6. Implement OpenRouter provider
7. Update config loading (TOML, provider configs)
8. Add provider selection to CLI
9. Update cost tracking with new provider pricing
10. Add database migration for provider column

## Testing

### Unit Tests
- Message translation (hex format ↔ provider format)
- Streaming parsing
- Error handling
- Each provider in isolation

### Integration Tests
- Real API calls to each provider
- End-to-end: request → streaming → tool calling
- Cost tracking accuracy

### Scenario Tests (using hex)
```bash
# Test provider switching
./hex --provider openai --model gpt-4o "test message"
./hex --provider gemini --model gemini-2.5-flash "test message"

# Test config file
./hex "test message"  # Uses config default

# Test conversation resumption
./hex --resume <id>   # Resumes with original provider
```

### Dogfooding
Use hex with each provider to develop the next feature.

## Verification Checklist

- [ ] Send message via Anthropic (already works)
- [ ] Send message via OpenAI
- [ ] Send message via Gemini
- [ ] Send message via OpenRouter
- [ ] Tool calling works on all providers
- [ ] Streaming works on all providers
- [ ] Config file loads provider settings
- [ ] CLI flags override provider/model
- [ ] Cost tracking accurate for all providers
- [ ] Resume conversation preserves provider
- [ ] Backward compatibility maintained

## Implementation

Use `superpowers:subagent-driven-development` to dispatch subagents for:
1. Provider abstraction and factory
2. Anthropic refactor
3. OpenAI implementation
4. Gemini implementation
5. OpenRouter implementation
6. Config system updates
7. Database migration
8. Cost tracking updates

Each subagent implements one component with tests and verification.

## Success Criteria

1. Users can select provider with `--provider` flag
2. Users can configure provider API keys in `~/.hex/config.toml`
3. Tool calling works across all providers
4. Streaming works across all providers
5. Cost tracking accurate for all providers
6. Existing Anthropic users experience no breaking changes
7. All tests pass
