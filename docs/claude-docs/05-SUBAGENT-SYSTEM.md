# Subagent System

## Overview

Subagents are isolated instances of Claude with separate context windows, each configured for specific tasks. They enable parallel work, specialized expertise, and context management in Claude Code.

## What Are Subagents?

A subagent is a distinct Claude instance with:
- **Isolated context window**: Separate conversation history and working memory
- **Custom system prompt**: Specialized instructions and behavioral guidelines
- **Specific tool access**: Controlled permissions for file operations, shell access, etc.
- **Model selection**: Can use different Claude models (Sonnet, Opus, Haiku)
- **Focused responsibility**: Designed for particular types of tasks

Think of subagents as specialized team members - each has a specific role and expertise area.

## Architecture

### File Structure

Subagents are Markdown files with YAML frontmatter:

```markdown
---
name: code-reviewer
description: Reviews code for bugs, security issues, and best practices
model: claude-sonnet-4-5-20250929
tools:
  - Read
  - Grep
  - Glob
temperature: 0.3
---

You are a meticulous code reviewer focused on:
- Security vulnerabilities
- Logic errors and edge cases
- Performance issues
- Code maintainability

When reviewing code:
1. Read the entire implementation first
2. Check against requirements
3. Look for edge cases
4. Verify error handling
5. Provide specific, actionable feedback

Always cite line numbers and file paths in your feedback.
```

### Frontmatter Fields

| Field | Required | Description |
|-------|----------|-------------|
| `name` | Yes | Unique identifier (lowercase, hyphens) |
| `description` | Yes | One-line summary shown to main agent |
| `model` | No | Claude model ID (default: inherits from session) |
| `tools` | No | Array of allowed tool names |
| `temperature` | No | 0.0-1.0, controls randomness (default: 1.0) |
| `maxTokens` | No | Maximum response length |
| `cached` | No | Whether to use prompt caching (default: true) |

### Storage Locations

Subagents are discovered in this order:

1. **User-global**: `~/.claude/agents/`
   - Personal subagents available across all projects
   - Example: `~/.claude/agents/security-auditor.md`

2. **Project-local**: `.claude/agents/`
   - Project-specific subagents
   - Example: `.claude/agents/database-expert.md`

3. **Plugin-provided**: `~/.claude/plugins/{plugin}/agents/`
   - Distributed with Claude Code plugins
   - Example: `~/.claude/plugins/superpowers/agents/code-reviewer.md`

Later locations override earlier ones if names conflict.

## Built-in Subagents

Claude Code ships with several subagents:

### 1. Explore
**Purpose**: Deep codebase investigation and research

**When to use**:
- Analyzing unfamiliar code
- Understanding complex architectures
- Finding patterns across many files
- Researching implementation details

**Tools**: Read, Grep, Glob, Bash (read-only)

**Characteristics**:
- Thorough and methodical
- Generates detailed reports
- Good at connecting disparate pieces
- Returns to main agent with findings

### 2. Plan
**Purpose**: Breaking down complex tasks into actionable steps

**When to use**:
- Large features requiring multiple changes
- Architectural decisions
- Migration planning
- Multi-step refactoring

**Tools**: Read, Grep, Glob

**Characteristics**:
- Strategic thinking
- Creates ordered task lists
- Identifies dependencies
- Estimates complexity

### 3. General-purpose
**Purpose**: Parallel execution of independent tasks

**When to use**:
- Multiple unrelated bugs
- Batch processing
- Independent feature work
- Non-conflicting file changes

**Tools**: Full toolset

**Characteristics**:
- Autonomous execution
- Works in parallel with other agents
- Reports back on completion
- Handles failures gracefully

### 4. Code-reviewer
**Purpose**: Quality assurance and code review

**When to use**:
- Before merging branches
- After implementing features
- Security audits
- Performance reviews

**Tools**: Read, Grep, Glob

**Characteristics**:
- Critical but constructive
- Checks against requirements
- Identifies edge cases
- Provides specific feedback

## Automatic Delegation

Claude automatically delegates to subagents when:

1. **User explicitly requests**: "Explore the authentication system"
2. **Task matches subagent expertise**: Security questions → security-auditor
3. **Parallel work is beneficial**: Multiple independent bugs
4. **Context window is full**: Offload research to fresh context
5. **Deep investigation needed**: Complex analysis requiring focus

### Delegation Flow

```
Main Agent
    │
    ├─ Recognizes task type
    │
    ├─ Selects appropriate subagent
    │
    ├─ Provides context and instructions
    │
    └─ Spawns subagent instance
            │
            ├─ Subagent works independently
            │
            ├─ Maintains separate context
            │
            └─ Reports findings back
                    │
                    └─ Main agent integrates results
```

## Creating Custom Subagents

### Example: Database Expert

