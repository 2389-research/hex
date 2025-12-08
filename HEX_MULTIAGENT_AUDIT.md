# Hex Multi-Agent System Audit

**Date:** 2025-12-07
**Auditor:** Claude (using building-multiagent-systems skill)
**Subject:** Hex - AI-Powered Terminal Assistant

---

## Executive Summary

Hex implements a **hierarchical multi-agent system** with an interactive orchestrator that spawns specialized subagents for complex tasks. The architecture demonstrates strong adherence to multi-agent best practices with the **Task tool** serving as the primary coordination mechanism. However, there are opportunities to strengthen the four-layer architecture and improve agent lifecycle management.

**Overall Grade: B+ (Strong, with room for improvement)**

---

## 1. Discovery Questions Analysis

### Q1: Starting Point
**Current State:** Mature system with established agent architecture
**Assessment:** ✅ Clear evolution path from single-agent to multi-agent

### Q2: Primary Use Case
**Pattern:** Recursive delegation + fan-out for parallel tasks
**Implementation:** Task tool spawns subagents (Explore, general-purpose, code-reviewer)
**Assessment:** ✅ Appropriate pattern for development workflows

### Q3: Scale Expectations
**Current:** Small scale (1-5 concurrent subagents)
**Design capacity:** Medium scale capable (parent + ~10 children)
**Assessment:** ✅ Scaled appropriately for CLI tool use case

### Q4: State Requirements
**Implementation:** Session-based with SQLite persistence
**Conversation continuity:** ✅ Messages + context survive restarts
**Agent state:** ⚠️ Subagent state NOT persisted across sessions
**Assessment:** ⚠️ Partial - main agent state survives, subagent state ephemeral

### Q5: Tool Coordination
**Pattern:** Shared read-only tool registry + permission inheritance
**File access:** Multiple agents can read simultaneously
**Write coordination:** ⚠️ No explicit file locking detected
**Assessment:** ⚠️ Race conditions possible with concurrent file writes

### Q6: Existing Constraints
**Language:** Go
**Framework:** Custom (no framework dependency)
**Performance:** Excellent (native binary, <10MB)
**Compliance:** Local-first (privacy-preserving)
**Assessment:** ✅ Strong architectural choices

---

## 2. Four-Layer Architecture Analysis

### Current Architecture

```
┌─────────────────────────────────────────┐
│  Layer 1: Reasoning (Claude API)        │  ✅ Clear LLM boundary
├─────────────────────────────────────────┤
│  Layer 2: Orchestration (Model.Update)  │  ⚠️ Mixed concerns
├─────────────────────────────────────────┤
│  Layer 3: Tool Bus (tools.Executor)     │  ✅ Schema validation
├─────────────────────────────────────────┤
│  Layer 4: Adapters (Read/Write/Bash)    │  ✅ Deterministic
└─────────────────────────────────────────┘
```

### Layer 1: Reasoning ✅
**Implementation:** `internal/core/client.go` - Claude API calls
**Strengths:**
- Clean API boundary
- No LLM calls leak into lower layers
- Proper streaming support

**Weaknesses:**
- None identified

**Grade: A**

### Layer 2: Orchestration ⚠️
**Implementation:** `internal/ui/update.go`, `internal/subagents/`
**Strengths:**
- Task tool spawns subagents correctly
- Permission model exists (`approval.Rules`)
- Tool approval workflow (safety-first mode)

**Weaknesses:**
- **Mixed concerns:** TUI logic + orchestration in same file (update.go)
- **No clear orchestrator abstraction:** Orchestration scattered across UI model
- **Subagent lifecycle unclear:** No explicit spawn/stop/cleanup protocol
- **State machine not formalized:** Agent states tracked but not enforced

**Issues Found:**

1. **Orchestration + UI coupling** (`internal/ui/update.go:1-1400`)
   - Problem: 1400-line Update() function mixes TUI rendering with agent coordination
   - Impact: Hard to test orchestration logic independently
   - Recommendation: Extract `AgentOrchestrator` separate from UI

2. **No cascading stop protocol**
   - Problem: When main agent stops, no code ensures subagents stop first
   - Searched for: "StopAllChildren", "cascading stop"
   - Found: None
   - Impact: Potential orphaned subagent processes

3. **Subagent spawn without timeout**
   - File: `internal/subagents/subagent.go`
   - Problem: Task tool has no max execution time limit
   - Impact: Infinite-running subagent could block parent

**Grade: C+**

