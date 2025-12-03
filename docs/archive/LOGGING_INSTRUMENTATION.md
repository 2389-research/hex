# Comprehensive Logging Instrumentation for Tool Execution Debugging

**Date:** 2025-11-29
**Purpose:** Debug potential tool loop issues by adding extensive logging throughout the tool execution flow

---

## Overview

This document describes the comprehensive logging and validation instrumentation added to track tool execution from streaming detection through approval, execution, and result sending back to the API.

The logging will help diagnose:
- If tools are being requested repeatedly
- If tool_use and tool_result blocks are properly matched
- If the message history structure is correct
- Where in the flow any issues occur

---

## Logging Points Added

### 1. Streaming: Tool Detection

**Location:** `internal/ui/update.go:474`

**Triggered:** When a tool_use content block starts streaming from the API

```
[STREAM_TOOL_START] tool_use detected: id=%s, name=%s
```

**What it tells us:** A tool request has been detected in the streaming response

---

### 2. Streaming: Tool Input Parsing

**Location:** `internal/ui/update.go:506-509`

**Triggered:** When the tool's JSON input parameters are parsed

```
[STREAM_TOOL_INPUT] parsed input for tool_use_id=%s, input=%+v
```

**Or on error:**

```
[STREAM_TOOL_INPUT_ERROR] failed to parse input JSON for tool_use_id=%s: %v
```

**What it tells us:** Whether the tool parameters were successfully extracted

---

### 3. Streaming: Tool Complete

**Location:** `internal/ui/update.go:514`

**Triggered:** When all tool parameters have been received and the tool_use block is complete

```
[STREAM_TOOL_COMPLETE] tool_use complete, storing as pending: id=%s, name=%s
```

**What it tells us:** The tool_use is fully assembled and stored as pending

---

### 4. Streaming: Message Stop

**Location:** `internal/ui/update.go:549`

**Triggered:** When the streaming message ends

```
[STREAM_STOP] message stream ended, pendingToolUse=%v
```

**What it tells us:** Whether the stream ended with a pending tool waiting for approval

---

### 5. Streaming: Assistant Message Creation

**Location:** `internal/ui/update.go:553-588`

**Triggered:** When creating the assistant message with tool_use content block

```
[STREAM_STOP_WITH_TOOL] creating assistant message with tool_use: id=%s, name=%s
[STREAM_STOP_WITH_TOOL] including text block (%d chars)
[STREAM_STOP_WITH_TOOL] added tool_use block to assistant message
[STREAM_STOP_WITH_TOOL] assistant message added to history (total messages: %d)
[STREAM_STOP_WITH_TOOL] enabling tool approval mode
```

**What it tells us:**
- The assistant message is being constructed correctly
- Text and tool_use blocks are being added
- The message was added to history
- Tool approval mode is being enabled

**Also triggers:** `dumpMessages("AFTER stream completion with tool_use")`

---

### 6. Tool Approval

**Location:** `internal/ui/model.go:454-457`

**Triggered:** When user approves a tool

```
[TOOL_APPROVAL] tool_use_id=%s, tool=%s
[VALIDATION] Looking for tool_use with ID: %s
[VALIDATION] ✓ Found tool_use at message[%d].ContentBlock[%d]
```

**Or on validation failure:**

```
[VALIDATION] ✗ WARNING: tool_use with ID %s NOT FOUND in message history!
```

**What it tells us:**
- Which tool is being approved
- Whether the tool_use block exists in message history before approval

---

### 7. Tool Execution Start

**Location:** `internal/ui/model.go:473`

**Triggered:** When tool execution begins in background

```
[TOOL_EXEC_START] tool_use_id=%s, tool=%s
```

**What it tells us:** The tool has started executing

---

### 8. Tool Execution Complete

**Location:** `internal/ui/model.go:476`

**Triggered:** When tool execution finishes (before returning result)

```
[TOOL_EXEC_DONE] tool_use_id=%s, success=%v
```

**What it tells us:** Whether the tool execution succeeded or failed

---

### 9. Tool Result Received

**Location:** `internal/ui/update.go:41`

**Triggered:** When the tool execution result message is received by the UI

```
[TOOL_RESULT_RECEIVED] tool_use_id=%s, err=%v
```

**What it tells us:** The tool execution result has been received by the update loop

---

### 10. Tool Result Processing

**Location:** `internal/ui/update.go:52-67`

**Triggered:** When processing tool results (success or error)

**On error:**
```
[TOOL_RESULT_ERROR] tool_use_id=%s, error=%s
```

**On success:**
```
[TOOL_RESULT_SUCCESS] tool_use_id=%s, storing result
[VALIDATION] Looking for tool_use with ID: %s
[VALIDATION] ✓ Found tool_use at message[%d].ContentBlock[%d]
[TOOL_RESULTS_QUEUE] current queue length: %d
```

