// ABOUTME: Tests for event subscription functionality
// ABOUTME: Verifies UI properly subscribes to and handles service events

package ui

import (
	"context"
	"testing"
	"time"

	"github.com/2389-research/hex/internal/pubsub"
	"github.com/2389-research/hex/internal/services"
	tea "github.com/charmbracelet/bubbletea"
)

// mockConversationService implements ConversationService for testing
type mockConversationService struct {
	events chan pubsub.Event[services.Conversation]
}

func (m *mockConversationService) Subscribe(_ context.Context) <-chan pubsub.Event[services.Conversation] {
	return m.events
}

func (m *mockConversationService) Create(_ context.Context, _ string) (*services.Conversation, error) {
	return nil, nil
}

func (m *mockConversationService) Get(_ context.Context, _ string) (*services.Conversation, error) {
	return nil, nil
}

func (m *mockConversationService) List(_ context.Context) ([]*services.Conversation, error) {
	return nil, nil
}

func (m *mockConversationService) Update(_ context.Context, _ *services.Conversation) error {
	return nil
}

func (m *mockConversationService) Delete(_ context.Context, _ string) error {
	return nil
}

func (m *mockConversationService) UpdateTokenUsage(_ context.Context, _ string, _, _ int64) error {
	return nil
}

// mockMessageService implements MessageService for testing
type mockMessageService struct {
	events chan pubsub.Event[services.Message]
}

func (m *mockMessageService) Subscribe(_ context.Context) <-chan pubsub.Event[services.Message] {
	return m.events
}

func (m *mockMessageService) Add(_ context.Context, _ *services.Message) error {
	return nil
}

func (m *mockMessageService) GetByConversation(_ context.Context, _ string) ([]*services.Message, error) {
	return nil, nil
}

func (m *mockMessageService) GetSummaries(_ context.Context, _ string) ([]*services.Message, error) {
	return nil, nil
}

// TestStartEventSubscriptions verifies that event subscriptions are properly initialized
func TestStartEventSubscriptions(t *testing.T) {
	// Create model
	model := NewModel("test-conv", "test-model")

	// Create mock services with event channels
	convEvents := make(chan pubsub.Event[services.Conversation], 1)
	msgEvents := make(chan pubsub.Event[services.Message], 1)

	convSvc := &mockConversationService{events: convEvents}
	msgSvc := &mockMessageService{events: msgEvents}

	// Set services
	model.SetServices(convSvc, msgSvc, nil)

	// Start event subscriptions
	cmd := model.StartEventSubscriptions()

	// Verify context was created
	if model.eventCtx == nil {
		t.Fatal("eventCtx should be initialized")
	}
	if model.eventCancel == nil {
		t.Fatal("eventCancel should be initialized")
	}

	// Verify command is not nil
	if cmd == nil {
		t.Fatal("StartEventSubscriptions should return a command")
	}

	// Clean up
	model.eventCancel()
	close(convEvents)
	close(msgEvents)
}

// TestConversationEventHandling verifies that conversation events are handled correctly
func TestConversationEventHandling(t *testing.T) {
	// Create model
	model := NewModel("test-conv", "test-model")

	// Create mock services
	convEvents := make(chan pubsub.Event[services.Conversation], 1)
	msgEvents := make(chan pubsub.Event[services.Message], 1)

	convSvc := &mockConversationService{events: convEvents}
	msgSvc := &mockMessageService{events: msgEvents}

	model.SetServices(convSvc, msgSvc, nil)
	model.StartEventSubscriptions()

	// Send a conversation update event
	conv := services.Conversation{
		ID:         "test-conv",
		IsFavorite: true,
	}
	event := pubsub.Event[services.Conversation]{
		Type:    pubsub.Updated,
		Payload: conv,
	}

	// Create event message
	eventMsg := conversationEventMsg{event: event}

	// Handle the event
	_, cmd := model.handleConversationEvent(eventMsg)

	// Verify favorite status was updated
	if !model.IsFavorite {
		t.Error("IsFavorite should be true after event")
	}

	// Verify command to continue listening was returned
	if cmd == nil {
		t.Error("handleConversationEvent should return command to continue listening")
	}

	// Clean up
	model.eventCancel()
	close(convEvents)
	close(msgEvents)
}

