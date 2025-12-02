# Context Management

How Claude Code manages token limits, context windows, memory, and session continuity.

## Token Budget

### Overview

Claude Code operates with a **200,000 token budget** per conversation:

```
Total budget: 200,000 tokens
Includes:
  - System instructions
  - Project-specific guidance
  - Tool definitions
  - Conversation history
  - File contents read
  - Command outputs
  - Tool responses

When budget exhausted:
  → Conversation ends
  → Use subagents for fresh context
```

### Token Consumption

**What Uses Tokens:**

```
High cost (10k-50k+ tokens):
- Reading large files (>1000 lines)
- Long command outputs
- Extensive grep results
- Large diffs
- Multiple file reads

Medium cost (1k-10k tokens):
- Normal file reads (100-1000 lines)
- Command outputs
- Conversation messages
- Tool responses

Low cost (<1k tokens):
- User messages
- Short responses
- Simple commands
- File paths
```

**Example Token Usage:**

```
Action                          Approximate Tokens
---------------------------------------------------
Read 100-line Python file       2,000-3,000
Read 1000-line file             15,000-20,000
git log --oneline -100          3,000-5,000
pytest output (10 tests)        1,000-2,000
grep result (50 matches)        2,000-4,000
User message (short)            50-200
Claude response (detailed)      500-2,000
Project guidance                    5,000-10,000
Tool definitions                20,000-30,000
```

### Monitoring Token Usage

**System Warnings:**
```
You'll see warnings like:
<system_warning>Token usage: 44393/200000; 155607 remaining</system_warning>

Meaning:
- Used: 44,393 tokens (22%)
- Remaining: 155,607 tokens (78%)
- Safe to continue
```

**Critical Levels:**

```
Token Usage             Status          Action
-----------------------------------------------------------------
0-50,000 (0-25%)       Plenty          Operate normally
50,000-100,000 (25-50%) Good           Be mindful of large reads
100,000-150,000 (50-75%) Warning       Avoid large files
150,000-180,000 (75-90%) Critical      Finish task quickly
180,000+ (90%+)        Emergency       Use subagent immediately
```

## Context Window Management

### What's in the Context

**Always Present:**
```
1. System instructions (core behavior)
2. Tool definitions (all available tools)
3. Environment info (cwd, git status, date)
4. Project guidance files:
   - Organization-wide defaults
   - Project-level overrides
5. Conversation history (all messages)
```

**Added During Conversation:**
```
1. File contents (from Read tool)
2. Search results (from Grep/Glob)
3. Command outputs (from Bash)
4. Todo lists (from TodoWrite)
5. User uploads (screenshots, docs)
```

### Context Accumulation

**How Context Grows:**

```
Turn 1: 30k tokens (system + initial message)
  User: "Read config.py"
  Read: +5k tokens (file content)
  Total: 35k

Turn 2: 35k tokens (previous + new)
  User: "Read api.py"
  Read: +10k tokens (larger file)
  Total: 45k

Turn 3: 45k tokens
  User: "Find all TODOs"
  Grep: +3k tokens (results)
  Total: 48k

...continues growing each turn
```

**Context Never Shrinks:**
```
Context is append-only within a conversation.
Old messages and file contents remain in context.
Only way to clear: Start new conversation or use subagent.
```

### Efficient Context Usage

**Do's:**

✅ **Read files strategically**
```
Instead of: Read all 50 modules
Do: Grep to find relevant modules, read only those
```

✅ **Use targeted search**
```
Instead of: Grep with broad pattern, get 1000 matches
Do: Grep with specific pattern, get 10 matches
```

✅ **Limit command output**
```
Instead of: git log (entire history)
Do: git log --oneline -10 (last 10 commits)
```

✅ **Use head_limit in Grep**
```
# Limit results to 20 matches
Grep(pattern="TODO", head_limit=20)
```

✅ **Progressive reading**
```
# Read file in chunks if large
Read(file_path="/path/to/file.py", limit=100, offset=0)
Read(file_path="/path/to/file.py", limit=100, offset=100)
```

**Don'ts:**

❌ **Don't read huge files**
```
Avoid: Read 5000-line generated file
Consider: Grep for specific sections
```

