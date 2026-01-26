# Phase 2: Tool Support Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Enable HexAgent to handle tool calls, approvals, and results so tux can display and manage tools.

**Architecture:** Extend HexAgent to parse tool_use chunks from the API, emit EventToolCall/EventApproval/EventToolResult events, integrate with hex's tool executor, and send results back to continue the conversation.

**Tech Stack:** Go, hex's existing tool system (internal/tools), tux event types

---

## Background

### Current State (Phase 1)
- HexAgent streams text via `EventText` events
- Only handles `content_block_delta` (text) and `message_stop` chunks
- No tool support

### Target State (Phase 2)
- HexAgent handles tool_use content blocks
- Emits `EventToolCall` when tool is parsed
- Emits `EventApproval` and waits for user decision
- Executes approved tools via hex's Executor
- Emits `EventToolResult` with output
- Sends tool results to API and resumes streaming

### Key Data Flow
```
API chunks → HexAgent → tux Events → tux UI
                ↓
         tool execution
                ↓
         tool results → API (continue conversation)
```

---

## Task 1: Add Tool Executor to HexAgent

**Files:**
- Modify: `internal/tui/agent.go`

**Step 1: Add imports and executor field**

Add to imports:
```go
"github.com/2389-research/hex/internal/tools"
```

Add to HexAgent struct:
```go
// Tool execution
executor *tools.Executor
```

**Step 2: Update constructor**

```go
// NewHexAgent creates a new HexAgent with the given API client and tool executor.
func NewHexAgent(client *core.Client, model string, systemPrompt string, executor *tools.Executor) *HexAgent {
	if client == nil {
		panic("client cannot be nil")
	}
	return &HexAgent{
		client:       client,
		model:        model,
		systemPrompt: systemPrompt,
		messages:     make([]core.Message, 0),
		executor:     executor,
	}
}
```

**Step 3: Update call sites**

In `cmd/hex/tux.go`, update `runTuxMode` to pass executor:
```go
func runTuxMode(apiKey, model, systemPrompt string, executor *tools.Executor) error {
	client := core.NewClient(apiKey)
	agent := tui.NewHexAgent(client, model, systemPrompt, executor)
	// ...
}
```

In `cmd/hex/root.go`, create and pass executor:
```go
// Inside the useTux block, after API key validation:

// Create tool registry and executor
registry := tools.NewRegistry()
executor := tools.NewExecutor(registry, nil, nil, nil) // Simplified for now

return runTuxMode(providerCfg.APIKey, tuxModel, systemPrompt, executor)
```

**Step 4: Update test helper**

In `internal/tui/agent_test.go`, update `newTestAgent`:
```go
func newTestAgent(model, systemPrompt string) *HexAgent {
	return &HexAgent{
		client:       nil,
		model:        model,
		systemPrompt: systemPrompt,
		messages:     make([]core.Message, 0),
		executor:     nil, // Tests that don't execute tools
	}
}
```

**Step 5: Verify build**

```bash
go build ./...
```

**Step 6: Commit**

```bash
git add internal/tui/agent.go internal/tui/agent_test.go cmd/hex/tux.go cmd/hex/root.go
git commit -m "feat: add tool executor to HexAgent"
```

---

## Task 2: Add Tool Parsing State

**Files:**
- Modify: `internal/tui/agent.go`

**Step 1: Add tool parsing state to struct**

Add to HexAgent struct:
```go
// Tool parsing state (within a single Run)
assemblingTool   *core.ToolUse
toolInputJSONBuf strings.Builder
pendingTools     []*core.ToolUse
```

**Step 2: Add reset method**

```go
// resetToolState clears tool parsing state for a new run.
func (a *HexAgent) resetToolState() {
	a.assemblingTool = nil
	a.toolInputJSONBuf.Reset()
	a.pendingTools = nil
}
```

**Step 3: Verify build**

```bash
go build ./internal/tui/...
```

**Step 4: Commit**

```bash
git add internal/tui/agent.go
git commit -m "feat: add tool parsing state to HexAgent"
```

---

## Task 3: Handle content_block_start for Tools

**Files:**
- Modify: `internal/tui/agent.go`

**Step 1: Add handler method**

