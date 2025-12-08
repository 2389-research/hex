// ABOUTME: Test suite for state machine implementation
// ABOUTME: Validates state transitions, thread safety, and observer notifications
package orchestrator

import (
	"sync"
	"testing"
	"time"
)

// TestNewStateMachine verifies state machine initialization
func TestNewStateMachine(t *testing.T) {
	sm := NewStateMachine(StateIdle)

	if sm == nil {
		t.Fatal("NewStateMachine returned nil")
	}

	if sm.Current() != StateIdle {
		t.Errorf("Expected initial state %s, got %s", StateIdle, sm.Current())
	}
}

// TestValidTransitions verifies all valid state transitions work
func TestValidTransitions(t *testing.T) {
	tests := []struct {
		name    string
		from    AgentState
		to      AgentState
		wantErr bool
	}{
		// Valid transitions from plan
		{"Idle->Streaming", StateIdle, StateStreaming, false},
		{"Streaming->AwaitingApproval", StateStreaming, StateAwaitingApproval, false},
		{"Streaming->Complete", StateStreaming, StateComplete, false},
		{"Streaming->Error", StateStreaming, StateError, false},
		{"Streaming->Idle", StateStreaming, StateIdle, false}, // Allow Stop
		{"AwaitingApproval->ExecutingTool", StateAwaitingApproval, StateExecutingTool, false},
		{"AwaitingApproval->Error", StateAwaitingApproval, StateError, false},
		{"AwaitingApproval->Idle", StateAwaitingApproval, StateIdle, false}, // Allow Stop
		{"ExecutingTool->Streaming", StateExecutingTool, StateStreaming, false},
		{"ExecutingTool->Error", StateExecutingTool, StateError, false},
		{"ExecutingTool->Idle", StateExecutingTool, StateIdle, false}, // Allow Stop
		{"Error->Idle", StateError, StateIdle, false},
		{"Complete->Idle", StateComplete, StateIdle, false}, // Allow reset
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := NewStateMachine(tt.from)
			err := sm.Transition(tt.to, nil)

			if (err != nil) != tt.wantErr {
				t.Errorf("Transition() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil && sm.Current() != tt.to {
				t.Errorf("Expected state %s, got %s", tt.to, sm.Current())
			}
		})
	}
}

// TestInvalidTransitions_ReturnError verifies invalid transitions are rejected
func TestInvalidTransitions_ReturnError(t *testing.T) {
	tests := []struct {
		name string
		from AgentState
		to   AgentState
	}{
		// Invalid transitions (updated after allowing Stop->Idle from any state)
		{"Idle->Complete", StateIdle, StateComplete},
		{"Idle->Error", StateIdle, StateError},
		{"Idle->AwaitingApproval", StateIdle, StateAwaitingApproval},
		{"Idle->ExecutingTool", StateIdle, StateExecutingTool},
		{"Complete->Streaming", StateComplete, StateStreaming},
		{"Complete->Error", StateComplete, StateError},
		{"Complete->AwaitingApproval", StateComplete, StateAwaitingApproval},
		{"Complete->ExecutingTool", StateComplete, StateExecutingTool},
		{"Streaming->ExecutingTool", StateStreaming, StateExecutingTool},
		{"AwaitingApproval->Streaming", StateAwaitingApproval, StateStreaming},
		{"AwaitingApproval->Complete", StateAwaitingApproval, StateComplete},
		{"ExecutingTool->Complete", StateExecutingTool, StateComplete},
		{"ExecutingTool->AwaitingApproval", StateExecutingTool, StateAwaitingApproval},
		{"Error->Streaming", StateError, StateStreaming},
		{"Error->Complete", StateError, StateComplete},
		{"Error->AwaitingApproval", StateError, StateAwaitingApproval},
		{"Error->ExecutingTool", StateError, StateExecutingTool},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := NewStateMachine(tt.from)
			err := sm.Transition(tt.to, nil)

			if err == nil {
				t.Errorf("Expected error for invalid transition %s->%s, got nil", tt.from, tt.to)
			}

			// State should not have changed
			if sm.Current() != tt.from {
				t.Errorf("State changed on invalid transition: expected %s, got %s", tt.from, sm.Current())
			}
		})
	}
}

// TestConcurrentTransitions_ThreadSafe verifies thread safety
func TestConcurrentTransitions_ThreadSafe(t *testing.T) {
	sm := NewStateMachine(StateIdle)

	// Transition to streaming first
	if err := sm.Transition(StateStreaming, nil); err != nil {
		t.Fatalf("Initial transition failed: %v", err)
	}

	var wg sync.WaitGroup
	numGoroutines := 100

	// Try to transition concurrently from streaming
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Each goroutine tries to transition to a different valid state
			targets := []AgentState{StateAwaitingApproval, StateComplete, StateError}
			target := targets[id%len(targets)]

			// We don't care if it succeeds (only one should), just that it doesn't panic
			_ = sm.Transition(target, nil)
		}(i)
	}

	// Also read state concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = sm.Current()
		}()
	}

	wg.Wait()

	// Final state should be valid (one of the target states)
	finalState := sm.Current()
	validStates := map[AgentState]bool{
		StateAwaitingApproval: true,
		StateComplete:         true,
		StateError:            true,
		StateStreaming:        true, // Could still be streaming if all transitions failed
	}

	if !validStates[finalState] {
		t.Errorf("Final state %s is not valid", finalState)
	}
}

