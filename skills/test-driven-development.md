---
name: test-driven-development
description: Write tests before implementation, ensuring tests fail first before writing code to pass them
tags:
  - testing
  - tdd
  - methodology
  - quality
activationPatterns:
  - "write.*test"
  - "implement.*feature"
  - "add.*functionality"
  - "create.*test"
priority: 8
version: 1.0.0
---

# Test-Driven Development (TDD)

## The Process

TDD is a disciplined approach to writing code that ensures quality through a tight feedback loop:

1. **Write a failing test**: Create a test that defines the desired functionality
2. **Run the test**: Confirm it fails for the right reason
3. **Write minimal code**: Only enough to make the test pass
4. **Run the test again**: Verify it passes
5. **Refactor**: Improve code quality while keeping tests green
6. **Repeat**: Move to the next piece of functionality

## Core Rules

- **NEVER** write implementation code before writing a test
- **ONLY** write enough code to make the current test pass
- **ALWAYS** run tests before claiming success
- **REFACTOR** continuously while tests remain green
- **TEST MUST FAIL FIRST** - if a new test passes immediately, it's not testing anything

## Example Workflow

### 1. Write the Test First

```go
func TestCalculateTotal(t *testing.T) {
    numbers := []int{1, 2, 3, 4, 5}
    expected := 15

    result := CalculateTotal(numbers)

    if result != expected {
        t.Errorf("CalculateTotal(%v) = %d; want %d", numbers, result, expected)
    }
}
```

### 2. Run Test - Watch It Fail

```bash
$ go test
# ./calculator_test.go:8:15: undefined: CalculateTotal
FAIL
```

Good! The test fails because the function doesn't exist yet.

### 3. Write Minimal Implementation

```go
func CalculateTotal(numbers []int) int {
    total := 0
    for _, n := range numbers {
        total += n
    }
    return total
}
```

### 4. Run Test - Watch It Pass

```bash
$ go test
PASS
ok      calculator      0.001s
```

### 5. Refactor If Needed

The implementation is already clean, but we could add edge case handling:

```go
func CalculateTotal(numbers []int) int {
    if len(numbers) == 0 {
        return 0
    }
    total := 0
    for _, n := range numbers {
        total += n
    }
    return total
}
```

Add test for edge case:

```go
func TestCalculateTotalEmpty(t *testing.T) {
    result := CalculateTotal([]int{})
    if result != 0 {
        t.Errorf("CalculateTotal([]) = %d; want 0", result)
    }
}
```

## When to Apply TDD

**Use TDD when:**
- Implementing new features with clear requirements
- Fixing bugs (write a test that reproduces the bug first)
- Refactoring existing code (tests verify behavior preservation)
- Working on critical code (authentication, payments, data processing)
- Building library APIs (tests document expected behavior)

**Skip TDD when:**
- Prototyping UI layouts (visual exploration)
- Exploring unfamiliar APIs (learning mode)
- Writing throwaway/spike code
- Extremely simple getters/setters

## Common Mistakes

### ❌ Writing Implementation First

```go
// WRONG: Implementation written first
func ProcessOrder(order Order) error {
    // ... complex logic ...
}

// THEN writing test
func TestProcessOrder(t *testing.T) { /* ... */ }
```

This defeats the purpose - you can't verify the test actually catches bugs.

### ✅ Test First

```go
// RIGHT: Test written first
func TestProcessOrder(t *testing.T) {
    order := Order{ID: "123", Items: []Item{{SKU: "ABC"}}}
    err := ProcessOrder(order)
    if err != nil {
        t.Errorf("unexpected error: %v", err)
    }
}

// THEN implementation
func ProcessOrder(order Order) error {
    // Minimal code to pass test
    return nil
}
```

### ❌ Not Running Test to See It Fail

You must see the test fail before writing code. Otherwise:
- Test might already pass (not testing anything new)
- Test might be broken/incorrect
- You don't know what failure looks like

### ✅ Always Verify Failure First

```bash
$ go test                    # MUST see red/failure
$ # Write implementation
$ go test                    # MUST see green/pass
```

### ❌ Writing Too Much Code at Once

```go
// WRONG: Implementing multiple features at once
func CalculateOrderTotal(order Order) float64 {
    subtotal := calculateSubtotal(order)
    tax := calculateTax(subtotal, order.TaxRate)
    shipping := calculateShipping(order)
    discount := applyDiscounts(subtotal, order.Coupons)
    return subtotal + tax + shipping - discount
}
```

This is too much. Write one test, implement one piece.

### ✅ Incremental Implementation

```go
// Test 1: Just subtotal
func TestCalculateOrderTotal_Subtotal(t *testing.T) { /* ... */ }
// Implement: return calculateSubtotal(order)

// Test 2: Add tax
func TestCalculateOrderTotal_WithTax(t *testing.T) { /* ... */ }
// Implement: return calculateSubtotal(order) + calculateTax(...)

// And so on...
```

## Benefits of TDD

1. **Confidence**: Know your code works before integration
2. **Documentation**: Tests show how code should be used
3. **Design**: Writing tests first forces better API design
4. **Safety**: Refactor without fear of breaking things
5. **Focus**: Work on one small piece at a time
6. **Debugging**: Failures caught immediately, not later

## Red-Green-Refactor Cycle

```
┌──────────────────┐
│   WRITE TEST     │  (Red - Test Fails)
│   (Failing)      │
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│ WRITE MINIMAL    │  (Green - Test Passes)
│ IMPLEMENTATION   │
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│   REFACTOR       │  (Green - Tests Still Pass)
│  (Improve Code)  │
└────────┬─────────┘
         │
         └──────────┐
                    │
                    ▼
              Next Feature
```

## Integration with Hex Workflow

When implementing features in Clem:

1. **Create failing test** in `internal/*/your_feature_test.go`
2. **Run test** with `go test ./internal/...`
3. **Write minimal code** to make it pass
4. **Run test again** to verify success
5. **Refactor** if needed
6. **Commit** with test and implementation together

## Summary

TDD is not about writing more tests - it's about writing tests FIRST. This simple discipline:
- Catches bugs before they exist
- Forces clear thinking about requirements
- Produces better-designed code
- Gives confidence to refactor
- Creates living documentation

**Remember**: Red → Green → Refactor. Always in that order.