❌ **Don't re-read unnecessarily**
```
If already read file this conversation:
→ Reference it from context
→ Don't read again
```

❌ **Don't run verbose commands**
```
Avoid: pytest -vvv (extremely verbose)
Use: pytest -q (quiet)
```

❌ **Don't glob everything**
```
Avoid: Glob("**/*") (all files)
Use: Glob("**/*.py") (specific type)
```

## When to Use Subagents

### Subagent Benefits

**Fresh Context:**
```
Subagent starts with:
- 0 tokens used
- Clean context
- Only necessary information

Ideal when:
- Main context getting full
- New independent task
- Complex multi-step work
```

**Isolation:**
```
Subagent operates independently:
- Own token budget
- Own conversation history
- Own working directory (if using worktree)

Benefits:
- Won't pollute main context
- Can work in parallel (multiple subagents)
- Can fail without affecting main
```

### When to Delegate

**Use Subagent When:**

```
1. User preference (documented policies):
   "I highly prefer all work to be done via the
    subagent development skill"

2. Token budget concerns:
   - Main context >50% full
   - About to read many large files
   - Long-running complex task

3. Task characteristics:
   - Independent from current work
   - Has clear scope/boundaries
   - Requires fresh perspective

4. Multiple tasks:
   - Can work in parallel
   - Each has distinct purpose
   - No shared state needed
```

**Example Delegation:**

```
Scenario: Main context at 120k tokens (60%)
User: "Implement three new API endpoints"

Decision: Use subagent-driven-development
Reasoning:
  - User prefers subagents
  - 3 independent tasks
  - Each task will read/modify files
  - Fresh context beneficial

Implementation:
1. Create implementation plan
2. Dispatch subagent for endpoint 1
3. Review results
4. Dispatch subagent for endpoint 2
5. Review results
6. Dispatch subagent for endpoint 3
7. Final integration test in main
```

**Stay in Main When:**

```
1. Simple query/answer
2. Quick file read/search
3. Running single command
4. Continuation of current work
5. Token budget still healthy (<50k)
6. User explicitly wants main session
```

## Memory and Persistence

### What Persists Across Sessions

**Git History:**
```
Commits persist permanently:
- All code changes
- Commit messages
- Author information

Use for:
- Tracking work done
- Understanding code evolution
- Recovering old versions
```

**File System:**
```
All file changes persist:
- Code modifications
- New files created
- Deleted files gone

Use for:
- Actual project state
- Generated artifacts
- Configuration
```

**Project Files:**
```
.claude/ directory (if exists):
- Project guidance artifacts
- Commands (slash commands)
- Skills (project-specific)

Use for:
- Project-specific behavior
- Shared team conventions
```

### What Doesn't Persist

**Conversation History:**
```
NOT saved between sessions:
- What was discussed
- Files that were read
- Commands that were run
- Decisions that were made

Each new conversation starts fresh.
Exception: Subagents can report back to main.
```

**Context State:**
```
NOT saved:
- Token count
- Files in context
- Search results
- Command outputs

Every session rebuilds context from scratch.
```

**Temporary State:**
```
NOT saved:
- Environment variables (unless in shell config)
- Running processes
- Network connections
- In-memory data
```

### Building Context Across Sessions

**Use Git History:**
```
New session:
1. Check git log to see what was done
   git log --oneline -20

2. Read recent commit messages

3. Understand recent changes:
   git show HEAD
   git show HEAD~1

4. Continue from there
```

**Use Documentation:**
```
If project has docs:
1. Read README.md
2. Read project guidance
3. Read architecture docs

These explain context without consuming many tokens.
```

**User Provides Context:**
```
User can say:
"Last session we implemented auth.
 Now let's add rate limiting."

This gives context without needing to read all code.
```

## Session Continuity

### Maintaining Context in Long Tasks

**Progressive Work:**
```
Instead of: Do everything in one session
Do: Break into chunks across sessions

Session 1: Design and plan
  - Discuss approach
  - Create implementation plan
  - Save plan to file

Session 2: Implement part 1
  - Read plan from file
  - Implement first feature
  - Commit

Session 3: Implement part 2
  - Check git log to see progress
  - Implement second feature
  - Commit

Benefits:
- Each session has fresh context
- Work is saved in git
- Can pause/resume easily
```

