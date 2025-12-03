# Integration Tests for Hex Phase 2

This directory contains end-to-end integration tests for the Hex CLI Phase 2 implementation.

## Test Files

1. **helpers.go** - Common test utilities and helper functions
2. **storage_test.go** - Database persistence and conversation/message CRUD
3. **tools_test.go** - Tool registry, executor, and individual tool tests
4. **api_test.go** - Anthropic API client and streaming tests
5. **ui_test.go** - UI model state and lifecycle tests (simplified)
6. **scenarios_test.go** - End-to-end realistic user workflows

## Running Tests

```bash
# Run all integration tests
go test ./test/integration/...

# Run in short mode (skips slow/API tests)
go test -short ./test/integration/...

# Run with coverage
go test -cover ./test/integration/...

# Run specific test
go test -run TestStorageIntegration ./test/integration/...
```

## Test Requirements

### Storage Tests
- SQLite database created and initialized
- Conversations and messages persisted
- Foreign key constraints work
- Timestamps update correctly

### Tools Tests
- All three tools (Read, Write, Bash) registered
- Tools execute successfully
- Approval/denial flow works
- Concurrent execution is safe

### API Tests (Skipped in Short Mode)
- Require `ANTHROPIC_API_KEY` environment variable
- Test streaming responses
- Test token counting
- Test error handling

### UI Tests
- Model initialization
- State transitions
- Message handling
- Basic lifecycle

### Scenario Tests
- Complete workflows combining all components
- Error recovery
- Resume conversation
- Multi-tool sequences

## Notes

- Tests use temporary databases (`t.TempDir()`)
- All tests clean up after themselves
- API tests are skipped in `-short` mode
- Some tests check for specific error messages (may be brittle)
