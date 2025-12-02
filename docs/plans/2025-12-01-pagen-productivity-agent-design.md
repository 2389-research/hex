# Pagen - Personal Productivity Agent Design

**Date**: 2025-12-01
**Status**: Design Complete - Ready for Implementation
**Authors**: Harper, Claude (pagen_architect)

## Executive Summary

Pagen is a personal productivity agent forked from Clem (Claude Code CLI). Instead of code manipulation tools, Pagen provides productivity-focused tools (email, calendar, tasks) routed through pluggable provider implementations. The agent uses domain-specific tool primitives and subagent composition to build emergent productivity workflows.

## Design Decisions

### Key Choices

1. **Provider Architecture**: In-process interface-based plugins (v1) → HashiCorp go-plugin (future v2)
2. **Tool Granularity**: Domain-specific primitives (19 tools) over generic abstractions
3. **Provider Routing**: Single active provider (v1) → Intelligent multi-provider routing (v2)
4. **Skill Development**: Emergent via subagent composition + conversation memory

### Trade-offs Accepted

- **Go-only providers** (v1): Simpler to build, easier to debug vs polyglot flexibility
- **More tools per provider**: Better agent mental models vs more implementation work
- **Single provider initially**: Faster to ship vs ultimate flexibility

## Architecture

### Three-Layer Architecture

```
┌─────────────────────────────────────────────────────────┐
│ Layer 1: Agent Core (forked from Clem)                  │
│ - Bubbletea TUI                                          │
│ - Anthropic API client with streaming                   │
│ - Conversation persistence (SQLite)                      │
│ - Tool execution loop                                    │
│ - Keep: Task (subagents), AskUserQuestion, TodoWrite,   │
│         WebFetch, WebSearch                              │
└─────────────────────────────────────────────────────────┘
                          │
                          │ Calls productivity tools
                          ▼
┌─────────────────────────────────────────────────────────┐
│ Layer 2: Tool Router & Registry                         │
│ - Domain-specific tool definitions (19 tools)           │
│ - Provider interface contract                           │
│ - Active provider selection (v1: single provider)       │
│ - Tool → Provider routing (future: intelligent)         │
└─────────────────────────────────────────────────────────┘
                          │
                          │ Routes to active provider
                          ▼
┌─────────────────────────────────────────────────────────┐
│ Layer 3: Provider Plugins (in-process, interface-based) │
│                                                          │
│ ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│ │Gmail Provider│  │Outlook Plugin│  │Future: Notion│  │
│ │- Auth        │  │- Auth        │  │   Linear     │  │
│ │- API calls   │  │- API calls   │  │   Slack      │  │
│ └──────────────┘  └──────────────┘  └──────────────┘  │
└─────────────────────────────────────────────────────────┘
```

### Changes from Clem

**Remove**:
- File tools: Read, Write, Edit, Grep, Glob
- Code-focused tools and workflows

**Keep**:
- Core: Conversation system, TUI, SQLite persistence, streaming
- Meta-tools: Task (subagents), AskUserQuestion, TodoWrite
- Web: WebFetch, WebSearch

**Add**:
- 19 productivity tool primitives (email, calendar, tasks)
- Provider system (interface, registry, routing)
- OAuth authentication flows
- Provider CLI commands

## Provider Interface

### Go Interface Definition

```go
type Provider interface {
    // Metadata
    Name() string                    // "gmail", "outlook", etc.
    SupportedTools() []string        // which tools this provider implements

    // Lifecycle
    Initialize(config map[string]string) error  // setup, auth
    Authenticate() error             // OAuth flow or API key
    Close() error                    // cleanup

    // Health (for future routing)
    Status() ProviderStatus          // healthy, degraded, offline
    Capabilities() ProviderCapabilities  // rate limits, features

    // Tool execution
    ExecuteTool(toolName string, params map[string]interface{}) (ToolResult, error)
}

type ProviderStatus struct {
    Healthy   bool
    Message   string
    LastCheck time.Time
}

type ProviderCapabilities struct {
    RateLimits map[string]int  // tool -> calls per hour
    Features   []string         // "attachments", "calendar_sharing", etc.
}

type ToolResult struct {
    Success  bool                    `json:"success"`
    Data     interface{}            `json:"data,omitempty"`
    Error    string                 `json:"error,omitempty"`
    Metadata map[string]interface{} `json:"metadata,omitempty"`
}
```

