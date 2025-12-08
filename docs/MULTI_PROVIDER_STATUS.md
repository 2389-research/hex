# Multi-Provider Support Status

## Overview

Hex is in the process of adding support for multiple LLM providers (OpenAI, Gemini, OpenRouter) in addition to the existing Anthropic/Claude integration.

**Current Status**: Infrastructure complete, only Anthropic functional

## Completed Work (Tasks 1-10)

### ✅ Task 1: Provider Interface and Factory
- Created `Provider` interface matching orchestrator's `APIClient`
- Defined `ProviderConfig` for provider-specific configuration
- Created `Factory` for provider registration and retrieval
- **Location**: `internal/providers/provider.go`, `internal/providers/factory.go`

### ✅ Task 2-5: Provider Implementations (Removed)
- Initial implementations of Anthropic, OpenAI, Gemini, and OpenRouter providers were created
- **Removed in Task 8** when we discovered interface mismatch with orchestrator
- Need to be re-implemented using correct interface (see "Remaining Work" below)

### ✅ Task 6: Config System Update (YAML → TOML)
- Migrated from YAML to TOML format
- Added per-provider API key configuration
- Auto-migration from old config.yaml to config.toml
- Environment variable support for all providers
- **Location**: `internal/core/config.go`, `internal/core/migration.go`

**New config format**:
```toml
provider = "anthropic"  # Default provider
model = "claude-sonnet-4-5-20250929"

[providers.anthropic]
api_key = "sk-ant-..."

[providers.openai]
api_key = "sk-..."

[providers.gemini]
api_key = "..."

[providers.openrouter]
api_key = "..."
```

### ✅ Task 7: Database Migration for Provider Column
- Added `provider` column to `conversations` table
- Defaults to 'anthropic' for backward compatibility
- Migration files: `000003_add_provider_column.{up,down}.sql`
- **Location**: `internal/storage/migrations/`, `internal/storage/conversations.go`

### ✅ Task 8: Update Orchestrator to Use Provider Interface
- Defined `Provider` interface matching `APIClient` signature
- Created `AnthropicAdapter` wrapping `core.Client`
- Added `model` field to orchestrator (removed hardcoded model)
- Updated stream handler to use injected model
- **Location**: `internal/orchestrator/orchestrator.go`, `internal/providers/anthropic_adapter.go`

### ✅ Task 9: CLI Provider Flag and Validation
- Added `--provider` flag to CLI
- Provider precedence: flag > config > default ('anthropic')
- Validation with helpful error messages for unsupported providers
- Store provider in new conversations
- **Location**: `cmd/hex/root.go`

**Usage**:
```bash
# Default (anthropic)
hex "hello world"

# Explicit provider
hex --provider anthropic "hello"

# Unsupported provider (shows error)
hex --provider openai "test"
# Error: provider 'openai' not yet supported. Currently only 'anthropic'
# is available. Other providers (openai, gemini, openrouter) coming soon
```

### ✅ Task 10: Cost Tracking for Multiple Providers
- Added pricing for OpenAI models (gpt-4o, gpt-4o-mini, o3, o4-mini)
- Added pricing for Gemini models (gemini-2.5-flash, gemini-2.5-pro, etc.)
- Cost tracking will work automatically when providers are implemented
- **Location**: `internal/cost/pricing.go`

## Remaining Work

### 🔨 Task 11: Re-implement Provider Clients

The provider implementations need to be recreated with the correct interface.

**Required interface** (from `internal/providers/provider.go`):
```go
type Provider interface {
    // Must return channel of core.StreamChunk, not custom types
    CreateMessageStream(ctx context.Context, req core.MessageRequest) (<-chan *core.StreamChunk, error)

    Name() string
    ValidateConfig(cfg ProviderConfig) error
}
```

**What needs to be built**:
1. **Anthropic Provider** - Refactor existing `core.Client` into `internal/providers/anthropic/`
2. **OpenAI Provider** - Implement in `internal/providers/openai/`
   - Translate `core.MessageRequest` to OpenAI format
   - Parse SSE stream and convert to `core.StreamChunk`
   - Handle `choices[0].delta` structure
3. **Gemini Provider** - Implement in `internal/providers/gemini/`
   - Translate to Gemini's `contents`/`parts` structure
   - Handle JSON streaming format
   - Model ID in URL path, API key as query param
4. **OpenRouter Provider** - Implement in `internal/providers/openrouter/`
   - Can wrap OpenAI provider (OpenAI-compatible API)
   - Different base URL: `https://openrouter.ai/api/v1/chat/completions`

**Key requirement**: All providers must emit `core.StreamChunk` with proper structure for tool handling:
- `Type`: "message_start", "content_block_delta", "message_stop"
- `Delta.Text`: Incremental text content
- `ContentBlock`: Tool use information
- `Done`: Stream completion flag

### 🔨 Task 12: Update CLI to Use Provider Factory

Once providers are re-implemented, update CLI initialization:

**Current** (uses `core.Client` directly):
```go
client := core.NewClient(apiKey)
uiModel.SetAPIClient(client)
```

**Target** (use provider factory):
```go
// Create provider factory
factory := providers.NewFactory()

// Register all providers
anthropicProvider := anthropic.NewProvider(providers.ProviderConfig{
    APIKey:  cfg.ProviderConfigs["anthropic"].APIKey,
    BaseURL: cfg.ProviderConfigs["anthropic"].BaseURL,
})
factory.Register("anthropic", anthropicProvider)

openaiProvider := openai.NewProvider(providers.ProviderConfig{
    APIKey: cfg.ProviderConfigs["openai"].APIKey,
})
factory.Register("openai", openaiProvider)

// Get selected provider
provider, err := factory.GetProvider(providerName)
if err != nil {
    return fmt.Errorf("get provider: %w", err)
}

// Use provider (it already implements APIClient interface!)
uiModel.SetAPIClient(provider)
```

**Locations to update**:
- `cmd/hex/root.go` - Interactive mode
- `cmd/hex/interactive.go` - Shared interactive setup

### 🔨 Task 13: Integration Testing

Once providers are implemented, test:
- Switching between providers with `--provider` flag
- Each provider's streaming behavior
- Tool use with each provider
- Cost tracking for each provider
- Config file provider settings

### 🔨 Task 14: Documentation Updates

Update user-facing documentation:
- README with `--provider` flag examples
- Config file format guide
- Provider-specific setup instructions
- Migration guide from single-provider to multi-provider config

## Architecture Summary

### Data Flow

```
CLI (root.go)
  ├─> Load config (TOML)
  ├─> Validate provider
  ├─> Create Provider instance
  │     ├─> AnthropicProvider (ready)
  │     ├─> OpenAIProvider (todo)
  │     ├─> GeminiProvider (todo)
  │     └─> OpenRouterProvider (todo)
  │
  └─> Pass Provider to UI Model
        └─> AgentService uses Provider
              └─> Orchestrator uses Provider
                    └─> StreamHandler calls CreateMessageStream()
                          └─> Returns <-chan *core.StreamChunk
```

### Key Design Decisions

1. **Provider interface matches APIClient** - Providers can be used directly as API clients
2. **Central streaming format** - All providers must convert to `core.StreamChunk`
3. **Config precedence** - Flag > config file > default
4. **Backward compatibility** - Existing code works with AnthropicAdapter

## Testing Strategy

### Manual Testing Checklist

- [ ] Default provider (anthropic) works without flags
- [ ] `--provider anthropic` explicit flag works
- [ ] `--provider openai` shows helpful error (not yet implemented)
- [ ] Config file `provider = "anthropic"` is respected
- [ ] Provider is saved in conversation records
- [ ] Cost tracking shows correct prices for each model
- [ ] Old YAML configs auto-migrate to TOML

### Automated Testing

- [ ] Provider factory registration and retrieval
- [ ] Provider interface compliance tests
- [ ] Config loading with different provider settings
- [ ] Database migration up/down
- [ ] Cost calculation for all models

## Dependencies

### External Packages
- `github.com/pelletier/go-toml/v2` - TOML config parsing
- `github.com/golang-migrate/migrate/v4` - Database migrations

### Internal Packages
- `internal/core` - Core types (MessageRequest, StreamChunk, etc.)
- `internal/providers` - Provider interface and implementations
- `internal/storage` - Database operations with provider column
- `internal/cost` - Cost tracking with multi-provider pricing

## Migration Notes for Users

### Existing Users

If you're currently using hex with Anthropic:
1. **No action required** - Old config.yaml will auto-migrate
2. Provider defaults to 'anthropic'
3. Existing conversations continue to work

### New Config Format

The new `~/.hex/config.toml` format:
```toml
# Global settings
provider = "anthropic"  # Which provider to use
model = "claude-sonnet-4-5-20250929"

# Provider-specific API keys
[providers.anthropic]
api_key = "sk-ant-..."  # Or set ANTHROPIC_API_KEY env var

[providers.openai]
api_key = "sk-..."  # Or set OPENAI_API_KEY env var

[providers.gemini]
api_key = "..."  # Or set GEMINI_API_KEY env var

[providers.openrouter]
api_key = "..."  # Or set OPENROUTER_API_KEY env var
```

### Environment Variables

All providers support environment variables:
- `ANTHROPIC_API_KEY`
- `OPENAI_API_KEY`
- `GEMINI_API_KEY`
- `OPENROUTER_API_KEY`

Environment variables take precedence over config file.

## Timeline

- **December 7, 2025**: Tasks 1-6 completed (infrastructure and config)
- **December 8, 2025**: Tasks 7-10 completed (database, orchestrator, CLI, cost tracking)
- **Next**: Tasks 11-14 (re-implement providers, testing, docs)

## Contact

For questions or issues related to multi-provider support:
- Check this document for current status
- See implementation in `internal/providers/`
- Reference design doc: `docs/plans/2025-12-08-multi-provider-support-design.md`
