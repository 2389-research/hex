# TUI Code Review Report
## Hex Internal UI Package Review

**Date**: 2025-12-07
**Reviewer**: Claude (Code Review Agent)
**Scope**: `/Users/harper/Public/src/2389/hex/internal/ui/` (model.go, update.go, view.go)

---

## Executive Summary

This review identifies **21 issues** across the TUI codebase, ranging from critical race conditions to code quality improvements. The most serious issues involve concurrent access to shared state, potential nil pointer dereferences, and complex state management that could lead to bugs.

**Priority Breakdown**:
- **CRITICAL**: 5 issues (require immediate attention)
- **HIGH**: 8 issues (should be addressed soon)
- **MEDIUM**: 5 issues (quality improvements)
- **LOW**: 3 issues (nice to have)

---

## CRITICAL Issues

### 1. Race Condition: Context Assignment in Async Command
**Severity**: CRITICAL
**Location**: `update.go:1140-1145`
**Category**: Concurrency Bug

**Description**: The `sendToolResults()` function assigns context to model fields (`m.streamCtx`, `m.streamCancel`) synchronously before launching an async command, but this creates a race condition. The async command in the returned function may execute before or after the BubbleTea Update loop processes the model changes.

```go
// Create cancellable context for this stream BEFORE the async command
ctx, cancel := context.WithCancel(context.Background())
// Store these in the model NOW (synchronously)
m.streamCtx = ctx
m.streamCancel = cancel
_, _ = fmt.Fprintf(os.Stderr, "[DEBUG] sendToolResults: created new context\n")

// Capture necessary state for the command
apiClient := m.apiClient
// ... later in async function, context may not be set yet
```

**Impact**:
- Stream cancellation may not work correctly
- Context leaks if cancel function is overwritten before being called
- Unpredictable behavior during rapid user interactions

**Suggested Fix**:
```go
// Option 1: Pass context to command closure
return func() tea.Msg {
    ctx, cancel := context.WithCancel(context.Background())
    defer func() {
        // Store context in a wrapper message
        return contextCreatedMsg{ctx: ctx, cancel: cancel}
    }()
    // ... use ctx locally
}

// Option 2: Use atomic operations for context storage
// Option 3: Redesign to avoid storing context in model
```

---

### 2. Nil Pointer Dereference: Approval Rules
**Severity**: CRITICAL
**Location**: `model.go:869-885`
**Category**: Nil Safety

**Description**: The code checks `m.approvalRules != nil` before calling methods, but doesn't handle the case where `approvalRules` becomes nil between the check and the method call (TOCTOU - Time Of Check, Time Of Use).

```go
if len(m.pendingToolUses) > 0 && m.approvalRules != nil {
    toolName := m.pendingToolUses[0].Name
    rule := m.approvalRules.Check(toolName)  // approvalRules could be nil here
```

**Impact**:
- Panic if approval rules are cleared concurrently
- Application crash during tool approval flow

**Suggested Fix**:
```go
// Capture approvalRules reference once
approvalRules := m.approvalRules
if len(m.pendingToolUses) > 0 && approvalRules != nil {
    toolName := m.pendingToolUses[0].Name
    rule := approvalRules.Check(toolName)
    // ... rest of logic
}
```

---

### 3. Message History Corruption: ContentBlock Filtering
**Severity**: CRITICAL
**Location**: `update.go:613-616`
**Category**: Logic Error

**Description**: The viewport rendering skips messages with empty `Content` but non-empty `ContentBlock`, which are critical API messages (tool results). However, these messages ARE part of the conversation history and should be visible to users for transparency.

```go
// Skip messages with empty Content but non-empty ContentBlock (tool result messages)
// These are internal API messages that shouldn't be displayed
if msg.Content == "" && len(msg.ContentBlock) > 0 {
    continue  // This hides tool results from user!
}
```

**Impact**:
- Tool results are invisible to users
- Confusing UX - users don't see what tools returned
- Makes debugging tool issues nearly impossible
- Violates transparency principle

**Suggested Fix**:
```go
// Render ContentBlock messages with a special format
if msg.Content == "" && len(msg.ContentBlock) > 0 {
    // Format ContentBlock for display
    content.WriteString(m.renderContentBlocks(msg.ContentBlock))
    continue
}

// Add method to format ContentBlocks
func (m *Model) renderContentBlocks(blocks []core.ContentBlock) string {
    var b strings.Builder
    for _, block := range blocks {
        switch block.Type {
        case "tool_result":
            b.WriteString(fmt.Sprintf("🔧 Tool Result [%s]: %s\n",
                block.ToolUseID, block.Content))
        case "tool_use":
            b.WriteString(fmt.Sprintf("🛠 Tool Call: %s\n", block.Name))
        }
    }
    return b.String()
}
```

