# Multi-Agent System Improvements - Implementation Plan

**Created**: 2025-12-07
**Based on**: HEX_MULTIAGENT_AUDIT.md
**Execution Method**: Subagent-Driven Development
**Grade Target**: A (from current B+)

---

## Overview

This plan implements all recommendations from the multi-agent audit across 3 phases:
- **Phase 1**: Critical Safety (cascading stop, file locking, rate limiting)
- **Phase 2**: Architectural Improvements (orchestrator extraction, state machine, max depth)
- **Phase 3**: Observability & Control (cost tracking, event-sourcing, visualizer)

Each task will be executed by a fresh subagent with code review between tasks.

---

## Phase 1: Critical Safety (Priority: CRITICAL)

### Task 1.1: Implement Cascading Stop Protocol

**Problem**: When parent agent stops, child agents continue running (orphans, leaked goroutines, incomplete work)

**Location**: `internal/tools/task_tool.go`, `internal/core/client.go`

**Current Code Issue**:
```go
// internal/tools/task_tool.go:161-180
// Creates subprocess with context but no tracking
cmd := exec.CommandContext(ctx, "hex", args...)

// Parent stops → context cancelled → child killed abruptly
// No graceful shutdown, no cleanup
```

**Requirements**:

1. **Create Process Registry** (`internal/registry/process_registry.go`):
   ```go
   type ProcessRegistry struct {
       mu        sync.RWMutex
       processes map[string]*ManagedProcess  // agentID -> process
   }

   type ManagedProcess struct {
       AgentID    string
       ParentID   string
       Cmd        *exec.Cmd
       Cancel     context.CancelFunc
       Children   []string
       State      ProcessState  // Running, Stopping, Stopped
       StartedAt  time.Time
       StoppedAt  *time.Time
   }

   func (r *ProcessRegistry) Register(parent, child string, cmd *exec.Cmd, cancel context.CancelFunc) error
   func (r *ProcessRegistry) StopCascading(agentID string) error  // RECURSIVE
   func (r *ProcessRegistry) GetOrphans() []*ManagedProcess
   ```

2. **Modify Task Tool** to register processes:
   ```go
   // internal/tools/task_tool.go
   func (t *TaskTool) Execute(params map[string]interface{}) (string, error) {
       // ... existing code ...

       // CHANGE: Register before starting
       agentID := generateAgentID()
       parentID := os.Getenv("HEX_AGENT_ID")  // Pass down via env

       registry.Global().Register(parentID, agentID, cmd, cancel)

       // Start subprocess
       if err := cmd.Start(); err != nil {
           registry.Global().Unregister(agentID)
           return "", err
       }

       // ... existing code ...
   }
   ```

3. **Add Graceful Shutdown Handler** (`internal/shutdown/handler.go`):
   ```go
   func InitShutdownHandler() {
       sigChan := make(chan os.Signal, 1)
       signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

       go func() {
           <-sigChan
           agentID := os.Getenv("HEX_AGENT_ID")
           if agentID != "" {
               registry.Global().StopCascading(agentID)
           }
           os.Exit(0)
       }()
   }
   ```

4. **Tests** (`internal/registry/process_registry_test.go`):
   - TestRegisterProcess
   - TestStopCascading_ThreeLevels
   - TestGetOrphans
   - TestConcurrentRegistration

**Success Criteria**:
- ✅ Parent stop triggers recursive child stops
- ✅ No orphaned processes after parent exits
- ✅ Processes stopped in correct order (leaves first, root last)
- ✅ All goroutines cleaned up (verified with pprof)
- ✅ All tests passing

**Files to Create**:
- internal/registry/process_registry.go
- internal/registry/process_registry_test.go
- internal/shutdown/handler.go
- internal/shutdown/handler_test.go

**Files to Modify**:
- internal/tools/task_tool.go (add registration)
- cmd/hex/root.go (call InitShutdownHandler)

---

### Task 1.2: Add File Locking for Concurrent Writes

**Problem**: Multiple agents can write to the same file simultaneously (race condition, corruption)

**Location**: `internal/tools/write_file_tool.go`, `internal/tools/edit_tool.go`

**Current Code Issue**:
```go
// internal/tools/write_file_tool.go:87
// No locking before write
if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
    return "", fmt.Errorf("failed to write file: %w", err)
}

// RACE: Two agents writing same file → undefined result
```

**Requirements**:

1. **Create File Lock Manager** (`internal/filelock/manager.go`):
   ```go
   type LockManager struct {
       mu    sync.Mutex
       locks map[string]*FileLock
   }

   type FileLock struct {
       Path      string
       OwnerID   string
       LockedAt  time.Time
       mu        sync.RWMutex
   }

   func (m *LockManager) Acquire(path, ownerID string, timeout time.Duration) error
   func (m *LockManager) Release(path, ownerID string) error
   func (m *LockManager) ForceRelease(path string) error  // For cleanup
   func (m *LockManager) IsLocked(path string) bool
   ```

