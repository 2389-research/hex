// ABOUTME: AgentService implementation with message queuing
// ABOUTME: Handles concurrent LLM requests with FIFO queuing per conversation

package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/2389-research/hex/internal/core"
	"github.com/google/uuid"
)

// LLMClient defines the interface for interacting with LLM APIs
type LLMClient interface {
	CreateMessage(ctx context.Context, req core.MessageRequest) (*core.MessageResponse, error)
	CreateMessageStream(ctx context.Context, req core.MessageRequest) (<-chan *core.StreamChunk, error)
}

// agentService implements AgentService with concurrent message queuing
type agentService struct {
	client         LLMClient
	convSvc        ConversationService
	msgSvc         MessageService
	messageQueue   *sync.Map // map[string][]AgentCall
	activeRequests *sync.Map // map[string]context.CancelFunc
	mu             sync.Mutex
}

// NewAgentService creates a new agent service
func NewAgentService(client LLMClient, convSvc ConversationService, msgSvc MessageService) AgentService {
	return &agentService{
		client:         client,
		convSvc:        convSvc,
		msgSvc:         msgSvc,
		messageQueue:   &sync.Map{},
		activeRequests: &sync.Map{},
	}
}

// Run executes a prompt (queues if conversation is busy)
func (s *agentService) Run(ctx context.Context, call AgentCall) (*AgentResult, error) {
	// Check if conversation is busy
	if s.IsConversationBusy(call.ConversationID) {
		// Queue the call
		s.enqueueCall(call)
		return nil, nil
	}

	// Execute immediately
	return s.executeCall(ctx, call)
}

// Stream executes a prompt with streaming response
func (s *agentService) Stream(ctx context.Context, call AgentCall) (<-chan StreamEvent, error) {
	// Check if conversation is busy
	if s.IsConversationBusy(call.ConversationID) {
		// Queue the call - for streaming we still queue but return empty channel
		s.enqueueCall(call)
		ch := make(chan StreamEvent)
		close(ch)
		return ch, nil
	}

	// Execute streaming request
	return s.executeStream(ctx, call)
}

// IsConversationBusy returns true if conversation has active request
func (s *agentService) IsConversationBusy(convID string) bool {
	_, busy := s.activeRequests.Load(convID)
	return busy
}

// QueuedPrompts returns number of queued messages for conversation
func (s *agentService) QueuedPrompts(convID string) int {
	val, ok := s.messageQueue.Load(convID)
	if !ok {
		return 0
	}
	queue := val.([]AgentCall)
	return len(queue)
}

// CancelConversation cancels active request and clears queue
func (s *agentService) CancelConversation(convID string) {
	// Cancel active request
	if cancel, ok := s.activeRequests.LoadAndDelete(convID); ok {
		cancelFunc := cancel.(context.CancelFunc)
		cancelFunc()
	}

	// Clear queue
	s.messageQueue.Delete(convID)
}

// enqueueCall adds a call to the conversation's queue
func (s *agentService) enqueueCall(call AgentCall) {
	s.mu.Lock()
	defer s.mu.Unlock()

	val, _ := s.messageQueue.LoadOrStore(call.ConversationID, []AgentCall{})
	queue := val.([]AgentCall)
	queue = append(queue, call)
	s.messageQueue.Store(call.ConversationID, queue)
}

// dequeueCall removes and returns the next call from the queue
func (s *agentService) dequeueCall(convID string) *AgentCall {
	s.mu.Lock()
	defer s.mu.Unlock()

	val, ok := s.messageQueue.Load(convID)
	if !ok {
		return nil
	}

	queue := val.([]AgentCall)
	if len(queue) == 0 {
		return nil
	}

	// Get first item (FIFO)
	call := queue[0]
	queue = queue[1:]

	if len(queue) == 0 {
		s.messageQueue.Delete(convID)
	} else {
		s.messageQueue.Store(convID, queue)
	}

	return &call
}

