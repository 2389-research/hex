# Multi-Agent Features Integration Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Integrate event-sourcing, cost tracking, process registry, and visualization tools into production hex flows

**Architecture:** Wire up the Phase 1-3 multi-agent improvements (already built and tested) into the main execution paths. Enable event recording by default, add cost tracking to agent lifecycle, integrate process registry with cleanup, and expose visualization tools via CLI commands.

**Tech Stack:** Go, SQLite (for events), existing internal packages (orchestrator, events, cost, registry)

---

## Task 1: Wire Event Store to Root Command

**Files:**
- Modify: `cmd/hex/root.go:130-150`
- Test: Manual verification with event file
- Reference: `internal/events/store.go` (already implemented)

**Step 1: Add event store initialization to root command**

Location: `cmd/hex/root.go` around line 135 (after config loading, before agent execution)

```go
// Initialize event store
eventFile := filepath.Join(os.TempDir(), fmt.Sprintf("hex_events_%s.jsonl", time.Now().Format("20060102_150405")))
eventStore, err := events.NewEventStore(eventFile)
if err != nil {
	return fmt.Errorf("failed to create event store: %w", err)
}
defer func() { _ = eventStore.Close() }()

// Store in context or pass to orchestrator
ctx = context.WithValue(ctx, "event_store", eventStore)

// Record session start event
agentID := os.Getenv("HEX_AGENT_ID")
if agentID == "" {
	agentID = "root"
	_ = os.Setenv("HEX_AGENT_ID", agentID)
}

_ = eventStore.Record(events.Event{
	ID:        uuid.New().String(),
	AgentID:   agentID,
	Type:      "SessionStart",
	Timestamp: time.Now(),
	Data: map[string]interface{}{
		"prompt": userPrompt,
		"tools":  enabledTools,
	},
})
```

**Step 2: Add required imports**

Add to imports in `cmd/hex/root.go`:
```go
"github.com/2389-research/hex/internal/events"
"github.com/google/uuid"
```

**Step 3: Test manually**

Run: `./hex -p "Hello, test event recording"`
Expected: Creates file `/tmp/hex_events_*.jsonl` with SessionStart event

Verify: `cat /tmp/hex_events_*.jsonl`
Expected: See JSON line with SessionStart event

**Step 4: Commit**

```bash
git add cmd/hex/root.go
git commit -m "feat: wire event store to root command

- Initialize event store on session start
- Record SessionStart event with prompt and tools
- Store event file in /tmp with timestamp"
```

---

## Task 2: Enable Cost Tracking in Agent Lifecycle

**Files:**
- Modify: `cmd/hex/root.go:140-160`
- Modify: `internal/core/client.go:45-60` (add usage tracking hook)
- Test: Manual verification with cost output

**Step 1: Add cost tracker initialization**

Location: `cmd/hex/root.go` after event store init

```go
// Initialize cost tracker
costTracker := cost.NewCostTracker()

// Store in context
ctx = context.WithValue(ctx, "cost_tracker", costTracker)

// Defer cost summary on exit
defer func() {
	summary := costTracker.GetSummary(agentID)
	if summary.TotalCost > 0 {
		fmt.Fprintf(os.Stderr, "\n💰 Session Cost: $%.4f\n", summary.TotalCost)
		fmt.Fprintf(os.Stderr, "   Input tokens: %d, Output tokens: %d\n",
			summary.InputTokens, summary.OutputTokens)
	}
}()
```

**Step 2: Add imports**

Add to `cmd/hex/root.go`:
```go
"github.com/2389-research/hex/internal/cost"
```

**Step 3: Hook cost tracking into API client**

Location: `internal/core/client.go` in `CreateMessage` method after API response

