# Event Architecture Refactor Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Refactor Hex to use event-driven architecture with service boundaries, message queuing, database migrations, and token tracking.

**Architecture:** Implement PubSub broker for decoupled events, extract business logic into service interfaces with concrete implementations, add message queuing to agent service, implement database migrations for schema versioning, add token and cost tracking.

**Tech Stack:** Go 1.24, SQLite, golang-migrate/migrate, sync.Map for concurrency

---

## Phase 1: PubSub Foundation

### Task 1: Create PubSub Package Structure

**Files:**
- Create: `internal/pubsub/broker.go`
- Create: `internal/pubsub/broker_test.go`
- Create: `internal/pubsub/types.go`

**Step 1: Write failing broker test**

```go
// internal/pubsub/broker_test.go
package pubsub

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBroker_PublishSubscribe(t *testing.T) {
	t.Parallel()
	broker := NewBroker[string]()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	events := broker.Subscribe(ctx)

	broker.Publish(Created, "test-payload")

	select {
	case event := <-events:
		assert.Equal(t, Created, event.Type)
		assert.Equal(t, "test-payload", event.Payload)
	case <-ctx.Done():
		t.Fatal("timeout waiting for event")
	}
}

func TestBroker_MultipleSubscribers(t *testing.T) {
	t.Parallel()
	broker := NewBroker[int]()

	ctx := context.Background()
	sub1 := broker.Subscribe(ctx)
	sub2 := broker.Subscribe(ctx)

	broker.Publish(Created, 42)

	val1 := <-sub1
	val2 := <-sub2

	assert.Equal(t, 42, val1.Payload)
	assert.Equal(t, 42, val2.Payload)
}

func TestBroker_Unsubscribe(t *testing.T) {
	t.Parallel()
	broker := NewBroker[string]()

	ctx, cancel := context.WithCancel(context.Background())
	events := broker.Subscribe(ctx)

	cancel() // Unsubscribe
	time.Sleep(10 * time.Millisecond)

	broker.Publish(Created, "test")

	select {
	case <-events:
		t.Fatal("received event after unsubscribe")
	case <-time.After(50 * time.Millisecond):
		// Good - channel closed or no event
	}
}
```

**Step 2: Run test to verify it fails**

```bash
cd /Users/harper/workspace/2389/clem/.worktrees/refactor-event-architecture
go test ./internal/pubsub -v
```

Expected: FAIL with "undefined: NewBroker"

**Step 3: Write event types**

```go
// internal/pubsub/types.go
// ABOUTME: Event types and interfaces for PubSub system
// ABOUTME: Defines EventType enum and Event wrapper struct

package pubsub

// EventType represents the type of event being published
type EventType int

const (
	Created EventType = iota
	Updated
	Deleted
)

// Event wraps a payload with its event type
type Event[T any] struct {
	Type    EventType
	Payload T
}

// Subscriber interface for services that publish events
type Subscriber[T any] interface {
	Subscribe(ctx context.Context) <-chan Event[T]
}
```

**Step 4: Write broker implementation**

```go
// internal/pubsub/broker.go
// ABOUTME: Generic event broker for publish-subscribe pattern
// ABOUTME: Manages subscriptions and broadcasts events to all subscribers

package pubsub

import (
	"context"
	"sync"
)

// Broker manages subscriptions and event publishing for a specific type
type Broker[T any] struct {
	subscribers map[chan Event[T]]bool
	mu          sync.RWMutex
}

// NewBroker creates a new event broker
func NewBroker[T any]() *Broker[T] {
	return &Broker[T]{
		subscribers: make(map[chan Event[T]]bool),
	}
}

// Subscribe creates a new subscription that receives events until context is cancelled
func (b *Broker[T]) Subscribe(ctx context.Context) <-chan Event[T] {
	ch := make(chan Event[T], 10) // Buffer to prevent blocking publishers

	b.mu.Lock()
	b.subscribers[ch] = true
	b.mu.Unlock()

	// Cleanup when context is cancelled
	go func() {
		<-ctx.Done()
		b.mu.Lock()
		delete(b.subscribers, ch)
		close(ch)
		b.mu.Unlock()
	}()

	return ch
}

// Publish sends an event to all subscribers
func (b *Broker[T]) Publish(eventType EventType, payload T) {
	event := Event[T]{
		Type:    eventType,
		Payload: payload,
	}

	b.mu.RLock()
	defer b.mu.RUnlock()

	for ch := range b.subscribers {
		select {
		case ch <- event:
		default:
			// Subscriber channel full, skip to prevent blocking
		}
	}
}
```

