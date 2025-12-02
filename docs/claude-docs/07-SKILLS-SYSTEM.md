# Skills System

## Overview

Skills are reusable, model-invoked capabilities that extend Claude's knowledge and processes. They're like reference manuals or procedural guides that Claude can consult when needed.

## What Are Skills?

A skill is a specialized knowledge module that:

- **Lives as a Markdown file**: Human-readable documentation with metadata
- **Teaches Claude**: Provides domain knowledge, processes, or best practices
- **Is discovered automatically**: Claude finds and activates relevant skills
- **Extends capabilities**: Adds expertise without modifying core prompts
- **Can include supporting files**: Scripts, templates, configuration examples

Think of skills as expertise modules - each one makes Claude better at a specific type of task.

## Architecture

### File Structure

Skills are Markdown files with YAML frontmatter:

```markdown
---
name: test-driven-development
description: Write tests before implementation, ensuring tests fail first before writing code
tags:
  - testing
  - tdd
  - methodology
activationPatterns:
  - "write.*test"
  - "implement.*feature"
  - "add.*functionality"
model: claude-sonnet-4-5-20250929
---

# Test-Driven Development (TDD)

## The Process

1. **Write a failing test**: Create a test that defines desired functionality
2. **Run the test**: Confirm it fails for the right reason
3. **Write minimal code**: Only enough to make the test pass
4. **Run the test again**: Verify it passes
5. **Refactor**: Improve code while keeping tests green
6. **Repeat**: Move to next piece of functionality

## Rules

- NEVER write implementation before writing the test
- ONLY write enough code to make the test pass
- ALWAYS run tests before claiming success
- REFACTOR continuously while tests remain green

## Example

```typescript
// 1. Write the test first
describe('calculateTotal', () => {
  it('sums array of numbers', () => {
    expect(calculateTotal([1, 2, 3])).toBe(6);
  });
});

// 2. Run test - it fails (calculateTotal doesn't exist)

// 3. Write minimal implementation
function calculateTotal(numbers: number[]): number {
  return numbers.reduce((sum, n) => sum + n, 0);
}

// 4. Run test - it passes

// 5. Refactor if needed (already simple)
```

## When to Apply

Apply TDD when:
- Implementing new features
- Fixing bugs (write test that reproduces bug first)
- Refactoring (tests verify behavior preservation)
- Working on critical code (authentication, payments, etc.)

Skip TDD when:
- Prototyping UI layouts
- Exploring unfamiliar APIs
- Writing throwaway code

## Common Mistakes

- Writing test after implementation (defeats the purpose)
- Not running test to see it fail first
- Writing too much implementation at once
- Skipping refactoring step
- Testing implementation details instead of behavior
```

### Frontmatter Fields

| Field | Required | Description |
|-------|----------|-------------|
| `name` | Yes | Unique identifier (lowercase, hyphens) |
| `description` | Yes | One-line summary for discovery |
| `tags` | No | Categories for organization and search |
| `activationPatterns` | No | Regex patterns that trigger skill |
| `model` | No | Preferred Claude model for this skill |
| `priority` | No | 1-10, higher loads first (default: 5) |
| `dependencies` | No | Other skill names required |
| `version` | No | Semantic version (e.g., "1.0.0") |

### Storage Locations

Skills are discovered in this order:

1. **Plugin-provided**: `~/.claude/plugins/{plugin}/skills/`
   - Distributed with Claude Code plugins
   - Example: `~/.claude/plugins/superpowers/skills/systematic-debugging.md`

2. **User-global**: `~/.claude/skills/`
   - Personal skills available across all projects
   - Example: `~/.claude/skills/database-optimization.md`

3. **Project-local**: `.claude/skills/`
   - Project-specific skills and conventions
   - Example: `.claude/skills/api-versioning.md`

Later locations override earlier ones if names conflict.

## How Skills Work

### Discovery Process

