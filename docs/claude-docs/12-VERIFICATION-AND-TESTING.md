# Verification and Testing

**ABOUTME: Comprehensive guide to testing strategies, verification workflows, and quality gates in Claude Code**
**ABOUTME: Covers pre-commit hooks, test output interpretation, coverage checking, and verification best practices**

## Overview

Claude Code enforces a **verify-before-complete** philosophy. This document covers testing strategies, verification workflows, and quality gates that ensure code correctness before declaring tasks complete.

## Core Principles

### 1. Verification Protocol

**NEVER claim work is complete without verification:**

```bash
# ❌ WRONG - No verification
# "I've fixed the bug" [without running tests]

# ✅ CORRECT - Verify first
npm test                    # Run and observe output
# "Tests pass - bug is fixed"
```

**Verification checklist before completion:**
- [ ] Tests run and pass
- [ ] Test output is clean (no warnings/errors)
- [ ] Functionality verified in target environment
- [ ] Pre-commit hooks pass
- [ ] No regression in existing features

### 2. Test-Driven Development (TDD)

Claude Code follows strict TDD principles:

**RED → GREEN → REFACTOR cycle:**

```python
# 1. RED: Write failing test
def test_user_authentication():
    result = authenticate_user("admin", "password123")
    assert result.is_authenticated == True
    assert result.user_id is not None

# Run: pytest test_auth.py
# Expected: FAIL (authenticate_user doesn't exist yet)

# 2. GREEN: Write minimal code to pass
def authenticate_user(username, password):
    # Minimal implementation
    if username == "admin" and password == "password123":
        return AuthResult(is_authenticated=True, user_id=1)
    return AuthResult(is_authenticated=False, user_id=None)

# Run: pytest test_auth.py
# Expected: PASS

# 3. REFACTOR: Improve design while keeping tests green
def authenticate_user(username, password):
    # Proper implementation with database lookup
    user = db.users.find_by_username(username)
    if user and user.verify_password(password):
        return AuthResult(is_authenticated=True, user_id=user.id)
    return AuthResult(is_authenticated=False, user_id=None)

# Run: pytest test_auth.py
# Expected: PASS (still green after refactor)
```

**TDD workflow in Claude Code:**

1. **Write test before implementation** - Always
2. **Run test to confirm it fails** - Verify the test actually tests something
3. **Write minimal code to pass** - Don't over-engineer
4. **Refactor while keeping tests green** - Improve design iteratively
5. **Never skip steps** - Each phase has a purpose

### 3. Test Coverage Requirements

**NO EXCEPTIONS POLICY**:

> Under no circumstances should you mark any test type as "not applicable". Every project, regardless of size or complexity, MUST have unit tests, integration tests, AND end-to-end tests.

**Required test types:**

| Test Type | Purpose | Example |
|-----------|---------|---------|
| **Unit Tests** | Test individual functions/classes in isolation | `test_calculate_total()` |
| **Integration Tests** | Test component interactions | `test_api_database_integration()` |
| **End-to-End Tests** | Test complete user workflows | `test_user_signup_flow()` |

**Skipping tests requires explicit authorization:**

```
User: "I AUTHORIZE YOU TO SKIP WRITING TESTS THIS TIME"
```

Without this exact phrase, ALL test types are mandatory.

## Testing Strategies

### Unit Testing

**Test individual components in isolation:**

```python
# test_calculator.py
import pytest
from calculator import add, subtract, multiply, divide

def test_add():
    assert add(2, 3) == 5
    assert add(-1, 1) == 0
    assert add(0, 0) == 0

def test_divide():
    assert divide(10, 2) == 5
    with pytest.raises(ZeroDivisionError):
        divide(10, 0)

def test_multiply_edge_cases():
    assert multiply(5, 0) == 0
    assert multiply(-3, 4) == -12
```

**Run unit tests:**

```bash
# Python
pytest tests/unit/ -v

# JavaScript/TypeScript
npm test -- --testPathPattern=unit

# Go
go test ./... -short

# Rust
cargo test --lib
```

### Integration Testing

**Test component interactions:**

