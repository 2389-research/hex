// ABOUTME: types.go defines the core event types and structures for the event-sourcing system.
// ABOUTME: Events are recorded to disk for debugging, replay, and audit trails.

package events

import "time"

// EventType represents the type of event that occurred
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

// Event represents a single event in the system
type Event struct {
	ID        string      `json:"id"`        // UUID
	AgentID   string      `json:"agent_id"`  // Hierarchical (root, root.1, root.1.2)
	ParentID  string      `json:"parent_id"` // Parent agent ID
	Type      EventType   `json:"type"`      // Event type
	Timestamp time.Time   `json:"timestamp"` // When the event occurred
	Data      interface{} `json:"data"`      // Event-specific data
}
