# Bash and Command Execution in Claude Code

This document explains how Claude Code executes shell commands, manages persistent sessions, handles git operations, and chains commands effectively.

## The Bash Tool

### Purpose

The Bash tool executes shell commands in your development environment. It's how Claude Code:
- Runs tests
- Executes git operations
- Builds projects
- Installs dependencies
- Manages processes
- Performs system operations

### Architecture

```
┌─────────────────────────────────────────┐
│      Claude Code Issues Command         │
│  Bash(command="pytest tests/")          │
└───────────────┬─────────────────────────┘
                │
                ▼
┌─────────────────────────────────────────┐
│        Shell Environment                 │
│  - Uses your actual shell (bash/zsh)    │
│  - Your PATH and environment vars       │
│  - Your installed tools                 │
│  - Your filesystem permissions          │
└───────────────┬─────────────────────────┘
                │
                ▼
┌─────────────────────────────────────────┐
│         Command Execution               │
│  - Runs command to completion           │
│  - Captures stdout/stderr               │
│  - Returns exit code                    │
└───────────────┬─────────────────────────┘
                │
                ▼
┌─────────────────────────────────────────┐
│      Results Back to Claude Code        │
│  {                                       │
│    stdout: "...",                        │
│    stderr: "...",                        │
│    exit_code: 0                          │
│  }                                       │
└─────────────────────────────────────────┘
```

### Bash Tool Signature

```python
Bash(
    command: str,               # Shell command to execute (required)
    description: str = None,    # Brief description (optional but recommended)
    timeout: int = 120000,      # Timeout in milliseconds (optional, default 2min)
    run_in_background: bool = False  # Run asynchronously (optional)
)
```

### Return Value

```python
{
    "stdout": "command output",
    "stderr": "error output",
    "exit_code": 0  # 0 = success, non-zero = failure
}
```

## Basic Command Execution

### Simple Commands

```python
# Run tests
Bash(
    command="pytest tests/",
    description="Run test suite"
)

# Check git status
Bash(
    command="git status",
    description="Check repository status"
)

# List files
Bash(
    command="ls -la src/",
    description="List source files"
)
```

### Reading Output

**Always parse the output** to verify success:

```python
# Run command
result = Bash("pytest tests/")

# Check exit code
if result.exit_code == 0:
    # Success
    print("Tests passed")
else:
    # Failure
    print(f"Tests failed: {result.stderr}")
```

**Critical**: Don't assume success. Always check exit_code or parse output.

### Command Descriptions

The `description` parameter helps with clarity:

```python
# ✓ GOOD: Clear description
Bash(
    command="npm run build",
    description="Build production bundle"
)

# ❌ WEAK: No description
Bash(command="npm run build")
```

Descriptions should be:
- 5-10 words
- Action-oriented ("Run tests", "Build project")
- Clear about what's happening

## Working Directory Behavior

### Critical Rule: Directory Resets Between Calls

**Each Bash call starts in the working directory**. The directory does NOT persist:

```python
# ❌ WRONG: Second call won't be in /tmp
Bash("cd /tmp")
Bash("ls")  # Runs in working directory, NOT /tmp

# ✓ CORRECT: Chain with &&
Bash("cd /tmp && ls")

# ✓ CORRECT: Use absolute paths
Bash("ls /tmp")
```

### Handling Directories

#### Option 1: Chain Commands

```bash
cd /path && command
```

Example:
```python
Bash("cd /Users/harper/project && pytest")
```

#### Option 2: Use Absolute Paths

```python
Bash("pytest /Users/harper/project/tests/")
```

#### Option 3: Multiple Commands in One Call

```python
Bash("""
cd /Users/harper/project
source venv/bin/activate
pytest tests/
""")
```

### Getting Current Directory

```python
result = Bash("pwd")
current_dir = result.stdout.strip()
# Use current_dir in subsequent operations
```

## Command Chaining

### Sequential Execution (&&)

**Use `&&`** to run commands sequentially, where each depends on the previous succeeding:

```bash
command1 && command2 && command3
```

**Behavior**:
- command2 runs only if command1 succeeds (exit 0)
- command3 runs only if command2 succeeds
- If any fails, chain stops

