---
name: refactor
description: Systematic refactoring workflow with safety checks
args:
  target: Code to refactor - function, file, or module (optional)
  goal: Refactoring goal - "simplify", "performance", "readability" (optional)
---

# Code Refactoring Workflow

You are refactoring {{if .target}}{{.target}}{{else}}the code{{end}}{{if .goal}} to improve {{.goal}}{{end}}.

## Refactoring Principles

### The Golden Rule

**Change behavior OR structure, never both at once.**

Refactoring means improving code structure without changing behavior.

## Pre-Refactoring Checklist

Before touching any code:

1. **Ensure Tests Exist**
   - [ ] Code has comprehensive tests
   - [ ] All tests are passing
   - [ ] Coverage is adequate

   **If tests don't exist**: Write them first before refactoring.

2. **Understand Current Code**
   - [ ] Read and understand what it does
   - [ ] Identify dependencies
   - [ ] Note performance characteristics
   - [ ] Document assumptions

3. **Define Goal**
   - [ ] What are you improving? (readability, performance, maintainability)
   - [ ] What is the success criteria?
   - [ ] What are you NOT changing? (public APIs, behavior)

4. **Commit Current State**
   - [ ] Commit or stash any uncommitted changes
   - [ ] Have a clean baseline to revert to

## Refactoring Process

### Phase 1: Identify Smells

Common code smells:

**Complexity**
- Long functions (>20 lines)
- Deep nesting (>3 levels)
- Complex conditionals
- Many parameters (>3-4)

**Duplication**
- Copy-pasted code
- Similar logic in multiple places
- Repeated patterns

**Naming**
- Unclear variable names
- Misleading function names
- Inconsistent terminology

**Structure**
- God objects (classes doing too much)
- Feature envy (using other object's data)
- Tight coupling
- Poor abstraction boundaries

### Phase 2: Plan Changes

Choose refactoring techniques:

**Extract Method/Function**
```go
// Before: Long function
func ProcessOrder(order Order) {
    // 50 lines of code
}

// After: Extracted smaller functions
func ProcessOrder(order Order) {
    ValidateOrder(order)
    CalculateTotal(order)
    ApplyDiscounts(order)
    ProcessPayment(order)
}
```

**Extract Variable**
```go
// Before: Complex expression
if (user.age >= 18 && user.hasLicense && !user.isSuspended) {
    // ...
}

// After: Named variable
canDrive := user.age >= 18 && user.hasLicense && !user.isSuspended
if canDrive {
    // ...
}
```

**Rename**
```go
// Before: Unclear
func calc(a, b int) int { return a * b }

// After: Clear
func CalculateArea(width, height int) int { return width * height }
```

**Simplify Conditionals**
```go
// Before: Complex nested ifs
if user != nil {
    if user.IsActive {
        if user.HasPermission("admin") {
            // do thing
        }
    }
}

// After: Guard clauses
if user == nil || !user.IsActive || !user.HasPermission("admin") {
    return
}
// do thing
```

### Phase 3: Refactor in Small Steps

**For each change:**

1. **Make One Change**
   - Focus on one improvement at a time
   - Keep the change small and atomic

2. **Run Tests**
   - Execute full test suite
   - All tests must pass
   - If tests fail, revert and try differently

3. **Commit**
   - Commit after each successful change
   - Use descriptive commit messages
   - Small commits are easy to review and revert

4. **Repeat**
   - Move to next improvement
   - Build incrementally

### Phase 4: Verify Improvement

After refactoring:

1. **Tests Still Pass**
   ```bash
   go test ./...
   ```

2. **Behavior Unchanged**
   - Run integration tests
   - Test manually if needed
   - Verify edge cases

3. **Goal Achieved**
   - Is code more readable?
   - Is performance better?
   - Is it more maintainable?

4. **No Regressions**
   - Check for performance degradation
   - Verify no new bugs introduced
   - Ensure compatibility maintained

## Common Refactoring Patterns

### 1. Extract Function
When to use: Function is too long or does multiple things

Steps:
1. Identify cohesive block of code
2. Extract to new function with descriptive name
3. Replace original code with function call
4. Run tests

### 2. Inline Function
When to use: Function is trivial and adds no value

Steps:
1. Replace function calls with function body
2. Remove function definition
3. Run tests

### 3. Replace Magic Numbers
When to use: Unexplained constants in code

```go
// Before
if user.age > 18 { ... }

// After
const LegalAdultAge = 18
if user.age > LegalAdultAge { ... }
```

### 4. Consolidate Conditional
When to use: Multiple conditions checking same thing

```go
// Before
if user.IsAdmin() { return true }
if user.IsOwner() { return true }
if user.IsModerator() { return true }
return false

// After
return user.IsAdmin() || user.IsOwner() || user.IsModerator()
```

### 5. Replace Nested Conditionals
When to use: Deep nesting reduces readability

```go
// Before
if x != nil {
    if x.valid {
        if x.ready {
            process(x)
        }
    }
}

// After
if x == nil { return }
if !x.valid { return }
if !x.ready { return }
process(x)
```

## Refactoring Checklist

Before you finish:

- [ ] All tests pass
- [ ] No new warnings or errors
- [ ] Code is more readable than before
- [ ] Complexity reduced (if that was the goal)
- [ ] Performance maintained or improved
- [ ] Public APIs unchanged (unless intentional)
- [ ] Documentation updated
- [ ] Changes committed with clear messages

## Output Format

Document your refactoring:

```markdown
## Refactoring Report: [Target]

### Goal
[What you're improving and why]

### Current State
[Code smells and issues identified]

### Refactoring Plan
1. [Step 1]: [Technique to use]
2. [Step 2]: [Technique to use]
...

### Changes Made

#### Change 1: [Description]
**Before**:
```go
[old code]
```

**After**:
```go
[new code]
```

**Improvement**: [Why this is better]
**Tests**: ✓ All passing

#### Change 2: ...

### Verification
- [ ] All tests pass
- [ ] Behavior unchanged
- [ ] Goal achieved
- [ ] No regressions

### Metrics
- Lines of code: [before] → [after]
- Cyclomatic complexity: [before] → [after]
- Test coverage: [before] → [after]
```

## Critical Rules

1. **Tests First**: Never refactor without tests
2. **Small Steps**: One change at a time
3. **Green Bar**: Keep tests passing always
4. **Commit Often**: Commit after each successful change
5. **Revert Freely**: If stuck, revert and try differently
6. **Behavior Unchanged**: Refactoring should not change what code does

Begin the refactoring process now.