---

### 4. Stream Cancellation Memory Leak
**Severity**: CRITICAL
**Location**: `model.go:532-538`, `update.go:279-282`, `update.go:359-362`
**Category**: Resource Leak

**Description**: Multiple places cancel `streamCancel` but don't set it to nil afterwards, and some don't check for nil before calling. This can lead to double-cancellation or panic.

```go
// Cancel any active stream
if m.streamCancel != nil {
    m.streamCancel()  // Cancel is called but pointer not cleared
}
// Later code may call cancel again, or context may leak
```

**Impact**:
- Memory leaks from unclosed contexts
- Possible panic if cancel is called twice
- Goroutine leaks if streams aren't properly cleaned up

**Suggested Fix**:
```go
// Always nil out after canceling
if m.streamCancel != nil {
    m.streamCancel()
    m.streamCancel = nil
    m.streamCtx = nil
    m.streamChan = nil
}

// Create helper method
func (m *Model) cancelStream() {
    if m.streamCancel != nil {
        m.streamCancel()
    }
    m.streamCancel = nil
    m.streamCtx = nil
    m.streamChan = nil
}
```

---

### 5. Tool Approval Form State Corruption
**Severity**: CRITICAL
**Location**: `update.go:549-576`
**Category**: State Management

**Description**: The approval form is forwarded ALL messages in the Update loop, including its own internal messages. This creates a feedback loop and can cause state corruption or infinite loops.

```go
// CRITICAL: Forward ALL messages to approval form when in approval mode
// The form's Init() command generates messages that must be processed
if m.toolApprovalMode && m.toolApprovalForm != nil {
    // This forwards EVERYTHING - including the form's own messages
    var formCmd tea.Cmd
    m.toolApprovalForm, formCmd = m.toolApprovalForm.Update(msg)
    if formCmd != nil {
        cmds = append(cmds, formCmd)
    }
    // ... potentially processes result multiple times
}
```

**Impact**:
- Form state may become corrupted
- Duplicate approval events
- UI freezes or infinite loops
- Unpredictable behavior

**Suggested Fix**:
```go
// Only forward specific message types to form
if m.toolApprovalMode && m.toolApprovalForm != nil {
    switch msg.(type) {
    case tea.KeyMsg, tea.WindowSizeMsg, tea.MouseMsg:
        // Only forward user input and resize events
        var formCmd tea.Cmd
        m.toolApprovalForm, formCmd = m.toolApprovalForm.Update(msg)
        if formCmd != nil {
            cmds = append(cmds, formCmd)
        }
    }
}
```

---

## HIGH Priority Issues

### 6. Viewport Update Throttling Logic Error
**Severity**: HIGH
**Location**: `update.go:586-598`
**Category**: Logic Error

**Description**: The throttling logic is flawed - it checks time but then immediately updates, defeating the purpose. Also, the throttle only applies during streaming, but expensive renders can happen anytime.

```go
timeSinceLastUpdate := time.Since(m.lastViewportUpdate)
if m.Streaming && timeSinceLastUpdate < 16*time.Millisecond {
    return  // Skip update
}
// Immediately update timestamp - this defeats throttling!
m.lastViewportUpdate = time.Now()
```

**Impact**:
- CPU spikes during streaming
- No actual throttling effect
- Poor performance on slower machines

**Suggested Fix**:
```go
func (m *Model) updateViewport() {
    // Always throttle expensive renders, not just during streaming
    timeSinceLastUpdate := time.Since(m.lastViewportUpdate)
    if timeSinceLastUpdate < 16*time.Millisecond {
        m.needsViewportUpdate = true  // Flag for deferred update
        return
    }

    m.lastViewportUpdate = time.Now()
    m.needsViewportUpdate = false
    // ... actual render logic
}

// Add ticker to flush deferred updates
func (m *Model) flushDeferredUpdates() tea.Cmd {
    if m.needsViewportUpdate {
        m.updateViewport()
    }
    return tea.Tick(16*time.Millisecond, func(t time.Time) tea.Msg {
        return viewportRefreshMsg{}
    })
}
```

---

### 7. Unsafe Slice Modification During Iteration
**Severity**: HIGH
**Location**: `update.go:610-643`
**Category**: Data Race