### Layer 3: Tool Bus ✅
**Implementation:** `internal/tools/executor.go`, `internal/tools/registry.go`
**Strengths:**
- Schema-first design (`registry.go:159-220` - JSON schemas)
- Tool validation before execution
- Permission checking (`executor.go:94`)
- Result caching (`cache.go`) - performance optimization
- Clean tool interface

**Weaknesses:**
- No explicit rate limiting across agents
- File locking not implemented (concurrent write risk)

**Grade: A-**

### Layer 4: Adapters (Deterministic Tools) ✅
**Implementation:** `internal/tools/*_tool.go`
**Checked for LLM calls in tools:**
- `read_tool.go` ✅ Pure file I/O
- `write_tool.go` ✅ Pure file I/O
- `edit_tool.go` ✅ String replacement only
- `bash_tool.go` ✅ Shell execution only
- `grep_tool.go` ✅ Ripgrep wrapper
- `task_tool.go` ⚠️ **Spawns subagent with LLM** (acceptable - this IS the coordination tool)

**Strengths:**
- All adapters are deterministic
- No LLM calls inside tools (except Task, which is meta)
- Proper error handling
- Timeout support in Bash tool

**Grade: A**

---

## 3. Coordination Pattern Analysis

### Primary Pattern: Recursive Delegation

**Implementation:** Task tool (`internal/subagents/`)

```go
// Task tool spawns specialized subagents
subagent := &Subagent{
    Type:        subagentType,    // Explore, general-purpose, etc.
    Prompt:      prompt,
    Model:       model,
    WorkingDir:  workingDir,
}

result, err := subagent.Run(ctx)
```

**Strengths:**
- ✅ Hierarchical thread IDs (`task-123.1.2` pattern) NOT FOUND but should be added
- ✅ Subagent type specialization (Explore vs general-purpose)
- ✅ Context propagation (working directory, model selection)

**Weaknesses:**
- ❌ **No max depth limit** - recursive delegation could infinite loop
- ❌ **No hierarchical cost tracking** - can't aggregate subagent costs
- ⚠️ **Event-sourcing incomplete** - conversation persisted, but not agent spawn/stop events

### Secondary Pattern: Tool Inheritance

**Implementation:** `internal/permissions/` + Task tool

**Current State:**
```go
// Subagent gets tools via --tools flag
// Example: ./hex task "..." --tools=read_file,grep
```

**Assessment:**
- ✅ Subagents can have tool subsets
- ⚠️ Permission inheritance model unclear
- ❌ Subagents cannot escalate privileges (good!) but not formally enforced

**Recommendation:** Formalize permission inheritance:
```go
type AgentPermissions struct {
    Parent      *AgentPermissions  // Cannot exceed parent
    AllowedTools []string
    FileRead    bool
    FileWrite   bool
    ShellExec   bool
}

func (p *AgentPermissions) CanEscalate() bool {
    if p.Parent == nil {
        return true  // Root agent
    }
    // Check each permission against parent
    return p.FileWrite <= p.Parent.FileWrite &&
           p.ShellExec <= p.Parent.ShellExec
}
```

---

## 4. Agent Lifecycle Management

### Current Lifecycle

```
Created → Running → (Streaming) → Completed → Destroyed
```

**Implementation:**
- Created: Subagent struct initialized
- Running: LLM stream starts
- Completed: Result returned
- Destroyed: ⚠️ **No explicit cleanup detected**

### Issues Found

#### 1. No State Machine Enforcement
**Problem:** Agent states tracked informally
**File:** `internal/ui/model.go:114` - `Streaming bool`
**Issue:** Boolean flags instead of formal state machine

**Recommendation:**
```go
type AgentState int

const (
    StateIdle AgentState = iota
    StateThinking
    StateStreaming
    StateToolExecution
    StateStopped
)

// Enforce valid transitions
func (a *Agent) Transition(to AgentState) error {
    validTransitions := map[AgentState][]AgentState{
        StateIdle:          {StateThinking, StateStopped},
        StateThinking:      {StateStreaming, StateToolExecution, StateStopped},
        StateStreaming:     {StateToolExecution, StateIdle, StateStopped},
        StateToolExecution: {StateIdle, StateThinking, StateStopped},
        StateStopped:       {}, // Terminal state
    }

    if !contains(validTransitions[a.state], to) {
        return fmt.Errorf("invalid transition: %v → %v", a.state, to)
    }
    a.state = to
    return nil
}
```

#### 2. Missing Cascading Stop

**Current code analysis:**
- Searched: `internal/subagents/subagent.go`
- Found: `Run()` returns result, no `Stop()` method
- Searched: Goroutine cleanup
- Found: Context cancellation in `streamMessage()` ✅

