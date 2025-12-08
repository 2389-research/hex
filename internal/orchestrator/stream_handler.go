// ABOUTME: Stream handling logic for agent orchestrator
// ABOUTME: Processes streaming chunks and manages stream lifecycle
package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/2389-research/hex/internal/core"
	"github.com/2389-research/hex/internal/cost"
	"github.com/2389-research/hex/internal/events"
	"github.com/google/uuid"
)

const MaxPendingTools = 100

// handleStream processes the stream in a background goroutine
func (o *AgentOrchestrator) handleStream(ctx context.Context) {
	// Create stream request
	req := core.MessageRequest{
		Model:     "claude-sonnet-4-5-20250929",
		Messages:  o.getMessageHistorySafe(),
		MaxTokens: 4096,
		Stream:    true,
	}

	// Start stream
	streamChan, err := o.client.CreateMessageStream(ctx, req)
	if err != nil {
		o.emitEvent(EventError, err)
		o.setState(StateError)
		return
	}

	o.mu.Lock()
	o.streamChan = streamChan
	o.mu.Unlock()

	// Process stream chunks
	for {
		select {
		case <-ctx.Done():
			// Stream cancelled
			if os.Getenv("HEX_DEBUG") != "" {
				fmt.Fprintf(os.Stderr, "[ORCHESTRATOR] Stream cancelled: %v\n", ctx.Err())
			}
			return

		case chunk, ok := <-streamChan:
			if !ok {
				// Stream closed
				if os.Getenv("HEX_DEBUG") != "" {
					fmt.Fprintf(os.Stderr, "[ORCHESTRATOR] Stream closed\n")
				}
				o.handleStreamComplete()
				return
			}

			// Process chunk
			o.processChunk(chunk)
		}
	}
}

// processChunk handles a single stream chunk
func (o *AgentOrchestrator) processChunk(chunk *core.StreamChunk) {
	if chunk == nil {
		return
	}

	if os.Getenv("HEX_DEBUG") != "" {
		fmt.Fprintf(os.Stderr, "[ORCHESTRATOR] Processing chunk type: %s\n", chunk.Type)
	}

	// Emit chunk event
	o.emitEvent(EventStreamChunk, chunk)

	switch chunk.Type {
	case "content_block_start":
		o.handleContentBlockStart(chunk)
	case "content_block_delta":
		o.handleContentBlockDelta(chunk)
	case "content_block_stop":
		o.handleContentBlockStop(chunk)
	case "message_delta":
		o.handleMessageDelta(chunk)
	case "message_stop":
		o.handleStreamComplete()
	}

	// Also handle Done flag
	if chunk.Done {
		o.handleStreamComplete()
	}
}

// handleContentBlockStart processes the start of a content block (tool_use)
func (o *AgentOrchestrator) handleContentBlockStart(chunk *core.StreamChunk) {
	if chunk.ContentBlock == nil {
		return
	}

	// Only handle tool_use blocks
	if chunk.ContentBlock.Type != "tool_use" {
		return
	}

	if os.Getenv("HEX_DEBUG") != "" {
		fmt.Fprintf(os.Stderr, "[ORCHESTRATOR] Tool use started: %s (id=%s)\n",
			chunk.ContentBlock.Name, chunk.ContentBlock.ID)
	}

	o.mu.Lock()
	o.assemblingTool = &core.ToolUse{
		Type:  "tool_use",
		ID:    chunk.ContentBlock.ID,
		Name:  chunk.ContentBlock.Name,
		Input: make(map[string]interface{}),
	}
	o.toolInputBuf = ""
	o.mu.Unlock()
}

// handleContentBlockDelta processes delta events for content blocks
func (o *AgentOrchestrator) handleContentBlockDelta(chunk *core.StreamChunk) {
	if chunk.Delta == nil {
		return
	}

	o.mu.Lock()
	defer o.mu.Unlock()

	// Text delta
	if chunk.Delta.Text != "" {
		o.streamingText += chunk.Delta.Text
		if os.Getenv("HEX_DEBUG") != "" {
			fmt.Fprintf(os.Stderr, "[ORCHESTRATOR] Text delta: %d chars\n", len(chunk.Delta.Text))
		}
	}

	// Tool input JSON delta
	if chunk.Delta.Type == "input_json_delta" && chunk.Delta.PartialJSON != "" {
		o.toolInputBuf += chunk.Delta.PartialJSON
		if os.Getenv("HEX_DEBUG") != "" {
			fmt.Fprintf(os.Stderr, "[ORCHESTRATOR] Tool input delta: %d chars\n", len(chunk.Delta.PartialJSON))
		}
	}
}