**Description**: The code iterates over `m.Messages` slice with range but takes pointers to elements. If messages are added/removed during rendering (from another goroutine or event), this causes corruption.

```go
for i := range m.Messages {
    msg := &m.Messages[i]  // Pointer to slice element - unsafe if slice reallocates
    // ... later, messages might be appended elsewhere
```

**Impact**:
- Pointer to freed memory if slice grows
- Incorrect message rendering
- Possible panic

**Suggested Fix**:
```go
// Copy messages before iterating
messages := make([]Message, len(m.Messages))
copy(messages, m.Messages)

for i := range messages {
    msg := &messages[i]
    // Safe to use pointer now
```

---

### 8. Missing Error Handling in Event Subscriptions
**Severity**: HIGH
**Location**: `model.go:629-644`
**Category**: Error Handling

**Description**: Event subscription channels can close unexpectedly, but the code only checks `ok` and returns nil, silently losing subscription.

```go
func waitForConversationEvent(ch <-chan pubsub.Event[services.Conversation]) tea.Cmd {
    return func() tea.Msg {
        event, ok := <-ch
        if !ok {
            return nil  // Silently stops listening - no error reported
        }
        return conversationEventMsg{event: event}
    }
}
```

**Impact**:
- UI stops receiving events with no indication
- Stale data displayed to user
- Appears to work but becomes disconnected

**Suggested Fix**:
```go
type subscriptionClosedMsg struct {
    subscriptionType string
}

func waitForConversationEvent(ch <-chan pubsub.Event[services.Conversation]) tea.Cmd {
    return func() tea.Msg {
        event, ok := <-ch
        if !ok {
            return subscriptionClosedMsg{subscriptionType: "conversation"}
        }
        return conversationEventMsg{event: event}
    }
}

// In Update handler:
case subscriptionClosedMsg:
    m.ErrorMessage = fmt.Sprintf("%s subscription closed", msg.subscriptionType)
    // Attempt to resubscribe
    return m, m.StartEventSubscriptions()
```

---

### 9. Tool Execution Button Mashing Vulnerability
**Severity**: HIGH
**Location**: `model.go:770-813`
**Category**: Race Condition

**Description**: `ApproveToolUse()` doesn't check if tools are already executing. Rapid key presses could execute the same tools multiple times.

```go
func (m *Model) ApproveToolUse() tea.Cmd {
    if len(m.pendingToolUses) == 0 || m.toolExecutor == nil {
        // ... early return checks
    }

    // No check for m.executingTool == true!
    // User could press 'y' multiple times
    toolUses := m.pendingToolUses
    m.executingToolUses = toolUses
    m.executingTool = true
```

**Impact**:
- Same tool executed multiple times
- Dangerous for destructive operations
- Wasted API calls and money
- Unpredictable system state

**Suggested Fix**:
```go
func (m *Model) ApproveToolUse() tea.Cmd {
    // Guard against double-execution
    if m.executingTool {
        return nil
    }

    if len(m.pendingToolUses) == 0 || m.toolExecutor == nil {
        m.toolApprovalMode = false
        return nil
    }

    // ... rest of logic
```

---

### 10. Unchecked Channel Read in readStreamChunks
**Severity**: HIGH
**Location**: `update.go:989-1017`
**Category**: Nil Safety

**Description**: While there's a nil check for the channel, the code doesn't handle context cancellation properly. Reading from a channel after context is cancelled can block forever.

```go
func (m *Model) readStreamChunks(streamChan <-chan *core.StreamChunk) tea.Cmd {
    return func() tea.Msg {
        if streamChan == nil {
            // Handle nil case
        }

        // But what if context was cancelled? This blocks!
        chunk, ok := <-streamChan
```

**Impact**:
- Goroutine leak if context is cancelled
- UI freezes during cancellation
- Resource exhaustion over time

**Suggested Fix**:
```go
func (m *Model) readStreamChunks(streamChan <-chan *core.StreamChunk) tea.Cmd {
    // Capture context
    ctx := m.streamCtx

    return func() tea.Msg {
        if streamChan == nil || ctx == nil {
            return &StreamChunkMsg{
                Chunk: &core.StreamChunk{Type: "message_stop", Done: true},
            }
        }

        select {
        case chunk, ok := <-streamChan:
            if !ok {
                return &StreamChunkMsg{
                    Chunk: &core.StreamChunk{Type: "message_stop", Done: true},
                }
            }
            return &StreamChunkMsg{Chunk: chunk}
        case <-ctx.Done():
            return &StreamChunkMsg{
                Chunk: &core.StreamChunk{Type: "message_stop", Done: true},
            }
        }
    }
}
```

