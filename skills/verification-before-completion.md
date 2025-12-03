---
name: verification-before-completion
description: Always verify work by running commands and confirming output before claiming completion
tags:
  - verification
  - testing
  - quality
  - completion
activationPatterns:
  - "done"
  - "completed"
  - "finished"
  - "fixed"
  - "passing"
priority: 10
version: 1.0.0
---

# Verification Before Completion

## Core Principle

**NEVER claim work is complete, fixed, or passing without running verification commands and confirming the output.**

Evidence before assertions. Always.

## The Problem

Claiming success without verification leads to:
- Features that don't actually work
- "Fixes" that don't fix anything
- Tests that still fail
- Broken builds
- Wasted time debugging "completed" work

## The Solution

Before saying "done", "fixed", "passing", or "working":

1. **Run the verification command**
2. **Read the output completely**
3. **Confirm success**
4. **Only then claim completion**

## Verification Checklist

### Before Claiming "Tests Pass"

```bash
# ❌ WRONG: Don't just assume
"I've added tests, they should pass"

# ✅ RIGHT: Actually run them
$ go test ./internal/skills/...
ok      github.com/2389-research/hex/internal/skills  0.123s
```

**Verify**:
- [ ] `go test` command actually ran
- [ ] All tests passed (no FAIL)
- [ ] No panics or errors in output
- [ ] Coverage is adequate

### Before Claiming "Build Succeeds"

```bash
# ❌ WRONG: Assume it compiles
"The code looks good, should build fine"

# ✅ RIGHT: Build it
$ go build ./cmd/hex
$ echo $?
0
```

**Verify**:
- [ ] Build command completed
- [ ] Exit code is 0
- [ ] No compilation errors
- [ ] Binary was created

### Before Claiming "Linter Passes"

```bash
# ❌ WRONG: Assume no issues
"I followed the style guide"

# ✅ RIGHT: Run linter
$ golangci-lint run ./internal/skills/
$ # No output = success
```

**Verify**:
- [ ] Linter ran successfully
- [ ] No warnings or errors
- [ ] Exit code is 0

### Before Claiming "Bug Fixed"

```bash
# ❌ WRONG: Assume fix worked
"I changed the code, the bug should be gone"

# ✅ RIGHT: Reproduce and verify
# 1. Reproduce original bug
$ go test -run TestBuggyFunction
FAIL

# 2. Apply fix

# 3. Verify bug is gone
$ go test -run TestBuggyFunction
PASS
```

**Verify**:
- [ ] Could reproduce bug before fix
- [ ] Bug doesn't occur after fix
- [ ] Related tests pass
- [ ] No new bugs introduced

### Before Claiming "Feature Implemented"

```bash
# ❌ WRONG: Wrote code, didn't test
"Feature is implemented"

# ✅ RIGHT: Test it end-to-end
$ go build ./cmd/hex
$ ./hex --new-feature
Expected output appears
```

**Verify**:
- [ ] Feature builds
- [ ] Feature runs without errors
- [ ] Feature produces expected output
- [ ] Tests for feature pass
- [ ] Documentation matches behavior

### Before Claiming "Deployment Succeeded"

```bash
# ❌ WRONG: Pushed to server
"I deployed it"

# ✅ RIGHT: Verify it's running
$ curl https://api.example.com/health
{"status": "ok", "version": "1.2.3"}

$ ssh server 'systemctl status hex'
● hex.service - Hex Service
   Active: active (running)
```

**Verify**:
- [ ] Service is running
- [ ] Health check passes
- [ ] Version is correct
- [ ] Logs show no errors

## Common Scenarios

### Scenario: Test Addition

**❌ Wrong Approach**:
```
"I've added tests for the new feature"
```

**✅ Right Approach**:
```bash
$ go test ./internal/skills/ -v
=== RUN   TestSkillParsing
--- PASS: TestSkillParsing (0.00s)
=== RUN   TestSkillLoader
--- PASS: TestSkillLoader (0.00s)
PASS
ok      github.com/2389-research/hex/internal/skills  0.123s

"Tests added and verified passing. Coverage increased from 75% to 85%."
```

### Scenario: Bug Fix

**❌ Wrong Approach**:
```
"Fixed the nil pointer bug"
```

