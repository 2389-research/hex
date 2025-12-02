# Claude Code Overview

## What is Claude Code?

Claude Code is Anthropic's official command-line interface that embeds Claude AI as a terminal-based development assistant. It's not a chatbot with coding capabilities - it's an **agentic system** designed for direct integration into software development workflows.

### Core Identity

```
Claude Code = AI Assistant + Unix Tools + Agentic Behavior
```

- **Terminal-native**: Runs in your shell, operates on your filesystem
- **Tool-based**: Uses discrete capabilities (Read, Edit, Bash, etc.) rather than freeform text generation
- **Agentic**: Makes decisions, executes actions, verifies results
- **Context-aware**: Understands git state, file structure, project context

## Architecture

### High-Level Design

```
┌─────────────────────────────────────────────────┐
│                   User Input                     │
│         (natural language request)               │
└────────────────┬────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────┐
│              Claude Code Core                    │
│  ┌──────────────────────────────────────────┐   │
│  │     Language Model (Claude)              │   │
│  │  - Task understanding                    │   │
│  │  - Tool selection                        │   │
│  │  - Decision making                       │   │
│  └──────────────────────────────────────────┘   │
└────────────────┬────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────┐
│                 Tool Layer                       │
│  ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐  │
│  │ Read │ │ Edit │ │Write │ │ Bash │ │ Grep │  │
│  └──────┘ └──────┘ └──────┘ └──────┘ └──────┘  │
│  ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐           │
│  │ Glob │ │ Task │ │ Skill│ │ MCP  │  ...      │
│  └──────┘ └──────┘ └──────┘ └──────┘           │
└────────────────┬────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────┐
│            Your Development Environment          │
│  - Filesystem (read/write)                       │
│  - Shell (command execution)                     │
│  - Git (version control)                         │
│  - Running processes (tests, builds, servers)    │
└─────────────────────────────────────────────────┘
```

### Execution Model

1. **Receive Request**: User provides natural language instruction
2. **Plan Actions**: Claude determines which tools to use and in what order
3. **Execute Tools**: Tools interact with filesystem, shell, or external systems
4. **Receive Results**: Tools return output (file contents, command results, errors)
5. **Iterate or Complete**: Continue with more tools or provide final response

### Key Constraint: Stateless Between Calls

Each Claude Code invocation starts fresh:
- No memory of previous conversations (unless using project memory/MCP)
- Must re-read files to understand state
- Git status provides critical context
- Working directory resets between Bash calls

Exception: Persistent shells and background processes maintain state.

## Core Capabilities

### 1. Feature Development

**What it does**: Implements new functionality end-to-end

**Typical workflow**:
```
User: "Add a --verbose flag to the CLI that shows debug output"

Claude Code:
1. Read CLI argument parsing code
2. Read logger configuration
3. Edit argument parser to add --verbose flag
4. Edit logger setup to respect flag
5. Run tests to verify
6. Commit changes with descriptive message
```

**Key strength**: Understands full context from request to implementation to verification.

### 2. Debugging and Investigation

**What it does**: Diagnoses problems and proposes fixes

**Typical workflow**:
```
User: "Tests are failing in test_parser.py"

Claude Code:
1. Run tests to see actual error
2. Read test file
3. Read implementation being tested
4. Identify mismatch (e.g., expected vs actual output)
5. Edit implementation to fix root cause
6. Re-run tests to verify fix
```

**Key strength**: Systematic investigation rather than guessing.

### 3. Codebase Exploration

**What it does**: Navigates unfamiliar codebases to understand structure

**Typical workflow**:
```
User: "How does authentication work in this app?"

Claude Code:
1. Grep for "auth" patterns
2. Glob for auth-related files
3. Read key authentication modules
4. Trace data flow through imports
5. Summarize findings with file references
```

**Key strength**: Fast pattern matching + deep understanding of found code.

### 4. Code Review and Refactoring

**What it does**: Analyzes code quality and improves structure

**Typical workflow**:
```
User: "This module has too many responsibilities, can you refactor it?"

Claude Code:
1. Read target module
2. Identify distinct concerns
3. Create new focused modules (Write)
4. Move code to appropriate modules (Edit)
5. Update imports across codebase
6. Run tests to ensure no breakage
```

**Key strength**: Cross-file changes with consistency and testing.

### 5. Documentation and Explanation

**What it does**: Generates or improves documentation

**Typical workflow**:
```
User: "Document the API endpoints in this FastAPI app"

Claude Code:
1. Read route definitions
2. Read request/response models
3. Generate OpenAPI-compatible docstrings
4. Edit files to add documentation
5. Verify docs render correctly
```

**Key strength**: Understands code deeply enough to explain accurately.

## Unix Philosophy Approach

Claude Code follows Unix principles:

### Do One Thing Well

Each tool has a focused purpose:
- **Read**: Only reads files
- **Edit**: Only modifies existing files
- **Write**: Only creates new files or rewrites
- **Bash**: Only executes commands
- **Grep**: Only searches content
- **Glob**: Only finds files by pattern

No tool tries to do everything.

### Tools Compose Together

Complex workflows emerge from simple tool combinations:

```python
# Example: Find all Python files importing 'logging',
# read them, and add a new log level

1. Glob("**/*.py")           # Find Python files
2. Grep("import logging")    # Filter to logging users
3. Read(each_file)           # Understand current state
4. Edit(each_file)           # Add new log level
5. Bash("pytest")            # Verify changes work
```

### Text-Based Interfaces

All tool inputs and outputs are text or structured data:
- File contents as plain text with line numbers
- Command output as stdout/stderr
- Search results as file paths and matching lines
- Edit operations as string-to-string transformations

This makes everything inspectable and debuggable.

### Small, Focused Operations

Rather than "rewrite this entire module", Claude Code prefers:
1. Read module
2. Edit function A
3. Edit function B
4. Edit imports
5. Run tests

Each step is small, verifiable, and reversible.

## Installation and Setup

### Installation Methods

#### Official Installation (Recommended)
```bash
# macOS/Linux
curl -fsSL https://claude.ai/install.sh | bash

# Or via package manager
brew install anthropic/claude/claude-code
```

#### Configuration
```bash
# Initialize in project
cd your-project/
claude-code init

# Creates .claude/ directory with:
# - Project guidance (policy doc)
# - commands/ (custom slash commands)
# - docs/ (project documentation)
```

### Authentication

```bash
# Set API key
export ANTHROPIC_API_KEY="your-key-here"

# Or configure via CLI
claude-code auth login
```

### Project Setup

**Critical artifact**: Project guidance stored under `.claude/`

This content contains project-specific instructions that Claude Code reads on every invocation:

```markdown
# Project Instructions

## Technology Stack
- Python 3.12 with uv package manager
- FastAPI for web framework
- PostgreSQL database

## Testing Requirements
- Must run tests before committing
- Coverage must stay above 80%

## Code Style
- Use ruff for linting
- Follow PEP 8
- Max line length: 100
```

Claude Code treats this guidance as **higher priority than default behavior**.

## Agentic Behavior

Claude Code is **agentic**, meaning it:

### Makes Decisions Autonomously

When asked "fix the broken tests", Claude Code:
1. Decides to run tests first (see actual error)
2. Decides which files to read based on error
3. Decides whether to fix code or test
4. Decides when fix is complete

No hand-holding required.

### Verifies Its Own Work

After making changes:
```
1. Edit file
2. Run tests      ← Verification
3. Check output   ← Parsing results
4. Fix if needed  ← Self-correction
5. Re-verify      ← Iterative improvement
```

### Asks When Uncertain

Claude Code follows a decision framework:

**Autonomous (no permission)**:
- Fix failing tests
- Correct typos
- Add missing imports
- Refactor single files

**Collaborative (propose first)**:
- Multi-file changes
- New features
- API changes
- Schema migrations

**Always ask**:
- Rewriting working code from scratch
- Security changes
- Data loss risks

This prevents both over-caution (asking about every semicolon) and recklessness (breaking production).

## Working with Context

### Context Window

Claude Code operates within a context window (token limit). For Claude Sonnet:
- ~200k tokens input
- Enough for dozens of files
- Must be strategic about what to read

### Context Management Strategies

**1. Search before reading**:
```
Bad:  Read all 500 Python files
Good: Grep for error message → Read 3 relevant files
```

**2. Read incrementally**:
```
Bad:  Read entire codebase upfront
Good: Read as needed based on findings
```

**3. Use git context**:
Git status shows:
- Modified files (likely relevant)
- Untracked files (recently created)
- Current branch (feature context)

Claude Code automatically gets git status at conversation start.

**4. Focus on essentials**:
```
Essential:
- Files being modified
- Tests verifying behavior
- Interfaces/contracts

Optional:
- Implementation details of dependencies
- Historical code not being changed
- Tangentially related modules
```

## Model Details

Claude Code is powered by **Claude Sonnet 4.5** (as of December 2025):
- Model ID: `claude-sonnet-4-5-20250929`
- Knowledge cutoff: January 2025
- Optimized for coding tasks
- Strong at reasoning and tool use

### What This Means

**Strengths**:
- Complex multi-step reasoning
- Accurate code generation
- Deep language understanding
- Reliable tool selection