**Step 5: Run tests to verify they pass**

```bash
go test ./internal/pubsub -v
```

Expected: PASS (3 tests)

**Step 6: Commit**

```bash
git add internal/pubsub/
git commit -m "feat: add PubSub broker for event-driven architecture

Implement generic event broker with publish-subscribe pattern.
Supports multiple subscribers, automatic cleanup on context
cancellation, and buffered channels to prevent blocking.

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## Phase 2: Service Interfaces

### Task 2: Define Service Interfaces

**Files:**
- Create: `internal/services/conversation.go`
- Create: `internal/services/message.go`
- Create: `internal/services/agent.go`
- Create: `internal/services/types.go`

**Step 1: Write shared types**

```go
// internal/services/types.go
// ABOUTME: Shared types and models for service layer
// ABOUTME: Defines domain models used across services

package services

import "time"

// Conversation represents a chat session
type Conversation struct {
	ID               int64
	Title            string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	PromptTokens     int64
	CompletionTokens int64
	TotalCost        float64
	SummaryMessageID *int64
}

// Message represents a single message in a conversation
type Message struct {
	ID             int64
	ConversationID int64
	Role           string
	Content        string
	Provider       string
	Model          string
	IsSummary      bool
	CreatedAt      time.Time
}

// AgentCall represents a request to the agent
type AgentCall struct {
	ConversationID int64
	Prompt         string
	Attachments    []string
	MaxTokens      int
}

// AgentResult represents the agent's response
type AgentResult struct {
	Text         string
	ToolCalls    []ToolCall
	PromptTokens int64
	OutputTokens int64
}

// ToolCall represents a tool execution request
type ToolCall struct {
	ID     string
	Name   string
	Input  map[string]interface{}
	Result string
}

// StreamEvent represents a streaming update from the agent
type StreamEvent struct {
	Type         string // "text", "tool_call", "done", "error"
	Text         string
	ToolCall     *ToolCall
	Error        error
	PromptTokens int64
	OutputTokens int64
}
```

**Step 2: Write ConversationService interface**

```go
// internal/services/conversation.go
// ABOUTME: Service interface for conversation management
// ABOUTME: Handles CRUD operations and token tracking for conversations

package services

import (
	"context"

	"github.com/2389-research/hex/internal/pubsub"
)

// ConversationService manages conversation lifecycle and state
type ConversationService interface {
	pubsub.Subscriber[Conversation]

	// Create creates a new conversation
	Create(ctx context.Context, title string) (*Conversation, error)

	// Get retrieves a conversation by ID
	Get(ctx context.Context, id int64) (*Conversation, error)

	// List returns all conversations ordered by updated_at DESC
	List(ctx context.Context) ([]*Conversation, error)

	// Update saves conversation changes
	Update(ctx context.Context, conv *Conversation) error

	// Delete removes a conversation and its messages
	Delete(ctx context.Context, id int64) error

	// UpdateTokenUsage updates token counts and calculates cost
	UpdateTokenUsage(ctx context.Context, id int64, promptTokens, completionTokens int64) error
}
```

**Step 3: Write MessageService interface**

```go
// internal/services/message.go
// ABOUTME: Service interface for message management
// ABOUTME: Handles message storage and retrieval with event publishing

package services

