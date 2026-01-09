// ABOUTME: HexAgent implements tux.Agent interface
// ABOUTME: Wraps hex's API client to emit tux-compatible events

package tui

import (
	"context"
	"strings"
	"sync"

	"github.com/2389-research/hex/internal/core"
	"github.com/2389-research/tux"
)

// HexAgent implements tux.Agent by wrapping hex's API client.
type HexAgent struct {
	client       *core.Client
	model        string
	systemPrompt string

	// Conversation state
	messages []core.Message
	mu       sync.Mutex

	// Current run state
	events chan tux.Event
	cancel context.CancelFunc
}

// NewHexAgent creates a new HexAgent with the given API client.
func NewHexAgent(client *core.Client, model string, systemPrompt string) *HexAgent {
	if client == nil {
		panic("client cannot be nil")
	}
	return &HexAgent{
		client:       client,
		model:        model,
		systemPrompt: systemPrompt,
		messages:     make([]core.Message, 0),
	}
}

// Run starts the agent with the given prompt.
// It runs until completion or context cancellation.
func (a *HexAgent) Run(ctx context.Context, prompt string) error {
	// Create cancellable context
	ctx, cancel := context.WithCancel(ctx)
	a.mu.Lock()
	a.cancel = cancel
	a.mu.Unlock()

	defer func() {
		a.mu.Lock()
		if a.events != nil {
			close(a.events)
			a.events = nil
		}
		a.cancel = nil
		a.mu.Unlock()
	}()

	// Add user message to history
	a.mu.Lock()
	a.messages = append(a.messages, core.Message{
		Role:    "user",
		Content: prompt,
	})
	messages := make([]core.Message, len(a.messages))
	copy(messages, a.messages)
	a.mu.Unlock()

	// Build request
	req := core.MessageRequest{
		Model:     a.model,
		Messages:  messages,
		MaxTokens: 8192,
		Stream:    true,
		System:    a.systemPrompt,
	}

	// Start streaming
	chunks, err := a.client.CreateMessageStream(ctx, req)
	if err != nil {
		a.emit(tux.Event{Type: tux.EventError, Error: err})
		return err
	}

	// Process stream
	var responseText strings.Builder
	for chunk := range chunks {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Handle different chunk types
		switch chunk.Type {
		case "content_block_delta":
			if chunk.Delta != nil && chunk.Delta.Type == "text_delta" {
				responseText.WriteString(chunk.Delta.Text)
				a.emit(tux.Event{
					Type: tux.EventText,
					Text: chunk.Delta.Text,
				})
			}
		case "message_stop":
			// Add assistant response to history
			a.mu.Lock()
			a.messages = append(a.messages, core.Message{
				Role:    "assistant",
				Content: responseText.String(),
			})
			a.mu.Unlock()

			a.emit(tux.Event{Type: tux.EventComplete})
		}
	}

	return nil
}

// Subscribe returns a channel of events from the agent.
// The channel is closed when the agent completes.
func (a *HexAgent) Subscribe() <-chan tux.Event {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Close existing channel if any
	if a.events != nil {
		close(a.events)
	}

	// Create new events channel
	a.events = make(chan tux.Event, 100)
	return a.events
}

// Cancel cancels the current agent run.
func (a *HexAgent) Cancel() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.cancel != nil {
		a.cancel()
	}
}

// emit sends an event to subscribers.
func (a *HexAgent) emit(event tux.Event) {
	a.mu.Lock()
	ch := a.events
	a.mu.Unlock()

	if ch != nil {
		select {
		case ch <- event:
		default:
			// Channel full, drop event (shouldn't happen with buffered channel)
		}
	}
}

// AddSystemContext adds context to the system prompt.
func (a *HexAgent) AddSystemContext(context string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.systemPrompt += "\n\n" + context
}

// ClearHistory clears the conversation history.
func (a *HexAgent) ClearHistory() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.messages = make([]core.Message, 0)
}