**Using Todo Lists:**
```
TodoWrite persists in conversation:
- Track progress
- See what's done
- Know what's next

But doesn't persist across sessions.

For cross-session tracking:
- Use GitHub Issues
- Use project management tools
- Use comments in code
- Use TODO comments
```

### Handoff Between Sessions

**Ending Session:**
```
Before conversation ends:
1. Commit all work:
   git status
   git add ...
   git commit

2. Push if appropriate:
   git push

3. Summarize what was done:
   "Completed:
    - Feature X implemented
    - Tests passing
    - Committed as abc123

    Next steps:
    - Feature Y
    - Integration testing"

4. Update documentation if needed
```

**Starting Session:**
```
When resuming work:
1. Check git status:
   git log --oneline -10
   git status

2. Understand current state

3. Ask user what to work on:
   "I see we last implemented X.
    What should we work on now?"

4. Build necessary context:
   - Read relevant files
   - Review recent commits
   - Run tests to verify state
```

## Large File Handling

### Strategies for Large Files

**Avoid Reading Entire File:**
```
# Instead of:
Read("/path/to/huge_file.py")  # 5000 lines = 60k tokens

# Do:
Grep(pattern="class UserAuth", path="/path/to/huge_file.py")
# Find the class you need

# Then read just that section:
Read("/path/to/huge_file.py", offset=100, limit=50)
# Read 50 lines starting at line 100
```

**Use Search First:**
```
Workflow:
1. Grep to find relevant sections
2. Note line numbers
3. Read only those sections
4. Edit specific sections

This uses 1/10th the tokens vs reading entire file.
```

**Progressive Exploration:**
```
# Start with structure
Grep(pattern="^class |^def ", path="large_file.py")
# See classes and functions

# Then read specific parts
Read("large_file.py", offset=<function_start>, limit=50)
```

### Handling Large Outputs

**Command Output:**
```
# Avoid:
git log                    # Entire history (huge)

# Use:
git log --oneline -20      # Last 20 (small)
git log --oneline --since="1 week ago"  # Recent only
```

**Test Output:**
```
# Avoid:
pytest -vv                 # Verbose output

# Use:
pytest -q                  # Quiet (failures only)
pytest -x                  # Stop at first failure
pytest tests/test_one.py   # Run specific test
```

**Grep Output:**
```
# Avoid:
Grep(pattern="TODO")       # All TODOs (might be 1000+)

# Use:
Grep(pattern="TODO", head_limit=20)  # First 20 only
Grep(pattern="TODO: URGENT")         # More specific
Grep(pattern="TODO", path="src/api") # Limited scope
```

## Efficient Tool Usage

### Parallel Tool Calls

**When Independent:**
```python
# These don't depend on each other - run in parallel
Bash("git status")
Bash("git log --oneline -10")
Bash("git diff")

# Single response with 3 tool calls is faster
# and uses fewer tokens than 3 separate responses
```

**When Dependent:**
```python
# These depend on each other - run sequentially
result = Bash("git status")
# Parse result to see what changed
Bash(f"git add {changed_file}")
# Can't add file until we know what changed
```

### Tool Selection for Efficiency

**File Discovery:**
```
Fastest to slowest:
1. Know exact path → Read directly
2. Know pattern → Glob for files
3. Know content → Grep for content
4. Explore → Bash ls, then narrow down
```

**Code Search:**
```
Grep is efficient:
✅ Grep(pattern="def login")
   Returns: file:line matches (compact)

Avoid inefficient:
❌ Read all files, search manually
   Uses: 10x-100x more tokens
```

**File Modification:**
```
Edit is efficient:
✅ Edit(old_string="...", new_string="...")
   Changes: Only specified parts
   Tokens: Minimal

Write is expensive:
⚠️  Write(file, entire_content)
   Replaces: Entire file
   Tokens: Must send full content
   Use only for: New files or complete rewrites
```

## Conversation Summarization

### When Context Gets Large