import (
	"context"

	"github.com/2389-research/hex/internal/pubsub"
)

// MessageService manages message lifecycle
type MessageService interface {
	pubsub.Subscriber[Message]

	// Add stores a new message
	Add(ctx context.Context, msg *Message) error

	// GetByConversation returns all messages for a conversation
	GetByConversation(ctx context.Context, convID int64) ([]*Message, error)

	// GetSummaries returns only summary messages for a conversation
	GetSummaries(ctx context.Context, convID int64) ([]*Message, error)
}
```

**Step 4: Write AgentService interface**

```go
// internal/services/agent.go
// ABOUTME: Service interface for LLM agent interactions
// ABOUTME: Handles message queuing, execution, and conversation state

package services

import "context"

// AgentService coordinates LLM interactions with queuing support
type AgentService interface {
	// Run executes a prompt (queues if conversation is busy)
	Run(ctx context.Context, call AgentCall) (*AgentResult, error)

	// Stream executes a prompt with streaming response
	Stream(ctx context.Context, call AgentCall) (<-chan StreamEvent, error)

	// IsConversationBusy returns true if conversation has active request
	IsConversationBusy(convID int64) bool

	// QueuedPrompts returns number of queued messages for conversation
	QueuedPrompts(convID int64) int

	// CancelConversation cancels active request and clears queue
	CancelConversation(convID int64)
}
```

**Step 5: Commit**

```bash
git add internal/services/
git commit -m "feat: define service interfaces for event architecture

Add ConversationService, MessageService, and AgentService interfaces
with domain models. Services will publish events and coordinate
business logic independently of UI layer.

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## Phase 3: Database Migrations

### Task 3: Set Up Migration System

**Files:**
- Create: `internal/storage/migrations/000001_initial_schema.up.sql`
- Create: `internal/storage/migrations/000001_initial_schema.down.sql`
- Create: `internal/storage/migrations/000002_add_token_tracking.up.sql`
- Create: `internal/storage/migrations/000002_add_token_tracking.down.sql`
- Modify: `internal/storage/storage.go`
- Modify: `go.mod`

**Step 1: Add migration dependency**

```bash
go get -u github.com/golang-migrate/migrate/v4
go get -u github.com/golang-migrate/migrate/v4/database/sqlite3
go get -u github.com/golang-migrate/migrate/v4/source/file
```

**Step 2: Extract current schema to initial migration**

```sql
-- internal/storage/migrations/000001_initial_schema.up.sql
CREATE TABLE IF NOT EXISTS conversations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    conversation_id INTEGER NOT NULL,
    role TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_messages_conversation
    ON messages(conversation_id);

CREATE INDEX IF NOT EXISTS idx_conversations_updated
    ON conversations(updated_at DESC);
```

```sql
-- internal/storage/migrations/000001_initial_schema.down.sql
DROP INDEX IF EXISTS idx_conversations_updated;
DROP INDEX IF EXISTS idx_messages_conversation;
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS conversations;
```

**Step 3: Create token tracking migration**

```sql
-- internal/storage/migrations/000002_add_token_tracking.up.sql
-- Add token tracking to conversations
ALTER TABLE conversations ADD COLUMN prompt_tokens INTEGER DEFAULT 0;
ALTER TABLE conversations ADD COLUMN completion_tokens INTEGER DEFAULT 0;
ALTER TABLE conversations ADD COLUMN total_cost REAL DEFAULT 0.0;

-- Add summary tracking
ALTER TABLE messages ADD COLUMN is_summary BOOLEAN DEFAULT 0;
ALTER TABLE conversations ADD COLUMN summary_message_id INTEGER REFERENCES messages(id);

-- Add provider tracking to messages
ALTER TABLE messages ADD COLUMN provider TEXT;
ALTER TABLE messages ADD COLUMN model TEXT;

-- Add indexes for performance
CREATE INDEX IF NOT EXISTS idx_messages_is_summary
    ON messages(conversation_id, is_summary)
    WHERE is_summary = 1;
```