**What it tells us:**
- Whether result processing succeeded
- Whether the tool_use block exists before storing result
- How many results are queued to send back to API

---

### 11. Tool Results Sending

**Location:** `internal/ui/update.go:77`

**Triggered:** Before calling sendToolResults()

```
[TOOL_RESULTS_SENDING] about to send %d tool results back to API
```

**What it tells us:** How many tool results are being sent to the API

**Also triggers:** `dumpMessages("BEFORE adding tool results")` (in sendToolResults function)

---

### 12. Tool Denial

**Location:** `internal/ui/model.go:503-510`

**Triggered:** When user denies a tool

```
[TOOL_DENIAL] tool_use_id=%s, tool=%s
[VALIDATION] Looking for tool_use with ID: %s
[VALIDATION] ✓ Found tool_use at message[%d].ContentBlock[%d]
```

**What it tells us:**
- Which tool was denied
- Whether the tool_use block exists in history

---

## Message Dumps

### dumpMessages() Function

**Location:** `internal/ui/model.go:658-692`

**Triggers:**
1. `BEFORE adding tool results` - Before tool_result blocks are added to messages
2. `AFTER adding tool results` - After tool_result blocks are added
3. `AFTER stream completion with tool_use` - After assistant message with tool_use is added

**Output Format:**
```
========== MESSAGE DUMP: [label] ==========
[0] Role: user
    Content (string): "..."
[1] Role: assistant
    ContentBlocks (2):
      [0] Type: text
          Text: "..."
      [1] Type: tool_use
          ID: toolu_...
          Name: read_file
          Input: map[file_path:/path/to/file]
[2] Role: user
    ContentBlocks (1):
      [0] Type: tool_result
          ToolUseID: toolu_...
          Content: "..."
========================================
```

**What it tells us:**
- Complete message history structure
- All content blocks with their types and data
- Proper nesting and formatting of tool_use and tool_result blocks

---

## Validation Function

### validateToolUseExists()

**Location:** `internal/ui/model.go:694-712`

**Purpose:** Verify that a tool_use block with a given ID exists in message history

**Called:**
- Before tool approval
- Before storing tool results (success case)
- Before sending tool denial results

**Output:**
- Success: `[VALIDATION] ✓ Found tool_use at message[%d].ContentBlock[%d]`
- Failure: `[VALIDATION] ✗ WARNING: tool_use with ID %s NOT FOUND in message history!`

**What it tells us:** Whether the conversation structure is valid according to Anthropic API requirements

---

## How to Use This Logging

### Running with Logging

1. Build and run hex:
   ```bash
   go build ./cmd/hex
   ./hex 2>debug.log
   ```

2. The main UI will display on stdout, debug logs will go to stderr (debug.log)

### Analyzing a Tool Loop

Look for this sequence:

1. **Normal flow:**
   ```
   [STREAM_TOOL_START] tool_use detected
   [STREAM_TOOL_COMPLETE] tool_use complete
   [STREAM_STOP_WITH_TOOL] creating assistant message
   MESSAGE DUMP: AFTER stream completion with tool_use
   [TOOL_APPROVAL] tool_use_id=...
   [VALIDATION] ✓ Found tool_use
   [TOOL_EXEC_START] tool_use_id=...
   [TOOL_EXEC_DONE] tool_use_id=..., success=true
   [TOOL_RESULT_RECEIVED] tool_use_id=...
   [TOOL_RESULT_SUCCESS] tool_use_id=...
   [VALIDATION] ✓ Found tool_use
   [TOOL_RESULTS_SENDING] about to send 1 tool results
   MESSAGE DUMP: BEFORE adding tool results
   MESSAGE DUMP: AFTER adding tool results
   ```

2. **If tool loops:**
   - Look for repeated `[STREAM_TOOL_START]` with same tool_use_id
   - Check if `tool_result` blocks are properly formatted in message dumps
   - Check if validation warnings appear
   - Look for mismatched IDs between tool_use and tool_result

3. **Check message structure:**
   - Examine the "AFTER adding tool results" dump
   - Verify the user message contains tool_result blocks (not plain text)
   - Verify each tool_result has matching tool_use_id from previous assistant message

---

## Files Modified

- `internal/ui/update.go` - Added os import, streaming logging, tool result logging
- `internal/ui/model.go` - Added validation function, dumpMessages function, approval logging

---

## Next Steps

1. User should run the application with logging enabled
2. Trigger a tool execution
3. Examine the debug.log file for the complete flow
4. If tool loop occurs, compare the log output to expected flow above
5. Look for validation warnings or malformed message structures

---

**Status:** ✅ Code compiles successfully
**Ready for testing:** Yes