```markdown
---
name: database-expert
description: PostgreSQL query optimization and schema design specialist
model: claude-sonnet-4-5-20250929
tools:
  - Read
  - Grep
  - Glob
  - Bash
temperature: 0.2
---

You are a PostgreSQL database expert specializing in:
- Query optimization and EXPLAIN analysis
- Index strategy and performance tuning
- Schema design and normalization
- Migration safety and data integrity

## Your Process

1. **Understand the schema**: Read migration files and models
2. **Analyze queries**: Look for N+1 problems, missing indexes, inefficient joins
3. **Check indexes**: Verify coverage and identify redundant indexes
4. **Review constraints**: Ensure data integrity at database level
5. **Suggest improvements**: Specific, actionable recommendations with SQL

## Guidelines

- Always use EXPLAIN ANALYZE when discussing query performance
- Consider both read and write patterns
- Think about connection pooling and transaction boundaries
- Warn about breaking changes in migrations
- Provide before/after comparisons for optimizations

## Output Format

For each issue found:
- File path and line number
- Current implementation
- Problem explanation
- Recommended solution
- Expected performance impact
```

### Example: API Documentation Generator

```markdown
---
name: api-doc-generator
description: Generates OpenAPI/Swagger documentation from code
model: claude-sonnet-4-5-20250929
tools:
  - Read
  - Grep
  - Glob
  - Write
temperature: 0.4
---

You are an API documentation specialist who generates comprehensive OpenAPI 3.0 specifications.

## Your Task

1. **Discover endpoints**: Find route definitions across the codebase
2. **Extract schemas**: Identify request/response models
3. **Document parameters**: Query params, path params, headers
4. **Add examples**: Realistic request/response examples
5. **Generate OpenAPI**: Valid OpenAPI 3.0 YAML

## Analysis Process

For each endpoint:
- HTTP method and path
- Authentication requirements
- Request body schema (if applicable)
- Response schemas (success and error cases)
- Query parameters and validation rules
- Description of what the endpoint does

## Output

Generate a complete `openapi.yaml` file with:
- Server configuration
- Security schemes
- All paths and operations
- Reusable components/schemas
- Meaningful descriptions and examples

Validate the output is valid OpenAPI 3.0 before writing the file.
```

## Tool Access Configuration

### Available Tools

Subagents can access any tool the main agent has:

**File Operations**:
- `Read`: Read file contents
- `Write`: Create new files
- `Edit`: Modify existing files
- `Glob`: Find files by pattern

**Search**:
- `Grep`: Search file contents

**Execution**:
- `Bash`: Run shell commands

**Git**:
- Usually inherited via Bash access

**MCP Tools**:
- Any MCP server tools (prefix: `mcp__`)

### Permission Patterns

**Read-only researcher**:
```yaml
tools:
  - Read
  - Grep
  - Glob
```

**Code modifier**:
```yaml
tools:
  - Read
  - Edit
  - Grep
  - Glob
```

**Full autonomy**:
```yaml
tools:
  - Read
  - Write
  - Edit
  - Grep
  - Glob
  - Bash
```

**Specialized API access**:
```yaml
tools:
  - Read
  - mcp__browser__navigate
  - mcp__browser__click
  - mcp__browser__extract
```

### Security Considerations

- **Principle of least privilege**: Only grant necessary tools
- **Bash access**: Powerful but dangerous - use carefully
- **Write permissions**: Can create files anywhere
- **MCP tools**: May access external services or data
- **Model selection**: Higher-capability models have more reasoning but cost more

## Model Selection

Different tasks benefit from different models:

### Claude Sonnet 4.5 (Default)
- **Best for**: General tasks, coding, analysis
- **Characteristics**: Fast, capable, cost-effective
- **Use when**: Most situations

### Claude Opus 3.5
- **Best for**: Complex reasoning, architecture decisions
- **Characteristics**: Deepest thinking, slower, expensive
- **Use when**: Strategic planning, difficult debugging

### Claude Haiku 3.5
- **Best for**: Simple repetitive tasks, quick checks
- **Characteristics**: Very fast, inexpensive, less capable
- **Use when**: Linting, formatting, simple validations

Example configuration:
```yaml
model: claude-opus-3-5-20250929  # Use Opus for complex reasoning
temperature: 0.3                  # Lower temperature for consistency
```

## Best Practices

### 1. Focused Responsibility

**Good**: Single, well-defined purpose
```markdown
---
name: security-auditor
description: Identifies security vulnerabilities in authentication code
---
```

**Bad**: Vague or overlapping responsibilities
```markdown
---
name: helper
description: Does various coding tasks
---
```

### 2. Detailed Instructions

