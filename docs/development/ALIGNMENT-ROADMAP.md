# Hex → Claude Code Alignment Roadmap

**Created:** 2025-12-01
**Purpose:** Phased plan to align Hex with Claude Code's feature set
**Source:** [CLAUDE-CODE-AUDIT.md](./CLAUDE-CODE-AUDIT.md)

## Executive Summary

Hex has a solid foundation with excellent core functionality. To reach feature parity with Claude Code, we need to focus on **extensibility systems** that enable users to customize and automate workflows.

**Current State:** ~65% feature parity
**Target State:** 95%+ feature parity
**Timeline:** 6 phases over ~16 weeks

## Strategic Priorities

### What We're Doing Well
✅ Core tool system
✅ Session persistence
✅ TUI/UX
✅ MCP integration
✅ Basic CLI interface

### Critical Gaps
❌ Hooks system (automation)
❌ Skills system (knowledge)
❌ Plugin system (distribution)
❌ Slash commands (shortcuts)
⚠️ Permission system (granular control)

### Philosophy

**80/20 Approach:** Focus on systems that provide maximum extensibility with minimal complexity.

**Hooks + Skills > Plugins:** Implementing hooks and skills gives 80% of the value without the complexity of a full plugin marketplace.

## Phase Breakdown

### Phase 1: Hooks System (2 weeks)

**Goal:** Enable workflow automation through lifecycle events

**Deliverables:**
- Hooks configuration system (`.claude/settings.json`)
- Hook execution engine (shell script support)
- 10 lifecycle events:
  - SessionStart
  - SessionEnd
  - PreToolUse
  - PostToolUse
  - ToolBlocked
  - UserPromptSubmit
  - BeforeMessageSend
  - AfterMessageReceive
  - PreCompact
  - SubagentStop

**Implementation:**
```
internal/hooks/
├── engine.go       # Hook execution
├── config.go       # Hook configuration
├── events.go       # Event definitions
└── executor.go     # Shell script runner
```

**Success Criteria:**
- [ ] Can run shell commands on SessionStart
- [ ] Can auto-format files after Edit tool (PostToolUse)
- [ ] Can inject context on UserPromptSubmit
- [ ] Hooks configured via .claude/settings.json
- [ ] All 10 events implemented

**Example Use Case:**
```json
{
  "hooks": {
    "PostToolUse": {
      "Edit": ["biome format ${file}"],
      "Write": ["git add ${file}"]
    },
    "SessionStart": ["echo 'Welcome to Hex!' >&2"]
  }
}
```

---

### Phase 2: Skills System (2-3 weeks)

**Goal:** Enable domain knowledge modules that guide behavior

**Deliverables:**
- Skill file format (markdown with YAML frontmatter)
- Skill discovery system (`.claude/skills/`, `~/.claude/skills/`)
- Skill tool (invoke skills from Claude)
- 5-10 starter skills:
  - TDD workflow
  - Systematic debugging
  - Code review checklist
  - Testing patterns
  - Git workflows

**Implementation:**
```
internal/skills/
├── loader.go       # Skill discovery
├── matcher.go      # Pattern matching
├── registry.go     # Skill registry
└── tool.go         # Skill tool
```

**Success Criteria:**
- [ ] Can load .md skills from disk
- [ ] Skill tool invokes skill content
- [ ] Skills auto-match based on keywords
- [ ] User/project skill directories work
- [ ] 5+ production skills shipped

**Example Skill:**
```markdown
---
name: test-driven-development
description: Use when implementing features - enforces TDD workflow
triggers: ["implement", "feature", "add functionality"]
---

# Test-Driven Development

## Process

1. Write failing test first
2. Run test to see it fail
3. Write minimal code to pass
4. Run test to see it pass
5. Refactor if needed

## Checklist

- [ ] Test written and failing
- [ ] Implementation passes test
- [ ] Edge cases covered
```

---

### Phase 3: Permission System Enhancement (1 week)

**Goal:** Granular permission control for tools

**Deliverables:**
- `--permission-mode` flag (auto, ask, deny)
- `--allowed-tools` / `--disallowed-tools` flags
- Permission rules configuration
- PermissionRequest hook

**Implementation:**
```
internal/permissions/
├── modes.go        # Permission modes
├── rules.go        # Allow/deny rules
├── checker.go      # Permission checking
└── config.go       # Configuration
```

