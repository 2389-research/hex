# Hex Modernization Spec: Mux-Native Code Agent + Claude/Codex-Style Skills

**Proposed filename:** `docs/plans/2026-06-20-hex-modernization-spec.md`
**Status:** Draft
**Scope:** `2389-research/hex` with targeted supporting changes in `2389-research/mux`
**Primary outcome:** Make Hex the robust local code-agent product and make mux the reusable orchestration kernel.

---

## 1. Executive Summary

Hex already has most of the pieces needed for a robust code agent: CLI/TUI, conversation persistence, tool execution, approvals, MCP, subagents, background shell processes, structured logging, and a partially wired mux execution path.

Mux provides the right reusable core: agent orchestration, typed tool registry/executor, approval callbacks, event streaming, hooks, child agents, LLM provider abstraction, and MCP tool adaptation.

The modernization should not build a new product beside Hex. It should make Hex's mux path the production path and use Hex as the reference implementation of a robust local coding agent.

The modernization has two tightly connected tracks:

1. **Code-agent modernization** — mux-native execution, durable runs, workspace isolation, audit logs, patch-first editing, richer policy, approval broker, git/test tools, final reports, and safer subagent orchestration.
2. **Skill system modernization** — upgrade Hex's existing Markdown skill loader into a Claude/Codex-style skill runtime with `SKILL.md` directory packages, progressive disclosure, explicit slash invocation, automatic activation, skill-scoped tool policy, and trust controls.

The product line should become:

```text
hex = robust local code agent, CLI/TUI/runtime/product shell
mux = reusable Go agent orchestration framework
```

---

## 2. Current Audit: What Hex Already Has

### 2.1 Product/runtime shell

Hex is already positioned as a full Claude-like CLI with print mode, interactive TUI, SQLite persistence, tool support, MCP integration, structured logging, and distribution paths.

Current strengths:

- Print mode for one-off queries.
- Interactive Bubbletea TUI.
- Conversation persistence in SQLite.
- Built-in code tools.
- MCP integration.
- Structured logging.
- Background process support.
- Subagent support through `Task`.
- Existing skill package skeleton.

### 2.2 Current project structure

Observed structure from `docs/ARCHITECTURE.md`:

```text
hex/
├── cmd/hex/              # CLI entry point and commands
├── internal/
│   ├── core/             # API client, types, config
│   ├── ui/               # Bubbletea TUI
│   ├── storage/          # SQLite persistence
│   └── tools/            # Tool execution system
├── test/integration/
├── docs/
├── go.mod
└── Makefile
```

This is the correct split. The modernization should add `internal/runtime`, expand `internal/skills`, and improve `internal/tools`, rather than flattening or replacing the architecture.

### 2.3 Current mux integration

Hex already has `cmd/hex/mux_runner.go`.

The current mux runner:

1. Loads Hex config.
2. Determines provider and model.
3. Creates a mux LLM client.
4. Wraps thinking if enabled.
5. Builds Hex tools for mux.
6. Builds system prompt from Hex defaults.
7. Loads `AGENTS.md` context.
8. Loads project memory context.
9. Creates a mux approval function.
10. Creates hook engine and mux hook bridge.
11. Creates either a root mux agent or mux subagent.
12. Subscribes to mux events.
13. Runs the mux agent and streams output.

This is the right foundation. The modernization should make this path default and production-grade.

### 2.4 Existing adapter layer

Hex has `internal/adapter/tool.go`, which adapts Hex's `tools.Tool` interface into mux's `tool.Tool` interface.

Current behavior:

- `Name()` maps directly.
- `Description()` maps directly.
- `RequiresApproval()` converts `map[string]any` to `map[string]interface{}` and delegates to the Hex tool.
- `Execute()` converts params, calls the Hex tool, then converts Hex `Result` to mux `Result`.
- `InputSchema()` calls `tools.GetToolSchema()` so mux can expose schemas to the model.

This is good as a compatibility bridge. Longer term, mux's public `tool.Tool` should become the canonical interface, but the adapter should stay during migration.

### 2.5 Existing tools

Hex currently exposes the core code-agent tool surface:

```text
read_file
write_file
edit
bash
grep
glob
task
bash_output
kill_shell
ask_user_question
todo_write
web_fetch
web_search
Skill
```

Not all of these are wired into the mux runner today. The mux runner currently builds a smaller set: read/write/edit/bash/grep/glob/task.

### 2.6 Existing storage

Hex has SQLite storage for conversations, messages, todos, and history. The migration initializes:

- `conversations`
- `messages`
- `todos`
- `history`
- `history_fts`
- relevant indexes

The storage layer enables foreign keys and WAL mode.

This is good for conversation persistence, but it is not enough for robust code-agent runs. We need run/event/tool/artifact persistence.

### 2.7 Existing skills system

Hex already has a `skills` package:

```text
internal/skills/
├── loader.go
├── registry.go
├── skill.go
├── tool.go
└── tool_adapter.go
```

Current capabilities:

- Loads `.md` skills from builtin, user, plugin, and project directories.
- Project skills can override earlier sources.
- Parses YAML frontmatter and Markdown content.
- Supports `name`, `description`, `tags`, `activationPatterns`, `model`, `priority`, `dependencies`, and `version`.
- Compiles activation regexes.
- Registry supports lookup, list, tag search, pattern matching, text search, count, and clear.
- `Skill` tool loads named skill content into the conversation.
- `ToolAdapter` exposes the skill tool as a normal Hex `tools.Tool`.

Current limitation: skills are mainly context snippets, not first-class workflow packages.

---

## 3. Audit Findings and Modernization Issues

### 3.1 Mux runner should become default

Current mux integration looks like an alternate path. The modernization should make mux the canonical execution path and keep the legacy runner as fallback until parity is proven.

Target:

```text
hex default runner = mux
legacy runner      = compatibility / emergency fallback
```

### 3.2 Plan mode likely loses the generated plan

Current mux plan mode appears to run planning as one mux agent run, then execution as another mux agent run.