**Problem:** When parent agent stops, subagents continue until completion

**Recommendation:**
```go
type Orchestrator struct {
    mu       sync.Mutex
    children []*Subagent
}

func (o *Orchestrator) SpawnSubAgent(config SubagentConfig) (*Subagent, error) {
    o.mu.Lock()
    defer o.mu.Unlock()

    agent := NewSubagent(config)
    o.children = append(o.children, agent)
    return agent, nil
}

func (o *Orchestrator) Stop() error {
    o.mu.Lock()
    defer o.mu.Unlock()

    // 1. Stop all children first
    var wg sync.WaitGroup
    for _, child := range o.children {
        wg.Add(1)
        go func(a *Subagent) {
            defer wg.Done()
            a.Stop()
        }(child)
    }
    wg.Wait()

    // 2. Now stop self
    o.cancel()  // Cancel context
    return nil
}
```

#### 3. No Orphan Detection

**Missing:** Heartbeat monitoring for orphaned agents

**Recommendation:**
```go
// Add to main agent
type Heartbeat struct {
    agentID   string
    parentID  string
    lastBeat  time.Time
}

func OrphanDetector(ctx context.Context) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            agents := GetRunningAgents()
            for _, agent := range agents {
                if agent.parentID != "" {
                    parent := GetAgent(agent.parentID)
                    if parent == nil || parent.IsStopped() {
                        log.Printf("Orphaned agent detected: %s", agent.ID)
                        agent.Stop()
                    }
                }
            }
        case <-ctx.Done():
            return
        }
    }
}
```

---

## 5. Tool Coordination Issues

### Shared Resource Locking (Missing)

**Risk:** Multiple agents editing same file simultaneously

**Current Implementation:**
- File: `internal/tools/edit_tool.go`
- No locks detected
- Race condition possible:
  ```
  Agent A: Read file.go
  Agent B: Read file.go
  Agent A: Edit line 10
  Agent B: Edit line 10
  Agent A: Write file.go
  Agent B: Write file.go  ← OVERWRITES A's changes!
  ```

**Recommendation:**
```go
type FileLockManager struct {
    locks sync.Map  // map[string]*sync.RWMutex
}

func (m *FileLockManager) AcquireWrite(path string) (unlock func(), err error) {
    lockVal, _ := m.locks.LoadOrStore(path, &sync.RWMutex{})
    lock := lockVal.(*sync.RWMutex)

    lock.Lock()
    return func() { lock.Unlock() }, nil
}

// In edit_tool.go Execute():
unlock, err := lockManager.AcquireWrite(params.FilePath)
if err != nil {
    return "", err
}
defer unlock()

// Now safe to edit
```

### Rate Limiting (Missing)

**Risk:** Multiple agents hitting Claude API simultaneously → rate limit errors

**Current:** Each agent makes API calls independently
**Issue:** No coordination across agent tree

**Recommendation:**
```go
type TokenBucket struct {
    tokens     int
    capacity   int
    refillRate time.Duration
    mu         sync.Mutex
}

func (b *TokenBucket) Acquire(ctx context.Context) error {
    b.mu.Lock()
    defer b.mu.Unlock()

    for b.tokens == 0 {
        select {
        case <-time.After(b.refillRate):
            b.tokens = min(b.tokens+1, b.capacity)
        case <-ctx.Done():
            return ctx.Err()
        }
    }

    b.tokens--
    return nil
}

// Global rate limiter shared across all agents
var apiLimiter = &TokenBucket{
    tokens:     50,
    capacity:   50,
    refillRate: 100 * time.Millisecond,
}
```

---

## 6. State Management & Event-Sourcing

### Current Implementation

**Persisted:**
- ✅ Conversations (`internal/storage/`)
- ✅ Messages with content blocks
- ✅ Approval rules

**NOT Persisted:**
- ❌ Agent spawn events
- ❌ Tool execution events (for subagents)
- ❌ Cost breakdown per agent
- ❌ Subagent hierarchy

**Assessment:** Partial event-sourcing - conversation-level only

### Recommendation: Complete Event-Sourcing

```go
type AgentEvent struct {
    ID         string
    Type       EventType  // AgentSpawned, ToolExecuted, AgentStopped
    Timestamp  time.Time
    AgentID    string
    ParentID   string
    Data       json.RawMessage
}

type EventType int

const (
    EventAgentSpawned EventType = iota
    EventToolExecuted
    EventAgentStopped
    EventCostIncurred
)

// Benefits:
// 1. Replay agent execution for debugging
// 2. Cost attribution to specific agents
// 3. Audit trail of all agent actions
// 4. Time-travel debugging
```

