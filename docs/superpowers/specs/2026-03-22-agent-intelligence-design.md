# Agent Intelligence Improvements

Based on findings from "From Code Foundation Models to Agents and Applications" (arXiv:2511.18538) and companion papers on code agent success/failure trajectories.

## Problem

Hex's agent loop is functional but lacks the intelligence layers that research shows make code agents effective. The system prompt is one sentence. There is no self-correction, no stuck detection, no planning phase, no project memory, and context pruning is primitive.

## Changes

### 1. Enhanced System Prompt

**File:** `internal/core/constants.go`

Replace the one-line identity prompt with comprehensive guidance covering:

- Tool usage patterns (read before edit, verify after changes)
- Self-correction protocol (analyze errors before retrying, don't repeat failed approaches)
- Clarification protocol (ask when ambiguous rather than guessing)
- Code quality expectations (match existing style, minimal changes)
- Planning guidance (outline approach for complex multi-step tasks)

The prompt stays in `DefaultSystemPrompt` as a single const string. Target length: 40-50 lines of focused guidance. No bloat.

**Research basis:** Claude Code configuration study (arXiv:2511.09268) found architecture specs in 100% of top-performing agent configs. Development guidelines appear in 44.8%, testing specs in 35.4%.

### 2. Stuck Detection & Adaptive Turn Management

**Files:** `cmd/hex/print.go`, `cmd/hex/root.go`

Track tool execution patterns across turns in the legacy print loop:

- Count consecutive failures of the same tool
- After 2+ consecutive failures of the same tool, inject a stuck hint into the next user message: "You appear to be stuck. The same tool has failed multiple times. Try a different approach."
- Add `--max-turns` flag (default 20, currently hardcoded)

Data structures:
```go
type turnTracker struct {
    lastFailedTool string
    failCount      int
}
```

After each turn, check tool results for errors. If the same tool failed again, increment. If a different tool was used or succeeded, reset. When failCount >= 2, append hint text to the tool results message.

The mux path benefits from the improved system prompt but does not get the stuck detection logic (mux manages its own loop).

**Research basis:** Paper on success/failure trajectories (arXiv:2511.00197) found failed trajectories are consistently longer with higher variance. Agents lack mechanisms to detect and abandon dead ends.

### 3. Verification Hints After Mutations

**File:** `cmd/hex/print.go`

After a tool execution turn where `edit` or `write_file` was used, append a brief verification nudge to the tool results:

```
"[hex: Files were modified. Consider verifying your changes compile or pass tests before proceeding.]"
```

This is appended as additional text in one of the tool_result content blocks, not a separate message. The LLM decides whether to act on it.

Detection logic: scan the `toolUses` slice for tools named `"edit"` or `"write_file"`. If any are present, append the hint to the last tool result block.

**Research basis:** Paper Section 5.1.2 states "1-3 refinement rounds yield the largest performance gains." Feedback is "the engine of effective code search."

### 4. Smarter Context Pruning

**File:** `internal/convcontext/manager.go`

Three improvements:

**a) ContentBlock-aware token estimation:**
Current `EstimateMessageTokens` only counts `msg.Content` (string field). It ignores `msg.ContentBlock` (the array of structured blocks used for tool calls/results). Add estimation for content blocks.

**b) Summary-based pruning:**
When removing old messages during pruning, replace tool_result messages with 1-line summaries instead of deleting entirely. Format: `"[Previously: tool_name executed successfully]"` or `"[Previously: read_file on path/to/file.go]"`.

**c) Error prioritization:**
When deciding what to keep during pruning, weight messages containing error strings higher than successful tool results. Errors carry more diagnostic value for the agent's ongoing work.

### 5. Project Memory

**New files:** `internal/memory/project.go`, `internal/memory/project_test.go`

Lightweight project context scanner that:

