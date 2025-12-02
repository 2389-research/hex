# Tool System in Claude Code

This document provides a complete inventory of tools available in Claude Code, when to use each tool, and how they compose together.

## Tool Architecture

### What is a Tool?

A **tool** is a discrete capability that Claude Code can invoke to interact with the environment:

```
┌────────────────────────────────────────┐
│         Claude Code (LLM)              │
│  - Understands request                 │
│  - Selects appropriate tools           │
│  - Interprets results                  │
└───────────────┬────────────────────────┘
                │
                ▼
┌────────────────────────────────────────┐
│           Tool Invocation              │
│  function_call(                        │
│    name="Read",                        │
│    parameters={"file_path": "..."}     │
│  )                                     │
└───────────────┬────────────────────────┘
                │
                ▼
┌────────────────────────────────────────┐
│         Tool Execution                 │
│  - Reads file from filesystem          │
│  - Returns contents with line numbers  │
└───────────────┬────────────────────────┘
                │
                ▼
┌────────────────────────────────────────┐
│           Result to LLM                │
│  File contents or error message        │
└────────────────────────────────────────┘
```

### Tool Properties

Every tool has:

1. **Name**: Unique identifier (e.g., "Read", "Bash", "Edit")
2. **Parameters**: Input values (file paths, commands, patterns)
3. **Return value**: Output or error
4. **Side effects**: What changes in the environment (if any)

### Tool vs. Function Call

Tools are **not** arbitrary code execution. They are:
- Predefined capabilities
- Sandboxed operations
- Validated inputs
- Structured outputs

You cannot "write a tool" during execution. You can only use available tools.

## Complete Tool Inventory

### File Operations

#### Read
**Purpose**: Retrieve file contents

**Parameters**:
- `file_path` (required): Absolute path to file
- `offset` (optional): Start line number
- `limit` (optional): Max lines to read (default 2000)

**Returns**: File contents with line numbers

**Side effects**: None

**Use when**:
- Need to understand code before editing
- Investigating file structure
- Extracting information

**Example**:
```python
Read(file_path="/Users/harper/project/main.py")
```

---

#### Edit
**Purpose**: Modify existing files via exact string replacement

**Parameters**:
- `file_path` (required): Absolute path to file
- `old_string` (required): Exact text to find
- `new_string` (required): Replacement text
- `replace_all` (optional): Replace all occurrences (default False)

**Returns**: Success confirmation or error

**Side effects**: Modifies file on disk

**Use when**:
- Changing code in existing files
- Renaming variables/functions
- Updating imports

**Constraints**:
- Must Read file first
- old_string must be exact match
- old_string must be unique (unless replace_all=True)

**Example**:
```python
Edit(
    file_path="/Users/harper/project/main.py",
    old_string="def old_name():",
    new_string="def new_name():"
)
```

---

#### Write
**Purpose**: Create new files or completely rewrite existing files

**Parameters**:
- `file_path` (required): Absolute path to file
- `content` (required): Complete file contents

**Returns**: Success confirmation or error

**Side effects**: Creates or overwrites file on disk

**Use when**:
- Creating new files (configs, scripts, docs)
- Complete file rewrites (rare)

**Constraints**:
- Must Read existing files before overwriting
- Prefer Edit for modifications

**Example**:
```python
Write(
    file_path="/Users/harper/project/config.json",
    content='{"debug": true}'
)
```

---

### Search and Discovery

#### Grep
**Purpose**: Search file contents using regex patterns

**Parameters**:
- `pattern` (required): Regex pattern to match
- `path` (optional): Directory or file to search (default: cwd)
- `output_mode` (optional): "files_with_matches", "content", or "count"
- `glob` (optional): File pattern filter (e.g., "*.py")
- `type` (optional): File type filter (e.g., "py", "js")
- `-i` (optional): Case-insensitive search
- `-A`, `-B`, `-C` (optional): Context lines after/before/around
- `multiline` (optional): Enable multiline matching

