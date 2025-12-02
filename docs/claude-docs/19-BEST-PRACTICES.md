# Best Practices

Learned patterns, proven workflows, efficiency tips, and anti-patterns to avoid when working with Claude Code.

## Core Development Workflows

### The TDD Cycle

**Standard Flow:**
```
1. Write failing test (RED)
   ✅ Test fails as expected
   ✅ Error message is clear

2. Write minimal implementation (GREEN)
   ✅ Test passes
   ✅ No extra features

3. Refactor (REFACTOR)
   ✅ Improve code quality
   ✅ Tests still pass

4. Commit
   ✅ All tests passing
   ✅ Meaningful commit message
```

**Example:**

```python
# Step 1: RED - Write failing test
def test_user_registration():
    user = create_user("alice", "alice@example.com")
    assert user.username == "alice"
    assert user.email == "alice@example.com"
    assert user.id is not None

# Run: pytest -x
# ❌ FAIL: NameError: name 'create_user' is not defined
# Good! Expected failure.

# Step 2: GREEN - Minimal implementation
def create_user(username: str, email: str) -> User:
    return User(
        id=generate_id(),
        username=username,
        email=email
    )

# Run: pytest
# ✅ PASS: test_user_registration

# Step 3: REFACTOR - Improve
def create_user(username: str, email: str) -> User:
    """Create a new user with validated email."""
    if not is_valid_email(email):
        raise ValueError(f"Invalid email: {email}")

    return User(
        id=generate_id(),
        username=username.strip(),
        email=email.lower()
    )

# Run: pytest
# ✅ PASS: All tests still pass

# Step 4: COMMIT
git add src/users.py tests/test_users.py
git commit -m "feat: add user registration with email validation"
```

### Feature Implementation Pattern

**Proven Sequence:**

```
1. Understand requirement
   → Ask clarifying questions if needed
   → Propose approach
   → Get user approval

2. Plan implementation
   → Break into small tasks
   → Create TodoWrite if complex
   → Identify files to modify

3. Write test first (TDD)
   → Test describes desired behavior
   → Run test (should fail)
   → Commit test

4. Implement feature
   → Write minimal code to pass test
   → Run test (should pass)
   → Commit implementation

5. Refactor
   → Improve code quality
   → Keep tests passing
   → Commit refactoring

6. Integration test
   → Test feature in full context
   → Verify nothing broken
   → Commit if changes needed

7. Documentation
   → Update README if needed
   → Add code comments
   → Update API docs
   → Commit docs

8. Final verification
   → All tests pass
   → Linting passes
   → Hooks pass
   → Ready for review/PR
```

### Bug Fix Pattern

**Systematic Approach:**

```
1. Reproduce bug
   → Write failing test that demonstrates bug
   → Confirm test fails
   → Commit test

2. Debug systematically
   → Use superpowers:systematic-debugging skill
   → Find root cause, not symptoms
   → Add debug logging if needed

3. Fix root cause
   → Implement fix
   → Test passes
   → Commit fix

4. Verify fix
   → Run full test suite
   → Check related functionality
   → Ensure no regressions

5. Clean up
   → Remove debug logging
   → Refactor if needed
   → Commit cleanup

6. Document
   → Add comment explaining why fix works
   → Update docs if bug revealed gap
   → Commit docs
```

**Example:**

```python
# 1. REPRODUCE - Write failing test
def test_login_with_null_email():
    """Regression test for #123: Crash on null email"""
    response = client.post("/login", json={
        "username": "alice",
        "email": None,  # This crashes
        "password": "secret"
    })
    assert response.status_code == 400
    assert "email" in response.json()["error"]

# Run test:
# ❌ FAIL: KeyError on user.email

# 2. DEBUG - Find root cause
# Problem: Code assumes email exists
# Location: api/routes/auth.py:42
# Issue: No validation before accessing email

# 3. FIX - Handle null email
def login(request):
    email = request.json.get("email")
    if email is None:
        return {"error": "Email is required"}, 400

    # ... rest of login logic

# Run test:
# ✅ PASS: test_login_with_null_email

# 4. VERIFY - Run full suite
# ✅ PASS: All 156 tests pass

# 5. COMMIT
git add api/routes/auth.py tests/test_auth.py
git commit -m "fix: handle null email in login (closes #123)"
```

## Code Quality Practices

### Writing Maintainable Code