---

### 11. Message Cache Invalidation Missing
**Severity**: HIGH
**Location**: `model.go:348-379`
**Category**: Caching Bug

**Description**: The `RenderMessage` function caches rendered output but never invalidates the cache if the message content changes (which can happen during streaming or editing).

```go
func (m *Model) RenderMessage(msg *Message) (string, error) {
    // Performance: Use cached render if available
    if msg.renderedCache != "" {
        return msg.renderedCache, nil  // Returns stale cache!
    }
    // ... render and cache
}
```

**Impact**:
- Stale message content displayed
- Editing messages doesn't update view
- Confusing UX

**Suggested Fix**:
```go
// Add content hash for cache invalidation
type Message struct {
    // ... existing fields
    renderedCache string
    contentHash   uint64  // Hash of content for cache validation
}

func (m *Model) RenderMessage(msg *Message) (string, error) {
    // Compute content hash
    h := fnv.New64a()
    h.Write([]byte(msg.Content))
    currentHash := h.Sum64()

    // Check if cache is valid
    if msg.renderedCache != "" && msg.contentHash == currentHash {
        return msg.renderedCache, nil
    }

    // ... render
    msg.renderedCache = rendered
    msg.contentHash = currentHash
    return rendered, nil
}
```

---

### 12. Input Value Race During Autocomplete
**Severity**: HIGH
**Location**: `update.go:534-546`
**Category**: Race Condition

**Description**: The code reads `Input.Value()` twice (oldValue and newValue) but the textarea can be updated by concurrent events between these calls.

```go
oldValue := m.Input.Value()
m.Input, cmd = m.Input.Update(msg)
cmds = append(cmds, cmd)

// Phase 6C Task 4: Update autocomplete as user types
if m.autocomplete != nil && m.autocomplete.IsActive() {
    newValue := m.Input.Value()  // Could have changed again!
```

**Impact**:
- Autocomplete gets wrong value
- Suggestions don't match input
- Flickering autocomplete menu

**Suggested Fix**:
```go
oldValue := m.Input.Value()
m.Input, cmd = m.Input.Update(msg)
newValue := m.Input.Value()  // Read immediately after update
cmds = append(cmds, cmd)

// Use captured newValue consistently
if m.autocomplete != nil && m.autocomplete.IsActive() {
    if newValue != oldValue {
        m.autocomplete.Update(newValue)
    }
}

// Use same newValue for suggestions
if newValue != oldValue {
    m.AnalyzeSuggestions()
}
```

---

### 13. Escape Key Ambiguity
**Severity**: HIGH
**Location**: `update.go:258-274`
**Category**: UX Issue

**Description**: Escape key has multiple behaviors depending on state, but the priority order is unclear and could lead to confusing behavior.

```go
// Handle Esc key - dismiss suggestions or exit modes (does NOT quit)
if msg.Type == tea.KeyEsc {
    if m.showSuggestions {
        m.DismissSuggestions()
        return m, nil
    }
    if m.SearchMode {
        m.ExitSearchMode()
        return m, nil
    }
    if m.helpVisible {
        m.ToggleHelp()
        return m, nil
    }
    // What if multiple are true?
```

**Impact**:
- User expects to exit help but suggestions dismissed instead
- Non-obvious interaction priority
- Training users requires trial and error

**Suggested Fix**:
```go
// Document and enforce clear priority order
// Priority: Autocomplete > Suggestions > Search > Help
if msg.Type == tea.KeyEsc {
    // Highest priority: Active autocomplete
    if m.autocomplete != nil && m.autocomplete.IsActive() {
        m.autocomplete.Hide()
        return m, nil
    }

    // Second: Visible suggestions
    if m.showSuggestions {
        m.DismissSuggestions()
        return m, nil
    }

    // Third: Search mode
    if m.SearchMode {
        m.ExitSearchMode()
        return m, nil
    }

    // Lowest: Help panel
    if m.helpVisible {
        m.ToggleHelp()
        return m, nil
    }

    return m, nil
}
```

---

## MEDIUM Priority Issues

### 14. Inefficient String Building in renderIntroView
**Severity**: MEDIUM
**Location**: `view.go:86-142`
**Category**: Performance

**Description**: The function uses `strings.Builder` efficiently but then splits and re-joins strings unnecessarily for padding.

