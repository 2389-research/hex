// Package subagents provides isolated execution contexts for subagent tasks.
//
// ABOUTME: Context isolation for subagent execution
// ABOUTME: Ensures subagents run with separate conversation history and working memory
package subagents

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// IsolatedContext represents an isolated execution context for a subagent
// Each subagent gets its own context that cannot see the parent's history
type IsolatedContext struct {
	// ID is a unique identifier for this context
	ID string

	// ParentID is the ID of the parent context (if any)
	ParentID string

	// Type is the subagent type
	Type SubagentType

	// CreatedAt is when this context was created
	CreatedAt time.Time

	// ConversationHistory stores messages specific to this subagent
	// This is isolated from the parent agent's history
	ConversationHistory []Message

	// WorkingMemory stores ephemeral data for this execution
	WorkingMemory map[string]interface{}

	// mu protects concurrent access to conversation history and working memory
	mu sync.RWMutex
}

// Message represents a message in the subagent's isolated conversation
type Message struct {
	Role      string
	Content   string
	Timestamp time.Time
}

// NewIsolatedContext creates a new isolated context for a subagent
func NewIsolatedContext(parentID string, agentType SubagentType) *IsolatedContext {
	return &IsolatedContext{
		ID:                  generateContextID(),
		ParentID:            parentID,
		Type:                agentType,
		CreatedAt:           time.Now(),
		ConversationHistory: make([]Message, 0),
		WorkingMemory:       make(map[string]interface{}),
	}
}

// AddMessage adds a message to this context's isolated conversation history
func (c *IsolatedContext) AddMessage(role, content string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.ConversationHistory = append(c.ConversationHistory, Message{
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
	})
}

// GetMessages returns a copy of the conversation history
// This prevents external code from modifying the internal state
func (c *IsolatedContext) GetMessages() []Message {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Return a copy to prevent external modification
	messages := make([]Message, len(c.ConversationHistory))
	copy(messages, c.ConversationHistory)
	return messages
}

// SetMemory stores a value in working memory
func (c *IsolatedContext) SetMemory(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.WorkingMemory[key] = value
}

// GetMemory retrieves a value from working memory
func (c *IsolatedContext) GetMemory(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	value, ok := c.WorkingMemory[key]
	return value, ok
}

// Clear resets the context (useful for testing or cleanup)
func (c *IsolatedContext) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.ConversationHistory = make([]Message, 0)
	c.WorkingMemory = make(map[string]interface{})
}

// MessageCount returns the number of messages in this context
func (c *IsolatedContext) MessageCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.ConversationHistory)
}

// ContextManager manages isolated contexts for multiple subagents
type ContextManager struct {
	contexts map[string]*IsolatedContext
	mu       sync.RWMutex
}

// NewContextManager creates a new context manager
func NewContextManager() *ContextManager {
	return &ContextManager{
		contexts: make(map[string]*IsolatedContext),
	}
}

// CreateContext creates a new isolated context
func (m *ContextManager) CreateContext(parentID string, agentType SubagentType) *IsolatedContext {
	m.mu.Lock()
	defer m.mu.Unlock()

	ctx := NewIsolatedContext(parentID, agentType)
	m.contexts[ctx.ID] = ctx
	return ctx
}

// GetContext retrieves a context by ID
func (m *ContextManager) GetContext(id string) (*IsolatedContext, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ctx, ok := m.contexts[id]
	if !ok {
		return nil, fmt.Errorf("context not found: %s", id)
	}
	return ctx, nil
}

// DeleteContext removes a context by ID
func (m *ContextManager) DeleteContext(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.contexts[id]; !ok {
		return fmt.Errorf("context not found: %s", id)
	}

	delete(m.contexts, id)
	return nil
}

// ListContexts returns all active context IDs
func (m *ContextManager) ListContexts() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ids := make([]string, 0, len(m.contexts))
	for id := range m.contexts {
		ids = append(ids, id)
	}
	return ids
}

// ContextCount returns the number of active contexts
func (m *ContextManager) ContextCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.contexts)
}

// CleanupOldContexts removes contexts older than the specified duration
func (m *ContextManager) CleanupOldContexts(maxAge time.Duration) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	cutoff := time.Now().Add(-maxAge)
	removed := 0

	for id, ctx := range m.contexts {
		if ctx.CreatedAt.Before(cutoff) {
			delete(m.contexts, id)
			removed++
		}
	}

	return removed
}

// generateContextID generates a unique context ID
var (
	contextIDCounter int64
	contextIDMutex   sync.Mutex
)

func generateContextID() string {
	contextIDMutex.Lock()
	defer contextIDMutex.Unlock()

	contextIDCounter++
	// Combine timestamp and counter for uniqueness
	return fmt.Sprintf("ctx_%d_%d", time.Now().UnixNano(), contextIDCounter)
}

// ExecutionContext wraps a standard context.Context with subagent-specific data
type ExecutionContext struct {
	context.Context
	Isolated *IsolatedContext
}

// NewExecutionContext creates a context for subagent execution
func NewExecutionContext(ctx context.Context, isolated *IsolatedContext) *ExecutionContext {
	return &ExecutionContext{
		Context:  ctx,
		Isolated: isolated,
	}
}