**Example**:
```python
Bash(
    command="git add -A && git commit -m 'feat: add feature' && git push",
    description="Stage, commit, and push changes"
)
```

**If commit fails** (e.g., pre-commit hook), push won't run.

### Parallel Execution (&)

**Use `&`** to run commands in parallel:

```bash
command1 & command2 & command3
```

**Behavior**:
- All commands start simultaneously
- Don't wait for each other
- Useful for independent operations

**Example**:
```python
Bash(
    command="pytest tests/unit & pytest tests/integration & pytest tests/e2e",
    description="Run all test suites in parallel"
)
```

### Unconditional Sequential (;)

**Use `;`** to run commands sequentially regardless of success:

```bash
command1 ; command2 ; command3
```

**Behavior**:
- command2 runs even if command1 fails
- command3 runs even if command2 fails
- Use when failures are acceptable

**Example**:
```python
Bash(
    command="make clean ; make build",
    description="Clean (if possible) then build"
)
```

### OR Logic (||)

**Use `||`** for fallback commands:

```bash
command1 || command2
```

**Behavior**:
- command2 runs only if command1 fails
- Useful for defaults or error handling

**Example**:
```python
Bash("python3 script.py || python script.py")  # Try python3, fallback to python
```

### Complex Chains

Combine operators for sophisticated logic:

```python
Bash("""
    cd /project && \
    source venv/bin/activate && \
    (pytest tests/ || echo "Tests failed but continuing") && \
    python deploy.py
""")
```

## Persistent Shell Sessions

### Background Execution

Run long-running commands without blocking:

```python
# Start background process
Bash(
    command="npm run dev",
    run_in_background=True,
    description="Start development server"
)
# Returns immediately with shell_id
```

**Returns**:
```python
{
    "shell_id": "bash_12345",
    "message": "Command started in background"
}
```

### Monitoring Background Processes

Use BashOutput to check progress:

```python
# Start background job
result = Bash("npm run dev", run_in_background=True)
shell_id = result.shell_id

# Later, check output
BashOutput(bash_id=shell_id)
# Returns new output since last check

# Check again
BashOutput(bash_id=shell_id)
# Returns only NEW output (incremental)
```

### Filtering Output

```python
# Only show errors
BashOutput(
    bash_id=shell_id,
    filter="ERROR.*"  # Regex pattern
)
```

### Killing Background Processes

```python
KillShell(shell_id=shell_id)
```

### Background Process Examples

#### Example 1: Development Server

```python
# Start server
result = Bash("python manage.py runserver", run_in_background=True)
server_id = result.shell_id

# Do other work...

# Check if server started
BashOutput(bash_id=server_id)

# Stop server when done
KillShell(shell_id=server_id)
```

#### Example 2: Long-Running Tests

```python
# Start long test suite
result = Bash("pytest tests/ --slow", run_in_background=True)
test_id = result.shell_id

# Check progress periodically
BashOutput(bash_id=test_id, filter="PASSED|FAILED|ERROR")

# Wait for completion (check until process ends)
while True:
    output = BashOutput(bash_id=test_id)
    if "finished" in output or "exited" in output:
        break
```

## Git Operations

### Common Git Commands

#### Git Status

**Purpose**: See current repository state

```python
Bash(
    command="git status",
    description="Check git status"
)
```

**Use for**:
- Seeing modified files
- Checking branch
- Finding untracked files

#### Git Diff

**Purpose**: See what changed

```python
# Unstaged changes
Bash("git diff")

# Staged changes
Bash("git diff --cached")

# All changes
Bash("git diff HEAD")

# Specific file
Bash("git diff path/to/file.py")
```

#### Git Log

**Purpose**: See commit history

```python
# Recent commits
Bash("git log --oneline -n 10")

# Commits in current branch
Bash("git log main..HEAD --oneline")

# Full log with details
Bash("git log --stat")
```

#### Git Add

**Purpose**: Stage changes

```python
# Stage all changes
Bash("git add -A")

# Stage specific file
Bash("git add path/to/file.py")

# Stage by pattern
Bash("git add '*.py'")
```

#### Git Commit

**Purpose**: Create commit

