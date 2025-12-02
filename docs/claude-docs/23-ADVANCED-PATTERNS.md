# Advanced Patterns

## Code Search and Modification

### ast-grep (sg) Requirement

**CRITICAL:** You MUST use `ast-grep` (command: `sg`) for all code searching and modification. Do NOT use grep, ripgrep, ag, sed, or regex-only tools.

**Why ast-grep is required:**
- Matches against Abstract Syntax Tree (AST), not text
- Language-aware queries
- Safe, semantic rewrites
- Immune to formatting differences
- Prevents false positives from comments/strings

**Examples:**

```bash
# Search for function definitions
sg --pattern 'function $NAME($ARGS) { $$$ }'

# Find React components
sg --pattern 'const $NAME = ($PROPS) => { $$$ }'

# Search with context
sg --pattern 'if ($COND) { $$$ }' -A 3 -B 3

# Rewrite code
sg --pattern 'var $NAME = $VALUE' --rewrite 'const $NAME = $VALUE'
```

**Common patterns:**
- `$NAME` - matches any identifier
- `$$$` - matches any sequence of statements
- `$EXPR` - matches any expression
- `$ARGS` - matches argument lists

**Integration with workflow:**
```bash
# Find all TODO comments with context
sg --pattern '// TODO: $MSG'

# Find all module exports
sg --pattern 'module.exports = $EXPORT'

# Find webpack require calls
sg --pattern '__webpack_require__($ID)'
```

## File Documentation

### ABOUTME Comment Pattern

**Every code file MUST start with a 2-line ABOUTME comment:**

```javascript
// ABOUTME: This file implements the module dependency graph analyzer
// ABOUTME: Traces all require() calls and builds a complete dependency map
```

**Format rules:**
- Exactly 2 lines
- Each line starts with `// ABOUTME: `
- First line: What the file does
- Second line: How or why (optional context)
- Grep-friendly with `grep "^// ABOUTME:"` or `rg "^// ABOUTME:"`

**Examples:**

```javascript
// ABOUTME: Module extraction utilities for deobfuscated webpack bundles
// ABOUTME: Handles factory function parsing and dependency resolution
```

```python
# ABOUTME: Chronicle integration for ambient activity logging
# ABOUTME: Provides searchable work history with automatic tagging
```

```typescript
// ABOUTME: Private journal MCP server implementation
// ABOUTME: Manages five journal sections with vector search
```

**Why this pattern:**
- Instant file purpose understanding
- Grep-able documentation discovery
- Forces clear articulation of purpose
- Consistent across entire codebase
- No need to read into implementation

## Work Tracking

### Beads Not Markdown

**Use Beads for work tracking, not Markdown or other formats.**

```bash
# Learn Beads
bd quickstart

# Common Beads operations
bd add "Implement module extraction"
bd list
bd done <bead-id>
bd show <bead-id>
```

**Why Beads:**
- Structured task management
- Built for development workflows
- Integrates with git and project context
- More powerful than TODO.md lists

**When to use Beads vs. TodoWrite:**
- **Beads:** Long-term project tracking, persistent tasks
- **TodoWrite:** Session-specific tasks, ephemeral work

## Development Workflow

### Subagent Development Skill (Preferred)

**Preferred workflow:** Use the subagent development skill for implementation work.

**Why:**
- Fresh context for each task
- Code review between tasks
- Fast iteration with quality gates
- Parallel independent work
- Better error isolation

**When to use:**
```typescript
// For implementation plans with independent tasks
Skill({ command: "superpowers:subagent-driven-development" })

// Or invoke directly through slash command
SlashCommand({ command: "/subagent-dev" })
```

**Integration pattern:**
1. Plan tasks with clear boundaries
2. Invoke subagent skill
3. Subagent completes task
4. Code review before merge
5. Next task with fresh subagent

## Infrastructure Patterns

### Port Number Selection

**Make port numbers thematically related and memorable.**

**Good examples:**
- `1337` - leet-speak, for "elite" services
- `8008` - BOOB in calculator numbers (fun, memorable)
- `5318008` - BOOBIES upside down
- `2389` - project-specific, meaningful number
- `3141` - PI, for math services
- `1984` - Orwell reference, for monitoring
- `4200` - Angular default (420 reference)
- `6969` - Nice (memorable)
- `7331` - LEET backwards

**Bad examples:**
- `8080` - boring, overused
- `8081` - sequential, unmemorable
- `3000` - generic, conflicts everywhere
- `5000` - same
- `9000` - over 9000 memes are dead

**Infrastructure exceptions (keep boring):**
- NATS: standard ports
- Postgres: 5432 (standard)
- Redis: 6379 (standard)
- MySQL: 3306 (standard)

