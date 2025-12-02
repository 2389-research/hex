# Error Handling

How Claude Code handles errors, recovers from failures, and knows when to ask for help.

## Error Handling Philosophy

```
1. Errors are information, not failures
2. Read error messages carefully - they contain clues
3. Fix root causes, not symptoms
4. Learn from each error
5. Ask for help when stuck
6. Never hide or bypass errors
```

## Error Categories

### Tool Failures
- Read/Write/Edit errors
- Bash command failures
- Grep/Glob failures
- Skill execution errors

### Code Errors
- Test failures
- Linting violations
- Type errors
- Runtime errors

### System Errors
- Permission denied
- File not found
- Network errors
- Resource exhaustion

### Hook Errors
- Pre-commit failures
- Pre-push failures
- Commit-msg failures

## Tool Failure Recovery

### Read Tool Errors

**Error: File Not Found**
```python
Error: No such file or directory: /path/to/file.py

Decision tree:
1. Is this expected (checking if exists)?
   → Use information, continue
   → Example: "File doesn't exist, I'll create it"

2. Did user specify this path?
   → Ask user to verify path
   → Example: "Can't find /path/to/file.py. Is the path correct?"

3. Was I inferring the path?
   → Search for the file
   → Example: Use Glob "**/*file.py"
```

**Example Recovery:**
```
User: "Update the config in settings.py"

Attempt:
Read("/app/settings.py")
Error: No such file or directory

Recovery:
1. Search for it:
   Glob("**/settings.py")

2. Found: /app/config/settings.py

3. Explain:
   "Found settings.py at /app/config/settings.py (not /app/settings.py).
    Updating that file now."

4. Continue:
   Read("/app/config/settings.py")
```

**Error: Permission Denied**
```python
Error: Permission denied: /etc/system.conf

Decision:
→ ALWAYS ask user
→ NEVER try to bypass (sudo, chmod)

Response:
"I don't have permission to read /etc/system.conf.
 You'll need to either:
 1. Run claude code with sudo (not recommended)
 2. Change file permissions: chmod +r /etc/system.conf
 3. Copy the file to a readable location

 Which would you prefer?"
```

### Write Tool Errors

**Error: Must Read Before Write**
```python
Error: File exists but was not read first

Cause: Trying to Write to existing file without Read

Recovery:
1. Read the file first
2. Then Write
```

**Example:**
```
Attempt:
Write("/app/config.py", "new content")
Error: Must read file first

Recovery:
1. Read("/app/config.py")
2. Analyze existing content
3. Decide: Edit (partial) vs Write (full replacement)
4. Execute correct operation
```

**Error: Directory Doesn't Exist**
```python
Error: No such file or directory: /new/path/file.py

Recovery:
1. Check if this is intentional (new directory structure)
2. Create directory first using Bash
3. Then write file

Example:
```bash
mkdir -p /new/path
```
Write("/new/path/file.py", content)
```

### Edit Tool Errors

**Error: old_string Not Found**
```python
Error: old_string not found in file

Causes:
1. String doesn't exist
2. Whitespace mismatch
3. String appears multiple times

Recovery:
1. Read file again to see actual content
2. Find correct string (exact match including whitespace)
3. Make old_string more unique (add surrounding context)
4. Or use replace_all=True if intentional
```

**Example Recovery:**
```
Attempt:
Edit(
  file_path="/app/api.py",
  old_string="def login():",
  new_string="def login(username: str):"
)
Error: old_string not found in file

Recovery:
1. Read("/app/api.py") to see actual content

2. Found actual string has different indentation:
   "    def login():"  # 4 spaces

3. Retry with correct whitespace:
Edit(
  file_path="/app/api.py",
  old_string="    def login():",
  new_string="    def login(username: str):"
)

Success!
```

**Error: old_string Not Unique**
```python
Error: old_string appears 3 times, must be unique

Recovery options:
1. Add more context to make unique:
   old_string: "# User routes\ndef login():"

2. Use replace_all=True if want all:
   replace_all=True

3. Edit each occurrence separately with unique context
```