2. **Modify Write Tool**:
   ```go
   // internal/tools/write_file_tool.go
   func (t *WriteTool) Execute(params map[string]interface{}) (string, error) {
       filePath := params["file_path"].(string)
       content := params["content"].(string)

       agentID := os.Getenv("HEX_AGENT_ID")
       lockMgr := filelock.Global()

       // ACQUIRE LOCK (30s timeout)
       if err := lockMgr.Acquire(filePath, agentID, 30*time.Second); err != nil {
           return "", fmt.Errorf("could not acquire file lock: %w", err)
       }
       defer lockMgr.Release(filePath, agentID)

       // Existing write logic
       if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
           return "", fmt.Errorf("failed to write file: %w", err)
       }

       return fmt.Sprintf("Successfully wrote to %s", filePath), nil
   }
   ```

3. **Modify Edit Tool** (similar pattern):
   ```go
   // internal/tools/edit_tool.go
   func (t *EditTool) Execute(params map[string]interface{}) (string, error) {
       filePath := params["file_path"].(string)

       // Acquire lock BEFORE reading
       agentID := os.Getenv("HEX_AGENT_ID")
       lockMgr := filelock.Global()

       if err := lockMgr.Acquire(filePath, agentID, 30*time.Second); err != nil {
           return "", fmt.Errorf("could not acquire file lock: %w", err)
       }
       defer lockMgr.Release(filePath, agentID)

       // Existing edit logic (read, replace, write)
       // ...
   }
   ```

4. **Add Cleanup on Shutdown**:
   ```go
   // internal/shutdown/handler.go
   func InitShutdownHandler() {
       // ... existing code ...

       // Release all locks held by this agent
       agentID := os.Getenv("HEX_AGENT_ID")
       filelock.Global().ReleaseAll(agentID)
   }
   ```

5. **Tests** (`internal/filelock/manager_test.go`):
   - TestAcquireRelease
   - TestConcurrentAcquire_BlocksSecond
   - TestTimeout_ReturnsError
   - TestForceRelease
   - TestDeadlockPrevention

**Success Criteria**:
- ✅ Concurrent writes to same file are serialized
- ✅ Lock timeout prevents infinite blocking
- ✅ Locks released on agent exit
- ✅ No deadlocks (verified with stress test)
- ✅ All tests passing

**Files to Create**:
- internal/filelock/manager.go
- internal/filelock/manager_test.go

**Files to Modify**:
- internal/tools/write_file_tool.go
- internal/tools/edit_tool.go
- internal/shutdown/handler.go (add lock cleanup)

---

### Task 1.3: Add Shared API Rate Limiter

**Problem**: Multiple parallel agents can cause API rate limit errors (429 responses)

**Location**: `internal/core/client.go`

**Current Code Issue**:
```go
// internal/core/client.go:74
// Each agent creates independent client
c := &Client{
    apiKey:     apiKey,
    httpClient: &http.Client{Timeout: 10 * time.Minute},
}

// NO COORDINATION between agents
// Parallel agents → burst requests → rate limit errors
```

**Requirements**:

1. **Create Shared Rate Limiter** (`internal/ratelimit/limiter.go`):
   ```go
   type SharedLimiter struct {
       mu            sync.Mutex
       tokens        int
       maxTokens     int           // 50 (Anthropic limit)
       refillRate    time.Duration // 1 token per minute
       lastRefill    time.Time
       waitQueue     chan struct{}
   }

   func NewSharedLimiter(maxTokens int, refillRate time.Duration) *SharedLimiter
   func (l *SharedLimiter) Acquire(ctx context.Context) error
   func (l *SharedLimiter) refill()  // Background goroutine
   ```

2. **Implement Token Bucket Algorithm**:
   ```go
   func (l *SharedLimiter) Acquire(ctx context.Context) error {
       for {
           l.mu.Lock()

           // Refill tokens based on time elapsed
           now := time.Now()
           elapsed := now.Sub(l.lastRefill)
           tokensToAdd := int(elapsed / l.refillRate)
           if tokensToAdd > 0 {
               l.tokens = min(l.tokens + tokensToAdd, l.maxTokens)
               l.lastRefill = now
           }

           // Try to acquire
           if l.tokens > 0 {
               l.tokens--
               l.mu.Unlock()
               return nil
           }

           l.mu.Unlock()

           // Wait for refill or context cancellation
           select {
           case <-ctx.Done():
               return ctx.Err()
           case <-time.After(l.refillRate):
               continue
           }
       }
   }
   ```