// executeCall executes an agent call and handles queued messages
func (s *agentService) executeCall(ctx context.Context, call AgentCall) (*AgentResult, error) {
	// Create cancellable context
	callCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Register active request
	s.activeRequests.Store(call.ConversationID, cancel)
	defer s.activeRequests.Delete(call.ConversationID)

	// Store user message
	userMsg := &Message{
		ID:             uuid.New().String(),
		ConversationID: call.ConversationID,
		Role:           "user",
		Content:        call.Prompt,
		CreatedAt:      time.Now(),
	}

	if err := s.msgSvc.Add(callCtx, userMsg); err != nil {
		return nil, fmt.Errorf("store user message: %w", err)
	}

	// Get conversation history for context
	messages, err := s.msgSvc.GetByConversation(callCtx, call.ConversationID)
	if err != nil {
		return nil, fmt.Errorf("get conversation history: %w", err)
	}

	// Build message request
	apiMessages := make([]core.Message, 0, len(messages))
	for _, msg := range messages {
		apiMessages = append(apiMessages, core.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	req := core.MessageRequest{
		Model:     "claude-3-5-sonnet-20241022",
		Messages:  apiMessages,
		MaxTokens: call.MaxTokens,
	}

	// Call LLM
	resp, err := s.client.CreateMessage(callCtx, req)
	if err != nil {
		// Check if context was cancelled
		if callCtx.Err() != nil {
			return nil, callCtx.Err()
		}
		return nil, fmt.Errorf("create message: %w", err)
	}

	// Extract text response
	responseText := resp.GetTextContent()

	// Store assistant message
	assistantMsg := &Message{
		ID:             uuid.New().String(),
		ConversationID: call.ConversationID,
		Role:           "assistant",
		Content:        responseText,
		Provider:       "anthropic",
		Model:          resp.Model,
		CreatedAt:      time.Now(),
	}

	if err := s.msgSvc.Add(callCtx, assistantMsg); err != nil {
		return nil, fmt.Errorf("store assistant message: %w", err)
	}

	// Update token usage
	if err := s.convSvc.UpdateTokenUsage(
		callCtx,
		call.ConversationID,
		int64(resp.Usage.InputTokens),
		int64(resp.Usage.OutputTokens),
	); err != nil {
		return nil, fmt.Errorf("update token usage: %w", err)
	}

	// Create result
	result := &AgentResult{
		Text:         responseText,
		PromptTokens: int64(resp.Usage.InputTokens),
		OutputTokens: int64(resp.Usage.OutputTokens),
	}

	// Process next queued message if any
	go s.processNextQueued(call.ConversationID)

	return result, nil
}

// executeStream executes a streaming agent call
func (s *agentService) executeStream(ctx context.Context, call AgentCall) (<-chan StreamEvent, error) {
	// Create cancellable context
	callCtx, cancel := context.WithCancel(ctx)

	// Register active request
	s.activeRequests.Store(call.ConversationID, cancel)

	// Create output channel
	outCh := make(chan StreamEvent, 10)

	go func() {
		defer close(outCh)
		defer s.activeRequests.Delete(call.ConversationID)
		defer cancel()

		// Store user message
		userMsg := &Message{
			ID:             uuid.New().String(),
			ConversationID: call.ConversationID,
			Role:           "user",
			Content:        call.Prompt,
			CreatedAt:      time.Now(),
		}

		if err := s.msgSvc.Add(callCtx, userMsg); err != nil {
			outCh <- StreamEvent{Type: "error", Error: fmt.Errorf("store user message: %w", err)}
			return
		}

		// Get conversation history
		messages, err := s.msgSvc.GetByConversation(callCtx, call.ConversationID)
		if err != nil {
			outCh <- StreamEvent{Type: "error", Error: fmt.Errorf("get conversation history: %w", err)}
			return
		}

		// Build message request
		apiMessages := make([]core.Message, 0, len(messages))
		for _, msg := range messages {
			apiMessages = append(apiMessages, core.Message{
				Role:    msg.Role,
				Content: msg.Content,
			})
		}

		req := core.MessageRequest{
			Model:     "claude-3-5-sonnet-20241022",
			Messages:  apiMessages,
			MaxTokens: call.MaxTokens,
		}

		// Call streaming API
		chunks, err := s.client.CreateMessageStream(callCtx, req)
		if err != nil {
			outCh <- StreamEvent{Type: "error", Error: fmt.Errorf("create message stream: %w", err)}
			return
		}

		// Accumulate response
		accumulator := core.NewStreamAccumulator()
		var lastUsage *core.Usage

		for chunk := range chunks {
			if chunk.Delta != nil && chunk.Delta.Text != "" {
				accumulator.Add(chunk)
				outCh <- StreamEvent{
					Type: "text",
					Text: chunk.Delta.Text,
				}
			}

			if chunk.Usage != nil {
				lastUsage = chunk.Usage
			}

			if chunk.Done {
				break
			}
		}

		// Store assistant message
		responseText := accumulator.GetText()
		assistantMsg := &Message{
			ID:             uuid.New().String(),
			ConversationID: call.ConversationID,
			Role:           "assistant",
			Content:        responseText,
			Provider:       "anthropic",
			Model:          req.Model,
			CreatedAt:      time.Now(),
		}

		if err := s.msgSvc.Add(callCtx, assistantMsg); err != nil {
			outCh <- StreamEvent{Type: "error", Error: fmt.Errorf("store assistant message: %w", err)}
			return
		}

		// Update token usage if available
		if lastUsage != nil {
			if err := s.convSvc.UpdateTokenUsage(
				callCtx,
				call.ConversationID,
				int64(lastUsage.InputTokens),
				int64(lastUsage.OutputTokens),
			); err != nil {
				outCh <- StreamEvent{Type: "error", Error: fmt.Errorf("update token usage: %w", err)}
				return
			}

			outCh <- StreamEvent{
				Type:         "done",
				PromptTokens: int64(lastUsage.InputTokens),
				OutputTokens: int64(lastUsage.OutputTokens),
			}
		} else {
			outCh <- StreamEvent{Type: "done"}
		}

		// Process next queued message
		s.processNextQueued(call.ConversationID)
	}()

	return outCh, nil
}

// processNextQueued processes the next queued message for a conversation
func (s *agentService) processNextQueued(convID string) {
	call := s.dequeueCall(convID)
	if call == nil {
		return
	}

	// Execute the queued call with a background context
	ctx := context.Background()
	_, _ = s.executeCall(ctx, *call)
}
