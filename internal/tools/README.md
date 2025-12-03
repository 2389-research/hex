# Tool System Architecture

This package implements the tool execution system for Hex, providing a clean abstraction for executing tools (Read, Write, Bash, etc.) with permission management and API integration.

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                        Tool System                          │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌──────────────┐      ┌──────────────┐                    │
│  │   Registry   │◄─────┤   Executor   │                    │
│  │              │      │              │                    │
│  │  - Register  │      │  - Execute   │                    │
│  │  - Get       │      │  - Approval  │                    │
│  │  - List      │      └──────┬───────┘                    │
│  └──────┬───────┘             │                            │
│         │                     │                            │
│         │ manages             │ uses                       │
│         │                     │                            │
│         ▼                     ▼                            │
│  ┌─────────────────────────────────────┐                   │
│  │          Tool Interface             │                   │
│  ├─────────────────────────────────────┤                   │
│  │  - Name() string                    │                   │
│  │  - Description() string             │                   │
│  │  - RequiresApproval(params) bool    │                   │
│  │  - Execute(ctx, params) Result      │                   │
│  └─────────────────────────────────────┘                   │
│                      ▲                                      │
│                      │ implements                          │
│         ┌────────────┴────────────┬───────────┐            │
│         │                         │           │            │
│  ┌──────┴──────┐          ┌──────┴──────┐   ┌─┴─────┐     │
│  │  ReadTool   │          │ WriteTool   │   │ Bash  │     │
│  │             │          │             │   │ Tool  │     │
│  │ (Phase 2-9) │          │ (Phase 2-10)│   │(2-11) │     │
│  └─────────────┘          └─────────────┘   └───────┘     │
│                                                             │
└─────────────────────────────────────────────────────────────┘
                            │
                            │ converts to/from
                            ▼
                ┌────────────────────────┐
                │   API Integration      │
                ├────────────────────────┤
                │  ToolUse    (request)  │
                │  ToolResult (response) │
                └────────────────────────┘
```

## Core Components

### 1. Tool Interface

The `Tool` interface defines the contract that all tools must implement:

```go
type Tool interface {
    Name() string
    Description() string
    RequiresApproval(params map[string]interface{}) bool
    Execute(ctx context.Context, params map[string]interface{}) (*Result, error)
}
```

**Key Design Decisions:**

- **Context-aware**: All tools receive a context for cancellation and timeout support
- **Dynamic approval**: Tools decide at runtime if they need approval based on parameters
- **Generic parameters**: Uses `map[string]interface{}` for flexibility with different tool types
- **Structured results**: Returns `Result` with success/failure, output, error, and metadata

### 2. Registry

Thread-safe storage for available tools. Provides:

- **Registration**: Add tools to the registry
- **Retrieval**: Get tools by name
- **Listing**: Get all registered tool names (sorted alphabetically)
- **Thread Safety**: Uses `sync.RWMutex` for concurrent access

**Design Pattern**: Singleton-style registry that persists for the application lifetime.

### 3. Executor

Manages tool execution with permission checking:

```go
type Executor struct {
    registry     *Registry
    approvalFunc ApprovalFunc
}
```

**Execution Flow:**

1. Retrieve tool from registry
2. Check if approval is needed (via `RequiresApproval`)
3. If needed, call `approvalFunc` and block if denied
4. Execute tool with context and parameters
5. Return result or error

**Permission System:**

- `ApprovalFunc`: `func(toolName string, params map[string]interface{}) bool`
- Called only when `RequiresApproval(params)` returns `true`
- If `nil`, no approval checks are performed (useful for testing)
- If returns `false`, tool is not executed (permission denied)

### 4. Result

Represents tool execution outcome:

```go
type Result struct {
    ToolName string                 // Tool that was executed
    Success  bool                   // Did it succeed?
    Output   string                 // Standard output/result
    Error    string                 // Error message if failed
    Metadata map[string]interface{} // Additional metadata
}
```

**Design Philosophy:**

- Clear success/failure indication
- Separate `Output` and `Error` fields (only one populated)
- Extensible `Metadata` for tool-specific information (exit codes, file paths, etc.)

### 5. API Integration Types

Maps between internal execution and Anthropic API format:

**ToolUse** (API → Internal):
```go
type ToolUse struct {
    Type  string                 `json:"type"`  // "tool_use"
    ID    string                 `json:"id"`    // Unique ID
    Name  string                 `json:"name"`  // Tool name
    Input map[string]interface{} `json:"input"` // Parameters
}
```

**ToolResult** (Internal → API):
```go
type ToolResult struct {
    Type      string `json:"type"`        // "tool_result"
    ToolUseID string `json:"tool_use_id"` // Original ID
    Content   string `json:"content"`     // Output or error
    IsError   bool   `json:"is_error"`    // Error flag
}
```

**Conversion**: `ResultToToolResult(result *Result, toolUseID string) ToolResult`

## Usage Examples

### Basic Tool Registration and Execution

```go
// Create registry
registry := tools.NewRegistry()

