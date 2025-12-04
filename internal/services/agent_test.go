// ABOUTME: Test suite for AgentService implementation
// ABOUTME: Covers queuing, cancellation, concurrency, and token tracking

package services

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/2389-research/hex/internal/core"
)

// MockClient implements core.Client interface for testing
type MockClient struct {
	mu                sync.Mutex
	createMessageFunc func(ctx context.Context, req core.MessageRequest) (*core.MessageResponse, error)
	streamFunc        func(ctx context.Context, req core.MessageRequest) (<-chan *core.StreamChunk, error)
	callCount         int
}

func (m *MockClient) CreateMessage(ctx context.Context, req core.MessageRequest) (*core.MessageResponse, error) {
	m.mu.Lock()
	m.callCount++
	m.mu.Unlock()

	if m.createMessageFunc != nil {
		return m.createMessageFunc(ctx, req)
	}

	// Default mock response
	return &core.MessageResponse{
		ID:   "msg_123",
		Role: "assistant",
		Content: []core.Content{
			{Type: "text", Text: "Mock response"},
		},
		Usage: core.Usage{
			InputTokens:  10,
			OutputTokens: 20,
		},
	}, nil
}

func (m *MockClient) CreateMessageStream(ctx context.Context, req core.MessageRequest) (<-chan *core.StreamChunk, error) {
	m.mu.Lock()
	m.callCount++
	m.mu.Unlock()

	if m.streamFunc != nil {
		return m.streamFunc(ctx, req)
	}

	// Default mock stream response
	ch := make(chan *core.StreamChunk, 3)
	go func() {
		defer close(ch)
		ch <- &core.StreamChunk{
			Type:  "content_block_delta",
			Delta: &core.Delta{Type: "text_delta", Text: "Mock "},
		}
		ch <- &core.StreamChunk{
			Type:  "content_block_delta",
			Delta: &core.Delta{Type: "text_delta", Text: "stream"},
		}
		ch <- &core.StreamChunk{
			Type:  "message_stop",
			Done:  true,
			Usage: &core.Usage{InputTokens: 10, OutputTokens: 20},
		}
	}()
	return ch, nil
}

func (m *MockClient) GetCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.callCount
}

func TestAgentService_Run(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	convSvc := NewConversationService(db)
	msgSvc := NewMessageService(db)
	client := &MockClient{}
	agentSvc := NewAgentService(client, convSvc, msgSvc)

	// Create test conversation
	ctx := context.Background()
	conv, err := convSvc.Create(ctx, "Test Conversation")
	if err != nil {
		t.Fatalf("Failed to create conversation: %v", err)
	}

	// Execute agent call
	call := AgentCall{
		ConversationID: conv.ID,
		Prompt:         "Hello, world!",
		MaxTokens:      1000,
	}

	result, err := agentSvc.Run(ctx, call)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if result.Text != "Mock response" {
		t.Errorf("Expected 'Mock response', got '%s'", result.Text)
	}

	if result.PromptTokens != 10 {
		t.Errorf("Expected PromptTokens=10, got %d", result.PromptTokens)
	}

	if result.OutputTokens != 20 {
		t.Errorf("Expected OutputTokens=20, got %d", result.OutputTokens)
	}

	// Verify token usage was updated
	updatedConv, err := convSvc.Get(ctx, conv.ID)
	if err != nil {
		t.Fatalf("Failed to get updated conversation: %v", err)
	}

	if updatedConv.PromptTokens != 10 {
		t.Errorf("Expected conversation PromptTokens=10, got %d", updatedConv.PromptTokens)
	}

	if updatedConv.CompletionTokens != 20 {
		t.Errorf("Expected conversation CompletionTokens=20, got %d", updatedConv.CompletionTokens)
	}

	// Verify messages were stored
	messages, err := msgSvc.GetByConversation(ctx, conv.ID)
	if err != nil {
		t.Fatalf("Failed to get messages: %v", err)
	}

	if len(messages) != 2 {
		t.Fatalf("Expected 2 messages (user + assistant), got %d", len(messages))
	}

	if messages[0].Role != "user" || messages[0].Content != "Hello, world!" {
		t.Errorf("First message should be user message with prompt")
	}

	if messages[1].Role != "assistant" || messages[1].Content != "Mock response" {
		t.Errorf("Second message should be assistant response")
	}
}