**Returns**: Matching files, content, or counts

**Side effects**: None

**Use when**:
- Finding where a function is defined
- Searching for error messages
- Locating import statements
- Finding TODO comments

**Example**:
```python
# Find files containing "authenticate"
Grep(
    pattern="authenticate",
    output_mode="files_with_matches",
    type="py"
)

# Show matching lines with context
Grep(
    pattern="def process",
    output_mode="content",
    glob="**/*.py",
    A=2,  # 2 lines after
    B=2   # 2 lines before
)
```

---

#### Glob
**Purpose**: Find files matching patterns

**Parameters**:
- `pattern` (required): Glob pattern (e.g., "**/*.py")
- `path` (optional): Directory to search (default: cwd)

**Returns**: List of matching file paths

**Side effects**: None

**Use when**:
- Finding all files of a type
- Listing files in directory structure
- Before reading multiple related files

**Example**:
```python
# Find all Python files
Glob(pattern="**/*.py")

# Find all test files
Glob(pattern="tests/**/test_*.py")

# Find all JSON configs
Glob(pattern="config/**/*.json")
```

---

### Command Execution

#### Bash
**Purpose**: Execute shell commands

**Parameters**:
- `command` (required): Shell command to execute
- `description` (optional): Brief description of command
- `timeout` (optional): Max execution time in ms (default 120000)
- `run_in_background` (optional): Run asynchronously (default False)

**Returns**: stdout, stderr, exit code

**Side effects**: Depends on command (file creation, process execution, etc.)

**Use when**:
- Running tests
- Git operations
- Building projects
- Installing dependencies
- Running scripts

**Constraints**:
- Working directory resets between calls
- Use absolute paths or chain commands with &&
- Never use --no-verify with git

**Example**:
```python
# Run tests
Bash(command="pytest tests/", description="Run test suite")

# Git status
Bash(command="git status", description="Check git status")

# Chain commands
Bash(command="cd /path && python script.py", description="Run script in directory")
```

---

#### BashOutput
**Purpose**: Retrieve output from background Bash process

**Parameters**:
- `bash_id` (required): ID of background shell
- `filter` (optional): Regex to filter output lines

**Returns**: New output since last check

**Side effects**: None

**Use when**:
- Monitoring long-running processes
- Checking progress of background jobs
- Tailing logs from running services

**Example**:
```python
# Start background process
Bash(command="npm run dev", run_in_background=True)
# Returns: bash_id = "shell_123"

# Later, check output
BashOutput(bash_id="shell_123")
```

---

#### KillShell
**Purpose**: Terminate background shell

**Parameters**:
- `shell_id` (required): ID of shell to kill

**Returns**: Success confirmation

**Side effects**: Terminates process

**Use when**:
- Stopping long-running background jobs
- Cleaning up after testing

**Example**:
```python
KillShell(shell_id="shell_123")
```

---

### Task Management

#### TodoWrite
**Purpose**: Create and manage task lists

**Parameters**:
- `todos` (required): Array of todo objects with:
  - `content`: Description (imperative form)
  - `activeForm`: Present continuous form
  - `status`: "pending", "in_progress", or "completed"

**Returns**: Updated todo list

**Side effects**: Updates task tracking

**Use when**:
- Complex multi-step tasks
- Tracking progress across operations
- User requests explicit task breakdown

**Constraints**:
- Only one task in_progress at a time
- Mark completed immediately after finishing
- Don't use for simple single tasks

**Example**:
```python
TodoWrite(todos=[
    {
        "content": "Fix authentication bug",
        "activeForm": "Fixing authentication bug",
        "status": "completed"
    },
    {
        "content": "Add test coverage",
        "activeForm": "Adding test coverage",
        "status": "in_progress"
    },
    {
        "content": "Update documentation",
        "activeForm": "Updating documentation",
        "status": "pending"
    }
])
```

---

#### Task
**Purpose**: Launch isolated subagent instances for specialized work