```python
# Simple commit
Bash('git commit -m "feat: add new feature"')

# Multi-line commit message using heredoc
Bash("""git commit -m "$(cat <<'EOF'
feat: add user authentication

- Implement JWT token generation
- Add login/logout endpoints
- Add auth middleware

🤖 Generated with Claude Code

Co-Authored-By: Claude <noreply@anthropic.com>
EOF
)"
""")
```

**Critical**: Never use `--no-verify`:

```python
# ❌ FORBIDDEN:
Bash("git commit -m 'message' --no-verify")

# ✓ CORRECT:
Bash("git commit -m 'message'")
# If pre-commit fails, fix issues and retry
```

#### Git Push

**Purpose**: Push commits to remote

```python
# Push current branch
Bash("git push")

# Push and set upstream
Bash("git push -u origin feature-branch")

# Force push (dangerous - ask permission first)
Bash("git push --force")  # Only with explicit user permission
```

#### Git Branch

**Purpose**: Manage branches

```python
# List branches
Bash("git branch -a")

# Create branch
Bash("git branch feature-name")

# Switch branch
Bash("git checkout feature-name")

# Create and switch
Bash("git checkout -b feature-name")
```

### Git Workflow Pattern

#### Standard Feature Development

```python
# 1. Check current state
Bash("git status")

# 2. Create feature branch
Bash("git checkout -b feature/add-logging")

# 3. Make code changes (Edit tool)
Edit(...)

# 4. Run tests
Bash("pytest")

# 5. Stage changes
Bash("git add -A")

# 6. Check what's staged
Bash("git diff --cached")

# 7. Commit
Bash("""git commit -m "$(cat <<'EOF'
feat: add logging to data processor

- Add logging import
- Log processing start/end
- Log errors with context

🤖 Generated with Claude Code

Co-Authored-By: Claude <noreply@anthropic.com>
EOF
)"
""")

# 8. Push
Bash("git push -u origin feature/add-logging")
```

#### Handling Pre-Commit Failures

```python
# Attempt commit
result = Bash('git commit -m "feat: add feature"')

if result.exit_code != 0:
    # Pre-commit failed
    print("Pre-commit hooks failed:")
    print(result.stderr)

    # Read error output to understand issue
    # Fix issues (e.g., run formatter)
    Bash("ruff format .")

    # Re-add changes
    Bash("git add -A")

    # Retry commit
    Bash('git commit -m "feat: add feature"')
```

**Never bypass with --no-verify**.

### Git Safety Rules

#### Rule 1: Never Use --no-verify

```python
# ❌ FORBIDDEN:
git commit --no-verify
git push --no-verify

# ✓ CORRECT:
# Fix pre-commit issues, then commit normally
```

#### Rule 2: Check Status Before Committing

```python
# ✓ CORRECT pattern:
Bash("git status")  # See what will be committed
Bash("git diff --cached")  # Review staged changes
Bash("git commit -m '...'")  # Then commit
```

#### Rule 3: Ask Before Force Push to Main/Master

```python
# ❌ DON'T DO WITHOUT PERMISSION:
Bash("git push --force origin main")

# ✓ ASK FIRST:
# "I need to force push to main. This is dangerous. Do you want me to proceed?"
```

#### Rule 4: Verify Tests Before Pushing

```python
# ✓ CORRECT workflow:
Edit(...)  # Make changes
Bash("pytest")  # Verify tests pass
Bash("git add -A && git commit -m '...'")  # Then commit
Bash("git push")  # Then push
```

### Creating Pull Requests

Use `gh` CLI for GitHub operations:

```python
# Create PR
Bash("""gh pr create --title "Add feature X" --body "$(cat <<'EOF'
## Summary
- Implemented feature X
- Added tests
- Updated docs

## Test plan
- [x] Unit tests pass
- [x] Integration tests pass
- [x] Manual testing complete

🤖 Generated with Claude Code
EOF
)"
""")

# View PR
Bash("gh pr view")

# List PRs
Bash("gh pr list")
```

### Git Worktrees

For isolated work:

```python
# Create worktree
Bash("git worktree add ../project-feature feature-branch")

# Work in worktree (different directory)
Bash("cd ../project-feature && make test")

# Remove worktree when done
Bash("git worktree remove ../project-feature")
```

