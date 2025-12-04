# Architectural Refactor Design

**Date:** December 3, 2025
**Author:** Claude (via Hex)
**Status:** Approved
**Target:** Hex v1.4.0

## Goal

Refactor Hex to use event-driven architecture with proper service boundaries, message queuing, database migrations, and token tracking. This foundation enables LSP integration and UX improvements in subsequent releases.

## Context

The Crush audit revealed architectural patterns that improve maintainability, testability, and user experience. We will adopt these patterns before adding LSP integration (v1.5.0) and UX improvements.

Current problems:
- UI directly manipulates database state
- No separation between business logic and presentation
- Users cannot queue messages while agent works
- No token or cost tracking
- Schema changes require manual SQL updates

## Design

### 1. Event Architecture

We implement a PubSub system for decoupled communication.

**Implementation:**
```go
// internal/pubsub/broker.go
type Broker[T any] struct {
    subscribers map[chan Event[T]]bool
    mu          sync.RWMutex
}

type Event[T] struct {
    Type    EventType  // Created, Updated, Deleted
    Payload T
}
```

**Event Types:**
- ConversationEvent - conversation lifecycle
- MessageEvent - message additions and streaming
- ToolEvent - tool execution and approvals
- TokenEvent - usage tracking updates

**Flow:**
1. Services publish events when state changes
2. UI components subscribe to relevant events
3. Components update reactively without direct coupling

### 2. Service Layer

We extract business logic into interfaces with concrete implementations.

**Services:**

```go
// internal/services/conversation.go
type ConversationService interface {
    pubsub.Subscriber[Conversation]
    Create(ctx context.Context, title string) (*Conversation, error)
    Get(ctx context.Context, id int64) (*Conversation, error)
    List(ctx context.Context) ([]*Conversation, error)
    Update(ctx context.Context, conv *Conversation) error
    Delete(ctx context.Context, id int64) error
    UpdateTokenUsage(ctx context.Context, id int64, prompt, completion int64) error
}

// internal/services/message.go
type MessageService interface {
    pubsub.Subscriber[Message]
    Add(ctx context.Context, msg *Message) error
    GetByConversation(ctx context.Context, convID int64) ([]*Message, error)
    GetSummaries(ctx context.Context, convID int64) ([]*Message, error)
}

// internal/services/agent.go
type AgentService interface {
    Run(ctx context.Context, call AgentCall) (*AgentResult, error)
    Stream(ctx context.Context, call AgentCall) (<-chan StreamEvent, error)
    IsConversationBusy(convID int64) bool
    QueuedPrompts(convID int64) int
    CancelConversation(convID int64)
}
```

**Benefits:**
- Test UI without database
- Clear API boundaries
- Coordinate operations internally
- Mock for unit tests

### 3. Message Queuing

AgentService queues messages when conversations are busy.

**Implementation:**
```go
type agentService struct {
    messageQueue     *sync.Map  // map[int64][]AgentCall
    activeRequests   *sync.Map  // map[int64]context.CancelFunc
}

func (a *agentService) Run(ctx context.Context, call AgentCall) (*AgentResult, error) {
    if a.IsConversationBusy(call.ConversationID) {
        a.queueMessage(call)
        return nil, nil  // Queued, not executed
    }

    // Execute and process next queued message when done
}
```

**UI Changes:**
- Input remains enabled during responses
- Status bar shows: "Agent working... (2 queued)"
- Add "Clear queue" action
- Cancel clears queue and active request

### 4. Database Migrations

We version schema changes using `golang-migrate/migrate`.

**Structure:**
```
internal/storage/migrations/
  ├── 000001_initial_schema.up.sql
  ├── 000001_initial_schema.down.sql
  ├── 000002_add_token_tracking.up.sql
  └── 000002_add_token_tracking.down.sql
```

**Schema Additions:**
```sql
-- Token tracking
ALTER TABLE conversations ADD COLUMN prompt_tokens INTEGER DEFAULT 0;
ALTER TABLE conversations ADD COLUMN completion_tokens INTEGER DEFAULT 0;
ALTER TABLE conversations ADD COLUMN total_cost REAL DEFAULT 0.0;

-- Summary tracking
ALTER TABLE messages ADD COLUMN is_summary BOOLEAN DEFAULT 0;
ALTER TABLE conversations ADD COLUMN summary_message_id INTEGER REFERENCES messages(id);

-- Provider tracking
ALTER TABLE messages ADD COLUMN provider TEXT;
ALTER TABLE messages ADD COLUMN model TEXT;
```