```go
// Track usage for cost calculation
if costTracker := ctx.Value("cost_tracker"); costTracker != nil {
	if tracker, ok := costTracker.(*cost.CostTracker); ok {
		agentID := os.Getenv("HEX_AGENT_ID")
		if agentID == "" {
			agentID = "root"
		}

		_ = tracker.RecordUsage(agentID, "", req.Model, cost.Usage{
			InputTokens:  int64(resp.Usage.InputTokens),
			OutputTokens: int64(resp.Usage.OutputTokens),
			CacheReads:   int64(resp.Usage.CacheReadInputTokens),
			CacheWrites:  int64(resp.Usage.CacheCreationInputTokens),
		})
	}
}
```

**Step 4: Test manually**

Run: `./hex -p "What is 2+2?"`
Expected: See cost summary at end like "💰 Session Cost: $0.0012"

**Step 5: Commit**

```bash
git add cmd/hex/root.go internal/core/client.go
git commit -m "feat: enable cost tracking in agent lifecycle

- Initialize cost tracker on session start
- Hook into API client to record usage
- Display cost summary on exit"
```

---

## Task 3: Integrate Process Registry with Agent Cleanup

**Files:**
- Modify: `cmd/hex/root.go:100-120` (signal handling)
- Modify: `internal/tools/task_tool.go:150-170` (already has registration, verify cleanup)
- Test: Manual test with Ctrl+C

**Step 1: Add graceful shutdown handler**

Location: `cmd/hex/root.go` early in root command (before main execution)

```go
// Set up signal handling for graceful shutdown
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

ctx, cancel := context.WithCancel(context.Background())
defer cancel()

go func() {
	<-sigChan
	fmt.Fprintln(os.Stderr, "\n🛑 Interrupt received, shutting down gracefully...")

	// Stop all child processes
	agentID := os.Getenv("HEX_AGENT_ID")
	if agentID == "" {
		agentID = "root"
	}

	if err := registry.Global().StopCascading(agentID, nil); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: cleanup error: %v\n", err)
	}

	cancel()
	os.Exit(0)
}()
```

**Step 2: Add imports**

Add to `cmd/hex/root.go`:
```go
"os/signal"
"syscall"
"github.com/2389-research/hex/internal/registry"
```

**Step 3: Test graceful shutdown**

Run: `./hex -p "Run 'sleep 60'"`
While running: Press Ctrl+C
Expected: See "🛑 Interrupt received, shutting down gracefully..." and clean exit

**Step 4: Commit**

```bash
git add cmd/hex/root.go
git commit -m "feat: integrate process registry with graceful shutdown

- Add signal handler for SIGINT/SIGTERM
- Cascade stop to all child processes on interrupt
- Clean exit with proper cleanup"
```

---

## Task 4: Add Visualization Commands to CLI

**Files:**
- Create: `cmd/hex/visualize.go` (new subcommand)
- Modify: `cmd/hex/root.go:50` (register subcommand)
- Test: `go run cmd/hex/main.go visualize --help`

**Step 1: Create visualize subcommand**

Create file: `cmd/hex/visualize.go`

```go
// ABOUTME: Visualization subcommand for multi-agent execution analysis
// ABOUTME: Provides access to hexviz and hexreplay tools

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var visualizeCmd = &cobra.Command{
	Use:   "visualize [event-file]",
	Short: "Visualize multi-agent execution from event file",
	Long: `Visualize multi-agent execution history using the hexviz tool.

By default shows tree view. Use --view flag to select:
  tree     - Hierarchical agent structure (default)
  timeline - Chronological event sequence
  cost     - Cost breakdown by agent