// Register a tool
tool := &MyTool{}
registry.Register(tool)

// Create executor with approval
executor := tools.NewExecutor(registry, func(name string, params map[string]interface{}) bool {
    // Ask user for approval
    return askUser(name, params)
})

// Execute tool
result, err := executor.Execute(context.Background(), "my_tool", map[string]interface{}{
    "param1": "value1",
})
```

### Implementing a Custom Tool

```go
type MyTool struct{}

func (t *MyTool) Name() string {
    return "my_tool"
}

func (t *MyTool) Description() string {
    return "Does something useful"
}

func (t *MyTool) RequiresApproval(params map[string]interface{}) bool {
    // Check if this operation is dangerous based on params
    path := params["path"].(string)
    return strings.HasPrefix(path, "/etc")
}

func (t *MyTool) Execute(ctx context.Context, params map[string]interface{}) (*tools.Result, error) {
    // Do the work
    output, err := doWork(params)
    if err != nil {
        return &tools.Result{
            ToolName: "my_tool",
            Success:  false,
            Error:    err.Error(),
        }, nil
    }

    return &tools.Result{
        ToolName: "my_tool",
        Success:  true,
        Output:   output,
        Metadata: map[string]interface{}{
            "path": params["path"],
        },
    }, nil
}
```

### Handling API Integration

```go
// Receive from API
var toolUse tools.ToolUse
json.Unmarshal(apiData, &toolUse)

// Execute
result, err := executor.Execute(ctx, toolUse.Name, toolUse.Input)
if err != nil {
    // Handle error
}

// Convert back to API format
toolResult := tools.ResultToToolResult(result, toolUse.ID)
apiResponse, _ := json.Marshal(toolResult)
```

## Testing

### Unit Testing

Each component has comprehensive unit tests:

- `tool_test.go`: Tool interface and mock implementations
- `registry_test.go`: Registration, retrieval, thread safety
- `executor_test.go`: Execution flow, approval, error handling
- `result_test.go`: Result structures and metadata
- `types_test.go`: API integration types and JSON marshaling

**Coverage**: 100% statement coverage

### Mock Tool

The `MockTool` type provides configurable behavior for testing:

```go
mock := &tools.MockTool{
    NameValue: "test_tool",
    RequiresApprovalValue: true,
    ExecuteFunc: func(ctx context.Context, params map[string]interface{}) (*tools.Result, error) {
        // Custom behavior
        return &tools.Result{Success: true, Output: "test"}, nil
    },
}
```

## Thread Safety

- **Registry**: Uses `sync.RWMutex` for concurrent access
  - Multiple readers can access simultaneously
  - Writers get exclusive access
  - Tested with concurrent registration and retrieval

- **Executor**: Stateless, safe for concurrent use
  - Each execution is independent
  - Context propagation for cancellation

## Error Handling

**Two levels of errors:**

1. **Execution errors** (returned as `error`):
   - Tool not found in registry
   - Catastrophic failures during tool execution
   - Context cancellation

2. **Tool errors** (in `Result.Error`):
   - Expected failures (file not found, permission denied, etc.)
   - User-facing error messages
   - Tool still executed, but operation failed

**Philosophy**: Distinguish between "the tool ran but failed" vs "we couldn't run the tool."

## Next Steps (Phase 2 Tasks 9-11)

This architecture provides the foundation for:

- **Task 9**: Read Tool - File reading with safety checks
- **Task 10**: Write Tool - File writing with confirmation
- **Task 11**: Bash Tool - Sandboxed command execution

Each will implement the `Tool` interface and integrate with this system.

## Design Principles

1. **Separation of Concerns**: Registry, execution, and approval are separate
2. **Testability**: Clean interfaces, mockable components
3. **Extensibility**: Easy to add new tool types
4. **Safety**: Permission system prevents dangerous operations
5. **Thread Safety**: Concurrent tool execution support
6. **API Integration**: Clean mapping to/from Anthropic API format

## Files

- `tool.go`: Tool interface definition
- `registry.go`: Tool registry implementation
- `executor.go`: Tool executor with approval
- `result.go`: Result structures
- `types.go`: API integration types
- `mock_tool.go`: Testing utilities
- `*_test.go`: Comprehensive test coverage
- `example_test.go`: Usage examples
- `README.md`: This documentation