**Clarity Over Cleverness:**
```python
# ❌ Clever but confusing
result = [x for x in data if x not in [y for y in filter(lambda z: z > 0, data)]]

# ✅ Clear and maintainable
positive_values = [x for x in data if x > 0]
result = [x for x in data if x not in positive_values]
```

**Meaningful Names:**
```python
# ❌ Unclear
def f(x, y):
    return x * y + 10

# ✅ Clear
def calculate_total_with_tax(subtotal: float, tax_rate: float) -> float:
    return subtotal * tax_rate + 10
```

**Small Functions:**
```python
# ❌ Too large
def process_order(order):
    # 100 lines of code
    # Validation
    # Calculation
    # Database update
    # Email sending
    # Logging
    pass

# ✅ Well-factored
def process_order(order):
    validate_order(order)
    total = calculate_order_total(order)
    save_order(order, total)
    send_confirmation_email(order)
    log_order_processed(order)
```

**Comments for Why, Not What:**
```python
# ❌ Obvious comment
# Increment counter
counter += 1

# ❌ Outdated comment
# This will be refactored later (from 2019)
legacy_code()

# ✅ Explains why
# Use exponential backoff to avoid overwhelming API
# See: https://api-docs.example.com/rate-limiting
time.sleep(2 ** attempt)

# ✅ Explains business logic
# Orders over $100 get free shipping per marketing campaign
if order.total > 100:
    order.shipping_cost = 0
```

### File Organization

**ABOUTME Comments:**
```python
# ABOUTME: User authentication endpoints
# ABOUTME: Handles login, logout, token refresh via JWT

from fastapi import APIRouter, HTTPException
# ... rest of file
```

**Logical Grouping:**
```
project/
  src/
    api/
      __init__.py
      routes/           # Route handlers
        auth.py
        users.py
        orders.py
      middleware/       # Request/response processing
        auth.py
        logging.py
      models/           # Data models
        user.py
        order.py
    db/
      __init__.py
      connection.py     # Database setup
      queries.py        # SQL queries
    utils/
      __init__.py
      email.py          # Email utilities
      validation.py     # Input validation
  tests/
    test_auth.py        # Mirror src structure
    test_users.py
    test_orders.py
```

### Dependency Management

**Using uv Effectively:**

```bash
# ✅ Add dependencies properly
uv add requests          # Production dependency
uv add --dev pytest      # Development dependency
uv add --dev ruff        # Dev tool

# ✅ Lock dependencies
uv lock                  # Create/update uv.lock

# ✅ Sync environment
uv sync --locked         # Install from lock file

# ✅ Run commands in env
uv run pytest           # Use project environment
uv run python -m app    # Run module

# ❌ Don't use old tools
pip install package     # NO - use uv
poetry add package      # NO - use uv
virtualenv venv         # NO - use uv venv
```

**Dependency Selection:**
```
Choose dependencies based on:
1. Maintenance: Active development, recent updates
2. Popularity: Well-tested, community support
3. License: Compatible with project
4. Size: Minimal transitive dependencies
5. Quality: Good documentation, tests

Example decision:
  Need: HTTP client
  Options:
    - requests: ✅ Popular, simple, well-maintained
    - httpx: ✅ Modern, async support
    - urllib: ⚠️  Stdlib but complex API
  Choose: requests for sync, httpx for async
```

## Git Practices

### Commit Hygiene

**Atomic Commits:**
```
Each commit should:
  ✅ Have single purpose
  ✅ Pass all tests
  ✅ Be potentially revertable
  ❌ Mix multiple features
  ❌ Leave code broken

Example:
  ❌ "Add auth and fix logging and update deps"
  ✅ "feat: add JWT authentication"
  ✅ "fix: handle null values in logger"
  ✅ "chore: update dependencies"
```

**Commit Messages:**
```
Format:
  <type>: <summary>

  [optional body]

  🤖 Generated with [Claude Code](https://claude.com/claude-code)

  Co-Authored-By: Claude <noreply@anthropic.com>

Types:
  feat: New feature
  fix: Bug fix
  refactor: Code restructuring
  test: Add/update tests
  docs: Documentation
  chore: Maintenance
  style: Formatting
  perf: Performance

Good messages:
  ✅ "feat: add rate limiting to API endpoints"
  ✅ "fix: handle null email in user registration (closes #123)"
  ✅ "refactor: extract email validation to separate function"

Bad messages:
  ❌ "updates"
  ❌ "fix stuff"
  ❌ "WIP"
  ❌ "asdf"
```

