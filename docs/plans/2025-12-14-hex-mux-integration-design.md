# Hex-Mux Integration Design

## Overview

Hex will use mux as its core agent engine. This replaces hex's `AgentOrchestrator` with mux's `orchestrator.Orchestrator` while keeping hex's subprocess-based subagent model.

## Architecture

```
┌─────────────────────────────────────────────────┐
│  hex CLI (cmd/hex)                              │
│  - Parses args, loads config                    │
│  - Creates mux Agent with full tool registry    │
│  - Runs agent.Run() or agent.Subscribe()        │
└─────────────────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────┐
│  mux/agent.Agent                                │
│  - FilteredRegistry (full access for root)      │
│  - mux/orchestrator.Orchestrator                │
│  - mux/tool.Executor                            │
└─────────────────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────┐
│  hex/internal/adapter                           │
│  - Wraps hex tools → mux tool.Tool interface    │
│  - Registered with mux tool.Registry            │
└─────────────────────────────────────────────────┘
```

### Subagent Spawning

Subprocess model is preserved. Parent passes tool restrictions via environment variables:

```
Parent (hex process)              Child (hex subprocess)
        │                                  │
        │ spawn with env vars:             │
        │ HEX_ALLOWED_TOOLS=Read,Grep      │
        │ HEX_DENIED_TOOLS=Bash            │
        │ ────────────────────────────────►│
        │                                  │ Creates FilteredRegistry
        │                                  │ with restrictions
        │◄─────────────────────────────────│
        │ stdout result                    │
```

## Design Decisions

| Question | Decision | Rationale |
|----------|----------|-----------|
| Integration depth | Full orchestrator replacement | Single orchestrator reduces complexity |
| Tool adaptation | Adapter pattern | Minimal changes to existing hex tools |
| Subagent spawning | Keep subprocess model | Process isolation is valuable |
| Subprocess config | Env vars | Follows existing hex patterns |
| Orchestrator in child | Mux orchestrator | Consistency top-to-bottom |
| Root agent | Also uses mux | One code path for all agents |
| LLM client | Mux provides Anthropic client | Clean separation, reusable |
| Tool adapter location | Lives in hex | Clean dependency direction |

## Components

### New in mux

**`mux/llm/anthropic.go`** - Anthropic API client implementing `llm.Client`

```go
type AnthropicClient struct {
    client *anthropic.Client
    model  string
}

func NewAnthropicClient(apiKey, model string) *AnthropicClient

func (a *AnthropicClient) CreateMessage(ctx context.Context, req *Request) (*Response, error)
func (a *AnthropicClient) CreateMessageStream(ctx context.Context, req *Request) (<-chan StreamEvent, error)
```

Uses official `anthropic-sdk-go` package. Handles API auth, model selection, system prompts, tool definitions, streaming.

### New in hex

**`hex/internal/adapter/tool.go`** - Tool adapter

```go
type adaptedTool struct {
    hex tools.Tool
}

func (a *adaptedTool) Name() string        { return a.hex.Name() }
func (a *adaptedTool) Description() string { return a.hex.Description() }

func (a *adaptedTool) InputSchema() map[string]any {
    return convertToJSONSchema(a.hex.Parameters())
}

func (a *adaptedTool) Execute(ctx context.Context, input map[string]any) (map[string]any, error) {
    result, err := a.hex.Execute(ctx, input)
    if err != nil {
        return nil, err
    }
    return map[string]any{
        "content": result.Content,
        "success": result.Success,
    }, nil
}

func AdaptTool(t tools.Tool) tool.Tool {
    return &adaptedTool{hex: t}
}
```

**`hex/internal/adapter/bootstrap.go`** - Agent bootstrap