```go
// handleContentBlockStart processes a content_block_start chunk.
func (a *HexAgent) handleContentBlockStart(chunk *core.StreamChunk) {
	if chunk.ContentBlock == nil {
		return
	}

	if chunk.ContentBlock.Type == "tool_use" {
		a.assemblingTool = &core.ToolUse{
			Type:  "tool_use",
			ID:    chunk.ContentBlock.ID,
			Name:  chunk.ContentBlock.Name,
			Input: make(map[string]interface{}),
		}
		a.toolInputJSONBuf.Reset()

		// Emit tool call event (parameters will come in deltas)
		a.emit(tux.Event{
			Type:       tux.EventToolCall,
			ToolID:     chunk.ContentBlock.ID,
			ToolName:   chunk.ContentBlock.Name,
			ToolParams: nil, // Params not yet available
		})
	}
}
```

**Step 2: Verify build**

```bash
go build ./internal/tui/...
```

**Step 3: Commit**

```bash
git add internal/tui/agent.go
git commit -m "feat: handle content_block_start for tool_use"
```

---

## Task 4: Handle content_block_delta for Tool Params

**Files:**
- Modify: `internal/tui/agent.go`

**Step 1: Update chunk processing in Run()**

Replace the current chunk handling switch with expanded logic:

```go
// Handle different chunk types
switch chunk.Type {
case "content_block_start":
	a.handleContentBlockStart(chunk)

case "content_block_delta":
	if chunk.Delta != nil {
		switch chunk.Delta.Type {
		case "text_delta":
			responseText.WriteString(chunk.Delta.Text)
			a.emit(tux.Event{
				Type: tux.EventText,
				Text: chunk.Delta.Text,
			})
		case "input_json_delta":
			// Accumulate tool parameter JSON
			if a.assemblingTool != nil {
				a.toolInputJSONBuf.WriteString(chunk.Delta.PartialJSON)
			}
		}
	}

case "content_block_stop":
	a.handleContentBlockStop()

case "message_stop":
	// Add assistant response to history
	a.mu.Lock()
	a.messages = append(a.messages, core.Message{
		Role:    "assistant",
		Content: responseText.String(),
	})
	a.mu.Unlock()

	// Process any pending tools
	if len(a.pendingTools) > 0 {
		if err := a.processTools(ctx); err != nil {
			return err
		}
	} else {
		a.emit(tux.Event{Type: tux.EventComplete})
	}
}
```

**Step 2: Verify build**

```bash
go build ./internal/tui/...
```

**Step 3: Commit**

```bash
git add internal/tui/agent.go
git commit -m "feat: handle content_block_delta for tool params"
```

---

## Task 5: Handle content_block_stop

**Files:**
- Modify: `internal/tui/agent.go`

**Step 1: Add handler method**

```go
// handleContentBlockStop processes a content_block_stop chunk.
func (a *HexAgent) handleContentBlockStop() {
	if a.assemblingTool == nil {
		return
	}

	// Parse accumulated JSON into Input map
	jsonStr := a.toolInputJSONBuf.String()
	if jsonStr != "" {
		if err := json.Unmarshal([]byte(jsonStr), &a.assemblingTool.Input); err != nil {
			// Log error but continue - malformed params
			a.emit(tux.Event{
				Type:  tux.EventError,
				Error: fmt.Errorf("parse tool params: %w", err),
			})
		}
	}

	// Add to pending tools
	a.pendingTools = append(a.pendingTools, a.assemblingTool)
	a.assemblingTool = nil
	a.toolInputJSONBuf.Reset()
}
```

**Step 2: Add json import**

Add to imports:
```go
"encoding/json"
```

**Step 3: Verify build**

```bash
go build ./internal/tui/...
```

**Step 4: Commit**

```bash
git add internal/tui/agent.go
git commit -m "feat: handle content_block_stop to finalize tool"
```

---

## Task 6: Implement Tool Processing Loop

**Files:**
- Modify: `internal/tui/agent.go`

**Step 1: Add processTools method**

```go
// processTools handles pending tools: approval, execution, results.
func (a *HexAgent) processTools(ctx context.Context) error {
	var toolResults []core.ToolResult

	for _, tool := range a.pendingTools {
		// Check if tool needs approval
		needsApproval := true
		if a.executor != nil {
			// Could check executor's approval rules here
			// For now, always require approval
		}

		var approved bool
		if needsApproval {
			// Request approval via event
			decision, err := a.requestApproval(ctx, tool)
			if err != nil {
				return err
			}
			approved = (decision == tux.DecisionApprove || decision == tux.DecisionAlwaysAllow)
		} else {
			approved = true
		}

		if approved {
			// Execute tool
			result := a.executeTool(ctx, tool)
			toolResults = append(toolResults, result)

			// Emit result event
			a.emit(tux.Event{
				Type:       tux.EventToolResult,
				ToolID:     tool.ID,
				ToolName:   tool.Name,
				ToolOutput: result.Content,
				Success:    !result.IsError,
			})
		} else {
			// Tool denied
			toolResults = append(toolResults, core.ToolResult{
				Type:      "tool_result",
				ToolUseID: tool.ID,
				Content:   "Tool execution denied by user",
				IsError:   true,
			})

			a.emit(tux.Event{
				Type:       tux.EventToolResult,
				ToolID:     tool.ID,
				ToolName:   tool.Name,
				ToolOutput: "Tool execution denied by user",
				Success:    false,
			})
		}
	}

	// Clear pending tools
	a.pendingTools = nil

	// Continue conversation with tool results
	return a.continueWithToolResults(ctx, toolResults)
}
```

