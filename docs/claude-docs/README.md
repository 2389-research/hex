# Claude Code - How It Works

**Complete Internal Documentation**

**Purpose**: Comprehensive guide to how Claude Code CLI works internally - for developers, AI researchers, and anyone building similar systems.

**Date**: 2025-12-01

---

## What This Is

This documentation explains **how Claude Code works** - the AI assistant CLI tool from Anthropic. This is NOT about any specific project or deobfuscation work. This is about the tool itself.

## Quick Start

**New to Claude Code?** Start here:
1. [01-OVERVIEW.md](01-OVERVIEW.md) - What Claude Code is
2. [02-FILE-OPERATIONS.md](02-FILE-OPERATIONS.md) - How it edits files
3. [03-TOOL-SYSTEM.md](03-TOOL-SYSTEM.md) - Available tools
4. [04-BASH-AND-COMMAND-EXECUTION.md](04-BASH-AND-COMMAND-EXECUTION.md) - Command execution
5. [32-CLI-REFERENCE.md](32-CLI-REFERENCE.md) - Runtime flags and subcommands

**Building similar tools?** Go here:
1. [01-OVERVIEW.md](01-OVERVIEW.md) - Architecture overview
2. [05-SUBAGENT-SYSTEM.md](05-SUBAGENT-SYSTEM.md) - Parallel agent design
3. [06-HOOKS-SYSTEM.md](06-HOOKS-SYSTEM.md) - Lifecycle events
4. [14-DECISION-FRAMEWORK.md](14-DECISION-FRAMEWORK.md) - How Claude decides

**Extending Claude Code?** Check these:
1. [07-SKILLS-SYSTEM.md](07-SKILLS-SYSTEM.md) - Adding capabilities
2. [08-PLUGINS.md](08-PLUGINS.md) - Plugin architecture
3. [09-MCP-SERVERS.md](09-MCP-SERVERS.md) - MCP integration
4. [10-SLASH-COMMANDS.md](10-SLASH-COMMANDS.md) - Custom commands

---

## Documentation Structure

### Core Architecture (6 docs)
- **[00-INDEX.md](00-INDEX.md)** - Navigation and quick reference
- **[01-OVERVIEW.md](01-OVERVIEW.md)** - What Claude Code is and how it works
- **[02-FILE-OPERATIONS.md](02-FILE-OPERATIONS.md)** - Read, Edit, Write tool deep dive
- **[03-TOOL-SYSTEM.md](03-TOOL-SYSTEM.md)** - Complete tool inventory
- **[04-BASH-AND-COMMAND-EXECUTION.md](04-BASH-AND-COMMAND-EXECUTION.md)** - Shell execution
- **[32-CLI-REFERENCE.md](32-CLI-REFERENCE.md)** - Command-line options and subcommands

### Extensibility Systems (4 docs)
- **[05-SUBAGENT-SYSTEM.md](05-SUBAGENT-SYSTEM.md)** - Isolated task delegation
- **[06-HOOKS-SYSTEM.md](06-HOOKS-SYSTEM.md)** - Lifecycle event automation
- **[07-SKILLS-SYSTEM.md](07-SKILLS-SYSTEM.md)** - Model-invoked capabilities
- **[08-PLUGINS.md](08-PLUGINS.md)** - Plugin architecture and creation

### Integration & Config (4 docs)
- **[09-MCP-SERVERS.md](09-MCP-SERVERS.md)** - Model Context Protocol
- **[10-SLASH-COMMANDS.md](10-SLASH-COMMANDS.md)** - Built-in and custom commands
- **[11-CONFIGURATION.md](11-CONFIGURATION.md)** - Settings and hierarchy
- **[12-VERIFICATION-AND-TESTING.md](12-VERIFICATION-AND-TESTING.md)** - Testing workflows

### Internal Workflows (8 docs)
- **[13-TODOWRITE-WORKFLOW.md](13-TODOWRITE-WORKFLOW.md)** - Task tracking system
- **[14-DECISION-FRAMEWORK.md](14-DECISION-FRAMEWORK.md)** - How Claude decides what to do
- **[15-GIT-WORKFLOWS.md](15-GIT-WORKFLOWS.md)** - Commits, PRs, branches
- **[16-ERROR-HANDLING.md](16-ERROR-HANDLING.md)** - Recovery and debugging
- **[17-CONTEXT-MANAGEMENT.md](17-CONTEXT-MANAGEMENT.md)** - Token limits and memory
- **[18-SECURITY-MODEL.md](18-SECURITY-MODEL.md)** - Permissions and sandboxing
- **[19-BEST-PRACTICES.md](19-BEST-PRACTICES.md)** - Learned patterns
- **[20-GLOSSARY.md](20-GLOSSARY.md)** - Terms and definitions

