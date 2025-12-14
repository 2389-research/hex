# Hex-Mux Integration Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Replace hex's AgentOrchestrator with mux's orchestrator while preserving subprocess-based subagent spawning.

**Architecture:** hex CLI bootstraps a mux Agent with adapted hex tools. Subagents spawn as subprocesses with tool restrictions passed via env vars, creating FilteredRegistry on startup.

**Tech Stack:** Go, mux library, anthropic-sdk-go

---

## Task 1: Add Anthropic SDK to mux

**Files:**
- Modify: `/Users/harper/Public/src/2389/mux/go.mod`

**Step 1: Add anthropic-sdk-go dependency**

```bash
cd /Users/harper/Public/src/2389/mux && go get github.com/anthropics/anthropic-sdk-go
```

**Step 2: Verify dependency added**

Run: `grep anthropic go.mod`
Expected: `github.com/anthropics/anthropic-sdk-go`

**Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "feat(llm): add anthropic-sdk-go dependency"
```

---

## Task 2: Implement Anthropic Client - Types and Constructor

**Files:**
- Create: `/Users/harper/Public/src/2389/mux/llm/anthropic.go`
- Test: `/Users/harper/Public/src/2389/mux/llm/anthropic_test.go`

**Step 1: Write the failing test**

```go
// /Users/harper/Public/src/2389/mux/llm/anthropic_test.go
// ABOUTME: Tests for the Anthropic LLM client implementation.
// ABOUTME: Verifies API communication, streaming, and error handling.
package llm

import (
	"testing"
)

func TestNewAnthropicClient(t *testing.T) {
	client := NewAnthropicClient("test-api-key", "claude-sonnet-4-20250514")
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.model != "claude-sonnet-4-20250514" {
		t.Errorf("expected model claude-sonnet-4-20250514, got %s", client.model)
	}
}

