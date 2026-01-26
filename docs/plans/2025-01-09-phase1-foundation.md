# Phase 1: Foundation Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Create HexAgent implementing tux's Agent interface and wire up basic streaming chat.

**Architecture:** HexAgent wraps hex's API client to emit tux Events. A new entry point (`cmd/hex/tux.go`) uses `tux.New(agent)` instead of the old `ui.NewModel()`. Phase 1 focuses only on text streaming - tool support comes in Phase 2.

**Tech Stack:** Go, tux (github.com/2389-research/tux), hex's existing core/services

---

## Task 1: Add tux dependency

**Files:**
- Modify: `go.mod`

**Step 1: Add tux module**

```bash
cd /Users/dylanr/work/hex/.worktrees/tux-migration
go get github.com/2389-research/tux@latest
```

**Step 2: Verify import works**

```bash
go mod tidy
go build ./...
```

Expected: Build succeeds

**Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "deps: add tux dependency"
```

---

## Task 2: Create HexAgent struct

**Files:**
- Create: `internal/tui/agent.go`

**Step 1: Create the tui package and agent file**

```go
// ABOUTME: HexAgent implements tux.Agent interface
// ABOUTME: Wraps hex's API client to emit tux-compatible events

package tui

import (
	"context"
	"sync"

	"github.com/2389-research/hex/internal/core"
	"github.com/2389-research/tux"
)

// HexAgent implements tux.Agent by wrapping hex's API client.
type HexAgent struct {
	client       *core.Client
	model        string
	systemPrompt string

	// Conversation state
	messages []core.Message
	mu       sync.Mutex

	// Current run state
	events chan tux.Event
	cancel context.CancelFunc
}

// NewHexAgent creates a new HexAgent with the given API client.
func NewHexAgent(client *core.Client, model string, systemPrompt string) *HexAgent {
	return &HexAgent{
		client:       client,
		model:        model,
		systemPrompt: systemPrompt,
		messages:     make([]core.Message, 0),
	}
}

// Run starts the agent with the given prompt.
// It runs until completion or context cancellation.
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
	req := &core.MessageRequest{
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
	var responseText string
	for chunk := range chunks {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Handle different chunk types
		switch chunk.Type {
		case "content_block_delta":
			if chunk.Delta != nil && chunk.Delta.Type == "text_delta" {
				responseText += chunk.Delta.Text
				a.emit(tux.Event{
					Type: tux.EventText,
					Text: chunk.Delta.Text,
				})
			}
		case "message_stop":
			// Add assistant response to history
			a.mu.Lock()
			a.messages = append(a.messages, core.Message{
				Role:    "assistant",
				Content: responseText,
			})
			a.mu.Unlock()

			a.emit(tux.Event{Type: tux.EventComplete})
		}
	}

	return nil
}

// Subscribe returns a channel of events from the agent.
// The channel is closed when the agent completes.
func (a *HexAgent) Subscribe() <-chan tux.Event {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Create new events channel
	a.events = make(chan tux.Event, 100)
	return a.events
}

// Cancel cancels the current agent run.
func (a *HexAgent) Cancel() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.cancel != nil {
		a.cancel()
	}
}

// emit sends an event to subscribers.
func (a *HexAgent) emit(event tux.Event) {
	a.mu.Lock()
	ch := a.events
	a.mu.Unlock()

	if ch != nil {
		select {
		case ch <- event:
		default:
			// Channel full, drop event (shouldn't happen with buffered channel)
		}
	}
}

// AddSystemContext adds context to the system prompt.
func (a *HexAgent) AddSystemContext(context string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.systemPrompt += "\n\n" + context
}

// ClearHistory clears the conversation history.
func (a *HexAgent) ClearHistory() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.messages = make([]core.Message, 0)
}
```

**Step 2: Verify compilation**

```bash
go build ./internal/tui/...
```

Expected: Build succeeds

**Step 3: Commit**

```bash
git add internal/tui/agent.go
git commit -m "feat: add HexAgent implementing tux.Agent interface"
```

---

## Task 3: Create HexAgent tests

**Files:**
- Create: `internal/tui/agent_test.go`

**Step 1: Write tests**

```go
// ABOUTME: Tests for HexAgent
// ABOUTME: Verifies tux.Agent interface implementation

package tui