**Step 2: Verify build**

```bash
go build ./internal/tui/...
```

**Step 3: Commit**

```bash
git add internal/tui/agent.go
git commit -m "feat: add tool processing loop"
```

---

## Task 7: Implement Approval Request

**Files:**
- Modify: `internal/tui/agent.go`

**Step 1: Add requestApproval method**

```go
// requestApproval emits an approval event and waits for user decision.
func (a *HexAgent) requestApproval(ctx context.Context, tool *core.ToolUse) (tux.ApprovalDecision, error) {
	// Create response channel
	responseChan := make(chan tux.ApprovalDecision, 1)

	// Emit approval event with params now available
	a.emit(tux.Event{
		Type:       tux.EventApproval,
		ToolID:     tool.ID,
		ToolName:   tool.Name,
		ToolParams: tool.Input,
		Response:   responseChan,
	})

	// Wait for decision
	select {
	case decision := <-responseChan:
		return decision, nil
	case <-ctx.Done():
		return tux.DecisionDeny, ctx.Err()
	}
}
```

**Step 2: Verify build**

```bash
go build ./internal/tui/...
```

**Step 3: Commit**

```bash
git add internal/tui/agent.go
git commit -m "feat: add approval request with response channel"
```

---

## Task 8: Implement Tool Execution

**Files:**
- Modify: `internal/tui/agent.go`

**Step 1: Add executeTool method**

```go
// executeTool runs a tool and returns the result.
func (a *HexAgent) executeTool(ctx context.Context, tool *core.ToolUse) core.ToolResult {
	if a.executor == nil {
		return core.ToolResult{
			Type:      "tool_result",
			ToolUseID: tool.ID,
			Content:   "Tool executor not configured",
			IsError:   true,
		}
	}

	result, err := a.executor.Execute(ctx, tool.Name, tool.Input)
	if err != nil {
		return core.ToolResult{
			Type:      "tool_result",
			ToolUseID: tool.ID,
			Content:   fmt.Sprintf("Error: %v", err),
			IsError:   true,
		}
	}

	content := result.Output
	if !result.Success && result.Error != "" {
		content = fmt.Sprintf("Error: %s", result.Error)
	}

	return core.ToolResult{
		Type:      "tool_result",
		ToolUseID: tool.ID,
		Content:   content,
		IsError:   !result.Success,
	}
}
```

**Step 2: Verify build**

```bash
go build ./internal/tui/...
```

**Step 3: Commit**

```bash
git add internal/tui/agent.go
git commit -m "feat: add tool execution via executor"
```

---

## Task 9: Implement Continue with Tool Results

**Files:**
- Modify: `internal/tui/agent.go`

**Step 1: Add continueWithToolResults method**

```go
// continueWithToolResults sends tool results to API and resumes streaming.
func (a *HexAgent) continueWithToolResults(ctx context.Context, results []core.ToolResult) error {
	// Build tool result message
	var content []interface{}
	for _, r := range results {
		content = append(content, map[string]interface{}{
			"type":        "tool_result",
			"tool_use_id": r.ToolUseID,
			"content":     r.Content,
			"is_error":    r.IsError,
		})
	}

	// Add to message history
	a.mu.Lock()
	a.messages = append(a.messages, core.Message{
		Role:    "user",
		Content: content,
	})
	messages := make([]core.Message, len(a.messages))
	copy(messages, a.messages)
	a.mu.Unlock()

	// Build continuation request
	req := core.MessageRequest{
		Model:     a.model,
		Messages:  messages,
		MaxTokens: 8192,
		Stream:    true,
		System:    a.systemPrompt,
	}

	// Start new stream
	chunks, err := a.client.CreateMessageStream(ctx, req)
	if err != nil {
		a.emit(tux.Event{Type: tux.EventError, Error: err})
		return err
	}

	// Process continuation stream (recursive call to same logic)
	return a.processStream(ctx, chunks)
}
```