3. **Modify Client** to use shared limiter:
   ```go
   // internal/core/client.go
   var globalLimiter = ratelimit.NewSharedLimiter(50, time.Minute)

   func (c *Client) CreateMessage(ctx context.Context, req MessageRequest) (*MessageResponse, error) {
       // ACQUIRE BEFORE API CALL
       if err := globalLimiter.Acquire(ctx); err != nil {
           return nil, fmt.Errorf("rate limit acquire failed: %w", err)
       }

       // Existing API call
       resp, err := c.httpClient.Do(httpReq)
       // ...
   }
   ```

4. **Add Metrics** for debugging:
   ```go
   type Metrics struct {
       TotalAcquired   atomic.Int64
       TotalWaits      atomic.Int64
       CurrentTokens   atomic.Int64
       AvgWaitTimeMs   atomic.Int64
   }

   func (l *SharedLimiter) GetMetrics() Metrics
   ```

5. **Tests** (`internal/ratelimit/limiter_test.go`):
   - TestAcquire_ImmediateWhenTokensAvailable
   - TestAcquire_WaitsWhenEmpty
   - TestRefill_AddsTokensOverTime
   - TestConcurrentAcquire_Fair
   - TestContextCancellation

**Success Criteria**:
- ✅ No 429 rate limit errors with parallel agents
- ✅ Tokens refilled at correct rate
- ✅ Fair distribution across agents
- ✅ Context cancellation works
- ✅ All tests passing

**Files to Create**:
- internal/ratelimit/limiter.go
- internal/ratelimit/limiter_test.go

**Files to Modify**:
- internal/core/client.go (add Acquire before API calls)

---

## Phase 2: Architectural Improvements (Priority: HIGH)

### Task 2.1: Extract AgentOrchestrator from UI Layer

**Problem**: 1400-line update.go mixes orchestration logic with UI rendering (hard to test, hard to reason about)

**Location**: `internal/ui/update.go`, `internal/ui/model.go`

**Current Code Issue**:
```go
// internal/ui/update.go:1-1400
// All agent lifecycle logic embedded in BubbleTea Update function
// Impossible to test without UI, hard to understand control flow
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case subscriptionMsg:        // Agent logic
    case streamToolStartMsg:      // Agent logic
    case streamContentBlockMsg:   // Agent logic
    case approvalRequiredMsg:     // Agent logic
    // ... 50+ message types mixing UI and orchestration
    }
}
```

**Requirements**:

1. **Create AgentOrchestrator** (`internal/orchestrator/orchestrator.go`):
   ```go
   type AgentOrchestrator struct {
       client          *core.Client
       toolExecutor    *tools.Executor
       messageHistory  []core.Message
       state           AgentState
       eventChan       chan Event  // Send events to UI
       stopChan        chan struct{}
   }

   type Event struct {
       Type      EventType  // StreamStart, ToolCall, ToolResult, Complete, Error
       Data      interface{}
       Timestamp time.Time
   }

   func (o *AgentOrchestrator) Start(ctx context.Context, prompt string) error
   func (o *AgentOrchestrator) HandleToolApproval(toolUseID string, approved bool) error
   func (o *AgentOrchestrator) Stop() error
   func (o *AgentOrchestrator) GetState() AgentState
   func (o *AgentOrchestrator) Subscribe() <-chan Event
   ```

2. **Extract State Machine**:
   ```go
   type AgentState string

   const (
       StateIdle           AgentState = "idle"
       StateStreaming      AgentState = "streaming"
       StateAwaitingApproval AgentState = "awaiting_approval"
       StateExecutingTool  AgentState = "executing_tool"
       StateComplete       AgentState = "complete"
       StateError          AgentState = "error"
   )

   func (s AgentState) CanTransitionTo(next AgentState) bool {
       validTransitions := map[AgentState][]AgentState{
           StateIdle:             {StateStreaming},
           StateStreaming:        {StateAwaitingApproval, StateComplete, StateError},
           StateAwaitingApproval: {StateExecutingTool, StateError},
           StateExecutingTool:    {StateStreaming, StateError},
           StateComplete:         {},
           StateError:            {StateIdle},
       }
       // ...
   }
   ```

3. **Refactor Update Function** to consume events:
   ```go
   // internal/ui/update.go (SIMPLIFIED)
   func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
       switch msg := msg.(type) {
       case orchestrator.Event:
           return m.handleOrchestratorEvent(msg)

       case tea.KeyMsg:
           return m.handleKeyPress(msg)

       // UI-only messages
       case tickMsg:
           return m.handleTick()
       }
   }

   func (m model) handleOrchestratorEvent(event orchestrator.Event) (tea.Model, tea.Cmd) {
       switch event.Type {
       case orchestrator.StreamStart:
           m.streaming = true
           return m, nil

       case orchestrator.ToolCall:
           m.pendingToolCall = event.Data.(ToolCall)
           return m, nil

       // ... much simpler
       }
   }
   ```