**Parameters**:
- `description` (required): Short task description (for logging)
- `prompt` (required): Detailed instructions for the subagent
- `subagent_type` (required): Which agent to use (e.g., "Explore", "Plan", "general-purpose")
- `model` (optional): Override default model (haiku/sonnet/opus)

**Returns**: Subagent's final response as string

**Side effects**: Creates isolated Claude instance, executes task, returns result

**Use when**:
- Need fresh context window (token limit relief)
- Specialized exploration or analysis
- Parallel independent tasks
- Complex research requiring isolation

**Constraints**:
- Subagent has NO access to main conversation context
- Must provide ALL necessary information in prompt parameter
- Context is discarded after completion

**Example**:
```python
Task(
    description="Explore authentication system",
    prompt="""Explore the authentication system in this codebase.

Focus on:
1. How users log in (OAuth, JWT, sessions?)
2. Password storage and validation
3. Authorization middleware
4. Security vulnerabilities

Provide:
- Architecture summary
- Security assessment
- List of auth-related files""",
    subagent_type="Explore"
)
```

**See also**: 05-SUBAGENT-SYSTEM.md, 26-SUBAGENT-DEEP-DIVE.md

---

#### NotebookEdit
**Purpose**: Edit Jupyter notebook (.ipynb) cells

**Parameters**:
- `notebook_path` (required): Absolute path to .ipynb file
- `new_source` (required): New cell content
- `cell_id` (optional): Cell ID to edit
- `cell_type` (optional): "code" or "markdown"
- `edit_mode` (optional): "replace", "insert", or "delete"

**Returns**: Success confirmation

**Side effects**: Modifies notebook cell on disk

**Use when**:
- Editing data science notebooks
- Updating analysis code
- Modifying markdown documentation in notebooks
- Adding/removing notebook cells

**Constraints**:
- Must Read notebook first
- Cell IDs are specific to the notebook

**Example**:
```python
# Replace a code cell
NotebookEdit(
    notebook_path="/Users/harper/analysis.ipynb",
    cell_id="abc-123",
    new_source="import pandas as pd\ndf = pd.read_csv('data.csv')",
    cell_type="code",
    edit_mode="replace"
)

# Insert new markdown cell
NotebookEdit(
    notebook_path="/Users/harper/analysis.ipynb",
    new_source="# Data Analysis Results",
    cell_type="markdown",
    edit_mode="insert"
)
```

---

#### AskUserQuestion
**Purpose**: Ask user questions during execution with structured options

**Parameters**:
- `questions` (required): Array of question objects (1-4 questions)
  - `question`: The question text
  - `header`: Short label (max 12 chars)
  - `options`: Array of 2-4 options with label and description
  - `multiSelect`: Boolean (allow multiple selections)

**Returns**: User's selected answers

**Side effects**: Pauses execution for user input

**Use when**:
- Need user preference or decision
- Multiple valid approaches exist
- Clarifying ambiguous requirements
- Offering implementation choices

**Constraints**:
- Maximum 4 questions per call
- Each question needs 2-4 options
- "Other" option provided automatically

**Example**:
```python
AskUserQuestion(questions=[
    {
        "question": "Which authentication method should we implement?",
        "header": "Auth method",
        "multiSelect": False,
        "options": [
            {
                "label": "OAuth2",
                "description": "Industry standard, supports multiple providers"
            },
            {
                "label": "JWT",
                "description": "Stateless tokens, good for APIs"
            },
            {
                "label": "Sessions",
                "description": "Traditional server-side sessions"
            }
        ]
    }
])
```

**See also**: 14-DECISION-FRAMEWORK.md for when to ask vs proceed autonomously

---

#### ExitPlanMode
**Purpose**: Exit plan mode after creating implementation plan

**Parameters**:
- `plan` (required): The implementation plan (markdown format)

**Returns**: Presents plan to user and exits plan mode

**Side effects**: Transitions from planning to ready-to-implement state