```go
func NewRootAgent(apiKey string, model string) (*agent.Agent, error) {
    llmClient := anthropic.New(apiKey, model)

    registry := tool.NewRegistry()
    for _, hexTool := range tools.AllTools() {
        registry.Register(AdaptTool(hexTool))
    }

    return agent.New(agent.Config{
        Name:         "hex-root",
        Registry:     registry,
        LLMClient:    llmClient,
        SystemPrompt: loadSystemPrompt(),
    }), nil
}

func NewSubagent(apiKey string, model string) (*agent.Agent, error) {
    llmClient := anthropic.New(apiKey, model)

    registry := tool.NewRegistry()
    for _, hexTool := range tools.AllTools() {
        registry.Register(AdaptTool(hexTool))
    }

    allowed := parseCSV(os.Getenv("HEX_ALLOWED_TOOLS"))
    denied := parseCSV(os.Getenv("HEX_DENIED_TOOLS"))

    return agent.New(agent.Config{
        Name:         os.Getenv("HEX_AGENT_ID"),
        Registry:     registry,
        LLMClient:    llmClient,
        SystemPrompt: loadSubagentPrompt(),
        AllowedTools: allowed,
        DeniedTools:  denied,
    }), nil
}
```

### Modified in hex

**`hex/cmd/hex/main.go`** - Entry point

```go
func main() {
    var ag *agent.Agent
    if os.Getenv("HEX_SUBAGENT_TYPE") != "" {
        ag = adapter.NewSubagent(apiKey, model)
    } else {
        ag = adapter.NewRootAgent(apiKey, model)
    }
    ag.Run(ctx, userPrompt)
}
```

**`hex/internal/tools/task_tool.go`** - Add tool restriction env vars

```go
config := subagents.DefaultConfig(req.Type)

cmd.Env = append(cmd.Env,
    fmt.Sprintf("HEX_SUBAGENT_TYPE=%s", req.Type),
    fmt.Sprintf("HEX_AGENT_ID=%s", agentID),
    fmt.Sprintf("HEX_PARENT_AGENT_ID=%s", parentAgentID),
    fmt.Sprintf("HEX_ALLOWED_TOOLS=%s", strings.Join(config.AllowedTools, ",")),
    fmt.Sprintf("HEX_DENIED_TOOLS=%s", strings.Join(config.DeniedTools, ",")),
)
```

### Removed from hex

- `internal/orchestrator/orchestrator.go` - replaced by mux
- `internal/orchestrator/state_machine.go` - mux handles state
- `internal/orchestrator/state_history.go` - mux handles this
- `internal/core/client.go` - replaced by mux Anthropic client
- `internal/core/streaming.go` - replaced by mux streaming

## Tool Restrictions by Subagent Type

| Type | AllowedTools | DeniedTools |
|------|-------------|-------------|
| Explore | Read, Grep, Glob, Bash | - |
| Plan | Read, Grep, Glob | - |
| CodeReviewer | Read, Grep, Glob | - |
| GeneralPurpose | (all) | - |

These restrictions are now **enforced** via mux's `FilteredRegistry`, not just documented.

## Adding New LLM Providers

The `llm.Client` interface enables multiple providers:

```go
var llmClient llm.Client
switch provider {
case "anthropic":
    llmClient = llm.NewAnthropicClient(apiKey, model)
case "openai":
    llmClient = llm.NewOpenAIClient(apiKey, model)
}

agent.New(agent.Config{
    LLMClient: llmClient,
})
```

Each provider implements `CreateMessage()` and `CreateMessageStream()`.

## Migration Order

1. Add mux as dependency
2. Implement Anthropic client in mux
3. Create adapter package in hex
4. Modify main.go to use mux agent
5. Update Task tool to pass tool restrictions
6. Test root agent works
7. Test subagent spawning works
8. Remove old orchestrator code
9. Remove old core/client code

## Estimates

| Location | Component | Lines (est.) |
|----------|-----------|--------------|
| mux | `llm/anthropic.go` | ~200 |
| hex | `internal/adapter/tool.go` | ~80 |
| hex | `internal/adapter/bootstrap.go` | ~100 |
| hex | `cmd/hex/main.go` changes | ~50 |
| hex | `internal/tools/task_tool.go` changes | ~20 |

Code removed: ~1500 lines (orchestrator, state machine, core client)

Net result: hex becomes a CLI wrapper around mux with hex-specific tools.
