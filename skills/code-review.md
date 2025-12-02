---
name: code-review
description: Comprehensive code review checklist for pull requests and code changes
tags:
  - code-review
  - quality
  - pr
  - pull-request
  - review
activationPatterns:
  - "review.*(pr|pull request|code|changes)"
  - "code review"
  - "review.*before.*merge"
priority: 7
version: 1.0.0
---

# Code Review Checklist

## Purpose

Code review ensures quality, catches bugs, shares knowledge, and maintains consistency. Every pull request should be reviewed thoroughly before merging.

## Pre-Review: Automated Checks

Before starting manual review, verify automated checks pass:

- [ ] CI/CD pipeline passes (all stages green)
- [ ] All tests pass locally and in CI
- [ ] Code coverage maintained or improved
- [ ] Linting passes (no warnings)
- [ ] Type checking passes (for typed languages)
- [ ] Security scanner passes
- [ ] Build succeeds

**If automated checks fail, send back without review.** Don't review code that doesn't build or pass tests.

## 1. Functionality Review

Does the code actually work correctly?

- [ ] Code implements what PR description claims
- [ ] Requirements from ticket/issue are met
- [ ] Edge cases are handled appropriately
- [ ] Error cases are handled gracefully
- [ ] Input validation is present where needed
- [ ] No obvious bugs or logic errors
- [ ] Boundary conditions are correct (off-by-one, etc.)

**Test**: Try to think of inputs that might break it.

## 2. Code Quality

Is the code readable, maintainable, and well-structured?

### Naming

- [ ] Variables/functions have clear, descriptive names
- [ ] Names follow project conventions (camelCase, snake_case, etc.)
- [ ] Boolean variables/functions use positive names (`isValid` not `isNotInvalid`)
- [ ] Acronyms are used consistently

### Structure

- [ ] Functions are focused (single responsibility)
- [ ] Functions are reasonably sized (< 50 lines ideal)
- [ ] No code duplication (DRY principle)
- [ ] Complex logic is broken into smaller functions
- [ ] Nesting depth is reasonable (< 3 levels ideal)

### Comments

- [ ] Comments explain WHY, not WHAT
- [ ] Complex algorithms have explanations
- [ ] Public APIs have documentation
- [ ] No commented-out code (use git history)
- [ ] No misleading or outdated comments

**Examples**:

```go
// ❌ Bad: States the obvious
// Increment counter by 1
counter++

// ✅ Good: Explains why
// Skip first line since it's a header
counter++

// ❌ Bad: Commented-out code
// oldImplementation()
newImplementation()

// ✅ Good: Removed, check git history if needed
newImplementation()
```

## 3. Testing

Is the code adequately tested?

- [ ] New functionality has tests
- [ ] Tests cover happy path (normal cases)
- [ ] Tests cover edge cases (empty, null, boundary values)
- [ ] Tests cover error cases (invalid input, failures)
- [ ] Test names clearly describe what they test
- [ ] Tests are reliable (not flaky)
- [ ] No tests of implementation details (test behavior, not internals)
- [ ] Mocks/stubs are justified (not overused)

**Example**:

```go
// ✅ Good test structure
func TestCalculateTotal_WithValidItems(t *testing.T) { /* ... */ }
func TestCalculateTotal_WithEmptyCart(t *testing.T) { /* ... */ }
func TestCalculateTotal_WithNegativePrices(t *testing.T) { /* ... */ }
```

## 4. Security

Are there security vulnerabilities?

- [ ] No hardcoded secrets (API keys, passwords, tokens)
- [ ] User input is sanitized/validated
- [ ] SQL injection is prevented (parameterized queries)
- [ ] XSS is prevented (escaped output)
- [ ] Authentication is required where needed
- [ ] Authorization checks are in place
- [ ] Sensitive data is encrypted (passwords, PII)
- [ ] No logging of sensitive information

**Common Issues**:

```go
// ❌ SQL Injection vulnerability
query := fmt.Sprintf("SELECT * FROM users WHERE id = %s", userID)

// ✅ Parameterized query
query := "SELECT * FROM users WHERE id = ?"
db.Query(query, userID)

// ❌ Hardcoded secret
apiKey := "sk_live_abc123xyz"

// ✅ Environment variable
apiKey := os.Getenv("API_KEY")
```

## 5. Performance

Will this code perform well?

- [ ] No N+1 query problems
- [ ] Database queries have appropriate indexes
- [ ] No unnecessary loops or redundant operations
- [ ] Large datasets are handled efficiently (streaming, pagination)
- [ ] No unbounded memory allocations
- [ ] Caching is used appropriately
- [ ] Expensive operations are async where possible