Examples:
  hex visualize /tmp/hex_events_20251208.jsonl
  hex visualize events.jsonl --view timeline
  hex visualize events.jsonl --view cost --agent root.1`,
	Args: cobra.ExactArgs(1),
	RunE: runVisualize,
}

var (
	viewMode    string
	agentFilter string
	typeFilter  string
	htmlOutput  string
)

func init() {
	visualizeCmd.Flags().StringVar(&viewMode, "view", "tree", "view mode: tree, timeline, or cost")
	visualizeCmd.Flags().StringVar(&agentFilter, "agent", "", "filter by agent ID")
	visualizeCmd.Flags().StringVar(&typeFilter, "type", "", "filter by event type")
	visualizeCmd.Flags().StringVar(&htmlOutput, "html", "", "export to HTML file")
}

func runVisualize(cmd *cobra.Command, args []string) error {
	eventFile := args[0]

	// Check if event file exists
	if _, err := os.Stat(eventFile); os.IsNotExist(err) {
		return fmt.Errorf("event file not found: %s", eventFile)
	}

	// Find hexviz binary
	hexvizPath, err := exec.LookPath("hexviz")
	if err != nil {
		// Try in bin/ directory
		binPath := filepath.Join("bin", "hexviz")
		if _, err := os.Stat(binPath); err == nil {
			hexvizPath = binPath
		} else {
			return fmt.Errorf("hexviz not found. Run: go build -o bin/hexviz ./cmd/hexviz")
		}
	}

	// Build command arguments
	cmdArgs := []string{
		"-events", eventFile,
		"-view", viewMode,
	}

	if agentFilter != "" {
		cmdArgs = append(cmdArgs, "-agent", agentFilter)
	}

	if typeFilter != "" {
		cmdArgs = append(cmdArgs, "-type", typeFilter)
	}

	if htmlOutput != "" {
		cmdArgs = append(cmdArgs, "-html", htmlOutput)
	}

	// Execute hexviz
	vizCmd := exec.Command(hexvizPath, cmdArgs...)
	vizCmd.Stdout = os.Stdout
	vizCmd.Stderr = os.Stderr

	return vizCmd.Run()
}
```

**Step 2: Register subcommand**

Location: `cmd/hex/root.go` in `init()` function

```go
func init() {
	// ... existing flags ...

	// Register subcommands
	rootCmd.AddCommand(visualizeCmd)
	rootCmd.AddCommand(replayCmd) // We'll add this next
}
```

**Step 3: Create replay subcommand**

Create file: `cmd/hex/replay.go`

```go
// ABOUTME: Replay subcommand for event timeline viewing
// ABOUTME: Provides access to hexreplay tool

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var replayCmd = &cobra.Command{
	Use:   "replay [event-file]",
	Short: "Replay event timeline from event file",
	Long: `Display chronological timeline of agent events.

Examples:
  hex replay /tmp/hex_events_20251208.jsonl
  hex replay events.jsonl --agent root.1
  hex replay events.jsonl --type ToolCall`,
	Args: cobra.ExactArgs(1),
	RunE: runReplay,
}

var (
	replayAgentFilter string
	replayTypeFilter  string
)

func init() {
	replayCmd.Flags().StringVar(&replayAgentFilter, "agent", "", "filter by agent ID")
	replayCmd.Flags().StringVar(&replayTypeFilter, "type", "", "filter by event type")
}

func runReplay(cmd *cobra.Command, args []string) error {
	eventFile := args[0]

	// Check if event file exists
	if _, err := os.Stat(eventFile); os.IsNotExist(err) {
		return fmt.Errorf("event file not found: %s", eventFile)
	}

	// Find hexreplay binary
	hexreplayPath, err := exec.LookPath("hexreplay")
	if err != nil {
		// Try in bin/ directory
		binPath := filepath.Join("bin", "hexreplay")
		if _, err := os.Stat(binPath); err == nil {
			hexreplayPath = binPath
		} else {
			return fmt.Errorf("hexreplay not found. Run: go build -o bin/hexreplay ./cmd/hexreplay")
		}
	}

	// Build command arguments
	cmdArgs := []string{"-events", eventFile}

	if replayAgentFilter != "" {
		cmdArgs = append(cmdArgs, "-agent", replayAgentFilter)
	}

	if replayTypeFilter != "" {
		cmdArgs = append(cmdArgs, "-type", replayTypeFilter)
	}

	// Execute hexreplay
	replayExec := exec.Command(hexreplayPath, cmdArgs...)
	replayExec.Stdout = os.Stdout
	replayExec.Stderr = os.Stderr

	return replayExec.Run()
}
```