// TestMessageEventHandling verifies that message events are handled correctly
func TestMessageEventHandling(t *testing.T) {
	// Create model
	model := NewModel("test-conv", "test-model")

	// Create mock services
	convEvents := make(chan pubsub.Event[services.Conversation], 1)
	msgEvents := make(chan pubsub.Event[services.Message], 1)

	convSvc := &mockConversationService{events: convEvents}
	msgSvc := &mockMessageService{events: msgEvents}

	model.SetServices(convSvc, msgSvc, nil)
	model.StartEventSubscriptions()

	// Send a message created event
	msg := services.Message{
		ConversationID: "test-conv",
		Role:           "user",
		Content:        "test message",
	}
	event := pubsub.Event[services.Message]{
		Type:    pubsub.Created,
		Payload: msg,
	}

	// Create event message
	eventMsg := messageEventMsg{event: event}

	// Handle the event
	_, cmd := model.handleMessageEvent(eventMsg)

	// Verify command to continue listening was returned
	if cmd == nil {
		t.Error("handleMessageEvent should return command to continue listening")
	}

	// Clean up
	model.eventCancel()
	close(convEvents)
	close(msgEvents)
}

// TestEventCleanupOnQuit verifies that event subscriptions are canceled on quit
func TestEventCleanupOnQuit(t *testing.T) {
	// Create model
	model := NewModel("test-conv", "test-model")

	// Create mock services
	convEvents := make(chan pubsub.Event[services.Conversation], 1)
	msgEvents := make(chan pubsub.Event[services.Message], 1)

	convSvc := &mockConversationService{events: convEvents}
	msgSvc := &mockMessageService{events: msgEvents}

	model.SetServices(convSvc, msgSvc, nil)
	model.StartEventSubscriptions()

	// Verify context is active
	select {
	case <-model.eventCtx.Done():
		t.Fatal("eventCtx should not be canceled initially")
	default:
		// Good - context is still active
	}

	// Cancel subscriptions (simulating quit)
	model.eventCancel()

	// Verify context is canceled
	select {
	case <-model.eventCtx.Done():
		// Good - context was canceled
	case <-time.After(100 * time.Millisecond):
		t.Error("eventCtx should be canceled after calling eventCancel")
	}

	// Clean up
	close(convEvents)
	close(msgEvents)
}

// TestEventSubscriptionsInInit verifies that Init starts subscriptions when services are available
func TestEventSubscriptionsInInit(t *testing.T) {
	// Create model
	model := NewModel("test-conv", "test-model")

	// Create mock services
	convEvents := make(chan pubsub.Event[services.Conversation], 1)
	msgEvents := make(chan pubsub.Event[services.Message], 1)

	convSvc := &mockConversationService{events: convEvents}
	msgSvc := &mockMessageService{events: msgEvents}

	// Set services before calling Init
	model.SetServices(convSvc, msgSvc, nil)

	// Call Init
	cmd := model.Init()

	// Verify command was returned
	if cmd == nil {
		t.Fatal("Init should return a command when services are available")
	}

	// Execute the command to trigger subscription
	msg := cmd()

	// Should return a batch command
	if msg == nil {
		t.Error("Init command should return a message")
	}

	// Clean up
	if model.eventCancel != nil {
		model.eventCancel()
	}
	close(convEvents)
	close(msgEvents)
}