### Provider Registry Pattern

```go
// internal/providers/registry.go
type ProviderRegistry struct {
    providers map[string]Provider  // name -> provider instance
    active    string                // currently active provider
    mu        sync.RWMutex
}

func (r *ProviderRegistry) Register(p Provider) error
func (r *ProviderRegistry) SetActive(name string) error
func (r *ProviderRegistry) ExecuteTool(toolName string, params map[string]interface{}) (ToolResult, error)
func (r *ProviderRegistry) GetProvider(name string) (Provider, error)
func (r *ProviderRegistry) ListProviders() []ProviderInfo
```

## Tool Primitives

### Email Domain (8 tools)

1. **send_email**(to, subject, body, cc?, bcc?, attachments?)
   - Send new email message
   - Supports multiple recipients, CC, BCC, file attachments

2. **reply_email**(message_id, body, reply_all?)
   - Reply to existing email
   - Option to reply-all

3. **search_emails**(query, from?, to?, after?, before?, is_unread?)
   - Search emails with filters
   - Returns list of matching messages

4. **read_email**(message_id)
   - Get full email content and metadata
   - Returns sender, recipients, subject, body, attachments

5. **archive_email**(message_id)
   - Move email to archive/remove from inbox

6. **mark_email_read**(message_id) / **mark_email_unread**(message_id)
   - Change read/unread status

7. **label_email**(message_id, label)
   - Add label/tag to email

8. **delete_email**(message_id)
   - Delete email (move to trash)

### Calendar Domain (6 tools)

1. **create_event**(title, start, end, attendees?, location?, description?)
   - Create new calendar event
   - Supports inviting attendees, location, description

2. **update_event**(event_id, ...updates)
   - Modify existing event
   - Can update any field (title, time, attendees, etc.)

3. **delete_event**(event_id)
   - Delete calendar event

4. **list_events**(start_date, end_date, calendar?)
   - List events in date range
   - Optional calendar filter (for multi-calendar setups)

5. **search_events**(query, start?, end?)
   - Search events by title/description
   - Optional date range filter

6. **find_free_time**(duration_minutes, within_days, attendees?)
   - Find available time slots
   - Optionally check attendee availability

### Tasks Domain (5 tools)

1. **create_task**(title, due_date?, notes?, priority?)
   - Create new task/todo item
   - Supports due dates, notes, priority levels

2. **update_task**(task_id, ...updates)
   - Modify existing task
   - Can update title, due date, notes, priority

3. **complete_task**(task_id)
   - Mark task as completed

4. **list_tasks**(filter?, due_before?)
   - List tasks with optional filtering
   - Filter by status, priority, due date

5. **delete_task**(task_id)
   - Delete task

### Total: 19 Tool Primitives

## Authentication

### OAuth Flow (Gmail, Outlook, etc.)

**Pattern**: Local callback server (similar to `gh auth login`)

```go
func (p *GmailProvider) Authenticate() error {
    // 1. Start local HTTP server on random port
    callbackServer := startCallbackServer()
    redirectURI := fmt.Sprintf("http://localhost:%d/callback", callbackServer.Port)

    // 2. Generate OAuth URL
    authURL := p.oauth2Config.AuthCodeURL("state",
        oauth2.AccessTypeOffline,
        oauth2.ApprovalForce)

    // 3. Open browser (or print URL if can't open)
    fmt.Printf("Opening browser for authentication...\n")
    fmt.Printf("If browser doesn't open, visit: %s\n", authURL)
    browser.OpenURL(authURL)

    // 4. Wait for callback with auth code
    authCode := <-callbackServer.CodeChan

    // 5. Exchange code for token
    token, err := p.oauth2Config.Exchange(context.Background(), authCode)
    if err != nil {
        return err
    }

    // 6. Store token for future use
    return p.saveToken(token)
}
```