Problem: mux `Run()` starts fresh and replaces message history with the new prompt. Therefore the second run likely does not see the first generated plan unless the plan text is explicitly included or `Continue()` is used.

Fix options:

```go
// Option A: use Continue for second turn
finalText, err := runMuxAgentRun(planPrompt)
finalText, err = runMuxAgentContinue(execPrompt)
```

or:

```go
// Option B: explicitly include the plan
execPrompt := "Here is the plan:\n\n" + finalText + "\n\nNow execute it step by step."
```

Recommendation: use `Continue()` and also include the plan text defensively in the second prompt.

### 3.3 Mux subagent output extraction likely returns empty output

`adapter.AgentRunner.RunAgent` scans messages and extracts `messages[i].Content` from the last assistant message.

Mux assistant messages are stored as `Blocks`, not necessarily `Content`.

Fix:

```go
func assistantText(m llm.Message) string {
    if m.Content != "" {
        return m.Content
    }

    var b strings.Builder
    for _, block := range m.Blocks {
        if block.Type == llm.ContentTypeText {
            b.WriteString(block.Text)
        }
    }
    return b.String()
}
```

Use this in `AgentRunner.RunAgent`.

### 3.4 `Task` tool schema is incomplete under mux

`GetToolSchema()` has explicit schemas for read/write/bash/Skill/grep/glob/edit, but `task` falls through to an empty object schema.

The `TaskTool` requires at least:

```text
prompt
description
subagent_type
```

Add explicit `task` schema so the mux model can call the tool reliably.

### 3.5 Skill tool is not wired into mux tools

`initializeSkills()` creates a skill registry and skill tool adapter, but `getHexToolsWithMuxSubagents()` currently does not include the skill tool in the mux tool list.

Fix: include `skillTool` in both root tool list and subagent `toolFactory`.

### 3.6 Permission checker is too coarse

Current `permissions.Checker.Check(toolName, params)` ignores params and makes decisions mainly from permission mode and allowed/disallowed tool lists.

This is not enough for a robust code agent.

Examples that require param-aware policy:

```text
bash: allow `go test ./...`, ask for `npm install`, deny `curl | sh`
read_file: allow workspace file, ask/deny ~/.ssh/id_rsa
write_file: allow workspace patch after approval, deny outside workspace
edit: allow repo-relative source edit, deny absolute path escape
git: allow status/diff/log, ask for commit/push
```

### 3.7 Workspace boundary is too loose

`read_file`, `write_file`, and `edit` clean and absolutize paths, but they do not enforce a workspace root boundary.

For a code agent, every file operation should be workspace-rooted by default.

Target:

```go
type Workspace struct {
    Root string
}

func (w Workspace) Resolve(rel string) (string, error) {
    clean := filepath.Clean(rel)
    if filepath.IsAbs(clean) {
        return "", fmt.Errorf("absolute paths require explicit approval")
    }

    abs := filepath.Join(w.Root, clean)
    root := filepath.Clean(w.Root)

    if abs != root && !strings.HasPrefix(abs, root+string(os.PathSeparator)) {
        return "", fmt.Errorf("path escapes workspace")
    }

    return abs, nil
}
```

### 3.8 Full-file writes should not be the primary edit path

Current `write_file` can create/overwrite/append files, and `edit` does exact string replacement. These are useful, but a robust code agent should prefer patch-first edits.

Add:

```text
apply_patch
propose_patch
```

Preferred mutation flow:

```text
read_file / grep / glob
→ propose_patch
→ apply_patch
→ git_diff
→ run_tests
→ final report
```

### 3.9 Bash tool needs policy before approval

Hex Bash always requires approval, which is good. But approval should not be the only control. The tool runs `sh -c`, supports arbitrary command strings, and can launch background processes.

Add a `CommandPolicy` before approval.

Suggested default policy:

```text
auto-allow:
  go test ./...
  go test ./... -run ...
  go test ./... -race
  npm test
  pytest
  rg ...
  git diff/status/log/show
  ls/find/cat/head/tail/sed inside workspace

ask:
  package installs
  network access
  long-running commands
  git commit
  git push
  gh pr create
  background process launch

deny:
  sudo
  rm -rf /
  chmod -R /
  chown -R /
  curl | sh
  wget | sh
  shell reading ~/.ssh, tokens, or env files unless explicitly approved
  writes outside workspace
```

### 3.10 Event stream is live but not durable

Mux events are useful for UI streaming, but a robust code agent needs durable event storage.

Add a run event log that persists every meaningful event:

```text
run_started
llm_request_started
llm_response_chunk
tool_call_requested
approval_requested
approval_resolved
tool_started
tool_finished
patch_applied
test_started
test_finished
run_completed
run_failed
```

### 3.11 Need explicit run model

Hex conversation persistence is not enough. Add first-class code-agent runs.

A run is the auditable unit of work:

```text
user request
workspace
branch/base ref
tool executions
approvals
diff
tests
artifacts
final report
```

---

## 4. Target Architecture

```text
CLI / TUI / future API
        │
        ▼
Run Service
  - create run
  - attach workspace
  - load skills
  - select runner
  - stream events
  - persist audit trail
        │
        ▼
Mux Agent / Orchestrator
  - LLM loop
  - tool calls
  - hooks
  - events
  - subagents
        │
        ▼
Tool Executor + Policy + Approval Broker
  - workspace policy
  - command policy
  - skill-scoped policy
  - user approvals
        │
        ▼
Code Tools
  - read/search/glob/grep
  - apply_patch/edit/write
  - bash/process tools
  - git/test/lint/format
  - Skill
  - Task/subagents
        │
        ▼
Durable Store
  - conversations
  - runs
  - run_events
  - tool_executions
  - approvals
  - artifacts
  - trusted_skills
```

Recommended package additions:

```text
internal/runtime/
├── run.go
├── run_store.go
├── event_store.go
├── workspace.go
├── policy.go
├── command_policy.go
├── approval_broker.go
├── report.go
└── artifacts.go

internal/tools/
├── patch_tool.go
├── git_tool.go
├── test_tool.go
├── lint_tool.go
└── format_tool.go

internal/skills/
├── selector.go
├── cards.go
├── renderer.go
├── trust.go
├── substitutions.go
└── directory.go
```

---

## 5. Modernization Goals

### 5.1 Product goals

1. Make Hex a reliable local coding agent for real repos.
2. Preserve the CLI/TUI interaction model.
3. Make mux the default execution engine.
4. Keep all code-changing actions auditable and reversible.
5. Make every run resumable, inspectable, and reportable.
6. Support Claude/Codex-style skills as reusable workflow modules.
7. Keep Hex hackable and Go-native.

### 5.2 Engineering goals

1. Use mux as the orchestration kernel.
2. Use Hex tools as concrete tool implementations during migration.
3. Add durable run/event storage without breaking conversation storage.
4. Add workspace-rooted file access.
5. Move to patch-first editing.
6. Add command/path/tool policy before approval.
7. Add skill selection, invocation, and policy semantics.
8. Add regression tests for identified mux integration bugs.

### 5.3 Non-goals

1. Do not rewrite Hex from scratch.
2. Do not move all tools into mux immediately.
3. Do not add remote/multi-user server architecture in the first phase.
4. Do not execute skill scripts automatically.
5. Do not build a skill marketplace/package manager initially.
6. Do not remove the legacy runner until mux parity is proven.

---

## 6. Run Model and Storage

### 6.1 New tables

Add migration:

```sql
CREATE TABLE IF NOT EXISTS runs (
    id TEXT PRIMARY KEY,
    conversation_id TEXT REFERENCES conversations(id) ON DELETE SET NULL,
    prompt TEXT NOT NULL,
    status TEXT NOT NULL CHECK(status IN (
        'pending',
        'running',
        'awaiting_approval',
        'completed',
        'failed',
        'cancelled'
    )),
    runner TEXT NOT NULL DEFAULT 'mux',
    model TEXT NOT NULL,
    provider TEXT,
    workspace_root TEXT,
    repo_root TEXT,
    base_ref TEXT,
    branch TEXT,
    final_text TEXT,
    error TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS run_events (
    id TEXT PRIMARY KEY,
    run_id TEXT NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
    seq INTEGER NOT NULL,
    type TEXT NOT NULL,
    payload JSON,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(run_id, seq)
);

CREATE TABLE IF NOT EXISTS tool_executions (
    id TEXT PRIMARY KEY,
    run_id TEXT NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
    tool_name TEXT NOT NULL,
    input JSON NOT NULL,
    output TEXT,
    error TEXT,
    success BOOLEAN,
    approval_id TEXT,
    started_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS approvals (
    id TEXT PRIMARY KEY,
    run_id TEXT NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
    tool_name TEXT NOT NULL,
    input JSON NOT NULL,
    risk TEXT,
    reason TEXT,
    status TEXT NOT NULL CHECK(status IN ('pending', 'approved', 'denied', 'expired')),
    requested_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    resolved_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS artifacts (
    id TEXT PRIMARY KEY,
    run_id TEXT NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
    kind TEXT NOT NULL,
    path TEXT,
    mime_type TEXT,
    content TEXT,
    metadata JSON,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_runs_conversation ON runs(conversation_id);
CREATE INDEX IF NOT EXISTS idx_runs_status ON runs(status);
CREATE INDEX IF NOT EXISTS idx_run_events_run_seq ON run_events(run_id, seq);
CREATE INDEX IF NOT EXISTS idx_tool_executions_run ON tool_executions(run_id);
CREATE INDEX IF NOT EXISTS idx_approvals_run_status ON approvals(run_id, status);
CREATE INDEX IF NOT EXISTS idx_artifacts_run ON artifacts(run_id);
```

### 6.2 Run status semantics

```text
pending            run created but not started
running            mux agent loop active
awaiting_approval  blocked on approval broker
completed          success or useful no-op
failed             unrecoverable model/tool/runtime error
cancelled          user/system cancelled run
```

### 6.3 Final run report

Every run should end with a report artifact:

```markdown
# Run Report

## Status
completed | failed | cancelled | awaiting_approval

## Request
<original prompt>

## Summary
<agent final summary>

## Files inspected
- path

## Files changed
- path: summary

## Commands run
- command: exit code, duration

## Tests / validation
- command: pass/fail, output summary

## Approvals
- tool, risk, approved/denied

## Diff summary
<git diff --stat + human summary>

## Remaining issues
<any caveats>
```

---

## 7. Workspace Manager

### 7.1 Goals

1. Prevent file tools from escaping the current workspace by default.
2. Record repo state at run start.
3. Make changes reversible.
4. Make diffs and reports deterministic.

### 7.2 Minimal local workspace behavior

For current local mode:

```text
workspace root = current git repository root, else current working directory
```

At run start:

```text
git rev-parse --show-toplevel
git status --porcelain=v1
git rev-parse HEAD
```

Record:

```go
type WorkspaceSnapshot struct {
    Root string
    RepoRoot string
    Head string
    Branch string
    Dirty bool
    Status string
}
```

### 7.3 Future worktree mode

Later:

```bash
git worktree add .hex/workspaces/<run-id> -b hex/run-<run-id>
```

This allows parallel agents safely.

### 7.4 Path resolution

All file tools should call a shared resolver:

```go
type Resolver struct {
    WorkspaceRoot string
    AllowAbsolute bool
}

func (r Resolver) Resolve(path string) (string, error)
func (r Resolver) IsInside(path string) bool
func (r Resolver) DisplayPath(abs string) string
```

Default:

```text
relative path inside workspace: allow
absolute path inside workspace: normalize and allow
absolute path outside workspace: ask or deny depending tool
path traversal escaping workspace: deny
symlink escaping workspace: ask/deny after EvalSymlinks
```

---

