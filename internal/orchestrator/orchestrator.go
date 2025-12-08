// ABOUTME: Agent orchestrator for managing AI agent lifecycle
// ABOUTME: Handles streaming, tool execution, and state management independent of UI
package orchestrator

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/2389-research/hex/internal/core"
	"github.com/2389-research/hex/internal/events"
	"github.com/2389-research/hex/internal/tools"
	"github.com/google/uuid"
)

// EventType represents the type of orchestrator event
type EventType string

const (
	// EventStreamStart indicates stream has started
	EventStreamStart EventType = "stream_start"

	// EventStreamChunk indicates a chunk of streaming data
	EventStreamChunk EventType = "stream_chunk"

	// EventToolCall indicates a tool is being requested
	EventToolCall EventType = "tool_call"

	// EventToolResult indicates a tool has completed
	EventToolResult EventType = "tool_result"

	// EventComplete indicates stream/execution is complete
	EventComplete EventType = "complete"

	// EventError indicates an error occurred
	EventError EventType = "error"
)

// Event represents an event emitted by the orchestrator
type Event struct {
	Type      EventType
	Data      interface{}
	Timestamp time.Time
}

// APIClient is the interface for API streaming operations
type APIClient interface {
	CreateMessageStream(ctx context.Context, req core.MessageRequest) (<-chan *core.StreamChunk, error)
}

// ToolExecutor is the interface for tool execution operations
type ToolExecutor interface {
	Execute(ctx context.Context, toolName string, params map[string]interface{}) (*tools.Result, error)
}

// AgentOrchestrator manages the agent lifecycle independent of UI layer
type AgentOrchestrator struct {
	client          APIClient
	toolExecutor    ToolExecutor
	messageHistory  []core.Message
	stateMachine    *StateMachine
	stateHistory    *StateHistory
	eventChan       chan Event
	stopChan        chan struct{}
	mu              sync.RWMutex
	streamCtx       context.Context
	streamCancel    context.CancelFunc
	streamChan      <-chan *core.StreamChunk
	pendingToolUses []*core.ToolUse
	streamingText   string
	assemblingTool  *core.ToolUse
	toolInputBuf    string
	toolCtx         context.Context
	toolCancel      context.CancelFunc
}

// NewOrchestrator creates a new agent orchestrator
func NewOrchestrator(client APIClient, executor ToolExecutor) *AgentOrchestrator {
	// Initialize state machine with observers
	sm := NewStateMachine(StateIdle)
	history := &StateHistory{
		history: make([]StateTransition, 0),
	}

	sm.AddObserver(&LoggingObserver{})
	sm.AddObserver(history)

	return &AgentOrchestrator{
		client:         client,
		toolExecutor:   executor,
		stateMachine:   sm,
		stateHistory:   history,
		eventChan:      make(chan Event, 100), // Buffered for async emission
		stopChan:       make(chan struct{}),
		messageHistory: []core.Message{},
	}
}

// Start begins agent execution with the given prompt
func (o *AgentOrchestrator) Start(ctx context.Context, prompt string) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	// Record agent started event
	if store := events.Global(); store != nil {
		_ = store.Record(events.Event{
			ID:        uuid.New().String(),
			AgentID:   os.Getenv("HEX_AGENT_ID"),
			ParentID:  os.Getenv("HEX_PARENT_AGENT_ID"),
			Type:      events.EventAgentStarted,
			Timestamp: time.Now(),
			Data:      map[string]interface{}{"prompt": prompt},
		})
	}

	// Add user message to history
	o.messageHistory = append(o.messageHistory, core.Message{
		Role:    "user",
		Content: prompt,
	})

	// Transition to streaming state
	if err := o.stateMachine.Transition(StateStreaming, nil); err != nil {
		return fmt.Errorf("failed to transition to streaming state: %w", err)
	}

	// Emit stream start event
	o.emitEventLocked(EventStreamStart, nil)

	// Record stream started event
	if store := events.Global(); store != nil {
		_ = store.Record(events.Event{
			ID:        uuid.New().String(),
			AgentID:   os.Getenv("HEX_AGENT_ID"),
			ParentID:  os.Getenv("HEX_PARENT_AGENT_ID"),
			Type:      events.EventStreamStarted,
			Timestamp: time.Now(),
			Data:      nil,
		})
	}

	// Create cancellable context for stream
	o.streamCtx, o.streamCancel = context.WithCancel(ctx)

	// Create cancellable context for tools
	o.toolCtx, o.toolCancel = context.WithCancel(context.Background())

	// Start stream handling in background
	go o.handleStream(o.streamCtx)

	return nil
}