## Testing Workflows

### Running Tests

#### Basic Test Execution

```python
# Run all tests
Bash("pytest")

# Run specific file
Bash("pytest tests/test_auth.py")

# Run specific test
Bash("pytest tests/test_auth.py::test_login")

# Verbose output
Bash("pytest -v")

# Show print statements
Bash("pytest -s")
```

#### Test Output Parsing

**Always read test output**:

```python
result = Bash("pytest tests/")

if result.exit_code == 0:
    print("✓ All tests passed")
else:
    # Parse failures
    if "FAILED" in result.stdout:
        print("Some tests failed:")
        print(result.stdout)
```

#### Coverage Reports

```python
# Run with coverage
Bash("pytest --cov=src --cov-report=term-missing")

# Generate HTML report
Bash("pytest --cov=src --cov-report=html")

# Check coverage threshold
Bash("pytest --cov=src --cov-fail-under=80")
```

### Test-Driven Development Pattern

```python
# 1. Write failing test
Write("tests/test_feature.py", content="""
def test_new_feature():
    result = new_feature()
    assert result == expected
""")

# 2. Run test (should fail)
result = Bash("pytest tests/test_feature.py")
assert result.exit_code != 0, "Test should fail initially"

# 3. Implement feature
Edit("src/module.py", old_string="...", new_string="...")

# 4. Run test (should pass)
result = Bash("pytest tests/test_feature.py")
assert result.exit_code == 0, "Test should pass after implementation"

# 5. Commit
Bash('git add -A && git commit -m "feat: implement feature"')
```

## Build and Deployment

### Building Projects

```python
# Python build
Bash("python -m build")

# Node.js build
Bash("npm run build")

# Rust build
Bash("cargo build --release")

# Docker build
Bash("docker build -t myapp:latest .")
```

### Installing Dependencies

```python
# Python (uv)
Bash("uv sync")

# Python (pip)
Bash("pip install -r requirements.txt")

# Node.js
Bash("npm install")

# System packages (with permission)
Bash("brew install <package>")
```

### Running Services

```python
# Start database
Bash("docker-compose up -d postgres", run_in_background=True)

# Start development server
Bash("uvicorn app:main --reload", run_in_background=True)

# Start background worker
Bash("celery worker -A tasks", run_in_background=True)
```

## Environment Variables

### Setting Variables

```python
# Single command
Bash("API_KEY=secret python script.py")

# Multiple commands
Bash("""
export API_KEY=secret
export DEBUG=true
python script.py
""")
```

### Reading Variables

```python
result = Bash("echo $HOME")
home_dir = result.stdout.strip()
```

### .env Files

```python
# Load .env and run command
Bash("set -a && source .env && set +a && python script.py")
```

## Error Handling

### Checking Exit Codes

```python
result = Bash("command_that_might_fail")

if result.exit_code == 0:
    # Success
    handle_success(result.stdout)
else:
    # Failure
    handle_error(result.stderr)
```

### Common Exit Codes

- `0`: Success
- `1`: General error
- `2`: Misuse of shell command
- `126`: Command cannot execute
- `127`: Command not found
- `130`: Terminated by Ctrl+C
- `255`: Exit status out of range

### Parsing Error Output

```python
result = Bash("pytest tests/")

if "FAILED" in result.stdout:
    # Extract failed test names
    failures = re.findall(r"FAILED (.*?) -", result.stdout)
    print(f"Failed tests: {failures}")
```

## Advanced Patterns

### Conditional Execution

```python
# Run command2 only if command1 succeeds
Bash("test -f config.json && python script.py")

# Run command2 only if command1 fails
Bash("test -f config.json || echo 'Config missing'")
```

### Looping

```python
# Process multiple files
Bash("""
for file in src/*.py; do
    echo "Processing $file"
    ruff check "$file"
done
""")
```

### Here Documents

```python
# Create file with heredoc
Bash("""
cat > config.txt <<'EOF'
setting1=value1
setting2=value2
EOF
""")
```

### Pipes and Filters