**Use when**:
- Finished creating detailed implementation plan
- Ready to present plan for approval
- Working in plan mode (/superpowers:write-plan)

**Constraints**:
- Only available in plan mode
- Plan should be comprehensive and actionable

**Example**:
```python
ExitPlanMode(plan="""
## Implementation Plan: User Authentication

### Phase 1: Database Schema
1. Create users table with email, password_hash, created_at
2. Add index on email column
3. Migration file: db/migrations/001_create_users.sql

### Phase 2: Authentication Logic
1. Implement password hashing (bcrypt, 12 rounds)
2. Create login endpoint POST /api/auth/login
3. Generate JWT tokens (expiry: 24h)

### Phase 3: Testing
1. Unit tests for password hashing
2. Integration tests for login flow
3. Security tests for common attacks
""")
```

**See also**: 28-PLAN-MODE-WORKFLOW.md

---

### Advanced Tools

#### Skill
**Purpose**: Invoke specialized skills (subagent patterns)

**Parameters**:
- `command` (required): Skill name

**Returns**: Skill-specific behavior

**Side effects**: Varies by skill

**Use when**:
- Following structured workflows (TDD, debugging, etc.)
- Need specialized knowledge (e.g., test-driven development)
- Complex multi-phase tasks

**Available skills**:
- `superpowers:test-driven-development`
- `superpowers:systematic-debugging`
- `superpowers:brainstorming`
- `superpowers:subagent-driven-development`
- Many others (see Skill inventory)

**Example**:
```python
Skill(command="superpowers:test-driven-development")
```

---

#### SlashCommand
**Purpose**: Execute custom project commands

**Parameters**:
- `command` (required): Command with arguments

**Returns**: Command-specific output

**Side effects**: Varies by command

**Use when**:
- Project-specific workflows
- Repeated operations
- Custom automation

**Example**:
```python
SlashCommand(command="/superpowers:write-plan")
```

---

#### WebFetch
**Purpose**: Fetch and process web content

**Parameters**:
- `url` (required): URL to fetch
- `prompt` (required): What to extract from page

**Returns**: Extracted information

**Side effects**: None (read-only)

**Use when**:
- Looking up documentation
- Checking library versions
- Researching APIs

**Example**:
```python
WebFetch(
    url="https://docs.python.org/3/library/json.html",
    prompt="Explain how to parse JSON with error handling"
)
```

---

#### WebSearch
**Purpose**: Search the web for current information

**Parameters**:
- `query` (required): Search query
- `allowed_domains` (optional): Whitelist of domains
- `blocked_domains` (optional): Blacklist of domains

**Returns**: Search results with excerpts

**Side effects**: None

**Use when**:
- Finding recent information beyond knowledge cutoff
- Looking up error messages
- Checking current library versions

**Example**:
```python
WebSearch(query="FastAPI async database connection 2025")
```

---

#### ListMcpResourcesTool
**Purpose**: List available resources from MCP servers

**Parameters**:
- `server` (optional): Specific MCP server name to query

**Returns**: Array of resources with metadata (name, URI, description, server)

**Side effects**: None (read-only)

**Use when**:
- Discovering available MCP resources
- Browsing capabilities of MCP servers
- Finding resource URIs for ReadMcpResourceTool

**Example**:
```python
# List all resources from all servers
ListMcpResourcesTool()

# List resources from specific server
ListMcpResourcesTool(server="filesystem")
```

**See also**: 09-MCP-SERVERS.md

---

#### ReadMcpResourceTool
**Purpose**: Read specific resource from MCP server

**Parameters**:
- `server` (required): MCP server name
- `uri` (required): Resource URI

**Returns**: Resource content

**Side effects**: None (read-only)

**Use when**:
- Accessing MCP-provided data
- Reading resources discovered via ListMcpResourcesTool
- Integrating with external systems via MCP

**Example**:
```python
ReadMcpResourceTool(
    server="filesystem",
    uri="file:///Users/harper/project/README.md"
)
```

**See also**: 09-MCP-SERVERS.md