4. **Move Stream Handling** to orchestrator:
   ```go
   // internal/orchestrator/stream_handler.go
   func (o *AgentOrchestrator) handleStream(ctx context.Context) error {
       stream := o.client.CreateMessageStream(ctx, req)

       for {
           select {
           case <-ctx.Done():
               return ctx.Err()

           case chunk, ok := <-stream:
               if !ok {
                   return nil
               }

               o.processChunk(chunk)
           }
       }
   }

   func (o *AgentOrchestrator) processChunk(chunk StreamChunk) {
       switch chunk.Type {
       case "content_block_start":
           if chunk.ContentBlock.Type == "tool_use" {
               o.eventChan <- Event{Type: ToolCall, Data: chunk.ContentBlock}
           }
       // ...
       }
   }
   ```

5. **Tests** (`internal/orchestrator/orchestrator_test.go`):
   - TestStart_SendsStreamStartEvent
   - TestToolApproval_TransitionsState
   - TestStop_CancelsStream
   - TestStateTransitions_Valid
   - TestStateTransitions_Invalid
   - TestConcurrentEventEmission

**Success Criteria**:
- ✅ Orchestrator is UI-independent (100% unit testable)
- ✅ update.go reduced to <500 lines (UI-only logic)
- ✅ All state transitions validated
- ✅ Events emitted in correct order
- ✅ All tests passing (orchestrator + UI integration)

**Files to Create**:
- internal/orchestrator/orchestrator.go
- internal/orchestrator/stream_handler.go
- internal/orchestrator/state.go
- internal/orchestrator/orchestrator_test.go

**Files to Modify**:
- internal/ui/update.go (simplified to consume events)
- internal/ui/model.go (embed orchestrator)

---

### Task 2.2: Formalize Agent State Machine

**Problem**: Agent states are implicit, transitions not validated, hard to debug state-related bugs

**Location**: Extracted from orchestrator work above

**Requirements**:

1. **Define Formal State Machine** (`internal/orchestrator/state_machine.go`):
   ```go
   type StateMachine struct {
       current     AgentState
       mu          sync.RWMutex
       transitions map[AgentState][]AgentState
       observers   []StateObserver
   }

   type StateObserver interface {
       OnStateChange(from, to AgentState, data interface{})
   }

   func NewStateMachine(initial AgentState) *StateMachine {
       sm := &StateMachine{
           current: initial,
           transitions: map[AgentState][]AgentState{
               StateIdle:             {StateStreaming},
               StateStreaming:        {StateAwaitingApproval, StateComplete, StateError},
               StateAwaitingApproval: {StateExecutingTool, StateError},
               StateExecutingTool:    {StateStreaming, StateError},
               StateComplete:         {},
               StateError:            {StateIdle},
           },
       }
       return sm
   }

   func (sm *StateMachine) Transition(to AgentState, data interface{}) error {
       sm.mu.Lock()
       defer sm.mu.Unlock()

       if !sm.canTransition(to) {
           return fmt.Errorf("invalid transition: %s -> %s", sm.current, to)
       }

       from := sm.current
       sm.current = to

       // Notify observers
       for _, obs := range sm.observers {
           obs.OnStateChange(from, to, data)
       }

       return nil
   }

   func (sm *StateMachine) Current() AgentState {
       sm.mu.RLock()
       defer sm.mu.RUnlock()
       return sm.current
   }

   func (sm *StateMachine) AddObserver(obs StateObserver) {
       sm.observers = append(sm.observers, obs)
   }
   ```

2. **Add State Logging Observer**:
   ```go
   type LoggingObserver struct{}

   func (l *LoggingObserver) OnStateChange(from, to AgentState, data interface{}) {
       logging.Debug("State transition",
           "from", from,
           "to", to,
           "data", fmt.Sprintf("%+v", data))
   }
   ```

3. **Add State History for Debugging**:
   ```go
   type StateHistory struct {
       mu      sync.Mutex
       history []StateTransition
   }

   type StateTransition struct {
       From      AgentState
       To        AgentState
       Timestamp time.Time
       Data      interface{}
   }

   func (h *StateHistory) OnStateChange(from, to AgentState, data interface{}) {
       h.mu.Lock()
       defer h.mu.Unlock()

       h.history = append(h.history, StateTransition{
           From: from, To: to, Timestamp: time.Now(), Data: data,
       })
   }

   func (h *StateHistory) GetHistory() []StateTransition {
       h.mu.Lock()
       defer h.mu.Unlock()
       return append([]StateTransition{}, h.history...)
   }
   ```

4. **Integrate with Orchestrator**:
   ```go
   type AgentOrchestrator struct {
       // ... existing fields ...
       stateMachine *StateMachine
       stateHistory *StateHistory
   }

   func NewOrchestrator(...) *AgentOrchestrator {
       sm := NewStateMachine(StateIdle)
       history := &StateHistory{}

       sm.AddObserver(&LoggingObserver{})
       sm.AddObserver(history)

       return &AgentOrchestrator{
           stateMachine: sm,
           stateHistory: history,
       }
   }

   func (o *AgentOrchestrator) Start(ctx context.Context, prompt string) error {
       if err := o.stateMachine.Transition(StateStreaming, nil); err != nil {
           return err
       }
       // ...
   }
   ```

