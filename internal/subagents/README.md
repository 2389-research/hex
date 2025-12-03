# Subagent Framework

Complete implementation of Phase 5 from the Alignment Roadmap: isolated subagent execution with parallel dispatch.

## Overview

The subagent framework enables Hex to spawn isolated Claude instances for specialized tasks. Each subagent runs with:
- **Isolated context**: Separate conversation history and working memory
- **Specialized configuration**: Type-specific tools, temperature, and token limits
- **Parallel execution**: Multiple subagents can run concurrently
- **Hook integration**: SubagentStop event fires on completion

## Subagent Types

### 1. `general-purpose`
- **Purpose**: General tasks (current default behavior)
- **Tools**: All tools available
- **Temperature**: 1.0
- **Max Tokens**: 4096

### 2. `Explore`
- **Purpose**: Fast codebase exploration and research
- **Tools**: Read, Grep, Glob, Bash (read-only)
- **Temperature**: 0.7
- **Max Tokens**: 8192
- **Use case**: "Find all authentication code"

### 3. `Plan`
- **Purpose**: Design and planning work
- **Tools**: Read, Grep, Glob
- **Temperature**: 0.6
- **Max Tokens**: 6144
- **Use case**: "Create implementation plan for feature X"

### 4. `code-reviewer`
- **Purpose**: Code review and quality checks
- **Tools**: Read, Grep, Glob
- **Temperature**: 0.3 (consistent, thorough)
- **Max Tokens**: 6144
- **Use case**: "Review PR #123 for security issues"

## Architecture

```
internal/subagents/
├── types.go          # Type definitions and configuration
├── context.go        # Context isolation (separate conversation history)
├── executor.go       # Isolated execution engine (spawns hex processes)
├── dispatcher.go     # Parallel dispatch coordination
└── subagents_test.go # Comprehensive tests (81.3% coverage)
```

## Usage Examples

### Basic Execution

```go
executor := subagents.NewExecutor()

req := &subagents.ExecutionRequest{
    Type:        subagents.TypeExplore,
    Prompt:      "Find all authentication code in src/",
    Description: "Explore auth system",
}

result, err := executor.Execute(ctx, req)
if err != nil {
    return err
}

fmt.Println(result.Output)
```

### Parallel Dispatch

```go
dispatcher := subagents.NewDispatcher(executor)

requests := []*subagents.DispatchRequest{
    {
        ID: "bug1",
        Request: &subagents.ExecutionRequest{
            Type:        subagents.TypeGeneralPurpose,
            Prompt:      "Fix authentication bug",
            Description: "Fix auth bug",
        },
    },
    {
        ID: "bug2",
        Request: &subagents.ExecutionRequest{
            Type:        subagents.TypeGeneralPurpose,
            Prompt:      "Fix payment bug",
            Description: "Fix payment bug",
        },
    },
}

results := dispatcher.DispatchParallel(ctx, requests)

// Process results
for _, r := range results {
    fmt.Printf("Task %s: %v\n", r.ID, r.Result.Success)
}
```

### With Hooks

```go
// Hook engine must implement subagents.HookEngine interface
result, err := executor.ExecuteWithHooks(ctx, req, hookEngine)

// SubagentStop hook fires with:
// - taskDescription
// - subagentType
// - responseLength
// - tokensUsed
// - success
// - executionTime
```

### Using the Enhanced Task Tool

```go
// Legacy mode (backward compatible)
tool := tools.NewTaskTool()

// New framework mode
tool := tools.NewTaskToolWithFramework()

// Execute via tool
result, err := tool.Execute(ctx, map[string]interface{}{
    "prompt":        "Find all auth code",
    "description":   "Explore authentication",
    "subagent_type": "Explore",  // Must be valid type
    "model":         "claude-sonnet-4-5-20250929",  // Optional
})
```

## Context Isolation

Each subagent gets its own `IsolatedContext`:

```go
// Created automatically by executor
ctx := subagents.NewIsolatedContext("parent-id", subagents.TypeExplore)

// Isolated conversation history
ctx.AddMessage("user", "Hello")
messages := ctx.GetMessages()  // Returns copy, prevents pollution

// Working memory
ctx.SetMemory("key", "value")
value, ok := ctx.GetMemory("key")
```