**Success Criteria:**
- [ ] `--permission-mode auto` skips all prompts
- [ ] `--permission-mode deny` blocks all tools
- [ ] `--allowed-tools Read,Write` restricts to subset
- [ ] `--disallowed-tools Bash` blocks specific tools
- [ ] PermissionRequest hook fires before prompts

**Example:**
```bash
# Audit mode - deny destructive operations
hex "analyze codebase" --disallowed-tools Edit,Write,Bash

# CI mode - auto-approve safe operations
hex "run tests" --permission-mode auto --allowed-tools Bash,Read
```

---

### Phase 4: Slash Commands (1-2 weeks)

**Goal:** Custom prompt shortcuts for common workflows

**Deliverables:**
- `.claude/commands/` directory support
- SlashCommand tool
- Command discovery
- 5-10 built-in commands:
  - `/plan` - create implementation plan
  - `/brainstorm` - design exploration
  - `/review` - code review
  - `/debug` - systematic debugging
  - `/test` - write tests
  - `/commit` - review and commit changes

**Implementation:**
```
internal/commands/
├── loader.go       # Command discovery
├── registry.go     # Command registry
├── tool.go         # SlashCommand tool
└── builtins.go     # Built-in commands
```

**Success Criteria:**
- [ ] `/plan` expands to planning prompt
- [ ] Custom commands in `.claude/commands/*.md` work
- [ ] Commands have descriptions and arguments
- [ ] Tab completion for slash commands
- [ ] 5+ built-in commands available

**Example Command:**
```markdown
---
name: review
description: Perform code review on changes
args: [file]
---

Review the changes in {{.file}}. Check for:

1. Code quality and readability
2. Potential bugs or edge cases
3. Test coverage
4. Security vulnerabilities
5. Performance concerns

Provide specific, actionable feedback.
```

---

### Phase 5: Subagent Framework (2-3 weeks)

**Goal:** Proper isolated subagent execution

**Deliverables:**
- Enhance Task tool with context isolation
- Subagent types (general, Explore, Plan, code-reviewer)
- Parallel dispatch support
- SubagentStop hook

**Implementation:**
```
internal/subagents/
├── types.go        # Subagent type definitions
├── executor.go     # Isolated execution
├── dispatcher.go   # Parallel dispatch
└── context.go      # Context management
```

**Success Criteria:**
- [ ] Subagents run with isolated context
- [ ] Multiple subagents dispatch in parallel
- [ ] Explore agent for codebase navigation
- [ ] Plan agent for design work
- [ ] code-reviewer agent for quality checks
- [ ] SubagentStop hook fires on completion

**Example:**
```
Task(
  subagent_type: "Explore",
  prompt: "Find all authentication code",
  thoroughness: "medium"
)
```

---

### Phase 6: Plugin System (4-6 weeks)

**Goal:** Distributable extension bundles

**Deliverables:**
- Plugin manifest format (plugin.json)
- Plugin discovery (`~/.hex/plugins/`)
- Plugin lifecycle (install, enable, disable, uninstall)
- `hex plugin` subcommands
- Plugin marketplace integration (optional)

**Implementation:**
```
internal/plugins/
├── manifest.go     # plugin.json schema
├── loader.go       # Plugin discovery
├── installer.go    # Install/uninstall
├── registry.go     # Plugin registry
└── cmd.go          # Plugin commands

cmd/hex/plugin.go  # Subcommand
```

**Success Criteria:**
- [ ] `hex plugin install <url>` works
- [ ] `hex plugin list` shows installed plugins
- [ ] Plugins can bundle skills + hooks + MCP servers
- [ ] Plugin enable/disable toggles functionality
- [ ] Plugin updates supported

**Example Plugin:**
```json
{
  "name": "go-development",
  "version": "1.0.0",
  "description": "Go development tools and workflows",
  "skills": ["go-tdd", "go-testing", "go-debugging"],
  "hooks": {
    "PostToolUse": {
      "Edit": ["gofmt -w ${file}"]
    }
  },
  "mcpServers": {
    "go-tools": {
      "command": "go-mcp-server"
    }
  }
}
```

---

## Advanced Features (Phase 7+)

### Additional CLI Flags
- `--fork-session` - Branch from checkpoint
- `--max-thinking-tokens` - Extended thinking
- `--max-turns` - Conversation turn limit
- `--max-budget-usd` - Cost controls
- `--fallback-model` - Model failover
- `--json-schema` - Structured output validation
- Stream JSON format

