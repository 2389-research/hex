# Claude Code Documentation Index

This documentation suite explains how **Claude Code CLI** works as an AI-powered terminal assistant. These docs are for understanding the tool itself, not any specific project.

## Navigation

### Core Architecture
- **[01-OVERVIEW.md](./01-OVERVIEW.md)** - What Claude Code is and how it works
- **[02-FILE-OPERATIONS.md](./02-FILE-OPERATIONS.md)** - Read, Edit, Write tool deep dive
- **[03-TOOL-SYSTEM.md](./03-TOOL-SYSTEM.md)** - Complete tool inventory and selection
- **[04-BASH-AND-COMMAND-EXECUTION.md](./04-BASH-AND-COMMAND-EXECUTION.md)** - Shell execution patterns
- **[32-CLI-REFERENCE.md](./32-CLI-REFERENCE.md)** - Command-line interface reference

### Extensibility Systems
- **[05-SUBAGENT-SYSTEM.md](./05-SUBAGENT-SYSTEM.md)** - Isolated task delegation
- **[06-HOOKS-SYSTEM.md](./06-HOOKS-SYSTEM.md)** - Lifecycle event automation
- **[07-SKILLS-SYSTEM.md](./07-SKILLS-SYSTEM.md)** - Model-invoked capabilities
- **[08-PLUGINS.md](./08-PLUGINS.md)** - Plugin architecture and creation

### Integration & Configuration
- **[09-MCP-SERVERS.md](./09-MCP-SERVERS.md)** - Model Context Protocol integration
- **[10-SLASH-COMMANDS.md](./10-SLASH-COMMANDS.md)** - Built-in and custom commands
- **[11-CONFIGURATION.md](./11-CONFIGURATION.md)** - Settings hierarchy and management
- **[12-VERIFICATION-AND-TESTING.md](./12-VERIFICATION-AND-TESTING.md)** - Testing workflows

### Internal Workflows
- **[13-TODOWRITE-WORKFLOW.md](./13-TODOWRITE-WORKFLOW.md)** - Task tracking system
- **[14-DECISION-FRAMEWORK.md](./14-DECISION-FRAMEWORK.md)** - Autonomous vs collaborative actions
- **[15-GIT-WORKFLOWS.md](./15-GIT-WORKFLOWS.md)** - Commits, PRs, pre-commit protocol
- **[16-ERROR-HANDLING.md](./16-ERROR-HANDLING.md)** - Recovery and debugging
- **[17-CONTEXT-MANAGEMENT.md](./17-CONTEXT-MANAGEMENT.md)** - Token limits and memory
- **[18-SECURITY-MODEL.md](./18-SECURITY-MODEL.md)** - Permissions and sandboxing
- **[19-BEST-PRACTICES.md](./19-BEST-PRACTICES.md)** - Learned patterns and workflows
- **[20-GLOSSARY.md](./20-GLOSSARY.md)** - Complete terminology reference

### Advanced Features
- **[21-CHRONICLE-INTEGRATION.md](./21-CHRONICLE-INTEGRATION.md)** - Ambient activity logging
- **[22-PRIVATE-JOURNAL.md](./22-PRIVATE-JOURNAL.md)** - Private reflection and learning
- **[23-ADVANCED-PATTERNS.md](./23-ADVANCED-PATTERNS.md)** - Secret techniques and magic
- **[24-SESSION-CAPABILITIES.md](./24-SESSION-CAPABILITIES.md)** - MCP capability model
- **[25-PROMPTING-STRATEGIES.md](./25-PROMPTING-STRATEGIES.md)** - Prompting patterns and techniques
- **[26-SUBAGENT-DEEP-DIVE.md](./26-SUBAGENT-DEEP-DIVE.md)** - Subagent implementation details
- **[27-EDIT-STRATEGIES.md](./27-EDIT-STRATEGIES.md)** - Editing strategies
- **[28-PLAN-MODE-WORKFLOW.md](./28-PLAN-MODE-WORKFLOW.md)** - Plan mode workflow
- **[29-TOOL-DEEP-DIVES.md](./29-TOOL-DEEP-DIVES.md)** - Advanced tool usage guides
- **[30-MCP-SERVER-GUIDES.md](./30-MCP-SERVER-GUIDES.md)** - Practical MCP server patterns
- **[31-JJ-WORKFLOWS.md](./31-JJ-WORKFLOWS.md)** - JJ (Jujutsu) VCS workflows

## Quick Start Paths

### For AI Assistants Learning Claude Code
1. Start with [01-OVERVIEW.md](./01-OVERVIEW.md) to understand architecture
2. Read [02-FILE-OPERATIONS.md](./02-FILE-OPERATIONS.md) - most critical for daily work
3. Study [03-TOOL-SYSTEM.md](./03-TOOL-SYSTEM.md) for tool selection
4. Review [04-BASH-AND-COMMAND-EXECUTION.md](./04-BASH-AND-COMMAND-EXECUTION.md) for command patterns

### For Understanding File Editing
1. [02-FILE-OPERATIONS.md](./02-FILE-OPERATIONS.md) - Complete file operation guide
2. Focus on Edit tool constraints (exact matching, uniqueness)
3. Study line number prefix handling
4. Review real-world examples

### For Command Execution
1. [04-BASH-AND-COMMAND-EXECUTION.md](./04-BASH-AND-COMMAND-EXECUTION.md) - Bash tool deep dive
2. Understand persistent sessions
3. Learn command chaining patterns
4. Study git operation workflows
5. Reference [32-CLI-REFERENCE.md](./32-CLI-REFERENCE.md) for flags and subcommands

## Document Structure