**Limitations**:
- No internet access (unless via MCP)
- Knowledge cutoff (may not know newest libraries)
- Token limits (can't read infinite files)
- No inherent memory between conversations

## Comparison to Other Tools

### vs. GitHub Copilot

| Aspect | Claude Code | Copilot |
|--------|-------------|---------|
| Scope | Full tasks (multi-file) | Line/block completion |
| Execution | Runs commands | Suggests code only |
| Testing | Runs and verifies tests | No execution |
| Autonomy | Agentic (decides steps) | Reactive (suggests) |
| Interface | Terminal | IDE integration |

### vs. ChatGPT Code Interpreter

| Aspect | Claude Code | Code Interpreter |
|--------|-------------|------------------|
| Environment | Your local machine | Sandboxed cloud |
| Tools | Full shell access | Limited sandbox |
| Files | Your actual project | Uploaded files only |
| Persistence | Changes persist | Session-only |
| Integration | Git, tests, builds | Isolated |

### vs. Cursor/Cody

| Aspect | Claude Code | Cursor/Cody |
|--------|-------------|-------------|
| Interface | CLI/terminal | IDE |
| Model | Claude Sonnet | Various |
| Tool Use | Rich tool set | Limited |
| Workflow | Command-driven | Editor-driven |
| Power Users | Terminal lovers | IDE lovers |

## Philosophy and Design Principles

### 1. Read Before Edit (Always)

Claude Code **must** read a file before editing it. This is not optional.

**Why**: Prevents blind changes based on assumptions.

**How enforced**: Edit tool errors if file wasn't read in current conversation.

### 2. Verify Before Commit

Never commit without verification:
```
1. Make changes
2. Run tests        ← Verification
3. Check output     ← Parse results
4. Only then commit ← Safe to proceed
```

**Why**: Broken commits pollute git history and break CI.

### 3. Test Before Complete

When a task involves code changes:
```
1. Implement feature
2. Run tests
3. Fix failures
4. Report success only after tests pass
```

**Why**: "Working" means "tested and verified", not "looks right".

### 4. Ask When Uncertain

If the correct approach isn't obvious:
```
User: "The API is slow"

Claude Code: "I can investigate the slowness. Would you like me to:
1. Profile the code to find bottlenecks
2. Check database query performance
3. Review caching strategy
4. All of the above"
```

**Why**: User knows context AI doesn't have.

### 5. Explain Complex Changes

For non-trivial edits:
```
Claude Code: "I'm refactoring the authentication module:
1. Extracting token validation to separate function
2. Adding type hints for clarity
3. Caching decoded tokens to avoid repeated work

This should improve performance and maintainability."
```

**Why**: User should understand what's changing and why.

## Typical Session Flow

### Example: Implement a Feature

```
User: "Add a --json flag to output results as JSON instead of text"

Claude Code workflow:

1. UNDERSTAND CONTEXT
   - Bash: git status (see current state)
   - Glob: *.py (find Python files)
   - Grep: "argparse" or "click" (find CLI framework)

2. READ CURRENT STATE
   - Read: main.py (see current CLI setup)
   - Read: output.py (see current output format)

3. PLAN CHANGES
   - Add --json argument to parser
   - Modify output function to support JSON
   - Update tests

4. IMPLEMENT
   - Edit: main.py (add --json flag)
   - Edit: output.py (add JSON formatting)
   - Edit: test_output.py (add JSON test cases)

5. VERIFY
   - Bash: pytest
   - Check output: all tests pass

6. COMMIT
   - Bash: git add -A
   - Bash: git commit -m "feat: add --json flag for JSON output"

7. REPORT
   "✓ Added --json flag
    ✓ Updated output formatting
    ✓ All tests passing
    ✓ Changes committed

   Usage: python main.py --json"
```

### Example: Debug a Failure

```
User: "The login endpoint returns 500 error"

Claude Code workflow:

1. REPRODUCE
   - Bash: curl -X POST /api/login (see actual error)
   - Note: Returns "Internal Server Error"

2. INVESTIGATE LOGS
   - Bash: tail -n 50 app.log
   - Find: "AttributeError: 'NoneType' object has no attribute 'verify'"

3. TRACE ERROR
   - Grep: "def login" (find login handler)
   - Read: routes/auth.py
   - Identify: Line 42 calls user.verify() but user can be None

4. UNDERSTAND ROOT CAUSE
   - Read: models/user.py
   - See: get_user() returns None if not found
   - Confirm: No null check before user.verify()

5. FIX
   - Edit: routes/auth.py
   - Add: Check if user is None, return 401 if so

6. VERIFY
   - Bash: pytest tests/test_auth.py
   - Bash: curl -X POST /api/login (manual test)
   - Confirm: Returns 401 Unauthorized (correct behavior)

7. REPORT
   "Fixed login 500 error:

   Root cause: Missing null check after user lookup
   Fix: Return 401 Unauthorized when user not found
   Verified: Tests passing, manual test returns proper 401"
```

## Success Criteria

Claude Code is working well when:

1. **Changes are correct**: Code works as intended
2. **Tests pass**: Automated verification succeeds
3. **Commits are clean**: No test failures in history
4. **User understands**: Changes are explained clearly
5. **Workflow is smooth**: Minimal back-and-forth needed

Claude Code is struggling when:

1. **Making blind edits**: Editing without reading first
2. **Breaking tests**: Committing failing code
3. **Over-asking**: Requesting permission for trivial changes
4. **Under-asking**: Making risky changes without discussion
5. **Context overload**: Reading unnecessary files

## Next Steps

- **[02-FILE-OPERATIONS.md](./02-FILE-OPERATIONS.md)**: Deep dive into Read, Edit, Write tools
- **[03-TOOL-SYSTEM.md](./03-TOOL-SYSTEM.md)**: Complete tool inventory and selection logic
- **[04-BASH-AND-COMMAND-EXECUTION.md](./04-BASH-AND-COMMAND-EXECUTION.md)**: Command execution patterns

---

*This document describes what Claude Code is and how it works at a high level. For specific tool usage, see the detailed guides.*