**Step 4: Test commands**

Build tools:
```bash
go build -o bin/hexviz ./cmd/hexviz
go build -o bin/hexreplay ./cmd/hexreplay
```

Test help:
```bash
go run cmd/hex/main.go visualize --help
go run cmd/hex/main.go replay --help
```

Expected: See help text with all options

**Step 5: Commit**

```bash
git add cmd/hex/visualize.go cmd/hex/replay.go cmd/hex/root.go
git commit -m "feat: add visualization commands to CLI

- Add 'hex visualize' subcommand for hexviz
- Add 'hex replay' subcommand for hexreplay
- Support all view modes and filtering options"
```

---

## Task 5: Add Event File Location to Session Output

**Files:**
- Modify: `cmd/hex/root.go:145-150` (add output after event store init)
- Test: Run hex and verify message

**Step 1: Print event file location**

Location: `cmd/hex/root.go` after event store initialization

```go
// Initialize event store
eventFile := filepath.Join(os.TempDir(), fmt.Sprintf("hex_events_%s.jsonl", time.Now().Format("20060102_150405")))
eventStore, err := events.NewEventStore(eventFile)
if err != nil {
	return fmt.Errorf("failed to create event store: %w", err)
}
defer func() { _ = eventStore.Close() }()

// Inform user about event recording
if os.Getenv("HEX_DEBUG") != "" {
	fmt.Fprintf(os.Stderr, "📊 Recording events to: %s\n", eventFile)
	fmt.Fprintf(os.Stderr, "   Replay: hex replay %s\n", eventFile)
	fmt.Fprintf(os.Stderr, "   Visualize: hex visualize %s\n", eventFile)
}
```

**Step 2: Test with debug mode**

Run: `HEX_DEBUG=1 ./hex -p "Hello"`
Expected: See event file location and usage commands

**Step 3: Commit**

```bash
git add cmd/hex/root.go
git commit -m "feat: show event file location in debug mode

- Print event file path when HEX_DEBUG=1
- Show replay and visualize command examples
- Help users discover visualization features"
```

---

## Task 6: End-to-End Integration Test

**Files:**
- Create: `test/integration/multiagent_integration_test.go`
- Test: `go test ./test/integration -v -run TestMultiAgentIntegration`

**Step 1: Write integration test**

Create file: `test/integration/multiagent_integration_test.go`