// TestUpdateHandlesConversationEvent verifies Update properly handles conversation events
func TestUpdateHandlesConversationEvent(t *testing.T) {
	// Create model
	model := NewModel("test-conv", "test-model")

	// Create mock services
	convEvents := make(chan pubsub.Event[services.Conversation], 1)
	msgEvents := make(chan pubsub.Event[services.Message], 1)

	convSvc := &mockConversationService{events: convEvents}
	msgSvc := &mockMessageService{events: msgEvents}

	model.SetServices(convSvc, msgSvc, nil)
	model.StartEventSubscriptions()

	// Create conversation event
	conv := services.Conversation{
		ID:         "test-conv",
		IsFavorite: true,
	}
	event := pubsub.Event[services.Conversation]{
		Type:    pubsub.Updated,
		Payload: conv,
	}
	eventMsg := conversationEventMsg{event: event}

	// Process through Update
	updatedModel, cmd := model.Update(eventMsg)

	// Verify model was updated
	m, ok := updatedModel.(*Model)
	if !ok {
		t.Fatal("Update should return *Model")
	}

	if !m.IsFavorite {
		t.Error("IsFavorite should be updated")
	}

	// Verify command was returned
	if cmd == nil {
		t.Error("Update should return command to continue listening")
	}

	// Clean up
	model.eventCancel()
	close(convEvents)
	close(msgEvents)
}

// TestUpdateHandlesMessageEvent verifies Update properly handles message events
func TestUpdateHandlesMessageEvent(t *testing.T) {
	// Create model
	model := NewModel("test-conv", "test-model")

	// Create mock services
	convEvents := make(chan pubsub.Event[services.Conversation], 1)
	msgEvents := make(chan pubsub.Event[services.Message], 1)

	convSvc := &mockConversationService{events: convEvents}
	msgSvc := &mockMessageService{events: msgEvents}

	model.SetServices(convSvc, msgSvc, nil)
	model.StartEventSubscriptions()

	// Create message event
	msg := services.Message{
		ConversationID: "test-conv",
		Role:           "user",
		Content:        "test",
	}
	event := pubsub.Event[services.Message]{
		Type:    pubsub.Created,
		Payload: msg,
	}
	eventMsg := messageEventMsg{event: event}

	// Process through Update
	updatedModel, cmd := model.Update(eventMsg)

	// Verify model type
	_, ok := updatedModel.(*Model)
	if !ok {
		t.Fatal("Update should return *Model")
	}

	// Verify command was returned
	if cmd == nil {
		t.Error("Update should return command to continue listening")
	}

	// Clean up
	model.eventCancel()
	close(convEvents)
	close(msgEvents)
}

// TestQuitHandlersCleanupEvents verifies that Ctrl+C cleans up event subscriptions
// Note: Esc no longer quits, only Ctrl+C and exit commands do
func TestQuitHandlersCleanupEvents(t *testing.T) {
	tests := []struct {
		name   string
		keyMsg tea.KeyMsg
	}{
		{
			name:   "Ctrl+C",
			keyMsg: tea.KeyMsg{Type: tea.KeyCtrlC},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create model
			model := NewModel("test-conv", "test-model")

			// Create mock services
			convEvents := make(chan pubsub.Event[services.Conversation], 1)
			msgEvents := make(chan pubsub.Event[services.Message], 1)

			convSvc := &mockConversationService{events: convEvents}
			msgSvc := &mockMessageService{events: msgEvents}

			model.SetServices(convSvc, msgSvc, nil)
			model.StartEventSubscriptions()

			// Verify context is active
			select {
			case <-model.eventCtx.Done():
				t.Fatal("eventCtx should not be canceled initially")
			default:
				// Good
			}

			// Send quit key
			_, cmd := model.Update(tt.keyMsg)

			// Verify quit command was returned
			if cmd == nil {
				t.Fatal("Update should return quit command")
			}

			// Verify context was canceled
			select {
			case <-model.eventCtx.Done():
				// Good - context was canceled
			case <-time.After(100 * time.Millisecond):
				t.Error("eventCtx should be canceled after quit")
			}

			// Clean up
			close(convEvents)
			close(msgEvents)
		})
	}
}

