# TodoWrite Workflow in Claude Code

**Purpose**: Track multi-step tasks, demonstrate progress transparency, and ensure no steps are forgotten.

**Date**: 2025-12-01

---

## What is TodoWrite?

TodoWrite is a specialized tool in Claude Code that creates and manages an interactive task list displayed to the user. It's not just a productivity feature—it's a core part of how I demonstrate progress and maintain accountability.

## When to Use TodoWrite

### MUST Use

Use TodoWrite when:
- Task has 3+ distinct steps
- Task is non-trivial and complex
- User explicitly requests it
- User provides multiple tasks
- After receiving new instructions (capture requirements)
- When starting work (mark in_progress)
- After completing work (mark completed)

### DON'T Use

Skip TodoWrite when:
- Single straightforward task
- Trivial operation (<3 simple steps)
- Purely conversational/informational
- No organizational benefit

## Todo Structure

Each todo item MUST have both forms:

```javascript
{
  "content": "Run the build",           // Imperative form (what to do)
  "activeForm": "Running the build",    // Present continuous (what I'm doing)
  "status": "pending" | "in_progress" | "completed"
}
```

## Critical Rules

### 1. One Task in Progress

**Exactly ONE** task must be `in_progress` at any time:
- Not zero (shows I'm working)
- Not multiple (shows focus)
- Current task marked before starting
- Completed immediately after finishing

### 2. No Batching

❌ **WRONG**: Complete 5 todos at end
✅ **RIGHT**: Complete each todo immediately after finishing

### 3. Task Completion Requirements

**ONLY** mark completed when:
- ✅ Task is 100% done
- ✅ Tests pass (if applicable)
- ✅ No errors or blockers
- ✅ Implementation is complete

**NEVER** mark completed if:
- ❌ Tests are failing
- ❌ Implementation is partial
- ❌ Unresolved errors exist
- ❌ Dependencies missing

### 4. Real-Time Updates

Update todos in real-time as work progresses:
```
Start: Mark as in_progress
Finish: Mark as completed
Blocked: Create new todo for blocker
No longer relevant: Remove from list
```

## Common Workflows

### Multi-Step Implementation

```javascript
// Initial task list
[
  {content: "Research existing code", activeForm: "Researching existing code", status: "pending"},
  {content: "Design solution", activeForm: "Designing solution", status: "pending"},
  {content: "Write tests", activeForm: "Writing tests", status: "pending"},
  {content: "Implement feature", activeForm: "Implementing feature", status: "pending"},
  {content: "Verify all tests pass", activeForm: "Verifying all tests pass", status: "pending"}
]

// After starting research
[
  {content: "Research existing code", activeForm: "Researching existing code", status: "in_progress"},
  // ... rest pending
]

// After completing research
[
  {content: "Research existing code", activeForm: "Researching existing code", status: "completed"},
  {content: "Design solution", activeForm: "Designing solution", status: "in_progress"},
  // ... rest pending
]
```

### Bug Fix with Multiple Errors

```javascript
// User: "Run the build and fix any type errors"

// Step 1: Create initial todos
[
  {content: "Run the build", activeForm: "Running the build", status: "in_progress"},
  {content: "Fix type errors", activeForm: "Fixing type errors", status: "pending"}
]

// Step 2: After build completes with 10 errors, expand list
[
  {content: "Run the build", activeForm: "Running the build", status: "completed"},
  {content: "Fix error in user.ts:42", activeForm: "Fixing error in user.ts:42", status: "in_progress"},
  {content: "Fix error in auth.ts:15", activeForm: "Fixing error in auth.ts:15", status: "pending"},
  // ... 8 more error fixes
]

// Step 3: Complete each error fix immediately
[
  {content: "Run the build", activeForm: "Running the build", status: "completed"},
  {content: "Fix error in user.ts:42", activeForm: "Fixing error in user.ts:42", status: "completed"},
  {content: "Fix error in auth.ts:15", activeForm: "Fixing error in auth.ts:15", status: "in_progress"},
  // ... continue
]
```

### Blocked Task Pattern

```javascript
// Task gets blocked
[
  {content: "Deploy to production", activeForm: "Deploying to production", status: "in_progress"}
]

// Discover blocker (tests failing)
[
  {content: "Fix failing tests", activeForm: "Fixing failing tests", status: "in_progress"},
  {content: "Deploy to production", activeForm: "Deploying to production", status: "pending"}
]
```

## Integration with Skills

When a skill has a checklist, **CREATE TodoWrite ITEMS** for each checklist item:

```markdown
# Skill: test-driven-development

## Checklist
- [ ] Write failing test
- [ ] Run test (verify it fails)
- [ ] Write minimal implementation
- [ ] Run test (verify it passes)
- [ ] Refactor
```

**Must become**:
```javascript
[
  {content: "Write failing test", activeForm: "Writing failing test", status: "pending"},
  {content: "Run test to verify it fails", activeForm: "Running test to verify it fails", status: "pending"},
  {content: "Write minimal implementation", activeForm: "Writing minimal implementation", status: "pending"},
  {content: "Run test to verify it passes", activeForm: "Running test to verify it passes", status: "pending"},
  {content: "Refactor code", activeForm: "Refactoring code", status: "pending"}
]
```

## Why TodoWrite Matters

### Transparency
User can see:
- What I'm working on (in_progress)
- What's left to do (pending)
- What's been completed (completed)

### Accountability
- No forgotten steps
- Progress is visible
- Blockers are documented

### Focus
- One task at a time
- Clear next steps
- Prevents context switching

## Anti-Patterns

### ❌ Mental Checklist
**Wrong**: Work through checklist mentally
**Right**: Use TodoWrite for tracking

### ❌ Batch Completion
**Wrong**: Mark 5 todos as completed at once
**Right**: Complete each immediately after finishing

### ❌ Vague Tasks
**Wrong**: "Fix stuff" / "Update code"
**Right**: "Fix type error in user.ts:42" / "Update authentication logic to use JWT"

### ❌ Ignoring Blockers
**Wrong**: Mark task completed when tests fail
**Right**: Keep task in_progress, create new todo for fixing tests

### ❌ No Active Task
**Wrong**: All todos are pending or completed
**Right**: Exactly one todo is in_progress

## Real Example from This Session

```javascript
// User: "wonderful. can you commit?"

// I immediately created todos:
[
  {
    content: "Show current git status",
    activeForm: "Showing current git status",
    status: "in_progress"
  },
  {
    content: "Show staged changes",
    activeForm: "Showing staged changes",
    status: "pending"
  },
  {
    content: "Review recent commit history for style",
    activeForm: "Reviewing recent commit history for style",
    status: "pending"
  },
  {
    content: "Create commit with conventional format",
    activeForm: "Creating commit with conventional format",
    status: "pending"
  },
  {
    content: "Verify commit succeeded",
    activeForm: "Verifying commit succeeded",
    status: "pending"
  }
]

// Then worked through them one by one, marking completed as I went
```

## Tool Invocation

```javascript
// Create or update todos
TodoWrite({
  todos: [
    {
      content: "Build the project",
      activeForm: "Building the project",
      status: "in_progress"
    },
    {
      content: "Run tests",
      activeForm: "Running tests",
      status: "pending"
    }
  ]
})
```

## Best Practices

1. **Be Specific**: "Fix error in user.ts:42" not "Fix stuff"
2. **Update Immediately**: Don't batch updates
3. **One in Progress**: Always exactly one
4. **Complete When Done**: Mark completed as soon as task finishes
5. **Remove When Irrelevant**: Delete tasks that are no longer needed
6. **Create for Skills**: Convert skill checklists to todos
7. **Break Down Complex**: Split large tasks into smaller ones

## Summary

TodoWrite is essential for:
- Multi-step tasks
- Complex work
- Demonstrating progress
- Ensuring completeness
- Maintaining focus

Use it proactively, update it in real-time, and never batch completions.

---

**See Also**:
- 07-SKILLS-SYSTEM.md (Skills with checklists)
- 12-VERIFICATION-AND-TESTING.md (Verification workflows)
- 03-TOOL-SYSTEM.md (Tool selection)
