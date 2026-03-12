# Task Tool - Sub-Agent System

## Overview

The Task tool enables Hex to spawn sub-agent processes to handle complex, multi-step tasks autonomously. This is similar to Claude Code's task delegation system, where a parent agent can launch child agents to work on specific problems.

## How It Works

The Task tool supports three execution modes:

1. **Legacy subprocess mode**: Spawns a new `hex --print` process
2. **Framework mode**: Uses `subagents.Executor` for structured agent execution
3. **Mux mode**: Runs agents in-process using the mux library

The mode is selected automatically based on configuration. In all modes:
- The prompt and configuration are passed to the sub-agent
- API keys and environment variables are inherited from parent
- Output is captured and returned as the tool result

## Architecture

```
Parent Hex Process
    |
    |- Tool: Task
         |
         |- exec.Command("hex", "--print", prompt)
              |
              |- Child Hex Process
                   |
                   |- Executes task
                   |- Returns output
```

## Tool Specification

**Name**: `task`

**Description**: Launch a sub-agent to handle complex, multi-step tasks autonomously

**Always Requires Approval**: Yes (spawns processes, uses API, costs money)

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `prompt` | string | Yes | The task for the agent to perform |
| `description` | string | Yes | Short 3-5 word description of the task |
| `subagent_type` | string | Yes | Type of agent (e.g., "general-purpose") |
| `model` | string | No | Model to use (inherits parent's model if not specified) |
| `resume` | string | No | Agent ID to resume from a previous conversation |
| `timeout` | number | No | Timeout in seconds (default: 300, max: 1800) |

## Example Usage

### Basic Task

```json
{
  "name": "task",
  "input": {
    "prompt": "Analyze the codebase and find all TODOs",
    "description": "Find TODOs",
    "subagent_type": "general-purpose"
  }
}
```

### Task with Specific Model

```json
{
  "name": "task",
  "input": {
    "prompt": "Write comprehensive tests for the user authentication module",
    "description": "Write auth tests",
    "subagent_type": "general-purpose",
    "model": "claude-sonnet-4-5-20250929"
  }
}
```

### Resuming a Previous Task

```json
{
  "name": "task",
  "input": {
    "prompt": "Continue the previous analysis",
    "description": "Resume analysis",
    "subagent_type": "general-purpose",
    "resume": "conv-1234567890"
  }
}
```

## Implementation Details

### Command Execution

The tool builds a command like:

```bash
hex --print "Your prompt here"
```

With optional flags:

```bash
hex --model claude-sonnet-4 --print "Your prompt here"
```

Or:

```bash
hex --resume conv-123 --print "Your prompt here"
```

### Environment Inheritance

The subprocess inherits:
- `ANTHROPIC_API_KEY` - Required for API access
- `PATH` - For finding executables
- All other environment variables from parent process
- Current working directory

### Binary Discovery

The tool tries to find the `hex` binary in this order:

1. If `HexbinPath` is set, use that
2. Look for `hex` in PATH using `exec.LookPath`
3. Try to build from source by:
   - Finding `go.mod` in current directory or up to 5 levels up
   - Running `go build -o /tmp/hex-<timestamp> ./cmd/hex`
   - Using the temporary binary

### Timeout Handling

- **Default**: 5 minutes (300 seconds)
- **Maximum**: 30 minutes (1800 seconds)
- **Behavior**: If timeout is exceeded, the subprocess is killed and an error is returned

### Output Capture

- Both `stdout` and `stderr` are captured and combined
- The `--print` flag ensures non-interactive output
- Output is returned in the `Output` field of the Result

## Result Structure

### Success

```json
{
  "ToolName": "task",
  "Success": true,
  "Output": "The sub-agent's response text...",
  "Error": "",
  "Metadata": {
    "exit_code": 0,
    "duration": 12.345,
    "prompt": "Your prompt",
    "description": "Short description",
    "subagent_type": "general-purpose",
    "model": "claude-sonnet-4-5-20250929"
  }
}
```

### Failure

```json
{
  "ToolName": "task",
  "Success": false,
  "Output": "(combined stdout/stderr)",
  "Error": "sub-agent exited with code 1",
  "Metadata": {
    "exit_code": 1,
    "duration": 5.678,
    "prompt": "Your prompt",
    "description": "Short description",
    "subagent_type": "general-purpose"
  }
}
```

### Timeout

```json
{
  "ToolName": "task",
  "Success": false,
  "Output": "",
  "Error": "task timed out after 5m0s",
  "Metadata": {
    "timeout": 300,
    "duration": 300.123,
    "prompt": "Your prompt",
    "description": "Short description",
    "subagent_type": "general-purpose"
  }
}
```

## Limitations vs Full MCP

This implementation is simplified compared to Claude Code's full MCP (Model Context Protocol) support:

### What We Have

- ✅ Subprocess spawning
- ✅ Environment inheritance
- ✅ Output capture
- ✅ Timeout handling
- ✅ Context cancellation
- ✅ Basic parameter passing

### What We're Missing

- ❌ Bidirectional communication during execution
- ❌ Streaming updates from sub-agent
- ❌ Tool use by sub-agent (sub-agent can't call tools)
- ❌ Conversation state sharing beyond initial prompt
- ❌ Resource limits (memory, CPU)
- ❌ Sub-agent termination signals
- ❌ Progress reporting

### Why This Works Anyway

Despite the limitations, this approach works well because:

1. **Print Mode**: The `--print` flag makes Hex execute non-interactively and return a single response
2. **Simple Tasks**: Many tasks don't need bidirectional communication
3. **Environment Sharing**: API keys and config are inherited automatically
4. **Clean Output**: Combined stdout/stderr provides complete output

## Future Enhancements

Potential improvements:

1. **Streaming**: Capture and stream subprocess output in real-time
2. **Progress**: Parse and report progress indicators
3. **Tool Support**: Enable sub-agents to use tools (requires MCP protocol)
4. **Resource Limits**: Add memory and CPU constraints
5. **Interactive**: Support interactive mode for complex workflows
6. **State Sharing**: Pass conversation history to sub-agent
7. **Cancellation**: Better handling of graceful shutdown

## Testing

The tool includes comprehensive tests covering:

- Tool metadata (name, description)
- Approval requirements
- Parameter validation (all required parameters)
- Type checking (all parameters)
- Subprocess execution
- Output capture
- Error handling
- Timeout behavior
- Context cancellation
- Metadata population

Run tests with:

```bash
go test ./internal/tools -run TestTaskTool -v
```

Note: Some tests require `ANTHROPIC_API_KEY` to be set and will be skipped otherwise.

## Security Considerations

1. **API Key**: Sub-agent inherits API key from environment
2. **Working Directory**: Sub-agent uses same working directory as parent
3. **Environment**: All environment variables are inherited
4. **Approval Required**: Every task execution requires user approval
5. **Timeout**: Maximum 30 minute execution time prevents runaway processes
6. **Exit Codes**: Non-zero exit codes are treated as failures

## Example Scenarios

### Scenario 1: Code Analysis

```go
result, err := taskTool.Execute(ctx, map[string]interface{}{
    "prompt": "Analyze all Go files in internal/tools and list potential bugs",
    "description": "Analyze tools code",
    "subagent_type": "general-purpose",
})
```

### Scenario 2: Test Generation

```go
result, err := taskTool.Execute(ctx, map[string]interface{}{
    "prompt": "Generate comprehensive tests for internal/storage/messages.go",
    "description": "Generate tests",
    "subagent_type": "general-purpose",
    "model": "claude-sonnet-4-5-20250929",
})
```

### Scenario 3: Documentation

```go
result, err := taskTool.Execute(ctx, map[string]interface{}{
    "prompt": "Create API documentation for all exported functions in internal/core",
    "description": "Create API docs",
    "subagent_type": "general-purpose",
    "timeout": 600, // 10 minutes
})
```

## Debugging

Enable verbose output to see subprocess execution:

```bash
# Run with debug output
go test ./internal/tools -run TestTaskTool_Execute_SimpleTask -v

# Check subprocess errors
result, _ := taskTool.Execute(ctx, params)
if !result.Success {
    fmt.Printf("Error: %s\n", result.Error)
    fmt.Printf("Output: %s\n", result.Output)
    fmt.Printf("Exit Code: %d\n", result.Metadata["exit_code"])
}
```

## Conclusion

The Task tool provides a practical sub-agent system for Hex that works within the constraints of subprocess execution. While not as sophisticated as full MCP support, it enables complex task delegation and autonomous operation for many common scenarios.