**✅ Right Approach**:
```bash
# Before fix
$ go test -run TestUserHandler
panic: runtime error: invalid memory address or nil pointer dereference
FAIL

# After fix
$ go test -run TestUserHandler
PASS

"Fixed nil pointer bug in user handler. Verified with TestUserHandler which now passes."
```

### Scenario: Refactoring

**❌ Wrong Approach**:
```
"Refactored the authentication code"
```

**✅ Right Approach**:
```bash
$ go test ./internal/auth/...
ok      github.com/2389-research/hex/internal/auth    0.456s

$ go build ./cmd/hex
$ echo $?
0

"Refactored authentication code. All tests still pass, build succeeds."
```

### Scenario: Configuration Change

**❌ Wrong Approach**:
```
"Updated the configuration"
```

**✅ Right Approach**:
```bash
$ hex --help | grep "new-flag"
  --new-flag string   New configuration option (default "value")

$ hex --new-flag=test
Configuration loaded successfully
Using new-flag: test

"Updated configuration to add new-flag. Verified with --help and test run."
```

## What to Verify

### For Code Changes

- [ ] Tests pass (`go test`)
- [ ] Builds successfully (`go build`)
- [ ] Linter passes (`golangci-lint run`)
- [ ] No new warnings
- [ ] Coverage maintained/improved

### For New Features

- [ ] Feature works as intended
- [ ] Tests cover new functionality
- [ ] Documentation updated
- [ ] Examples work
- [ ] Edge cases handled

### For Bug Fixes

- [ ] Bug reproduced first
- [ ] Bug no longer occurs
- [ ] Test added to prevent regression
- [ ] Related functionality still works

### For Refactoring

- [ ] All tests still pass
- [ ] Behavior unchanged
- [ ] No performance regression
- [ ] API compatibility maintained

### For Documentation

- [ ] Examples actually work
- [ ] Commands produce shown output
- [ ] Links are valid
- [ ] Code samples compile

## Verification Commands by Type

### Go Projects

```bash
# Run tests
go test ./...
go test -race ./...
go test -cover ./...

# Build
go build ./cmd/...

# Lint
golangci-lint run

# Format check
gofmt -l .

# Vet
go vet ./...
```

### Git Operations

```bash
# Verify commit
git log -1 --stat

# Verify push
git push origin feature-branch
git ls-remote origin feature-branch

# Verify branch
git branch -a | grep feature-branch
```

### API Testing

```bash
# Verify endpoint
curl -X GET http://localhost:8080/api/health

# Verify with data
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"name": "test"}'
```

## Red Flags

If you find yourself saying:
- "Should work" → Stop. Verify it works.
- "Probably fine" → Stop. Confirm it's fine.
- "Tests should pass" → Stop. Run the tests.
- "I think it's deployed" → Stop. Check deployment.

## Benefits

Verification before completion:
- Catches bugs before they spread
- Builds confidence in changes
- Provides proof of completion
- Prevents wasted review time
- Creates documentation (command output)

## Implementation

### Every Time You Claim Completion

1. **Identify verification command**
   - What command proves this works?

2. **Run the command**
   - Actually execute it, don't assume

3. **Read output completely**
   - Don't skim, read every line

4. **Confirm success**
   - Green tests, zero exit code, expected output

5. **Document verification**
   - Include command and output in commit/PR

**Example Commit Message**:
```
feat: add skill system to Hex

Implements skill loading, registry, and tool integration.

Verified:
$ go test ./internal/skills/
ok      github.com/2389-research/hex/internal/skills  0.234s

$ go build ./cmd/hex
$ echo $?
0

$ golangci-lint run ./internal/skills/
(no output - success)
```

## Summary

Before you say:
- "Done" → Run verification
- "Fixed" → Confirm it's fixed
- "Passing" → See the passes
- "Working" → Watch it work

**Remember**: If you didn't verify it, you don't know it works. Evidence before assertions. Always.

## Minimum Verification Standard

At minimum, before ANY claim of completion:

```bash
# 1. Tests pass
go test ./...

# 2. Builds successfully
go build ./cmd/...

# 3. No linting errors
golangci-lint run
```

If any of these fail, work is NOT complete.