**Example**:

```go
// ❌ N+1 query problem
for _, user := range users {
    posts := db.GetPostsByUserID(user.ID) // Query in loop!
}

// ✅ Batch query
userIDs := extractIDs(users)
posts := db.GetPostsByUserIDs(userIDs) // Single query
```

## 6. Error Handling

Are errors handled properly?

- [ ] Errors are checked (not ignored)
- [ ] Error messages are informative
- [ ] Errors are wrapped with context
- [ ] Errors are logged at appropriate level
- [ ] Resources are cleaned up on error (defer, try-finally)
- [ ] Partial failures are handled (atomicity, rollbacks)

**Example**:

```go
// ❌ Ignored error
file, _ := os.Open("file.txt")

// ✅ Error checked and handled
file, err := os.Open("file.txt")
if err != nil {
    return fmt.Errorf("open config file: %w", err)
}
defer file.Close()
```

## 7. Documentation

Is the change properly documented?

- [ ] Public API changes are documented
- [ ] Breaking changes are clearly noted
- [ ] README updated if needed
- [ ] Architecture docs updated if needed
- [ ] Migration guide provided if needed
- [ ] Examples updated if API changed

## 8. Consistency

Does the code fit with the existing codebase?

- [ ] Follows project coding standards
- [ ] Matches existing patterns/idioms
- [ ] Dependencies are justified
- [ ] No reinventing the wheel (uses existing utilities)
- [ ] File/directory structure matches conventions

## Review Process

### 1. Read PR Description

Understand what the change is supposed to do and why.

### 2. Review Changes File-by-File

- Start with tests to understand expected behavior
- Then review implementation
- Check for unintended changes (formatting, whitespace)

### 3. Test Locally (Optional)

For complex changes:
```bash
git fetch origin pull/123/head:pr-123
git checkout pr-123
go test ./...
# Manual testing if needed
```

### 4. Leave Specific Feedback

**Good Feedback**:
```markdown
**File**: internal/auth/handler.go
**Line**: 42
**Severity**: High
**Issue**: Password is compared with == instead of constant-time comparison
**Fix**: Use bcrypt.CompareHashAndPassword(hash, password)
**Why**: Timing attacks can leak password information
```

**Poor Feedback**:
```markdown
This doesn't look right.
```

### 5. Categorize Feedback

Use labels:
- **MUST FIX**: Bugs, security issues, breaking problems
- **SHOULD FIX**: Code quality, best practices
- **NICE TO HAVE**: Suggestions, alternatives
- **QUESTION**: Clarifications needed

### 6. Approve or Request Changes

- **Approve**: All MUST FIX items resolved, code is good
- **Request Changes**: MUST FIX items remain
- **Comment**: Feedback provided but no blocking issues

## Giving Good Feedback

### Be Specific

❌ "This function is confusing"
✅ "This function does three different things. Consider splitting into `validateInput()`, `processData()`, and `saveResult()`"

### Be Kind

❌ "This is terrible code"
✅ "We typically use X pattern for this. See example in auth/handler.go:42"

### Explain Why

❌ "Use a constant here"
✅ "Use a constant here so the timeout value can be configured without code changes"

### Suggest Solutions

❌ "This won't work"
✅ "This assumes users is non-empty. Add `if len(users) == 0` check or document this precondition"

### Acknowledge Good Work

Don't only point out problems:
✅ "Nice refactoring of the error handling!"
✅ "Great test coverage on the edge cases"

## Common Review Findings

### Magic Numbers

```go
// ❌ Magic number
time.Sleep(300 * time.Millisecond)

// ✅ Named constant
const retryDelay = 300 * time.Millisecond
time.Sleep(retryDelay)
```

### Error Shadowing

```go
// ❌ Error variable shadowed
result, err := step1()
if err != nil {
    result, err := step2() // Shadows outer err!
    // Outer err is not returned
}
return result, err
```

### Missing Validation

```go
// ❌ No validation
func ProcessUser(user *User) {
    db.Save(user) // What if user is nil?
}

// ✅ Validated
func ProcessUser(user *User) error {
    if user == nil {
        return errors.New("user cannot be nil")
    }
    return db.Save(user)
}
```

## Summary

Code review is about:
- **Correctness**: Does it work?
- **Quality**: Is it maintainable?
- **Security**: Is it safe?
- **Performance**: Is it efficient?
- **Consistency**: Does it fit?

Take time to review thoroughly. Finding bugs in review is much cheaper than finding them in production.

**Remember**: Review the code, not the person. We're all working toward better software together.
