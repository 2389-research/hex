# Subagent System Deep Dive

## Overview

This document explains how subagents work in Claude Code at a technical and implementation level. While 05-SUBAGENT-SYSTEM.md covers what subagents are and when to use them, this document explains **how they actually work under the hood**.

## What Happens When You Invoke a Subagent

### The Invocation Flow

```
Main Agent (You)
    │
    ├─ User says: "Explore the authentication system"
    │
    ├─ OR: You recognize task needs specialist
    │
    ├─ Use Task tool with subagent_type parameter
    │
    └─ Task({
          description: "Explore authentication",
          prompt: "Detailed task instructions...",
          subagent_type: "Explore"  ← Triggers subagent
        })
            │
            ├─ Claude Code CLI receives Task tool call
            │
            ├─ Looks up "Explore" subagent definition
            │   - Checks ~/.claude/agents/Explore.md
            │   - Checks .claude/agents/Explore.md
            │   - Checks plugin agents
            │
            ├─ Parses YAML frontmatter
            │   - model: which Claude model to use
            │   - tools: which tools agent can access
            │   - temperature: response randomness
            │   - cached: whether to use prompt caching
            │
            ├─ Creates NEW isolated Claude instance
            │   - Fresh context window (empty history)
            │   - Custom system prompt (agent's instructions)
            │   - Configured tool permissions
            │   - Separate conversation thread
            │
            ├─ Injects task prompt as first user message
            │
            ├─ Subagent begins work
            │   - Has no knowledge of main agent's context
            │   - Only knows what's in the task prompt
            │   - Can use approved tools
            │   - Builds own conversation history
            │
            ├─ Subagent completes task
            │   - Generates final response
            │   - Returns result to main agent
            │
            └─ Main agent receives result
                - Result appears as Task tool result
                - Main agent integrates findings
                - Subagent context is discarded
```

## Technical Architecture

### Context Isolation

**Key insight**: Subagents are completely isolated Claude instances.

```
Main Agent Context Window:
┌─────────────────────────────────────┐
│ User: "Fix the auth bug"            │
│ You: "I'll investigate..."          │
│ [... 50 messages ...]               │
│ You: Let me explore with subagent   │
└─────────────────────────────────────┘
         │
         │ Task() tool call
         ▼
Subagent Context Window:
┌─────────────────────────────────────┐
│ System: [Explore agent instructions]│
│ User: "Explore authentication sys..." │
│ [Subagent's isolated work]          │
└─────────────────────────────────────┘
         │
         │ Returns result
         ▼
Main Agent Context Window (continued):
┌─────────────────────────────────────┐
│ Task Result: [Subagent findings]    │
│ You: "Based on exploration..."      │
└─────────────────────────────────────┘
```

**Implications**:
- Subagent doesn't see main conversation history
- Subagent doesn't know what user originally asked
- Subagent only knows what you put in the `prompt` parameter
- Main agent must extract/summarize subagent work for user

### Prompt Construction

When a subagent is invoked, Claude Code constructs its system prompt:

```
System Prompt = [Agent Markdown Content] + [Claude Code Base Instructions]

Example for Explore agent:
┌─────────────────────────────────────┐
│ You are a meticulous explorer...    │ ← From Explore.md
│                                      │
│ Your process:                        │
│ 1. Analyze codebase structure       │
│ 2. Search for patterns              │
│ 3. Document findings                │
│                                      │
│ [Standard Claude Code instructions] │ ← Base system prompt
│ - How to use tools                   │
│ - File operation rules              │
│ - Safety constraints                │
└─────────────────────────────────────┘

First User Message = [Your prompt parameter]
```

**Critical**: The markdown content in the agent file IS the system prompt. It shapes the agent's behavior, personality, and approach.

### Tool Access Control

Subagents only see tools listed in their frontmatter `tools` array:

```yaml
---
name: explorer
tools:
  - Read
  - Grep
  - Glob
---
```

**From subagent's perspective**:
```typescript
Available tools: [
  {
    name: "Read",
    description: "Read file contents...",
    parameters: { file_path: string, ... }
  },
  {
    name: "Grep",
    description: "Search file contents...",
    parameters: { pattern: string, ... }
  },
  {
    name: "Glob",
    description: "Find files by pattern...",
    parameters: { pattern: string }
  }
]

// Edit, Write, Bash, etc. are NOT visible to this agent
```