```sql
-- internal/storage/migrations/000002_add_token_tracking.down.sql
DROP INDEX IF EXISTS idx_messages_is_summary;

-- SQLite doesn't support DROP COLUMN, so we'd need to recreate tables
-- For now, down migration is not supported for this change
-- In production, we'd implement full table recreation
```

**Step 4: Modify storage.go to run migrations**

Add to `internal/storage/storage.go` after imports:

```go
import (
	"database/sql"
	"embed"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func RunMigrations(db *sql.DB) error {
	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	sourceDriver, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("failed to create migration source: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", sourceDriver, "sqlite3", driver)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration failed: %w", err)
	}

	return nil
}
```

Modify `OpenDatabase` function to call migrations:

```go
func OpenDatabase(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, err
	}

	// Run migrations
	if err := RunMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}
```

**Step 5: Test migrations**

```bash
# Remove old test database if exists
rm -f /tmp/test-migrations.db

# Create test to verify migrations work
cd /Users/harper/workspace/2389/clem/.worktrees/refactor-event-architecture
go test ./internal/storage -v -run TestMigrations
```

Create test in `internal/storage/storage_test.go`:

```go
func TestMigrations(t *testing.T) {
	// Create temporary database
	tmpDB := filepath.Join(t.TempDir(), "test.db")

	db, err := OpenDatabase(tmpDB)
	require.NoError(t, err)
	defer db.Close()

	// Verify tables exist
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table'").Scan(&count)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, count, 2) // At least conversations and messages

	// Verify new columns exist in conversations
	rows, err := db.Query("PRAGMA table_info(conversations)")
	require.NoError(t, err)
	defer rows.Close()

	columns := make(map[string]bool)
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dfltValue sql.NullString
		require.NoError(t, rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk))
		columns[name] = true
	}

	assert.True(t, columns["prompt_tokens"], "missing prompt_tokens column")
	assert.True(t, columns["completion_tokens"], "missing completion_tokens column")
	assert.True(t, columns["total_cost"], "missing total_cost column")
}
```

**Step 6: Commit**

```bash
git add internal/storage/migrations/ internal/storage/storage.go internal/storage/storage_test.go go.mod go.sum
git commit -m "feat: add database migration system with token tracking

Implement golang-migrate for schema versioning. Migrate existing
schema to initial migration, add token tracking fields to
conversations and messages tables.

Migrations run automatically on database open.

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## Phase 4: Implement Services

### Task 4: Implement ConversationService

**Files:**
- Create: `internal/services/conversation_impl.go`
- Create: `internal/services/conversation_test.go`

**Step 1: Write failing test**

```go
// internal/services/conversation_test.go
package services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/2389-research/hex/internal/pubsub"
	"github.com/2389-research/hex/internal/storage"
)

func setupTestDB(t *testing.T) *sql.DB {
	tmpDB := filepath.Join(t.TempDir(), "test.db")
	db, err := storage.OpenDatabase(tmpDB)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db
}

func TestConversationService_Create(t *testing.T) {
	db := setupTestDB(t)
	svc := NewConversationService(db)

	conv, err := svc.Create(context.Background(), "Test Conversation")
	require.NoError(t, err)

	assert.NotZero(t, conv.ID)
	assert.Equal(t, "Test Conversation", conv.Title)
	assert.NotZero(t, conv.CreatedAt)
}

func TestConversationService_PublishesEvents(t *testing.T) {
	db := setupTestDB(t)
	svc := NewConversationService(db)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	events := svc.Subscribe(ctx)

	conv, err := svc.Create(context.Background(), "Test")
	require.NoError(t, err)

	select {
	case event := <-events:
		assert.Equal(t, pubsub.Created, event.Type)
		assert.Equal(t, conv.ID, event.Payload.ID)
	case <-ctx.Done():
		t.Fatal("timeout waiting for event")
	}
}