// handleContentBlockStop processes content block completion (tool parameters complete)
func (o *AgentOrchestrator) handleContentBlockStop(chunk *core.StreamChunk) {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.assemblingTool == nil {
		return
	}

	// Enforce maximum pending tools
	if len(o.pendingToolUses) >= MaxPendingTools {
		if os.Getenv("HEX_DEBUG") != "" {
			fmt.Fprintf(os.Stderr, "[ORCHESTRATOR] WARNING: Max pending tools reached (%d), dropping %s\n",
				MaxPendingTools, o.assemblingTool.Name)
		}
		o.emitEventLocked(EventError, fmt.Errorf("maximum pending tools exceeded (%d)", MaxPendingTools))
		o.assemblingTool = nil
		o.toolInputBuf = ""
		return
	}

	// Parse accumulated JSON
	if o.toolInputBuf != "" {
		var input map[string]interface{}
		if err := json.Unmarshal([]byte(o.toolInputBuf), &input); err == nil {
			o.assemblingTool.Input = input
		} else if os.Getenv("HEX_DEBUG") != "" {
			fmt.Fprintf(os.Stderr, "[ORCHESTRATOR] Failed to parse tool input JSON: %v\n", err)
		}
	}

	if os.Getenv("HEX_DEBUG") != "" {
		fmt.Fprintf(os.Stderr, "[ORCHESTRATOR] Tool use complete: %s (id=%s)\n",
			o.assemblingTool.Name, o.assemblingTool.ID)
	}

	// Add to pending tools
	o.pendingToolUses = append(o.pendingToolUses, o.assemblingTool)

	// Record tool call requested event
	if store := events.Global(); store != nil {
		_ = store.Record(events.Event{
			ID:        uuid.New().String(),
			AgentID:   os.Getenv("HEX_AGENT_ID"),
			ParentID:  os.Getenv("HEX_PARENT_AGENT_ID"),
			Type:      events.EventToolCallRequested,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"tool_use_id": o.assemblingTool.ID,
				"tool_name":   o.assemblingTool.Name,
			},
		})
	}

	// Emit tool call event
	o.emitEventLocked(EventToolCall, o.assemblingTool)

	// Clear assembling state
	o.assemblingTool = nil
	o.toolInputBuf = ""
}

// handleMessageDelta processes message-level deltas (usage info)
func (o *AgentOrchestrator) handleMessageDelta(chunk *core.StreamChunk) {
	// Can emit usage info if needed
	if chunk.Usage != nil && os.Getenv("HEX_DEBUG") != "" {
		fmt.Fprintf(os.Stderr, "[ORCHESTRATOR] Usage: input=%d, output=%d\n",
			chunk.Usage.InputTokens, chunk.Usage.OutputTokens)
	}

	// Record cost tracking for streaming
	if chunk.Usage != nil {
		agentID := os.Getenv("HEX_AGENT_ID")
		parentID := os.Getenv("HEX_PARENT_AGENT_ID")

		if agentID != "" {
			// Convert core.Usage to cost.Usage
			costUsage := cost.Usage{
				InputTokens:      chunk.Usage.InputTokens,
				OutputTokens:     chunk.Usage.OutputTokens,
				CacheReadTokens:  chunk.Usage.CacheReadTokens,
				CacheWriteTokens: chunk.Usage.CacheWriteTokens,
			}
			if err := cost.Global().RecordUsage(agentID, parentID, "claude-sonnet-4-5-20250929", costUsage); err != nil {
				// Log error but don't fail the stream
				if os.Getenv("HEX_DEBUG") != "" {
					fmt.Fprintf(os.Stderr, "[ORCHESTRATOR] Cost tracking failed: %v\n", err)
				}
			}
		}
	}
}

// handleStreamComplete processes stream completion
func (o *AgentOrchestrator) handleStreamComplete() {
	o.mu.Lock()
	defer o.mu.Unlock()

	if os.Getenv("HEX_DEBUG") != "" {
		fmt.Fprintf(os.Stderr, "[ORCHESTRATOR] Stream complete, pending tools: %d\n",
			len(o.pendingToolUses))
	}

	// If there are pending tools, transition to awaiting approval
	if len(o.pendingToolUses) > 0 {
		if err := o.stateMachine.Transition(StateAwaitingApproval, map[string]interface{}{
			"pending_tools": len(o.pendingToolUses),
		}); err != nil && os.Getenv("HEX_DEBUG") != "" {
			fmt.Fprintf(os.Stderr, "[ORCHESTRATOR] Warning: failed to transition to awaiting approval: %v\n", err)
		}
		// Emit complete with pending tools indicator
		o.emitEventLocked(EventComplete, map[string]interface{}{
			"pending_tools": len(o.pendingToolUses),
		})

		// Clear stream state but keep pending tools
		o.streamCtx = nil
		o.streamCancel = nil
		o.streamChan = nil
		return
	}

	// No tools, just complete
	// Add assistant message to history if there's text
	if o.streamingText != "" {
		o.messageHistory = append(o.messageHistory, core.Message{
			Role:    "assistant",
			Content: o.streamingText,
		})
	}

	if err := o.stateMachine.Transition(StateComplete, nil); err != nil && os.Getenv("HEX_DEBUG") != "" {
		fmt.Fprintf(os.Stderr, "[ORCHESTRATOR] Warning: failed to transition to complete: %v\n", err)
	}
	o.emitEventLocked(EventComplete, nil)

	// Clear stream state
	o.streamingText = ""
	o.streamCtx = nil
	o.streamCancel = nil
	o.streamChan = nil
}

// getMessageHistorySafe returns a copy of message history (thread-safe)
func (o *AgentOrchestrator) getMessageHistorySafe() []core.Message {
	o.mu.RLock()
	defer o.mu.RUnlock()

	// Return a copy to prevent concurrent access issues
	history := make([]core.Message, len(o.messageHistory))
	copy(history, o.messageHistory)
	return history
}