5. **Tests** (`internal/orchestrator/state_machine_test.go`):
   - TestValidTransitions
   - TestInvalidTransitions_ReturnError
   - TestConcurrentTransitions_ThreadSafe
   - TestObserverNotification
   - TestStateHistory

**Success Criteria**:
- ✅ All transitions validated at runtime
- ✅ Invalid transitions return errors (no silent failures)
- ✅ State history captured for debugging
- ✅ Thread-safe
- ✅ All tests passing

**Files to Create**:
- internal/orchestrator/state_machine.go
- internal/orchestrator/state_machine_test.go
- internal/orchestrator/state_observers.go

**Files to Modify**:
- internal/orchestrator/orchestrator.go (integrate state machine)

---

### Task 2.3: Add Max Recursion Depth to Task Tool

**Problem**: Infinite recursion possible if agent repeatedly creates subagents

**Location**: `internal/tools/task_tool.go`

**Current Code Issue**:
```go
// No depth tracking
// Agent can spawn subagent, which spawns subagent, which spawns...
// Infinite recursion until OOM or rate limit
```

**Requirements**:

1. **Add Depth Tracking**:
   ```go
   const MaxAgentDepth = 5  // Configurable via env var

   func (t *TaskTool) Execute(params map[string]interface{}) (string, error) {
       currentDepth := getAgentDepth()

       if currentDepth >= MaxAgentDepth {
           return "", fmt.Errorf("max agent depth (%d) exceeded, cannot create subagent", MaxAgentDepth)
       }

       // Pass depth to child via environment
       env := os.Environ()
       env = append(env, fmt.Sprintf("HEX_AGENT_DEPTH=%d", currentDepth+1))
       cmd.Env = env

       // ... rest of execution
   }

   func getAgentDepth() int {
       depthStr := os.Getenv("HEX_AGENT_DEPTH")
       if depthStr == "" {
           return 0  // Root agent
       }
       depth, _ := strconv.Atoi(depthStr)
       return depth
   }
   ```

2. **Make Limit Configurable**:
   ```go
   func getMaxAgentDepth() int {
       maxDepthStr := os.Getenv("HEX_MAX_AGENT_DEPTH")
       if maxDepthStr == "" {
           return 5  // Default
       }
       maxDepth, err := strconv.Atoi(maxDepthStr)
       if err != nil || maxDepth < 1 {
           return 5
       }
       return maxDepth
   }
   ```

3. **Add Depth to Logging**:
   ```go
   func (t *TaskTool) Execute(params map[string]interface{}) (string, error) {
       depth := getAgentDepth()

       logging.Debug("Task tool execution",
           "depth", depth,
           "max_depth", getMaxAgentDepth(),
           "subagent_type", params["subagent_type"])

       // ...
   }
   ```

4. **Error Message Guidance**:
   ```go
   if currentDepth >= maxDepth {
       return "", fmt.Errorf(
           "max agent depth (%d) exceeded - this usually means:\n"+
           "1. The task is too complex for recursive decomposition\n"+
           "2. The agent is stuck in a loop\n"+
           "3. You need to break down the task differently\n"+
           "Set HEX_MAX_AGENT_DEPTH to increase limit (use with caution)",
           maxDepth)
   }
   ```

5. **Tests** (`internal/tools/task_tool_test.go`):
   - TestMaxDepth_Enforced
   - TestDepth_PassedToChild
   - TestConfigurableLimit
   - TestErrorMessage_Helpful

**Success Criteria**:
- ✅ Recursion stops at max depth
- ✅ Depth passed correctly to children
- ✅ Configurable via env var
- ✅ Clear error message
- ✅ All tests passing

**Files to Modify**:
- internal/tools/task_tool.go
- internal/tools/task_tool_test.go (add depth tests)

---

## Phase 3: Observability & Control (Priority: MEDIUM)

### Task 3.1: Implement Hierarchical Cost Tracking

**Problem**: Can't see per-agent costs, can't set budgets, can't track spend across subagent trees

**Location**: `internal/core/client.go`, new `internal/cost/` package

**Requirements**:

1. **Create Cost Tracker** (`internal/cost/tracker.go`):
   ```go
   type CostTracker struct {
       mu     sync.RWMutex
       costs  map[string]*AgentCost  // agentID -> cost
   }

   type AgentCost struct {
       AgentID     string
       ParentID    string
       Model       string
       InputTokens  int64
       OutputTokens int64
       CacheReads   int64
       CacheWrites  int64
       InputCost    float64   // USD
       OutputCost   float64   // USD
       CacheCost    float64   // USD
       TotalCost    float64   // USD
       StartedAt    time.Time
       CompletedAt  *time.Time
   }

   func (t *CostTracker) RecordUsage(agentID, parentID, model string, usage core.Usage) error
   func (t *CostTracker) GetAgentCost(agentID string) (*AgentCost, error)
   func (t *CostTracker) GetTreeCost(rootAgentID string) (float64, error)  // Recursive sum
   func (t *CostTracker) GetAllCosts() []*AgentCost
   ```

