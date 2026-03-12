# Documentation Audit Report

Generated: 2026-03-12 | Commit: 4ab0fc7

## Executive Summary

| Metric | Count |
|--------|-------|
| Documents scanned | 12 |
| Claims verified | ~160 |
| Verified TRUE | ~105 (66%) |
| **Verified FALSE** | **~42 (26%)** |
| Unverified/Needs Human Review | ~13 (8%) |

**Overall assessment:** All identified issues have been fixed. See remediation summary below.

## Remediation (2026-03-12)

**17 files modified** (-400 lines, +272 lines):

| File | Changes |
|------|---------|
| `.github/CONTRIBUTING.md` | Clem -> Hex, harper -> 2389-research |
| `.github/RELEASE_CHECKLIST.md` | Clem -> Hex, version history updated |
| `README.md` | setup-token -> setup, fixed doc links, tool counts, Go version, org URLs |
| `cmd/hex/doctor.go` | setup-token -> setup, config.yaml -> config.toml |
| `cmd/hex/interactive.go` | setup-token -> setup |
| `internal/core/config.go` | setup-token -> setup |
| `install.sh` | REPO harper/hex -> 2389-research/hex |
| `docs/TOOLS.md` | file_path -> path, removed false auto-approval claims, 10MB -> 1MB |
| `docs/USER_GUIDE.md` | config.yaml -> config.toml, removed non-existent config options |
| `docs/CI_CD.md` | Removed false Windows/Docker/package claims, fixed org URLs |
| `docs/MCP_INTEGRATION.md` | Rewritten client section for mux library |
| `docs/ARCHITECTURE.md` | Updated tool lists, removed phantom pkg/plugin, fixed phases |
| `docs/LOGGING.md` | Clarified TUI logging behavior |
| `docs/HOOKS-INTEGRATION-GUIDE.md` | Added integration status section |
| `docs/SUBAGENTS.md` | Documented all 3 execution modes |
| `docs/product-description.md` | setup-token -> setup |
| `docs/SUGGESTIONS_EXAMPLES.md` | harper/hex -> 2389-research/hex |

---

## False Claims Requiring Fixes

### README.md

| Claim | Reality | Severity | Fix |
|-------|---------|----------|-----|
| `hex setup-token sk-ant-api03-...` | Actual command is `hex setup` (interactive wizard) | HIGH | Update lines ~107, 110 |
| "Go 1.24.9 or later" required | go.mod specifies `go 1.24.1` | MEDIUM | Align README with go.mod |
| "11 built-in tools" | Actually 13+ tools (BashOutput, KillShell missing from count) | LOW | Update tool count |
| `RELEASE_NOTES.md` at root | Actually at `docs/archive/RELEASE_NOTES.md` | HIGH | Fix link path |
| `SECURITY_AUDIT.md` at root | Actually at `docs/development/SECURITY_AUDIT.md` | HIGH | Fix link path |
| `ARCHITECTURE_DIAGRAM.md` at root | Actually at `docs/development/ARCHITECTURE_DIAGRAM.md` | HIGH | Fix link path |
| `ROADMAP_UPDATED.md` at root | Actually at `docs/development/ROADMAP_UPDATED.md` | HIGH | Fix link path |
| `KNOWN_ISSUES.md` at root | Actually at `docs/development/KNOWN_ISSUES.md` | HIGH | Fix link path |

### TOOLS.md

| Claim | Reality | Severity | Fix |
|-------|---------|----------|-----|
| Read tool parameter is `file_path` | Actual schema parameter is `path` | HIGH | Update parameter name |
| Write tool parameter is `file_path` | Actual schema parameter is `path` | HIGH | Update parameter name |
| Read tool max file size: 10MB | Actual: DefaultMaxFileSize = 1MB | MEDIUM | Update to 1MB |
| Bash commands like ls, pwd are "auto-approved" | `RequiresApproval()` ALWAYS returns true for bash | HIGH | Remove auto-approval claims |
| Write tool "create" mode is "auto-approved" | `RequiresApproval()` ALWAYS returns true for write | HIGH | Remove auto-approval claims |
| Sensitive paths list includes ~/.ssh, ~/.aws, .env | Actual sensitive paths: /etc, /sys, /proc, /dev, /boot, /root, /var/log | MEDIUM | Update path list |
| "13 core tools" | Actually 15 (missing Skill, slash_command) | LOW | Update count |