### Configuration Structure

```yaml
# ~/.pagen/config.yaml
active_provider: gmail

providers:
  gmail:
    type: gmail
    client_id: your-client-id.apps.googleusercontent.com
    client_secret: your-secret
    token_file: ~/.pagen/tokens/gmail.json

  outlook:
    type: outlook
    client_id: your-app-id
    client_secret: your-secret
    token_file: ~/.pagen/tokens/outlook.json
```

### Token Storage

- Tokens stored in `~/.pagen/tokens/`
- One file per provider: `gmail.json`, `outlook.json`, etc.
- Token refresh handled transparently by OAuth2 client
- Consider encryption at rest for production

### CLI Commands for Provider Management

```bash
# Initial setup
pagen init                              # creates ~/.pagen/ directory structure

# Add a provider
pagen provider add gmail                # interactive OAuth setup

# Set active provider
pagen provider use gmail

# List providers
pagen provider list
# Output:
# * gmail (active) - authenticated
#   outlook - not authenticated

# Test provider connectivity
pagen provider test gmail               # verify auth + API access

# Re-authenticate
pagen provider reauth gmail             # refresh OAuth flow
```

## Error Handling

### Provider Error Flow

Errors flow back to agent with metadata for adaptation:

```go
type ToolResult struct {
    Success bool                    `json:"success"`
    Data    interface{}            `json:"data,omitempty"`
    Error   string                 `json:"error,omitempty"`
    Metadata map[string]interface{} `json:"metadata,omitempty"`
}
```

**Example: Rate limit error**
```json
{
    "success": false,
    "error": "Rate limit exceeded. Retry after 60 seconds.",
    "metadata": {
        "retry_after": 60,
        "rate_limit": "100/hour",
        "current_usage": 100
    }
}
```

### Agent Adaptation Strategies

When tools fail, agent can:
- **Retry with backoff**: Wait and retry based on `retry_after`
- **Alternative approach**: Search fewer emails, use different query
- **User notification**: Use TodoWrite to remind user to check quota
- **Graceful degradation**: Partial results if some operations fail

## Skill Development

### How Pagen Learns Workflows

**1. Primitives in System Prompt**
- All 19 tools documented with descriptions, parameters, examples
- Agent knows what operations are possible

**2. Conversation Memory**
- SQLite stores full conversation history
- Successful patterns are preserved
- Agent can reference past approaches

**3. Subagent Composition**
- Complex requests spawn focused subagents via Task tool
- Each subagent uses primitives to accomplish focused goal
- Main agent orchestrates and synthesizes results

**4. Emergent Learning**
- No explicit "skill saving" in v1
- Agent learns from repetition and memory
- Future: Persist successful workflows as reusable skills

### Example Workflow: Meeting Preparation

```
User: "Help me prepare for tomorrow's 2pm meeting with Alice"

Pagen main agent:
  1. [uses list_events(tomorrow)]
     → finds "Project Review with Alice, 2pm"

  2. [spawns subagent: "gather context for meeting with Alice"]
     Subagent uses:
       - search_emails(from="alice", after=last_week)
       - read_email() on most recent 3
       - search_emails(subject="project review")
       Returns: Summary of recent discussion points

  3. [spawns subagent: "create prep checklist"]
     Subagent uses:
       - create_task("Review Alice's latest feedback email")
       - create_task("Update project status doc")
       - create_task("Prepare Q4 timeline")
       Returns: 3 tasks created

  4. [main agent synthesizes]
     Response: "I've found 3 recent emails from Alice about the project.
               Created 3 prep tasks. Want me to draft talking points?"
```

