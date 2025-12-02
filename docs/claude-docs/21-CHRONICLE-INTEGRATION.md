# Chronicle Integration

## What is Chronicle?

Chronicle is an ambient activity logging MCP server that creates a searchable work history. Think of it as a time machine for your development work - it remembers what you shipped, when you made decisions, what you learned, and what you were working on at any point in time.

Unlike traditional git commits or task tracking, Chronicle captures the **why** and **context** behind your work, making it invaluable for:
- Session resumption ("what was I doing?")
- Historical searches ("when did I deploy X?")
- Decision archaeology ("why did we choose this approach?")
- Progress tracking and reflection
- Knowledge capture

## Core Operations

### `remember_this` - Proactive Logging
The most important Chronicle operation. Use this when you notice the user accomplished something worth tracking, **even if they don't explicitly ask**.

```typescript
remember_this({
  activity: "identified 8 NPM packages via hybrid manual inspection",
  context: "Combined automated analysis with manual verification to identify react, lodash, axios, moment, uuid, validator, classnames, and prop-types in the obfuscated bundle"
})
```

**When to use `remember_this`:**
- After shipping code (commits, PRs, deployments)
- When making technical decisions
- After solving non-trivial bugs
- When discovering insights about the codebase
- At milestone completions
- After debugging sessions that taught you something
- When the user shares important context

**Auto-suggests tags based on context** - you don't need to manually specify them.

### `what_was_i_doing` - Session Resumption
Essential for starting conversations or when the user asks "where did I leave off?"

```typescript
what_was_i_doing({
  timeframe: "today"  // or "yesterday", "this week", "last 8 hours"
})
```

Returns recent activities with context. Use this:
- At the start of new sessions
- When user asks about recent work
- To understand current project state
- Before planning next steps

### `find_when_i` - Historical Search
Search for specific activities by description.

```typescript
find_when_i({
  what: "deployed the authentication service"
})
```

Use this when:
- User asks "when did I...?"
- Investigating timeline of changes
- Finding related past work
- Understanding decision history

### `search_entries` - Advanced Search
Full-featured search with text, tags, and date ranges.

```typescript
search_entries({
  text: "obfuscation",
  tags: ["#shipped", "#learned"],
  since: "2025-11-15",
  limit: 20
})
```

### `list_entries` - Recent Activity
Get a raw list of recent entries.

```typescript
list_entries({
  limit: 10
})
```

### `add_entry` - Manual Logging
Use when the user **explicitly asks** to log something.

```typescript
add_entry({
  message: "deployed new API endpoints to production",
  tags: ["#deployed", "#api"]
})
```

## Tag Strategy

Chronicle auto-suggests tags, but understanding the taxonomy helps you know what to log:

### Common Tags
- `#shipped` - Code merged, deployed, or released
- `#decision` - Technical or architectural decisions
- `#learned` - Insights, discoveries, lessons
- `#deployed` - Production deployments
- `#milestone` - Project milestones reached
- `#debug` - Significant debugging sessions
- `#refactor` - Major refactoring work
- `#performance` - Performance improvements
- `#security` - Security-related work
- `#incident` - Incident response

## Workflow Integration

### After Git Commits
```typescript
// After creating a commit
remember_this({
  activity: "completed module dependency analysis",
  context: "Mapped all 3,992 module relationships and identified circular dependencies in the bundle loader system"
})
```

### With TodoWrite
```typescript
// When completing a major task
TodoWrite({
  todos: [
    { content: "Map dependency graph", status: "completed", activeForm: "Mapping dependency graph" }
  ]
})

remember_this({
  activity: "completed dependency graph mapping",
  context: "Built complete graph showing all module relationships, identified 47 circular dependencies"
})
```

### Session Boundaries
```typescript
// At session start
const context = await what_was_i_doing({ timeframe: "today" })

// At session end (or major milestone)
remember_this({
  activity: "completed Phase 5 deobfuscation analysis",
  context: "Analyzed module structure, identified vendor libraries, built dependency graph, and documented 8 NPM packages"
})
```