1. **Scan**: Claude Code scans all skill directories at session start
2. **Parse**: Extracts frontmatter metadata from each `.md` file
3. **Index**: Builds searchable index by name, description, tags
4. **Monitor**: Watches for new skills added during session

### Activation Mechanisms

Skills can be activated in four ways:

#### 1. User Invokes Directly

User uses the Skill tool:
```
Use the test-driven-development skill to implement the login feature.
```

Claude loads and applies the skill.

#### 2. Automatic Pattern Matching

User message matches `activationPatterns`:
```
User: "Write tests for the authentication module"
```

Regex `write.*test` matches → TDD skill loads automatically.

#### 3. Claude Discovers Relevance

Claude searches skills when user request matches tags/description:
```
User: "Optimize the slow database queries"
```

Claude finds skill tagged `database`, `performance` → loads it.

#### 4. Dependency Loading

Skill A declares dependency on skill B:
```yaml
dependencies:
  - test-driven-development
```

When A loads, B loads automatically.

### Skill Context

When activated, skill content is added to Claude's context:

```
<skill name="test-driven-development">
[Full skill markdown content]
</skill>
```

Claude now has access to all knowledge and processes in the skill.

## Creating Effective Skills

### 1. Clear, Focused Purpose

**Good**: Single, well-defined domain
```yaml
name: api-error-handling
description: Standardized error handling for REST APIs
```

**Bad**: Vague or overlapping
```yaml
name: best-practices
description: Various coding tips
```

### 2. Actionable Content

**Good**: Specific steps and examples
```markdown
## Process

1. Define error types:
   - ValidationError (400)
   - AuthenticationError (401)
   - NotFoundError (404)
   - ServerError (500)

2. Create error classes:
```typescript
class APIError extends Error {
  statusCode: number;
  constructor(message: string, statusCode: number) {
    super(message);
    this.statusCode = statusCode;
  }
}
```

3. Use in handlers:
```typescript
if (!user) {
  throw new APIError('User not found', 404);
}
```
```

**Bad**: Generic advice
```markdown
Handle errors properly.
Use appropriate status codes.
Log errors.
```

### 3. Relevant Examples

Include code examples in the project's language/framework:

```markdown
## Example (Express.js)

```javascript
app.use((err, req, res, next) => {
  if (err instanceof APIError) {
    res.status(err.statusCode).json({
      error: err.message,
      code: err.statusCode
    });
  } else {
    res.status(500).json({
      error: 'Internal server error'
    });
  }
});
```
```

### 4. Decision Criteria

Help Claude know when to apply:

```markdown
## When to Apply

Use this pattern when:
- Building REST APIs with multiple endpoints
- Need consistent error format for clients
- Want centralized error logging
- Errors should map to HTTP status codes

Don't use when:
- Building GraphQL APIs (use GraphQL error format)
- Internal microservices (use domain exceptions)
- Command-line tools (use exit codes)
```

### 5. Anti-patterns

Show what NOT to do:

```markdown
## Common Mistakes

### ❌ Catching and ignoring
```javascript
try {
  dangerousOperation();
} catch (e) {
  // Silent failure
}
```

### ✅ Catch, log, re-throw
```javascript
try {
  dangerousOperation();
} catch (e) {
  logger.error('Operation failed', { error: e });
  throw new APIError('Operation failed', 500);
}
```
```

## Activation Patterns

Regex patterns that trigger automatic activation:

### Literal Matching

```yaml
activationPatterns:
  - "debug"
  - "troubleshoot"
```

Matches messages containing those exact words.

### Wildcard Matching

```yaml
activationPatterns:
  - "write.*test"
  - "add.*feature"
  - "implement.*"
```

Matches phrases with any words between.

### Complex Patterns

```yaml
activationPatterns:
  - "(create|build|implement).*(api|endpoint)"
  - "optimize.*(query|database|performance)"
  - "fix.*(bug|issue|error)"
```