// HandleToolApproval processes tool approval/denial
func (o *AgentOrchestrator) HandleToolApproval(toolUseID string, approved bool) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	// Find the tool
	var toolUse *core.ToolUse
	for _, t := range o.pendingToolUses {
		if t.ID == toolUseID {
			toolUse = t
			break
		}
	}

	if toolUse == nil {
		return fmt.Errorf("tool use ID not found: %s", toolUseID)
	}

	if !approved {
		// Record tool denied event
		if store := events.Global(); store != nil {
			_ = store.Record(events.Event{
				ID:        uuid.New().String(),
				AgentID:   os.Getenv("HEX_AGENT_ID"),
				ParentID:  os.Getenv("HEX_PARENT_AGENT_ID"),
				Type:      events.EventToolCallDenied,
				Timestamp: time.Now(),
				Data:      map[string]interface{}{"tool_use_id": toolUseID, "tool_name": toolUse.Name},
			})
		}

		// Tool denied - emit error result
		result := &tools.Result{
			ToolName: toolUse.Name,
			Success:  false,
			Error:    "User denied permission",
		}
		o.emitEventLocked(EventToolResult, map[string]interface{}{
			"tool_use_id": toolUseID,
			"result":      result,
		})

		// Remove from pending
		o.removePendingTool(toolUseID)

		// If no more pending tools, transition to error state
		if len(o.pendingToolUses) == 0 {
			if err := o.stateMachine.Transition(StateError, "tool denied"); err != nil {
				return fmt.Errorf("failed to transition to error state: %w", err)
			}
		}

		return nil
	}

	// Record tool approved event
	if store := events.Global(); store != nil {
		_ = store.Record(events.Event{
			ID:        uuid.New().String(),
			AgentID:   os.Getenv("HEX_AGENT_ID"),
			ParentID:  os.Getenv("HEX_PARENT_AGENT_ID"),
			Type:      events.EventToolCallApproved,
			Timestamp: time.Now(),
			Data:      map[string]interface{}{"tool_use_id": toolUseID, "tool_name": toolUse.Name},
		})
	}

	// Tool approved - transition to executing
	if err := o.stateMachine.Transition(StateExecutingTool, toolUse.Name); err != nil {
		return fmt.Errorf("failed to transition to executing tool state: %w", err)
	}

	// Execute tool in background
	go o.executeToolAsync(toolUse)

	// Remove from pending
	o.removePendingTool(toolUseID)

	return nil
}

// Stop halts the orchestrator and cancels any active streams
func (o *AgentOrchestrator) Stop() error {
	o.mu.Lock()
	defer o.mu.Unlock()

	// Record agent stopped event
	if store := events.Global(); store != nil {
		_ = store.Record(events.Event{
			ID:        uuid.New().String(),
			AgentID:   os.Getenv("HEX_AGENT_ID"),
			ParentID:  os.Getenv("HEX_PARENT_AGENT_ID"),
			Type:      events.EventAgentStopped,
			Timestamp: time.Now(),
			Data:      nil,
		})
	}

	// Cancel stream if active
	if o.streamCancel != nil {
		o.streamCancel()
		o.streamCancel = nil
	}

	// Cancel tools if active
	if o.toolCancel != nil {
		o.toolCancel()
		o.toolCancel = nil
	}

	// Clear state
	o.streamCtx = nil
	o.streamChan = nil
	o.streamingText = ""
	o.assemblingTool = nil
	o.toolInputBuf = ""

	// Transition to idle from any state (Stop is a special case)
	if err := o.stateMachine.Transition(StateIdle, "stop"); err != nil {
		// Log but don't fail on state transition error during stop
		if os.Getenv("HEX_DEBUG") != "" {
			fmt.Fprintf(os.Stderr, "[ORCHESTRATOR] Warning: failed to transition to idle during stop: %v\n", err)
		}
	}

	return nil
}

// GetState returns the current agent state
func (o *AgentOrchestrator) GetState() AgentState {
	return o.stateMachine.Current()
}

// Subscribe returns a channel for receiving orchestrator events
func (o *AgentOrchestrator) Subscribe() <-chan Event {
	return o.eventChan
}

// emitEvent emits an event to subscribers (thread-safe)
func (o *AgentOrchestrator) emitEvent(typ EventType, data interface{}) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.emitEventLocked(typ, data)
}