### Core Architecture (6 docs)
- **00-INDEX.md** - This file (navigation and quick reference)
- **01-OVERVIEW.md** - What Claude Code is and how it works
- **02-FILE-OPERATIONS.md** - Read, Edit, Write tool deep dive
- **03-TOOL-SYSTEM.md** - Complete tool inventory and selection
- **04-BASH-AND-COMMAND-EXECUTION.md** - Shell execution patterns
- **32-CLI-REFERENCE.md** - Command-line options and subcommands

### Extensibility Systems (4 docs)
- **05-SUBAGENT-SYSTEM.md** - Isolated task delegation
- **06-HOOKS-SYSTEM.md** - Lifecycle event automation (10 events)
- **07-SKILLS-SYSTEM.md** - Model-invoked capabilities
- **08-PLUGINS.md** - Plugin architecture and creation

### Integration & Configuration (4 docs)
- **09-MCP-SERVERS.md** - Model Context Protocol integration
- **10-SLASH-COMMANDS.md** - Built-in and custom commands
- **11-CONFIGURATION.md** - Settings hierarchy and management
- **12-VERIFICATION-AND-TESTING.md** - Testing workflows and quality gates

### Internal Workflows (8 docs)
- **13-TODOWRITE-WORKFLOW.md** - Task tracking system
- **14-DECISION-FRAMEWORK.md** - Autonomous vs collaborative actions
- **15-GIT-WORKFLOWS.md** - Commits, PRs, pre-commit protocol
- **16-ERROR-HANDLING.md** - Recovery and debugging strategies
- **17-CONTEXT-MANAGEMENT.md** - Token limits and memory (200k budget)
- **18-SECURITY-MODEL.md** - Permissions and sandboxing
- **19-BEST-PRACTICES.md** - Learned patterns and workflows
- **20-GLOSSARY.md** - Complete terminology reference

### Advanced Features (11 docs)
- **21-CHRONICLE-INTEGRATION.md** - Ambient activity logging (remember_this, what_was_i_doing)
- **22-PRIVATE-JOURNAL.md** - Private reflection and learning
- **23-ADVANCED-PATTERNS.md** - Secret techniques and magic (ast-grep, ABOUTME, Beads)
- **24-SESSION-CAPABILITIES.md** - MCP capability model
- **25-PROMPTING-STRATEGIES.md** - Prompting patterns and techniques
- **26-SUBAGENT-DEEP-DIVE.md** - Technical implementation of subagent architecture (invocation, context isolation, lifecycle)
- **27-EDIT-STRATEGIES.md** - File editing strategies (surgical precision, context calibration, whitespace discipline)
- **28-PLAN-MODE-WORKFLOW.md** - Plan mode workflow (/superpowers:write-plan, ExitPlanMode, implementation planning)
- **29-TOOL-DEEP-DIVES.md** - Advanced tools (NotebookEdit, AskUserQuestion, ExitPlanMode, MCP resources)
- **30-MCP-SERVER-GUIDES.md** - Practical MCP server usage (Chronicle, Playwright, Toki, Chrome, Social Media)
- **31-JJ-WORKFLOWS.md** - JJ (Jujutsu) VCS workflows and git equivalents

### Overview
- **README.md** - Documentation overview and navigation

## Key Principles

### Unix Philosophy
- Do one thing well
- Tools compose together
- Text-based interfaces
- Small, focused operations

### Agentic Approach
- Read before edit (always)
- Verify before commit
- Test before complete
- Ask when uncertain

### Safety First
- No destructive operations without verification
- Read files before modifying
- Check test output
- Respect pre-commit hooks

## Common Workflows

### Feature Development
```
1. Search codebase (Grep/Glob)
2. Read relevant files (Read)
3. Edit code (Edit)
4. Run tests (Bash)
5. Commit changes (Bash + git)
```

### Bug Investigation
```
1. Reproduce error (Bash)
2. Search for error patterns (Grep)
3. Read suspicious files (Read)
4. Add instrumentation (Edit)
5. Test hypothesis (Bash)
```

### Codebase Exploration
```
1. Find file patterns (Glob)
2. Search for implementations (Grep)
3. Read key files (Read)
4. Map dependencies (analysis)
5. Document findings (Write/Edit)
```

## Terminology

- **Tool** - A capability Claude Code can invoke (Read, Edit, Bash, etc.)
- **Working Directory** - Current directory context (can change between operations)
- **Line Number Prefix** - Format: `spaces + line_number + tab` in Read output
- **Exact String Matching** - Edit tool requires precise text matches
- **Parallel Execution** - Multiple independent tool calls in one response
- **Sequential Execution** - Tool calls that depend on previous results
- **Persistent Shell** - Bash sessions that maintain state across calls
- **Pre-commit Hooks** - Git hooks that run before commits (never bypass)

## Reading This Documentation

### Notation Conventions
- `tool_name` - References to specific tools
- **Bold** - Important concepts
- *Italic* - Emphasis
- `code blocks` - Actual code or commands
- > Quotes - Important warnings or notes

### Code Examples
All examples are real-world patterns from actual Claude Code usage. They demonstrate both correct and incorrect approaches.

### Diagrams
Where helpful, ASCII diagrams show:
- Tool execution flow
- File operation sequences
- Decision trees for tool selection
- State transitions

## Contributing to These Docs

These docs evolve as Claude Code capabilities expand. When adding new sections:

1. Focus on HOW the tool works, not what to build with it
2. Include real examples from actual usage
3. Document both correct usage and common pitfalls
4. Explain the "why" behind constraints
5. Cross-reference related sections

## Version

These docs describe Claude Code as of December 2025. Core tools (Read, Edit, Write, Bash, Grep, Glob) are stable. Newer features (MCP servers, skills, slash commands) may evolve.

---

*Start with [01-OVERVIEW.md](./01-OVERVIEW.md) to understand what Claude Code is and how it works.*
