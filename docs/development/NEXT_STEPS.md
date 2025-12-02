# Clem - Next Steps for Phase 2

## Current Status

✅ **Phase 1 Complete:** Foundation working
- CLI framework (Cobra)
- Configuration system (Viper)
- API client with message creation
- Print mode functional
- Setup and doctor commands
- All tests passing (15+ tests)

## Phase 2: Interactive Mode

**Goal:** Full-featured TUI with streaming, storage, and history

**Reference:** See `docs/plans/2025-11-26-clem-phase2-interactive.md` for detailed implementation plan

### Tasks Overview

#### Task 1: SQLite Storage Schema
**Estimated:** 4-6 hours

Files to create:
- `internal/storage/schema.go`
- `internal/storage/schema_test.go`
- `internal/storage/migrations/001_initial.sql`

What to build:
- Database schema (conversations, messages tables)
- Migration system
- CRUD operations for conversations
- CRUD operations for messages
- Tests for all operations

#### Task 2: Streaming API Client
**Estimated:** 4-6 hours

Files to create:
- `internal/core/streaming.go`
- `internal/core/streaming_test.go`

What to build:
- SSE (Server-Sent Events) client
- Progressive message streaming
- Chunk parsing and accumulation
- Error handling for stream interruptions
- Tests with mock streams

#### Task 3: Bubbletea TUI
**Estimated:** 8-12 hours

Files to create:
- `internal/ui/model.go` - App state model
- `internal/ui/update.go` - Update function
- `internal/ui/view.go` - View rendering
- `internal/ui/components/` - Reusable components
  - `header.go` - Status bar
  - `messages.go` - Message list
  - `input.go` - Prompt input
  - `spinner.go` - Loading indicator

What to build:
- Main TUI loop (Bubbletea)
- Message rendering with syntax highlighting
- Input handling
- Status bar with model/token count
- Streaming response visualization
- Keyboard shortcuts (Ctrl+C to exit, etc.)

#### Task 4: Conversation History
**Estimated:** 4-6 hours

Files to create:
- `internal/storage/history.go`
- `internal/storage/history_test.go`
- Update `cmd/clem/root.go` for --continue and --resume

What to build:
- Save conversations to SQLite
- Load most recent conversation (--continue)
- List and select conversations (--resume)
- Session metadata tracking
- Auto-save on exit

#### Task 5: Background Tasks
**Estimated:** 2-4 hours

Files to create:
- `internal/ui/tasks.go`
- `internal/ui/tasks_test.go`

What to build:
- Task queue management
- Progress tracking
- Status updates in TUI
- Concurrent task execution

### How to Execute

**Option 1: Use Superpowers Skill (Recommended)**

```bash
# From project root
claude /superpowers:execute-plan docs/plans/2025-11-26-clem-phase2-interactive.md
```

This will:
- Load the detailed plan
- Execute tasks in batches
- Review after each batch
- Ensure test coverage

**Option 2: Manual Implementation**

Follow TDD for each task:
1. Write the test first (see examples in plan)
2. Run test to see it fail
3. Write minimal code to pass
4. Refactor while keeping tests green
5. Move to next test

### Testing Strategy

**Unit Tests:**
- Every new file gets a `_test.go` companion
- Use table-driven tests for multiple cases
- Mock external dependencies (SQLite uses in-memory DB)

**Integration Tests:**
- Create `tests/integration/phase2_test.go`
- Test full workflows:
  - Start conversation → send message → save → resume
  - Stream response → render in TUI
  - Background tasks → status updates

**No Mocks Policy:**
- Use real SQLite (in-memory for tests)
- Use real HTTP clients (with VCR for API calls)
- Use real file system (temp directories)

### Success Criteria

Phase 2 is complete when:
- [ ] `clem` (no flags) launches interactive TUI
- [ ] Can send messages and see streaming responses
- [ ] Conversations save to SQLite automatically
- [ ] `clem --continue` resumes last conversation
- [ ] `clem --resume` shows conversation picker
- [ ] Status bar shows model, tokens, task status
- [ ] All tests pass (unit + integration)
- [ ] Documentation updated

### Estimated Timeline

| Task | Estimated Hours | Dependencies |
|------|----------------|--------------|
| 1. SQLite Schema | 4-6h | None |
| 2. Streaming Client | 4-6h | None |
| 3. Bubbletea TUI | 8-12h | Task 2 (streaming) |
| 4. History | 4-6h | Task 1 (schema) |
| 5. Background Tasks | 2-4h | Task 3 (TUI) |
| **Total** | **22-34h** | **~1 week** |

### Tech Stack

**UI:**
- [Bubbletea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Styling
- [Glamour](https://github.com/charmbracelet/glamour) - Markdown rendering

**Storage:**
- [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) - Pure Go SQLite
- Hybrid schema (normalized tables + JSON for complex data)

**Streaming:**
- Standard library `net/http` for SSE
- `bufio.Scanner` for line-by-line parsing
- Channels for async updates

### Debugging Tips

**TUI Issues:**
- Use `--debug` flag to write logs to file
- Bubbletea has `tea.LogToFile()` for debugging
- Test view rendering with snapshot tests

**Storage Issues:**
- Use SQLite CLI to inspect database
- Enable SQL query logging in tests
- Check for locking issues (use WAL mode)

**Streaming Issues:**
- Log raw SSE events before parsing
- Test with recorded cassettes
- Handle reconnection gracefully

### Resources

**Documentation:**
- Bubbletea tutorial: https://github.com/charmbracelet/bubbletea/tree/master/tutorials
- SQLite Go driver: https://pkg.go.dev/modernc.org/sqlite
- SSE spec: https://html.spec.whatwg.org/multipage/server-sent-events.html

**Examples:**
- Bubbletea examples: https://github.com/charmbracelet/bubbletea/tree/master/examples
- Glamour demo: https://github.com/charmbracelet/glamour

**Design Docs:**
- Full Phase 2 plan: `docs/plans/2025-11-26-clem-phase2-interactive.md`
- Architecture overview: `README.md`
- Phase 1 report: `docs/PHASE1.md`

### Getting Help

If stuck:
1. Check the detailed plan in `docs/plans/`
2. Review Bubbletea examples for TUI patterns
3. Look at Phase 1 code for API client patterns
4. Ask Claude for specific implementation guidance

### Pre-flight Checklist

Before starting Phase 2:
- [ ] All Phase 1 tests passing
- [ ] Go 1.21+ installed
- [ ] API key configured (for testing)
- [ ] Read Phase 2 plan thoroughly
- [ ] Understand TDD workflow
- [ ] Have Bubbletea docs handy

---

## Quick Start

```bash
# Install dependencies
cd clean
go get github.com/charmbracelet/bubbletea
go get github.com/charmbracelet/lipgloss
go get github.com/charmbracelet/glamour
go get modernc.org/sqlite

# Verify Phase 1 still works
make test
./clem --version

# Start Phase 2 Task 1 (SQLite Schema)
# Create test first:
mkdir -p internal/storage/migrations
touch internal/storage/schema.go
touch internal/storage/schema_test.go

# Follow TDD cycle for each task
# See detailed plan for exact code to write
```

---

**Ready to begin Phase 2!** 🚀
