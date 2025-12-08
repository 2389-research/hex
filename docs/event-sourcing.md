# Event-Sourcing System

## Overview

The event-sourcing system records all agent events to disk for debugging, replay, and audit trails. Events are stored in JSON Lines format (one event per line) for easy parsing and analysis.

## Architecture

### Event Types

The system tracks the following event types:

- `agent_started` - Agent begins execution
- `agent_stopped` - Agent completes execution
- `stream_started` - Streaming starts
- `stream_chunk` - Streaming chunk received
- `tool_call_requested` - Agent requests tool execution
- `tool_call_approved` - User approves tool execution
- `tool_call_denied` - User denies tool execution
- `tool_execution_start` - Tool execution begins
- `tool_execution_end` - Tool execution completes
- `state_transition` - State machine transition
- `error` - Error occurred

### Event Structure

Each event contains:

```json
{
  "id": "uuid",
  "agent_id": "root.1.2",
  "parent_id": "root.1",
  "type": "agent_started",
  "timestamp": "2025-12-07T23:23:42Z",
  "data": {
    "prompt": "example prompt"
  }
}
```

### Hierarchical Agent IDs

Agents are tracked with hierarchical IDs:

- `root` - Main agent
- `root.1` - First subagent of root
- `root.1.1` - First subagent of root.1
- `root.2` - Second subagent of root

This allows filtering events by agent tree.

## Usage

### Recording Events

Events are automatically recorded throughout the system:

```go
// In orchestrator
if store := events.Global(); store != nil {
    store.Record(events.Event{
        ID:        uuid.New().String(),
        AgentID:   os.Getenv("HEX_AGENT_ID"),
        ParentID:  os.Getenv("HEX_PARENT_AGENT_ID"),
        Type:      events.EventAgentStarted,
        Timestamp: time.Now(),
        Data:      map[string]interface{}{"prompt": prompt},
    })
}
```

### Environment Variables

- `HEX_AGENT_ID` - Current agent's hierarchical ID
- `HEX_PARENT_AGENT_ID` - Parent agent's ID

These are automatically set by the subagent framework.

### Event Store

The event store is initialized at application startup:

```go
eventStore, err := events.NewEventStore("hex_events.jsonl")
if err != nil {
    logging.WarnWith("Failed to create event store", "error", err)
} else {
    events.SetGlobal(eventStore)
    defer eventStore.Close()
}
```

### Retrieving Events

Get all events for an agent and its descendants:

```go
events := store.GetEvents("root.1")  // Gets root.1, root.1.1, root.1.2, etc.
```

### Replay Tool

The `hexreplay` CLI tool replays events from disk:

```bash
# Show all events
./hexreplay -events hex_events.jsonl

# Filter by agent
./hexreplay -events hex_events.jsonl -agent root.1

# Filter by event type
./hexreplay -events hex_events.jsonl -type tool_call_requested

# Verbose output (show event data)
./hexreplay -events hex_events.jsonl -v
```

## Implementation Details

### File Format

Events are stored in JSON Lines format (`.jsonl`):

```
{"id":"...", "agent_id":"root", "type":"agent_started", ...}
{"id":"...", "agent_id":"root", "type":"stream_started", ...}
{"id":"...", "agent_id":"root.1", "type":"agent_started", ...}
```

This format allows:
- Easy appending of new events
- Line-by-line parsing
- Standard tooling (jq, grep, etc.)

### Thread Safety

The event store uses `sync.RWMutex` for thread-safe concurrent access:

```go
func (s *EventStore) Record(event Event) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    // ...
}
```

### Persistence

Events are written to disk immediately on recording:

```go
line, err := json.Marshal(event)
if err != nil {
    return err
}
_, err = s.file.Write(append(line, '\n'))
```

### Loading Existing Events

When restarting, the store can load existing events:

```go
if err := store.LoadExisting(); err != nil {
    return err
}
```

## Integration Points

Events are recorded at these key locations:

1. **Orchestrator** (`internal/orchestrator/orchestrator.go`)
   - Agent start/stop
   - Tool approval/denial
   - Tool execution start/end

2. **Stream Handler** (`internal/orchestrator/stream_handler.go`)
   - Stream started
   - Tool call requested

3. **Subagent Executor** (`internal/subagents/executor.go`)
   - Sets up agent IDs for subagents

## Testing

Run event-sourcing tests:

```bash
go test ./internal/events/...
```

Tests cover:
- Event recording
- Hierarchical filtering
- Persistence to disk
- Concurrent recording
- Agent ID generation
- Integration scenarios

## Debugging

### Enable Debug Mode

Set `HEX_DEBUG=1` to see event recording in real-time.

### Query Events

Use standard tools to query events:

```bash
# Count events
wc -l hex_events.jsonl

# Find all tool calls
grep tool_call_requested hex_events.jsonl

# Pretty-print events
jq . hex_events.jsonl

# Filter by agent
grep '"agent_id":"root.1"' hex_events.jsonl
```

### Analyze Event Timeline

```bash
# Show event types in order
jq -r '.type' hex_events.jsonl

# Show agent activity
jq -r '"\(.timestamp) [\(.agent_id)] \(.type)"' hex_events.jsonl
```

## Future Enhancements

Potential improvements:

1. **Event Filtering** - More sophisticated queries
2. **Event Replay** - Replay events to reproduce behavior
3. **Event Visualization** - Timeline view of events
4. **Event Metrics** - Statistics and analytics
5. **Event Compression** - Compress old events
6. **Event Rotation** - Rotate event files by size/time

## See Also

- `internal/events/types.go` - Event type definitions
- `internal/events/store.go` - Event store implementation
- `internal/events/agent_id.go` - Agent ID generation
- `cmd/hexreplay/main.go` - Replay tool