**Migration Runner:**
```go
func RunMigrations(db *sql.DB) error {
    driver, _ := sqlite3.WithInstance(db, &sqlite3.Config{})
    m, _ := migrate.NewWithDatabaseInstance(
        "file://internal/storage/migrations",
        "sqlite3", driver)
    return m.Up()
}
```

Old databases migrate automatically on open.

### 5. Token Tracking

ConversationService calculates cost from token usage.

**Flow:**
1. AgentService receives usage from LLM response
2. Calls `conversationSvc.UpdateTokenUsage()`
3. Service updates totals and calculates cost
4. Publishes ConversationEvent
5. UI components update displays

**Cost Calculation:**
```go
// Sonnet 4 pricing: $3/1M input, $15/1M output
inputCost := float64(promptTokens) / 1_000_000 * 3.0
outputCost := float64(completionTokens) / 1_000_000 * 15.0
totalCost = inputCost + outputCost
```

**UI Display:**
- Status bar: "150K tokens ($0.52)"
- Conversation list shows cost per conversation
- `/stats` command for detailed breakdown

### 6. Testing Strategy

**Service Tests:**
```go
func TestConversationService_Create(t *testing.T) {
    db := setupTestDB(t)
    svc := services.NewConversationService(db)

    conv, err := svc.Create(ctx, "Test")
    require.NoError(t, err)
    assert.Equal(t, "Test", conv.Title)
}
```

**PubSub Tests:**
```go
func TestBroker_PublishSubscribe(t *testing.T) {
    broker := pubsub.NewBroker[string]()
    events := broker.Subscribe(ctx)

    broker.Publish(pubsub.Created, "test")

    event := <-events
    assert.Equal(t, pubsub.Created, event.Type)
}
```

**UI Tests with Mocks:**
```go
mockAgent := &mocks.MockAgentService{
    RunFunc: func(ctx context.Context, call AgentCall) (*AgentResult, error) {
        return &AgentResult{Text: "response"}, nil
    },
}

model := NewModel(mockConv, mockMsg, mockAgent)
```

**Coverage Target:** >80% for new code

## Implementation Plan

### Phase 1: Foundation (Days 1-2)
1. Create `internal/pubsub` with Broker
2. Create `internal/services` with interfaces
3. Set up migration system
4. Write initial migrations with token tracking
5. All tests passing

### Phase 2: Service Implementation (Days 2-3)
1. Implement ConversationService with PubSub
2. Implement MessageService with PubSub
3. Migrate storage to use migrations
4. Add token tracking to ConversationService
5. Write service tests

### Phase 3: Agent Service & Queuing (Days 3-4)
1. Implement AgentService with queue
2. Move LLM interaction from UI to service
3. Add cancellation and queue management
4. Test queuing behavior

### Phase 4: UI Integration (Days 4-5)
1. Refactor UI Model to use injected services
2. Subscribe to events instead of direct updates
3. Update viewport to be event-driven
4. Add queue status to UI
5. Remove old coupling code

### Phase 5: Testing & Polish (Days 5-6)
1. Integration tests for full flow
2. Test migration from old databases
3. Performance test with queuing
4. Update documentation

## Success Criteria

- All existing features work identically
- Tests achieve >80% coverage of new code
- Old databases migrate without data loss
- Queue handles rapid user input smoothly
- No data races or deadlocks
- Performance matches current implementation

## Next Steps

After this foundation merges:
1. Add UX improvements (persistent permissions, safe commands, mouse throttling)
2. Implement LSP integration (v1.5.0)

## Risks

**Risk:** Large refactor breaks existing functionality
**Mitigation:** Comprehensive test coverage before migration

**Risk:** Performance regression from event overhead
**Mitigation:** Benchmark before/after, optimize hot paths

**Risk:** Migration bugs with user databases
**Mitigation:** Test with real user databases before release

## Open Questions

None - design approved.