func TestNewAnthropicClientDefaultModel(t *testing.T) {
	client := NewAnthropicClient("test-api-key", "")
	if client.model != "claude-sonnet-4-20250514" {
		t.Errorf("expected default model claude-sonnet-4-20250514, got %s", client.model)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/harper/Public/src/2389/mux && go test ./llm/... -run TestNewAnthropicClient -v`
Expected: FAIL with "undefined: NewAnthropicClient"

**Step 3: Write minimal implementation**

```go
// /Users/harper/Public/src/2389/mux/llm/anthropic.go
// ABOUTME: Anthropic API client implementing the llm.Client interface.
// ABOUTME: Handles both streaming and non-streaming message creation.
package llm

import (
	"context"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// AnthropicClient implements Client for the Anthropic API.
type AnthropicClient struct {
	client *anthropic.Client
	model  string
}

// NewAnthropicClient creates a new Anthropic API client.
func NewAnthropicClient(apiKey, model string) *AnthropicClient {
	if model == "" {
		model = "claude-sonnet-4-20250514"
	}
	return &AnthropicClient{
		client: anthropic.NewClient(option.WithAPIKey(apiKey)),
		model:  model,
	}
}

// CreateMessage sends a message and returns the complete response.
func (a *AnthropicClient) CreateMessage(ctx context.Context, req *Request) (*Response, error) {
	// TODO: implement in next task
	return nil, nil
}

// CreateMessageStream sends a message and returns a channel of streaming events.
func (a *AnthropicClient) CreateMessageStream(ctx context.Context, req *Request) (<-chan StreamEvent, error) {
	// TODO: implement in next task
	return nil, nil
}

// Compile-time interface assertion.
var _ Client = (*AnthropicClient)(nil)
```

**Step 4: Run test to verify it passes**

Run: `cd /Users/harper/Public/src/2389/mux && go test ./llm/... -run TestNewAnthropicClient -v`
Expected: PASS

**Step 5: Commit**

```bash
cd /Users/harper/Public/src/2389/mux
git add llm/anthropic.go llm/anthropic_test.go
git commit -m "feat(llm): add AnthropicClient constructor"
```

---

## Task 3: Implement CreateMessage

**Files:**
- Modify: `/Users/harper/Public/src/2389/mux/llm/anthropic.go`
- Modify: `/Users/harper/Public/src/2389/mux/llm/anthropic_test.go`

**Step 1: Write the failing test**

Add to `anthropic_test.go`:

```go
func TestAnthropicClientCreateMessage_ConvertsRequest(t *testing.T) {
	// This test verifies request conversion logic
	// We can't call real API without key, so test the conversion helpers
	req := &Request{
		Model:     "claude-sonnet-4-20250514",
		MaxTokens: 1024,
		System:    "You are helpful.",
		Messages: []Message{
			{Role: RoleUser, Content: "Hello"},
		},
		Tools: []ToolDefinition{
			{Name: "read", Description: "Read a file", InputSchema: map[string]any{"type": "object"}},
		},
	}

	params := convertRequest(req)
	if params.Model != "claude-sonnet-4-20250514" {
		t.Errorf("expected model claude-sonnet-4-20250514, got %s", params.Model)
	}
	if params.MaxTokens != 1024 {
		t.Errorf("expected max_tokens 1024, got %d", params.MaxTokens)
	}
	if len(params.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(params.Messages))
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/harper/Public/src/2389/mux && go test ./llm/... -run TestAnthropicClientCreateMessage -v`
Expected: FAIL with "undefined: convertRequest"

**Step 3: Write implementation**

Add to `anthropic.go`:

```go
// convertRequest converts our Request to Anthropic's MessageNewParams.
func convertRequest(req *Request) anthropic.MessageNewParams {
	params := anthropic.MessageNewParams{
		Model:     req.Model,
		MaxTokens: int64(req.MaxTokens),
	}

	// Convert messages
	for _, msg := range req.Messages {
		var content []anthropic.ContentBlockParamUnion
		if msg.Content != "" {
			content = append(content, anthropic.NewTextBlock(msg.Content))
		}
		for _, block := range msg.Blocks {
			switch block.Type {
			case ContentTypeText:
				content = append(content, anthropic.NewTextBlock(block.Text))
			case ContentTypeToolResult:
				content = append(content, anthropic.NewToolResultBlock(block.ToolUseID, block.Text, block.IsError))
			}
		}
		params.Messages = append(params.Messages, anthropic.MessageParam{
			Role:    anthropic.MessageParamRole(msg.Role),
			Content: content,
		})
	}

	// Set system prompt
	if req.System != "" {
		params.System = []anthropic.TextBlockParam{{Text: req.System}}
	}

	// Convert tools
	for _, tool := range req.Tools {
		params.Tools = append(params.Tools, anthropic.ToolParam{
			Name:        tool.Name,
			Description: anthropic.String(tool.Description),
			InputSchema: anthropic.ToolInputSchemaParam{
				Properties: tool.InputSchema,
			},
		})
	}

	return params
}

// convertResponse converts Anthropic's Message to our Response.
func convertResponse(msg *anthropic.Message) *Response {
	resp := &Response{
		ID:         msg.ID,
		Model:      msg.Model,
		StopReason: StopReason(msg.StopReason),
		Usage: Usage{
			InputTokens:  int(msg.Usage.InputTokens),
			OutputTokens: int(msg.Usage.OutputTokens),
		},
	}

	for _, block := range msg.Content {
		switch b := block.AsAny().(type) {
		case anthropic.TextBlock:
			resp.Content = append(resp.Content, ContentBlock{
				Type: ContentTypeText,
				Text: b.Text,
			})
		case anthropic.ToolUseBlock:
			resp.Content = append(resp.Content, ContentBlock{
				Type:  ContentTypeToolUse,
				ID:    b.ID,
				Name:  b.Name,
				Input: b.Input,
			})
		}
	}

	return resp
}

// CreateMessage sends a message and returns the complete response.
func (a *AnthropicClient) CreateMessage(ctx context.Context, req *Request) (*Response, error) {
	if req.Model == "" {
		req.Model = a.model
	}
	if req.MaxTokens == 0 {
		req.MaxTokens = 4096
	}

	params := convertRequest(req)
	msg, err := a.client.Messages.New(ctx, params)
	if err != nil {
		return nil, err
	}

	return convertResponse(msg), nil
}
```

**Step 4: Run test to verify it passes**

Run: `cd /Users/harper/Public/src/2389/mux && go test ./llm/... -run TestAnthropicClientCreateMessage -v`
Expected: PASS

**Step 5: Commit**

```bash
cd /Users/harper/Public/src/2389/mux
git add llm/anthropic.go llm/anthropic_test.go
git commit -m "feat(llm): implement CreateMessage for Anthropic"
```

---

## Task 4: Implement CreateMessageStream

**Files:**
- Modify: `/Users/harper/Public/src/2389/mux/llm/anthropic.go`
- Modify: `/Users/harper/Public/src/2389/mux/llm/anthropic_test.go`

**Step 1: Write implementation**

Replace the stub `CreateMessageStream` in `anthropic.go`:

```go
// CreateMessageStream sends a message and returns a channel of streaming events.
func (a *AnthropicClient) CreateMessageStream(ctx context.Context, req *Request) (<-chan StreamEvent, error) {
	if req.Model == "" {
		req.Model = a.model
	}
	if req.MaxTokens == 0 {
		req.MaxTokens = 4096
	}

	params := convertRequest(req)
	stream := a.client.Messages.NewStreaming(ctx, params)

	eventChan := make(chan StreamEvent, 100)

	go func() {
		defer close(eventChan)

		for stream.Next() {
			event := stream.Current()
			switch e := event.AsAny().(type) {
			case anthropic.MessageStartEvent:
				eventChan <- StreamEvent{
					Type:     EventMessageStart,
					Response: convertResponse(&e.Message),
				}
			case anthropic.ContentBlockStartEvent:
				eventChan <- StreamEvent{
					Type:  EventContentStart,
					Index: int(e.Index),
				}
			case anthropic.ContentBlockDeltaEvent:
				var text string
				if delta, ok := e.Delta.AsAny().(anthropic.TextDelta); ok {
					text = delta.Text
				}
				eventChan <- StreamEvent{
					Type:  EventContentDelta,
					Index: int(e.Index),
					Text:  text,
				}
			case anthropic.ContentBlockStopEvent:
				eventChan <- StreamEvent{
					Type:  EventContentStop,
					Index: int(e.Index),
				}
			case anthropic.MessageDeltaEvent:
				eventChan <- StreamEvent{
					Type: EventMessageDelta,
				}
			case anthropic.MessageStopEvent:
				eventChan <- StreamEvent{
					Type: EventMessageStop,
				}
			}
		}

		if err := stream.Err(); err != nil {
			eventChan <- StreamEvent{
				Type:  EventError,
				Error: err,
			}
		}
	}()

	return eventChan, nil
}
```

**Step 2: Run all tests**

Run: `cd /Users/harper/Public/src/2389/mux && go test ./llm/... -v`
Expected: PASS

**Step 3: Commit**

```bash
cd /Users/harper/Public/src/2389/mux
git add llm/anthropic.go
git commit -m "feat(llm): implement CreateMessageStream for Anthropic"
```

---

## Task 5: Add mux as dependency to hex

**Files:**
- Modify: `/Users/harper/Public/src/2389/hex/go.mod`

**Step 1: Add mux dependency**

```bash
cd /Users/harper/Public/src/2389/hex && go get github.com/2389-research/mux@latest
```

Note: If mux is not published, use replace directive:

```bash
echo 'replace github.com/2389-research/mux => ../mux' >> go.mod
go mod tidy
```

**Step 2: Verify dependency**

Run: `grep mux go.mod`
Expected: Shows mux dependency

**Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "feat: add mux as dependency"
```

---

## Task 6: Create hex tool adapter

**Files:**
- Create: `/Users/harper/Public/src/2389/hex/internal/adapter/tool.go`
- Create: `/Users/harper/Public/src/2389/hex/internal/adapter/tool_test.go`

**Step 1: Write the failing test**

```go
// /Users/harper/Public/src/2389/hex/internal/adapter/tool_test.go
// ABOUTME: Tests for the hex-to-mux tool adapter.
// ABOUTME: Verifies that hex tools are correctly wrapped for mux.
package adapter

import (
	"context"
	"testing"

	"github.com/2389-research/hex/internal/tools"
)

type mockHexTool struct {
	name        string
	description string
	result      *tools.Result
	err         error
}

func (m *mockHexTool) Name() string        { return m.name }
func (m *mockHexTool) Description() string { return m.description }
func (m *mockHexTool) RequiresApproval(params map[string]interface{}) bool { return false }
func (m *mockHexTool) Execute(ctx context.Context, params map[string]interface{}) (*tools.Result, error) {
	return m.result, m.err
}

func TestAdaptTool(t *testing.T) {
	hexTool := &mockHexTool{
		name:        "test_tool",
		description: "A test tool",
		result: &tools.Result{
			ToolName: "test_tool",
			Success:  true,
			Output:   "test output",
		},
	}

	adapted := AdaptTool(hexTool)

	if adapted.Name() != "test_tool" {
		t.Errorf("expected name test_tool, got %s", adapted.Name())
	}
	if adapted.Description() != "A test tool" {
		t.Errorf("expected description 'A test tool', got %s", adapted.Description())
	}

	result, err := adapted.Execute(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Error("expected success true")
	}
	if result.Output != "test output" {
		t.Errorf("expected output 'test output', got %s", result.Output)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/harper/Public/src/2389/hex && go test ./internal/adapter/... -run TestAdaptTool -v`
Expected: FAIL with package not found or "undefined: AdaptTool"

**Step 3: Write implementation**

```go
// /Users/harper/Public/src/2389/hex/internal/adapter/tool.go
// ABOUTME: Adapter that wraps hex tools to implement mux's tool.Tool interface.
// ABOUTME: Allows hex's existing tools to work with mux's orchestrator.
package adapter

import (
	"context"

	"github.com/2389-research/hex/internal/tools"
	muxtool "github.com/2389-research/mux/tool"
)

// adaptedTool wraps a hex Tool to implement mux's Tool interface.
type adaptedTool struct {
	hex tools.Tool
}

// AdaptTool wraps a hex tool for use with mux.
func AdaptTool(t tools.Tool) muxtool.Tool {
	return &adaptedTool{hex: t}
}

// AdaptAll wraps multiple hex tools for use with mux.
func AdaptAll(hexTools []tools.Tool) []muxtool.Tool {
	adapted := make([]muxtool.Tool, len(hexTools))
	for i, t := range hexTools {
		adapted[i] = AdaptTool(t)
	}
	return adapted
}

func (a *adaptedTool) Name() string {
	return a.hex.Name()
}

func (a *adaptedTool) Description() string {
	return a.hex.Description()
}

func (a *adaptedTool) RequiresApproval(params map[string]any) bool {
	// Convert map[string]any to map[string]interface{}
	converted := make(map[string]interface{}, len(params))
	for k, v := range params {
		converted[k] = v
	}
	return a.hex.RequiresApproval(converted)
}

func (a *adaptedTool) Execute(ctx context.Context, params map[string]any) (*muxtool.Result, error) {
	// Convert map[string]any to map[string]interface{}
	converted := make(map[string]interface{}, len(params))
	for k, v := range params {
		converted[k] = v
	}

	result, err := a.hex.Execute(ctx, converted)
	if err != nil {
		return nil, err
	}

	return &muxtool.Result{
		ToolName: result.ToolName,
		Success:  result.Success,
		Output:   result.Output,
		Error:    result.Error,
	}, nil
}

// Compile-time interface assertion.
var _ muxtool.Tool = (*adaptedTool)(nil)
```

**Step 4: Run test to verify it passes**

Run: `cd /Users/harper/Public/src/2389/hex && go test ./internal/adapter/... -run TestAdaptTool -v`
Expected: PASS

**Step 5: Commit**

```bash
cd /Users/harper/Public/src/2389/hex
git add internal/adapter/tool.go internal/adapter/tool_test.go
git commit -m "feat(adapter): add hex-to-mux tool adapter"
```

---

## Task 7: Create bootstrap package

**Files:**
- Create: `/Users/harper/Public/src/2389/hex/internal/adapter/bootstrap.go`
- Create: `/Users/harper/Public/src/2389/hex/internal/adapter/bootstrap_test.go`

**Step 1: Write the failing test**

```go
// /Users/harper/Public/src/2389/hex/internal/adapter/bootstrap_test.go
// ABOUTME: Tests for agent bootstrap functions.
// ABOUTME: Verifies root and subagent creation with proper tool filtering.
package adapter

import (
	"os"
	"testing"
)

func TestParseCSV(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"", nil},
		{"Read", []string{"Read"}},
		{"Read,Grep,Glob", []string{"Read", "Grep", "Glob"}},
		{"Read, Grep, Glob", []string{"Read", "Grep", "Glob"}},
	}

	for _, tc := range tests {
		result := parseCSV(tc.input)
		if len(result) != len(tc.expected) {
			t.Errorf("parseCSV(%q): expected %v, got %v", tc.input, tc.expected, result)
			continue
		}
		for i := range result {
			if result[i] != tc.expected[i] {
				t.Errorf("parseCSV(%q)[%d]: expected %q, got %q", tc.input, i, tc.expected[i], result[i])
			}
		}
	}
}

func TestIsSubagent(t *testing.T) {
	// Clean env
	os.Unsetenv("HEX_SUBAGENT_TYPE")

	if IsSubagent() {
		t.Error("expected IsSubagent() to return false when env not set")
	}

	os.Setenv("HEX_SUBAGENT_TYPE", "Explore")
	defer os.Unsetenv("HEX_SUBAGENT_TYPE")

	if !IsSubagent() {
		t.Error("expected IsSubagent() to return true when env is set")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/harper/Public/src/2389/hex && go test ./internal/adapter/... -run 'TestParseCSV|TestIsSubagent' -v`
Expected: FAIL

**Step 3: Write implementation**

```go
// /Users/harper/Public/src/2389/hex/internal/adapter/bootstrap.go
// ABOUTME: Bootstrap functions for creating mux agents in hex.
// ABOUTME: Handles root agent and subagent creation with proper tool filtering.
package adapter

import (
	"os"
	"strings"

	"github.com/2389-research/hex/internal/tools"
	"github.com/2389-research/mux/agent"
	"github.com/2389-research/mux/llm"
	muxtool "github.com/2389-research/mux/tool"
)

// IsSubagent returns true if this process is running as a subagent.
func IsSubagent() bool {
	return os.Getenv("HEX_SUBAGENT_TYPE") != ""
}

// parseCSV parses a comma-separated string into a slice.
func parseCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// Config holds configuration for creating an agent.
type Config struct {
	APIKey       string
	Model        string
	SystemPrompt string
	HexTools     []tools.Tool
}

// NewRootAgent creates a root agent with full tool access.
func NewRootAgent(cfg Config) *agent.Agent {
	llmClient := llm.NewAnthropicClient(cfg.APIKey, cfg.Model)

	registry := muxtool.NewRegistry()
	for _, hexTool := range cfg.HexTools {
		registry.Register(AdaptTool(hexTool))
	}

	return agent.New(agent.Config{
		Name:         "hex-root",
		Registry:     registry,
		LLMClient:    llmClient,
		SystemPrompt: cfg.SystemPrompt,
	})
}

// NewSubagent creates a subagent with filtered tool access based on env vars.
func NewSubagent(cfg Config) *agent.Agent {
	llmClient := llm.NewAnthropicClient(cfg.APIKey, cfg.Model)

	registry := muxtool.NewRegistry()
	for _, hexTool := range cfg.HexTools {
		registry.Register(AdaptTool(hexTool))
	}

	allowed := parseCSV(os.Getenv("HEX_ALLOWED_TOOLS"))
	denied := parseCSV(os.Getenv("HEX_DENIED_TOOLS"))

	agentID := os.Getenv("HEX_AGENT_ID")
	if agentID == "" {
		agentID = "hex-subagent"
	}

	return agent.New(agent.Config{
		Name:         agentID,
		Registry:     registry,
		LLMClient:    llmClient,
		SystemPrompt: cfg.SystemPrompt,
		AllowedTools: allowed,
		DeniedTools:  denied,
	})
}
```

**Step 4: Run test to verify it passes**

Run: `cd /Users/harper/Public/src/2389/hex && go test ./internal/adapter/... -v`
Expected: PASS

**Step 5: Commit**

```bash
cd /Users/harper/Public/src/2389/hex
git add internal/adapter/bootstrap.go internal/adapter/bootstrap_test.go
git commit -m "feat(adapter): add agent bootstrap functions"
```

---

## Task 8: Update Task tool to pass tool restrictions

**Files:**
- Modify: `/Users/harper/Public/src/2389/hex/internal/tools/task_tool.go`

**Step 1: Find the subprocess spawning code**

Search for where env vars are set:
```bash
grep -n "HEX_SUBAGENT_TYPE" /Users/harper/Public/src/2389/hex/internal/tools/task_tool.go
```

**Step 2: Add tool restriction env vars**

Find the section where `cmd.Env` is set and add:

```go
// Get default tools for subagent type
config := subagents.DefaultConfig(req.Type)

// Add tool restrictions
if len(config.AllowedTools) > 0 {
	cmd.Env = append(cmd.Env,
		fmt.Sprintf("HEX_ALLOWED_TOOLS=%s", strings.Join(config.AllowedTools, ",")),
	)
}
if len(config.DeniedTools) > 0 {
	cmd.Env = append(cmd.Env,
		fmt.Sprintf("HEX_DENIED_TOOLS=%s", strings.Join(config.DeniedTools, ",")),
	)
}
```

**Step 3: Run tests**

Run: `cd /Users/harper/Public/src/2389/hex && go test ./internal/tools/... -v`
Expected: PASS

**Step 4: Commit**

```bash
cd /Users/harper/Public/src/2389/hex
git add internal/tools/task_tool.go
git commit -m "feat(tools): pass tool restrictions to subagent env vars"
```

---

## Task 9: Integration test - Root agent with mux

**Files:**
- Create: `/Users/harper/Public/src/2389/hex/internal/adapter/integration_test.go`

**Step 1: Write integration test**

```go
// /Users/harper/Public/src/2389/hex/internal/adapter/integration_test.go
// ABOUTME: Integration tests for hex-mux adapter.
// ABOUTME: Verifies end-to-end agent creation and tool execution.
package adapter

import (
	"context"
	"testing"

	"github.com/2389-research/hex/internal/tools"
	muxtool "github.com/2389-research/mux/tool"
)

func TestRootAgentToolExecution(t *testing.T) {
	// Create a simple mock tool
	mockTool := &mockHexTool{
		name:        "echo",
		description: "Echo back input",
		result: &tools.Result{
			ToolName: "echo",
			Success:  true,
			Output:   "echoed: hello",
		},
	}

	// Create registry with adapted tool
	registry := muxtool.NewRegistry()
	registry.Register(AdaptTool(mockTool))

	// Verify tool is registered
	tool, ok := registry.Get("echo")
	if !ok {
		t.Fatal("expected echo tool to be registered")
	}

	// Execute tool through mux
	result, err := tool.Execute(context.Background(), map[string]any{"input": "hello"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Output != "echoed: hello" {
		t.Errorf("expected output 'echoed: hello', got %s", result.Output)
	}
}

func TestSubagentToolFiltering(t *testing.T) {
	// Create multiple mock tools
	readTool := &mockHexTool{name: "Read", description: "Read files"}
	writeTool := &mockHexTool{name: "Write", description: "Write files"}
	bashTool := &mockHexTool{name: "Bash", description: "Execute bash"}

	// Create registry with all tools
	registry := muxtool.NewRegistry()
	registry.Register(AdaptTool(readTool))
	registry.Register(AdaptTool(writeTool))
	registry.Register(AdaptTool(bashTool))

	// Create filtered registry (simulating Explore subagent)
	filtered := muxtool.NewFilteredRegistry(registry, []string{"Read", "Bash"}, nil)

	// Verify filtering
	if _, ok := filtered.Get("Read"); !ok {
		t.Error("expected Read to be allowed")
	}
	if _, ok := filtered.Get("Bash"); !ok {
		t.Error("expected Bash to be allowed")
	}
	if _, ok := filtered.Get("Write"); ok {
		t.Error("expected Write to be denied")
	}

	// Verify count
	if filtered.Count() != 2 {
		t.Errorf("expected 2 tools, got %d", filtered.Count())
	}
}
```

**Step 2: Run integration tests**

Run: `cd /Users/harper/Public/src/2389/hex && go test ./internal/adapter/... -v`
Expected: PASS

**Step 3: Commit**

```bash
cd /Users/harper/Public/src/2389/hex
git add internal/adapter/integration_test.go
git commit -m "test(adapter): add integration tests for mux adapter"
```

---

## Task 10: Full test run and cleanup

**Step 1: Run all mux tests**

```bash
cd /Users/harper/Public/src/2389/mux && go test ./... -v
```
Expected: All tests pass

**Step 2: Run all hex adapter tests**

```bash
cd /Users/harper/Public/src/2389/hex && go test ./internal/adapter/... -v
```
Expected: All tests pass

**Step 3: Run hex pre-commit hooks**

```bash
cd /Users/harper/Public/src/2389/hex && git add -A && git status
```

**Step 4: Final commit**

```bash
cd /Users/harper/Public/src/2389/hex
git add -A
git commit -m "feat: complete hex-mux integration foundation

- Add mux as dependency
- Create tool adapter (hex tools → mux interface)
- Create bootstrap functions for root and subagent
- Pass tool restrictions via env vars to subagents
- Add integration tests

Tool restrictions are now enforced via mux's FilteredRegistry."
```

---

## Summary

| Task | Description | Location |
|------|-------------|----------|
| 1 | Add Anthropic SDK to mux | mux |
| 2 | Anthropic client constructor | mux |
| 3 | Implement CreateMessage | mux |
| 4 | Implement CreateMessageStream | mux |
| 5 | Add mux dependency to hex | hex |
| 6 | Create tool adapter | hex |
| 7 | Create bootstrap package | hex |
| 8 | Update Task tool env vars | hex |
| 9 | Integration tests | hex |
| 10 | Full test run | both |

After these tasks, hex will have the foundation to use mux's agent system. The next phase would be to modify `cmd/hex/main.go` to actually use the mux agent instead of hex's orchestrator.
