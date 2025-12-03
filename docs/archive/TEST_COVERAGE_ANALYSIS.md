# Test Coverage Analysis - Tool Streaming Bug

**Date:** 2025-11-29

---

## Question: Do We Have Tests for the Bug We Just Fixed?

**Short Answer:** No, we don't have automated tests that would have caught this specific bug.

---

## What We Have

### 1. Tool Executor Level Tests (`test/integration/multi_tool_scenarios_test.go`)

These test the **tool execution system** in isolation:
- Batch execution of multiple tools
- Mixed tool types (read + write)
- Partial failures
- Tool denials
- Error recovery

**What They Don't Test:**
- Streaming from the API
- Message structure for API requests
- The UI layer's handleStreamChunk flow
- How streaming text and tool_use blocks are combined into messages

**Why They Missed the Bug:**
These tests bypass the UI layer entirely. They call `executor.Execute()` directly, never going through the streaming → message creation → API request cycle where the bug lived.

---

## What We Need (And Partially Created)

### 2. Streaming Message Structure Tests (`internal/ui/streaming_tool_test.go`)

Created but **not yet working** due to type mismatches. These would test:

**Scenario 1:** Text streams, then tool_use blocks arrive
- ✅ Should create ONE assistant message
- ✅ Message should have ContentBlocks: [text, tool_use, tool_use, ...]
- ❌ Should NOT create two separate assistant messages

**Scenario 2:** Multiple tool_use blocks without text
- ✅ Should create ONE assistant message
- ✅ Message should have ContentBlocks: [tool_use, tool_use, ...]

**Scenario 3:** Tool_use block arrives mid-stream
- ✅ StreamingText should NOT be committed when tool_use starts
- ✅ StreamingText should be combined with tool_use blocks at message_stop

**Status:** File created with compile errors - needs fixing:
- Replace `core.StreamDelta` → `core.Delta`
- Replace `*core.ContentBlock` → `*core.Content`
- Fix type compatibility issues

---

## The Gap

The bug was in **integration between three layers:**

1. **API Streaming** (core) - provides chunks
2. **UI Update Handler** (ui/update.go) - processes chunks
3. **Message Builder** (ui/update.go:577-625) - creates messages

Our existing tests cover #1 and parts of #2, but not the critical flow where:
- Text streams → tool_use detected → `CommitStreamingText()` called → TWO messages created

This is a **state management bug** that only shows up when you:
1. Stream a response with both text AND tool_use blocks
2. Process the stream through `handleStreamChunk`
3. Build messages for the API
4. Send those messages back

---

## Why Manual Testing is Currently Required

To properly test this bug fix, you need to:

1. Start the actual app (`./hex`)
2. Request multiple tools from Claude
3. Let the API stream back a response with text + tool_use blocks
4. Verify the app doesn't crash with a 400 error
5. Check `debug.log` to see message structure

**Automated Test Challenges:**
- Need to mock the entire Bubbletea event loop
- Need to mock API streaming responses
- Need to verify internal message state at specific points
- Complex setup with many dependencies

---

## Recommendation

### Short Term
Keep the manual testing approach:
```bash
./hex
> create 3 files: test1.txt, test2.txt, test3.txt
# Verify app continues working, check debug.log
```

### Medium Term
Fix the `streaming_tool_test.go` file to have proper regression tests:
- Correct the type issues (Delta vs StreamDelta, Content vs ContentBlock)
- Add integration with Bubbletea's test helpers if available
- Ensure tests verify the EXACT bug we fixed

### Long Term
Consider adding:
- Mock API client that can stream test responses
- Helper functions to simulate streaming scenarios
- Integration tests that run full request/response cycles

---

## Bottom Line

**Do we have tests?** Not automated ones that caught this specific bug.

**Should we?** Yes - this is a critical path that can easily regress.

**Can we?** Yes - but it requires more infrastructure than the quick tests we created.

**Priority?** Medium - manual testing works for now, but automation would prevent future regressions.