import (
	"context"
	"testing"
	"time"

	"github.com/2389-research/tux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHexAgent_ImplementsInterface(t *testing.T) {
	// Compile-time check that HexAgent implements tux.Agent
	var _ tux.Agent = (*HexAgent)(nil)
}

func TestHexAgent_Subscribe(t *testing.T) {
	agent := NewHexAgent(nil, "test-model", "test system")

	ch := agent.Subscribe()
	require.NotNil(t, ch)
}

func TestHexAgent_Cancel(t *testing.T) {
	agent := NewHexAgent(nil, "test-model", "test system")

	// Should not panic when no run is active
	agent.Cancel()
}

func TestHexAgent_ClearHistory(t *testing.T) {
	agent := NewHexAgent(nil, "test-model", "test system")

	// Add some messages manually
	agent.mu.Lock()
	agent.messages = append(agent.messages, struct {
		Role    string
		Content interface{}
	}{"user", "test"})
	agent.mu.Unlock()

	agent.ClearHistory()

	agent.mu.Lock()
	assert.Empty(t, agent.messages)
	agent.mu.Unlock()
}

func TestHexAgent_AddSystemContext(t *testing.T) {
	agent := NewHexAgent(nil, "test-model", "initial system")

	agent.AddSystemContext("additional context")

	agent.mu.Lock()
	assert.Contains(t, agent.systemPrompt, "initial system")
	assert.Contains(t, agent.systemPrompt, "additional context")
	agent.mu.Unlock()
}

func TestHexAgent_EmitToSubscribers(t *testing.T) {
	agent := NewHexAgent(nil, "test-model", "test system")

	ch := agent.Subscribe()

	// Emit an event
	agent.emit(tux.Event{Type: tux.EventText, Text: "hello"})

	// Should receive it
	select {
	case event := <-ch:
		assert.Equal(t, tux.EventText, event.Type)
		assert.Equal(t, "hello", event.Text)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for event")
	}
}
```

**Step 2: Run tests**

```bash
go test ./internal/tui/... -v
```

Expected: All tests pass

**Step 3: Commit**

```bash
git add internal/tui/agent_test.go
git commit -m "test: add HexAgent tests"
```

---

## Task 4: Create tux entry point

**Files:**
- Create: `cmd/hex/tux.go`

**Step 1: Create tux entry point**

```go
// ABOUTME: Entry point for tux-based TUI
// ABOUTME: Uses tux.New() instead of ui.NewModel()

package main

import (
	"fmt"
	"os"

	"github.com/2389-research/hex/internal/core"
	"github.com/2389-research/hex/internal/tui"
	"github.com/2389-research/tux"
	"github.com/2389-research/tux/theme"
)

// runTuxMode starts the tux-based TUI.
// This is called when --tux flag is passed.
func runTuxMode(apiKey, model, systemPrompt string) error {
	// Create API client
	client := core.NewClient(apiKey)

	// Create HexAgent
	agent := tui.NewHexAgent(client, model, systemPrompt)

	// Create tux app with Dracula theme
	app := tux.New(agent,
		tux.WithTheme(theme.NewDraculaTheme()),
	)

	// Run the app
	if err := app.Run(); err != nil {
		return fmt.Errorf("tux app error: %w", err)
	}

	return nil
}
```

**Step 2: Verify compilation**

```bash
go build ./cmd/hex/...
```

Expected: Build succeeds (even if function not yet wired in)

**Step 3: Commit**

```bash
git add cmd/hex/tux.go
git commit -m "feat: add tux entry point"
```

---

## Task 5: Wire --tux flag to root command

**Files:**
- Modify: `cmd/hex/root.go`

**Step 1: Find where flags are defined**

Look for the existing flag definitions (likely near `var rootCmd = &cobra.Command{...}`) and add:

```go
var useTux bool
```

In `init()`:

```go
rootCmd.PersistentFlags().BoolVar(&useTux, "tux", false, "Use tux-based TUI (experimental)")
```

**Step 2: Find the run function and add tux branch**

In the command's Run or RunE function, before the existing interactive mode code, add:

```go
if useTux {
    apiKey := os.Getenv("ANTHROPIC_API_KEY")
    if apiKey == "" {
        return fmt.Errorf("ANTHROPIC_API_KEY environment variable not set")
    }
    return runTuxMode(apiKey, model, systemPrompt)
}
```

Note: You'll need to determine how `model` and `systemPrompt` are currently obtained and use the same pattern.

**Step 3: Verify build**

```bash
go build ./cmd/hex/...
```

Expected: Build succeeds

**Step 4: Commit**

```bash
git add cmd/hex/root.go
git commit -m "feat: wire --tux flag to root command"
```

---

## Task 6: Test basic streaming manually

**Step 1: Build hex**

```bash
make build
```

**Step 2: Run with --tux flag**

```bash
export ANTHROPIC_API_KEY=your-key
./bin/hex --tux
```

**Step 3: Test basic interaction**

- Type "Hello" and press Enter
- Verify streaming text appears
- Verify response completes
- Type another message and verify multi-turn works

**Step 4: Document any issues**

If there are issues, note them for the next task. Common issues:
- Event channel not being read
- Stream not closing properly
- Theme/styling issues

**Step 5: Commit any fixes**

```bash
git add -A
git commit -m "fix: address issues from manual testing"
```

---

## Task 7: Update acceptance tests for tux mode

**Files:**
- Create: `test/acceptance/tux_adapter.go`

**Step 1: Create TuxAdapter skeleton**

```go
// ABOUTME: Tux adapter implementing TUIHarness
// ABOUTME: Wraps tux.App for acceptance testing