**Commit Frequency:**
```
Commit when:
  ✅ Feature is complete and tested
  ✅ Bug is fixed and verified
  ✅ Refactoring is done and tests pass
  ✅ Before switching tasks
  ✅ Before risky operations

Don't commit:
  ❌ Broken code
  ❌ Failing tests
  ❌ Half-finished features (unless explicit WIP)
  ❌ Debug code left in
```

### Branch Management

**When to Branch:**
```
Create branch for:
  ✅ Large features (multi-session)
  ✅ Experimental work
  ✅ Team collaboration
  ✅ Multiple features in parallel

Stay on main for:
  ✅ Small changes
  ✅ Personal projects
  ✅ Rapid iteration
```

**Branch Naming:**
```
Pattern: <type>/<description>

Examples:
  ✅ feature/user-authentication
  ✅ fix/null-email-handling
  ✅ refactor/extract-email-validation
  ✅ docs/api-documentation

Avoid:
  ❌ my-branch
  ❌ test
  ❌ temp
  ❌ asdf
```

**Branch Cleanup:**
```bash
# After merging
git branch -d feature/user-auth    # Safe delete (merged only)
git push origin --delete feature/user-auth  # Delete remote

# List merged branches
git branch --merged main

# List unmerged
git branch --no-merged main
```

## Testing Practices

### Test Coverage

**What to Test:**
```
✅ MUST test:
  - Core business logic
  - API endpoints
  - Data validation
  - Error handling
  - Edge cases

✅ SHOULD test:
  - User workflows (integration)
  - Database queries
  - External API interactions
  - Authentication/authorization

⚠️  MAYBE test:
  - Trivial getters/setters
  - Third-party library behavior
  - Framework internals

❌ DON'T test:
  - Private implementation details
  - Test the test framework
  - Mock behavior that hides integration issues
```

**Test Types:**

```
Unit Tests:
  - Test single function/class
  - Fast execution (<1ms)
  - No external dependencies
  Example: test_calculate_total()

Integration Tests:
  - Test multiple components together
  - Medium execution (10-100ms)
  - May use database/files
  Example: test_user_registration_flow()

End-to-End Tests:
  - Test full user workflows
  - Slow execution (1-10s)
  - Uses real system
  Example: test_complete_checkout_process()
```

### Test Quality

**Good Tests:**
```python
# ✅ Clear name
def test_login_fails_with_invalid_password():
    pass

# ✅ Arrange-Act-Assert
def test_user_creation():
    # Arrange
    username = "alice"
    email = "alice@example.com"

    # Act
    user = create_user(username, email)

    # Assert
    assert user.username == username
    assert user.email == email
    assert user.id is not None

# ✅ Test one thing
def test_email_validation_rejects_invalid_format():
    invalid_emails = ["notanemail", "@example.com", "user@"]
    for email in invalid_emails:
        with pytest.raises(ValueError):
            validate_email(email)

# ✅ Use real data (no mocks)
def test_user_can_login(db):  # Real database
    user = create_user("alice", "alice@example.com", "password123")
    result = login("alice", "password123")
    assert result.success is True
```

**Bad Tests:**
```python
# ❌ Unclear name
def test_1():
    pass

# ❌ Testing multiple things
def test_everything():
    test_creation()
    test_login()
    test_logout()
    test_deletion()

# ❌ Using mocks (forbidden)
def test_with_mock():
    mock_db = MagicMock()
    mock_db.get_user.return_value = fake_user
    # Testing mock behavior, not real code!

# ❌ No assertions
def test_user_creation():
    create_user("alice", "alice@example.com")
    # What are we testing?
```

## Performance Practices

### Efficient File Operations

**Read Strategically:**
```
# ❌ Inefficient
Read("huge_file.py")  # 5000 lines, 60k tokens

# ✅ Efficient
Grep(pattern="class UserAuth", path="huge_file.py")
# Find what you need
Read("huge_file.py", offset=100, limit=50)
# Read only that section
```

**Search Before Reading:**
```
Workflow:
1. Glob("**/*.py") to find files
2. Grep(pattern="def login") to find specific code
3. Read only the relevant files
4. Edit specific sections

Tokens saved: 80-90%
```

### Efficient Command Usage