### MCP Enhancements
- MCP resource tools (ListMcpResourcesTool, ReadMcpResourceTool)
- `hex mcp serve` - Run as MCP server
- `hex mcp tools` / `mcp resources` - Diagnostics
- Plugin-bundled MCP servers

### Missing Tools
- NotebookEdit - Jupyter notebook editing
- ExitPlanMode - Planning workflow integration

### Configuration System
- Multi-layer settings (user, project, local)
- Runtime config overrides (`--settings`)
- Config hierarchy and merging
- `--setting-sources` flag

---

## Implementation Timeline

| Phase | Duration | Cumulative | Feature Parity |
|-------|----------|------------|----------------|
| Phase 1: Hooks | 2 weeks | 2 weeks | 70% |
| Phase 2: Skills | 3 weeks | 5 weeks | 75% |
| Phase 3: Permissions | 1 week | 6 weeks | 80% |
| Phase 4: Slash Commands | 2 weeks | 8 weeks | 85% |
| Phase 5: Subagents | 3 weeks | 11 weeks | 90% |
| Phase 6: Plugins | 5 weeks | 16 weeks | 95% |

**Total:** ~16 weeks to 95% feature parity

---

## Success Metrics

### Phase 1 Success
- Can automate formatting on file edits
- Can inject context on every prompt
- Can track tool usage
- Hooks used in real workflows

### Phase 2 Success
- TDD skill enforces test-first workflow
- Debug skill guides systematic investigation
- Users create custom skills
- Skills reduce repeated explanations

### Phase 3 Success
- CI runs with `--permission-mode auto`
- Audit mode blocks destructive tools
- Permission rules prevent accidental operations
- PermissionRequest hook enables policy enforcement

### Phase 4 Success
- `/plan` becomes primary planning flow
- Custom project commands streamline workflows
- Users discover commands via tab completion
- Commands replace repeated prompts

### Phase 5 Success
- Parallel subagents speed up large tasks
- Explore agent finds code faster than manual search
- code-reviewer provides consistent feedback
- Subagents reduce context window pressure

### Phase 6 Success
- Users install plugins from marketplace
- Community plugins extend functionality
- Organizations distribute internal plugins
- Plugin ecosystem emerges

---

## Risk Mitigation

### Technical Risks

**Risk:** Hooks slow down tool execution
**Mitigation:** Make hooks optional, async where possible, timeout protection

**Risk:** Skills bloat system prompts
**Mitigation:** Lazy loading, keyword matching, context budget limits

**Risk:** Plugins introduce security issues
**Mitigation:** Sandboxing, permission system integration, code review process

**Risk:** Subagent isolation breaks workflows
**Mitigation:** Careful context design, parent context injection, debugging tools

### Product Risks

**Risk:** Too much complexity confuses users
**Mitigation:** Progressive disclosure, good defaults, excellent documentation

**Risk:** Breaking changes during implementation
**Mitigation:** Feature flags, backward compatibility, migration guides

**Risk:** Community doesn't adopt extensibility
**Mitigation:** Provide excellent built-in skills/hooks, showcase use cases

---

## Decision Framework

### When to Implement?

**Yes:** Hooks + Skills (Phase 1-2)
- High value, medium complexity
- Enable 80% of extensibility use cases
- Foundation for other features

**Maybe:** Plugins (Phase 6)
- High value, high complexity
- Distribution story matters
- Can wait for community demand

**Later:** Advanced flags (Phase 7)
- Lower priority, nice to have
- Polish items
- User-driven prioritization

---

## Next Steps

1. **Review and Approve Roadmap** - Stakeholder alignment
2. **Create Phase 1 Design Doc** - Detailed hooks implementation plan
3. **Prototype Hooks** - Validate approach with simple example
4. **Full Implementation** - Execute Phase 1
5. **Iterate** - Learn and adjust for Phase 2

---

## References

- [CLAUDE-CODE-AUDIT.md](./CLAUDE-CODE-AUDIT.md) - Feature comparison
- [../claude-docs/](../claude-docs/) - Claude Code documentation
- [ARCHITECTURE.md](../ARCHITECTURE.md) - Hex architecture

---

**Last Updated:** 2025-12-01
**Status:** Proposed
**Owner:** Hex Team