// TestClearCommand verifies that /clear command resets all context and UI state
func TestClearCommand(t *testing.T) {
	// Create model
	model := NewModel("test-conv", "test-model")

	// Add some messages
	model.AddMessage("user", "Hello")
	model.AddMessage("assistant", "Hi there!")

	// Set some state
	model.StreamingText = "streaming..."
	model.Streaming = true
	model.Status = StatusStreaming
	model.ErrorMessage = "some error"
	model.TokensInput = 100
	model.TokensOutput = 200
	model.CurrentView = ViewModeHistory
	model.SearchMode = true
	model.SearchQuery = "test query"
	model.ShowIntro = false
	model.lastKeyWasG = true
	model.helpVisible = true
	model.typewriterMode = true
	model.quickActionsMode = true
	model.quickActionsInput = "test"
	model.showSuggestions = true
	model.lastAnalyzedInput = "input"

	// Verify state before clear
	if len(model.Messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(model.Messages))
	}
	if model.StreamingText != "streaming..." {
		t.Error("Expected streaming text to be set")
	}
	if !model.Streaming {
		t.Error("Expected Streaming to be true")
	}
	if model.CurrentView != ViewModeHistory {
		t.Error("Expected CurrentView to be ViewModeHistory")
	}

	// Execute clear command
	model.ClearContext()

	// Verify all state was reset
	if len(model.Messages) != 0 {
		t.Errorf("Expected 0 messages after clear, got %d", len(model.Messages))
	}
	if model.StreamingText != "" {
		t.Errorf("Expected empty streaming text, got %q", model.StreamingText)
	}
	if model.Streaming {
		t.Error("Expected Streaming to be false")
	}
	if model.Status != StatusIdle {
		t.Errorf("Expected StatusIdle, got %v", model.Status)
	}
	if model.ErrorMessage != "" {
		t.Errorf("Expected empty error message, got %q", model.ErrorMessage)
	}
	if model.TokensInput != 0 {
		t.Errorf("Expected 0 input tokens, got %d", model.TokensInput)
	}
	if model.TokensOutput != 0 {
		t.Errorf("Expected 0 output tokens, got %d", model.TokensOutput)
	}
	if model.SearchMode {
		t.Error("Expected SearchMode to be false")
	}
	if model.SearchQuery != "" {
		t.Errorf("Expected empty search query, got %q", model.SearchQuery)
	}
	if !model.ShowIntro {
		t.Error("Expected ShowIntro to be true after clear")
	}
	if model.CurrentView != ViewModeChat {
		t.Errorf("Expected CurrentView to be ViewModeChat after clear, got %v", model.CurrentView)
	}
	if model.lastKeyWasG {
		t.Error("Expected lastKeyWasG to be false")
	}
	if model.helpVisible {
		t.Error("Expected helpVisible to be false")
	}
	if model.typewriterMode {
		t.Error("Expected typewriterMode to be false")
	}
	if model.quickActionsMode {
		t.Error("Expected quickActionsMode to be false")
	}
	if model.quickActionsInput != "" {
		t.Errorf("Expected empty quickActionsInput, got %q", model.quickActionsInput)
	}
	if model.showSuggestions {
		t.Error("Expected showSuggestions to be false")
	}
	if model.lastAnalyzedInput != "" {
		t.Errorf("Expected empty lastAnalyzedInput, got %q", model.lastAnalyzedInput)
	}
}

// TestClearCommandInput verifies that typing /clear triggers the clear
func TestClearCommandInput(t *testing.T) {
	// Create model
	model := NewModel("test-conv", "test-model")

	// Add some messages
	model.AddMessage("user", "First message")
	model.AddMessage("assistant", "Response")

	// Verify we have messages
	if len(model.Messages) != 2 {
		t.Fatalf("Expected 2 messages, got %d", len(model.Messages))
	}

	// Set input to /clear
	model.Input.SetValue("/clear")

	// Send Enter key to trigger the command
	keyMsg := tea.KeyMsg{Type: tea.KeyEnter}
	updatedModel, cmd := model.Update(keyMsg)

	// Verify model was updated
	m, ok := updatedModel.(*Model)
	if !ok {
		t.Fatal("Update should return *Model")
	}

	// Verify messages were cleared
	if len(m.Messages) != 0 {
		t.Errorf("Expected 0 messages after /clear, got %d", len(m.Messages))
	}

	// Verify intro is shown
	if !m.ShowIntro {
		t.Error("Expected ShowIntro to be true after /clear")
	}

	// Verify command was returned (nil is ok for /clear)
	_ = cmd
}
