---
name: test
description: Write comprehensive tests following TDD principles
args:
  target: What to test - function, file, or feature (optional)
  type: Test type - "unit", "integration", or "e2e" (optional)
---

# Test Writing Workflow

You are writing tests {{if .target}}for {{.target}}{{else}}for the current functionality{{end}}{{if .type}} ({{.type}} tests){{end}}.

## Test-Driven Development Process

Follow the Red-Green-Refactor cycle:

### 1. RED: Write Failing Test

**Before writing any implementation code:**

1. **Identify Behavior**
   - What should this code do?
   - What are the inputs and expected outputs?
   - What edge cases exist?

2. **Write Test First**
   ```
   func TestFeatureName(t *testing.T) {
       // Arrange: Set up test data
       input := "test input"

       // Act: Call the function
       result := FunctionToTest(input)

       // Assert: Verify expected behavior
       expected := "expected output"
       if result != expected {
           t.Errorf("Expected %v, got %v", expected, result)
       }
   }
   ```

3. **Run Test**
   - Verify the test fails
   - Check that failure message is clear
   - Confirm test is failing for the right reason

### 2. GREEN: Write Minimal Code

**Write just enough code to make the test pass:**

1. **Implement**
   - Add minimal code to satisfy the test
   - Don't add extra features or optimizations yet
   - Focus on making this one test pass

2. **Run Test Again**
   - Verify the test now passes
   - Check all other tests still pass
   - No regression allowed

### 3. REFACTOR: Improve Code

**Now that tests are green, improve the design:**

1. **Clean Up**
   - Remove duplication
   - Improve names
   - Simplify logic
   - Add comments if needed

2. **Verify Tests Still Pass**
   - Run all tests after each refactor
   - Tests should remain green
   - If tests fail, revert and try again

3. **Repeat**
   - Return to step 1 for next behavior
   - Build functionality incrementally

## Test Coverage Requirements

Write tests for:

### Unit Tests ({{if eq .type "unit"}}← FOCUS ON THIS{{end}})
- [ ] Happy path (normal inputs, expected outputs)
- [ ] Edge cases (empty, nil, zero, max values)
- [ ] Error conditions (invalid inputs, failures)
- [ ] Boundary conditions (min/max, first/last)
- [ ] State changes (before/after verification)

### Integration Tests ({{if eq .type "integration"}}← FOCUS ON THIS{{end}})
- [ ] Component interactions
- [ ] Database operations
- [ ] API calls
- [ ] File I/O
- [ ] External dependencies

### End-to-End Tests ({{if eq .type "e2e"}}← FOCUS ON THIS{{end}})
- [ ] Complete user workflows
- [ ] System-wide functionality
- [ ] Real data scenarios
- [ ] Performance under load

## Test Quality Checklist

**Every test should be:**

- [ ] **Independent**: Can run in any order
- [ ] **Repeatable**: Same results every time
- [ ] **Fast**: Runs in milliseconds (unit tests)
- [ ] **Clear**: Purpose obvious from name and structure
- [ ] **Focused**: Tests one thing at a time
- [ ] **Maintainable**: Easy to update when code changes

**Avoid:**
- ❌ Testing implementation details (test behavior, not internals)
- ❌ Brittle tests (that break with unrelated changes)
- ❌ Slow tests (that slow down development)
- ❌ Flaky tests (that pass/fail randomly)
- ❌ Unclear failures (error messages that don't help)

## Test Structure

Use Arrange-Act-Assert pattern:

```go
func TestCalculateTotal(t *testing.T) {
    // Arrange: Set up test conditions
    items := []Item{
        {Price: 10.00, Quantity: 2},
        {Price: 5.50, Quantity: 1},
    }

    // Act: Execute the behavior being tested
    total := CalculateTotal(items)

    // Assert: Verify the expected outcome
    expected := 25.50
    if total != expected {
        t.Errorf("Expected total %.2f, got %.2f", expected, total)
    }
}
```

## Table-Driven Tests

For multiple test cases:

```go
func TestValidateEmail(t *testing.T) {
    tests := []struct {
        name    string
        email   string
        want    bool
    }{
        {"valid email", "user@example.com", true},
        {"missing @", "userexample.com", false},
        {"missing domain", "user@", false},
        {"empty string", "", false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := ValidateEmail(tt.email)
            if got != tt.want {
                t.Errorf("ValidateEmail(%q) = %v, want %v",
                    tt.email, got, tt.want)
            }
        })
    }
}
```

## Output Format

Document your test plan:

```markdown
## Test Plan: [Feature/Function Name]

### Test Type
[Unit / Integration / E2E]

### Behaviors to Test
1. [Behavior 1]: [Expected outcome]
2. [Behavior 2]: [Expected outcome]
3. [Error case]: [Expected error]

### Test Cases

#### Test 1: [Test Name]
**Purpose**: Verify [specific behavior]
**Input**: [Test input]
**Expected**: [Expected output]
**Status**: [ ] TODO / [x] PASS

#### Test 2: [Test Name]
...

### Coverage
- [ ] Happy path
- [ ] Edge cases
- [ ] Error handling
- [ ] Boundary conditions
```

## Implementation Steps

1. Write failing test
2. Run test to verify it fails
3. Write minimal implementation
4. Run test to verify it passes
5. Refactor if needed
6. Run all tests to verify no regressions
7. Repeat for next test case

Begin writing tests now following TDD principles.