```python
# test_api_integration.py
import pytest
from app import create_app
from database import db

@pytest.fixture
def client():
    app = create_app('testing')
    with app.test_client() as client:
        with app.app_context():
            db.create_all()
            yield client
            db.drop_all()

def test_user_registration_flow(client):
    # POST to register endpoint
    response = client.post('/api/register', json={
        'username': 'testuser',
        'email': 'test@example.com',
        'password': 'secure123'
    })
    assert response.status_code == 201

    # Verify user in database
    user = db.users.find_by_username('testuser')
    assert user is not None
    assert user.email == 'test@example.com'

    # Verify login works
    response = client.post('/api/login', json={
        'username': 'testuser',
        'password': 'secure123'
    })
    assert response.status_code == 200
    assert 'token' in response.json
```

**Run integration tests:**

```bash
# Python with fixtures
pytest tests/integration/ -v

# JavaScript with test containers
npm test -- --testPathPattern=integration

# Go with test database
go test ./... -tags=integration
```

### End-to-End Testing

**Test complete user workflows:**

```python
# test_e2e_checkout.py
from playwright.sync_api import Page, expect

def test_complete_checkout_flow(page: Page):
    # Navigate to site
    page.goto('http://localhost:3000')

    # Browse products
    page.click('text=Shop Now')
    expect(page.locator('.product-grid')).to_be_visible()

    # Add item to cart
    page.click('.product-card:first-child .add-to-cart')
    expect(page.locator('.cart-badge')).to_have_text('1')

    # Proceed to checkout
    page.click('.cart-icon')
    page.click('text=Checkout')

    # Fill checkout form
    page.fill('#email', 'customer@example.com')
    page.fill('#card-number', '4242424242424242')
    page.fill('#expiry', '12/25')
    page.fill('#cvc', '123')

    # Complete purchase
    page.click('text=Place Order')
    expect(page.locator('.success-message')).to_contain_text('Order confirmed')
```

**Run e2e tests:**

```bash
# Playwright
npx playwright test

# Cypress
npx cypress run

# Selenium
pytest tests/e2e/ --browser=chrome
```

## Pre-Commit Verification

### Pre-Commit Hooks

**MANDATORY protocol:**