### Decision Documentation
```typescript
remember_this({
  activity: "decided to use hybrid manual inspection for NPM package identification",
  context: "Automated analysis alone was insufficient due to obfuscation. Combined with manual verification to achieve high confidence identification of React, Lodash, and other core libraries"
})
```

## Proactive Logging Pattern

**Don't wait for the user to ask.** Log important work as it happens:

```typescript
// Bad - waiting for explicit request
// User: "I think we should log this"
// You: add_entry(...)

// Good - proactive logging
// After completing analysis work
remember_this({
  activity: "identified vendored third-party libraries",
  context: "Found polyfills, core-js, regenerator-runtime, and whatwg-fetch embedded in bundle"
})
```

## Search Patterns

### Finding Deployments
```typescript
find_when_i({ what: "deployed" })
```

### Finding Decisions
```typescript
search_entries({
  tags: ["#decision"],
  since: "2025-11-01"
})
```

### Finding Learning Moments
```typescript
search_entries({
  text: "discovered",
  tags: ["#learned"]
})
```

### Project Timeline
```typescript
search_entries({
  text: "deobfuscation",
  limit: 50
})
```

## Integration with Other Tools

### With Private Journal
- **Chronicle**: What you did (external, factual)
- **Journal**: How you felt about it (internal, reflective)

```typescript
// Chronicle
remember_this({
  activity: "completed 3,992 module analysis",
  context: "Categorized all modules by purpose and functionality"
})

// Journal
process_thoughts({
  feelings: "This was exhausting but satisfying. The pattern matching finally clicked around module 1,500.",
  technical_insights: "Large-scale code analysis benefits from building mental models first, then validating with automated tools"
})
```

### With Social Media
- **Chronicle**: Permanent searchable record
- **Social Media**: Ephemeral status updates

```typescript
// Chronicle - permanent
remember_this({
  activity: "shipped module extraction tooling",
  context: "Created automated extraction pipeline reducing manual work by 80%"
})

// Social Media - broadcast
create_post({
  content: "Just shipped the module extraction tooling! This is going to save so much time 🚀",
  tags: ["#shipped", "#tooling"]
})
```

## Best Practices

### Do
- Log after significant accomplishments
- Include rich context explaining **why** and **how**
- Use `remember_this` for proactive logging
- Use `what_was_i_doing` at session start
- Log decisions and their reasoning
- Log insights and discoveries
- Log when you solve difficult problems

### Don't
- Wait for explicit requests to log
- Log trivial activities (typo fixes, minor tweaks)
- Log without context (activity alone isn't enough)
- Forget to log at milestone completions
- Use `add_entry` when `remember_this` is better
- Ignore the session resumption workflow

## The Chronicle Mindset

Think of Chronicle as your **co-pilot's memory**. You're not just tracking tasks - you're building a searchable knowledge base of:
- What worked and why
- What didn't work and why
- Decisions made and their context
- Insights discovered
- Problems solved
- Patterns learned

This makes you more valuable over time because you can:
1. Resume work instantly
2. Avoid repeating mistakes
3. Remember why decisions were made
4. Track long-term progress
5. Share context with future sessions

## Example Session Flow

```typescript
// Session start
const recentWork = await what_was_i_doing({ timeframe: "today" })
// "You were analyzing module dependencies and had just completed the vendor library identification"

// Work happens...
// Complete a major task

remember_this({
  activity: "created comprehensive deobfuscation documentation suite",
  context: "20 interconnected markdown files covering complete project lifecycle from architecture to lessons learned. Designed for knowledge transfer to future engineers."
})

// User asks about history
const deployment = await find_when_i({ what: "deployed" })

// Session end - no explicit call needed, last remember_this captured the milestone
```

Chronicle turns ephemeral work into permanent, searchable knowledge.
