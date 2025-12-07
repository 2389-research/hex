# TUI Performance Analysis

## Executive Summary

Found **6 major performance bottlenecks** in the Bubble Tea TUI that cause significant overhead during streaming and large conversations.

## Critical Issues

### 1. **No Message Rendering Cache** ⚠️ HIGH IMPACT

**Problem**: Every viewport update re-renders ALL messages from scratch using expensive glamour markdown rendering.

**Evidence**:
```go
// internal/ui/update.go:596-628
for i, msg := range m.Messages {
    messageContent := msg.Content
    if msg.Role == "assistant" {
        rendered, err := m.RenderMessage(msg)  // ← Calls glamour EVERY TIME
        if err == nil {
            messageContent = strings.TrimSpace(rendered)
        }
    }
    // ... build content
}
```

**Impact**:
- O(n) complexity where n = total messages
- Glamour rendering is expensive (markdown parsing, ANSI styling)
- Called on every streaming chunk (potentially 100+ times per response)

**Estimated Cost**:
- 10 messages with average rendering time of 5ms = 50ms per viewport update
- During streaming at 10 chunks/sec = 500ms/sec of rendering overhead
- **50% CPU waste**

**Solution**: Add rendered message cache with invalidation

---

### 2. **Redundant Code Block Constraints** ⚠️ MEDIUM IMPACT

**Problem**: `constrainLongCodeBlocks()` parses and processes entire message content on every render.

**Evidence**:
```go
// internal/ui/model.go:372-470
func (m *Model) constrainLongCodeBlocks(content string) string {
    lines := strings.Split(content, "\n")
    // ... 100 lines of string processing
}
```

**Impact**:
- String splitting and parsing for every message on every update
- Redundant work when message content hasn't changed

**Solution**: Memoize constrained content in message cache

---

### 3. **Viewport Update Frequency** ⚠️ HIGH IMPACT

**Problem**: `updateViewport()` called excessively during streaming.

**Call Sites**:
- `AddMessage()` → `updateViewport()` (model.go:311)
- `AppendStreamingText()` → `updateViewport()` (update.go:765)
- `ApproveToolUse()` → `updateViewport()` (model.go:782)
- Tool results → `updateViewport()` (update.go:100, 127)
- Every streaming chunk → `updateViewport()` (update.go:765)

**Impact**:
- During fast streaming (50 chunks/sec), viewport updates 50 times/sec
- Each update re-renders all messages
- Causes visible UI lag and dropped frames

**Solution**: Debounce viewport updates or use incremental rendering

---

### 4. **No Incremental Streaming Display** ⚠️ MEDIUM IMPACT

**Problem**: Streaming text appends to buffer but full re-render happens each time.

**Evidence**:
```go
// internal/ui/update.go:759
m.AppendStreamingText(chunk.Delta.Text)
// ...
m.updateViewport()  // ← Full re-render for one character
```

**Impact**:
- Full viewport rebuild for single character additions
- Inefficient for character-by-character streaming

**Solution**: Use incremental append to viewport content

---

### 5. **Message Dump I/O During Production** ⚠️ LOW IMPACT (DEBUG ONLY)

**Problem**: `dumpMessages()` writes to stderr during tool operations.

**Evidence**:
```go
// internal/ui/model.go:1078, 1103
m.dumpMessages("BEFORE adding tool results")
m.dumpMessages("AFTER adding tool results")
```

**Impact**:
- I/O blocking during critical path
- Noisy logs

**Solution**: Guard with HEX_DEBUG check or remove from production

---

### 6. **Glamour Renderer Recreation** ⚠️ LOW IMPACT

**Problem**: Glamour renderer is created once but could be optimized further.

**Current State**: Renderer is created in `NewModel()` (model.go:219-226)

**Potential Improvement**: Use glamour cache settings

---

## Performance Optimization Recommendations

### Immediate Wins (High ROI)

1. **Add Rendered Message Cache**
   ```go
   type RenderedMessage struct {
       Content      string
       RenderedHTML string
       Version      int  // Increment when content changes
   }
   ```

2. **Debounce Viewport Updates**
   - Use a 16ms timer (60fps) to batch updates during streaming
   - Prevents excessive re-renders

3. **Incremental Streaming Append**
   - Append to viewport content directly instead of full rebuild
   - Only re-render when message is committed

### Medium-Term Improvements

4. **Lazy Message Rendering**
   - Only render messages visible in viewport
   - Use virtual scrolling for conversations > 50 messages

5. **Optimize `constrainLongCodeBlocks()`**
   - Cache constrained output
   - Use streaming parser instead of full string split

### Long-Term Improvements

6. **Virtual Scrolling**
   - Render only visible portion of conversation
   - Dramatically reduces memory and CPU for long conversations

7. **Web Assembly Markdown Renderer**
   - Consider faster markdown rendering (goldmark with WASM)

---

## Benchmarks Needed

To validate these optimizations, we should measure:

1. **Viewport Update Time**
   - Baseline: Current implementation
   - Target: <16ms for 60fps

2. **Message Rendering Time**
   - Per-message glamour rendering cost
   - Cache hit/miss ratio

3. **Streaming Throughput**
   - Messages/second before UI lag
   - Target: 50 chunks/sec without drops

4. **Memory Usage**
   - Conversation size vs memory footprint
   - Ensure cache doesn't leak

---

## Implementation Priority

1. **P0 (Critical)**: Message rendering cache
2. **P0 (Critical)**: Viewport update debouncing
3. **P1 (High)**: Incremental streaming append
4. **P2 (Medium)**: Remove debug dumpMessages() calls
5. **P3 (Nice-to-have)**: Lazy rendering / virtual scrolling

---

## Estimated Performance Gains

| Optimization | Current | Optimized | Improvement |
|--------------|---------|-----------|-------------|
| Viewport update during streaming | 50ms | 5ms | **10x faster** |
| Message rendering (10 msgs) | 50ms | 0ms (cached) | **∞ (instant)** |
| Streaming throughput | 10 chunks/sec | 100+ chunks/sec | **10x faster** |
| Memory overhead | Low | Medium | Cache adds ~10% |

**Total Expected Improvement**:
- **5-10x faster UI during streaming**
- **Eliminates visible lag in fast conversations**
- **Supports 10x larger conversations**

---

## Next Steps

1. Implement message rendering cache
2. Add viewport update debouncing
3. Benchmark before/after
4. Document cache invalidation strategy
5. Add performance metrics to status bar