**Security benefit**: Agent cannot use tools it doesn't have. Trying to call `Edit` would fail immediately.

### Model Selection

Different subagents can use different Claude models:

```yaml
---
name: quick-checker
model: haiku  # Fast, cheap, less capable
---

---
name: deep-analyzer
model: sonnet  # Balanced (default)
---

---
name: strategic-planner
model: opus  # Slow, expensive, most capable
---
```

**Model mapping**:
- `haiku` → `claude-haiku-3-5-20250320`
- `sonnet` → `claude-sonnet-4-5-20250929`
- `opus` → `claude-opus-3-5-20250514`

**Cost implications**:
```
Task with 10,000 input tokens + 2,000 output tokens:

Haiku:  $0.025 input + $0.025 output = $0.05
Sonnet: $0.30 input  + $0.60 output  = $0.90
Opus:   $1.50 input  + $3.00 output  = $4.50

90x price difference between Haiku and Opus!
```

## The Task Tool Deep Dive

### Task Tool Parameters

```typescript
Task({
  description: string,     // Short label (for logging/UI)
  prompt: string,          // What you want the subagent to do
  subagent_type: string,   // Which agent to use
  model?: string           // Override agent's default model
})
```

### Example Invocation

**Main agent code**:
```javascript
I'll explore the authentication system using the Explore subagent.

Task({
  description: "Explore authentication system",
  prompt: `Explore the authentication system in this codebase.

Focus on:
1. How users log in (OAuth, JWT, sessions?)
2. Where passwords/tokens are stored
3. Authorization middleware implementation
4. Security vulnerabilities