Matches multiple variations.

### Case Sensitivity

Patterns are case-insensitive by default:
- "Debug" matches
- "DEBUG" matches
- "debug" matches

### Best Practices

1. **Be specific**: Avoid overly broad patterns like `".*"`
2. **Test patterns**: Use regex testers to verify
3. **Multiple patterns**: Cover variations of same intent
4. **Avoid conflicts**: Don't overlap with other skills
5. **Document reasoning**: Comment why pattern triggers skill

## Supporting Files

Skills can reference external files:

### Directory Structure

```
.claude/skills/
├── api-design/
│   ├── SKILL.md              # Main skill file
│   ├── openapi-template.yaml # Template file
│   ├── examples/
│   │   ├── rest-api.ts
│   │   └── graphql-api.ts
│   └── scripts/
│       └── generate-client.sh
```

### Referencing Files

In skill markdown:
```markdown
## OpenAPI Template

Use this template for new APIs:

```yaml
@api-design/openapi-template.yaml
```

## Example Implementation

```typescript
@api-design/examples/rest-api.ts
```

## Generate Client

Run this script after updating the OpenAPI spec:

```bash
@api-design/scripts/generate-client.sh
```
```

Claude can read these files when needed.

### File Types

**Templates**:
- Configuration files
- Boilerplate code
- Directory structures

**Examples**:
- Working code samples
- Complete implementations
- Before/after comparisons

**Scripts**:
- Automation tools
- Generators
- Validators

**Documentation**:
- Additional reading
- Specifications
- Diagrams

## Built-in Skills

Claude Code plugins provide many skills:

### From Superpowers Plugin

**systematic-debugging**:
- Four-phase debugging framework
- Root cause investigation
- Hypothesis testing
- Implementation with verification

**test-driven-development**:
- Red-Green-Refactor cycle
- Test-first methodology
- Refactoring with safety

**verification-before-completion**:
- Run verification commands
- Confirm output before claiming success
- Evidence-based assertions

**brainstorming**:
- Socratic method questioning
- Alternative exploration
- Incremental validation

**condition-based-waiting**:
- Replace timeouts with condition polling
- Eliminate flaky tests
- Wait for actual state changes

**defense-in-depth**:
- Multi-layer validation
- Prevent invalid data propagation
- Structural impossibility of bugs

**root-cause-tracing**:
- Trace errors backward through call stack
- Systematic instrumentation
- Find original invalid data source

### Creating Custom Skills

Example: Code Review Checklist

```markdown
---
name: code-review-checklist
description: Comprehensive code review process for pull requests
tags:
  - code-review
  - quality
  - pr
activationPatterns:
  - "review.*(pr|pull request|code)"
  - "before.*(merge|merging)"
---

# Code Review Checklist

## Automated Checks

Before manual review, verify:
- [ ] CI/CD pipeline passes
- [ ] All tests pass locally
- [ ] Code coverage maintained or improved
- [ ] No linting errors
- [ ] Types check (TypeScript)
- [ ] Security scan passes

## Functionality

- [ ] Code does what PR description claims
- [ ] Edge cases handled
- [ ] Error cases handled
- [ ] Input validation present
- [ ] No obvious bugs

## Code Quality

- [ ] Names are clear and descriptive
- [ ] Functions are focused (single responsibility)
- [ ] No code duplication
- [ ] Comments explain WHY, not WHAT
- [ ] Complex logic has explanations

## Testing

- [ ] New code has tests
- [ ] Tests cover happy path
- [ ] Tests cover edge cases
- [ ] Tests cover error cases
- [ ] Test names describe behavior
- [ ] No brittle tests (testing implementation details)

## Security

- [ ] No hardcoded secrets
- [ ] User input sanitized
- [ ] SQL injection prevented (parameterized queries)
- [ ] XSS prevented (escaped output)
- [ ] Authentication/authorization correct
- [ ] Sensitive data encrypted

## Performance

- [ ] No N+1 queries
- [ ] Database indexes appropriate
- [ ] No unnecessary loops
- [ ] Large data sets handled efficiently
- [ ] No memory leaks

## Documentation

- [ ] API changes documented
- [ ] Breaking changes noted
- [ ] README updated if needed
- [ ] Comments added for complex logic
- [ ] Migration guide if needed

## Process

1. Run automated checks first
2. Read PR description and requirements
3. Review code file by file
4. Check tests thoroughly
5. Look for security issues
6. Consider performance
7. Verify documentation
8. Leave specific, actionable feedback
9. Approve only if all critical items pass

## Feedback Format

```markdown
**File**: src/auth/login.ts
**Line**: 42
**Severity**: High
**Issue**: Password compared with == instead of secure comparison
**Fix**: Use `bcrypt.compare(input, hash)` instead
```
```