func TestConversationService_UpdateTokenUsage(t *testing.T) {
	db := setupTestDB(t)
	svc := NewConversationService(db)

	conv, err := svc.Create(context.Background(), "Test")
	require.NoError(t, err)

	err = svc.UpdateTokenUsage(context.Background(), conv.ID, 1000, 5000)
	require.NoError(t, err)

	updated, err := svc.Get(context.Background(), conv.ID)
	require.NoError(t, err)

	assert.Equal(t, int64(1000), updated.PromptTokens)
	assert.Equal(t, int64(5000), updated.CompletionTokens)

	// Cost calculation: (1000/1M * $3) + (5000/1M * $15)
	expectedCost := (1000.0/1_000_000*3.0) + (5000.0/1_000_000*15.0)
	assert.InDelta(t, expectedCost, updated.TotalCost, 0.0001)
}
```

**Step 2: Run test to verify it fails**

```bash
go test ./internal/services -v -run TestConversation
```

Expected: FAIL with "undefined: NewConversationService"

**Step 3: Implement ConversationService**

```go
// internal/services/conversation_impl.go
// ABOUTME: Concrete implementation of ConversationService
// ABOUTME: Manages conversation CRUD with event publishing and token tracking

package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/2389-research/hex/internal/pubsub"
)

type conversationService struct {
	*pubsub.Broker[Conversation]
	db *sql.DB
}

// NewConversationService creates a new conversation service
func NewConversationService(db *sql.DB) ConversationService {
	return &conversationService{
		Broker: pubsub.NewBroker[Conversation](),
		db:     db,
	}
}

func (s *conversationService) Create(ctx context.Context, title string) (*Conversation, error) {
	now := time.Now()

	result, err := s.db.ExecContext(ctx,
		`INSERT INTO conversations (title, created_at, updated_at) VALUES (?, ?, ?)`,
		title, now, now)
	if err != nil {
		return nil, fmt.Errorf("failed to create conversation: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation ID: %w", err)
	}

	conv := &Conversation{
		ID:        id,
		Title:     title,
		CreatedAt: now,
		UpdatedAt: now,
	}

	s.Publish(pubsub.Created, *conv)
	return conv, nil
}

func (s *conversationService) Get(ctx context.Context, id int64) (*Conversation, error) {
	conv := &Conversation{}
	var summaryID sql.NullInt64

	err := s.db.QueryRowContext(ctx,
		`SELECT id, title, created_at, updated_at,
		        prompt_tokens, completion_tokens, total_cost, summary_message_id
		 FROM conversations WHERE id = ?`, id).Scan(
		&conv.ID, &conv.Title, &conv.CreatedAt, &conv.UpdatedAt,
		&conv.PromptTokens, &conv.CompletionTokens, &conv.TotalCost, &summaryID)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("conversation not found: %d", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}

	if summaryID.Valid {
		conv.SummaryMessageID = &summaryID.Int64
	}

	return conv, nil
}

func (s *conversationService) List(ctx context.Context) ([]*Conversation, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, title, created_at, updated_at,
		        prompt_tokens, completion_tokens, total_cost, summary_message_id
		 FROM conversations ORDER BY updated_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("failed to list conversations: %w", err)
	}
	defer rows.Close()

	var conversations []*Conversation
	for rows.Next() {
		conv := &Conversation{}
		var summaryID sql.NullInt64

		err := rows.Scan(&conv.ID, &conv.Title, &conv.CreatedAt, &conv.UpdatedAt,
			&conv.PromptTokens, &conv.CompletionTokens, &conv.TotalCost, &summaryID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan conversation: %w", err)
		}

		if summaryID.Valid {
			conv.SummaryMessageID = &summaryID.Int64
		}

		conversations = append(conversations, conv)
	}

	return conversations, nil
}