**Limit Output:**
```bash
# ❌ Wasteful
git log                    # Entire history (huge)

# ✅ Efficient
git log --oneline -20      # Last 20 commits
git log --since="1 week"   # Recent only

# ❌ Wasteful
pytest -vv                 # Very verbose

# ✅ Efficient
pytest -q                  # Quiet mode
pytest -x                  # Stop at first failure
```

**Parallel Operations:**
```
# ✅ Run independent commands in parallel
Bash("git status")
Bash("git log --oneline -10")
Bash("uv run pytest --collect-only")

# All three execute at once
# Results come back together
# Faster than sequential
```

### Memory Management

**Streaming Large Data:**
```python
# ❌ Load entire file
with open("huge.csv") as f:
    data = f.read()  # 1GB in memory!
    process(data)

# ✅ Stream line by line
with open("huge.csv") as f:
    for line in f:
        process(line)  # Small memory footprint
```

**Batch Processing:**
```python
# ❌ Process all at once
results = [process(item) for item in million_items]

# ✅ Process in batches
batch_size = 1000
for i in range(0, len(items), batch_size):
    batch = items[i:i+batch_size]
    results = process_batch(batch)
    save_results(results)
```

## Team Collaboration

### Code Review

**Before Requesting Review:**
```
✅ Checklist:
  - All tests pass
  - Linting passes
  - Hooks pass
  - Code is refactored
  - Commits are clean
  - PR description is clear
  - Test plan is documented

Use superpowers:requesting-code-review skill
  → Automatic review by subagent
  → Catches issues before human review
  → Validates against requirements
```

**Writing PR Descriptions:**
```markdown
## Summary
- High-level change 1
- High-level change 2
- High-level change 3

## Changes
- `api/auth.py`: Added JWT token generation
- `api/routes/auth.py`: Added login/logout endpoints
- `api/middleware.py`: Added auth middleware
- `tests/test_auth.py`: Added 156 tests

## Test plan
- [ ] All unit tests pass (156/156)
- [ ] Integration tests pass
- [ ] Manually tested login flow
- [ ] Tested token expiration
- [ ] Tested invalid credentials

## Breaking changes
None / [describe breaking changes]

## Deployment notes
- Requires JWT_SECRET environment variable
- Database migration needed: `alembic upgrade head`

🤖 Generated with [Claude Code](https://claude.com/claude-code)
```

### Documentation

**When to Document:**
```
✅ Document:
  - Public APIs
  - Complex algorithms
  - Business logic
  - Setup instructions
  - Deployment process
  - Architecture decisions

❌ Don't document:
  - Obvious code
  - Temporary notes
  - Outdated information
```

**Documentation Types:**

```
README.md:
  - Project overview
  - Setup instructions
  - Quick start guide
  - Links to detailed docs

API.md:
  - Endpoint descriptions
  - Request/response formats
  - Authentication
  - Error codes

ARCHITECTURE.md:
  - System design
  - Component interactions
  - Technology choices
  - Deployment architecture

CHANGELOG.md:
  - Version history
  - Breaking changes
  - Migration guides
```

## Anti-Patterns to Avoid

### Code Anti-Patterns

**❌ God Objects:**
```python
# Everything in one class
class Application:
    def handle_request(self): pass
    def query_database(self): pass
    def send_email(self): pass
    def process_payment(self): pass
    def log_activity(self): pass
    # ... 50 more methods
```

**❌ Magic Numbers:**
```python
if user.age > 18:  # What's special about 18?
    pass

# ✅ Use constants
LEGAL_AGE = 18
if user.age > LEGAL_AGE:
    pass
```

**❌ Error Swallowing:**
```python
try:
    process_payment()
except Exception:
    pass  # What went wrong?!

# ✅ Handle errors
try:
    process_payment()
except PaymentError as e:
    log.error(f"Payment failed: {e}")
    notify_user("Payment failed")
    raise
```

**❌ Premature Optimization:**
```python
# Optimizing before profiling
def calculate(x):
    # 100 lines of complex optimization
    # Is this even a bottleneck?

# ✅ Start simple, optimize if needed
def calculate(x):
    return x * 2 + 10  # Simple, clear
    # Profile if slow, then optimize
```

### Process Anti-Patterns

**❌ Skipping Tests:**
```
"I'll add tests later"  → Tests never added
"This is simple, no test needed" → Bug appears
"Tests are slow to write" → Debugging is slower

✅ ALWAYS write tests (TDD)
```

**❌ Bypassing Hooks:**
```bash
# FORBIDDEN
git commit --no-verify
git push --no-verify

# ✅ Fix the issues instead
ruff check --fix .
pytest
git commit  # Hooks pass
```

