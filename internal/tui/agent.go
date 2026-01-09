// ABOUTME: HexAgent implements tux.Agent interface
// ABOUTME: Wraps hex's API client to emit tux-compatible events

package tui

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/2389-research/hex/internal/core"
	"github.com/2389-research/hex/internal/tools"
	"github.com/2389-research/tux"
)

// HexAgent implements tux.Agent by wrapping hex's API client.
type HexAgent struct {
	client       *core.Client
	model        string
	systemPrompt string

	// Tool execution
	executor *tools.Executor

	// Tool parsing state (within a single Run)
	assemblingTool   *core.ToolUse
	toolInputJSONBuf strings.Builder
	pendingTools     []*core.ToolUse

	// Conversation state
	messages []core.Message
	mu       sync.Mutex

	// Current run state
	events chan tux.Event
	cancel context.CancelFunc
}

// NewHexAgent creates a new HexAgent with the given API client and tool executor.
func NewHexAgent(client *core.Client, model string, systemPrompt string, executor *tools.Executor) *HexAgent {
	if client == nil {
		panic("client cannot be nil")
	}
	return &HexAgent{
		client:       client,
		model:        model,
		systemPrompt: systemPrompt,
		messages:     make([]core.Message, 0),
		executor:     executor,
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
		case "content_block_start":
			a.handleContentBlockStart(chunk)

		case "content_block_delta":
			if chunk.Delta != nil {
				switch chunk.Delta.Type {
				case "text_delta":
					responseText.WriteString(chunk.Delta.Text)
					a.emit(tux.Event{
						Type: tux.EventText,
						Text: chunk.Delta.Text,
					})
				case "input_json_delta":
					// Accumulate tool parameter JSON
					if a.assemblingTool != nil {
						a.toolInputJSONBuf.WriteString(chunk.Delta.PartialJSON)
					}
				}
			}

		case "content_block_stop":
			// Will be implemented in Task 5
			a.handleContentBlockStop()

		case "message_stop":
			// Add assistant response to history if there's text
			if responseText.Len() > 0 {
				a.mu.Lock()
				a.messages = append(a.messages, core.Message{
					Role:    "assistant",
					Content: responseText.String(),
				})
				a.mu.Unlock()
			}

			// Process any pending tools
			if len(a.pendingTools) > 0 {
				if err := a.processTools(ctx); err != nil {
					return err
				}
			} else {
				a.emit(tux.Event{Type: tux.EventComplete})
			}
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

// resetToolState clears tool parsing state for a new run.
func (a *HexAgent) resetToolState() {
	a.assemblingTool = nil
	a.toolInputJSONBuf.Reset()
	a.pendingTools = nil
}

// handleContentBlockStart processes a content_block_start chunk.
func (a *HexAgent) handleContentBlockStart(chunk *core.StreamChunk) {
	if chunk.ContentBlock == nil {
		return
	}

	if chunk.ContentBlock.Type == "tool_use" {
		a.assemblingTool = &core.ToolUse{
			Type:  "tool_use",
			ID:    chunk.ContentBlock.ID,
			Name:  chunk.ContentBlock.Name,
			Input: make(map[string]interface{}),
		}
		a.toolInputJSONBuf.Reset()

		// Emit tool call event (parameters will come in deltas)
		a.emit(tux.Event{
			Type:       tux.EventToolCall,
			ToolID:     chunk.ContentBlock.ID,
			ToolName:   chunk.ContentBlock.Name,
			ToolParams: nil, // Params not yet available
		})
	}
}

// handleContentBlockStop processes a content_block_stop chunk.
func (a *HexAgent) handleContentBlockStop() {
	if a.assemblingTool == nil {
		return
	}

	// Parse accumulated JSON into Input map
	jsonStr := a.toolInputJSONBuf.String()
	if jsonStr != "" {
		if err := json.Unmarshal([]byte(jsonStr), &a.assemblingTool.Input); err != nil {
			// Log error but continue - malformed params
			a.emit(tux.Event{
				Type:  tux.EventError,
				Error: fmt.Errorf("parse tool params: %w", err),
			})
		}
	}

	// Add to pending tools
	a.pendingTools = append(a.pendingTools, a.assemblingTool)
	a.assemblingTool = nil
	a.toolInputJSONBuf.Reset()
}

// processTools handles pending tools: approval, execution, results.
func (a *HexAgent) processTools(ctx context.Context) error {
	var toolResults []tools.ToolResult

	for _, tool := range a.pendingTools {
		// Check if tool needs approval
		needsApproval := true
		if a.executor != nil {
			// Could check executor's approval rules here
			// For now, always require approval
		}

		var approved bool
		if needsApproval {
			// Request approval via event
			decision, err := a.requestApproval(ctx, tool)
			if err != nil {
				return err
			}
			approved = (decision == tux.DecisionApprove || decision == tux.DecisionAlwaysAllow)
		} else {
			approved = true
		}

		if approved {
			// Execute tool
			result := a.executeTool(ctx, tool)
			toolResults = append(toolResults, result)

			// Emit result event
			a.emit(tux.Event{
				Type:       tux.EventToolResult,
				ToolID:     tool.ID,
				ToolName:   tool.Name,
				ToolOutput: result.Content,
				Success:    !result.IsError,
			})
		} else {
			// Tool denied
			toolResults = append(toolResults, tools.ToolResult{
				Type:      "tool_result",
				ToolUseID: tool.ID,
				Content:   "Tool execution denied by user",
				IsError:   true,
			})

			a.emit(tux.Event{
				Type:       tux.EventToolResult,
				ToolID:     tool.ID,
				ToolName:   tool.Name,
				ToolOutput: "Tool execution denied by user",
				Success:    false,
			})
		}
	}

	// Clear pending tools
	a.pendingTools = nil

	// Continue conversation with tool results
	return a.continueWithToolResults(ctx, toolResults)
}

// requestApproval emits an approval event and waits for user decision.
// Stub - will be implemented in Task 7.
func (a *HexAgent) requestApproval(ctx context.Context, tool *core.ToolUse) (tux.ApprovalDecision, error) {
	// TODO: Implement in Task 7
	return tux.DecisionApprove, nil // Auto-approve for now
}

// executeTool runs a tool and returns the result.
// Stub - will be implemented in Task 8.
func (a *HexAgent) executeTool(ctx context.Context, tool *core.ToolUse) tools.ToolResult {
	// TODO: Implement in Task 8
	return tools.ToolResult{
		Type:      "tool_result",
		ToolUseID: tool.ID,
		Content:   "Tool execution not yet implemented",
		IsError:   true,
	}
}

// continueWithToolResults sends tool results to API and resumes streaming.
// Stub - will be implemented in Task 9.
func (a *HexAgent) continueWithToolResults(ctx context.Context, results []tools.ToolResult) error {
	// TODO: Implement in Task 9
	a.emit(tux.Event{Type: tux.EventComplete})
	return nil
}