## 8. Tool Modernization

### 8.1 Keep existing tools

Keep:

```text
read_file
write_file
edit
bash
grep
glob
bash_output
kill_shell
ask_user_question
todo_write
web_fetch
web_search
task
Skill
```

### 8.2 Add code-agent tools

Add:

```text
apply_patch
git_status
git_diff
git_log
git_show
git_checkout_file
run_tests
run_lint
format
```

### 8.3 Patch tool

Schema:

```json
{
  "type": "object",
  "properties": {
    "patch": {
      "type": "string",
      "description": "Unified diff patch to apply."
    },
    "check_only": {
      "type": "boolean",
      "description": "If true, validate patch without applying."
    }
  },
  "required": ["patch"]
}
```

Behavior:

```text
1. Validate unified diff format.
2. Verify all changed paths are inside workspace.
3. Optionally run `git apply --check`.
4. Apply with `git apply`.
5. Return changed files and `git diff --stat`.
6. Persist patch artifact.
```

Approval:

```text
ask by default
allow auto only if skill/session policy explicitly grants apply_patch
```

### 8.4 Git tools

Read-only git tools should usually not require approval:

```text
git_status
git_diff
git_log
git_show
```

Mutating git tools should require approval:

```text
git_checkout_file
git_commit
git_push
gh_pr_create
```

Add mutating git tools only after run/event persistence is working.

### 8.5 Test/lint/format tools

`run_tests` should wrap common commands:

```json
{
  "command": "go test ./...",
  "timeout": 300,
  "working_dir": "."
}
```

It should persist structured metadata:

```json
{
  "command": "go test ./...",
  "exit_code": 0,
  "duration_seconds": 12.4,
  "stdout_lines": 100,
  "stderr_lines": 0,
  "passed": true
}
```

### 8.6 Tool result metadata

Ensure all tools return metadata consistently:

```go
type Result struct {
    ToolName string
    Success bool
    Output string
    Error string
    Metadata map[string]interface{}
}
```

For mux, map metadata through the adapter.

---

## 9. Policy and Approval Modernization

### 9.1 New policy model

```go
type DecisionKind string

const (
    DecisionAllow DecisionKind = "allow"
    DecisionAsk   DecisionKind = "ask"
    DecisionDeny  DecisionKind = "deny"
)

type PolicyInput struct {
    RunID string
    ToolName string
    Params map[string]interface{}
    WorkspaceRoot string
    ActiveSkill *skills.Skill
    PermissionMode permissions.Mode
}

type PolicyDecision struct {
    Kind DecisionKind
    Risk string
    Reason string
}

type Policy interface {
    Decide(ctx context.Context, in PolicyInput) (PolicyDecision, error)
}
```

### 9.2 Evaluation order

```text
1. Hard safety deny rules.
2. Workspace boundary rules.
3. Skill-scoped restrictions.
4. Session allow/disallow flags.
5. Command/path-specific rules.
6. Permission mode: auto / ask / deny.
7. Approval broker if ask.
```

### 9.3 Approval broker

Current mux approval is a callback. Keep it, but route through a broker:

```go
type Broker interface {
    Request(ctx context.Context, req ApprovalRequest) (ApprovalDecision, error)
}
```

Print mode behavior:

```text
ask → deny with useful message unless --approval-mode=prompt is supported
```

Interactive mode behavior:

```text
ask → show approval card → wait for user input
```

### 9.4 Approval event payload

```json
{
  "approval_id": "appr_...",
  "tool_name": "bash",
  "input": {"command": "go test ./..."},
  "risk": "medium",
  "reason": "Shell command execution requires approval",
  "policy_source": "default-command-policy"
}
```

---

## 10. Mux Runner Modernization

### 10.1 Default mux runner

Add config:

```toml
runner = "mux" # mux | legacy
```

CLI flags:

```bash
hex --runner mux
hex --runner legacy
```

Default should become mux after parity tests pass.

### 10.2 Tool construction

Replace manual tool list with a tool builder:

```go
type ToolBuildOptions struct {
    LLMClient llm.Client
    Workspace runtime.Workspace
    SkillRegistry *skills.Registry
    Policy runtime.Policy
    BackgroundRegistry *tools.BackgroundRegistry
}

func BuildTools(opts ToolBuildOptions) ([]tools.Tool, error)
```

### 10.3 Event persistence

`runMuxAgent` should persist mux events as they stream.

Pseudo:

```go
events := agent.Subscribe()

go func() {
    errChan <- agent.Run(ctx, prompt)
}()

for event := range events {
    runStore.AppendEvent(runID, ConvertMuxEvent(event))
    ui.Render(event)
}
```

### 10.4 Hooks

Keep the current hook bridge, but also add internal run hooks:

```text
SessionStart → run_started
Iteration    → iteration_started
Stop         → stop_check
Compaction   → context_compacted
SessionEnd   → run_finished
```

---

## 11. Skill System v2

### 11.1 Summary

Upgrade existing Hex skills from Markdown context snippets to Claude/Codex-style workflow packages.

Current Hex skill support is already close to the right shape. The missing pieces are:

1. Directory skills with `SKILL.md`.
2. Progressive disclosure.
3. Slash invocation.
4. Automatic selection integrated into mux.
5. Skill-scoped policy.
6. Supporting files.
7. Trust/security model.
8. CLI management commands.

### 11.2 Compatibility target

Support these layouts:

```text
# Existing Hex
.hex/skills/foo.md
~/.hex/skills/foo.md

# Claude-style
.claude/skills/foo/SKILL.md
~/.claude/skills/foo/SKILL.md

# Codex-style
.agents/skills/foo/SKILL.md
~/.agents/skills/foo/SKILL.md
/etc/codex/skills/foo/SKILL.md

# Hex-native directory
.hex/skills/foo/SKILL.md
~/.hex/skills/foo/SKILL.md
```

### 11.3 Directory skill structure

