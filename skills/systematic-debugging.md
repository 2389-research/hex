---
name: systematic-debugging
description: Four-phase debugging framework for investigating bugs methodically without guessing
tags:
  - debugging
  - troubleshooting
  - methodology
  - root-cause
activationPatterns:
  - "debug"
  - "fix.*bug"
  - "error"
  - "failing"
  - "not working"
  - "troubleshoot"
priority: 9
version: 1.0.0
---

# Systematic Debugging

## Overview

Systematic debugging is a disciplined approach to finding and fixing bugs. The goal is to **understand the problem before attempting solutions**. Guessing and randomly trying fixes wastes time and often introduces new bugs.

## The Four Phases

### Phase 1: Root Cause Investigation

**Goal**: Understand what is actually happening vs. what should happen

**Steps**:

1. **Reproduce the bug** consistently
   - Document exact steps to trigger it
   - Note any conditions required (environment, data, timing)
   - Verify you can make it happen repeatedly

2. **Gather evidence**
   - Read error messages completely (don't skim)
   - Check logs for relevant entries
   - Examine stack traces
   - Note what changed recently (code, config, dependencies)

3. **Form a hypothesis**
   - Based on evidence, what do you think is wrong?
   - What would cause this specific symptom?
   - Where in the code could this originate?

**Example**:

```
Bug: "User login fails with 500 error"

Evidence:
- Error log: "database connection timeout after 5s"
- Stack trace: auth_handler.go:42 → db_client.go:156
- Recent change: Upgraded database client library yesterday

Hypothesis: New DB client has different timeout defaults
```

### Phase 2: Pattern Analysis

**Goal**: Identify when bug happens and when it doesn't

**Questions to Ask**:

- Does it happen every time or intermittently?
- Does it happen for all users or specific ones?
- Does it happen in all environments or just production?
- Does it happen with all data or specific inputs?
- Is there a pattern in timing (time of day, load level)?

**Create a Truth Table**:

```
| Condition         | Bug Occurs? |
|-------------------|-------------|
| User A, Dev env   | No          |
| User A, Prod env  | Yes         |
| User B, Dev env   | No          |
| User B, Prod env  | Yes         |
| Admin, Prod env   | No          |
```

**Analysis**: Bug only affects regular users in production. Admins are unaffected.

### Phase 3: Hypothesis Testing

**Goal**: Prove or disprove your hypothesis with evidence

**Testing Methods**:

1. **Add logging/instrumentation**
   ```go
   log.Printf("DEBUG: About to query database, timeout=%v", timeout)
   result, err := db.Query(ctx, query)
   log.Printf("DEBUG: Query returned err=%v, elapsed=%v", err, time.Since(start))
   ```

2. **Use a debugger**
   - Set breakpoints at suspected locations
   - Inspect variable values
   - Step through execution

3. **Isolate the component**
   - Can you trigger the bug in a unit test?
   - Can you reproduce with minimal code?

4. **Compare working vs broken**
   - Diff configurations
   - Compare code versions (git diff)
   - Check environment variables

**Example Test**:

```go
// Hypothesis: Timeout is too short
func TestDatabaseTimeout(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    _, err := db.Query(ctx, "SELECT * FROM large_table")

    // If this fails with timeout, hypothesis confirmed
    if err == context.DeadlineExceeded {
        t.Logf("Confirmed: 5s timeout is insufficient")
    }
}
```

### Phase 4: Implementation

**Goal**: Fix the root cause, not the symptom

**Steps**:

1. **Identify the fix**
   - Based on proven hypothesis
   - Fix the cause, not the effect
   - Consider side effects

2. **Implement the fix**
   - Make minimal changes
   - Follow existing patterns
   - Add comments explaining WHY

3. **Verify the fix**
   - Reproduce original bug → confirm it's fixed
   - Run all tests → confirm nothing broke
   - Check edge cases → ensure complete fix

4. **Prevent recurrence**
   - Add test that catches this bug
   - Document the issue
   - Consider if similar bugs exist elsewhere

**Example Fix**:

```go
// Before: Timeout too short
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

// After: Increased timeout based on analysis
// NOTE: Large tables need more time; 30s timeout chosen based on prod metrics
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
```

## Anti-Patterns to Avoid

### ❌ Random Guessing

```go
// WRONG: Just trying things randomly
err := DoSomething()
if err != nil {
    // "Maybe it's a race condition? Let me add a sleep..."
    time.Sleep(100 * time.Millisecond)
    err = DoSomething() // Try again?
}
```

This doesn't understand the problem and often hides symptoms without fixing cause.

### ✅ Systematic Investigation

```go
// RIGHT: Understand the error first
err := DoSomething()
if err != nil {
    log.Printf("ERROR: DoSomething failed: %v", err)
    // Investigate: What error type? When does it occur?
    // THEN implement proper fix based on understanding
}
```

### ❌ Symptom Treatment

```go
// WRONG: Fixing symptom, not cause
if user == nil {
    user = &User{} // Prevent nil panic, but WHY is it nil?
}
```

### ✅ Root Cause Fix

```go
// RIGHT: Fix why user is nil
// Bug was: Forgot to load user before this point
user, err := loadUser(userID)
if err != nil {
    return fmt.Errorf("load user: %w", err)
}
```

### ❌ Skipping Reproduction

Starting to fix before you can reliably reproduce wastes time. You won't know if your fix works.

### ✅ Reliable Reproduction First

1. Write failing test that reproduces bug
2. Fix the bug
3. Test now passes
4. Bug won't return because test prevents it

## Debugging Tools

### Logging

Add strategic log statements:

```go
log.Printf("TRACE: Entering function with params: %+v", params)
// ... do work ...
log.Printf("TRACE: Exiting function with result: %+v, err: %v", result, err)
```

### Debugger (Delve for Go)

```bash
dlv debug ./cmd/hex
(dlv) break main.go:42
(dlv) continue
(dlv) print variable
(dlv) next
```

### Testing in Isolation

```go
func TestSpecificBug(t *testing.T) {
    // Minimal reproduction of bug
    input := "problematic input"
    result, err := BuggyFunction(input)

    // Document expected vs actual
    t.Logf("Input: %v", input)
    t.Logf("Expected: X, Got: %v", result)
    t.Logf("Error: %v", err)
}
```

## Common Bug Patterns

### Nil Pointers

**Symptom**: Panic: nil pointer dereference

**Investigation**:
- Where was the pointer supposed to be initialized?
- What path led to it being nil?
- Is there missing error checking?

### Race Conditions

**Symptom**: Intermittent failures, different results each run

**Investigation**:
- Are multiple goroutines accessing shared data?
- Is there missing synchronization (mutexes, channels)?
- Run with `go test -race`

### Resource Leaks

**Symptom**: Performance degrades over time

**Investigation**:
- Are resources being closed (files, connections, goroutines)?
- Check for `defer` statements
- Monitor resource usage over time

### Off-by-One Errors

**Symptom**: Wrong results, array bounds errors

**Investigation**:
- Check loop boundaries (`< vs <=`, `> vs >=`)
- Verify slice/array indexing
- Test edge cases (empty, single element, boundary values)

## Checklist for Every Bug

- [ ] Can reproduce the bug reliably
- [ ] Read and understood error message completely
- [ ] Checked logs for additional context
- [ ] Identified pattern (when bug occurs vs doesn't)
- [ ] Formed hypothesis about root cause
- [ ] Tested hypothesis with evidence
- [ ] Identified actual root cause (not symptom)
- [ ] Implemented minimal fix
- [ ] Verified fix resolves original bug
- [ ] Verified fix doesn't break anything else
- [ ] Added test to prevent regression
- [ ] Documented the issue if complex

## Summary

Systematic debugging replaces guesswork with methodology:

1. **Investigate**: Gather evidence, form hypothesis
2. **Analyze**: Find patterns in when bug occurs
3. **Test**: Prove hypothesis with data
4. **Fix**: Address root cause, verify completely

**Remember**: Understanding the problem IS the solution. Once you truly understand why a bug occurs, the fix is usually obvious.