> When pre-commit hooks fail, you MUST follow this exact sequence before any commit attempt:
>
> 1. Read the complete error output aloud (explain what you're seeing)
> 2. Identify which tool failed (biome, ruff, tests, etc.) and why
> 3. Explain the fix you will apply and why it addresses the root cause
> 4. Apply the fix and re-run hooks
> 5. Only proceed with commit after all hooks pass

**FORBIDDEN: `--no-verify`, `--no-hooks`, `--no-pre-commit-hook`**

### Common Hook Failures

#### Linter Failures (Biome, Ruff, ESLint)

```bash
# Example failure output
$ git commit -m "Add feature"
biome....................................................................Failed
- hook id: biome
- exit code: 1

src/utils.ts:15:1: error: Unexpected var, use let or const instead
  15 | var result = calculate(x, y);
     | ^^^

# ✅ CORRECT response:
# "The biome hook failed because line 15 uses 'var' instead of 'let/const'.
#  I'll fix this by changing var to const since result isn't reassigned."
```

**Fix and retry:**

```bash
# Fix the code
# (use Edit tool to change var to const)

# Re-run hooks
git add src/utils.ts
git commit -m "Add feature"
# ✅ All hooks pass - proceed
```

#### Test Failures in Hooks

```bash
# Example failure
$ git commit -m "Fix bug"
pytest...................................................................Failed
- hook id: pytest
- exit code: 1

tests/test_auth.py::test_login FAILED

def test_login():
>   assert login("admin", "wrong") == False
E   AssertionError: assert True == False

# ✅ CORRECT response:
# "The pytest hook failed - test_login expects wrong password to return False
#  but the function returns True. This indicates my bug fix broke authentication.
#  I need to investigate the login function logic."
```

**Fix and retry:**

```bash
# Fix the actual bug (not the test!)
# Re-run tests manually first
pytest tests/test_auth.py -v

# Once tests pass, retry commit
git commit -m "Fix bug"
# ✅ Tests pass - proceed
```

#### Formatting Failures (Black, Prettier, gofmt)

```bash
# Example failure
$ git commit -m "Update code"
black....................................................................Failed
- hook id: black
- exit code: 1

would reformat src/main.py
1 file would be reformatted

# ✅ CORRECT response:
# "Black hook failed because src/main.py needs reformatting.
#  I'll run black to auto-format the file."
```

**Fix automatically:**

```bash
# Run formatter
black src/main.py

# Retry commit
git add src/main.py
git commit -m "Update code"
# ✅ Formatting passes - proceed
```

### Hook Debugging

**If hooks fail unexpectedly:**

1. **Run hook manually:**
   ```bash
   pre-commit run <hook-id> --files <file>
   ```

2. **Check hook configuration:**
   ```bash
   cat .pre-commit-config.yaml
   ```

3. **Verify hook installation:**
   ```bash
   pre-commit install --install-hooks
   ```

4. **Update hooks:**
   ```bash
   pre-commit autoupdate
   ```

## Test Output Interpretation

### Reading Test Output

**Key elements to observe:**

```bash
# Example pytest output
================================ test session starts =================================
platform darwin -- Python 3.12.1, pytest-8.0.0, pluggy-1.4.0
rootdir: /Users/harper/project
collected 47 items

tests/test_auth.py ....                                                        [  8%]
tests/test_api.py .....F.                                                      [ 23%]
tests/test_database.py ...................                                     [ 63%]
tests/test_utils.py .................                                          [100%]

====================================== FAILURES ======================================
________________________________ test_api_rate_limit _________________________________

    def test_api_rate_limit():
>       response = client.get('/api/data', headers={'X-API-Key': 'test'})
E       assert response.status_code == 429
E       AssertionError: assert 200 == 429

tests/test_api.py:45: AssertionError
========================== 1 failed, 46 passed in 2.31s ==========================
```

**What to extract:**
- ✅ **46 tests passed** - Most functionality works
- ❌ **1 test failed** - `test_api_rate_limit` in `tests/test_api.py:45`
- 🔍 **Failure reason** - Expected status 429 (rate limit), got 200 (success)
- 📍 **Location** - `tests/test_api.py:45`

**Next steps:**
1. Read `tests/test_api.py:45` to understand test intent
2. Check rate limiting implementation
3. Fix bug and re-run

### Pristine Test Output Requirement

**Policy:**

> TEST OUTPUT MUST BE PRISTINE TO PASS

This means:
- ✅ All tests pass (100% pass rate)
- ✅ No warnings in output
- ✅ No deprecation notices
- ✅ No error logs (unless explicitly tested)
- ✅ Clean, expected output only

**Example of non-pristine output:**

```bash
# ❌ NOT PRISTINE - has warnings
$ pytest
======================== test session starts =========================
tests/test_main.py .....                                       [100%]

======================== warnings summary ============================
tests/test_main.py::test_deprecated
  /app/utils.py:12: DeprecationWarning: old_function is deprecated
    warnings.warn("old_function is deprecated")

-- Docs: https://docs.pytest.org/en/stable/warnings.html
===================== 5 passed, 1 warning in 0.12s ==================
```

**Fix warnings before declaring done:**

```python
# Fix by updating to new function
def test_deprecated():
    # result = old_function()  # Deprecated
    result = new_function()    # Updated
    assert result == expected
```

### Coverage Checking

**Verify test coverage:**

```bash
# Python - pytest-cov
pytest --cov=src --cov-report=term-missing

# JavaScript - jest
npm test -- --coverage

# Go - built-in
go test -cover ./...

# Rust - tarpaulin
cargo tarpaulin --out Stdout
```

**Interpreting coverage reports:**

```bash
# Example coverage output
Name                Stmts   Miss  Cover   Missing
-------------------------------------------------
src/auth.py            45      2    96%   78-79
src/api.py             67      0   100%
src/database.py        89     12    87%   45-52, 103-106
src/utils.py           34      1    97%   89
-------------------------------------------------
TOTAL                 235     15    94%
```

**What to look for:**
- ✅ Overall coverage >90% (adjust per project standards)
- 🔍 Missing lines indicate untested code paths
- ⚠️ Low coverage in critical files (auth, payments) is HIGH RISK

**Adding missing coverage:**

```python
# src/auth.py lines 78-79 not covered
def authenticate(username, password):
    user = db.find_user(username)
    if not user:
        return None
    if user.verify_password(password):  # Line 78
        return user                      # Line 79
    return None

# Add test for successful authentication
def test_authenticate_success():
    user = authenticate("admin", "correct_password")
    assert user is not None
    assert user.username == "admin"
```

## Verification Workflows

### Feature Verification Workflow

```
1. Run existing tests        → Ensure no regression
2. Run new tests             → Verify feature works
3. Manual verification       → Test in real environment
4. Check test coverage       → Ensure adequate coverage
5. Run pre-commit hooks      → Ensure code quality
6. Declare complete          → Only after all pass
```

**Example:**

```bash
# 1. Run existing tests
pytest tests/ -v
# ✅ 146 passed

# 2. Run new tests for feature
pytest tests/test_new_feature.py -v
# ✅ 12 passed

# 3. Manual verification
python -m app.cli test-feature --input sample.txt
# ✅ Output matches expected

# 4. Check coverage
pytest --cov=src.new_feature --cov-report=term
# ✅ 98% coverage

# 5. Pre-commit hooks
git add .
git commit -m "Add new feature"
# ✅ All hooks pass

# 6. Declare complete
# "Feature implementation complete - all tests pass, coverage at 98%"
```

### Bug Fix Verification Workflow

```
1. Reproduce bug             → Confirm bug exists
2. Write failing test        → Test demonstrates bug
3. Fix bug                   → Implement solution
4. Verify test passes        → Solution works
5. Run all tests             → No regression
6. Verify fix manually       → Real-world check
```

**Example:**

```bash
# 1. Reproduce bug
curl http://localhost:8000/api/users/123
# ❌ Returns 500 Internal Server Error

# 2. Write failing test
# test_api.py
def test_get_user_by_id():
    response = client.get('/api/users/123')
    assert response.status_code == 200

pytest tests/test_api.py::test_get_user_by_id
# ❌ FAILED - AssertionError: assert 500 == 200

# 3. Fix bug
# (Edit api.py to handle missing users correctly)

# 4. Verify test passes
pytest tests/test_api.py::test_get_user_by_id
# ✅ PASSED

# 5. Run all tests
pytest tests/ -v
# ✅ 147 passed

# 6. Manual verification
curl http://localhost:8000/api/users/123
# ✅ Returns 200 with user data
```

### Refactoring Verification Workflow

```
1. Run tests before          → Establish baseline (all green)
2. Refactor code             → Change structure, not behavior
3. Run tests after           → Verify still all green
4. Check performance         → Ensure no regression
5. Review diff               → Verify only structure changed
```

**Example:**

```bash
# 1. Run tests before
pytest tests/ -v
# ✅ 147 passed in 2.34s

# 2. Refactor code
# (Extract method, rename variables, reorganize)

# 3. Run tests after
pytest tests/ -v
# ✅ 147 passed in 2.31s

# 4. Check performance
pytest tests/ --durations=10
# ✅ No significant slowdown

# 5. Review diff
git diff src/refactored_module.py
# ✅ Only structural changes, no logic changes
```

## CI/CD Integration

### GitHub Actions Example

```yaml
# .github/workflows/test.yml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v4

      - name: Set up Python
        uses: actions/setup-python@v5
        with:
          python-version: '3.12'

      - name: Install dependencies
        run: |
          pip install -r requirements.txt
          pip install pytest pytest-cov

      - name: Run unit tests
        run: pytest tests/unit/ -v

      - name: Run integration tests
        run: pytest tests/integration/ -v

      - name: Run e2e tests
        run: pytest tests/e2e/ -v

      - name: Check coverage
        run: |
          pytest --cov=src --cov-report=term --cov-fail-under=90

      - name: Upload coverage
        uses: codecov/codecov-action@v3
```

### Pre-merge Quality Gates

**Minimum requirements before merge:**

- [ ] All tests pass (unit, integration, e2e)
- [ ] Code coverage ≥ 90% (or project standard)
- [ ] Pre-commit hooks pass locally
- [ ] CI/CD pipeline passes
- [ ] No linter errors or warnings
- [ ] Manual verification in staging environment
- [ ] Code review approved (if applicable)

## Common Testing Pitfalls

### ❌ Don't Mock Without Understanding

**Policy:**

> NEVER implement a mock mode for testing or for any purpose. Always test with real dependencies or sanctioned test environments.

**Wrong approach:**

```python
# ❌ BAD - Mocking hides bugs
def test_fetch_user_data():
    with mock.patch('api.fetch_user') as mock_fetch:
        mock_fetch.return_value = {'id': 1, 'name': 'Test'}
        result = process_user(1)
        assert result.name == 'Test'
```

**Correct approach:**

```python
# ✅ GOOD - Use real test environment
def test_fetch_user_data(test_client, test_database):
    # Set up real test user in test database
    test_user = test_database.create_user(id=1, name='Test')

    # Test with real API call to test environment
    result = process_user(1)
    assert result.name == 'Test'
```

### ❌ Don't Ignore Test Output

**Wrong:**

```bash
$ pytest
# Output shows warnings and 2 failures
# "Tests mostly pass, proceeding with commit"
```

**Correct:**

```bash
$ pytest
# Read entire output
# Fix all warnings
# Fix all failures
# Re-run until completely clean
# Only then proceed
```

### ❌ Don't Skip Test Types

**Wrong:**

```
# "This is a simple script, it doesn't need integration tests"
```

**Correct:**

```
# Write unit tests (function level)
# Write integration tests (component interaction)
# Write e2e tests (full workflow)
# OR get explicit authorization: "I AUTHORIZE YOU TO SKIP WRITING TESTS THIS TIME"
```

### ❌ Don't Test Implementation Details

**Wrong:**

```python
# ❌ BAD - Tests internal implementation
def test_calculate_total():
    calc = Calculator()
    assert calc._internal_cache == {}
    calc.add(5)
    assert calc._internal_cache['last'] == 5
```

**Correct:**

```python
# ✅ GOOD - Tests public behavior
def test_calculate_total():
    calc = Calculator()
    result = calc.add(5)
    assert result == 5

    result = calc.add(3)
    assert result == 8
```

## Best Practices

### 1. Test Naming Conventions

```python
# ✅ Descriptive test names
def test_user_login_with_valid_credentials_returns_token()
def test_user_login_with_invalid_password_returns_401()
def test_user_login_with_nonexistent_user_returns_404()

# ❌ Vague test names
def test_login()
def test_login_2()
def test_login_edge_case()
```

### 2. Arrange-Act-Assert Pattern

```python
def test_calculate_order_total_with_discount():
    # Arrange - Set up test data
    order = Order(items=[
        Item(price=100, quantity=2),
        Item(price=50, quantity=1)
    ])
    discount = Discount(percent=10)

    # Act - Perform action being tested
    total = calculate_total(order, discount)

    # Assert - Verify expected outcome
    assert total == 225  # (200 + 50) * 0.9
```

### 3. Test Isolation

```python
# ✅ Each test is independent
@pytest.fixture(autouse=True)
def reset_database():
    db.clear()
    yield
    db.clear()

def test_create_user():
    user = create_user("alice")
    assert db.count() == 1

def test_delete_user():
    user = create_user("bob")
    delete_user(user.id)
    assert db.count() == 0
# Tests can run in any order
```

### 4. Fast Feedback Loop

```bash
# Run specific test during development
pytest tests/test_feature.py::test_specific_case -v

# Run with watch mode (pytest-watch)
ptw tests/ -- -v

# Run in parallel (pytest-xdist)
pytest -n auto
```

### 5. Meaningful Assertions

```python
# ❌ Vague assertion
def test_api_response():
    response = api.get('/users')
    assert response

# ✅ Specific assertion
def test_api_response():
    response = api.get('/users')
    assert response.status_code == 200
    assert 'users' in response.json
    assert len(response.json['users']) > 0
    assert response.json['users'][0]['id'] is not None
```

## Summary

**Verification is mandatory before completion:**

1. **Write tests first** (TDD: RED → GREEN → REFACTOR)
2. **All test types required** (unit, integration, e2e)
3. **Test output must be pristine** (no warnings, no errors)
4. **Pre-commit hooks must pass** (NEVER use --no-verify)
5. **Manual verification in real environment**
6. **Coverage must meet standards** (typically ≥90%)

**Only declare work complete after ALL verification steps pass.**

## See Also

- [15-GIT-WORKFLOWS.md](./15-GIT-WORKFLOWS.md) - Pre-commit hook integration
- [16-ERROR-HANDLING.md](./16-ERROR-HANDLING.md) - Test failure debugging
- [19-BEST-PRACTICES.md](./19-BEST-PRACTICES.md) - Testing best practices
- [04-BASH-AND-COMMAND-EXECUTION.md](./04-BASH-AND-COMMAND-EXECUTION.md) - Running test commands
