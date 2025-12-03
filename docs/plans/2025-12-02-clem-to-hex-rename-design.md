# Full "hex" → "hex" Rename Design

**Date**: 2025-12-02
**Status**: Approved
**Context**: Early development, no external users, clean break rename

## Scope

This is a comprehensive codebase rename touching 8 categories:

1. **Go module path**: `github.com/2389-research/hex` → `github.com/2389-research/hex`
2. **All import paths**: Every file importing internal packages
3. **Binary/command name**: `cmd/hex/` → `cmd/hex/`
4. **Config directory references**: `~/.hex/` → `~/.hex/`
5. **Documentation**: README, docs/, CHANGELOG, code comments
6. **Default paths**: Database paths, log file references
7. **GitHub Actions**: Workflow files, badge URLs
8. **User-facing strings**: CLI help text, error messages, prompts

## Approach

**Automated + Manual Hybrid**:
- Use `gofmt -r` and `find/sed` for mechanical replacements
- Manual review for context-sensitive changes
- Testing checkpoint after each phase
- Single atomic commit for easy rollback

## Execution Plan

### Phase 1: Preparation
1. Ensure clean git state
2. Run tests to establish baseline
3. Create backup branch: `git branch backup-before-hex-rename`

### Phase 2: Core Go Changes
4. Update `go.mod`: Change module path to `github.com/2389-research/hex`
5. Update all Go imports:
   - `github.com/2389-research/hex/internal/` → `github.com/2389-research/hex/internal/`
   - `github.com/2389-research/hex/pkg/` → `github.com/2389-research/hex/pkg/`
6. Rename directory: `cmd/hex/` → `cmd/hex/`
7. Run `go mod tidy`
8. **Checkpoint**: `go build ./...` and `go test ./...`

### Phase 3: Configuration & Paths
9. Update all `~/.hex` references to `~/.hex`
10. Update database path defaults (`defaultDBPath()`)
11. Update log file paths and config references
12. **Checkpoint**: Build and run `./hex --help`

### Phase 4: Documentation & User-Facing
13. Update README.md (title, badges, all references)
14. Update CHANGELOG.md
15. Update all `docs/` files
16. Update CLI help text and error messages
17. Update code comments mentioning "hex"

### Phase 5: GitHub Actions & CI
18. Update `.github/workflows/*.yml`
19. Update GitHub badge URLs
20. Update release workflow artifact names
21. **Checkpoint**: Push to branch, verify CI passes

### Phase 6: Comprehensive Testing
22. Run full test suite: `go test ./...`
23. Build binary: `go build -o hex ./cmd/hex`
24. Manual smoke tests:
    - `./hex --help`
    - `./hex --version`
    - Interactive mode (verify UI says "hex")
    - Database creation (verify `~/.hex/` directory)
25. Search for missed references: `rg -i "hex"`
26. Check docs: `rg -i "hex" README.md docs/`

### Phase 7: Finalization
27. Commit all changes
28. Tag the rename commit
29. **GitHub repo rename**: Settings → Repository name → "hex"
30. Update local remote: `git remote set-url origin git@github.com:harper/hex.git`
31. Push and verify CI passes

## Edge Cases

- Template strings with "Hex" in user messages (update to "Hex")
- External MCP servers referencing binary name (document breaking change)
- Examples directory with hardcoded paths (update all)

## Rollback Plan

If issues arise:
1. `git reset --hard backup-before-hex-rename`
2. Rename GitHub repo back to "hex"
3. Update remote URL back

## Success Criteria

- All tests pass
- Binary builds as `hex`
- CLI says "hex" not "hex"
- Config directory is `~/.hex/`
- No "hex" references in code (except git history)
- CI passes on GitHub