**Step 2: Refactor Run() to use processStream**

Extract the streaming loop into a separate method so it can be reused:

```go
// processStream handles the streaming response from the API.
func (a *HexAgent) processStream(ctx context.Context, chunks <-chan *core.StreamChunk) error {
	var responseText strings.Builder

	for chunk := range chunks {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Handle different chunk types
		switch chunk.Type {
		case "content_block_start":
			a.handleContentBlockStart(chunk)

		case "content_block_delta":
			if chunk.Delta != nil {
				switch chunk.Delta.Type {
				case "text_delta":
					responseText.WriteString(chunk.Delta.Text)
					a.emit(tux.Event{
						Type: tux.EventText,
						Text: chunk.Delta.Text,
					})
				case "input_json_delta":
					if a.assemblingTool != nil {
						a.toolInputJSONBuf.WriteString(chunk.Delta.PartialJSON)
					}
				}
			}

		case "content_block_stop":
			a.handleContentBlockStop()

		case "message_stop":
			// Add assistant response to history if there's text
			if responseText.Len() > 0 {
				a.mu.Lock()
				a.messages = append(a.messages, core.Message{
					Role:    "assistant",
					Content: responseText.String(),
				})
				a.mu.Unlock()
			}

			// Process any pending tools
			if len(a.pendingTools) > 0 {
				if err := a.processTools(ctx); err != nil {
					return err
				}
			} else {
				a.emit(tux.Event{Type: tux.EventComplete})
			}
		}
	}

	return nil
}
```

**Step 3: Update Run() to use processStream**

```go
func (a *HexAgent) Run(ctx context.Context, prompt string) error {
	// Create cancellable context
	ctx, cancel := context.WithCancel(ctx)
	a.mu.Lock()
	a.cancel = cancel
	a.mu.Unlock()

	defer func() {
		a.mu.Lock()
		if a.events != nil {
			close(a.events)
			a.events = nil
		}
		a.cancel = nil
		a.mu.Unlock()
	}()

	// Reset tool state for new run
	a.resetToolState()

	// Add user message to history
	a.mu.Lock()
	a.messages = append(a.messages, core.Message{
		Role:    "user",
		Content: prompt,
	})
	messages := make([]core.Message, len(a.messages))
	copy(messages, a.messages)
	a.mu.Unlock()

	// Build request
	req := core.MessageRequest{
		Model:     a.model,
		Messages:  messages,
		MaxTokens: 8192,
		Stream:    true,
		System:    a.systemPrompt,
	}

	// Start streaming
	chunks, err := a.client.CreateMessageStream(ctx, req)
	if err != nil {
		a.emit(tux.Event{Type: tux.EventError, Error: err})
		return err
	}

	// Process stream
	return a.processStream(ctx, chunks)
}
```

**Step 4: Verify build**

```bash
go build ./internal/tui/...
```

**Step 5: Commit**

```bash
git add internal/tui/agent.go
git commit -m "feat: add tool result continuation and refactor stream processing"
```

---

## Task 10: Wire Up Tool Definitions

**Files:**
- Modify: `internal/tui/agent.go`
- Modify: `cmd/hex/root.go`

**Step 1: Add GetToolDefinitions method to HexAgent**

```go
// GetToolDefinitions returns tool definitions for API requests.
func (a *HexAgent) GetToolDefinitions() []core.ToolDefinition {
	if a.executor == nil || a.executor.Registry() == nil {
		return nil
	}
	return a.executor.Registry().GetDefinitions()
}
```

**Step 2: Update request building to include tools**

In the `Run()` method and `continueWithToolResults()`, update request building:

```go
req := core.MessageRequest{
	Model:     a.model,
	Messages:  messages,
	MaxTokens: 8192,
	Stream:    true,
	System:    a.systemPrompt,
	Tools:     a.GetToolDefinitions(),
}
```

**Step 3: Update root.go to properly initialize executor**

```go
// In the useTux block:

// Create tool registry with available tools
registry := tools.NewRegistry()
// Register tools (or use default registration if available)

// Create executor
executor := tools.NewExecutor(registry, nil, nil, nil)

return runTuxMode(providerCfg.APIKey, tuxModel, systemPrompt, executor)
```

**Step 4: Verify build**

```bash
go build ./...
```

**Step 5: Commit**

```bash
git add internal/tui/agent.go cmd/hex/root.go
git commit -m "feat: wire tool definitions to API requests"
```

---