---

## 7. Production Hardening Gaps

### ✅ Strengths

1. **Timeout handling** - Bash tool has timeout support
2. **Context cancellation** - Stream cancellation prevents goroutine leaks (post-refactor)
3. **Error propagation** - Errors bubble up correctly
4. **Logging** - Structured logging with slog

### ❌ Missing

1. **Orphan detection** - No periodic scan for abandoned agents
2. **Resource limits** - No max concurrent subagents limit
3. **Cost caps** - No per-agent or per-session cost limits
4. **Cascading cleanup** - Parent doesn't stop children first
5. **Hierarchical cost tracking** - Can't see cost per subagent tree

### ⚠️ Needs Improvement

1. **Checkpointing** - Conversation persisted, but not mid-execution state
2. **Partial failure handling** - If subagent fails, what happens to siblings?

---

## 8. Comparison to Multi-Agent Patterns

### Fan-Out/Fan-In: Not Implemented
**Use case:** Parallel code review (security + performance + style agents)
**Status:** Could be added for multi-file analysis
**Priority:** Low (not current use case)

### Sequential Pipeline: Not Implemented
**Use case:** Multi-stage code transformations
**Status:** Could be added for refactoring workflows
**Priority:** Medium

### Recursive Delegation: ✅ Implemented
**Implementation:** Task tool → Explore/general-purpose subagents
**Quality:** Good, but needs max-depth limit
**Grade:** B+

### Work-Stealing Queue: Not Needed
**Use case:** Batch processing 1000+ items
**Status:** Not applicable for interactive CLI
**Priority:** N/A

### Map-Reduce: Partially Implemented
**Current:** Subagent uses Haiku (cheap), main uses Sonnet (smart)
**Quality:** Good cost optimization
**Grade:** A

### Peer Collaboration: Not Implemented
**Use case:** Multi-model consensus (GPT vs Claude vs Gemini)
**Status:** Could add for critical decisions
**Priority:** Low

### MAKER: Not Needed
**Use case:** Zero-error tolerance, 100K+ steps
**Status:** Overkill for development workflows
**Priority:** N/A

---

## 9. Critical Findings Summary

### 🔴 Critical Issues

1. **No cascading stop protocol**
   - Impact: Orphaned subagent processes
   - File: `internal/subagents/subagent.go`
   - Fix: Add Stop() method, track children

2. **Race conditions on file writes**
   - Impact: Data corruption with concurrent edits
   - File: `internal/tools/edit_tool.go`
   - Fix: Add file locking

3. **No API rate limiting coordination**
   - Impact: Rate limit errors with multiple agents
   - File: `internal/core/client.go`
   - Fix: Shared token bucket

### 🟡 High Priority Issues

4. **Orchestration layer mixed with UI**
   - Impact: Hard to test, hard to extend
   - File: `internal/ui/update.go`
   - Fix: Extract AgentOrchestrator

5. **No max recursion depth for Task tool**
   - Impact: Infinite delegation loops possible
   - File: `internal/subagents/subagent.go`
   - Fix: Add depth tracking

6. **No hierarchical cost tracking**
   - Impact: Can't attribute costs to subagent trees
   - File: `internal/core/client.go`
   - Fix: Thread IDs + cost aggregation

### 🟢 Medium Priority Issues

7. **State machine not formalized**
   - Impact: Invalid state transitions possible
   - File: `internal/ui/model.go`
   - Fix: Explicit state enum + validation

8. **Event-sourcing incomplete**
   - Impact: Can't replay or debug agent interactions
   - File: `internal/storage/`
   - Fix: Add AgentEvent table

---

## 10. Recommendations Roadmap

### Phase 1: Critical Fixes (Week 1)
1. Implement cascading stop protocol
2. Add file locking for concurrent writes
3. Add shared API rate limiter

### Phase 2: Architecture (Week 2-3)
4. Extract AgentOrchestrator from UI layer
5. Formalize agent state machine
6. Add max recursion depth to Task tool

### Phase 3: Observability (Week 4)
7. Implement hierarchical cost tracking
8. Add complete event-sourcing
9. Build agent execution visualizer

### Phase 4: Advanced Features (Future)
10. Orphan detection with heartbeats
11. Resource limits (max concurrent agents)
12. Cost caps per session

---

## 11. Code Examples

### Example 1: Proper Cascading Stop