```python
# Chain commands with pipes
Bash("cat data.txt | grep ERROR | wc -l")

# Complex pipeline
Bash("git log --oneline | head -n 10 | awk '{print $1}'")
```

## Timeout Management

### Setting Timeouts

```python
# Default: 2 minutes (120000ms)
Bash("quick_command")

# Custom timeout: 10 seconds
Bash("command", timeout=10000)

# Long timeout: 10 minutes
Bash("slow_build", timeout=600000)

# Max timeout: 10 minutes
Bash("very_slow", timeout=600000)  # Maximum allowed
```

### Handling Timeouts

```python
result = Bash("slow_command", timeout=5000)

if "timeout" in result.stderr.lower():
    print("Command timed out")
    # Use background execution instead
    Bash("slow_command", run_in_background=True)
```

## Restrictions and Limitations

### Cannot Use Interactive Commands

**Problem**: Commands that expect user input will hang

```python
# ❌ WRONG: Interactive commands
Bash("vim file.py")  # Requires interaction
Bash("git rebase -i")  # Interactive rebase
Bash("python")  # Python REPL
```

**Solution**: Use non-interactive alternatives

```python
# ✓ CORRECT: Non-interactive
Edit("file.py", ...)  # Instead of vim
Bash("git rebase main")  # Non-interactive rebase
Bash("python script.py")  # Run script, not REPL
```

**Exception**: Use tmux pattern via MCP for truly interactive needs.

### No Access to Certain Commands

Some commands may be unavailable:

```python
# May not be available:
timeout  # Use Bash timeout parameter instead
gtimeout  # Not installed

# Available:
python
git
npm
make
# ... (standard Unix tools)
```

### Working Directory Resets

Covered earlier - use `&&` or absolute paths.

## Best Practices

### 1. Always Check Exit Codes

```python
✓ result = Bash("pytest")
  if result.exit_code != 0:
      handle_failure()

❌ Bash("pytest")  # Ignoring result
```

### 2. Use Descriptive Commands

```python
✓ Bash("pytest tests/", description="Run test suite")
❌ Bash("pytest tests/")
```

### 3. Chain Related Commands

```python
✓ Bash("cd /path && pytest")
❌ Bash("cd /path")
  Bash("pytest")  # Won't work - directory reset
```

### 4. Never Bypass Pre-Commit

```python
✓ Bash("git commit -m 'message'")
❌ Bash("git commit -m 'message' --no-verify")
```

### 5. Verify After Changes

```python
✓ Edit(...)
  Bash("pytest")  # Verify

❌ Edit(...)
  Bash("git commit -m 'message'")  # No verification
```

### 6. Use Background for Long Operations

```python
✓ Bash("npm run dev", run_in_background=True)
❌ Bash("npm run dev")  # Blocks forever
```

### 7. Parse Output, Don't Assume

```python
✓ result = Bash("command")
  if "SUCCESS" in result.stdout:
      proceed()

❌ Bash("command")
  proceed()  # Assuming success
```

## Common Command Patterns

### File Operations

```bash
# Find files
find . -name "*.py" -type f

# Count lines
wc -l file.txt

# Search and replace
sed -i 's/old/new/g' file.txt  # But prefer Edit tool
```

### Process Management

```bash
# List processes
ps aux | grep python

# Kill process
kill -9 <pid>

# Check port usage
lsof -i :8000
```

### System Information

```bash
# Disk usage
df -h

# Memory usage
free -h

# CPU info
top -n 1
```

## Debugging Bash Commands

### Verbose Execution

```python
# See what's happening
Bash("set -x && command")  # Print commands as executed
```

### Error Tracing

```python
# Stop on first error
Bash("set -e && command1 && command2 && command3")
```

### Dry Run

```python
# Check what would happen
Bash("make -n build")  # Dry run
```

## Summary

**Key Points**:
1. Bash executes commands in your actual shell environment
2. Working directory resets between calls - use `&&` or absolute paths
3. Always check exit codes and parse output
4. Never use `--no-verify` with git commands
5. Use background execution for long-running processes
6. Chain commands with `&&` for sequential, `&` for parallel
7. Verify tests pass before committing

