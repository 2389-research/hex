// ABOUTME: HexAgent implements tux.Agent interface
// ABOUTME: Wraps hex's API client to emit tux-compatible events

package tui

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/2389-research/hex/internal/core"
	"github.com/2389-research/hex/internal/tools"
	"github.com/2389-research/tux"
	"github.com/google/uuid"
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

	// Session management
	storage        *SessionStorage
	currentSession *Session
}

// NewHexAgent creates a new HexAgent with the given API client and tool executor.
// The storage parameter is optional; if nil, sessions will not be persisted.
func NewHexAgent(client *core.Client, model string, systemPrompt string, executor *tools.Executor, storage *SessionStorage) *HexAgent {
	if client == nil {
		panic("client cannot be nil")
	}
	return &HexAgent{
		client:       client,
		model:        model,
		systemPrompt: systemPrompt,
		messages:     make([]core.Message, 0),
		executor:     executor,
		storage:      storage,
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

	// Reset tool state for new run
	a.resetToolState()

	// Initialize session if needed
	a.mu.Lock()
	if a.currentSession == nil {
		now := time.Now()
		a.currentSession = &Session{
			ID:        uuid.New().String(),
			CreatedAt: now,
			UpdatedAt: now,
			Messages:  make([]SessionMessage, 0),
		}
	}

	// Add user message to history
	a.messages = append(a.messages, core.Message{
		Role:    "user",
		Content: prompt,
	})
	messages := make([]core.Message, len(a.messages))
	copy(messages, a.messages)

	// Add user message to session
	userMsg := SessionMessage{
		Role:      "user",
		Content:   prompt,
		Timestamp: time.Now(),
	}
	a.currentSession.Messages = append(a.currentSession.Messages, userMsg)

	// Set title from first user message
	if a.currentSession.Title == "" {
		a.currentSession.Title = GenerateTitle(prompt)
	}
	a.mu.Unlock()

	// Build request
	req := core.MessageRequest{
		Model:     a.model,
		Messages:  messages,
		MaxTokens: 8192,
		Stream:    true,
		System:    a.systemPrompt,
		Tools:     a.GetToolDefinitions(),
	}

	// Start streaming
	chunks, err := a.client.CreateMessageStream(ctx, req)
	if err != nil {
		a.emit(tux.Event{Type: tux.EventError, Error: err})
		return err
	}

	// Process stream
	return a.processStream(ctx, chunks)
}

// processStream handles the streaming response from the API.
func (a *HexAgent) processStream(ctx context.Context, chunks <-chan *core.StreamChunk) error {
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
					if a.assemblingTool != nil {
						a.toolInputJSONBuf.WriteString(chunk.Delta.PartialJSON)
					}
				}
			}

		case "content_block_stop":
			a.handleContentBlockStop()

		case "message_stop":
			// Build assistant message with text and/or tool_use blocks
			if len(a.pendingTools) > 0 {
				// Assistant message with tool_use blocks
				var contentBlocks []core.ContentBlock
				if responseText.Len() > 0 {
					contentBlocks = append(contentBlocks, core.NewTextBlock(responseText.String()))
				}
				for _, tool := range a.pendingTools {
					contentBlocks = append(contentBlocks, core.ContentBlock{
						Type:  "tool_use",
						ID:    tool.ID,
						Name:  tool.Name,
						Input: tool.Input,
					})
				}
				a.mu.Lock()
				a.messages = append(a.messages, core.Message{
					Role:         "assistant",
					ContentBlock: contentBlocks,
				})
				a.mu.Unlock()

				if err := a.processTools(ctx, responseText.String()); err != nil {
					return err
				}
			} else if responseText.Len() > 0 {
				// Text-only response
				a.mu.Lock()
				a.messages = append(a.messages, core.Message{
					Role:    "assistant",
					Content: responseText.String(),
				})

				// Add assistant message to session and save
				if a.currentSession != nil {
					assistantMsg := SessionMessage{
						Role:      "assistant",
						Content:   responseText.String(),
						Timestamp: time.Now(),
					}
					a.currentSession.Messages = append(a.currentSession.Messages, assistantMsg)
					a.currentSession.UpdatedAt = time.Now()
				}
				a.mu.Unlock()

				// Save session (outside lock)
				a.saveSession()

				a.emit(tux.Event{Type: tux.EventComplete})
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

// saveSession persists the current session to storage if configured.
// This is safe to call even if storage or currentSession is nil.
func (a *HexAgent) saveSession() {
	a.mu.Lock()
	storage := a.storage
	session := a.currentSession
	a.mu.Unlock()

	if storage != nil && session != nil {
		// Ignore errors for now - session saving is best-effort
		_ = storage.Save(session)
	}
}

// GetToolDefinitions returns tool definitions for all registered tools.
// Returns nil if executor or registry is not configured.
func (a *HexAgent) GetToolDefinitions() []core.ToolDefinition {
	if a.executor == nil {
		return nil
	}
	registry := a.executor.Registry()
	if registry == nil {
		return nil
	}
	return registry.GetDefinitions()
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
// responseText is the text content that preceded the tool calls (may be empty).
func (a *HexAgent) processTools(ctx context.Context, responseText string) error {
	var toolResults []tools.ToolResult
	var sessionToolCalls []SessionToolCall

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

			// Track for session
			sessionToolCalls = append(sessionToolCalls, SessionToolCall{
				ID:     tool.ID,
				Name:   tool.Name,
				Input:  tool.Input,
				Output: result.Content,
				Error:  result.IsError,
			})

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

			// Track denied tool for session
			sessionToolCalls = append(sessionToolCalls, SessionToolCall{
				ID:     tool.ID,
				Name:   tool.Name,
				Input:  tool.Input,
				Output: "Tool execution denied by user",
				Error:  true,
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

	// Add assistant message with tool calls to session
	a.mu.Lock()
	if a.currentSession != nil {
		assistantMsg := SessionMessage{
			Role:      "assistant",
			Content:   responseText,
			Timestamp: time.Now(),
			ToolCalls: sessionToolCalls,
		}
		a.currentSession.Messages = append(a.currentSession.Messages, assistantMsg)
		a.currentSession.UpdatedAt = time.Now()
	}
	a.mu.Unlock()

	// Save session after tool execution
	a.saveSession()

	// Clear pending tools
	a.pendingTools = nil

	// Continue conversation with tool results
	return a.continueWithToolResults(ctx, toolResults)
}

// requestApproval emits an approval event and waits for user decision.
func (a *HexAgent) requestApproval(ctx context.Context, tool *core.ToolUse) (tux.ApprovalDecision, error) {
	// Create response channel
	responseChan := make(chan tux.ApprovalDecision, 1)

	// Emit approval event with params now available
	a.emit(tux.Event{
		Type:       tux.EventApproval,
		ToolID:     tool.ID,
		ToolName:   tool.Name,
		ToolParams: tool.Input,
		Response:   responseChan,
	})

	// Wait for decision
	select {
	case decision := <-responseChan:
		return decision, nil
	case <-ctx.Done():
		return tux.DecisionDeny, ctx.Err()
	}
}

// executeTool runs a tool and returns the result.
func (a *HexAgent) executeTool(ctx context.Context, tool *core.ToolUse) tools.ToolResult {
	if a.executor == nil {
		return tools.ToolResult{
			Type:      "tool_result",
			ToolUseID: tool.ID,
			Content:   "Tool executor not configured",
			IsError:   true,
		}
	}

	result, err := a.executor.Execute(ctx, tool.Name, tool.Input)
	if err != nil {
		return tools.ToolResult{
			Type:      "tool_result",
			ToolUseID: tool.ID,
			Content:   err.Error(),
			IsError:   true,
		}
	}

	return tools.ResultToToolResult(result, tool.ID)
}

// continueWithToolResults sends tool results to API and resumes streaming.
func (a *HexAgent) continueWithToolResults(ctx context.Context, results []tools.ToolResult) error {
	// Reset tool state for continuation
	a.resetToolState()

	// Build tool result content blocks
	contentBlocks := make([]core.ContentBlock, 0, len(results))
	for _, r := range results {
		contentBlocks = append(contentBlocks, core.NewToolResultBlock(r.ToolUseID, r.Content))
	}

	// Add to message history
	a.mu.Lock()
	a.messages = append(a.messages, core.Message{
		Role:         "user",
		ContentBlock: contentBlocks,
	})
	messages := make([]core.Message, len(a.messages))
	copy(messages, a.messages)
	a.mu.Unlock()

	// Build continuation request
	req := core.MessageRequest{
		Model:     a.model,
		Messages:  messages,
		MaxTokens: 8192,
		Stream:    true,
		System:    a.systemPrompt,
		Tools:     a.GetToolDefinitions(),
	}

	// Start new stream
	chunks, err := a.client.CreateMessageStream(ctx, req)
	if err != nil {
		a.emit(tux.Event{Type: tux.EventError, Error: err})
		return err
	}

	// Process continuation stream
	return a.processStream(ctx, chunks)
}
