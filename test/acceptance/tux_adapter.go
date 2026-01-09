// ABOUTME: Tux adapter implementing TUIHarness
// ABOUTME: Wraps tux.App for acceptance testing

package acceptance

import (
	"context"
	"errors"
	"time"

	"github.com/2389-research/tux"
)

// errNotImplemented is returned by stub methods.
var errNotImplemented = errors.New("TuxAdapter not yet implemented")

// TuxAdapter wraps a tux.App to implement TUIHarness.
// This allows the same acceptance tests to run against both UIs.
type TuxAdapter struct {
	app    *tux.App
	agent  *MockAgent
	width  int
	height int
}

// MockAgent is a mock tux.Agent for testing.
type MockAgent struct {
	events chan tux.Event
}

// Run implements tux.Agent.
func (m *MockAgent) Run(ctx context.Context, prompt string) error {
	// Mock implementation - controlled by test
	return nil
}

// Subscribe implements tux.Agent.
func (m *MockAgent) Subscribe() <-chan tux.Event {
	return m.events
}

// Cancel implements tux.Agent.
func (m *MockAgent) Cancel() {
	// Mock implementation
}

// NewTuxAdapter creates a new TuxAdapter for acceptance testing.
func NewTuxAdapter() *TuxAdapter {
	return &TuxAdapter{
		agent: &MockAgent{
			events: make(chan tux.Event, 100),
		},
	}
}

// Init initializes the tux app with the given dimensions.
func (a *TuxAdapter) Init(width, height int) error {
	a.width = width
	a.height = height
	// TODO: Initialize tux app with mock agent
	// a.app = tux.New(a.agent)
	return errNotImplemented
}

// Shutdown cleans up the tux app.
func (a *TuxAdapter) Shutdown() {
	// TODO: Cleanup tux app if needed
}

// SendKey sends a key press to the tux app.
func (a *TuxAdapter) SendKey(key string) error {
	// TODO: Translate key string to tux key event and send to app
	return errNotImplemented
}

// SendText types text into the input area.
func (a *TuxAdapter) SendText(text string) error {
	// TODO: Send text input to tux app
	return errNotImplemented
}

// SubmitInput submits the current input.
func (a *TuxAdapter) SubmitInput() error {
	// TODO: Simulate Enter key press to submit
	return errNotImplemented
}

// SimulateStreamStart begins a mock streaming response.
func (a *TuxAdapter) SimulateStreamStart() error {
	// TODO: Send stream start event through mock agent
	return errNotImplemented
}

// SimulateStreamChunk sends a chunk of streaming text.
func (a *TuxAdapter) SimulateStreamChunk(text string) error {
	// TODO: Send text event through mock agent
	return errNotImplemented
}

// SimulateStreamEnd ends the streaming response.
func (a *TuxAdapter) SimulateStreamEnd() error {
	// TODO: Send stream complete event through mock agent
	return errNotImplemented
}

// SimulateToolCall simulates a tool call from the assistant.
func (a *TuxAdapter) SimulateToolCall(id, name string, params map[string]interface{}) error {
	// TODO: Send tool call event through mock agent
	return errNotImplemented
}

// SimulateToolResult simulates a tool result.
func (a *TuxAdapter) SimulateToolResult(id string, success bool, output string) error {
	// TODO: Send tool result event through mock agent
	return errNotImplemented
}

// GetView returns the current rendered view.
func (a *TuxAdapter) GetView() string {
	// TODO: Get rendered view from tux app
	return ""
}

// GetStatus returns the current status.
func (a *TuxAdapter) GetStatus() string {
	// TODO: Get status from tux app state
	return "unknown"
}

// IsStreaming returns whether the app is currently streaming.
func (a *TuxAdapter) IsStreaming() bool {
	// TODO: Check streaming state from tux app
	return false
}

// GetMessages returns the messages in the conversation.
func (a *TuxAdapter) GetMessages() []TestMessage {
	// TODO: Extract messages from tux app state
	return nil
}

// HasModal returns whether a modal is currently active.
func (a *TuxAdapter) HasModal() bool {
	// TODO: Check modal state from tux app
	return false
}

// GetModalType returns the type of the active modal.
func (a *TuxAdapter) GetModalType() string {
	// TODO: Get modal type from tux app state
	return ""
}

// WaitFor waits for a condition to be true within the timeout.
func (a *TuxAdapter) WaitFor(condition func() bool, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return nil
		}
		time.Sleep(10 * time.Millisecond)
	}
	return errors.New("timeout waiting for condition")
}

// Compile-time check that TuxAdapter implements TUIHarness.
var _ TUIHarness = (*TuxAdapter)(nil)