func (s *conversationService) Update(ctx context.Context, conv *Conversation) error {
	conv.UpdatedAt = time.Now()

	var summaryID sql.NullInt64
	if conv.SummaryMessageID != nil {
		summaryID = sql.NullInt64{Int64: *conv.SummaryMessageID, Valid: true}
	}

	_, err := s.db.ExecContext(ctx,
		`UPDATE conversations
		 SET title = ?, updated_at = ?,
		     prompt_tokens = ?, completion_tokens = ?, total_cost = ?,
		     summary_message_id = ?
		 WHERE id = ?`,
		conv.Title, conv.UpdatedAt,
		conv.PromptTokens, conv.CompletionTokens, conv.TotalCost,
		summaryID, conv.ID)

	if err != nil {
		return fmt.Errorf("failed to update conversation: %w", err)
	}

	s.Publish(pubsub.Updated, *conv)
	return nil
}

func (s *conversationService) Delete(ctx context.Context, id int64) error {
	conv, err := s.Get(ctx, id)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx, `DELETE FROM conversations WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete conversation: %w", err)
	}

	s.Publish(pubsub.Deleted, *conv)
	return nil
}

func (s *conversationService) UpdateTokenUsage(ctx context.Context, id int64, promptTokens, completionTokens int64) error {
	conv, err := s.Get(ctx, id)
	if err != nil {
		return err
	}

	conv.PromptTokens += promptTokens
	conv.CompletionTokens += completionTokens

	// Calculate cost: $3/1M input, $15/1M output (Sonnet 4 pricing)
	inputCost := float64(conv.PromptTokens) / 1_000_000 * 3.0
	outputCost := float64(conv.CompletionTokens) / 1_000_000 * 15.0
	conv.TotalCost = inputCost + outputCost

	return s.Update(ctx, conv)
}
```

**Step 4: Run tests to verify they pass**

```bash
go test ./internal/services -v -run TestConversation
```

Expected: PASS (3 tests)

**Step 5: Commit**

```bash
git add internal/services/
git commit -m "feat: implement ConversationService with token tracking

Add concrete implementation with PubSub event publishing.
Calculates cost automatically based on token usage.
Full test coverage for CRUD operations and events.

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

### Task 5: Implement MessageService

**Files:**
- Create: `internal/services/message_impl.go`
- Create: `internal/services/message_test.go`

**Step 1: Write failing test**

```go
// internal/services/message_test.go
package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/2389-research/hex/internal/pubsub"
)

func TestMessageService_Add(t *testing.T) {
	db := setupTestDB(t)
	convSvc := NewConversationService(db)
	msgSvc := NewMessageService(db)

	conv, err := convSvc.Create(context.Background(), "Test")
	require.NoError(t, err)

	msg := &Message{
		ConversationID: conv.ID,
		Role:           "user",
		Content:        "Hello",
		Provider:       "anthropic",
		Model:          "claude-sonnet-4",
	}

	err = msgSvc.Add(context.Background(), msg)
	require.NoError(t, err)
	assert.NotZero(t, msg.ID)
	assert.NotZero(t, msg.CreatedAt)
}

func TestMessageService_GetByConversation(t *testing.T) {
	db := setupTestDB(t)
	convSvc := NewConversationService(db)
	msgSvc := NewMessageService(db)

	conv, _ := convSvc.Create(context.Background(), "Test")

	msg1 := &Message{ConversationID: conv.ID, Role: "user", Content: "First"}
	msg2 := &Message{ConversationID: conv.ID, Role: "assistant", Content: "Second"}

	msgSvc.Add(context.Background(), msg1)
	msgSvc.Add(context.Background(), msg2)

	messages, err := msgSvc.GetByConversation(context.Background(), conv.ID)
	require.NoError(t, err)
	assert.Len(t, messages, 2)
	assert.Equal(t, "First", messages[0].Content)
	assert.Equal(t, "Second", messages[1].Content)
}

