# TUI Simplification & Robustness Plan

**Goal**: Make hex TUI rock-solid and remove jankiness

**Status**: 5 major bugs already fixed today! This plan covers remaining work.

---

## ✅ Already Fixed (Today's Work)

1. ✅ **Tool Results Visibility** - Tool outputs now visible with visual indicators
2. ✅ **Button Mashing Guard** - Can't execute tools multiple times
3. ✅ **Stream Cancellation Leaks** - Proper cleanup with cancelStream() helper
4. ✅ **Message Cache Invalidation** - Cache properly invalidates on content changes
5. ✅ **Viewport Throttling** - Throttles ALL updates, not just streaming

---

## 🔴 PHASE 1: Critical Robustness Fixes (Week 1)

### 1. Fix Escape Key Ambiguity
**Location**: `internal/ui/update.go:258-274`
**Problem**: Escape has 4+ behaviors with unclear priority
**Impact**: User confusion, have to press Escape multiple times

**Fix**:
```go
// Clear priority order:
// 1. Close help (highest priority)
// 2. Exit search mode
// 3. Dismiss suggestions
// 4. Clear quick actions
// (NOT quit - use Ctrl+C for that)

if msg.Type == tea.KeyEsc {
    if m.helpVisible {
        m.helpVisible = false
        return m, nil
    }
    if m.SearchMode {
        m.ExitSearchMode()
        return m, nil
    }
    if m.showSuggestions {
        m.showSuggestions = false
        return m, nil
    }
    if m.quickActionsMode {
        m.quickActionsMode = false
        return m, nil
    }
    // Don't quit on Escape - too dangerous
    return m, nil
}
```

**Effort**: 30 minutes
**Risk**: Low
**ROI**: High - removes daily annoyance

---

### 2. Fix Unsafe Slice Iteration
**Location**: `internal/ui/update.go:606-645`
**Problem**: Taking pointers to slice elements during iteration - unsafe if slice grows
**Impact**: Potential crashes if messages added during rendering

**Current Code**:
```go
for i := range m.Messages {
    msg := &m.Messages[i]  // ← Dangerous if slice reallocates
    rendered, err := m.RenderMessage(msg)
    ...
}
```

**Options**:

**Option A** (Simple - we need cache updates):
```go
// Since we need to update cache on the actual model messages,
// we're actually safe because BubbleTea is single-threaded
// and Update is the only place that modifies the model.
// Just add a comment explaining this:

// Safe: BubbleTea guarantees single-threaded Update calls,
// so Messages slice won't grow during this loop
for i := range m.Messages {
    msg := &m.Messages[i]
    ...
}
```

**Option B** (Paranoid):
```go
// Snapshot length to detect modifications
messageCount := len(m.Messages)
for i := 0; i < messageCount; i++ {
    msg := &m.Messages[i]
    ...
}
// After loop, check if messages were added
if len(m.Messages) != messageCount {
    logging.Warn("Messages modified during rendering - triggering re-render")
    m.updateViewport()
}
```

**Effort**: 15 minutes (Option A) or 1 hour (Option B)
**Risk**: Low
**ROI**: Medium - defensive programming

---

### 3. Fix Channel Blocking in readStreamChunks
**Location**: `internal/ui/update.go:989-1017`
**Problem**: Reading from channel without proper context cancellation handling
**Impact**: Goroutine leak if context cancelled, UI freeze

**Current Code**:
```go
func (m *Model) readStreamChunks(streamChan <-chan *core.StreamChunk) tea.Cmd {
    return func() tea.Msg {
        chunk, ok := <-streamChan  // ← Blocks forever if context cancelled
        ...
    }
}
```

**Fix**:
```go
func (m *Model) readStreamChunks(streamChan <-chan *core.StreamChunk, ctx context.Context) tea.Cmd {
    return func() tea.Msg {
        select {
        case chunk, ok := <-streamChan:
            if !ok {
                return StreamChunkMsg{Chunk: nil, Error: fmt.Errorf("stream closed")}
            }
            return StreamChunkMsg{Chunk: chunk}
        case <-ctx.Done():
            return StreamChunkMsg{Error: ctx.Err()}
        }
    }
}
```

**Effort**: 1 hour
**Risk**: Low
**ROI**: High - prevents goroutine leaks

---

## 🟡 PHASE 2: Simplification (Week 2)

### 4. Simplify Approval Form Message Forwarding
**Location**: `internal/ui/update.go:549-576`
**Problem**: Form receives ALL messages including its own - potential loops
**Impact**: Hard to reason about, potential infinite loops

**Fix**: Whitelist safe message types
```go
// Only forward these message types to approval form
if m.toolApprovalForm != nil {
    switch msg.(type) {
    case tea.KeyMsg, tea.WindowSizeMsg:
        // Safe to forward - these are user inputs
        formModel, formCmd := m.toolApprovalForm.Update(msg)
        m.toolApprovalForm = formModel
        cmds = append(cmds, formCmd)
    default:
        // Don't forward internal messages - prevents loops
    }
}
```

**Effort**: 30 minutes
**Risk**: Low
**ROI**: Medium - prevents edge cases

---

### 5. Add Error Handling for Subscription Failures
**Location**: `internal/ui/model.go:629-644`
**Problem**: Event channels close silently, no error shown
**Impact**: UI shows stale data with no indication why

