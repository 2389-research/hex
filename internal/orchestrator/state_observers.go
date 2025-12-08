// ABOUTME: State observer implementations for state machine
// ABOUTME: Includes logging observer and state history recorder
package orchestrator

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// LoggingObserver logs state transitions when debug mode is enabled
type LoggingObserver struct{}

// OnStateChange logs the state transition
func (l *LoggingObserver) OnStateChange(from, to AgentState, data interface{}) {
	if os.Getenv("HEX_DEBUG") != "" {
		fmt.Fprintf(os.Stderr, "[STATE] %s -> %s (data: %+v)\n", from, to, data)
	}
}

// StateHistory records all state transitions for debugging
type StateHistory struct {
	mu      sync.Mutex
	history []StateTransition
}

// StateTransition represents a single state transition
type StateTransition struct {
	From      AgentState
	To        AgentState
	Timestamp time.Time
	Data      interface{}
}

// OnStateChange records the state transition
func (h *StateHistory) OnStateChange(from, to AgentState, data interface{}) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.history = append(h.history, StateTransition{
		From:      from,
		To:        to,
		Timestamp: time.Now(),
		Data:      data,
	})
}

// GetHistory returns a copy of all recorded transitions
func (h *StateHistory) GetHistory() []StateTransition {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Return a copy to prevent external modification
	result := make([]StateTransition, len(h.history))
	copy(result, h.history)
	return result
}