**❌ Scope Creep:**
```
Task: "Add login endpoint"

❌ Also implements:
  - Password reset
  - Email verification
  - OAuth integration
  - Rate limiting

✅ Just implement login
  - Other features in separate PRs
  - Each feature tested independently
  - Easier to review
```

**❌ Unclear Commits:**
```
❌ "updates"
❌ "fix stuff"
❌ "WIP" (then never cleaned up)
❌ Mixing multiple unrelated changes

✅ Clear, atomic commits
✅ One purpose per commit
✅ Meaningful messages
```

### Communication Anti-Patterns

**❌ Assuming User Intent:**
```
User: "Make it better"
❌ Claude: *rewrites entire codebase*

✅ Claude: "What specifically should I improve?
           - Performance?
           - Readability?
           - Test coverage?"
```

**❌ Not Asking for Help:**
```
*Tries same failed approach 5 times*
❌ Keeps trying without asking

✅ After 2 attempts, ask user:
   "I've tried X and Y, both failed because Z.
    Can you help me understand what's wrong?"
```

**❌ Hiding Errors:**
```
Error occurs
❌ Silently continue, hoping it doesn't matter

✅ Report error:
   "I encountered an error: [details]
    Here's what I tried: [attempts]
    How should I proceed?"
```

## Efficiency Tips

### Keyboard Shortcuts for Users

```
# In terminal
Ctrl+R          # Search command history
Ctrl+A          # Jump to start of line
Ctrl+E          # Jump to end of line
Ctrl+U          # Clear line
Ctrl+C          # Cancel current command

# In git
git config --global alias.st status
git config --global alias.co checkout
git config --global alias.br branch
git config --global alias.cm 'commit -m'
git config --global alias.last 'log -1 HEAD'
```

### Project Templates

**Quick Start Template:**
```bash
# New Python project
uv init my-project
cd my-project
uv add fastapi uvicorn
uv add --dev pytest ruff mypy
uv add --dev pre-commit

# Setup pre-commit
cat > .pre-commit-config.yaml <<EOF
repos:
  - repo: https://github.com/astral-sh/ruff-pre-commit
    rev: v0.1.0
    hooks:
      - id: ruff
      - id: ruff-format
  - repo: https://github.com/pre-commit/mirrors-mypy
    rev: v1.5.1
    hooks:
      - id: mypy
EOF

pre-commit install

# Create structure
mkdir -p src/api tests
touch src/api/__init__.py
touch tests/test_api.py

# Ready to code!
```

### Debugging Checklist

```
When bug occurs:
1. ✅ Can you reproduce it?
   - Write test that reproduces bug
   - If can't reproduce, gather more info

2. ✅ What changed?
   - git log --oneline -10
   - git diff HEAD~1
   - Recent changes often cause bugs

3. ✅ What's the error message?
   - Read it carefully
   - Google if unclear
   - Check documentation

4. ✅ Where does it fail?
   - Stack trace shows location
   - Add logging/print statements
   - Use debugger if needed

5. ✅ What did you expect?
   - vs what actually happened
   - Mismatch reveals the issue

6. ✅ What's the minimal reproduction?
   - Remove unrelated code
   - Isolate the issue
   - Easier to fix

7. ✅ What's the root cause?
   - Not just symptoms
   - Why did this happen?
   - How to prevent in future?
```

## Summary

### Golden Rules

1. **TDD Always** - Write test first, implement second
2. **Commit Often** - Small, atomic, tested commits
3. **Never Bypass Hooks** - Fix issues, don't skip checks
4. **Ask When Unsure** - Clarify before implementing
5. **Read Before Write** - Understand before modifying
6. **Test Real, Not Mocks** - Use real dependencies
7. **Quality Over Speed** - Do it right, not fast
8. **Document Why** - Explain reasoning, not what
9. **Review Before Merge** - Use code review skill
10. **Learn From Errors** - Each error teaches something

### Workflow Checklist

For every feature:
- [ ] Understand requirement
- [ ] Write failing test
- [ ] Implement minimal solution
- [ ] Test passes
- [ ] Refactor code
- [ ] All tests pass
- [ ] Linting passes
- [ ] Hooks pass
- [ ] Commit with good message
- [ ] Documentation updated
- [ ] Ready for review

These practices ensure high-quality, maintainable code that the team can work with confidently.