2. **Add Pricing Model** (`internal/cost/pricing.go`):
   ```go
   type PricingModel struct {
       InputTokenPrice  float64  // per 1M tokens
       OutputTokenPrice float64
       CacheReadPrice   float64
       CacheWritePrice  float64
   }

   var modelPricing = map[string]PricingModel{
       "claude-sonnet-4-5-20250929": {
           InputTokenPrice:  3.00,
           OutputTokenPrice: 15.00,
           CacheReadPrice:   0.30,
           CacheWritePrice:  3.75,
       },
       // Add other models...
   }

   func calculateCost(model string, usage core.Usage) (float64, error) {
       pricing, ok := modelPricing[model]
       if !ok {
           return 0, fmt.Errorf("unknown model: %s", model)
       }

       inputCost := (float64(usage.InputTokens) / 1_000_000) * pricing.InputTokenPrice
       outputCost := (float64(usage.OutputTokens) / 1_000_000) * pricing.OutputTokenPrice
       cacheReadCost := (float64(usage.CacheReadTokens) / 1_000_000) * pricing.CacheReadPrice
       cacheWriteCost := (float64(usage.CacheWriteTokens) / 1_000_000) * pricing.CacheWritePrice

       return inputCost + outputCost + cacheReadCost + cacheWriteCost, nil
   }
   ```

3. **Integrate with Client**:
   ```go
   // internal/core/client.go
   func (c *Client) CreateMessage(ctx context.Context, req MessageRequest) (*MessageResponse, error) {
       // ... existing code ...

       resp, err := c.doRequest(ctx, req)
       if err != nil {
           return nil, err
       }

       // Record cost
       agentID := os.Getenv("HEX_AGENT_ID")
       parentID := os.Getenv("HEX_PARENT_AGENT_ID")

       if agentID != "" {
           cost.Global().RecordUsage(agentID, parentID, req.Model, resp.Usage)
       }

       return resp, nil
   }
   ```

4. **Add Budget Enforcement** (`internal/cost/budget.go`):
   ```go
   type BudgetEnforcer struct {
       maxCostUSD float64
       tracker    *CostTracker
   }

   func (b *BudgetEnforcer) CheckBudget(agentID string) error {
       treeCost, err := b.tracker.GetTreeCost(agentID)
       if err != nil {
           return err
       }

       if treeCost > b.maxCostUSD {
           return fmt.Errorf("budget exceeded: $%.4f > $%.4f", treeCost, b.maxCostUSD)
       }

       return nil
   }

   // Call before each API request
   ```

5. **Add Cost Reporting** (`internal/cost/reporter.go`):
   ```go
   func PrintCostSummary(rootAgentID string) {
       tracker := Global()
       treeCost, _ := tracker.GetTreeCost(rootAgentID)

       fmt.Fprintf(os.Stderr, "\n=== Cost Summary ===\n")
       fmt.Fprintf(os.Stderr, "Total Cost: $%.4f\n", treeCost)

       // Print breakdown
       costs := tracker.GetAllCosts()
       for _, cost := range costs {
           fmt.Fprintf(os.Stderr, "  Agent %s: $%.4f (%d in, %d out tokens)\n",
               cost.AgentID, cost.TotalCost, cost.InputTokens, cost.OutputTokens)
       }
   }
   ```

6. **Tests** (`internal/cost/tracker_test.go`):
   - TestRecordUsage
   - TestCalculateCost_Accurate
   - TestGetTreeCost_Recursive
   - TestBudgetEnforcement
   - TestConcurrentRecording

**Success Criteria**:
- ✅ Per-agent costs tracked accurately
- ✅ Tree cost calculated correctly (recursive sum)
- ✅ Budget enforcement works
- ✅ Cost summary displayed after run
- ✅ All tests passing

**Files to Create**:
- internal/cost/tracker.go
- internal/cost/pricing.go
- internal/cost/budget.go
- internal/cost/reporter.go
- internal/cost/tracker_test.go

**Files to Modify**:
- internal/core/client.go (record usage after each API call)
- cmd/hex/root.go (print cost summary on exit)

---

### Task 3.2: Add Complete Event-Sourcing

**Problem**: No audit trail, can't replay agent executions, hard to debug complex failures

**Location**: New `internal/events/` package

**Requirements**:

1. **Define Event Types** (`internal/events/types.go`):
   ```go
   type Event struct {
       ID        string      // UUID
       AgentID   string      // Hierarchical ID (root, root.1, root.1.2)
       ParentID  string
       Type      EventType
       Timestamp time.Time
       Data      interface{}
   }

   type EventType string

   const (
       EventAgentStarted       EventType = "agent_started"
       EventAgentStopped       EventType = "agent_stopped"
       EventStreamStarted      EventType = "stream_started"
       EventStreamChunk        EventType = "stream_chunk"
       EventToolCallRequested  EventType = "tool_call_requested"
       EventToolCallApproved   EventType = "tool_call_approved"
       EventToolCallDenied     EventType = "tool_call_denied"
       EventToolExecutionStart EventType = "tool_execution_start"
       EventToolExecutionEnd   EventType = "tool_execution_end"
       EventStateTransition    EventType = "state_transition"
       EventError              EventType = "error"
   )
   ```

2. **Create Event Store** (`internal/events/store.go`):
   ```go
   type EventStore struct {
       mu     sync.RWMutex
       events []Event
       file   *os.File  // Write to disk
   }

   func NewEventStore(filepath string) (*EventStore, error) {
       f, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
       if err != nil {
           return nil, err
       }

       return &EventStore{file: f}, nil
   }

   func (s *EventStore) Record(event Event) error {
       s.mu.Lock()
       defer s.mu.Unlock()

       s.events = append(s.events, event)

       // Write to disk (JSON lines format)
       line, err := json.Marshal(event)
       if err != nil {
           return err
       }

       _, err = s.file.Write(append(line, '\n'))
       return err
   }

   func (s *EventStore) GetEvents(agentID string) []Event {
       s.mu.RLock()
       defer s.mu.RUnlock()

       var filtered []Event
       for _, e := range s.events {
           if e.AgentID == agentID || strings.HasPrefix(e.AgentID, agentID+".") {
               filtered = append(filtered, e)
           }
       }
       return filtered
   }

   func (s *EventStore) Close() error {
       return s.file.Close()
   }
   ```

3. **Generate Hierarchical Agent IDs**:
   ```go
   // internal/events/agent_id.go
   func GenerateAgentID(parentID string) string {
       if parentID == "" {
           return "root"
       }

       // Count existing children
       childCount := getChildCount(parentID)
       return fmt.Sprintf("%s.%d", parentID, childCount+1)
   }

   // Examples:
   // root
   // root.1
   // root.2
   // root.1.1
   // root.1.2
   ```

4. **Integrate Event Recording**:
   ```go
   // Record events at key points

   // Agent start
   events.Global().Record(Event{
       AgentID: agentID,
       Type: EventAgentStarted,
       Data: map[string]interface{}{"prompt": prompt},
   })

   // Tool call
   events.Global().Record(Event{
       AgentID: agentID,
       Type: EventToolCallRequested,
       Data: ToolCall{ID: toolUseID, Name: toolName, Params: params},
   })

   // State transition
   events.Global().Record(Event{
       AgentID: agentID,
       Type: EventStateTransition,
       Data: map[string]interface{}{"from": from, "to": to},
   })
   ```

5. **Add Replay Tool** (`cmd/hexreplay/main.go`):
   ```go
   func main() {
       eventFile := flag.String("events", "hex_events.jsonl", "event file")
       flag.Parse()

       store, err := events.LoadEventStore(*eventFile)
       if err != nil {
           log.Fatal(err)
       }

       allEvents := store.GetEvents("root")

       // Replay events (print timeline)
       for _, e := range allEvents {
           fmt.Printf("[%s] %s: %s\n",
               e.Timestamp.Format(time.RFC3339),
               e.AgentID,
               e.Type)
       }
   }
   ```

6. **Tests** (`internal/events/store_test.go`):
   - TestRecordEvent
   - TestGetEvents_Filtered
   - TestHierarchicalIDs
   - TestPersistenceToDisk
   - TestConcurrentRecording

**Success Criteria**:
- ✅ All agent events captured
- ✅ Events persisted to disk (hex_events.jsonl)
- ✅ Hierarchical agent IDs work
- ✅ Replay tool shows timeline
- ✅ All tests passing

**Files to Create**:
- internal/events/types.go
- internal/events/store.go
- internal/events/agent_id.go
- internal/events/store_test.go
- cmd/hexreplay/main.go

**Files to Modify**:
- internal/orchestrator/orchestrator.go (record events)
- internal/tools/task_tool.go (record subagent events)
- cmd/hex/root.go (initialize event store)

---

### Task 3.3: Build Agent Execution Visualizer

**Problem**: Hard to understand complex multi-agent executions, need visual representation

**Location**: New `cmd/hexviz/` package

**Requirements**:

1. **Create ASCII Visualizer** (`cmd/hexviz/main.go`):
   ```go
   func main() {
       eventFile := flag.String("events", "hex_events.jsonl", "event file")
       flag.Parse()

       store, err := events.LoadEventStore(*eventFile)
       if err != nil {
           log.Fatal(err)
       }

       printTreeView(store)
       printTimelineView(store)
       printCostView(store)
   }
   ```

2. **Tree View** (agent hierarchy):
   ```
   root (StateComplete, $0.0234)
   ├─ root.1 (StateComplete, $0.0045) [explore subagent]
   │  ├─ root.1.1 (StateComplete, $0.0012) [nested task]
   │  └─ root.1.2 (StateComplete, $0.0008)
   └─ root.2 (StateComplete, $0.0125) [code-reviewer]
      └─ root.2.1 (StateError, $0.0004) [failed subtask]
   ```

3. **Timeline View** (chronological):
   ```
   00:00.000  [root    ] AgentStarted
   00:01.234  [root    ] StreamStarted
   00:03.456  [root    ] ToolCallRequested (task_tool)
   00:03.500  [root.1  ] AgentStarted (explore)
   00:05.123  [root.1  ] ToolCallRequested (grep_tool)
   00:05.678  [root.1  ] ToolExecutionEnd (grep_tool, 0.555s)
   00:07.890  [root.1  ] AgentStopped
   00:08.000  [root    ] ToolExecutionEnd (task_tool, 4.500s)
   00:09.123  [root    ] AgentStopped
   ```

4. **Cost View** (breakdown):
   ```
   === Cost Breakdown ===
   Total: $0.0234

   root:        $0.0234 (100%)  45k in, 12k out
     root.1:    $0.0045 ( 19%)  12k in, 3k out
       root.1.1 $0.0012 (  5%)  3k in, 1k out
       root.1.2 $0.0008 (  3%)  2k in, 500 out
     root.2:    $0.0125 ( 53%)  28k in, 8k out
       root.2.1 $0.0004 (  2%)  1k in, 200 out
   ```

5. **Add Filtering**:
   ```go
   agentFilter := flag.String("agent", "", "filter by agent ID")
   eventTypeFilter := flag.String("type", "", "filter by event type")

   // Filter events before visualization
   filtered := filterEvents(store.GetEvents("root"), *agentFilter, *eventTypeFilter)
   ```

6. **Export to HTML** (interactive version):
   ```go
   htmlOutput := flag.Bool("html", false, "generate HTML report")

   if *htmlOutput {
       generateHTMLReport(store, "hex_report.html")
   }
   ```

**Success Criteria**:
- ✅ Tree view shows hierarchy
- ✅ Timeline shows chronological events
- ✅ Cost view shows breakdown
- ✅ Filtering works
- ✅ HTML export works

**Files to Create**:
- cmd/hexviz/main.go
- cmd/hexviz/tree_view.go
- cmd/hexviz/timeline_view.go
- cmd/hexviz/cost_view.go
- cmd/hexviz/html_generator.go

---

## Execution Order

**Phase 1** (execute in order, one task per subagent):
1. Task 1.1 (Cascading Stop) → Code review → Commit
2. Task 1.2 (File Locking) → Code review → Commit
3. Task 1.3 (Rate Limiting) → Code review → Commit

**Phase 2** (execute in order):
1. Task 2.1 (Extract Orchestrator) → Code review → Commit
2. Task 2.2 (State Machine) → Code review → Commit
3. Task 2.3 (Max Depth) → Code review → Commit

**Phase 3** (execute in order):
1. Task 3.1 (Cost Tracking) → Code review → Commit
2. Task 3.2 (Event-Sourcing) → Code review → Commit
3. Task 3.3 (Visualizer) → Code review → Commit

---

## Testing Strategy

**Per-Task Testing**:
- Unit tests for each component (>80% coverage)
- Integration tests where needed
- Scenario tests updated/added

**Post-Phase Testing**:
- Run all 13 scenario tests
- Run integration test suite
- Manual testing of new features

**Final Testing** (after Phase 3):
- Full scenario suite (13 tests)
- Performance testing (no regressions)
- Memory leak detection (pprof)
- Fresh-eyes code review

---

## Success Metrics

**Phase 1**: No orphaned processes, no file corruption, no rate limit errors
**Phase 2**: Orchestrator tested independently, state transitions validated, recursion limited
**Phase 3**: Cost tracking accurate, complete audit trail, visualizations clear

**Overall**: Grade improves from B+ to A in multi-agent audit

---

## Notes for Subagent Execution

- Each task is completely independent (can be executed by fresh subagent)
- All file paths are absolute
- All code patterns are specified
- All test requirements are clear
- TDD approach: write tests first, then implementation
- Code review after each task before proceeding

**Subagent Instructions**:
1. Read this plan
2. Implement your assigned task following TDD
3. Write comprehensive tests
4. Ensure all tests pass
5. Request code review
6. Fix any issues found
7. Mark task complete

---

**End of Plan**