**Good**: Specific process and expectations
```markdown
You analyze authentication code for security issues.

Process:
1. Check password storage (bcrypt, proper salting)
2. Verify JWT token handling (expiry, validation)
3. Look for SQL injection vulnerabilities
4. Check authorization logic (IDOR, privilege escalation)
5. Review session management

Output format:
- Severity: Critical/High/Medium/Low
- Location: File:Line
- Issue: Description
- Fix: Specific code change
```

**Bad**: Generic instructions
```markdown
You review code and find problems.
```

### 3. Clear Tool Selection

**Good**: Minimal necessary permissions
```yaml
tools:
  - Read      # Need to read code
  - Grep      # Need to search for patterns
  - Glob      # Need to find files
```

**Bad**: Excessive permissions
```yaml
tools:
  - Read
  - Write
  - Edit
  - Bash
  - Grep
  - Glob
# Why does a read-only auditor need write access?
```

### 4. Appropriate Temperature

| Task Type | Temperature | Reasoning |
|-----------|-------------|-----------|
| Code review | 0.2-0.3 | Consistent, thorough |
| Bug fixing | 0.3-0.5 | Focused but creative |
| Planning | 0.5-0.7 | Strategic thinking |
| Documentation | 0.4-0.6 | Clear but engaging |
| Exploration | 0.7-0.9 | Creative connections |

### 5. Context Management

Subagents have limited context. Help them by:

**Providing key information**:
```
"Review the authentication changes in PR #123.
Focus on files: src/auth/*.ts
Requirements: Must support OAuth2 and JWT"
```

**Setting clear boundaries**:
```
"Analyze ONLY the payment processing module.
Do not review unrelated code."
```

**Defining success criteria**:
```
"Your review is complete when you've checked:
- All edge cases have tests
- Error handling is present
- Security best practices are followed"
```

## Advanced Patterns

### Chained Subagents

Main agent delegates to Explorer → Explorer's findings feed Planner → Planner's tasks executed by general-purpose agents.

### Parallel Swarms

Spawn multiple general-purpose agents for independent tasks:
- Agent A: Fix bug in auth module
- Agent B: Fix bug in payment module
- Agent C: Fix bug in notification module

All work simultaneously, report back independently.

### Specialized Pipelines

1. **Code-generator** creates initial implementation
2. **Test-writer** adds comprehensive tests
3. **Code-reviewer** audits both
4. Main agent integrates feedback

### Recursive Delegation

Subagents can spawn their own subagents (if configured):
```yaml
tools:
  - SubagentSpawn  # Allow this agent to create subagents
```

Use sparingly - can create complex dependency chains.

## Debugging Subagents

### Common Issues

**Subagent not found**:
- Check file is in correct location
- Verify filename matches name in frontmatter
- Ensure `.md` extension

**Wrong tools available**:
- Review `tools` array in frontmatter
- Check tool names are exact (case-sensitive)
- Verify MCP servers are running (for MCP tools)

**Unexpected behavior**:
- Review system prompt clarity
- Adjust temperature
- Try different model
- Add more specific instructions

### Testing Subagents

1. **Direct invocation**: Test by directly calling the subagent
2. **Isolated tasks**: Give simple, well-defined tasks first
3. **Iterative refinement**: Improve prompt based on results
4. **Tool verification**: Confirm it can access required tools
5. **Edge cases**: Test boundary conditions and failures

## Integration with Other Systems

### Subagents + Hooks
Hooks can trigger when subagents use tools:
```json
{
  "hooks": {
    "PostToolUse": {
      "command": "echo 'Subagent used a tool'",
      "match": {
        "isSubagent": true
      }
    }
  }
}
```

### Subagents + Skills
Skills can invoke subagents as part of their workflow:
```markdown
When you need deep analysis, spawn the Explore subagent.
```

### Subagents + MCP
Subagents can use MCP tools just like main agent:
```yaml
tools:
  - Read
  - mcp__playwright__browser_navigate
```

## When to Create a Subagent

### Create a Subagent When:
- Task requires deep, focused expertise
- You need parallel execution
- Context isolation is beneficial
- Consistent process should be followed
- Specific tool permissions needed

### Use Main Agent When:
- Task is simple and quick
- Context continuity is important
- Interactive back-and-forth needed
- One-off or exploratory work

### Use a Skill When:
- Reusable process/checklist
- Lightweight guidance needed
- No separate context required
- Main agent can handle with instructions

## Summary

Subagents are powerful tools for:
- **Specialization**: Focused expertise areas
- **Parallelization**: Multiple tasks simultaneously
- **Context management**: Isolating complex work
- **Consistency**: Enforcing specific processes

Key to effective subagents:
1. Clear, focused purpose
2. Detailed system prompts
3. Appropriate tool access
4. Right model for the task
5. Well-defined success criteria

With thoughtful design, subagents become force multipliers, enabling Claude Code to handle increasingly complex workflows.