func TestMessageService_GetSummaries(t *testing.T) {
	db := setupTestDB(t)
	convSvc := NewConversationService(db)
	msgSvc := NewMessageService(db)

	conv, _ := convSvc.Create(context.Background(), "Test")

	regular := &Message{ConversationID: conv.ID, Role: "user", Content: "Regular"}
	summary := &Message{ConversationID: conv.ID, Role: "assistant", Content: "Summary", IsSummary: true}

	msgSvc.Add(context.Background(), regular)
	msgSvc.Add(context.Background(), summary)

	summaries, err := msgSvc.GetSummaries(context.Background(), conv.ID)
	require.NoError(t, err)
	assert.Len(t, summaries, 1)
	assert.Equal(t, "Summary", summaries[0].Content)
	assert.True(t, summaries[0].IsSummary)
}
```

**Step 2: Run test to verify it fails**

```bash
go test ./internal/services -v -run TestMessage
```

Expected: FAIL with "undefined: NewMessageService"

**Step 3: Implement MessageService**

```go
// internal/services/message_impl.go
// ABOUTME: Concrete implementation of MessageService
// ABOUTME: Manages message storage with event publishing

package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/2389-research/hex/internal/pubsub"
)

type messageService struct {
	*pubsub.Broker[Message]
	db *sql.DB
}

// NewMessageService creates a new message service
func NewMessageService(db *sql.DB) MessageService {
	return &messageService{
		Broker: pubsub.NewBroker[Message](),
		db:     db,
	}
}

func (s *messageService) Add(ctx context.Context, msg *Message) error {
	msg.CreatedAt = time.Now()

	result, err := s.db.ExecContext(ctx,
		`INSERT INTO messages (conversation_id, role, content, provider, model, is_summary, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		msg.ConversationID, msg.Role, msg.Content, msg.Provider, msg.Model, msg.IsSummary, msg.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to add message: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get message ID: %w", err)
	}

	msg.ID = id
	s.Publish(pubsub.Created, *msg)
	return nil
}

func (s *messageService) GetByConversation(ctx context.Context, convID int64) ([]*Message, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, conversation_id, role, content, provider, model, is_summary, created_at
		 FROM messages WHERE conversation_id = ? ORDER BY created_at ASC`,
		convID)

	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}
	defer rows.Close()

	return s.scanMessages(rows)
}

func (s *messageService) GetSummaries(ctx context.Context, convID int64) ([]*Message, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, conversation_id, role, content, provider, model, is_summary, created_at
		 FROM messages WHERE conversation_id = ? AND is_summary = 1 ORDER BY created_at ASC`,
		convID)

	if err != nil {
		return nil, fmt.Errorf("failed to get summaries: %w", err)
	}
	defer rows.Close()

	return s.scanMessages(rows)
}

func (s *messageService) scanMessages(rows *sql.Rows) ([]*Message, error) {
	var messages []*Message
	for rows.Next() {
		msg := &Message{}
		var provider, model sql.NullString

		err := rows.Scan(&msg.ID, &msg.ConversationID, &msg.Role, &msg.Content,
			&provider, &model, &msg.IsSummary, &msg.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}

		if provider.Valid {
			msg.Provider = provider.String
		}
		if model.Valid {
			msg.Model = model.String
		}

		messages = append(messages, msg)
	}

	return messages, nil
}
```

**Step 4: Run tests to verify they pass**

```bash
go test ./internal/services -v -run TestMessage
```

Expected: PASS (3 tests)

**Step 5: Commit**

```bash
git add internal/services/message_impl.go internal/services/message_test.go
git commit -m "feat: implement MessageService with event publishing

Add message storage service with PubSub events.
Supports filtering by conversation and summary status.
Full test coverage.

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## Phase 5: Agent Service with Queuing

### Task 6: Implement AgentService

**Files:**
- Create: `internal/services/agent_impl.go`
- Create: `internal/services/agent_test.go`

**Step 1: Write failing test**

```go
// internal/services/agent_test.go
package services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/2389-research/hex/internal/core"
)