### Bash Tool Errors

**Error: Command Not Found**
```bash
bash: ruff: command not found

Recovery:
1. Is this a dependency we need to install?
   → Install: uv add --dev ruff

2. Is this a tool that should exist?
   → Ask user: "ruff command not found. Should I install it?"

3. Is there an alternative?
   → Use alternative: uv run ruff (uses project venv)
```

**Error: Command Failed (Non-Zero Exit)**
```bash
Command: pytest
Exit code: 1
Output: [test failure output]

Decision tree:
1. Is this expected? (TDD red phase)
   → Use information, continue to implementation

2. Is this a regression?
   → Fix the code causing failure

3. Is this a test environment issue?
   → Fix environment (install deps, setup data)
```

**Error: Syntax Error in Command**
```bash
bash: syntax error near unexpected token '('

Cause: Usually improper quoting

Recovery:
1. Review command for special characters
2. Add proper quoting:
   Bad:  cd My Documents
   Good: cd "My Documents"

3. Escape special characters:
   Bad:  echo $PATH
   Good: echo \$PATH
```

### Grep Tool Errors

**Error: No Matches Found**
```python
Grep(pattern="import requests")
Result: No matches found

Decision tree:
1. Is this informative? (checking if something exists)
   → Use information: "No requests imports found"

2. Was pattern wrong?
   → Try alternative patterns:
     - Case insensitive: -i=True
     - Regex variations: "import.*request"

3. Wrong directory?
   → Try different path
```

**Error: Invalid Regex**
```python
Error: Invalid regex pattern

Cause: Special regex characters not escaped

Recovery:
1. Escape special characters: . * + ? [ ] { } ( ) | \ ^  $
2. Use literal strings when possible
3. Test pattern before using

Example:
Bad:  "function()"       # () are regex groups
Good: "function\\(\\)"  # Escaped parentheses
```

### Glob Tool Errors

**Error: No Files Match Pattern**
```python
Glob("*.py")
Result: []

Decision tree:
1. Wrong directory?
   → Try with path: Glob("*.py", path="/app")

2. Wrong pattern?
   → Try recursive: Glob("**/*.py")

3. Files really don't exist?
   → Report to user: "No Python files found"
```

## Test Failure Protocols

### When Tests Fail

**Expected Failure (TDD)**
```
Context: Just wrote a new test

Action:
1. Confirm test failure is expected
   "The test fails as expected (red phase).
    Now implementing the feature..."

2. Implement minimal code to pass

3. Run test again to verify green

4. Refactor if needed
```

**Regression (Previously Passing)**
```
Context: Tests were passing, now failing after code change

Action:
1. Read test failure output carefully

2. Identify what changed
   git diff

3. Understand why change broke test

4. Decide fix:
   - Code is wrong → Fix code
   - Test is wrong → Fix test (rarely)
   - Both need adjustment → Fix both

5. Apply fix

6. Re-run tests

7. Verify all pass
```

**Test Environment Issue**
```
Context: Test fails due to missing dependency, wrong setup, etc.

Errors like:
- ModuleNotFoundError
- Connection refused
- File not found (test fixtures)

Action:
1. Identify missing component

2. Install/setup component:
   - Dependencies: uv add pytest-mock
   - Data: Create test fixtures
   - Services: Start test database

3. Re-run tests
```

**Unclear Test Failure**
```
Context: Don't understand why test is failing

Action:
1. Read test code to understand intent

2. Read test output carefully

3. Add debug output if needed:
   - Print statements
   - Better assertions
   - Logging

4. Run test with verbose: pytest -vv

5. If still unclear, ask user:
   "I'm not sure why test_login is failing.
    The error is: [error]
    But I expected: [expectation]
    Can you help me understand what's wrong?"
```

### Test Output Analysis