**Fix**:
```go
func waitForConversationEvent(ch <-chan pubsub.Event[services.Conversation]) tea.Cmd {
    return func() tea.Msg {
        event, ok := <-ch
        if !ok {
            // Channel closed - return error message instead of nil
            return errorMsg{err: fmt.Errorf("conversation event subscription closed")}
        }
        return conversationEventMsg{event: event}
    }
}
```

**Effort**: 30 minutes
**Risk**: Low
**ROI**: Medium - better error visibility

---

## 🟢 PHASE 3: Complexity Reduction (Week 3)

### 6. Refactor handleStreamChunk Mega-Function
**Location**: `internal/ui/update.go:680-941` (261 lines!)
**Problem**: One function does 6+ things - hard to debug
**Impact**: High cognitive load, bugs hide easily

**Strategy**: Extract specialized handlers

**Current**:
```go
func (m *Model) handleStreamChunk(msg StreamChunkMsg) (*Model, tea.Cmd) {
    // 261 lines doing:
    // - Error handling
    // - Tool use assembly
    // - Text delta handling
    // - Content block management
    // - Stop reason handling
    // - Viewport updates
}
```

**After Refactoring**:
```go
func (m *Model) handleStreamChunk(msg StreamChunkMsg) (*Model, tea.Cmd) {
    if msg.Error != nil {
        return m.handleStreamError(msg.Error)
    }

    chunk := msg.Chunk
    switch chunk.Type {
    case "content_block_start":
        return m.handleContentBlockStart(chunk)
    case "content_block_delta":
        return m.handleContentBlockDelta(chunk)
    case "content_block_stop":
        return m.handleContentBlockStop(chunk)
    case "message_delta":
        return m.handleMessageDelta(chunk)
    case "message_stop":
        return m.handleMessageStop()
    default:
        return m, nil
    }
}

// Extracted handlers (30-50 lines each):
func (m *Model) handleStreamError(err error) (*Model, tea.Cmd) { ... }
func (m *Model) handleContentBlockStart(chunk *StreamChunk) (*Model, tea.Cmd) { ... }
func (m *Model) handleContentBlockDelta(chunk *StreamChunk) (*Model, tea.Cmd) { ... }
func (m *Model) handleContentBlockStop(chunk *StreamChunk) (*Model, tea.Cmd) { ... }
func (m *Model) handleMessageDelta(chunk *StreamChunk) (*Model, tea.Cmd) { ... }
func (m *Model) handleMessageStop() (*Model, tea.Cmd) { ... }
```

**Effort**: 4-6 hours
**Risk**: Medium - requires careful testing
**ROI**: High - makes all future streaming work easier

---

### 7. Context Race Condition (Advanced)
**Location**: `internal/ui/update.go:1140-1145`
**Problem**: Context assigned to model before async command runs
**Impact**: Cancel might not work, contexts leak

**This is complex - needs architectural discussion**

Options:
A. Pass context in command closure (no model storage)
B. Use atomic operations for context storage
C. Redesign to avoid storing context in model

**Effort**: 4+ hours
**Risk**: High
**ROI**: Medium - rare but important

**Defer to Phase 4 - needs more thought**

---

## 📊 Effort vs Impact Matrix

| Fix | Effort | Impact | Risk | Phase |
|-----|--------|--------|------|-------|
| Escape key priority | 30m | High | Low | 1 |
| Unsafe slice iteration | 15m-1h | Med | Low | 1 |
| Channel blocking | 1h | High | Low | 1 |
| Form message filter | 30m | Med | Low | 2 |
| Subscription errors | 30m | Med | Low | 2 |
| Refactor handleStreamChunk | 4-6h | High | Med | 3 |
| Context race condition | 4h+ | Med | High | 4 |

**Total Phase 1**: ~2.5 hours for 3 critical fixes
**Total Phase 2**: ~1 hour for 2 simplifications
**Total Phase 3**: ~5 hours for major refactoring

---

## 🎯 Recommended Next Steps

### This Week (Quick Wins):
1. Fix Escape key priority (30m)
2. Add safety comment for slice iteration (15m)
3. Fix channel blocking in readStreamChunks (1h)

**Total: ~2 hours for solid robustness improvements**

### Next Week (Cleanup):
4. Filter form message forwarding (30m)
5. Add subscription error handling (30m)

### Future (Major Refactoring):
6. Refactor handleStreamChunk into modules (1-2 days)
7. Redesign context handling (requires architectural discussion)

---

## 🚫 Things to REMOVE (Simplification)

Candidates for removal/consolidation:

1. **typewriterMode** - When was this last used? Dead feature?
2. **lastKeyWasG** - Overly specific, could use general key sequence buffer
3. **Duplicate view modes** - Do we really need Intro, Chat, History, Tools? Consolidate?
4. **Multiple status indicators** - Status, ErrorMessage, Streaming - can these merge?

**Action**: Audit feature usage and remove dead code

---

## ✅ Success Criteria

After these fixes, the TUI should:
1. ✅ Never crash from user input
2. ✅ Escape key behavior is predictable
3. ✅ No goroutine/memory leaks
4. ✅ Clear error messages when things go wrong
5. ✅ Smooth streaming with no lag
6. ✅ Codebase easier to maintain (smaller functions)

---

**Next Action**: Start with Phase 1, Fix #1 (Escape key priority) - 30 minutes!
