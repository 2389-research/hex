// ABOUTME: Bubbletea adapter implementing TUIHarness
// ABOUTME: Wraps current ui.Model for acceptance testing

package acceptance

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/2389-research/hex/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

// BubbleteaAdapter wraps the current ui.Model to implement TUIHarness
type BubbleteaAdapter struct {
	model  *ui.Model
	width  int
	height int
}

// NewBubbleteaAdapter creates a new adapter for acceptance testing
func NewBubbleteaAdapter() *BubbleteaAdapter {
	return &BubbleteaAdapter{}
}

func (a *BubbleteaAdapter) Init(width, height int) error {
	a.width = width
	a.height = height
	a.model = ui.NewModel("test-conv-id", "claude-sonnet-4-5-20250929")

	// Send window size to initialize
	msg := tea.WindowSizeMsg{Width: width, Height: height}
	updatedModel, _ := a.model.Update(msg)
	a.model = updatedModel.(*ui.Model)

	return nil
}

func (a *BubbleteaAdapter) Shutdown() {
	// Cleanup if needed
}

func (a *BubbleteaAdapter) SendKey(key string) error {
	var msg tea.KeyMsg

	switch key {
	case KeyEnter:
		msg = tea.KeyMsg{Type: tea.KeyEnter}
	case KeyEsc:
		msg = tea.KeyMsg{Type: tea.KeyEsc}
	case KeyCtrlC:
		msg = tea.KeyMsg{Type: tea.KeyCtrlC}
	case KeyCtrlO:
		msg = tea.KeyMsg{Type: tea.KeyCtrlO}
	case KeyUp:
		msg = tea.KeyMsg{Type: tea.KeyUp}
	case KeyDown:
		msg = tea.KeyMsg{Type: tea.KeyDown}
	case KeyTab:
		msg = tea.KeyMsg{Type: tea.KeyTab}
	case KeyG:
		msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}}
	case KeyShiftG:
		msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}}
	default:
		// Single character
		if len(key) == 1 {
			msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
		} else {
			return fmt.Errorf("unknown key: %s", key)
		}
	}

	updatedModel, _ := a.model.Update(msg)
	a.model = updatedModel.(*ui.Model)
	return nil
}

func (a *BubbleteaAdapter) SendText(text string) error {
	// Type each character into the input
	for _, r := range text {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
		updatedModel, _ := a.model.Update(msg)
		a.model = updatedModel.(*ui.Model)
	}
	return nil
}

func (a *BubbleteaAdapter) SubmitInput() error {
	return a.SendKey(KeyEnter)
}

func (a *BubbleteaAdapter) SimulateStreamStart() error {
	a.model.Streaming = true
	a.model.SetStatus(ui.StatusStreaming)
	// Add placeholder message
	a.model.Messages = append(a.model.Messages, ui.Message{
		Role:    "assistant",
		Content: "",
	})
	a.model.ForceUpdateViewport()
	return nil
}

func (a *BubbleteaAdapter) SimulateStreamChunk(text string) error {
	a.model.AppendStreamingText(text)
	a.model.ForceUpdateViewport()
	return nil
}

func (a *BubbleteaAdapter) SimulateStreamEnd() error {
	a.model.CommitStreamingText()
	a.model.Streaming = false
	a.model.SetStatus(ui.StatusIdle)
	a.model.ForceUpdateViewport()
	return nil
}

func (a *BubbleteaAdapter) SimulateToolCall(id, name string, params map[string]interface{}) error {
	// Queue a tool for approval
	a.model.SetStatus(ui.StatusQueued)
	// The actual tool queuing is more complex; for now simulate the state
	return nil
}

func (a *BubbleteaAdapter) SimulateToolResult(id string, success bool, output string) error {
	// Simulate tool result being added
	return nil
}

func (a *BubbleteaAdapter) GetView() string {
	return a.model.View()
}

func (a *BubbleteaAdapter) GetStatus() string {
	switch a.model.Status {
	case ui.StatusIdle:
		return "idle"
	case ui.StatusStreaming:
		return "streaming"
	case ui.StatusQueued:
		return "queued"
	case ui.StatusError:
		return "error"
	default:
		return "unknown"
	}
}

func (a *BubbleteaAdapter) IsStreaming() bool {
	return a.model.Streaming
}

func (a *BubbleteaAdapter) GetMessages() []TestMessage {
	msgs := make([]TestMessage, len(a.model.Messages))
	for i, m := range a.model.Messages {
		msgs[i] = TestMessage{
			Role:    m.Role,
			Content: m.Content,
		}
	}
	return msgs
}

func (a *BubbleteaAdapter) HasModal() bool {
	return a.model.HasActiveOverlay()
}

func (a *BubbleteaAdapter) GetModalType() string {
	if !a.model.HasActiveOverlay() {
		return ""
	}
	// Return overlay type based on what's active
	return "overlay"
}

func (a *BubbleteaAdapter) WaitFor(condition func() bool, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return nil
		}
		time.Sleep(10 * time.Millisecond)
	}
	return errors.New("timeout waiting for condition")
}

// ViewContains checks if the rendered view contains the given text
func ViewContains(h TUIHarness, text string) bool {
	return strings.Contains(h.GetView(), text)
}

// ViewContainsAny checks if the rendered view contains any of the given texts
func ViewContainsAny(h TUIHarness, texts ...string) bool {
	view := h.GetView()
	for _, text := range texts {
		if strings.Contains(view, text) {
			return true
		}
	}
	return false
}