---

### MCP (Model Context Protocol) Tools

MCP servers provide additional tools beyond the resource tools above. Available tools depend on configured servers.

#### Examples from Common MCP Servers

**Chronicle (time tracking)**:
```python
mcp__chronicle__add_entry(
    message="Implemented user authentication",
    tags=["feature", "auth"]
)
```

**Playwright (browser automation)**:
```python
mcp__playwright__browser_navigate(url="https://example.com")
mcp__playwright__browser_click(element="Login button", ref="btn-login")
```

**Toki (todo management)**:
```python
mcp__toki__add_todo(
    description="Fix CORS issue",
    priority="high",
    tags=["bug", "backend"]
)
```

---

## Tool Selection Logic

### Decision Tree

```
Need to interact with files?
├─ Read existing file? → Read
├─ Modify existing file? → Edit (after Read)
├─ Create new file? → Write
└─ Find files? → Glob

Need to search content?
├─ Know file location? → Read + manual search
├─ Search across files? → Grep
└─ Find file by name? → Glob

Need to run commands?
├─ Quick command? → Bash
├─ Long-running? → Bash(run_in_background=True)
└─ Monitor background? → BashOutput

Need to track progress?
├─ Simple task? → Just do it
├─ Multi-step? → TodoWrite
└─ Complex workflow? → Skill

Need external info?
├─ Web page? → WebFetch
├─ Search? → WebSearch
└─ Documentation? → WebFetch or WebSearch

Need specialized behavior?
├─ Project command? → SlashCommand
├─ Structured workflow? → Skill
└─ MCP integration? → mcp__* tools
```

### Tool Combination Patterns

#### Pattern 1: Search → Read → Edit

```python
# Find files
Grep(pattern="def authenticate", output_mode="files_with_matches")
# → Returns: ["auth.py", "api.py"]

# Read relevant files
Read("auth.py")
Read("api.py")

# Make changes
Edit("auth.py", old_string="...", new_string="...")
```

#### Pattern 2: Read → Edit → Test

```python
# Understand current state
Read("module.py")

# Make change
Edit("module.py", old_string="...", new_string="...")

# Verify
Bash("pytest tests/test_module.py")
```

#### Pattern 3: Glob → Read Multiple → Analyze

```python
# Find all test files
Glob(pattern="tests/**/*.py")
# → Returns: list of test files

# Read them in parallel
Read("tests/test_a.py")
Read("tests/test_b.py")
Read("tests/test_c.py")

# Analyze patterns and report
```

#### Pattern 4: Grep → Read with Context → Edit

```python
# Find error location
Grep(pattern="TODO.*FIXME", output_mode="content", C=3)
# → Shows TODOs with 3 lines context

# Read full file
Read("flagged_file.py")

# Fix the TODO
Edit("flagged_file.py", old_string="# TODO FIXME", new_string="# Fixed")
```

## Parallel vs Sequential Execution

### Parallel Execution

**When**: Operations are independent (don't depend on each other's results)

**How**: Make all tool calls in single response

**Example**:
```python
# All independent - execute in parallel
Read("/path/to/file1.py")
Read("/path/to/file2.py")
Read("/path/to/file3.py")
Grep(pattern="error", glob="*.log")
Bash("git status")
```

**All five tools execute simultaneously.**

### Sequential Execution

**When**: Later operations depend on earlier results

**How**: Make tool calls in separate responses

**Example**:
```python
# Response 1: Find files
Glob(pattern="**/*.py")
# → Results: ["a.py", "b.py", "c.py"]

# Response 2: Read found files (depends on Glob results)
Read("a.py")
Read("b.py")
Read("c.py")

# Response 3: Edit based on what was read (depends on Read results)
Edit("a.py", ...)
Edit("b.py", ...)
```

### Chaining in Bash

Within a single Bash command, use `&&` for sequential:

```bash
# Sequential (second runs only if first succeeds)
Bash("cd /path && pytest")

# Parallel (both run regardless)
Bash("pytest tests/unit & pytest tests/integration")
```