```go
// ABOUTME: Integration tests for multi-agent feature integration
// ABOUTME: Tests event-sourcing, cost tracking, and process registry together

package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/2389-research/hex/internal/cost"
	"github.com/2389-research/hex/internal/events"
	"github.com/2389-research/hex/internal/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMultiAgentIntegration(t *testing.T) {
	// Set up test environment
	tmpDir := t.TempDir()
	eventFile := filepath.Join(tmpDir, "test_events.jsonl")

	// Initialize event store
	eventStore, err := events.NewEventStore(eventFile)
	require.NoError(t, err)
	defer func() { _ = eventStore.Close() }()

	// Initialize cost tracker
	costTracker := cost.NewCostTracker()

	// Set agent ID
	_ = os.Setenv("HEX_AGENT_ID", "test-root")
	defer func() { _ = os.Unsetenv("HEX_AGENT_ID") }()

	// Record test events
	ctx := context.Background()

	startEvent := events.Event{
		ID:        "evt-1",
		AgentID:   "test-root",
		Type:      "SessionStart",
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"test": true},
	}
	err = eventStore.Record(startEvent)
	require.NoError(t, err)

	// Record cost
	err = costTracker.RecordUsage("test-root", "", "claude-sonnet-4-5-20250929", cost.Usage{
		InputTokens:  1000,
		OutputTokens: 500,
	})
	require.NoError(t, err)

	// Close event store
	err = eventStore.Close()
	require.NoError(t, err)

	// Verify event file exists
	_, err = os.Stat(eventFile)
	require.NoError(t, err)

	// Reload and verify events
	eventStore2, err := events.NewEventStore(eventFile)
	require.NoError(t, err)
	defer func() { _ = eventStore2.Close() }()

	loadedEvents := eventStore2.GetEvents("test-root")
	assert.Len(t, loadedEvents, 1)
	assert.Equal(t, "SessionStart", loadedEvents[0].Type)

	// Verify cost tracking
	summary := costTracker.GetSummary("test-root")
	assert.Equal(t, int64(1000), summary.InputTokens)
	assert.Equal(t, int64(500), summary.OutputTokens)
	assert.Greater(t, summary.TotalCost, 0.0)

	// Verify process registry (basic check)
	reg := registry.NewProcessRegistry()
	err = reg.Register("test-child", "test-root", &os.Process{Pid: os.Getpid()})
	require.NoError(t, err)

	process := reg.Get("test-child")
	assert.NotNil(t, process)
	assert.Equal(t, "test-root", process.ParentID)
}

func TestEventStoreAndCostTrackerIntegration(t *testing.T) {
	tmpDir := t.TempDir()
	eventFile := filepath.Join(tmpDir, "events.jsonl")

	eventStore, err := events.NewEventStore(eventFile)
	require.NoError(t, err)

	costTracker := cost.NewCostTracker()

	// Simulate multi-agent scenario
	agents := []string{"root", "root.1", "root.2"}

	for _, agentID := range agents {
		// Record event
		err = eventStore.Record(events.Event{
			ID:        agentID + "-start",
			AgentID:   agentID,
			Type:      "AgentStart",
			Timestamp: time.Now(),
		})
		require.NoError(t, err)

		// Record cost
		err = costTracker.RecordUsage(agentID, "", "claude-sonnet-4-5-20250929", cost.Usage{
			InputTokens:  500,
			OutputTokens: 200,
		})
		require.NoError(t, err)
	}

	_ = eventStore.Close()

	// Verify all events recorded
	eventStore2, err := events.NewEventStore(eventFile)
	require.NoError(t, err)
	defer func() { _ = eventStore2.Close() }()

	allEvents := eventStore2.GetEvents("")
	assert.Len(t, allEvents, 3)

	// Verify total cost
	totalCost := 0.0
	for _, agentID := range agents {
		summary := costTracker.GetSummary(agentID)
		totalCost += summary.TotalCost
	}
	assert.Greater(t, totalCost, 0.0)
}
```

**Step 2: Run integration tests**

Run: `go test ./test/integration -v -run TestMultiAgent`
Expected: All tests pass

**Step 3: Commit**

```bash
git add test/integration/multiagent_integration_test.go
git commit -m "test: add multi-agent integration tests

- Test event store + cost tracker integration
- Test multi-agent scenario with 3 agents
- Verify event persistence and cost calculation"
```

---

## Task 7: Update Documentation

**Files:**
- Create: `docs/MULTIAGENT_FEATURES.md` (user guide)
- Modify: `README.md` (add features section)

**Step 1: Create user guide**

Create file: `docs/MULTIAGENT_FEATURES.md`