// emitEventLocked emits an event while holding the lock (caller must hold lock)
func (o *AgentOrchestrator) emitEventLocked(typ EventType, data interface{}) {
	now := time.Now()

	// Emit to orchestrator's event channel
	select {
	case o.eventChan <- Event{Type: typ, Data: data, Timestamp: now}:
		// Event sent successfully
	case <-time.After(100 * time.Millisecond):
		// Event channel full, drop event (log warning if debug enabled)
		if os.Getenv("HEX_DEBUG") != "" {
			fmt.Fprintf(os.Stderr, "[ORCHESTRATOR] Warning: dropped event %s (channel full)\n", typ)
		}
	}

	// Also record to global event store for persistence
	if store := events.Global(); store != nil {
		agentID := os.Getenv("HEX_AGENT_ID")
		if agentID == "" {
			agentID = "root"
		}

		// Map orchestrator event types to global event types
		var eventType events.EventType
		switch typ {
		case EventStreamStart:
			eventType = events.EventStreamStarted
		case EventStreamChunk:
			eventType = events.EventStreamChunk
		case EventToolCall:
			eventType = events.EventToolCallRequested
		case EventToolResult:
			eventType = events.EventToolExecutionEnd
		case EventError:
			eventType = events.EventError
		case EventComplete:
			// Complete is not a global event type, skip
			return
		default:
			// Unknown event type, skip
			return
		}

		// Record event to global store
		_ = store.Record(events.Event{
			ID:        uuid.New().String(),
			AgentID:   agentID,
			ParentID:  "", // TODO: Get parent ID from context if available
			Type:      eventType,
			Timestamp: now,
			Data:      data,
		})
	}
}

// setState sets the current state using the state machine
// This is kept for backward compatibility but uses the state machine internally
func (o *AgentOrchestrator) setState(state AgentState) {
	if err := o.stateMachine.Transition(state, nil); err != nil {
		// Log error but don't panic - some callers may not handle errors
		if os.Getenv("HEX_DEBUG") != "" {
			fmt.Fprintf(os.Stderr, "[ORCHESTRATOR] Warning: setState failed: %v\n", err)
		}
	}
}

// addPendingTool adds a tool to pending list (for testing)
func (o *AgentOrchestrator) addPendingTool(toolUse *core.ToolUse) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.pendingToolUses = append(o.pendingToolUses, toolUse)
}

// removePendingTool removes a tool from pending list by ID (caller must hold lock)
func (o *AgentOrchestrator) removePendingTool(toolUseID string) {
	filtered := []*core.ToolUse{}
	for _, t := range o.pendingToolUses {
		if t.ID != toolUseID {
			filtered = append(filtered, t)
		}
	}
	o.pendingToolUses = filtered
}

// executeToolAsync executes a tool in the background and emits result
func (o *AgentOrchestrator) executeToolAsync(toolUse *core.ToolUse) {
	o.mu.RLock()
	ctx := o.toolCtx
	o.mu.RUnlock()

	// Record tool execution start
	if store := events.Global(); store != nil {
		_ = store.Record(events.Event{
			ID:        uuid.New().String(),
			AgentID:   os.Getenv("HEX_AGENT_ID"),
			ParentID:  os.Getenv("HEX_PARENT_AGENT_ID"),
			Type:      events.EventToolExecutionStart,
			Timestamp: time.Now(),
			Data:      map[string]interface{}{"tool_use_id": toolUse.ID, "tool_name": toolUse.Name},
		})
	}

	result, err := o.toolExecutor.Execute(ctx, toolUse.Name, toolUse.Input)

	if err != nil {
		result = &tools.Result{
			ToolName: toolUse.Name,
			Success:  false,
			Error:    err.Error(),
		}
	}

	// Record tool execution end
	if store := events.Global(); store != nil {
		_ = store.Record(events.Event{
			ID:        uuid.New().String(),
			AgentID:   os.Getenv("HEX_AGENT_ID"),
			ParentID:  os.Getenv("HEX_PARENT_AGENT_ID"),
			Type:      events.EventToolExecutionEnd,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"tool_use_id": toolUse.ID,
				"tool_name":   toolUse.Name,
				"success":     result.Success,
			},
		})
	}

	// Emit tool result event
	o.emitEvent(EventToolResult, map[string]interface{}{
		"tool_use_id": toolUse.ID,
		"result":      result,
	})

	// Atomically update state
	o.mu.Lock()
	defer o.mu.Unlock()

	// Remove completed tool from pending
	for i, pending := range o.pendingToolUses {
		if pending.ID == toolUse.ID {
			o.pendingToolUses = append(o.pendingToolUses[:i], o.pendingToolUses[i+1:]...)
			break
		}
	}

	// Determine next state based on current queue
	var nextState AgentState
	if o.streamCtx != nil && o.streamChan != nil {
		// Stream still active
		nextState = StateStreaming
	} else if len(o.pendingToolUses) == 0 {
		// No more tools pending - complete or error based on result
		if result.Success {
			nextState = StateComplete
		} else {
			nextState = StateError
		}
	} else {
		nextState = StateAwaitingApproval
	}

	if err := o.stateMachine.Transition(nextState, map[string]interface{}{
		"tool":   toolUse.Name,
		"result": result,
	}); err != nil {
		if os.Getenv("HEX_DEBUG") != "" {
			fmt.Fprintf(os.Stderr, "[ORCHESTRATOR] Warning: failed to transition after tool execution: %v\n", err)
		}
	}
}