**Common Patterns**:
- Run tests: `pytest`
- Git workflow: `git status → add → commit → push`
- Build: `npm run build` or `python -m build`
- Background: `run_in_background=True` + `BashOutput`

**Restrictions**:
- No interactive commands (use alternatives)
- No persistent directory changes (use chaining)
- Never bypass pre-commit hooks
- Max timeout: 10 minutes

---

## Background Execution Patterns

### Overview

Background execution allows long-running commands to execute while Claude Code continues other work. This is essential for:
- Development servers (webpack-dev-server, vite, etc.)
- Database servers
- File watchers
- Long-running tests
- Build processes

### The Background Execution Workflow

```
1. Start command in background
   Bash(command="npm run dev", run_in_background=True)
   → Returns immediately with shell_id

2. Continue other work
   Read files, make edits, etc.

3. Check output when needed
   BashOutput(bash_id=shell_id)
   → Get new output since last check

4. Kill when done
   KillShell(shell_id=shell_id)
```

### Starting Background Commands

```python
# Start dev server
Bash(
    command="npm run dev",
    run_in_background=True,
    description="Start development server"
)

# Returns immediately:
# {
#   "shell_id": "shell_abc123",
#   "message": "Started in background"
# }
```

**Key difference from normal Bash**:
- Normal: Waits for completion, returns output
- Background: Returns immediately with shell_id

### Monitoring Background Output

Use `BashOutput` to check what the background process has produced:

```python
# Get all new output since last check
BashOutput(bash_id="shell_abc123")

# Returns:
# {
#   "stdout": "Server running on http://localhost:3000\n",
#   "stderr": "",
#   "status": "running"
# }
```

**Important**: Each call returns only **new** output since previous check.

### Filtering Output

Use regex to filter for specific patterns:

```python
# Only show error lines
BashOutput(
    bash_id="shell_abc123",
    filter="ERROR|WARN"
)

# Only show successful test results
BashOutput(
    bash_id="shell_test",
    filter="PASS|✓"
)
```

**Warning**: Filtered-out lines are **discarded** and won't be available in subsequent calls.

### Terminating Background Processes

```python
# Graceful shutdown
KillShell(shell_id="shell_abc123")

# Forceful if needed (SIGKILL)
KillShell(shell_id="shell_abc123")  # Automatically escalates if process doesn't stop
```

### Common Patterns

#### Pattern 1: Dev Server Management

```python
# Start server
result = Bash(
    command="npm run dev",
    run_in_background=True
)
server_id = result["shell_id"]

# Wait for server to be ready
ready = False
for attempt in range(10):  # Try 10 times
    output = BashOutput(bash_id=server_id)
    if "Server running" in output["stdout"]:
        ready = True
        break
    # Note: In practice, you'd check output periodically
    # Claude Code doesn't have direct sleep - use multiple BashOutput calls

if ready:
    # Run tests against server
    Bash("npm test")

    # Shut down server
    KillShell(shell_id=server_id)
```

#### Pattern 2: File Watcher

```python
# Start file watcher
watcher = Bash(
    command="npm run watch",
    run_in_background=True
)

# Make changes
Edit(file_path="src/app.js", ...)

# Check if watcher detected changes
output = BashOutput(bash_id=watcher["shell_id"])
if "Recompiled" in output["stdout"]:
    print("Changes detected and recompiled!")

# Clean up
KillShell(shell_id=watcher["shell_id"])
```

#### Pattern 3: Long-Running Build

```python
# Start build in background
build = Bash(
    command="npm run build:production",
    run_in_background=True
)

# Do other work while building
Read("README.md")
Edit("package.json", ...)

# Periodically check build progress
BashOutput(bash_id=build["shell_id"], filter="\\d+%")  # Show progress percentages

# Eventually check completion
status = BashOutput(bash_id=build["shell_id"])
if status["status"] == "completed":
    if status["exit_code"] == 0:
        print("Build succeeded!")
    else:
        print(f"Build failed with code {status['exit_code']}")
```

#### Pattern 4: Database Server