### System Prompt Addition

```
You are Pagen, a personal productivity agent with access to:

Email Tools:
- send_email, reply_email, search_emails, read_email
- archive_email, mark_email_read/unread, label_email, delete_email

Calendar Tools:
- create_event, update_event, delete_event
- list_events, search_events, find_free_time

Task Tools:
- create_task, update_task, complete_task, list_tasks, delete_task

Orchestration:
- Task tool: Spawn focused subagents for complex workflows
- TodoWrite: Track multi-step work
- AskUserQuestion: Clarify ambiguous requests

Your goal: Help users manage their productivity workflows efficiently.

Best practices:
1. Use subagents to break complex requests into focused tasks
2. Learn from successful patterns in conversation history
3. When tools fail, adapt and try alternative approaches
4. Always explain what you're doing and why
```

## Implementation Phases

### Phase 1: Foundation (Week 1)

**Goal**: Provider infrastructure + Gmail read-only

- [ ] Fork Clem codebase, rename to Pagen
- [ ] Implement Provider interface and registry
- [ ] Remove code-focused tools (Read, Write, Edit, Bash, Grep, Glob)
- [ ] Keep: Task, AskUserQuestion, TodoWrite, WebFetch, WebSearch
- [ ] Implement Gmail provider scaffold
- [ ] OAuth flow with local callback server
- [ ] Read-only email tools: search_emails, read_email
- [ ] CLI commands: `pagen provider add/use/list/test`
- [ ] Basic error handling and logging

**Success Criteria**: Can authenticate with Gmail and search/read emails

### Phase 2: Core Email Operations (Week 2)

**Goal**: Full email management

- [ ] Implement remaining email tools:
  - send_email (with attachments)
  - reply_email
  - archive_email, mark_read/unread
  - label_email, delete_email
- [ ] Enhanced error handling (rate limits, quota)
- [ ] Tool approval flow integration
- [ ] Test coverage for email tools
- [ ] Documentation: Email tool reference

**Success Criteria**: Can perform complete email workflows (triage inbox, send replies, organize)

### Phase 3: Calendar & Tasks (Week 3)

**Goal**: Multi-domain productivity

- [ ] Implement calendar tools:
  - create_event, update_event, delete_event
  - list_events, search_events
  - find_free_time
- [ ] Implement task tools:
  - create_task, update_task, complete_task
  - list_tasks, delete_task
- [ ] Google Tasks API integration
- [ ] Multi-domain test scenarios
- [ ] Documentation: Calendar and task tool reference

**Success Criteria**: Can manage calendar events and tasks alongside email

### Phase 4: Subagent Workflows (Week 4)

**Goal**: Emergent skill composition

- [ ] Verify Task tool works with new productivity tools
- [ ] Create example workflows:
  - Meeting preparation
  - Weekly email cleanup
  - Daily briefing
- [ ] System prompt refinement
- [ ] Workflow documentation and examples
- [ ] User guide: How to use Pagen effectively

**Success Criteria**: Can demonstrate complex multi-tool workflows via subagents

### Phase 5: Polish & Distribution (Week 5)

**Goal**: Production-ready v1.0

- [ ] Second provider: Outlook/Microsoft 365
- [ ] Enhanced OAuth error handling
- [ ] Token refresh edge cases
- [ ] Performance optimization
- [ ] Security audit (token storage, API key handling)
- [ ] Release builds (same as Clem: Homebrew, binaries, Docker)
- [ ] Comprehensive documentation
- [ ] Release notes and changelog

**Success Criteria**: Ready to ship v1.0 with Gmail and Outlook support

## Future Enhancements (v2.0+)

### Multi-Provider Routing

**Goal**: Agent chooses best provider for each task

```go
type ProviderRouter struct {
    registry *ProviderRegistry
    strategy RoutingStrategy  // health-based, cost-based, user-preference
}

func (r *ProviderRouter) RouteToolCall(toolName string, params) (Provider, error) {
    candidates := r.findCapableProviders(toolName)
    return r.strategy.SelectProvider(candidates, params)
}
```

