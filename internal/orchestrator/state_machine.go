// ABOUTME: Formal state machine for agent orchestrator
// ABOUTME: Provides validated state transitions, observers, and state history
package orchestrator

import (
	"fmt"
	"sync"
)

// StateObserver is notified when state transitions occur
type StateObserver interface {
	OnStateChange(from, to AgentState, data interface{})
}

// StateMachine manages state transitions with validation
type StateMachine struct {
	mu          sync.RWMutex
	current     AgentState
	transitions map[AgentState][]AgentState
	observers   []StateObserver
}

// NewStateMachine creates a state machine with the given initial state
func NewStateMachine(initial AgentState) *StateMachine {
	sm := &StateMachine{
		current: initial,
		transitions: map[AgentState][]AgentState{
			// Valid transitions from plan
			StateIdle:             {StateStreaming},
			StateStreaming:        {StateAwaitingApproval, StateComplete, StateError, StateIdle}, // Allow Idle for Stop
			StateAwaitingApproval: {StateExecutingTool, StateError, StateIdle},                   // Allow Idle for Stop
			StateExecutingTool:    {StateStreaming, StateError, StateIdle},                       // Allow Idle for Stop
			StateComplete:         {StateIdle},                                                   // Allow reset after completion
			StateError:            {StateIdle},                                                   // Can retry
		},
		observers: make([]StateObserver, 0),
	}
	return sm
}

// Transition attempts to transition to a new state
// Returns error if transition is invalid
func (sm *StateMachine) Transition(to AgentState, data interface{}) error {
	sm.mu.Lock()

	// Validate transition
	if !sm.canTransitionLocked(to) {
		from := sm.current
		sm.mu.Unlock()
		return fmt.Errorf("invalid state transition: %s -> %s", from, to)
	}

	// Perform transition
	from := sm.current
	sm.current = to

	// Get observers while holding lock
	observers := make([]StateObserver, len(sm.observers))
	copy(observers, sm.observers)

	sm.mu.Unlock()

	// Notify observers outside lock to prevent deadlocks
	for _, obs := range observers {
		obs.OnStateChange(from, to, data)
	}

	return nil
}

// Current returns the current state (thread-safe)
func (sm *StateMachine) Current() AgentState {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.current
}

// AddObserver registers a state change observer
func (sm *StateMachine) AddObserver(obs StateObserver) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.observers = append(sm.observers, obs)
}

// canTransitionLocked checks if transition to target state is valid
// Caller must hold lock
func (sm *StateMachine) canTransitionLocked(to AgentState) bool {
	validTransitions, exists := sm.transitions[sm.current]
	if !exists {
		return false
	}

	for _, valid := range validTransitions {
		if valid == to {
			return true
		}
	}

	return false
}