```markdown
# Multi-Agent Features Guide

Hex now includes comprehensive multi-agent capabilities for tracking, analyzing, and visualizing agent execution.

## Features Overview

### 📊 Event-Sourcing

All agent activities are recorded to an event file for complete audit trails.

**Automatic Recording:**
- Session start/end
- Tool calls and results
- Agent spawn/terminate
- Errors and warnings

**Event File Location:**
```bash
# Enable debug mode to see event file location
HEX_DEBUG=1 hex -p "Your prompt"
# Output: 📊 Recording events to: /tmp/hex_events_20251208_143022.jsonl
```

### 💰 Cost Tracking

Automatic cost calculation based on Anthropic API pricing.

**Session Summary:**
```bash
hex -p "Analyze this codebase"
# ... execution ...
# 💰 Session Cost: $0.1234
#    Input tokens: 12500, Output tokens: 3200
```

**Per-Agent Breakdown:**
Use `hex visualize` with cost view to see cost by agent.

### 🔍 Visualization

**Tree View** - See agent hierarchy:
```bash
hex visualize /tmp/hex_events_*.jsonl --view tree
```

Output:
```
root
├── Task: "Analyze codebase"
├── Cost: $0.09
└── Duration: 2m 15s
  ├─ root.1
  │ ├── Task: "Search imports"
  │ ├── Cost: $0.04
  │ └── Duration: 45s
```

**Timeline View** - Chronological events:
```bash
hex visualize /tmp/hex_events_*.jsonl --view timeline
```

**Cost View** - Financial breakdown:
```bash
hex visualize /tmp/hex_events_*.jsonl --view cost
```

**HTML Export** - Interactive visualization:
```bash
hex visualize events.jsonl --html report.html
```

### 🔄 Event Replay

View event timeline:
```bash
hex replay /tmp/hex_events_*.jsonl
```

**Filtering:**
```bash
# Only events from specific agent
hex replay events.jsonl --agent root.1

# Only specific event types
hex replay events.jsonl --type ToolCall
```

### 🛡️ Process Management

**Graceful Shutdown:**
Press Ctrl+C to cleanly stop all child agents:
```
🛑 Interrupt received, shutting down gracefully...
```

All spawned processes are tracked and cleaned up properly.

## Environment Variables

- `HEX_DEBUG=1` - Show event file location and debug info
- `HEX_AGENT_ID` - Current agent identifier (auto-set)
- `HEX_MAX_AGENT_DEPTH` - Maximum agent recursion depth (default: 5)

## Examples

### Basic Session with Visualization

```bash
# Run a task
HEX_DEBUG=1 hex -p "Refactor this code using subagents"

# Note the event file location from output
# Example: /tmp/hex_events_20251208_143022.jsonl

# Visualize the execution tree
hex visualize /tmp/hex_events_20251208_143022.jsonl

# See the cost breakdown
hex visualize /tmp/hex_events_20251208_143022.jsonl --view cost

# Export to HTML
hex visualize /tmp/hex_events_20251208_143022.jsonl --html execution.html
```

### Analyzing Multi-Agent Performance

```bash
# Run complex task with multiple agents
hex -p "Analyze codebase, run tests, and create documentation"

# View event timeline
hex replay /tmp/hex_events_*.jsonl

# Filter to specific agent
hex visualize /tmp/hex_events_*.jsonl --agent root.2

# Check cost by agent
hex visualize /tmp/hex_events_*.jsonl --view cost
```

## Architecture

All features integrate seamlessly:

1. **Event Store** - Records all events to JSON Lines file
2. **Cost Tracker** - Calculates costs from API usage
3. **Process Registry** - Tracks parent-child relationships
4. **Visualization** - Renders events in multiple views

Events are recorded automatically without performance impact.
```

**Step 2: Update README**

Add to `README.md` in features section:

```markdown
### Multi-Agent Capabilities

- 📊 **Event-Sourcing**: Complete audit trail of all agent activities
- 💰 **Cost Tracking**: Automatic cost calculation per agent and tree totals
- 🔍 **Visualization**: Tree, timeline, and cost views with HTML export
- 🛡️ **Process Management**: Graceful shutdown with cascading cleanup
- 📈 **Analytics**: Analyze multi-agent performance and costs

See [Multi-Agent Features Guide](docs/MULTIAGENT_FEATURES.md) for details.
```

**Step 3: Commit**

```bash
git add docs/MULTIAGENT_FEATURES.md README.md
git commit -m "docs: add multi-agent features user guide

- Comprehensive guide for all new features
- Examples for visualization and replay
- Environment variables reference
- Update README with feature highlights"
```