Agent can:
- Use Gmail for email, Outlook for calendar (if configured)
- Fail over to backup provider if primary is down
- Route based on rate limits (switch providers to avoid quota)

### Full go-plugin Architecture

**Goal**: Language-agnostic providers, process isolation

- Migrate to HashiCorp go-plugin (gRPC over stdio)
- Providers can be written in Python, Node.js, etc.
- Hot reload providers without restarting Pagen
- Better security isolation

### Skill Persistence

**Goal**: Save and reuse successful workflows

```bash
# Save a workflow
pagen skill save "weekly-cleanup" --last-conversation

# List saved skills
pagen skill list

# Run saved skill
pagen skill run weekly-cleanup

# Export/import skills
pagen skill export weekly-cleanup > skill.yaml
pagen skill import < skill.yaml
```

Skills are templatized subagent workflows:
```yaml
name: weekly-cleanup
description: Archive old emails and create action items
steps:
  - tool: search_emails
    params:
      is_unread: true
      older_than: 7days
  - for_each: email in results
    - tool: read_email
    - if: is_newsletter
      - tool: archive_email
    - elif: needs_action
      - tool: create_task
```

### Additional Providers

- **Notion**: Page creation, database queries
- **Linear**: Issue tracking, project management
- **Slack**: Messaging, channel management
- **Jira**: Issue tracking
- **GitHub**: Issues, PRs, notifications
- **Todoist**: Task management

### Enhanced Calendar Features

- **Smart scheduling**: Find optimal meeting times across attendees
- **Time blocking**: Automatically create focus time blocks
- **Meeting analysis**: Identify meeting overload, suggest consolidation

## Open Questions

1. **Rate limiting strategy**: Should Pagen implement its own rate limiting layer or rely on provider errors?
   - Leaning toward: Provider errors + metadata for transparency

2. **Attachment handling**: Local files vs cloud storage for email attachments?
   - Leaning toward: Local file paths (simpler), cloud upload as future enhancement

3. **Multi-calendar support**: Primary calendar only or support multiple?
   - v1: Primary calendar only
   - v2: Multi-calendar with routing

4. **Task provider**: Google Tasks, Todoist, or generic interface for both?
   - v1: Google Tasks (same OAuth as Gmail)
   - v2: Todoist provider option

5. **Offline mode**: Should Pagen cache data for offline operation?
   - v1: Online-only, fail gracefully
   - v2: Consider caching frequently accessed data

## Success Metrics

### Technical Metrics

- **Provider interface stability**: Can add new providers without changing core
- **Tool coverage**: All 19 primitives working across 2+ providers
- **Auth reliability**: OAuth flow succeeds >95% of time
- **Error recovery**: Agent successfully adapts to >80% of tool failures

### User Experience Metrics

- **Setup time**: Provider auth completes in <2 minutes
- **Response time**: Tool calls complete in <5 seconds (p95)
- **Workflow success**: Complex multi-tool workflows succeed >90% of time
- **User satisfaction**: Users report time savings on productivity tasks

### Learning Metrics

- **Pattern recognition**: Agent reuses successful approaches >70% of time
- **Subagent effectiveness**: Multi-step workflows decompose logically
- **Adaptation**: Agent recovers from errors without user intervention >60% of time

## Conclusion

Pagen represents a novel approach to AI-powered productivity: instead of hardcoding workflows, provide domain-specific tool primitives and let the agent compose them via subagents. The provider plugin architecture ensures extensibility while keeping the agent focused on orchestration.

**Next Steps**:
1. Review and approve this design document
2. Create Phase 1 implementation plan
3. Begin development: Fork Clem, implement provider infrastructure
4. Build Gmail provider as reference implementation

---

**Questions? Feedback?**
This design is ready for implementation, but open to refinement based on development discoveries.