```go
// Pad each line of the logo
logoLines := strings.Split(logo, "\n")
for _, line := range logoLines {
    if strings.TrimSpace(line) != "" {
        b.WriteString(padding + line + "\n")  // String concat in loop
    }
}
```

**Suggested Fix**:
```go
// More efficient: write directly
for _, line := range logoLines {
    if strings.TrimSpace(line) != "" {
        b.WriteString(padding)
        b.WriteString(line)
        b.WriteByte('\n')
    }
}
```

---

### 15. Magic Numbers in Code
**Severity**: MEDIUM
**Location**: Multiple locations
**Category**: Code Quality

**Description**: Lots of magic numbers without named constants (16ms throttle, 60fps, 500 char truncation, etc.)

```go
if m.Streaming && timeSinceLastUpdate < 16*time.Millisecond {  // What's 16?
```

**Suggested Fix**:
```go
const (
    viewportRefreshRate = 60 * time.Millisecond  // 60 FPS
    maxToolOutputChars  = 500
    maxCodeBlockLines   = 30
    codeBlockContextLines = 5
)
```

---

### 16. Duplicate Tool Approval Logic
**Severity**: MEDIUM
**Location**: `model.go:1302-1351`
**Category**: Code Duplication

**Description**: The `handleApprovalResult` function has repetitive switch cases with similar logic for each decision type.

```go
switch msg.Result.Decision {
case forms.DecisionApprove:
    // ... 3 lines
    return m, m.ApproveToolUse()
case forms.DecisionDeny:
    // ... 3 lines
    return m, m.DenyToolUse()
case forms.DecisionAlwaysAllow:
    // ... 7 lines with rule storage
    return m, m.ApproveToolUse()
case forms.DecisionNeverAllow:
    // ... 7 lines with rule storage
    return m, m.DenyToolUse()
```

**Suggested Fix**:
```go
func (m *Model) handleApprovalResult(msg *forms.ApprovalResultMsg) (tea.Model, tea.Cmd) {
    // ... error handling

    decision := msg.Result.Decision

    // Handle persistent rules first
    if decision == forms.DecisionAlwaysAllow || decision == forms.DecisionNeverAllow {
        ruleType := approval.RuleAlwaysAllow
        if decision == forms.DecisionNeverAllow {
            ruleType = approval.RuleNeverAllow
        }

        if m.approvalRules != nil {
            if err := m.approvalRules.SetRule(msg.Result.ToolUse.Name, ruleType); err != nil {
                fmt.Fprintf(os.Stderr, "[APPROVAL_RULES] Failed to persist rule: %v\n", err)
            }
        }
    }

    // Execute based on decision
    if decision == forms.DecisionApprove || decision == forms.DecisionAlwaysAllow {
        return m, m.ApproveToolUse()
    }
    return m, m.DenyToolUse()
}
```

---

### 17. Complex Function: handleStreamChunk
**Severity**: MEDIUM
**Location**: `update.go:680-941`
**Category**: Code Quality

**Description**: This function is 261 lines long and handles too many responsibilities (error handling, tool assembly, text streaming, completion).

**Suggested Fix**: Break into smaller functions:
```go
func (m *Model) handleStreamChunk(msg *StreamChunkMsg) (tea.Model, tea.Cmd) {
    if msg.Error != nil {
        return m.handleStreamError(msg.Error)
    }

    chunk := msg.Chunk

    // Dispatch to specialized handlers
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
        return m.handleMessageStop(chunk)
    default:
        return m.handleUnknownChunk(chunk)
    }
}
```

---

### 18. Inconsistent Nil Checks
**Severity**: MEDIUM
**Location**: Multiple locations
**Category**: Code Quality

**Description**: Some nil checks use `!= nil`, others check fields, and some are missing entirely. Not consistent.

Examples:
```go
// Some places check:
if m.renderer != nil {

// Others check result:
renderer, err := glamour.NewTermRenderer(...)
if err != nil {
    renderer = nil  // Explicit nil
}

// Others don't check at all:
m.spinner.Stop()  // Could be nil
```

**Suggested Fix**: Establish and enforce a consistent pattern:
```go
// Option 1: Always check before use
if m.spinner != nil {
    m.spinner.Stop()
}

// Option 2: Make methods nil-safe
func (s *Spinner) Stop() {
    if s == nil {
        return
    }
    // ... actual stop logic
}
```

---

## LOW Priority Issues

### 19. Debug Logging to File in Production Code
**Severity**: LOW
**Location**: `update.go:145-149`, `update.go:156-159`
**Category**: Code Quality

**Description**: Debug logging opens files directly in update loop without proper error handling or log rotation.

