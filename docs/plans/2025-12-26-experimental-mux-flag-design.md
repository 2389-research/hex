# Experimental Mux Flag Design

**Date:** 2025-12-26
**Status:** Approved
**Author:** Claude + Doctor Biz

## Overview

Add an `--experimental-mux` flag to hex that toggles between the existing custom agent orchestrator and the new mux-based agent framework. This enables A/B comparison of both approaches.

## Flag & Detection

### New Flag

```go
// In root.go init()
var experimentalMux bool
rootCmd.PersistentFlags().BoolVar(&experimentalMux, "experimental-mux", false,
    "Use experimental mux agent framework instead of built-in orchestrator")
```

### Detection Logic

Both `runPrintMode()` and `runInteractive()` check the flag early and branch:

```go
if experimentalMux {
    return runWithMuxAgent(prompt, ...)
}
// existing code continues
```

### Environment Propagation

When mux mode spawns subagents (via Task tool), propagate the flag so child processes also use mux:

```go
if experimentalMux {
    os.Setenv("HEX_EXPERIMENTAL_MUX", "1")
}
```

The flag also respects the env var, so subagents automatically inherit:

```go
experimentalMux = experimentalMux || os.Getenv("HEX_EXPERIMENTAL_MUX") == "1"
```

## Mux Agent Wiring

### New File: `cmd/hex/mux_runner.go`

```go
func runWithMuxAgent(prompt string, cfg *core.Config, registry *tools.Registry) error {
    // Get all hex tools as slice
    hexTools := registry.All()

    // Create mux agent via adapter
    agentCfg := adapter.Config{
        APIKey:       cfg.ProviderConfigs["anthropic"].APIKey,
        Model:        model,
        SystemPrompt: systemPrompt,
        HexTools:     hexTools,
    }

    var agent *agent.Agent
    if adapter.IsSubagent() {
        agent = adapter.NewSubagent(agentCfg)
    } else {
        agent = adapter.NewRootAgent(agentCfg)
    }

    // Run the agent
    result, err := agent.Run(context.Background(), prompt)
    if err != nil {
        return err
    }

    fmt.Print(result.Response)
    return nil
}
```

### Key Points

- Reuses existing tool registry - tools don't need changes
- Adapter wraps hex tools for mux's interface
- Same tool filtering works via `HEX_ALLOWED_TOOLS` / `HEX_DENIED_TOOLS` env vars
- Subagent detection happens automatically

## TUI Integration

### Streaming Support

Mux agents support streaming via callbacks. Wire this to the existing TUI message system:

```go
func runInteractiveWithMux(uiModel *ui.Model, agent *agent.Agent) error {
    // Mux provides streaming via StreamHandler
    handler := &tuiStreamHandler{model: uiModel}

    agent.SetStreamHandler(handler)

    // The TUI sends prompts via a channel, agent processes them
    for prompt := range uiModel.PromptChannel() {
        go func(p string) {
            result, err := agent.Run(ctx, p)
            if err != nil {
                uiModel.SendError(err)
                return
            }
            uiModel.SendComplete(result)
        }(prompt)
    }
}
```

### Tool Approval Flow

Mux has a `ToolApprovalFunc` hook that wires to the TUI's existing approval overlay:

```go
agent.SetToolApprovalFunc(func(toolName string, params map[string]any) bool {
    // Send to TUI, block until user responds
    approved := uiModel.RequestToolApproval(toolName, params)
    return approved
})
```

This reuses the existing `overlay_tool_approval.go` - no TUI changes needed.

## Testing & Comparison

### How to Compare Agents

Run the same prompts with and without the flag:

```bash
# Current orchestrator
hex -p "list files in current directory"

# Mux agent
hex --experimental-mux -p "list files in current directory"
```

### Comparison Metrics

Both paths already log to the event store:
- Tool calls made
- Token usage (via cost tracker)
- Latency
- Errors

### Test Coverage

```go
// mux_runner_test.go
func TestMuxAgentEquivalence(t *testing.T) {
    prompt := "read the file go.mod"

    // Run with current orchestrator
    result1 := runPrintMode(prompt)

    // Run with mux
    experimentalMux = true
    result2 := runPrintMode(prompt)

    // Both should successfully read the file
    assert.Contains(t, result1, "module github.com/2389-research/hex")
    assert.Contains(t, result2, "module github.com/2389-research/hex")
}
```

### Logging

When mux mode is active, log it clearly:

```go
logging.InfoWith("Using experimental mux agent", "model", model)
```

## Files to Create/Modify

| File | Change |
|------|--------|
| `cmd/hex/root.go` | Add `--experimental-mux` flag, env var check |
| `cmd/hex/mux_runner.go` | **NEW** - `runWithMuxAgent()`, `runInteractiveWithMux()` |
| `cmd/hex/mux_runner_test.go` | **NEW** - equivalence tests |
| `internal/adapter/bootstrap.go` | Add streaming handler support (if needed) |

## Implementation Order

1. Add flag to `root.go`
2. Create `mux_runner.go` with print mode support
3. Test print mode works
4. Add TUI streaming integration
5. Add equivalence tests