## Skills vs. Slash Commands vs. Subagents

### Skills
**What**: Knowledge and processes embedded in context
**When**: Teaching Claude how to approach tasks
**Activation**: Automatic or manual via Skill tool
**Context**: Added to current conversation
**Example**: TDD methodology, code review checklist

### Slash Commands
**What**: Shortcuts that expand to prompts
**When**: Frequently used instructions or questions
**Activation**: User types `/command`
**Context**: Replaces with prompt text
**Example**: `/brainstorm`, `/execute-plan`

### Subagents
**What**: Separate Claude instances with specialized roles
**When**: Need isolation, parallelization, or specialized expertise
**Activation**: Main agent delegates
**Context**: New, separate context window
**Example**: Deep codebase exploration, code review

### Choosing the Right Tool

| Need | Use |
|------|-----|
| Teach a methodology | Skill |
| Provide reference information | Skill |
| Shortcut for common prompts | Slash command |
| Complex multi-step process | Slash command + Skill |
| Parallel independent work | Subagent |
| Deep focused investigation | Subagent |
| Isolated context needed | Subagent |

### Combining Them

Powerful workflows combine all three:

```markdown
User: /brainstorm

Slash command expands to brainstorming prompt
        ↓
Brainstorming skill activates (provides Socratic method)
        ↓
User refines idea with Claude's questions
        ↓
User: /execute-plan

Slash command expands to execution prompt
        ↓
Claude spawns subagents for parallel tasks
        ↓
Each subagent has relevant skills loaded
```

## Advanced Patterns

### Skill Composition

Skills can build on each other:

**Base skill**: `testing-fundamentals.md`
```yaml
name: testing-fundamentals
description: Core testing principles
```

**Specialized skill**: `testing-react-components.md`
```yaml
name: testing-react-components
description: React-specific testing patterns
dependencies:
  - testing-fundamentals
```

When React skill loads, fundamentals load too.

### Conditional Skills

Skills that only apply in specific contexts:

```markdown
---
name: production-deployment
description: Deployment checklist for production releases
activationPatterns:
  - "deploy.*production"
  - "release.*prod"
---

## Pre-Deployment Checks

Only proceed if:
- Current branch is `main`
- All tests pass
- Security scan clean
- Staging deployment successful
- Database migrations tested
- Rollback plan documented
```

### Versioned Skills

Track skill evolution:

```yaml
name: api-versioning
description: API versioning strategy
version: 2.0.0
```

Projects can specify required version:
```yaml
requiredSkills:
  - name: api-versioning
    minVersion: 2.0.0
```

### Dynamic Skills

Skills that generate content:

```markdown
---
name: changelog-generator
description: Generates changelog from git commits
---

## Process

1. Read git commits since last release:
```bash
git log v1.0.0..HEAD --oneline
```

2. Categorize commits:
   - Features: commits starting with "feat:"
   - Fixes: commits starting with "fix:"
   - Breaking: commits with "BREAKING CHANGE:"

3. Generate markdown:
```markdown
# Changelog