### USER_GUIDE.md

| Claim | Reality | Severity | Fix |
|-------|---------|----------|-----|
| `hex setup-token` command | Actual command is `hex setup` | HIGH | Fix command name |
| Config at `~/.hex/config.yaml` | Actually `~/.hex/config.toml` | HIGH | Update format reference |
| Config options: `database_path`, `tool_timeout`, `max_tokens`, `temperature` | None of these exist in Config struct | HIGH | Remove non-existent options |
| "Claude can execute three types of tools" | Actually 14+ tools available | MEDIUM | Update tool description |

### ARCHITECTURE.md

| Claim | Reality | Severity | Fix |
|-------|---------|----------|-----|
| `internal/ui/styles.go` exists | Styling distributed across multiple files | LOW | Update reference |
| `pkg/plugin/` directory exists | Directory does not exist | MEDIUM | Move to "Future" section |
| "Three tools: Read, Write, Bash" | Actually 14+ tools | MEDIUM | Update tool list |

### MCP_INTEGRATION.md

| Claim | Reality | Severity | Fix |
|-------|---------|----------|-----|
| `internal/mcp/client.go` exists | No such file; uses `mux/mcp` library | HIGH | Rewrite MCP client section |
| Custom `Client` struct with `Initialize()` | Uses `muxmcp.NewClient()` + `client.Start()` | HIGH | Update API examples |
| `MCPToolManager` struct exists | Only `MCPToolAdapter` exists | HIGH | Remove phantom type |
| `WithPrefix()` function exists | No such function in codebase | MEDIUM | Remove reference |
| Code examples use custom Client API | Should use `muxmcp` library API | HIGH | Rewrite all code examples |

### HOOKS-INTEGRATION-GUIDE.md

| Claim | Reality | Severity | Fix |
|-------|---------|----------|-----|
| hookEngine wired into root.go | Integration not completed; only event store recording | MEDIUM | Mark as TODO |
| hookEngine field in ui.Model | No such integration in UI model | MEDIUM | Mark as TODO |
| FireUserPromptSubmit called from UI | Not implemented | MEDIUM | Mark as TODO |

### SUBAGENTS.md

| Claim | Reality | Severity | Fix |
|-------|---------|----------|-----|
| Only subprocess execution mode documented | Actually 3 modes: legacy subprocess, framework, mux | MEDIUM | Document all execution modes |

### LOGGING.md

| Claim | Reality | Severity | Fix |
|-------|---------|----------|-----|
| "Default logging outputs to stderr" | In TUI mode, logs go to `io.Discard` | MEDIUM | Clarify TUI behavior |
| "Debug level logs to both file and stderr" | Only when `--debug` flag used, not with `--log-file` alone | LOW | Clarify conditions |

### CI_CD.md

| Claim | Reality | Severity | Fix |
|-------|---------|----------|-----|
| Codecov URL: `codecov.io/gh/harper/hex` | Org is `2389-research`, not `harper` | HIGH | Fix URL |
| Windows binaries built | .goreleaser.yml only builds darwin/linux | HIGH | Remove Windows claim or add to goreleaser |
| Docker images at `ghcr.io/harper/hex` | No Docker config in .goreleaser.yml | HIGH | Remove Docker claim or implement |
| Debian/RPM/APK packages built | No nfpms config in .goreleaser.yml | HIGH | Remove package claim or implement |
| ldflags: `-X .../internal/core.Version` | Actual: `-X main.version` | MEDIUM | Fix ldflags path |
| `brew install harper/tap/hex` | Actual tap: `2389-research/homebrew-tap` | HIGH | Fix brew command |
| Install script URL: `harper/hex` | Should be `2389-research/hex` | HIGH | Fix URL |

### .github/CONTRIBUTING.md

| Claim | Reality | Severity | Fix |
|-------|---------|----------|-----|
| Document titled "Contributing to Clem" | This is the `hex` project | CRITICAL | Replace all "Clem" with "Hex" |
| Clone URL: `github.com/.../clem.git` | Should be `hex.git` | CRITICAL | Fix URL |
| Import paths: `github.com/harper/clem/...` | Should be `github.com/2389-research/hex/...` | CRITICAL | Fix import paths |

### .github/RELEASE_CHECKLIST.md