// MockClient implements a simple mock of core.Client for testing
type MockClient struct {
	Response string
	Tokens   core.Usage
	Delay    time.Duration
}

func (m *MockClient) CreateMessageStream(ctx context.Context, messages []core.Message, maxTokens int) (<-chan core.StreamChunk, error) {
	ch := make(chan core.StreamChunk, 1)

	go func() {
		defer close(ch)

		if m.Delay > 0 {
			time.Sleep(m.Delay)
		}

		ch <- core.StreamChunk{
			Type: core.ChunkTypeContentDelta,
			Text: m.Response,
		}

		ch <- core.StreamChunk{
			Type:  core.ChunkTypeDone,
			Usage: &m.Tokens,
		}
	}()

	return ch, nil
}

func TestAgentService_Run(t *testing.T) {
	db := setupTestDB(t)
	convSvc := NewConversationService(db)
	msgSvc := NewMessageService(db)

	client := &MockClient{
		Response: "Test response",
		Tokens:   core.Usage{InputTokens: 100, OutputTokens: 50},
	}

	agentSvc := NewAgentService(client, convSvc, msgSvc)

	conv, _ := convSvc.Create(context.Background(), "Test")

	result, err := agentSvc.Run(context.Background(), AgentCall{
		ConversationID: conv.ID,
		Prompt:         "Hello",
	})

	require.NoError(t, err)
	assert.Equal(t, "Test response", result.Text)
	assert.Equal(t, int64(100), result.PromptTokens)
	assert.Equal(t, int64(50), result.OutputTokens)
}

func TestAgentService_Queuing(t *testing.T) {
	db := setupTestDB(t)
	convSvc := NewConversationService(db)
	msgSvc := NewMessageService(db)

	client := &MockClient{
		Response: "Response",
		Delay:    100 * time.Millisecond, // Simulate slow response
	}

	agentSvc := NewAgentService(client, convSvc, msgSvc)

	conv, _ := convSvc.Create(context.Background(), "Test")

	// Start first request (will be slow)
	go agentSvc.Run(context.Background(), AgentCall{
		ConversationID: conv.ID,
		Prompt:         "First",
	})

	// Wait a bit to ensure first request starts
	time.Sleep(10 * time.Millisecond)

	// Second request should be queued
	assert.True(t, agentSvc.IsConversationBusy(conv.ID))

	result, err := agentSvc.Run(context.Background(), AgentCall{
		ConversationID: conv.ID,
		Prompt:         "Second",
	})

	// Should return nil immediately (queued)
	assert.NoError(t, err)
	assert.Nil(t, result)
	assert.Equal(t, 1, agentSvc.QueuedPrompts(conv.ID))
}

func TestAgentService_Cancel(t *testing.T) {
	db := setupTestDB(t)
	convSvc := NewConversationService(db)
	msgSvc := NewMessageService(db)

	client := &MockClient{
		Response: "Response",
		Delay:    1 * time.Second,
	}

	agentSvc := NewAgentService(client, convSvc, msgSvc)

	conv, _ := convSvc.Create(context.Background(), "Test")

	// Start request
	go agentSvc.Run(context.Background(), AgentCall{
		ConversationID: conv.ID,
		Prompt:         "Test",
	})

	time.Sleep(10 * time.Millisecond)
	assert.True(t, agentSvc.IsConversationBusy(conv.ID))

	// Cancel
	agentSvc.CancelConversation(conv.ID)

	time.Sleep(10 * time.Millisecond)
	assert.False(t, agentSvc.IsConversationBusy(conv.ID))
}
```

**Step 2: Run test to verify it fails**

```bash
go test ./internal/services -v -run TestAgent
```

Expected: FAIL with "undefined: NewAgentService"

**Step 3: Implement AgentService** (continuing in next message due to length...)

Would you like me to continue with the full implementation plan, or would you prefer to start executing what we have so far?