```text
my-skill/
├── SKILL.md           # required
├── scripts/           # optional executable helpers
├── references/        # optional detailed docs
├── templates/         # optional templates
├── examples/          # optional examples
├── assets/            # optional static resources
└── agents/
    └── openai.yaml    # optional Codex-style metadata
```

### 11.4 Extended manifest schema

Add fields while preserving existing ones.

```go
type Skill struct {
    Name        string   `yaml:"name"`
    Description string   `yaml:"description"`
    WhenToUse   string   `yaml:"when_to_use"`

    Tags               []string `yaml:"tags"`
    ActivationPatterns []string `yaml:"activationPatterns"`
    Paths              []string `yaml:"paths"`

    Version      string   `yaml:"version"`
    Priority     int      `yaml:"priority"`
    Dependencies []string `yaml:"dependencies"`

    Model  string `yaml:"model"`
    Effort string `yaml:"effort"`
    Context string `yaml:"context"` // inline | fork
    Agent   string `yaml:"agent"`

    UserInvocable          *bool `yaml:"user-invocable"`
    DisableModelInvocation bool  `yaml:"disable-model-invocation"`

    AllowedTools       []string `yaml:"allowed-tools"`
    DisallowedTools    []string `yaml:"disallowed-tools"`
    AllowedCommands    []string `yaml:"allowed-commands"`
    DisallowedCommands []string `yaml:"disallowed-commands"`

    Arguments    []string `yaml:"arguments"`
    ArgumentHint string   `yaml:"argument-hint"`
    Shell        string   `yaml:"shell"`

    Content string `yaml:"-"`
    FilePath string `yaml:"-"`
    RootDir string `yaml:"-"`
    Source string `yaml:"-"`
    QualifiedName string `yaml:"-"`
    ContentHash string `yaml:"-"`
}
```

Support aliases:

```text
activationPatterns ↔ activation-patterns
allowedTools ↔ allowed-tools
disallowedTools ↔ disallowed-tools
userInvocable ↔ user-invocable
disableModelInvocation ↔ disable-model-invocation
whenToUse ↔ when_to_use
argumentHint ↔ argument-hint
```

### 11.5 Example skill

```markdown
---
name: go-test-debugging
description: Diagnose and fix failing Go tests using minimal changes and targeted verification.
when_to_use: Use when the user mentions go test failures, panic output, race failures, or package-level test failures.
tags: [go, testing, debugging]
activationPatterns:
  - "go test"
  - "failing test"
  - "panic:"
  - "race detector"
paths:
  - "**/*.go"
priority: 80
dependencies:
  - verification-before-completion
allowed-tools:
  - read_file
  - grep
  - glob
  - edit
  - apply_patch
  - bash
  - todo_write
disallowed-tools:
  - write_file
allowed-commands:
  - "go test ./..."
  - "go test ./... -race"
  - "go test ./... -run *"
  - "go test ./... -count=1"
model: inherit
effort: high
---

# Go Test Debugging

Use the smallest failing test command first.

## Procedure

1. Reproduce the failure.
2. Identify the smallest package or test that fails.
3. Read only the files needed to explain the failure.
4. Make the smallest code change.
5. Re-run the exact failing test.
6. Re-run the package test.
7. Summarize root cause, patch, and verification.

## Stop Conditions

Stop and ask if:
- the failure is nondeterministic after two retries
- fixing requires architecture outside the requested area
- a command needs network access
```

### 11.6 Skill discovery order

Recommended Hex precedence:

```text
1. bundled Hex skills
2. system/admin skills
3. user skills
4. plugin skills
5. project skills
6. nested project skills relevant to touched files
```

Paths:

```text
bundled:
  ./skills
  /usr/local/share/hex/skills
  /opt/hex/skills
  ~/.hex/builtin-skills

system/admin:
  /etc/hex/skills
  /etc/codex/skills

user:
  ~/.hex/skills
  ~/.claude/skills
  ~/.agents/skills

project:
  .hex/skills
  .claude/skills
  .agents/skills
```

Conflict policy:

```text
same command name across source levels:
  project > plugin > user > system > bundled

same command name across nested project dirs:
  keep both; use qualified name for nested skill
```

Example:

```text
/deploy                root skill
/apps/web:deploy      nested skill
```

### 11.7 Progressive disclosure

Initial context should include only skill cards.

```go
type SkillCard struct {
    Command string
    Name string
    Description string
    WhenToUse string
    Source string
    Path string
    UserInvocable bool
    ModelInvocable bool
}
```

Budget:

```go
MaxInitialSkillCards = 20
MaxInitialSkillIndexChars = 8000
MaxDescriptionChars = 1536
```

Initial prompt insertion:

```markdown
## Available Relevant Skills

- go-test-debugging: Diagnose and fix failing Go tests. Use Skill(command="go-test-debugging") before editing.
- verification-before-completion: Verify changes before claiming completion. Use before final response.
```

Full skill body enters context only after:

```text
Skill(command="...")
/user slash invocation
explicit subagent preload
```

### 11.8 Automatic selection

Add:

```go
type Selector struct {
    Registry *Registry
    MaxSkills int
    MaxChars int
}

func (s *Selector) Select(prompt string, touchedPaths []string) []*Skill
```

Selection signals:

```text
1. explicit slash command
2. exact skill mention
3. activationPatterns match
4. path glob match
5. description / when_to_use lexical match
6. priority
```

### 11.9 Slash invocation

Support:

```bash
hex /go-test-debugging "failing TestRunLoop"
```

Interactive mode:

```text
/go-test-debugging failing TestRunLoop
```

Invocation behavior:

1. Resolve skill by command name or qualified name.
2. Render skill with arguments.
3. Inject rendered skill as a user/system-visible skill invocation message.
4. Apply skill policy for the turn.
5. Run mux agent.

### 11.10 Skill arguments and substitutions

Support:

```text
$ARGUMENTS
$ARGUMENTS[0]
$0
$1
$name
${HEX_SESSION_ID}
${HEX_RUN_ID}
${HEX_WORKSPACE_ROOT}
${HEX_MODEL}
${HEX_SKILL_DIR}
```