The `ContextManager` handles lifecycle:
- Creates contexts on execution
- Cleans up after completion
- Supports cleanup of old contexts

## Hook Integration

The SubagentStop hook fires when a subagent completes:

```json
{
  "hooks": {
    "SubagentStop": {
      "command": "echo 'Subagent ${CLAUDE_SUBAGENT_TYPE} completed in ${CLAUDE_EXECUTION_TIME}s'"
    }
  }
}
```

Environment variables available:
- `CLAUDE_TASK_DESCRIPTION`
- `CLAUDE_SUBAGENT_TYPE`
- `CLAUDE_RESPONSE_LENGTH`
- `CLAUDE_TOKENS_USED`
- `CLAUDE_SUBAGENT_SUCCESS`
- `CLAUDE_EXECUTION_TIME`
- `CLAUDE_IS_SUBAGENT=true`

## Advanced Patterns

### Batch Processing

```go
// Process 100 tasks in batches of 10
results := dispatcher.DispatchBatch(ctx, requests, 10)
```

### Wait for First Success

```go
// Race multiple approaches, use first success
result, err := dispatcher.WaitForAny(ctx, requests)
```

### With Aggregation

```go
// Aggregate results from multiple subagents
aggregated, err := dispatcher.DispatchWithAggregation(
    ctx,
    requests,
    func(results []*subagents.Result) (interface{}, error) {
        // Combine results
        return combinedAnalysis, nil
    },
)
```

### Statistics

```go
results := dispatcher.DispatchParallel(ctx, requests)
stats := subagents.CalculateStatistics(results)

fmt.Printf("Total: %d, Success: %d, Failed: %d\n",
    stats.Total, stats.Successful, stats.Failed)
```

## Testing

Comprehensive test suite with 81.3% coverage:

```bash
go test ./internal/subagents/... -v
go test ./internal/subagents/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

Tests cover:
- All four subagent types
- Configuration defaults
- Context isolation and thread safety
- Parallel dispatch
- Hook integration
- Error handling
- Validation

## Implementation Details

### Context Isolation

- Each subagent has separate `ConversationHistory`
- Working memory isolated via mutex-protected maps
- Context cleanup after execution prevents memory leaks

### Process Spawning

- Spawns `hex --print` subprocess per subagent
- Inherits environment (API keys, config)
- Sets `HEX_SUBAGENT_TYPE` and `HEX_SUBAGENT_CONTEXT_ID`
- Captures stdout/stderr
- Respects timeouts and cancellation

### Parallel Coordination

- Semaphore limits concurrency (default: 10)
- Goroutines for parallel execution
- WaitGroup for synchronization
- Context cancellation propagated

## Future Enhancements

Potential improvements:
1. **Token tracking**: Parse API response to get actual token usage
2. **Resume support**: Allow subagents to resume from checkpoints
3. **Streaming**: Add streaming support for long-running subagents
4. **Resource limits**: CPU/memory constraints per subagent
5. **Retry logic**: Automatic retry on transient failures
6. **Result caching**: Cache results for identical requests

## Integration Points

### Task Tool
- `/Users/harper/Public/src/2389/hex/internal/tools/task_tool.go`
- Enhanced to use framework via `NewTaskToolWithFramework()`
- Backward compatible with legacy mode

### Hooks Engine
- `/Users/harper/Public/src/2389/hex/internal/hooks/events.go`
- SubagentStop event defined
- `/Users/harper/Public/src/2389/hex/internal/hooks/engine.go`
- `FireSubagentStop()` method

## Performance Characteristics

- **Context creation**: ~1μs
- **Subprocess spawn**: ~500ms (one-time build) + API call time
- **Parallel dispatch**: ~10 concurrent by default
- **Memory**: Isolated contexts prevent memory leaks
- **Cleanup**: Automatic via defer in executor

## Backward Compatibility

The Task tool maintains full backward compatibility:
- `NewTaskTool()` uses legacy subprocess implementation
- `NewTaskToolWithFramework()` opts into new framework
- Same API surface for both modes
- Existing tests pass unchanged