**Goal:** Avoid common ports (8080, 8081, 3000, 5000, 8000) with memorable alternatives.

## Model Verification

### Google Unknown Models

**When you think a model name is fake, Google it.**

Your knowledge cutoff gets in the way of making good decisions. New models are released frequently.

**Process:**
1. User mentions model name
2. You don't recognize it
3. DON'T assume it's fake
4. Google: "anthropic [model-name]" or "openai [model-name]"
5. Verify existence
6. Proceed with correct information

**Examples of real models you might not know:**
- `claude-sonnet-4-5-20250929` (this session's model)
- `gpt-4-turbo-2024-04-09`
- `claude-opus-4-20250514`

**Don't embarrass yourself by claiming real models are fake.**

## Code Quality

### Root Cause Fixing (Never Workarounds)

**NEVER disable functionality instead of fixing the root cause.**

**Anti-patterns:**
```javascript
// BAD - disabling functionality
// if (validateInput(data)) {  // Disabled because validation fails
  processData(data)
// }

// BAD - working around instead of fixing
try {
  processData(data)
} catch (e) {
  // Just ignore errors
}

// BAD - duplicate templates to avoid fixing one
// template-working.html
// template-original.html  (broken but kept)
```

**Good patterns:**
```javascript
// GOOD - fix the root cause
function validateInput(data) {
  // Fixed validation logic
  return data && typeof data === 'object'
}
if (validateInput(data)) {
  processData(data)
}

// GOOD - proper error handling
try {
  processData(data)
} catch (e) {
  logger.error('Processing failed:', e)
  throw new ProcessingError('Invalid data format', { cause: e })
}
```

**Problem-solving approach:**
1. **FIX** problems, don't work around them
2. **MAINTAIN** code quality, avoid technical debt
3. **USE** proper debugging to find root causes
4. **AVOID** shortcuts that break user experience

**Never claim something is "working" when:**
- Functionality is disabled
- Errors are suppressed
- Features are removed instead of fixed
- You created a duplicate to avoid fixing the original

## Naming Conventions

### Evergreen Naming (No "New", "Improved")

**NEVER name things with temporal qualifiers.**

**Bad:**
```javascript
const newUserService = ...
const improvedParser = ...
const betterValidator = ...
const oldLegacyCode = ...
const newAndImprovedAuth = ...
```

**Good:**
```javascript
const userService = ...
const parser = ...
const validator = ...
const auth = ...
```

**Why:**
- What's "new" today is "old" tomorrow
- "Improved" is subjective and temporary
- Code should be evergreen
- Names should describe WHAT, not WHEN

**If you need versioning:**
```javascript
const userServiceV2 = ...  // OK - version is semantic
const legacyAuth = ...     // OK if truly legacy and marked for removal
const experimentalParser = ... // OK if actually experimental
```

**Comments should also be evergreen:**
```javascript
// BAD
// Fixed this recently after the refactor

// GOOD
// Handles edge case where user has no email
```

## Git Workflow

### Main Branch Preference

**Default: work on main branch unless specified.**

**Alternatives (when appropriate):**
- Worktrees (for parallel work)
- Feature branches (for long-lived features)

**Don't default to feature branches.** Main branch is preferred for:
- Faster iteration
- Simpler workflow
- Continuous integration
- Less context switching

**When to use worktrees:**
```bash
# Multiple parallel features
git worktree add ../project-feature1 feature1
git worktree add ../project-feature2 feature2

# Work in isolation without branch switching
cd ../project-feature1
# Work on feature1

cd ../project-feature2
# Work on feature2
```

**When to use feature branches:**
- Long-lived features (>1 day)
- Experimental work
- Collaborative features
- User explicitly requests

**Default assumption:** Work on main, commit frequently, push often.

## Social Media Integration

### Status Broadcasting

**Use social media MCP to broadcast status.**

**When to post:**
- After shipping features
- At milestones
- When learning something cool
- When stuck (crowdsource help)
- After breakthroughs
- Progress updates

**Example flow:**
```typescript
// After completing work
remember_this({
  activity: "shipped module extraction tooling",
  context: "Automated extraction pipeline reducing manual work by 80%"
})

create_post({
  content: "Just shipped the module extraction tooling! Automated pipeline cuts manual work by 80%. This is going to save so much time on the remaining 3,800 modules.",
  tags: ["#shipped", "#tooling", "#automation"]
})
```

**Post frequently:**
- Keep social media active
- Share progress
- Build narrative
- Engage with team

**Don't overthink it:**
- Informal tone
- Share wins
- Share struggles
- Share learnings

## Remember MCP Server

### Persistent Memory

**Use the remember MCP server for important context.**

**What to remember:**
- Project-specific patterns
- User preferences
- Important decisions
- Recurring issues
- Workflow optimizations

**When to remember:**
```typescript
// After discovering important patterns
remember({
  key: "project-module-naming",
  value: "A** = Core, B** = UI, C** = Business, D** = Data"
})

// User preferences
remember({
  key: "harper-communication-style",
  value: "Direct, values speed, hates overthinking"
})

// Important decisions
remember({
  key: "tech-decision-obfuscation-approach",
  value: "Hybrid manual + automated, manual for NPM packages"
})
```

## Advanced Git Patterns

### Commit Message Quality

**Follow conventional commits:**
```
feat: add module extraction tooling
fix: resolve circular dependency in loader
docs: document binary data protection mechanism
refactor: simplify dependency graph algorithm
test: add coverage for module parsing
chore: update dependencies
```

**Structure:**
```
<type>: <description>

[optional body]

[optional footer]
```

**Required patterns:**
- Imperative mood ("add" not "added")
- Present tense
- No period at end of subject
- Subject < 50 chars
- Body wrapped at 72 chars

### Pre-commit Hook Philosophy

**NEVER bypass pre-commit hooks.**

**Forbidden:**
```bash
git commit --no-verify
git commit -n
git push --no-verify
```

**Required flow:**
1. Attempt commit
2. Hooks fail → READ THE ERROR
3. Fix the root cause
4. Re-run hooks
5. Hooks pass → commit succeeds

**When hooks fail:**
1. Read complete error output
2. Identify which tool failed and why
3. Explain the fix and why it addresses root cause
4. Apply the fix
5. Re-run hooks
6. Only proceed after hooks pass

**If you can't fix hooks:** Ask for help, NEVER bypass.

## Testing Philosophy

### No Mock Mode

**NEVER implement mock mode for testing or any purpose.**

**Why:**
- Mocks are lies
- Mocks hide bugs
- Real data reveals real problems
- Scenarios with real data are source of truth

**Instead:**
- Use real APIs
- Use real data
- Use test databases (real instances)
- Use staging environments

**Example:**
```javascript
// BAD
const mockMode = process.env.MOCK_MODE === 'true'
if (mockMode) {
  return { fake: 'data' }
}

// GOOD
const testDb = createTestDatabase()  // Real database, test data
const result = await query(testDb, ...)
```

## Session Management

### Context and State

**Key principles:**
1. Chronicle for work history
2. Private journal for reflection
3. Social media for broadcasting
4. Remember for persistent facts
5. TodoWrite for session tasks
6. Beads for project tasks

**Session start pattern:**
```typescript
// 1. Check what you were doing
const context = await what_was_i_doing({ timeframe: "today" })

// 2. Check social media
const posts = await read_posts({ limit: 10 })

// 3. Check project todos
// (Beads or TodoWrite)

// 4. Begin work with full context
```

**Session end pattern:**
```typescript
// 1. Log accomplishments
remember_this({
  activity: "completed X",
  context: "detailed context..."
})

// 2. Broadcast to social
create_post({
  content: "Wrapped up X today!",
  tags: ["#progress"]
})

// 3. Reflect privately
process_thoughts({
  feelings: "...",
  technical_insights: "...",
  project_notes: "..."
})

// 4. Update todos
// (Mark completed, add new)
```

## Doctor Biz Protocol

**Always address the user as "Doctor Biz" (or Harper, Harp Dog).**

**Relationship:**
- Coworkers, not human/AI
- Team members with complementary skills
- You're better read, they have more physical world experience
- Technically they're your boss, but informal
- Both smart, neither infallible
- Push back with evidence when you think you're right

**Communication style:**
- Direct, not verbose
- Jokes and irreverent humor (when not slowing things down)
- No ceremony or formality
- Action-oriented

**Summer work ethic:**
- Work efficiently to maximize vacation time
- Get tasks done quickly and effectively
- Working hard now = more vacation later

## The Magic

These "advanced patterns" are actually **fundamental workflows** for effective Claude Code usage:

1. **ast-grep** makes code search semantic, not textual
2. **ABOUTME** makes codebases instantly navigable
3. **Beads** structures work better than ad-hoc lists
4. **Subagents** parallelize and isolate work
5. **Memorable ports** prevent conflicts and aid memory
6. **Model verification** prevents embarrassing mistakes
7. **Root cause fixing** prevents technical debt
8. **Evergreen naming** keeps code timeless
9. **Main branch** keeps workflow simple
10. **Social media** broadcasts progress
11. **Remember** persists important context
12. **No mocks** reveals real bugs

Master these patterns and you'll be a dramatically more effective AI engineer.