If `$ARGUMENTS` is absent but invocation has arguments, append:

```text

ARGUMENTS: <raw arguments>
```

### 11.11 Skill tool schema

Update `Skill` schema:

```json
{
  "type": "object",
  "properties": {
    "command": {
      "type": "string",
      "description": "Name of the skill to invoke"
    },
    "arguments": {
      "type": "string",
      "description": "Optional arguments passed to the skill"
    },
    "reason": {
      "type": "string",
      "description": "Why this skill is relevant"
    }
  },
  "required": ["command"]
}
```

### 11.12 Skill-scoped policy

When a skill is active:

```text
availableTools = sessionTools
if skill.allowed-tools non-empty:
  availableTools = intersection(availableTools, skill.allowed-tools)
availableTools = availableTools - skill.disallowed-tools
```

For Bash:

```text
if command matches skill.disallowed-commands: deny
if command matches skill.allowed-commands: allow/ask according to policy
otherwise: fall back to session policy
```

Restrictions should clear on the next user message unless the skill is still active in the run.

### 11.13 Skill trust model

Add table:

```sql
CREATE TABLE IF NOT EXISTS trusted_skills (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    source TEXT NOT NULL,
    path TEXT NOT NULL,
    content_hash TEXT NOT NULL,
    trusted_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

Trust behavior:

```text
builtin Markdown-only skill: trusted by default
user Markdown-only skill: trusted by ownership
project Markdown-only skill: trusted after workspace trust
plugin or remote skill: untrusted by default
skill with scripts/hooks/MCP config: explicit trust required
skill hash changed: trust invalidated
```

Do not auto-execute skill scripts. Scripts can be run only through normal tools and approval policy.

### 11.14 Skill CLI

Add:

```bash
hex skills list [--all] [--source project|user|builtin|plugin] [--json]
hex skills show <name> [--content] [--json]
hex skills search <query> [--json]
hex skills doctor [--json]
hex skills create <name> [--directory|--file]
hex skills trust <name>
hex skills untrust <name>
```

Phase 1 minimum:

```text
list
show
search
doctor
create
```

### 11.15 Builtin skills

Add or upgrade:

```text
verification-before-completion
test-driven-development
systematic-debugging
go-test-debugging
code-review
summarize-changes
```

---

## 12. Implementation Plan

### Phase 0: Confirm baseline

Tasks:

```bash
go test ./...
go test -race ./...
golangci-lint run ./...
```

If baseline is not green, record failures and fix only blockers needed for modernization.

### Phase 1: Fix mux integration bugs

Files:

```text
cmd/hex/mux_runner.go
internal/adapter/bootstrap.go
internal/tools/registry.go
```

Tasks:

1. Fix plan mode to use `Continue()` or explicitly include plan text.
2. Fix mux subagent output extraction from `Blocks`.
3. Add explicit `task` schema.
4. Wire existing `Skill` tool into mux tool list.
5. Add regression tests.

Acceptance:

```text
mux plan mode execution sees generated plan
Task tool has valid JSON schema
Skill tool appears in mux tools
mux subagent returns non-empty assistant text when blocks contain text
```

### Phase 2: Make mux default behind config flag

Files:

```text
cmd/hex/root.go
cmd/hex/interactive.go
cmd/hex/print.go
cmd/hex/mux_runner.go
internal/core/config.go
```

Tasks:

1. Add `runner = "mux" | "legacy"` config.
2. Add `--runner` flag.
3. Keep `--legacy` compatibility alias if present.
4. Route print mode and interactive mode through mux when selected.
5. Ensure legacy runner still works.

Acceptance:

```text
hex --runner mux --print "..." works
hex --runner legacy --print "..." works
config runner value is respected
```

### Phase 3: Add run/event persistence

Files:

```text
internal/storage/migrations/00000X_add_runs.up.sql
internal/runtime/run.go
internal/runtime/run_store.go
internal/runtime/event_store.go
cmd/hex/mux_runner.go
```

Tasks:

1. Add run/event/tool/approval/artifact tables.
2. Create run at start of mux execution.
3. Persist mux events.
4. Persist tool executions through executor hooks or adapter wrapper.
5. Add `hex runs list/show` later, or keep DB-only for first pass.

Acceptance:

```text
each mux run creates a runs row
each mux run stores ordered events
tool executions are persisted with input/output/error
```

### Phase 4: Workspace boundary and patch-first editing

Files:

```text
internal/runtime/workspace.go
internal/tools/read_tool.go
internal/tools/write_tool.go
internal/tools/edit_tool.go
internal/tools/patch_tool.go
internal/tools/git_tool.go
```

Tasks:

1. Add workspace resolver.
2. Update file tools to use resolver.
3. Add `apply_patch`.
4. Add `git_status` and `git_diff`.
5. Add tests for path traversal, symlinks, absolute paths.

Acceptance:

```text
file tools cannot escape workspace by default
apply_patch applies only workspace-safe patches
git_diff artifact is recorded
```

### Phase 5: Policy and approval broker

Files:

```text
internal/runtime/policy.go
internal/runtime/command_policy.go
internal/runtime/approval_broker.go
internal/permissions/checker.go
cmd/hex/mux_runner.go
internal/ui approval components
```

Tasks:

1. Add param-aware policy model.
2. Add command policy.
3. Route mux approval function through approval broker.
4. Interactive mode shows approval cards.
5. Print mode denies `ask` decisions unless explicitly configured.

Acceptance:

```text
`bash: go test ./...` can be auto-allowed by policy
`bash: curl ... | sh` is denied before approval
write outside workspace is denied
interactive ask shows approval UI
```

### Phase 6: Skill System v2 phase 1

Files:

```text
internal/skills/skill.go
internal/skills/loader.go
internal/skills/registry.go
internal/skills/tool.go
internal/skills/tool_adapter.go
cmd/hex/skills_cmd.go
cmd/hex/mux_runner.go
```

Tasks:

1. Expand manifest struct with new fields.
2. Support frontmatter aliases.
3. Add `arguments` and `reason` to `Skill` tool.
4. Add `hex skills list/show/search`.
5. Ensure mux runner always includes `Skill` tool.

Acceptance:

```text
existing .md skills still parse
Skill tool accepts command + arguments
hex skills list/show/search works
mux path can invoke Skill
```

### Phase 7: Directory skills and progressive disclosure

Files:

```text
internal/skills/directory.go
internal/skills/cards.go
internal/skills/selector.go
internal/skills/renderer.go
cmd/hex/mux_runner.go
```

Tasks:

1. Add `foo/SKILL.md` directory skill support.
2. Add `.claude/skills` and `.agents/skills` discovery.
3. Add skill cards.
4. Add automatic selector.
5. Inject compact relevant skill list into mux prompt.
6. Enforce initial skill index budget.

Acceptance:

```text
.hex/skills/foo/SKILL.md loads
.claude/skills/foo/SKILL.md loads
.agents/skills/foo/SKILL.md loads
prompt mentioning "go test" selects go-test-debugging
full skill body is not included until invoked
```

### Phase 8: Slash invocation and skill-scoped policy

Files:

```text
cmd/hex/root.go
cmd/hex/interactive.go
internal/skills/renderer.go
internal/runtime/policy.go
internal/runtime/tool_filter.go
```

Tasks:

1. Parse `/skill-name args...` in print and interactive mode.
2. Respect `user-invocable: false`.
3. Respect `disable-model-invocation: true` for automatic selection.
4. Apply `allowed-tools` and `disallowed-tools` during active skill turn.
5. Apply `allowed-commands` and `disallowed-commands` for Bash.

Acceptance:

```text
/user slash invocation works
non-user-invocable skills do not appear in slash list
disable-model-invocation skills are not auto-selected
skill-scoped tool restrictions are enforced
```

### Phase 9: Skill trust and scripts

Files:

```text
internal/skills/trust.go
internal/storage/migrations/00000Y_add_trusted_skills.up.sql
cmd/hex/skills_cmd.go
internal/runtime/policy.go
```

Tasks:

1. Add `trusted_skills` table.
2. Compute content hash for skill package.
3. Add `hex skills trust/untrust`.
4. Detect executable-supporting skills.
5. Block auto-invocation of untrusted executable skills.

Acceptance:

```text
changing SKILL.md invalidates trust
script-bearing project skill requires trust before automatic use
scripts still require normal tool approval to execute
```

---

## 13. Tests

### 13.1 Mux integration tests

```text
TestMuxPlanModePreservesPlan
TestMuxSubagentExtractsBlockText
TestMuxRunnerIncludesSkillTool
TestTaskToolHasSchema
TestMuxEventsPersisted
```

### 13.2 Runtime/storage tests

```text
TestRunStoreCreateUpdateComplete
TestRunEventStoreAppendsSequentially
TestToolExecutionPersisted
TestApprovalLifecycle
TestArtifactStore
```

### 13.3 Workspace tests

```text
TestWorkspaceResolveRelativePath
TestWorkspaceDenyTraversal
TestWorkspaceDenySymlinkEscape
TestWorkspaceAbsoluteInsideAllowed
TestWorkspaceAbsoluteOutsideDenied
```

### 13.4 Tool tests

```text
TestApplyPatchCheckOnly
TestApplyPatchRejectsPathEscape
TestApplyPatchRecordsChangedFiles
TestGitStatusReadOnlyNoApproval
TestGitDiffArtifact
```

### 13.5 Policy tests

```text
TestPolicyAllowsSafeGoTest
TestPolicyDeniesCurlPipeShell
TestPolicyAsksForGitCommit
TestPolicyDeniesWorkspaceEscape
TestSkillAllowedToolsRestrictsTools
TestSkillDisallowedToolsRemovesTools
```

### 13.6 Skill parser tests

```text
TestParseSingleFileSkill
TestParseDirectorySkill
TestParseSkillFrontmatterAliases
TestParseSkillDefaultsNameFromDirectory
TestParseSkillRejectsMissingDescriptionForDirectorySkill
TestParseSkillAllowsExistingHexFormat
```

### 13.7 Skill loader tests

```text
TestLoaderLoadsBuiltinUserPluginProject
TestLoaderProjectOverridesUser
TestLoaderDirectorySkill
TestLoaderIgnoresSupportingMarkdownFilesAsSkills
TestLoaderLoadsClaudeSkillsDirectory
TestLoaderLoadsAgentsSkillsDirectory
TestLoaderQualifiedNamesForNestedConflicts
```

### 13.8 Skill registry/selector tests

```text
TestRegistryFindByPattern
TestRegistrySearch
TestRegistryCardsRespectBudget
TestSelectorPrefersExplicitSlash
TestSelectorUsesActivationPatterns
TestSelectorUsesPaths
```

### 13.9 Skill tool tests

```text
TestSkillToolLoadsFullContent
TestSkillToolSubstitutesArguments
TestSkillToolSuggestsSimilarSkills
TestSkillToolReportsAvailableSkills
TestSkillToolRejectsDisabledSkillForModelInvocation
```

### 13.10 Security tests

```text
TestSkillHashChangesWhenContentChanges
TestScriptSkillRequiresTrust
TestUntrustedPluginSkillNotAutoInvoked
TestTrustedSkillRemainsAvailable
```

---

## 14. Documentation Updates

Update:

```text
README.md
docs/USER_GUIDE.md
docs/TOOLS.md
docs/MCP_INTEGRATION.md
docs/ARCHITECTURE.md
docs/SKILLS.md        # new
docs/RUNS.md          # new
docs/POLICY.md        # new
docs/SECURITY.md      # update or new section
```

Add user docs for:

```text
hex --runner mux
hex runs
hex skills
skill file format
SKILL.md directory skills
slash invocation
skill trust
workspace safety
approval behavior
patch-first editing
```

---

## 15. Migration Strategy

### 15.1 Backward compatibility

Keep supporting:

```text
existing .md skills
existing Hex tool names
existing conversation DB
existing permission flags
legacy runner
```

### 15.2 Deprecation path

After mux is default and stable:

```text
v1.x: legacy runner available behind --runner legacy
v2.0: mux runner only, legacy code removed or behind build tag
```

### 15.3 Tool interface migration

Short term:

```text
Hex tools.Tool → adapter → mux tool.Tool
```

Long term:

```text
mux tool.Tool = canonical interface
Hex tools implement mux tool.Tool directly
adapter removed
```

Do not do this in the same PR as runner modernization.

---

## 16. Acceptance Criteria

This modernization is done when:

```text
- mux is the default runner.
- legacy runner remains available during transition.
- mux plan mode preserves generated plans.
- mux subagents return assistant text correctly.
- Skill and Task tools have complete schemas.
- runs, run events, tool executions, approvals, and artifacts are persisted.
- file tools are workspace-rooted.
- apply_patch exists and is the preferred edit path.
- git_status and git_diff exist.
- Bash commands are governed by param-aware policy.
- interactive approvals are brokered and persisted.
- existing .md skills still work.
- SKILL.md directory skills work.
- .hex/skills, .claude/skills, and .agents/skills are discovered.
- relevant skill cards are injected with a bounded context budget.
- full skill bodies load only on invocation.
- slash skill invocation works.
- skill-scoped tool restrictions work.
- script-bearing skills require trust before automatic use.
- final run reports are produced as artifacts.
```

---

## 17. Open Decisions

1. **Default runner timing:** Should mux become default immediately behind config, or after a release cycle with opt-in telemetry/feedback?
2. **Workspace mode:** Should initial implementation mutate the current worktree or create git worktrees per run?
3. **Skill precedence:** Should Hex project skills override `.claude/skills` and `.agents/skills`, or should cross-tool compatibility paths take equal precedence?
4. **Dynamic skill context:** Should Hex implement Claude-style `!command` dynamic context injection? Recommendation: no for Phase 1; add safe placeholders later.
5. **Skill scripts:** Should scripts be executable only through Bash, or should there be a dedicated `skill_script` tool with tighter policy?
6. **Tool canonicalization:** When should Hex tools switch to mux's `tool.Tool` directly?
7. **MCP exposure:** Should Hex expose itself as an MCP server after run/event persistence lands?

---

## 18. Recommended First PRs

### PR 1: Mux runner correctness

```text
- Fix plan mode continuity.
- Fix mux subagent output extraction.
- Add Task schema.
- Wire Skill into mux tools.
- Add focused tests.
```

### PR 2: Skill CLI baseline

```text
- Add hex skills list/show/search.
- Expand Skill schema with arguments/reason.
- Add tests for existing .md skill compatibility.
```

### PR 3: Run persistence baseline

```text
- Add runs/run_events/tool_executions tables.
- Persist mux events.
- Add run report artifact skeleton.
```

### PR 4: Workspace + patch tool

```text
- Add workspace resolver.
- Update read/edit/write to use resolver.
- Add apply_patch.
- Add git_status/git_diff.
```

### PR 5: Directory skills + progressive disclosure

```text
- Add SKILL.md directory support.
- Add .claude/.agents discovery.
- Add skill cards/selector.
- Inject relevant skills into mux prompt.
```

---

## 19. Appendix: Current Repo Evidence

This spec is grounded in the current Hex/mux codebase audit:

- `README.md` describes Hex as a production-ready CLI with print mode, interactive TUI, SQLite persistence, tool system, MCP, logging, and distribution.
- `docs/ARCHITECTURE.md` describes the CLI, core, UI, storage, and tools package split.
- `cmd/hex/mux_runner.go` contains the mux runner path, provider/model selection, tool construction, system prompt construction, AGENTS.md loading, permission bridge, hook bridge, mux agent creation, and event streaming.
- `internal/adapter/tool.go` adapts Hex tools into mux tools and exposes schemas.
- `internal/adapter/bootstrap.go` creates root/subagent mux agents and contains the current subagent output extraction logic.
- `internal/tools/registry.go` contains tool schemas; `Task` currently needs an explicit schema.
- `internal/tools/read_tool.go`, `write_tool.go`, `edit_tool.go`, and `bash_tool.go` show the current file and command tool behavior.
- `internal/permissions/checker.go` shows permission checks are currently tool-name oriented and do not use params.
- `internal/storage/migrations/000001_initial_schema.up.sql` shows the current conversation/message/todo/history schema.
- `internal/skills` already implements Markdown skill parsing, loading, registry lookup/search/matching, and a `Skill` tool adapter.
- mux `orchestrator.Run()` starts fresh and `Continue()` preserves history; this matters for Hex plan mode.
- mux stores assistant responses as content blocks, which matters for subagent output extraction.

External compatibility targets:

- Claude Code skills use `SKILL.md`, support direct `/skill-name` invocation, automatic loading when relevant, supporting files, frontmatter controls, dynamic context injection, and skill-scoped tool controls.
- Codex skills use directory packages with `SKILL.md`, optional `scripts/`, `references/`, `assets/`, and `agents/openai.yaml`; Codex uses progressive disclosure with skill name, description, and path initially, then loads full `SKILL.md` when selected; Codex discovers repository skills under `.agents/skills`.

---

## 20. Final Direction

The modernization should make Hex feel like a serious local software-engineering agent:

```text
- it understands repo instructions,
- chooses relevant skills,
- plans without losing state,
- edits with patches,
- validates with tests,
- asks before risky actions,
- records every meaningful action,
- produces a final report,
- and exposes the whole system through a clean mux-powered architecture.
```

The clean split remains:

```text
Hex owns product/runtime/UI/persistence/policy.
Mux owns reusable orchestration/tool/agent primitives.
Skills bridge human workflow knowledge into the agent loop.
```