Files to prioritize:
- src/auth/**
- src/middleware/auth.js
- src/models/User.js

Provide:
- Architecture summary
- Security assessment
- List of all auth-related files with purposes`,

  subagent_type: "Explore"
})
```

**What the subagent sees**:
```
System: [Explore agent instructions from Explore.md]

User: Explore the authentication system in this codebase.

Focus on:
1. How users log in (OAuth, JWT, sessions?)
2. Where passwords/tokens are stored
3. Authorization middleware implementation
4. Security vulnerabilities

Files to prioritize:
- src/auth/**
- src/middleware/auth.js
- src/models/User.js

Provide:
- Architecture summary
- Security assessment
- List of all auth-related files with purposes
```

**What the subagent does NOT see**:
- The user's original question
- Your previous conversation
- Other files you've looked at
- Context about the project

**Why the prompt is critical**: It's the ONLY context the subagent has. If you don't include it in the prompt, the subagent doesn't know about it.

### Return Value

The Task tool returns the subagent's final response as a string:

```typescript
const result = await Task({...})

// result is a string containing the subagent's complete response
// You then summarize/integrate this for the user
```

**Example flow**:
```
Main agent: Task({ prompt: "Explore auth...", subagent_type: "Explore" })
            ↓
Subagent: [Does research, reads files, analyzes code]
          "I've explored the authentication system. Here are my findings:

          Architecture:
          - Uses JWT tokens stored in httpOnly cookies
          - OAuth2 integration with Google and GitHub
          - Session management via Redis

          Security Assessment:
          - ✅ Passwords hashed with bcrypt (12 rounds)
          - ✅ CSRF protection implemented
          - ⚠️  JWT tokens never expire (security issue)
          - ⚠️  No rate limiting on login endpoint

          Files:
          - src/auth/strategies/jwt.js - JWT verification
          - src/auth/strategies/oauth.js - OAuth handlers
          - src/middleware/requireAuth.js - Route protection
          ..."
            ↓
Main agent receives result:
            ↓
Main agent to user:
"I explored the authentication system. Here's what I found:

The app uses JWT tokens with OAuth2 support (Google/GitHub). Security is mostly solid - passwords are properly hashed and CSRF protection is in place.

However, I found two issues:
1. JWT tokens never expire (security risk)
2. No rate limiting on login endpoint (brute-force vulnerability)

Would you like me to fix these security issues?"
```

## Parallel vs Sequential Subagents

### Sequential Pattern

```javascript
// One subagent finishes before next starts
const exploration = await Task({
  prompt: "Explore authentication...",
  subagent_type: "Explore"
})

// Wait for Explore to finish, then use its findings
const plan = await Task({
  prompt: `Based on these findings:\n${exploration}\n\nCreate implementation plan...`,
  subagent_type: "Plan"
})

// Wait for Plan to finish, then execute
const implementation = await Task({
  prompt: `Execute this plan:\n${plan}`,
  subagent_type: "general-purpose"
})
```

**Characteristics**:
- Subagents run one at a time
- Each can build on previous results
- Slower but coordinated

### Parallel Pattern

Claude Code can launch multiple subagents simultaneously:

```javascript
// In a single message, make multiple Task calls:
Task({
  description: "Fix auth bug",
  prompt: "Fix the authentication bug in src/auth/login.js...",
  subagent_type: "general-purpose"
})

Task({
  description: "Fix payment bug",
  prompt: "Fix the payment processing bug in src/payments/charge.js...",
  subagent_type: "general-purpose"
})

Task({
  description: "Fix notification bug",
  prompt: "Fix the notification delivery bug in src/notify/email.js...",
  subagent_type: "general-purpose"
})

// All three subagents run in parallel
// All three return results independently
```

**Characteristics**:
- Subagents run simultaneously
- Can't share information with each other
- Much faster for independent tasks
- Main agent integrates all results

**Critical**: Tasks must be truly independent. If they modify the same files, conflicts will occur.

## Subagent Lifecycle

### Creation

```
1. Task tool invoked
    ↓
2. Claude Code CLI finds agent definition
    ↓
3. Parse YAML frontmatter
    ↓
4. Initialize new Claude API conversation
   - Set model (haiku/sonnet/opus)
   - Set temperature
   - Configure prompt caching
    ↓
5. Build system prompt (agent instructions + base)
    ↓
6. Send first user message (your prompt parameter)
    ↓
7. Subagent begins thinking
```

### Execution

```
Subagent runs autonomously:
    ↓
1. Receives task from prompt
    ↓
2. Plans approach
    ↓
3. Uses available tools (Read, Grep, etc.)
    ↓
4. Builds understanding through tool results
    ↓
5. Synthesizes findings
    ↓
6. Generates final response
```

**During execution**:
- Subagent maintains its own conversation history
- Tool calls/results accumulate in subagent's context
- Main agent has no visibility into progress
- Main agent is waiting for final result

### Completion

```
Subagent finishes:
    ↓
1. Returns final message
    ↓
2. Claude Code captures response
    ↓
3. Returns to main agent as Task result
    ↓
4. Subagent conversation is discarded
    ↓
5. Context/history is lost (unless logged)
```

**Implications**:
- Can't resume a subagent after completion
- Can't ask subagent follow-up questions
- Must spawn new subagent for additional work
- Main agent must extract important info from result

## Advanced Patterns

### Chaining with Context Accumulation

```javascript
// Stage 1: Exploration
const findings = await Task({
  prompt: "Explore the codebase and identify all payment-related modules...",
  subagent_type: "Explore"
})

// Stage 2: Planning (includes Stage 1 output)
const plan = await Task({
  prompt: `Based on these modules:\n${findings}\n\nCreate a refactoring plan to consolidate payment logic...`,
  subagent_type: "Plan"
})

// Stage 3: Execution (includes Stage 2 output)
const result = await Task({
  prompt: `Execute this refactoring plan:\n${plan}\n\nMake the code changes and run tests.`,
  subagent_type: "general-purpose"
})
```

**Pattern**: Each stage receives cumulative context from previous stages.

**Limitation**: Context grows with each stage. Can hit token limits after 3-4 stages.

### Parallel Swarm with Aggregation

```javascript
// Launch parallel exploration
Task({
  description: "Explore auth",
  prompt: "Explore authentication system...",
  subagent_type: "Explore"
})

Task({
  description: "Explore payments",
  prompt: "Explore payment processing...",
  subagent_type: "Explore"
})

Task({
  description: "Explore notifications",
  prompt: "Explore notification delivery...",
  subagent_type: "Explore"
})

// All three return results
// Main agent aggregates findings into unified architecture document
```

**Pattern**: Divide large codebase exploration into parallel zones.

**Benefit**: 3x faster than sequential exploration.

### Specialist Delegation

```javascript
// Security-focused review
const securityIssues = await Task({
  prompt: `Review this authentication code for security issues:

  ${code}

  Check for:
  - SQL injection
  - XSS vulnerabilities
  - Authentication bypasses
  - Insecure token handling`,

  subagent_type: "security-auditor"  // Custom specialist agent
})
```

**Pattern**: Route tasks to specialized agents with domain expertise.

### Recursive Delegation (Advanced)

Some agents can spawn their own subagents:

```yaml
---
name: orchestrator
tools:
  - Read
  - Task  # Can spawn subagents!
---

You are an orchestrator agent. When given complex tasks, break them down and delegate to specialist subagents.
```

**Usage**:
```javascript
// Main agent delegates to orchestrator
const result = await Task({
  prompt: "Refactor the entire authentication system...",
  subagent_type: "orchestrator"
})

// Orchestrator agent internally does:
// - Task({ subagent_type: "Explore" }) to understand current state
// - Task({ subagent_type: "Plan" }) to create refactoring plan
// - Task({ subagent_type: "general-purpose" }) multiple times for implementation
// - Returns aggregated result
```

**Caution**: Can create deep delegation trees. Use sparingly.

## Subagent Design Patterns

### Pattern 1: The Specialist

**Purpose**: Deep expertise in specific domain

```yaml
---
name: sql-optimizer
description: PostgreSQL query performance expert
model: sonnet
tools: [Read, Grep, Glob]
temperature: 0.2
---

You are a PostgreSQL performance specialist. Analyze queries and suggest optimizations.

Your process:
1. Find all SQL queries in codebase
2. Identify N+1 queries and missing indexes
3. Suggest specific index additions with CREATE INDEX statements
4. Estimate performance impact

Output format:
- Query: [file:line]
- Issue: [description]
- Fix: [SQL statement]
- Impact: [estimated improvement]
```

**When to use**: Need deep technical expertise in narrow area.

### Pattern 2: The Explorer

**Purpose**: Comprehensive research and discovery

```yaml
---
name: feature-explorer
description: Maps how a feature works across the codebase
model: sonnet
tools: [Read, Grep, Glob]
temperature: 0.5
---

You explore codebases to understand how features are implemented.

Process:
1. Search for entry points (routes, API endpoints)
2. Trace execution flow through layers
3. Identify data models and schemas
4. Map dependencies and side effects
5. Document the complete flow

Output: Comprehensive feature map with file references.
```

**When to use**: Understanding unfamiliar or complex systems.

### Pattern 3: The Executor

**Purpose**: Autonomous implementation

```yaml
---
name: test-writer
description: Writes comprehensive test suites
model: sonnet
tools: [Read, Write, Bash]
temperature: 0.4
---

You write test suites for code.

Process:
1. Read the implementation code
2. Identify edge cases and error conditions
3. Write comprehensive tests (unit + integration)
4. Run tests to verify they pass
5. Report coverage

Output: Test files + test run results.
```

**When to use**: Automating repetitive but valuable work.

### Pattern 4: The Reviewer

**Purpose**: Quality assurance and validation

```yaml
---
name: security-reviewer
description: Identifies security vulnerabilities
model: opus  # Need best reasoning for security
tools: [Read, Grep, Glob]
temperature: 0.1  # Consistent, thorough
---

You are a security expert reviewing code for vulnerabilities.

CRITICAL CHECKS:
- SQL injection (string concatenation in queries)
- XSS (unescaped user input in templates)
- CSRF (missing tokens in forms)
- Auth bypass (logic errors in permission checks)
- Secrets in code (API keys, passwords)

For each issue:
- Severity: Critical/High/Medium/Low
- Location: file:line
- Exploit: How it could be attacked
- Fix: Specific code change
```

**When to use**: Pre-commit validation, security audits.

## Common Pitfalls

### Pitfall 1: Insufficient Context in Prompt

❌ **Bad**:
```javascript
Task({
  prompt: "Fix the bug",
  subagent_type: "general-purpose"
})
```

**Problem**: Subagent has no idea what bug, where it is, or how to fix it.

✅ **Good**:
```javascript
Task({
  prompt: `Fix the authentication timeout bug.

Location: src/auth/session.js:45-67

Bug: Users are logged out after 5 minutes even with activity.

Expected: Session should extend on activity, timeout only after 30min idle.

Current code:
${codeSnippet}

Requirements:
- Update session timestamp on each request
- Check for idle time, not total time
- Maintain security (no infinite sessions)`,

  subagent_type: "general-purpose"
})
```

**Why better**: Subagent has location, behavior, expectations, and code context.

### Pitfall 2: Wrong Tool Permissions

❌ **Bad**: Read-only researcher agent can't complete task
```yaml
---
name: bug-fixer
tools: [Read, Grep]  # No Edit or Write!
---

You fix bugs in code.
```

**Problem**: Agent is asked to fix bugs but can't modify files.

✅ **Good**:
```yaml
---
name: bug-fixer
tools: [Read, Grep, Glob, Edit, Bash]  # Can modify and test
---

You fix bugs in code.
```

### Pitfall 3: Parallel Tasks That Conflict

❌ **Bad**: Both agents try to modify same file
```javascript
Task({
  prompt: "Refactor src/auth.js to use async/await",
  subagent_type: "general-purpose"
})

Task({
  prompt: "Add logging to src/auth.js",
  subagent_type: "general-purpose"
})
```

**Problem**: Race condition. Both agents read original file, both try to write changes. Second write overwrites first.

✅ **Good**: Sequential or separate files
```javascript
// Option 1: Sequential
const refactor = await Task({
  prompt: "Refactor src/auth.js to use async/await",
  subagent_type: "general-purpose"
})

const logging = await Task({
  prompt: "Add logging to src/auth.js (now using async/await)",
  subagent_type: "general-purpose"
})

// Option 2: Parallel with different files
Task({
  prompt: "Refactor src/auth.js...",
  subagent_type: "general-purpose"
})

Task({
  prompt: "Add logging to src/payments.js...",
  subagent_type: "general-purpose"
})
```

### Pitfall 4: Wrong Model for Task

❌ **Bad**: Using Haiku for complex reasoning
```yaml
---
name: architect
description: Designs system architecture
model: haiku  # Too simple for this task
---
```

**Problem**: Haiku lacks reasoning depth for architectural decisions.

✅ **Good**:
```yaml
---
name: architect
description: Designs system architecture
model: opus  # Need deep reasoning
temperature: 0.6
---
```

### Pitfall 5: Over-Delegation

❌ **Bad**: Spawning subagent for trivial task
```javascript
// Main agent could easily do this:
const result = await Task({
  prompt: "Read src/config.js and tell me the port number",
  subagent_type: "Explore"
})
```

**Problem**: Overhead of subagent (new API call, context setup, latency) for 2-second task.

✅ **Good**: Main agent does simple tasks directly
```javascript
// Just read the file directly
const config = Read({ file_path: "src/config.js" })
// Parse port from config
```

**Rule of thumb**: Use subagents for tasks that take >5 tool calls or need specialized expertise.

## Debugging Subagents

### Problem: Subagent returns incomplete results

**Symptoms**: Subagent response is vague or missing requested information.

**Diagnosis**:
1. Check if prompt was specific enough
2. Check if subagent had necessary tools
3. Check if files/paths existed

**Fix**:
```javascript
// Add explicit requirements checklist
Task({
  prompt: `Explore authentication system.

Required information:
☐ List of all auth-related files
☐ Authentication flow diagram
☐ Security assessment
☐ Technology stack used

Do not return until all items are complete.`,

  subagent_type: "Explore"
})
```

### Problem: Subagent uses wrong approach

**Symptoms**: Agent does something different than expected.

**Diagnosis**: Agent's system prompt may conflict with your task.

**Fix**: Override behavior in prompt:
```javascript
Task({
  prompt: `Analyze performance issues.

IMPORTANT: Do not modify code. Only analyze and report.
Even though you have Edit permissions, this is read-only analysis.

Provide:
- List of slow functions
- Specific bottlenecks
- Optimization suggestions (don't implement)`,

  subagent_type: "general-purpose"
})
```

### Problem: Subagent times out or runs too long

**Symptoms**: Task never completes or takes excessive time.

**Diagnosis**: Task too broad or complex.

**Fix**: Break into smaller tasks:
```javascript
// Instead of one huge task:
// ❌ Task({ prompt: "Refactor entire codebase..." })

// Break into modules:
✅ Task({ prompt: "Refactor auth module..." })
✅ Task({ prompt: "Refactor payment module..." })
✅ Task({ prompt: "Refactor notification module..." })
```

## Integration with Other Systems

### Subagents + Skills

Skills can recommend when to use subagents:

```markdown
# in superpowers:code-review skill

When you complete implementation of a major feature step:
1. Use the code-reviewer subagent to validate the work
2. Address any issues found
3. Only then declare the step complete

Example:
Task({
  prompt: "Review the implementation of user authentication...",
  subagent_type: "code-reviewer"
})
```

### Subagents + Hooks

Hooks can trigger on subagent tool use:

```json
{
  "hooks": {
    "PostToolUse": {
      "command": "echo 'Subagent modified file: $FILE'",
      "match": {
        "tool": "Edit",
        "isSubagent": true
      }
    }
  }
}
```

### Subagents + MCP Tools

Subagents can use MCP tools for specialized capabilities:

```yaml
---
name: browser-tester
description: Tests web interfaces
tools:
  - Read
  - mcp__playwright__browser_navigate
  - mcp__playwright__browser_click
  - mcp__playwright__browser_snapshot
---

You test web applications using browser automation.

Process:
1. Navigate to application
2. Perform user flows
3. Validate behavior
4. Report bugs
```

## Performance Considerations

### Cost

Each subagent invocation is a separate API call:

```
Main agent cost: ~$0.50 for session
Subagent 1 (Explore): ~$0.30
Subagent 2 (Plan): ~$0.20
Subagent 3 (Implement): ~$0.80

Total: $1.80 vs $0.50 without subagents
```

**When worth it**:
- Parallel work saves time
- Fresh context prevents errors
- Specialized expertise improves quality
- Alternative would exceed token limits

**When not worth it**:
- Simple tasks main agent can do
- Minimal cost/time savings
- Context continuity is critical

### Latency

Subagents add sequential latency:

```
Sequential: 3 subagents × 30 seconds each = 90 seconds

Parallel: 3 subagents × 30 seconds max = 30 seconds
```

**Optimization**: Use parallel Task calls for independent work.

### Token Limits

Each subagent has separate context window:

```
Main agent: 200k tokens used ← Near limit!

Subagent: 0 tokens used ← Fresh start!
```

**When critical**: Main agent approaching 200k token budget. Offload research to subagent to preserve main context for synthesis.

## Best Practices Summary

### DO

✅ Provide detailed, specific prompts
✅ Include all necessary context in prompt
✅ Give subagents appropriate tool permissions
✅ Choose right model for task complexity
✅ Use parallel subagents for independent work
✅ Extract important findings for main context
✅ Specialize subagents for narrow domains
✅ Time-box tasks to prevent runaway agents

### DON'T

❌ Assume subagent knows main context
❌ Give excessive tool permissions
❌ Use subagents for trivial tasks
❌ Run parallel tasks that modify same files
❌ Forget to integrate subagent results
❌ Over-delegate simple work
❌ Create overly generic subagents
❌ Exceed budget spawning many subagents

## Summary

Subagents in Claude Code are:

1. **Isolated Claude instances** - Separate context windows
2. **Configured via YAML + Markdown** - System prompt from agent file
3. **Invoked via Task tool** - With subagent_type parameter
4. **Tool-restricted** - Only access approved tools
5. **Model-flexible** - Can use different Claude models
6. **Context-isolated** - Don't see main conversation
7. **Prompt-dependent** - Only know what's in task prompt
8. **One-shot** - Return result then context is discarded

**Core workflow**:
```
Main Agent → Task({ prompt, subagent_type }) → Subagent → Result → Main Agent
```

**Key to success**: Comprehensive prompts that give subagents everything they need.

---

**See Also:**
- 05-SUBAGENT-SYSTEM.md - Subagent basics and usage
- 14-DECISION-FRAMEWORK.md - When to use subagents
- 17-CONTEXT-MANAGEMENT.md - Context isolation benefits
- 25-PROMPTING-STRATEGIES.md - Crafting effective prompts