## Task 11: Add Tool Support Tests

**Files:**
- Modify: `internal/tui/agent_test.go`

**Step 1: Add test for tool parsing state**

```go
func TestHexAgent_ResetToolState(t *testing.T) {
	agent := newTestAgent("test-model", "test system")

	// Set some state
	agent.assemblingTool = &core.ToolUse{ID: "test"}
	agent.toolInputJSONBuf.WriteString("test json")
	agent.pendingTools = []*core.ToolUse{{ID: "pending"}}

	// Reset
	agent.resetToolState()

	assert.Nil(t, agent.assemblingTool)
	assert.Equal(t, "", agent.toolInputJSONBuf.String())
	assert.Nil(t, agent.pendingTools)
}
```

**Step 2: Add test for content block start handling**

```go
func TestHexAgent_HandleContentBlockStart_ToolUse(t *testing.T) {
	agent := newTestAgent("test-model", "test system")
	ch := agent.Subscribe()

	chunk := &core.StreamChunk{
		Type: "content_block_start",
		ContentBlock: &core.Content{
			Type: "tool_use",
			ID:   "tool_123",
			Name: "read_file",
		},
	}

	agent.handleContentBlockStart(chunk)

	// Verify assembling tool is set
	require.NotNil(t, agent.assemblingTool)
	assert.Equal(t, "tool_123", agent.assemblingTool.ID)
	assert.Equal(t, "read_file", agent.assemblingTool.Name)

	// Verify event was emitted
	select {
	case event := <-ch:
		assert.Equal(t, tux.EventToolCall, event.Type)
		assert.Equal(t, "tool_123", event.ToolID)
		assert.Equal(t, "read_file", event.ToolName)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for event")
	}
}
```

**Step 3: Add test for content block stop handling**

```go
func TestHexAgent_HandleContentBlockStop(t *testing.T) {
	agent := newTestAgent("test-model", "test system")

	// Set up assembling tool with JSON
	agent.assemblingTool = &core.ToolUse{
		ID:    "tool_123",
		Name:  "read_file",
		Input: make(map[string]interface{}),
	}
	agent.toolInputJSONBuf.WriteString(`{"path":"/test.txt"}`)

	agent.handleContentBlockStop()

	// Verify tool was added to pending
	require.Len(t, agent.pendingTools, 1)
	assert.Equal(t, "tool_123", agent.pendingTools[0].ID)
	assert.Equal(t, "/test.txt", agent.pendingTools[0].Input["path"])

	// Verify assembling state was cleared
	assert.Nil(t, agent.assemblingTool)
	assert.Equal(t, "", agent.toolInputJSONBuf.String())
}
```

**Step 4: Run tests**

```bash
go test ./internal/tui/... -v
```

**Step 5: Commit**

```bash
git add internal/tui/agent_test.go
git commit -m "test: add tool support tests for HexAgent"
```

---

## Task 12: Manual Integration Test

**Files:**
- None (manual testing)

**Step 1: Build**

```bash
make build
```

**Step 2: Test with tool-using prompt**

```bash
cd /Users/dylanr/work/hex/.worktrees/tux-migration
source .env
./bin/hex --tux
```

Type: "Read the file internal/tui/agent.go and tell me what it does"

**Expected behavior:**
1. Streaming text starts
2. Tool call appears (read_file)
3. Approval modal shows
4. After approval, tool executes
5. Tool result appears
6. Assistant continues with response

**Step 3: Verify**
- [ ] Tool call shown in UI
- [ ] Approval modal appears
- [ ] Can approve/deny
- [ ] Tool result displayed
- [ ] Conversation continues

**Step 4: Commit integration test notes**

```bash
git add -A
git commit -m "chore: phase 2 tool support complete"
```

---

## Summary

Phase 2 adds complete tool support to HexAgent:

| Task | Description |
|------|-------------|
| 1 | Add tool executor to HexAgent |
| 2 | Add tool parsing state |
| 3 | Handle content_block_start for tools |
| 4 | Handle content_block_delta for tool params |
| 5 | Handle content_block_stop |
| 6 | Implement tool processing loop |
| 7 | Implement approval request |
| 8 | Implement tool execution |
| 9 | Implement continue with tool results |
| 10 | Wire up tool definitions |
| 11 | Add tool support tests |
| 12 | Manual integration test |

After Phase 2, the tux-based hex will be able to:
- Stream text responses ✓ (Phase 1)
- Parse and display tool calls
- Request user approval for tools
- Execute approved tools
- Display tool results
- Continue conversations after tool use