package acceptance

import (
	"errors"
	"time"

	"github.com/2389-research/tux"
)

// TuxAdapter wraps a tux.App to implement TUIHarness.
// This allows the same acceptance tests to run against both UIs.
type TuxAdapter struct {
	app    *tux.App
	agent  *MockAgent
	width  int
	height int
}

// MockAgent is a mock tux.Agent for testing.
type MockAgent struct {
	events chan tux.Event
}

func (m *MockAgent) Run(ctx context.Context, prompt string) error {
	// Mock implementation - controlled by test
	return nil
}

func (m *MockAgent) Subscribe() <-chan tux.Event {
	return m.events
}

func (m *MockAgent) Cancel() {
	// Mock implementation
}

// NewTuxAdapter creates a new TuxAdapter for acceptance testing.
func NewTuxAdapter() *TuxAdapter {
	return &TuxAdapter{
		agent: &MockAgent{
			events: make(chan tux.Event, 100),
		},
	}
}

func (a *TuxAdapter) Init(width, height int) error {
	// TODO: Initialize tux app with mock agent
	a.width = width
	a.height = height
	return errors.New("TuxAdapter not yet implemented")
}

// ... implement remaining TUIHarness methods as stubs that return errors
// This skeleton allows the code to compile while indicating work needed
```

**Step 2: Verify compilation**

```bash
go build ./test/acceptance/...
```

Expected: Build succeeds (even with stub implementation)

**Step 3: Commit**

```bash
git add test/acceptance/tux_adapter.go
git commit -m "feat: add TuxAdapter skeleton for acceptance tests"
```

---

## Task 8: Verify acceptance tests still pass

**Step 1: Run acceptance tests**

```bash
go test ./test/acceptance/... -v
```

Expected: All existing tests pass (they use BubbleteaAdapter)

**Step 2: Run full test suite**

```bash
go test ./... -short
```

Expected: All tests pass

**Step 3: Run with race detector**

```bash
go test ./test/acceptance/... -race
```

Expected: No race conditions

**Step 4: Final commit**

```bash
git add -A
git commit -m "test: verify Phase 1 foundation complete

Phase 1 delivers:
- HexAgent implementing tux.Agent interface
- tux entry point with --tux flag
- Basic streaming chat working
- TuxAdapter skeleton for future acceptance test migration"
```

---

## Success Criteria

Phase 1 is complete when:

- [ ] `./bin/hex --tux` launches tux-based TUI
- [ ] User can type a message and receive streaming response
- [ ] Multi-turn conversation works (history maintained)
- [ ] Ctrl+C cancels streaming
- [ ] All existing tests still pass
- [ ] No race conditions

## Known Limitations (Phase 1)

- No tool support (Phase 2)
- No conversation persistence (Phase 3)
- No autocomplete/suggestions (Phase 4)
- No configuration/theming customization (Phase 5)

## Tux Gaps to Report

If you encounter missing features in tux during implementation, report them:

```
Blocked - tux is missing: [describe the gap]
```

The maintainer will fix these in tux and you can continue.