1. On first run (or when `.hex/project.json` doesn't exist), scans the working directory for signals:
   - Language: look for `go.mod`, `package.json`, `pyproject.toml`, `Cargo.toml`, etc.
   - Build system: `Makefile`, `build.gradle`, `CMakeLists.txt`, etc.
   - Test framework: infer from language + config files
   - Project structure: list top-level directories
   - Entry points: `main.go`, `index.ts`, `app.py`, etc.

2. Saves results to `.hex/project.json`:
```json
{
  "language": "go",
  "build_command": "make build",
  "test_command": "go test ./...",
  "structure": ["cmd/", "internal/", "docs/"],
  "detected_at": "2026-03-22T12:00:00Z"
}
```

3. On subsequent runs, loads and injects a brief context block into the system prompt:
```
Project context: Go project. Build: make build. Test: go test ./...
Key directories: cmd/, internal/, docs/
```

4. Respects staleness: re-scan if `project.json` is older than 7 days or if `--refresh-memory` flag is set.

Integration points:
- `cmd/hex/print.go`: load project memory and append to system prompt
- `cmd/hex/mux_runner.go`: same
- `cmd/hex/root.go`: add `--refresh-memory` flag

**Research basis:** Paper Section 5.4 future trends: "agents will develop a dynamic mental model of a repository." Section 6.2.3 discusses code-based memory with skill libraries for cross-session persistence.

### 6. Plan-then-Execute Mode

**Files:** `cmd/hex/print.go`, `cmd/hex/root.go`

Add `--plan` flag that modifies the agent loop:

1. **Planning turn:** Prepend to the user's prompt:
   ```
   Before executing, create a numbered plan for this task. List what files you need to read, what changes to make, and how to verify. Output ONLY the plan, do not execute yet.
   ```

2. **Execution turns:** After receiving the plan, inject it as context and add:
   ```
   Now execute the plan above step by step. After each step, note which step you completed.
   ```

3. The regular tool execution loop then proceeds as normal, but with the plan as reference context.

This is opt-in only. Default behavior is unchanged.

**Research basis:** Paper Section 6.1.1 describes ReWOO's plan-then-execute outperforming ReAct. Planner-centric paper (arXiv:2511.10037) shows 59.8% success vs ReAct's 48.2%.

### 7. Built-in "agent" Spell

**New files:** `internal/spells/builtin/agent/system.md`, `internal/spells/builtin/agent/config.yaml`

A spell that activates maximum agent intelligence. Uses `layer` mode to augment the default prompt:

`config.yaml`:
```yaml
mode: layer
reasoning:
  effort: high
```

`system.md` contents emphasize:
- Always plan before executing
- Verify every change (run tests, check compilation)
- When stuck, step back and reconsider the approach
- Read existing code thoroughly before modifying
- Explain reasoning for non-obvious decisions

This gives users `--spell agent` as a one-flag way to activate careful, methodical behavior.

## Files Changed

| File | Change Type | Description |
|------|------------|-------------|
| `internal/core/constants.go` | Modified | Enhanced system prompt |
| `cmd/hex/print.go` | Modified | Stuck detection, verification hints, plan mode, project memory integration |
| `cmd/hex/mux_runner.go` | Modified | Project memory integration |
| `cmd/hex/root.go` | Modified | New flags: --max-turns, --plan, --refresh-memory |
| `internal/convcontext/manager.go` | Modified | Smarter pruning with summaries and error prioritization |
| `internal/memory/project.go` | New | Project memory scanner and loader |
| `internal/memory/project_test.go` | New | Tests for project memory |
| `internal/spells/builtin/agent/system.md` | New | Agent spell system prompt |
| `internal/spells/builtin/agent/config.yaml` | New | Agent spell config |

## Testing Strategy

- **Unit tests** for project memory detection logic (mock filesystem with different project types)
- **Unit tests** for improved context pruning (ContentBlock estimation, summary generation, error prioritization)
- **Unit tests** for stuck detection tracker
- **Integration tests** via `.scratch/` scenarios:
  - Test that enhanced system prompt produces better tool usage patterns
  - Test that `--plan` flag produces a plan before executing
  - Test that stuck detection fires after repeated failures
  - Test that project memory is created and loaded correctly
- **End-to-end**: Use hex to develop hex (the existing dogfooding workflow validates everything)

## Non-Goals

- Changes to the mux framework (separate repo)
- RL/fine-tuning the foundation model
- Multi-agent collaboration changes (TaskTool already exists)
- New tool types
- Changes to the TUI/interactive mode agent loop (focus on print mode)