```go
// internal/orchestrator/orchestrator.go
type Orchestrator struct {
    mu       sync.RWMutex
    children map[string]*Subagent
    ctx      context.Context
    cancel   context.CancelFunc
}

func (o *Orchestrator) SpawnSubAgent(cfg SubagentConfig) (*Subagent, error) {
    o.mu.Lock()
    defer o.mu.Unlock()

    agent := NewSubagent(cfg)
    agent.parentCtx = o.ctx  // Inherit cancellation
    o.children[agent.ID] = agent
    return agent, nil
}

func (o *Orchestrator) Stop() error {
    o.mu.Lock()
    defer o.mu.Unlock()

    // 1. Get all children
    children := make([]*Subagent, 0, len(o.children))
    for _, child := range o.children {
        children = append(children, child)
    }

    // 2. Stop all children in parallel
    var wg sync.WaitGroup
    for _, child := range children {
        wg.Add(1)
        go func(a *Subagent) {
            defer wg.Done()
            _ = a.Stop()  // Best effort
        }(child)
    }
    wg.Wait()

    // 3. Stop self
    o.cancel()

    return nil
}
```

### Example 2: Hierarchical Thread IDs

```go
type AgentID string

func (id AgentID) Child(index int) AgentID {
    return AgentID(fmt.Sprintf("%s.%d", id, index))
}

func (id AgentID) Parent() (AgentID, bool) {
    parts := strings.Split(string(id), ".")
    if len(parts) == 1 {
        return "", false  // Root agent
    }
    return AgentID(strings.Join(parts[:len(parts)-1], ".")), true
}

// Usage:
parent := AgentID("session-abc")
child1 := parent.Child(1)  // "session-abc.1"
child2 := parent.Child(2)  // "session-abc.2"
grandchild := child1.Child(1)  // "session-abc.1.1"
```

### Example 3: File Locking in Edit Tool

```go
// internal/tools/file_locks.go
var GlobalFileLocks = NewFileLockManager()

type FileLockManager struct {
    locks sync.Map  // map[string]*sync.RWMutex
}

func (m *FileLockManager) Lock(path string) (unlock func()) {
    // Normalize path
    absPath, _ := filepath.Abs(path)

    // Get or create lock for this file
    lockVal, _ := m.locks.LoadOrStore(absPath, &sync.RWMutex{})
    lock := lockVal.(*sync.RWMutex)

    lock.Lock()
    return func() { lock.Unlock() }
}

// In edit_tool.go:
func (t *EditTool) Execute(ctx context.Context, params EditParams) (string, error) {
    unlock := GlobalFileLocks.Lock(params.FilePath)
    defer unlock()

    // Now safe to read-modify-write
    content, err := os.ReadFile(params.FilePath)
    // ... edit logic ...
    err = os.WriteFile(params.FilePath, modified, 0644)
    return "success", nil
}
```

---

## 12. Final Grade Breakdown

| Category | Grade | Reasoning |
|----------|-------|-----------|
| **Four-Layer Architecture** | B+ | Layers present but orchestration mixed with UI |
| **Coordination Patterns** | A- | Recursive delegation well-implemented |
| **Tool Coordination** | C | No locking, no rate limiting |
| **Lifecycle Management** | C+ | No cascading stop, no orphan detection |
| **State Management** | B | Partial event-sourcing |
| **Production Hardening** | C+ | Missing critical safety mechanisms |
| **Schema-First Tools** | A | Excellent JSON schemas |
| **Deterministic Boundary** | A | No LLM calls in tools |

**Overall: B+ (Strong foundation, needs production hardening)**

---

## 13. Conclusion

Hex demonstrates a **strong understanding of multi-agent architecture fundamentals** with excellent tool design and clean deterministic boundaries. The Task tool provides a solid foundation for agent coordination.

However, **production readiness requires addressing critical gaps**:
1. Cascading cleanup to prevent orphans
2. File locking for concurrent access
3. API rate limiting across agents
4. Formalized agent lifecycle

The architecture is **well-positioned for enhancement** - the four-layer pattern is recognizable and the recursive delegation pattern is appropriate. Implementing the Phase 1-2 recommendations would elevate Hex from "strong prototype" to "production-ready multi-agent system."

**Recommended Next Steps:**
1. Start with cascading stop (highest impact, lowest effort)
2. Add file locking (critical for multi-agent correctness)
3. Extract orchestrator from UI (enables future scaling)

With these improvements, Hex would be a **textbook example of multi-agent system design**.

---

**Audit Complete**