## Tool Constraints and Permissions

### File Operation Constraints

#### Read
- ✓ Can read any file
- ✓ Can read binary files (images, PDFs)
- ✗ Cannot read directories (use Bash ls instead)
- ✗ Cannot read files that don't exist

#### Edit
- ✓ Can modify any existing file
- ✗ Cannot edit without reading first
- ✗ Cannot edit non-existent files (use Write)
- ✗ Cannot use relative paths

#### Write
- ✓ Can create new files
- ✓ Can overwrite existing files (after reading)
- ✗ Cannot write without reading existing files first
- ✗ Cannot use relative paths

### Bash Constraints

#### Always Allowed
- Running tests
- Git operations (status, diff, log, commit, push)
- Building projects
- Installing dependencies
- Reading files (cat, head, tail)

#### Restricted
- ✗ `git commit --no-verify` (never bypass hooks)
- ✗ `git push --force` (unless explicitly requested)
- ✗ Destructive operations without confirmation
- ✗ Commands that require interactive input (use tmux pattern)

#### Working Directory
- Resets between Bash calls
- Use absolute paths or `cd /path && command`
- Cannot rely on previous `cd` commands

### Search Constraints

#### Grep
- ✓ Search any files
- ✓ Regex patterns supported
- ✓ Case-sensitive or insensitive
- ✗ Limited to text files (binary files skipped)
- ✗ Very large result sets may be truncated

#### Glob
- ✓ Find files by pattern
- ✓ Recursive search
- ✗ Doesn't search content (use Grep)
- ✗ Returns paths only, not contents

### Permission Levels

Claude Code follows a permission framework:

**🟢 Green (Autonomous)**:
- Fix failing tests
- Correct typos
- Add missing imports
- Read any file
- Run tests

**🟡 Yellow (Collaborative - propose first)**:
- Multi-file refactors
- New features
- API changes
- Schema migrations

**🔴 Red (Always ask)**:
- Rewriting working code
- Security changes
- Data loss risks
- Force pushing to main/master

## Tool Composition Examples

### Example 1: Feature Implementation

```python
# Task: Add logging to a module

# Step 1: Find the module (Search)
Grep(pattern="def process_data", output_mode="files_with_matches")

# Step 2: Understand current state (Read)
Read("/path/to/processor.py")

# Step 3: Add import (Edit)
Edit(
    file_path="/path/to/processor.py",
    old_string="import json",
    new_string="import json\nimport logging"
)

# Step 4: Add logging calls (Edit)
Edit(
    file_path="/path/to/processor.py",
    old_string="def process_data(data):\n    result = transform(data)",
    new_string="def process_data(data):\n    logging.info('Processing data')\n    result = transform(data)"
)

# Step 5: Verify (Bash)
Bash("python -m pytest tests/test_processor.py")

# Step 6: Commit (Bash)
Bash('git add -A && git commit -m "feat: add logging to processor"')
```

### Example 2: Bug Investigation

```python
# Task: Figure out why tests are failing

# Step 1: Run tests to see error (Bash)
Bash("pytest tests/test_api.py -v")
# Output shows: "AssertionError: expected 200, got 500"

# Step 2: Find relevant code (Grep)
Grep(pattern="def test_api", output_mode="content", glob="tests/*.py")

# Step 3: Read test and implementation (Read)
Read("tests/test_api.py")
Read("app/api.py")

# Step 4: Find root cause (analysis)
# Test expects 200 OK, but endpoint returns 500
# Code shows: missing null check before database query

# Step 5: Fix (Edit)
Edit(
    file_path="app/api.py",
    old_string="user = db.get_user(id)\nuser.update(data)",
    new_string="user = db.get_user(id)\nif user is None:\n    return 404\nuser.update(data)"
)

# Step 6: Verify fix (Bash)
Bash("pytest tests/test_api.py -v")
# Tests now pass
```