## [2.0.0] - 2025-12-01

### Features
- Add user authentication
- Implement payment processing

### Fixes
- Fix memory leak in websocket handler

### Breaking Changes
- API now requires authentication header
```
```

## Best Practices

### 1. Single Responsibility

Each skill should teach ONE thing well.

**Good**: `database-indexing.md` focuses on index strategy
**Bad**: `database-everything.md` tries to cover all database topics

### 2. Concrete Examples

Always include working code examples.

**Good**:
```typescript
// ✅ Good
async function getUser(id: string): Promise<User> {
  return await db.user.findUnique({ where: { id } });
}
```

**Bad**:
```
Query the database using the ORM.
```

### 3. Progressive Detail

Start simple, then add complexity:

```markdown
## Basic Usage

```typescript
const result = await api.get('/users');
```

## With Error Handling

```typescript
try {
  const result = await api.get('/users');
} catch (error) {
  if (error.status === 404) {
    // Handle not found
  }
}
```

## Production-Ready

```typescript
async function fetchUsers(retries = 3): Promise<User[]> {
  try {
    const result = await api.get('/users', {
      timeout: 5000,
      validateStatus: (status) => status < 500
    });
    return result.data;
  } catch (error) {
    if (retries > 0 && error.status >= 500) {
      await sleep(1000);
      return fetchUsers(retries - 1);
    }
    throw error;
  }
}
```
```

### 4. Explain Trade-offs

Help Claude make informed decisions:

```markdown
## Approach 1: Eager Loading

```typescript
const users = await db.user.findMany({
  include: { posts: true }
});
```

**Pros**:
- Single query
- Fast for small datasets

**Cons**:
- Loads all data upfront
- Memory intensive for large datasets
- Wasteful if posts not always needed

## Approach 2: Lazy Loading

```typescript
const users = await db.user.findMany();
// Load posts only when needed
users[0].posts = await db.post.findMany({
  where: { userId: users[0].id }
});
```

**Pros**:
- Only loads what's needed
- Memory efficient

**Cons**:
- N+1 query problem
- Slower for large numbers of users

## Recommendation

Use eager loading when:
- Data set is small (< 100 records)
- Related data is always needed

Use lazy loading when:
- Data set is large
- Related data is rarely needed
- Can batch queries to avoid N+1
```

### 5. Keep Updated

Skills should reflect current best practices:

```yaml
version: 2.1.0
# Changelog:
# 2.1.0 - Updated for React 18 concurrent features
# 2.0.0 - Migrated from class components to hooks
# 1.0.0 - Initial version
```

## Debugging Skills

### Verify Discovery

Check if skill is found:
```bash
claude --list-skills
```

### Test Activation

Check if pattern matches:
```bash
echo "write tests for auth" | grep -E "write.*test"
```

### View Loaded Skills

During session:
```
Show me what skills are currently loaded.
```

Claude will list active skills.

### Common Issues

**Skill not loading**:
- Check YAML frontmatter is valid
- Verify file extension is `.md`
- Ensure file is in correct directory
- Check for name conflicts

**Pattern not triggering**:
- Test regex pattern separately
- Check for typos
- Ensure case-insensitive matching
- Try more specific pattern

**Skill conflicts**:
- Two skills with same name (later overrides)
- Overlapping activation patterns (both load)
- Dependency cycles (error)

## Summary

Skills are powerful knowledge modules that:

- **Extend Claude's expertise** in specific domains
- **Live as Markdown files** with YAML metadata
- **Activate automatically** via patterns or manually
- **Include examples and processes** for concrete guidance
- **Support files** for templates and scripts
- **Compose** to build complex capabilities

Use skills to:
- Codify team conventions
- Teach methodologies (TDD, code review)
- Provide domain expertise (database, security)
- Standardize processes (deployment, testing)
- Share knowledge across projects

With well-designed skills, Claude becomes a true expert in your team's specific practices and technologies.