func TestAgentService_Queuing(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	convSvc := NewConversationService(db)
	msgSvc := NewMessageService(db)

	// Create a client that blocks to simulate long-running request
	blockCh := make(chan struct{})
	client := &MockClient{
		createMessageFunc: func(_ context.Context, _ core.MessageRequest) (*core.MessageResponse, error) {
			<-blockCh // Block until we signal
			return &core.MessageResponse{
				ID:   "msg_123",
				Role: "assistant",
				Content: []core.Content{
					{Type: "text", Text: "Response"},
				},
				Usage: core.Usage{InputTokens: 10, OutputTokens: 20},
			}, nil
		},
	}

	agentSvc := NewAgentService(client, convSvc, msgSvc)

	// Create test conversation
	ctx := context.Background()
	conv, err := convSvc.Create(ctx, "Test Conversation")
	if err != nil {
		t.Fatalf("Failed to create conversation: %v", err)
	}

	// Start first request (will block)
	call1 := AgentCall{
		ConversationID: conv.ID,
		Prompt:         "First prompt",
		MaxTokens:      1000,
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, _ = agentSvc.Run(ctx, call1)
	}()

	// Give first request time to start
	time.Sleep(50 * time.Millisecond)

	// Verify conversation is busy
	if !agentSvc.IsConversationBusy(conv.ID) {
		t.Error("Expected conversation to be busy")
	}

	// Send second request while first is active (should queue)
	call2 := AgentCall{
		ConversationID: conv.ID,
		Prompt:         "Second prompt",
		MaxTokens:      1000,
	}

	result2, err := agentSvc.Run(ctx, call2)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	// Second request should return nil (queued)
	if result2 != nil {
		t.Error("Expected nil result for queued request")
	}

	// Verify queue depth
	queueDepth := agentSvc.QueuedPrompts(conv.ID)
	if queueDepth != 1 {
		t.Errorf("Expected queue depth=1, got %d", queueDepth)
	}

	// Unblock first request
	close(blockCh)

	// Wait for first request to complete
	wg.Wait()

	// Give time for queued request to process
	time.Sleep(100 * time.Millisecond)

	// Verify queue is now empty
	queueDepth = agentSvc.QueuedPrompts(conv.ID)
	if queueDepth != 0 {
		t.Errorf("Expected queue depth=0 after processing, got %d", queueDepth)
	}

	// Verify conversation is no longer busy
	if agentSvc.IsConversationBusy(conv.ID) {
		t.Error("Expected conversation to not be busy after processing")
	}
}

func TestAgentService_Cancel(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	convSvc := NewConversationService(db)
	msgSvc := NewMessageService(db)

	// Create a client that blocks indefinitely
	blockCh := make(chan struct{})
	client := &MockClient{
		createMessageFunc: func(ctx context.Context, _ core.MessageRequest) (*core.MessageResponse, error) {
			<-ctx.Done()
			return nil, ctx.Err()
		},
	}

	agentSvc := NewAgentService(client, convSvc, msgSvc)

	// Create test conversation
	ctx := context.Background()
	conv, err := convSvc.Create(ctx, "Test Conversation")
	if err != nil {
		t.Fatalf("Failed to create conversation: %v", err)
	}

	// Start request
	call := AgentCall{
		ConversationID: conv.ID,
		Prompt:         "Test prompt",
		MaxTokens:      1000,
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, _ = agentSvc.Run(ctx, call)
	}()

	// Give request time to start
	time.Sleep(50 * time.Millisecond)

	// Queue some additional requests
	for i := 0; i < 3; i++ {
		queuedCall := AgentCall{
			ConversationID: conv.ID,
			Prompt:         "Queued prompt",
			MaxTokens:      1000,
		}
		_, _ = agentSvc.Run(ctx, queuedCall)
	}

	// Verify queue depth
	if agentSvc.QueuedPrompts(conv.ID) != 3 {
		t.Errorf("Expected 3 queued prompts")
	}

	// Cancel conversation
	agentSvc.CancelConversation(conv.ID)
	defer close(blockCh)

	// Wait for goroutine to finish
	wg.Wait()

	// Verify queue was cleared
	if agentSvc.QueuedPrompts(conv.ID) != 0 {
		t.Error("Expected queue to be cleared after cancellation")
	}

	// Verify conversation is no longer busy
	if agentSvc.IsConversationBusy(conv.ID) {
		t.Error("Expected conversation to not be busy after cancellation")
	}
}

func TestAgentService_IsConversationBusy(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	convSvc := NewConversationService(db)
	msgSvc := NewMessageService(db)
	client := &MockClient{}
	agentSvc := NewAgentService(client, convSvc, msgSvc)

	// Create test conversation
	ctx := context.Background()
	conv, err := convSvc.Create(ctx, "Test Conversation")
	if err != nil {
		t.Fatalf("Failed to create conversation: %v", err)
	}

	// Initially not busy
	if agentSvc.IsConversationBusy(conv.ID) {
		t.Error("Expected conversation to not be busy initially")
	}

	// Non-existent conversation should not be busy
	if agentSvc.IsConversationBusy("non-existent-id") {
		t.Error("Expected non-existent conversation to not be busy")
	}
}

func TestAgentService_QueuedPrompts(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	convSvc := NewConversationService(db)
	msgSvc := NewMessageService(db)
	client := &MockClient{}
	agentSvc := NewAgentService(client, convSvc, msgSvc)

	// Create test conversation
	ctx := context.Background()
	conv, err := convSvc.Create(ctx, "Test Conversation")
	if err != nil {
		t.Fatalf("Failed to create conversation: %v", err)
	}

	// Initially no queued prompts
	if agentSvc.QueuedPrompts(conv.ID) != 0 {
		t.Error("Expected 0 queued prompts initially")
	}

	// Non-existent conversation should have 0 queued prompts
	if agentSvc.QueuedPrompts("non-existent-id") != 0 {
		t.Error("Expected 0 queued prompts for non-existent conversation")
	}
}