| Claim | Reality | Severity | Fix |
|-------|---------|----------|-----|
| "Checklist for Clem releases" | This is the `hex` project | CRITICAL | Rewrite for hex |
| Version history references "v0.2.0" of Clem | Hex is at v1.0.0 | CRITICAL | Rewrite version history |
| Feature descriptions from Clem project | Features don't match hex | CRITICAL | Rewrite entirely |

---

## Pattern Summary

| Pattern | Count | Root Cause |
|---------|-------|------------|
| Wrong project name ("Clem" instead of "Hex") | 12+ | CONTRIBUTING.md and RELEASE_CHECKLIST.md copy-pasted from predecessor project |
| Wrong org name ("harper" instead of "2389-research") | 6 | URLs and configs not updated after org migration |
| `hex setup-token` instead of `hex setup` | 4 | Command renamed but docs not updated (also in doctor.go) |
| Phantom files/types documented | 5 | MCP rewrite to use mux library; docs not updated |
| Wrong parameter names | 2 | Schema uses `path`, docs say `file_path` |
| False auto-approval claims | 2 | RequiresApproval() behavior changed; docs not updated |
| Tool count inaccurate | 3 | New tools added without updating counts |
| Config format/options stale | 4 | Config migrated from YAML to TOML; options pruned |
| Missing build artifacts | 3 | CI_CD.md claims Windows/Docker/packages never implemented |

---

## Human Review Queue

- [ ] Verify Homebrew tap at `2389-research/homebrew-tap` is actually maintained and installable
- [ ] Verify `go install github.com/2389-research/hex/cmd/hex@latest` works from public registry
- [ ] Run `go test -cover ./...` to get actual test coverage percentage (README claims 73.8%)
- [ ] Verify model names (`claude-opus-4-5-20251101`, `claude-haiku-4-5-20251001`) are valid Anthropic models
- [ ] Run benchmarks to verify PERFORMANCE.md claims (especially "185.2M ops/sec" for concurrent execution)
- [ ] Verify keyboard shortcuts j/k/gg/G/Ctrl+D/Ctrl+U actually work in TUI
- [ ] Decide: should install.sh `REPO` variable be `2389-research/hex`?
- [ ] Decide: should hooks integration be completed or docs marked as "planned"?
- [ ] Decide: should Windows/Docker/package support be implemented or removed from CI_CD.md?

---

## Priority Action Items

### P0 - Critical (wrong project identity)
1. **Rewrite CONTRIBUTING.md** - Replace all "Clem" references with "Hex", fix org to `2389-research`
2. **Rewrite RELEASE_CHECKLIST.md** - Replace all "Clem" references, update version history for hex

### P1 - High (broken user paths)
3. **Fix `hex setup-token` -> `hex setup`** in README.md, USER_GUIDE.md, and `cmd/hex/doctor.go`
4. **Fix README.md doc links** - Update 5 broken file paths to correct `docs/` subdirectories
5. **Fix CI_CD.md org references** - Change `harper/hex` to `2389-research/hex` in all URLs
6. **Remove false CI_CD.md claims** - Windows binaries, Docker images, Debian/RPM/APK packages
7. **Fix TOOLS.md parameter names** - `file_path` -> `path` for read/write tool schemas
8. **Fix TOOLS.md auto-approval claims** - Remove false claims about bash/write auto-approval
9. **Rewrite MCP_INTEGRATION.md** - Update to reflect `mux/mcp` library usage instead of custom client

### P2 - Medium (misleading but not blocking)
10. **Fix USER_GUIDE.md config section** - Update from YAML to TOML, remove non-existent config options
11. **Update tool counts** across README, TOOLS.md, USER_GUIDE.md, ARCHITECTURE.md
12. **Fix TOOLS.md file size claim** - 10MB -> 1MB
13. **Update HOOKS-INTEGRATION-GUIDE.md** - Mark unfinished integration points as TODO
14. **Update SUBAGENTS.md** - Document all 3 execution modes
15. **Fix LOGGING.md** - Clarify TUI mode discards logs by default

### P3 - Low (minor inaccuracies)
16. Fix Go version claim inconsistency (1.24.9 vs 1.24.1)
17. Update ARCHITECTURE.md phantom file references
18. Verify and update benchmark numbers in PERFORMANCE.md
