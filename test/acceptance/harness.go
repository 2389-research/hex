// ABOUTME: TUI acceptance test harness interface
// ABOUTME: Abstracts TUI implementation for portable acceptance tests

package acceptance

import (
	"time"
)

// TUIHarness abstracts a TUI implementation for acceptance testing.
// Both the current Bubbletea UI and future tux UI implement this interface.
type TUIHarness interface {
	// Lifecycle
	Init(width, height int) error
	Shutdown()

	// Input
	SendKey(key string) error     // "enter", "ctrl+c", "esc", "up", "down", "g", etc.
	SendText(text string) error   // Type text into input area
	SubmitInput() error           // Submit current input (like pressing Enter)

	// Simulation (for testing without real API)
	SimulateStreamStart() error
	SimulateStreamChunk(text string) error
	SimulateStreamEnd() error
	SimulateToolCall(id, name string, params map[string]interface{}) error
	SimulateToolResult(id string, success bool, output string) error

	// Observation
	GetView() string          // Current rendered view
	GetStatus() string        // Current status (idle, streaming, etc.)
	IsStreaming() bool
	GetMessages() []TestMessage // Messages in conversation
	HasModal() bool             // Is a modal/overlay active?
	GetModalType() string       // Type of active modal if any

	// Waiting
	WaitFor(condition func() bool, timeout time.Duration) error
}

// TestMessage represents a message for test assertions
type TestMessage struct {
	Role    string // "user", "assistant"
	Content string
}

// Common key constants for readability
const (
	KeyEnter  = "enter"
	KeyEsc    = "esc"
	KeyCtrlC  = "ctrl+c"
	KeyCtrlO  = "ctrl+o"
	KeyUp     = "up"
	KeyDown   = "down"
	KeyTab    = "tab"
	KeyG      = "g"
	KeyShiftG = "G"
)