```go
if f, err := os.OpenFile("/tmp/hex-approval-debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600); err == nil {
    _, _ = fmt.Fprintf(f, "[APPROVAL_KEY] received key: %v\n", msg.String())
    _ = f.Close()
}
```

**Impact**:
- Fills /tmp on long sessions
- Performance hit from file I/O in hot path
- Leaks file descriptors if Close() fails

**Suggested Fix**:
```go
// Use structured logging with proper levels
if os.Getenv("HEX_DEBUG") != "" {
    logging.Debug("approval key received", "key", msg.String())
}
```

---

### 20. Commented-Out Code
**Severity**: LOW
**Location**: `update.go:1095-1098`
**Category**: Code Quality

**Description**: Deprecated function left commented in production code.

```go
// func (m *Model) handleQuickActionsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
// 	// Quick actions are now handled by huh forms in LaunchQuickActionsForm()
// 	return m, nil
// }
```

**Suggested Fix**: Remove commented code. If history is needed, use git.

---

### 21. Missing Documentation for Complex State Transitions
**Severity**: LOW
**Location**: `model.go:530-614` (ClearContext)
**Category**: Documentation

**Description**: The `ClearContext()` function clears 20+ fields but doesn't document the state machine or what combinations are valid.

**Suggested Fix**:
```go
// ClearContext resets the conversation context and UI state (for /clear command)
// State transition: ANY -> Idle with empty context
//
// Clears:
// - All messages and streaming state
// - Tool execution state (pending, executing, results)
// - UI overlays (quick actions, suggestions, autocomplete)
// - Event streams and cancellation
//
// Preserves:
// - Services (convSvc, msgSvc, agentSvc)
// - Configuration (model, conversationID)
// - Registries (toolRegistry, approvalRules)
func (m *Model) ClearContext() {
```

---

## Recommendations Summary

### Immediate Actions (CRITICAL)
1. **Fix context race condition** in `sendToolResults()` - redesign context management
2. **Add nil safety** for approval rules with proper synchronization
3. **Make tool results visible** in UI - users need transparency
4. **Fix stream cancellation** to prevent memory leaks
5. **Prevent approval form message loops** with selective forwarding

### Short Term (HIGH)
6. Implement proper viewport throttling with deferred updates
7. Copy message slices before iteration to prevent corruption
8. Add subscription closed error handling with reconnection
9. Guard against tool execution button mashing
10. Fix stream chunk reading to respect context cancellation
11. Implement cache invalidation for message rendering
12. Fix input value race in autocomplete
13. Clarify and document Escape key priority

### Code Quality (MEDIUM + LOW)
14. Extract constants for magic numbers
15. Refactor `handleStreamChunk` into smaller functions
16. Standardize nil checking patterns
17. Remove debug file logging in favor of structured logging
18. Clean up commented code
19. Document complex state machines

---

## Testing Gaps

The test coverage is good for basic functionality but missing:

1. **Concurrency tests** - No tests for race conditions
2. **Error recovery tests** - What happens when subscriptions fail?
3. **State machine tests** - Valid and invalid state transitions
4. **Tool execution edge cases** - Multiple approvals, cancellations
5. **Stream interruption tests** - Network failures, cancellations
6. **Memory leak tests** - Long-running sessions

**Suggested Test Additions**:
```go
// Test concurrent stream cancellation
func TestStreamCancellationRace(t *testing.T)

// Test tool approval during execution
func TestToolApprovalWhileExecuting(t *testing.T)

// Test message rendering with cache
func TestMessageCacheInvalidation(t *testing.T)

// Test subscription reconnection
func TestSubscriptionRecovery(t *testing.T)
```

---

## Code Metrics

- **Total Lines**: ~3,400 (across 3 main files)
- **Longest Function**: `handleStreamChunk` (261 lines)
- **Cyclomatic Complexity**: High in Update() and handleStreamChunk()
- **File Count**: 3 core files + 20+ supporting files
- **Test Coverage**: ~60% (estimated, missing concurrency tests)

---

## Conclusion

The TUI implementation is feature-rich but has several serious issues related to concurrency, state management, and error handling. The critical issues should be addressed before production use, especially around stream cancellation and tool execution. The code would benefit from:

1. **Better separation of concerns** - Update() is too large
2. **Explicit state machine** - Document valid transitions
3. **Defensive programming** - More nil checks, error handling
4. **Testing for failure modes** - Not just happy paths

The architecture is sound overall, but needs hardening for production reliability.