### Example 3: Codebase Exploration

```python
# Task: Understand how authentication works

# Step 1: Find auth-related files (Glob)
Glob(pattern="**/*auth*.py")

# Step 2: Search for key patterns (Grep)
Grep(pattern="@requires_auth|authenticate|login", output_mode="content")

# Step 3: Read core auth modules (Read)
Read("app/auth/middleware.py")
Read("app/auth/handlers.py")
Read("app/models/user.py")

# Step 4: Trace flow (analysis + more reads as needed)
Read("app/auth/tokens.py")

# Step 5: Document findings (Write)
Write(
    file_path="docs/AUTH_FLOW.md",
    content="""# Authentication Flow

1. Client sends credentials to /login
2. handlers.py validates against database
3. tokens.py generates JWT
4. middleware.py validates JWT on protected routes
"""
)
```

## Tool Selection Best Practices

### 1. Use the Right Tool for the Job

```python
# ❌ WRONG: Using Bash when Read is better
Bash("cat /path/to/file.py")

# ✓ CORRECT: Use Read for file content
Read("/path/to/file.py")
```

### 2. Combine Tools Effectively

```python
# ❌ WRONG: Reading every file to find a function
Read("file1.py")
Read("file2.py")
Read("file3.py")  # ... (searching for function)

# ✓ CORRECT: Use Grep first to find it
Grep(pattern="def target_function", output_mode="files_with_matches")
# Then read only the relevant file
```

### 3. Minimize Tool Calls

```python
# ❌ WRONG: Sequential reads when parallel works
Read("file1.py")  # Response 1
Read("file2.py")  # Response 2
Read("file3.py")  # Response 3

# ✓ CORRECT: Parallel reads in one response
Read("file1.py")
Read("file2.py")
Read("file3.py")
```

### 4. Verify After Changes

```python
# ✓ CORRECT: Always test after editing
Edit("module.py", ...)
Bash("pytest")  # Verify

# ❌ WRONG: Edit without verification
Edit("module.py", ...)
# (no verification)
```

### 5. Read Before Edit (Always)

```python
# ✓ CORRECT:
Read("file.py")
Edit("file.py", ...)

# ❌ WRONG:
Edit("file.py", ...)  # Error: must read first
```

## Tool Limitations

### Context Window

**Problem**: Can't read infinite files

**Solution**:
- Use Grep/Glob to narrow down
- Read incrementally
- Focus on changed files

### Execution Time

**Problem**: Some commands take too long

**Solution**:
- Use background execution
- Set appropriate timeouts
- Monitor with BashOutput

### Working Directory

**Problem**: Bash resets cwd between calls

**Solution**:
- Use absolute paths
- Chain commands: `cd /path && command`
- Store paths in variables

### No Interactive Commands

**Problem**: Can't run vim, interactive git, etc.

**Solution**:
- Use non-interactive alternatives
- Use tmux pattern (via MCP or skill)
- Automate with flags (git commit -m "...")

## Summary Table

| Tool | Primary Use | Read First? | Side Effects? | Parallel? |
|------|-------------|-------------|---------------|-----------|
| **Read** | View files | N/A | No | Yes |
| **Edit** | Modify files | Yes | Yes (file changes) | Yes |
| **Write** | Create/rewrite files | If exists | Yes (file changes) | Yes |
| **Grep** | Search content | No | No | Yes |
| **Glob** | Find files | No | No | Yes |
| **Bash** | Run commands | No | Varies | Yes* |
| **TodoWrite** | Track tasks | No | Yes (task state) | N/A |
| **Skill** | Structured workflows | No | Varies | No |
| **WebFetch** | Fetch web content | No | No | Yes |
| **WebSearch** | Search web | No | No | Yes |

*Bash can run multiple commands in parallel using `&` or separate calls

---

*Next: [04-BASH-AND-COMMAND-EXECUTION.md](./04-BASH-AND-COMMAND-EXECUTION.md) for deep dive on command execution.*