**Reading pytest Output**
```
FAILED tests/test_api.py::test_login - AssertionError: assert 401 == 200

Parse:
- File: tests/test_api.py
- Test: test_login
- Error: AssertionError
- Details: Got 401, expected 200

Meaning:
- Login endpoint returned 401 (Unauthorized)
- Test expected 200 (OK)
- Likely: Authentication is failing
```

**Reading Stack Traces**
```
Traceback (most recent call last):
  File "/app/api/routes.py", line 42, in login
    user = db.get_user(username)
  File "/app/db.py", line 15, in get_user
    return self.users[username]
KeyError: 'alice'

Parse:
- Error: KeyError
- Location: db.py, line 15
- Cause: username 'alice' not in self.users dict
- Root: get_user doesn't handle missing users

Fix: Add KeyError handling in get_user
```

## Hook Failure Handling

### Pre-Commit Hook Failures

See [15-GIT-WORKFLOWS.md](./15-GIT-WORKFLOWS.md#pre-commit-hook-protocol) for complete protocol.

**Quick Reference:**
```
1. Read error aloud (explain what failed)
2. Identify tool and reason
3. Explain the fix
4. Apply fix
5. Re-run hooks
6. Only proceed after success

NEVER use --no-verify
```

**Common Hook Failures:**

**Formatter (ruff, black, biome)**
```
Error: Code not formatted correctly

Fix:
1. Run formatter:
   uv run ruff format .

2. Stage changes:
   git add <formatted files>

3. Retry commit
```

**Linter (ruff, eslint, pylint)**
```
Error: Linting violations found

Fix:
1. Read linter output
2. Fix each violation:
   - Unused imports → Remove
   - Undefined names → Import or fix typo
   - Style issues → Follow linter suggestion

3. Re-run linter to verify:
   uv run ruff check .

4. Retry commit
```

**Type Checker (mypy)**
```
Error: Type errors found

Fix:
1. Read mypy output
2. Add type annotations:
   - Function parameters
   - Return types
   - Variable annotations

3. Fix type mismatches
4. Re-run mypy
5. Retry commit
```

**Tests (pytest)**
```
Error: Tests failed

Fix:
1. Identify which tests failed
2. Fix code or tests (usually code)
3. Ensure all tests pass:
   uv run pytest
4. Retry commit
```

## Permission Denied Scenarios

### File System Permissions

**Can't Read File**
```
Error: Permission denied: /path/to/file

Response:
"I don't have permission to read /path/to/file.
 This file is owned by [owner] with permissions [mode].

 You can fix this by:
 1. chmod +r /path/to/file
 2. Run claude code as [owner]
 3. Copy file to accessible location

 Which approach should I use?"
```

**Can't Write File**
```
Error: Permission denied: /path/to/file

Response:
"I don't have permission to write to /path/to/file.

 You can fix this by:
 1. chmod +w /path/to/file
 2. Run claude code with appropriate permissions
 3. Write to a different location

 What would you like to do?"
```

**Can't Execute Command**
```
Error: Permission denied: /usr/local/bin/tool

Response:
"I can't execute /usr/local/bin/tool due to permissions.

 You can fix this by:
 1. chmod +x /usr/local/bin/tool
 2. Install tool in user directory: uv tool install tool
 3. Use alternative approach

 How should I proceed?"
```

### Git Permissions

**Push Denied**
```
Error: Permission denied (publickey)

Response:
"Can't push to remote - SSH key authentication failed.

 This usually means:
 1. SSH key not added to GitHub/GitLab
 2. SSH agent not running
 3. Wrong remote URL (should be git@github.com:...)

 Check your SSH setup:
 ssh -T git@github.com

 Would you like me to help debug this?"
```

**Protected Branch**
```
Error: Cannot push to protected branch 'main'

Response:
"The 'main' branch is protected and requires:
 - Pull request workflow
 - Code review
 - Status checks to pass

 I'll create a feature branch and PR instead:
 1. Create branch: feature/[name]
 2. Push changes there
 3. Create pull request

 Is that OK?"
```

## Network Errors

### API Request Failures

**Connection Refused**
```
Error: Connection refused: localhost:8000

Recovery:
1. Is service running?
   "The API server doesn't seem to be running.
    Start it with: uv run python -m app"

2. Wrong port?
   "Trying to connect to port 8000. Is the service
    running on a different port?"

3. Wrong host?
   "Should I connect to a different host?"
```

**Timeout**
```
Error: Request timeout after 30s

Recovery:
1. Service slow/overloaded?
   "Request timed out. The service might be:
    - Slow to respond (increase timeout?)
    - Overloaded (check service health)
    - Stuck (restart needed?)"

2. Network issue?
   "Can't reach the service. Network problem?"
```

**DNS Resolution Failed**
```
Error: Could not resolve host: api.example.com

Recovery:
"Can't resolve hostname api.example.com.

 Possible causes:
 1. Typo in hostname
 2. DNS not configured
 3. Not connected to network/VPN

 Can you verify the hostname is correct?"
```

### Package Installation Errors

**PyPI Connection Failed**
```
Error: Could not fetch https://pypi.org/simple/package

Recovery:
1. Retry (transient failure)
   "PyPI connection failed. Retrying..."

2. Check network
   "Can't reach PyPI. Network connected?"

3. Use cache
   UV_OFFLINE=1 uv sync
   "Using cached packages only"

4. Use mirror
   "Should I use a PyPI mirror?"
```

## Resource Exhaustion

### Out of Disk Space

```
Error: No space left on device

Recovery:
"Out of disk space. Need to free up space:

 Current usage:
 df -h

 Options:
 1. Clean uv cache: uv cache clean
 2. Remove old virtualenvs
 3. Clean Docker: docker system prune
 4. Free up system space

 What would you like to do?"
```

### Out of Memory

```
Error: Cannot allocate memory

Recovery:
"Process ran out of memory.

 Options:
 1. Close other applications
 2. Reduce batch size (if processing data)
 3. Use memory-efficient approach
 4. Add swap space

 This usually happens when processing large files.
 Can you provide more system resources?"
```

### Too Many Open Files

```
Error: Too many open files

Recovery:
"Hit open file limit.

 Current limit: ulimit -n

 Fixes:
 1. Close files properly (likely code bug)
 2. Increase limit: ulimit -n 4096
 3. Check for file descriptor leaks

 This suggests a bug in the code. Let me investigate..."
```

## Debugging Strategies

### Systematic Debugging Approach

**Use superpowers:systematic-debugging skill:**
```
Triggers:
- Unexpected behavior
- Hard to reproduce bugs
- Complex failures
- Multiple potential causes

Process:
1. Root cause investigation
2. Pattern analysis
3. Hypothesis testing
4. Implementation
```

### Adding Debug Information

**Strategic Print/Log Statements**
```python
# Before debugging
def process_user(user_id):
    user = db.get_user(user_id)
    return user.email

# With debug info
def process_user(user_id):
    print(f"DEBUG: Processing user_id={user_id}")
    user = db.get_user(user_id)
    print(f"DEBUG: Got user={user}")
    print(f"DEBUG: User email={user.email}")
    return user.email
```

**Verbose Flags**
```bash
# Add verbosity to commands
pytest -vv                    # Very verbose pytest
ruff check --verbose .        # Verbose linting
uv add --verbose package      # Verbose package install
git status --verbose          # Verbose git
```

**Environment Variables**
```bash
# Debug mode
DEBUG=1 uv run python app.py

# Rust tools debug
RUST_LOG=debug uv add package

# Python logging
LOGLEVEL=DEBUG uv run python app.py
```

### Isolating the Problem

**Binary Search Approach**
```
1. Does it work with half the code commented?
   YES → Problem in commented half
   NO → Problem in active half

2. Repeat until found
```

**Minimal Reproduction**
```
1. Create smallest code that reproduces bug
2. Remove everything unrelated
3. Isolate exact line causing issue
4. Fix that specific line
```

### Reading Error Messages Carefully

**Error messages contain:**
```
1. What went wrong (error type)
2. Where it happened (file, line, function)
3. Why it happened (context, values)
4. How to fix (sometimes suggestions)

Example:
TypeError: unsupported operand type(s) for +: 'int' and 'str'
  ^ Type      ^ What operation                ^ The types

Meaning: Tried to add int and str (can't do that)
Fix: Convert one to match the other
```

## When to Ask for Help

### Stuck After Multiple Attempts

```
If tried same approach 2+ times:
→ STOP
→ Explain what was tried
→ Ask for help

Example:
"I've tried fixing this test twice:
 1. First attempt: Added type annotation - still failed
 2. Second attempt: Changed return type - still failed

 The error is: [error]

 I think the issue is [hypothesis], but I'm not certain.
 Can you help me understand what's wrong?"
```

### Unclear Requirements

```
If user request is ambiguous:
→ DON'T guess
→ ASK for clarification

Example:
User: "Make it better"
Response: "I can improve the code. What specifically should
           I focus on?
           - Performance?
           - Readability?
           - Test coverage?
           - Error handling?
           - Something else?"
```

### Outside Expertise

```
If problem requires domain knowledge Claude Code lacks:
→ Acknowledge limitation
→ Ask user

Example:
"This involves Kubernetes networking, which I'm not deeply
 familiar with. The error suggests a pod can't reach the
 service, but I'm not sure if this is a:
 - Network policy issue
 - DNS issue
 - Service configuration issue

 Can you help me understand the Kubernetes setup?"
```

### User Knows Better

```
If user has more context:
→ Defer to user

Example:
"You mentioned this is legacy code with weird quirks.
 I found this pattern that looks like a bug, but maybe
 it's intentional?

 [show code]

 Should I change it or leave it as is?"
```

## Error Prevention

### Defensive Coding

```python
# Always check before using
if os.path.exists(file_path):
    content = read(file_path)
else:
    # Handle missing file

# Validate inputs
def process_user(user_id: int):
    if not isinstance(user_id, int):
        raise TypeError(f"user_id must be int, got {type(user_id)}")
    if user_id < 1:
        raise ValueError(f"user_id must be positive, got {user_id}")
    # Now safe to use user_id

# Handle expected errors
try:
    user = db.get_user(username)
except KeyError:
    return {"error": "User not found"}, 404
```

### Verification Before Action

```
Before committing:
→ Run tests: uv run pytest
→ Run linter: uv run ruff check .
→ Run formatter: uv run ruff format .
→ Verify: git status shows expected files

Before pushing:
→ Verify branch: git branch --show-current
→ Verify remote: git remote -v
→ Verify tests pass
→ Then push

Before deleting:
→ Verify correct file/directory
→ Confirm with user if important
→ Then delete
```

### Reading Before Writing

```
ALWAYS Read before Edit/Write on existing files:

1. Read shows current state
2. Understand what's there
3. Plan changes
4. Execute changes
5. Verify changes

Never skip Read - it prevents:
- Overwriting good code
- Duplicate additions
- Breaking existing functionality
```

## Error Documentation

### Logging Errors for Learning

When encountering new error types, Claude Code should (if journaling available):
```
1. Record the error
2. Note the solution
3. Reference for future

This builds pattern recognition for faster recovery.
```

## Summary of Error Handling

### Core Principles
1. Read error messages carefully
2. Fix root causes, not symptoms
3. Learn from each error
4. Ask for help when stuck
5. Never bypass or hide errors

### Recovery Patterns
- Tool failures → Retry with corrections
- Test failures → Fix code or environment
- Hook failures → Fix issues, never bypass
- Permission errors → Ask user for solution
- Network errors → Retry or use alternatives

### When to Ask Help
- Stuck after 2+ attempts
- Unclear requirements
- Outside expertise needed
- User has more context

### Prevention
- Defensive coding
- Verification before action
- Always Read before Write
- Test thoroughly

Effective error handling turns problems into learning opportunities and ensures Claude Code operates safely and reliably.