---

## Task 8: Build and Install Visualization Tools

**Files:**
- Modify: `Makefile` (add build targets)
- Test: `make install-viz`

**Step 1: Add Makefile targets**

Add to `Makefile`:

```makefile
# Build visualization tools
.PHONY: build-viz
build-viz:
	@echo "Building visualization tools..."
	go build -o bin/hexviz ./cmd/hexviz
	go build -o bin/hexreplay ./cmd/hexreplay
	@echo "✅ Built bin/hexviz and bin/hexreplay"

# Install visualization tools to PATH
.PHONY: install-viz
install-viz: build-viz
	@echo "Installing visualization tools..."
	cp bin/hexviz $(GOPATH)/bin/hexviz
	cp bin/hexreplay $(GOPATH)/bin/hexreplay
	@echo "✅ Installed hexviz and hexreplay to $(GOPATH)/bin"

# Clean visualization binaries
.PHONY: clean-viz
clean-viz:
	rm -f bin/hexviz bin/hexreplay
	@echo "✅ Cleaned visualization binaries"
```

**Step 2: Update main build target**

Modify existing `build` target in Makefile:

```makefile
.PHONY: build
build: build-viz
	@echo "Building hex..."
	go build -o bin/hex cmd/hex/main.go
	@echo "✅ Built bin/hex"
```

**Step 3: Test build**

Run: `make build-viz`
Expected: Creates `bin/hexviz` and `bin/hexreplay`

Verify: `ls -lh bin/`
Expected: See both binaries

**Step 4: Commit**

```bash
git add Makefile
git commit -m "build: add visualization tools to build process

- Add build-viz target for hexviz and hexreplay
- Add install-viz to install to GOPATH/bin
- Include visualization builds in main build target"
```

---

## Verification Steps

After completing all tasks, verify the integration:

**1. Full Session Test:**
```bash
# Build everything
make build

# Run with debug
HEX_DEBUG=1 ./bin/hex -p "Run 'echo hello' and explain what happened"

# Verify output shows:
# - 📊 Recording events to: /tmp/hex_events_*.jsonl
# - 💰 Session Cost: $0.00XX
```

**2. Visualization Test:**
```bash
# Get event file from previous test
EVENT_FILE=$(ls -t /tmp/hex_events_*.jsonl | head -1)

# Test all views
./bin/hexviz -events $EVENT_FILE -view tree
./bin/hexviz -events $EVENT_FILE -view timeline
./bin/hexviz -events $EVENT_FILE -view cost

# Test replay
./bin/hexreplay -events $EVENT_FILE
```

**3. Graceful Shutdown Test:**
```bash
# Start long-running task
./bin/hex -p "Run 'sleep 60'"

# Press Ctrl+C
# Expected: "🛑 Interrupt received, shutting down gracefully..."
```

**4. Integration Test:**
```bash
go test ./test/integration -v
# Expected: All tests pass
```

**5. Build Test:**
```bash
make clean
make build
# Expected: Builds hex, hexviz, hexreplay without errors
```

---

## Success Criteria

- ✅ Event store records all sessions automatically
- ✅ Cost tracking shows accurate totals on exit
- ✅ Graceful shutdown stops all child processes
- ✅ `hex visualize` command works with all view modes
- ✅ `hex replay` command displays event timeline
- ✅ HTML export creates valid standalone files
- ✅ All integration tests pass
- ✅ Documentation is complete and accurate
- ✅ Build process includes visualization tools

---

## Notes for Engineer

- **TDD Approach**: Each task includes test steps before commit
- **Incremental**: Each task is independently testable
- **DRY**: Reuse existing packages (events, cost, registry)
- **YAGNI**: No over-engineering, just wire up what exists
- **Frequent Commits**: Commit after each task completion

**Estimated Time:** 2-3 hours for all tasks

**Dependencies:** All required packages already implemented in Phases 1-3