**Internal Summarization:**
```
When context >100k tokens:

Claude Code mentally summarizes:
- Key decisions made
- Files read and their purpose
- Current state of work
- What's left to do

This helps:
- Stay focused
- Make consistent decisions
- Avoid redundant work
```

**External Summarization:**
```
Can create explicit summaries:
1. Write summary to file:
   Write("PROGRESS.md", "
     # Progress Summary
     - Implemented: X, Y, Z
     - Tested: All passing
     - Next: Feature A, B
   ")

2. Commit with detailed message:
   "Implemented user auth system
    - Added JWT token generation
    - Added login/logout endpoints
    - Added auth middleware
    - 156 tests added, all passing"

3. Update docs:
   Edit("README.md", add progress section)
```

### Communicating State to User

**Regular Updates:**
```
Long task in progress:
- Update todo list
- Report after each major step
- Show test results
- Commit frequently

This helps user understand progress without asking.
```

**Final Summary:**
```
After completing work:
"Completed user authentication implementation:

 ✅ JWT token generation (api/auth.py)
 ✅ Login endpoint (api/routes/auth.py)
 ✅ Logout endpoint (api/routes/auth.py)
 ✅ Auth middleware (api/middleware.py)
 ✅ 156 tests added, all passing
 ✅ Committed as: a1b2c3d

 The system is ready for:
 - Password reset (next feature)
 - Email verification (next feature)
 - Rate limiting (future enhancement)

 All tests pass: pytest shows 156/156 ✅"
```

## Best Practices

### Do's

✅ **Monitor token usage** - Check warnings, plan accordingly
✅ **Use subagents for complex work** - Fresh context, user preference
✅ **Search before reading** - Grep to find, then read specific parts
✅ **Limit command output** - Use flags like -q, --oneline, -10
✅ **Read progressively** - Use offset/limit for large files
✅ **Commit frequently** - Saves work across sessions
✅ **Provide summaries** - Help user understand progress
✅ **Use parallel tools** - When independent, call together

### Don'ts

❌ **Don't read huge files** - Search for sections first
❌ **Don't re-read files** - Reference from context
❌ **Don't glob everything** - Use specific patterns
❌ **Don't run verbose commands** - Use quiet modes
❌ **Don't let context fill up** - Use subagent before critical
❌ **Don't assume persistence** - Commit work, don't rely on memory
❌ **Don't waste tokens** - Be strategic about what to read

## Context Management Strategies

### For Different Task Types

**Quick Query:**
```
Tokens: Low (5k-10k)
Strategy: Direct answer, minimal reading
Example: "What's in package.json?"
  → Read package.json
  → Answer question
  → Done
```

**Single Feature:**
```
Tokens: Medium (20k-50k)
Strategy: Focused reading, implement, test
Example: "Add /health endpoint"
  → Read routes.py to understand structure
  → Write test
  → Implement endpoint
  → Run tests
  → Commit
```

**Multiple Features:**
```
Tokens: High (50k-150k)
Strategy: Use subagents or break into sessions
Example: "Add auth, rate limiting, and logging"
  → Use subagent-driven-development
  → Each feature gets fresh context
  → Review between features
```

**Large Refactoring:**
```
Tokens: Very High (>100k)
Strategy: Multiple sessions or planning + subagents
Example: "Refactor entire API layer"
  Session 1: Analyze and plan
  Session 2-4: Implement in chunks
  Session 5: Integration test

  Or: Use subagents for each module
```

## Summary

### Key Principles

1. **Token budget is finite** (200k)
   - Monitor usage
   - Use efficiently
   - Switch to subagent when needed

2. **Context is append-only**
   - Grows every turn
   - Never shrinks
   - Only clears with new conversation

3. **Subagents provide fresh context**
   - User preference
   - Independent tasks
   - Token budget relief

4. **Nothing persists except files**
   - Commit work frequently
   - Use git history
   - Document decisions

5. **Read strategically**
   - Search before reading
   - Read sections, not whole files
   - Limit command outputs

6. **Plan for continuity**
   - Commit with good messages
   - Update documentation
   - Summarize progress

Effective context management ensures Claude Code can handle both simple queries and complex multi-session projects while staying within token limits and maintaining quality.