```python
# Start local database
db = Bash(
    command="docker run -p 5432:5432 postgres",
    run_in_background=True
)

# Wait for database to be ready
for attempt in range(30):
    output = BashOutput(bash_id=db["shell_id"])
    if "database system is ready" in output["stdout"]:
        break
    # Check again after some time has passed
    # In practice, make multiple BashOutput calls

# Run migrations
Bash("python manage.py migrate")

# Run tests
Bash("pytest")

# Shut down database
KillShell(shell_id=db["shell_id"])
```

### Background Execution Best Practices

#### DO:

✅ **Store shell_id** - Save it in variable for later use
✅ **Check startup** - Verify process started successfully
✅ **Monitor errors** - Use filter to catch stderr
✅ **Clean up** - Always KillShell when done
✅ **Set timeouts** - Prevent runaway processes
✅ **Log output** - Save important output for debugging

#### DON'T:

❌ **Forget to kill** - Orphaned processes waste resources
❌ **Assume immediate startup** - Check readiness before depending on service
❌ **Ignore stderr** - Errors may only appear in stderr
❌ **Check too frequently** - Rate-limit BashOutput calls
❌ **Filter critical output** - Might miss important errors

### Debugging Background Processes

#### Problem: Process not starting

```python
# Check output immediately
result = Bash(command="npm run dev", run_in_background=True)
output = BashOutput(bash_id=result["shell_id"])

if output["status"] == "failed":
    print(f"Failed to start: {output['stderr']}")
```

#### Problem: Process exits unexpectedly

```python
# Monitor status by checking periodically
output = BashOutput(bash_id=shell_id)
if output["status"] in ["completed", "failed"]:
    print(f"Process ended unexpectedly:")
    print(f"Exit code: {output['exit_code']}")
    print(f"Last output: {output['stdout']}")
# In practice, check status periodically via multiple BashOutput calls
```

#### Problem: Can't tell if process is ready

```python
# Use specific ready message
ready_message = "Listening on port 3000"

output = BashOutput(bash_id=shell_id)
if ready_message not in output["stdout"]:
    # Not ready yet, check logs
    print("Process started but not ready. Current output:")
    print(output["stdout"])
```

### Limitations

**Background execution cannot**:
- Accept interactive input (stdin)
- Change directory permanently (each command starts in cwd)
- Persist environment variables across commands
- Share state with other background processes

**Workarounds**:
```python
# ❌ Can't do this:
Bash("read -p 'Enter name: ' name", run_in_background=True)

# ✅ Do this instead:
# Prompt user beforehand with AskUserQuestion
# Then pass value via command argument or environment variable
```

### Performance Considerations

**Each BashOutput call**:
- Makes a system call
- Reads from process buffers
- Filters if regex provided
- Returns delta since last call

**Optimization**:
```python
# ❌ Inefficient - avoid rapid polling:
# Checking too frequently wastes resources

# ✅ Efficient - check periodically:
# Space out BashOutput calls with reasonable intervals
# Use multiple sequential checks instead of tight loops
```

### Real-World Example: Full Development Workflow

```python
# 1. Start database
db = Bash(
    command="docker run --rm -p 5432:5432 -e POSTGRES_PASSWORD=dev postgres",
    run_in_background=True
)

# 2. Wait for database
for attempt in range(20):
    output = BashOutput(bash_id=db["shell_id"])
    if "ready to accept connections" in output["stdout"]:
        break
    # In practice, check multiple times with spacing between checks

# 3. Run migrations
Bash("python manage.py migrate")

# 4. Start dev server
server = Bash(
    command="python manage.py runserver",
    run_in_background=True
)

# 5. Wait for server
for i in range(10):
    output = BashOutput(bash_id=server["shell_id"])
    if "Starting development server" in output["stdout"]:
        break
    time.sleep(1)

# 6. Run tests
test_result = Bash("pytest tests/ -v")

# 7. Check test output
if test_result["exit_code"] == 0:
    print("All tests passed!")
else:
    print("Tests failed, checking server logs...")
    server_output = BashOutput(bash_id=server["shell_id"])
    print(server_output["stdout"])

# 8. Clean up
KillShell(shell_id=server["shell_id"])
KillShell(shell_id=db["shell_id"])
```

---

*This completes the core Claude Code documentation suite. For additional topics, see planned documentation in [00-INDEX.md](./00-INDEX.md).*