// TestObserverNotification verifies observers are called on transitions
func TestObserverNotification(t *testing.T) {
	sm := NewStateMachine(StateIdle)

	// Create a test observer
	observer := &testObserver{
		transitions: make([]transition, 0),
	}

	sm.AddObserver(observer)

	// Make a transition
	err := sm.Transition(StateStreaming, map[string]string{"key": "value"})
	if err != nil {
		t.Fatalf("Transition failed: %v", err)
	}

	// Wait briefly for observer notification
	time.Sleep(50 * time.Millisecond)

	// Verify observer was notified
	observer.mu.Lock()
	defer observer.mu.Unlock()

	if len(observer.transitions) != 1 {
		t.Fatalf("Expected 1 transition notification, got %d", len(observer.transitions))
	}

	trans := observer.transitions[0]
	if trans.from != StateIdle {
		t.Errorf("Expected from state %s, got %s", StateIdle, trans.from)
	}
	if trans.to != StateStreaming {
		t.Errorf("Expected to state %s, got %s", StateStreaming, trans.to)
	}
	if trans.data == nil {
		t.Error("Expected data to be passed to observer")
	}
}

// TestMultipleObservers verifies multiple observers all receive notifications
func TestMultipleObservers(t *testing.T) {
	sm := NewStateMachine(StateIdle)

	// Create multiple observers
	obs1 := &testObserver{transitions: make([]transition, 0)}
	obs2 := &testObserver{transitions: make([]transition, 0)}
	obs3 := &testObserver{transitions: make([]transition, 0)}

	sm.AddObserver(obs1)
	sm.AddObserver(obs2)
	sm.AddObserver(obs3)

	// Make a transition
	err := sm.Transition(StateStreaming, nil)
	if err != nil {
		t.Fatalf("Transition failed: %v", err)
	}

	// Wait for notifications
	time.Sleep(50 * time.Millisecond)

	// Verify all observers were notified
	observers := []*testObserver{obs1, obs2, obs3}
	for i, obs := range observers {
		obs.mu.Lock()
		if len(obs.transitions) != 1 {
			t.Errorf("Observer %d: expected 1 notification, got %d", i, len(obs.transitions))
		}
		obs.mu.Unlock()
	}
}

// TestStateHistory verifies history is recorded correctly
func TestStateHistory(t *testing.T) {
	history := &StateHistory{
		history: make([]StateTransition, 0),
	}

	// Simulate state transitions
	transitions := []struct {
		from AgentState
		to   AgentState
		data interface{}
	}{
		{StateIdle, StateStreaming, "start"},
		{StateStreaming, StateAwaitingApproval, "tool"},
		{StateAwaitingApproval, StateExecutingTool, "approve"},
		{StateExecutingTool, StateStreaming, "result"},
		{StateStreaming, StateComplete, "done"},
	}

	for _, trans := range transitions {
		history.OnStateChange(trans.from, trans.to, trans.data)
	}

	// Get history
	recorded := history.GetHistory()

	if len(recorded) != len(transitions) {
		t.Fatalf("Expected %d transitions in history, got %d", len(transitions), len(recorded))
	}

	// Verify each transition
	for i, expected := range transitions {
		actual := recorded[i]
		if actual.From != expected.from {
			t.Errorf("Transition %d: expected from %s, got %s", i, expected.from, actual.From)
		}
		if actual.To != expected.to {
			t.Errorf("Transition %d: expected to %s, got %s", i, expected.to, actual.To)
		}
		if actual.Data != expected.data {
			t.Errorf("Transition %d: expected data %v, got %v", i, expected.data, actual.Data)
		}
		if actual.Timestamp.IsZero() {
			t.Errorf("Transition %d: timestamp not set", i)
		}
	}
}

// TestLoggingObserver verifies logging observer doesn't panic
func TestLoggingObserver(t *testing.T) {
	obs := &LoggingObserver{}

	// Should not panic even without debug mode
	obs.OnStateChange(StateIdle, StateStreaming, nil)
	obs.OnStateChange(StateStreaming, StateComplete, map[string]string{"key": "value"})
}

// testObserver is a test implementation of StateObserver
type testObserver struct {
	mu          sync.Mutex
	transitions []transition
}

type transition struct {
	from AgentState
	to   AgentState
	data interface{}
}

func (o *testObserver) OnStateChange(from, to AgentState, data interface{}) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.transitions = append(o.transitions, transition{from, to, data})
}