### Advanced Features (11 docs)
- **[21-CHRONICLE-INTEGRATION.md](21-CHRONICLE-INTEGRATION.md)** - Ambient activity logging
- **[22-PRIVATE-JOURNAL.md](22-PRIVATE-JOURNAL.md)** - Private reflection and learning
- **[23-ADVANCED-PATTERNS.md](23-ADVANCED-PATTERNS.md)** - Secret techniques and magic
- **[24-SESSION-CAPABILITIES.md](24-SESSION-CAPABILITIES.md)** - MCP capability model
- **[25-PROMPTING-STRATEGIES.md](25-PROMPTING-STRATEGIES.md)** - Prompting patterns and techniques
- **[26-SUBAGENT-DEEP-DIVE.md](26-SUBAGENT-DEEP-DIVE.md)** - Technical implementation of subagent architecture
- **[27-EDIT-STRATEGIES.md](27-EDIT-STRATEGIES.md)** - File editing strategies and techniques
- **[28-PLAN-MODE-WORKFLOW.md](28-PLAN-MODE-WORKFLOW.md)** - Plan mode workflow and implementation planning
- **[29-TOOL-DEEP-DIVES.md](29-TOOL-DEEP-DIVES.md)** - Advanced tool usage guides
- **[30-MCP-SERVER-GUIDES.md](30-MCP-SERVER-GUIDES.md)** - Practical MCP server patterns
- **[31-JJ-WORKFLOWS.md](31-JJ-WORKFLOWS.md)** - JJ (Jujutsu) VCS workflows

---

## Key Concepts

### Agentic Behavior
Claude Code is not a chatbot. It's an **autonomous agent** that:
- Reads and writes files
- Executes shell commands
- Manages git operations
- Delegates to subagents
- Runs tests and verifies
- Creates commits and PRs

### Tool-Based Architecture
Everything happens through **tools**:
- **Read**: Load file contents
- **Edit**: Exact string replacement
- **Write**: Create or overwrite files
- **Bash**: Execute shell commands
- **Task**: Launch subagents
- **TodoWrite**: Track progress

### Verification First
Claude Code follows a strict verification protocol:
1. Read before edit
2. Test after change
3. Verify before claim
4. Evidence before assertion

### Unix Philosophy
- **Composable**: Works with existing tools
- **Scriptable**: Can be automated
- **Transparent**: Shows all operations
- **Deterministic**: Hooks for guaranteed behavior

---

## Architecture Diagram

```
┌─────────────────────────────────────────────┐
│          User Input (Terminal)              │
└──────────────────┬──────────────────────────┘
                   │
        ┌──────────▼──────────┐
        │   Claude Code CLI   │
        │   (Main Process)    │
        └──────────┬──────────┘
                   │
        ┌──────────▼──────────────────────────┐
        │      Decision Framework              │
        │  (Autonomous, Skills, Subagents)     │
        └──────────┬──────────────────────────┘
                   │
        ┌──────────▼──────────┐
        │    Tool Selection    │
        └──────────┬──────────┘
                   │
    ┌──────────────┼──────────────┐
    │              │              │
┌───▼────┐   ┌────▼─────┐  ┌────▼────┐
│  Read  │   │   Edit   │  │  Bash   │
│  Write │   │   Glob   │  │  Task   │
│  Grep  │   │   MCP    │  │  Todo   │
└───┬────┘   └────┬─────┘  └────┬────┘
    │              │              │
    └──────────────┼──────────────┘
                   │
        ┌──────────▼──────────┐
        │   Hooks (Events)     │
        │  - PreToolUse        │
        │  - PostToolUse       │
        │  - UserPromptSubmit  │
        └──────────┬──────────┘
                   │
        ┌──────────▼──────────┐
        │   File System /      │
        │   Shell / Git        │
        └──────────────────────┘
```

---

## Documentation Statistics

- **Total Files**: 29 comprehensive guides
- **Coverage**: All major systems documented
- **Code Examples**: Extensive throughout
- **Cross-references**: Fully linked
- **Depth**: Implementation-level detail

---

## How to Use This Documentation

1. **Read the overview** (01-OVERVIEW.md) to understand the big picture
2. **Choose your path** based on your goal (see Quick Start above)
3. **Follow cross-references** for deep dives
4. **Try examples** in your own environment
5. **Reference the index** (00-INDEX.md) for quick lookups

---

## What Makes Claude Code Unique

### vs GitHub Copilot
- **Autonomous**: Makes decisions without constant prompting
- **File-level**: Edits entire files, not just suggestions
- **Tool access**: Can run tests, git, build tools
- **Verification**: Tests changes automatically

### vs Cursor
- **Terminal-first**: CLI, not IDE extension (though VS Code extension exists)
- **Unix philosophy**: Composable, scriptable
- **Subagents**: Parallel task delegation
- **Extensible**: Plugins, hooks, skills, MCP

### vs ChatGPT Code Interpreter
- **Local execution**: Runs on your machine
- **Project context**: Full codebase access
- **Git integration**: Commits, PRs, branches
- **Persistent**: Continuous sessions

---

## Contributing

This documentation is maintained as part of the Claude Code project analysis. To update:

1. Keep docs focused on HOW Claude Code works (not project-specific usage)
2. Include real examples and code samples
3. Cross-reference related documents
4. Update the index (00-INDEX.md) when adding new docs

---

## See Also

- **Official Docs**: https://code.claude.com/docs/
- **Deobfuscation Project**: ../deobfuscation-analysis/ (separate docs)

---

**Last Updated**: 2025-12-01
**Version**: 1.0
**Maintained By**: Claude Code Documentation Team
