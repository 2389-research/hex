# File Editing Strategies

## Overview

This document explains the **strategies and techniques** that make Claude Code highly effective at file editing. While 02-FILE-OPERATIONS.md explains the Edit tool's mechanics and constraints, this document focuses on the **patterns, decision-making, and advanced techniques** for successful file modifications.

**Key insight**: The Edit tool's constraints (exact string matching, uniqueness requirement, read-before-edit) force disciplined practices that actually make editing more reliable.

## The Core Strategy: Surgical Precision

### Philosophy

Claude Code uses **surgical precision** rather than wholesale rewrites:

```
Traditional approach:
1. Read entire file
2. Modify entire content in memory
3. Write entire file back

Claude Code approach:
1. Read file once
2. Identify exact change needed
3. Replace only what changes (Edit tool)
4. Preserve everything else automatically
```

**Benefits**:
- Changes are explicit and reviewable
- No risk of accidentally modifying unchanged parts
- Clear before/after comparison
- Version control diffs show only actual changes

### The Read-Match-Edit Pattern

**Every edit follows this pattern**:

```
1. READ: Load file contents
2. IDENTIFY: Locate target text
3. EXTRACT: Copy exact text (with context)
4. CRAFT: Create replacement
5. VERIFY: Check old_string will be unique
6. EDIT: Apply change
```

**Example in practice**:

```python
# 1. READ
Read("src/auth.py")
# Output shows:
#     15  def login(username, password):
#     16      user = find_user(username)
#     17      if check_password(user, password):
#     18          return create_session(user)
#     19      return None

# 2-3. IDENTIFY & EXTRACT (lines 15-19)
old_string = """def login(username, password):
    user = find_user(username)
    if check_password(user, password):
        return create_session(user)
    return None"""

# 4. CRAFT replacement
new_string = """def login(username, password):
    user = find_user(username)
    if not user:
        return None
    if check_password(user, password):
        return create_session(user)
    return None"""

# 5. VERIFY - old_string should appear exactly once
# (Function name makes it unique)

# 6. EDIT
Edit(
    file_path="src/auth.py",
    old_string=old_string,
    new_string=new_string
)
```

## Strategy 1: Context Calibration

### The Goldilocks Principle

Include **just enough context** to make the match unique:

```python
# File contains:
def foo():
    return True

def bar():
    return True

def baz():
    return True
```

**Too little context** (not unique):
```python
❌ old_string = "return True"  # Appears 3 times!
```

**Too much context** (fragile):
```python
❌ old_string = """def foo():
    return True

def bar():
    return True"""  # Breaks if bar() changes
```

**Just right**:
```python
✅ old_string = """def foo():
    return True"""  # Unique, minimal
```

### Context Selection Heuristics

**Start narrow, expand if needed**:

```python
# Level 1: Just the line
old_string = "x = calculate(data)"

# Level 2: Include function signature
old_string = """def process():
    x = calculate(data)"""

# Level 3: Include surrounding statements
old_string = """def process():
    data = load()
    x = calculate(data)
    save(x)"""

# Level 4: Include entire block
old_string = """class Processor:
    def process(self):
        data = load()
        x = calculate(data)
        save(x)"""
```

**Rule of thumb**: Use the smallest context that guarantees uniqueness.

### Natural Boundaries

**Prefer natural boundaries** for context:

✅ **Good boundaries**:
- Function definitions
- Class definitions
- Block statements (if/for/while)
- Comment sections
- Blank lines

❌ **Avoid breaking in middle of**:
- Statements
- String literals
- Comment blocks
- Expression chains

**Example**:

```python
# File contains:
def calculate_total(items):
    """Calculate total with tax."""
    subtotal = sum(item.price for item in items)
    tax = subtotal * 0.08
    return subtotal + tax

# ✅ GOOD (natural boundary - entire function):
old_string = """def calculate_total(items):
    \"\"\"Calculate total with tax.\"\"\"
    subtotal = sum(item.price for item in items)
    tax = subtotal * 0.08
    return subtotal + tax"""

# ❌ BAD (breaks in middle of logic):
old_string = """subtotal = sum(item.price for item in items)
    tax = subtotal"""
```

## Strategy 2: Whitespace Discipline

### The Copy-Paste Rule

**Never type old_string manually. Always copy from Read output.**

```python
# Read output:
     5	def foo():
     6	    print("test")
     ^^^^^ ← IGNORE THIS
          ^^^^^^^^^^^^^^^^^ ← COPY THIS (after tab)

# ✅ CORRECT (copied):
old_string = 'def foo():\n    print("test")'

# ❌ WRONG (typed manually):
old_string = "def foo():\n\tprint('test')"  # Tabs vs spaces mismatch!
```

### Tab vs Space Detection

**When in doubt, reveal whitespace**:

```python
# Use repr() to see actual characters:
>>> repr('    print("test")')  # 4 spaces
"    print('test')"

>>> repr('\tprint("test")')  # tab
"\tprint('test')"
```

**Or check Read output carefully**:
```
     6→    print("test")    ← Spaces (indent continues)
     6→	print("test")        ← Tab (immediate jump)
```

### Preserving Mixed Whitespace

Some files mix tabs and spaces (unfortunately common):

```python
# File has:
def foo():
\t    if True:  # Tab then spaces
\t\treturn "ok"  # Two tabs

# Must match exactly:
old_string = 'def foo():\n\t    if True:\n\t\treturn "ok"'
#                          ^tab ^spaces    ^tab ^tab
```

**Strategy**: Copy-paste guarantees exact preservation.

## Strategy 3: Multi-Line Editing

### The Heredoc Pattern

For multi-line strings, use heredoc-style formatting for clarity:

```python
Edit(
    file_path="src/app.py",
    old_string="""
def calculate(x, y):
    result = x + y
    return result
""".strip(),
    new_string="""
def calculate(x, y):
    result = x * y
    return result
""".strip()
)
```

**Note**: `.strip()` removes leading/trailing newlines from the literal but preserves internal formatting.

### Line-by-Line Construction

For complex edits, build strings line by line:

```python
old_lines = [
    "def process(data):",
    "    cleaned = clean(data)",
    "    validated = validate(cleaned)",
    "    return validated"
]
old_string = '\n'.join(old_lines)

new_lines = [
    "def process(data):",
    "    cleaned = clean(data)",
    "    if not cleaned:",
    "        raise ValueError('Invalid data')",
    "    validated = validate(cleaned)",
    "    return validated"
]
new_string = '\n'.join(new_lines)

Edit(file_path="processor.py", old_string=old_string, new_string=new_string)
```

**Benefits**:
- Clear diff visible in code
- Easy to verify each line
- Prevents string escaping errors

## Strategy 4: Incremental Edits

### One Change at a Time

**Break complex changes into atomic edits**:

```python
# Want to: 1) Add import, 2) Add parameter, 3) Add logic

# ❌ WRONG (all at once):
Edit(
    old_string="""import os

def process(data):
    return clean(data)""",
    new_string="""import os
import logging

def process(data, strict=False):
    logging.info('Processing')
    result = clean(data)
    if strict:
        validate(result)
    return result"""
)

# ✅ CORRECT (three separate edits):

# Edit 1: Add import
Edit(
    old_string="import os",
    new_string="import os\nimport logging"
)

# Edit 2: Add parameter
Edit(
    old_string="def process(data):",
    new_string="def process(data, strict=False):"
)

# Edit 3: Add logic
Edit(
    old_string="""def process(data, strict=False):
    return clean(data)""",
    new_string="""def process(data, strict=False):
    logging.info('Processing')
    result = clean(data)
    if strict:
        validate(result)
    return result"""
)
```

**Benefits**:
- Each edit is simple and clear
- Easier to verify correctness
- If one fails, others may still succeed
- Git history shows logical progression

### Sequential Dependencies

**When edits depend on each other, sequence them**:

```python
# Step 1: Rename function
Edit(
    old_string="def old_name():",
    new_string="def new_name():"
)

# Step 2: Update call sites (depends on step 1)
Edit(
    old_string="result = old_name()",
    new_string="result = new_name()",
    replace_all=True
)
```

**Rule**: If new_string of Edit A appears in old_string of Edit B, they must be sequential.

## Strategy 5: Replace All Tactics

### When to Use replace_all

**Use replace_all for**:
- Variable/function renaming throughout file
- Updating string literals globally
- Changing import names
- Fixing repeated typos

**Example - Safe rename**:

```python
# Rename variable throughout file
Edit(
    file_path="calculator.py",
    old_string="user_input",
    new_string="raw_input",
    replace_all=True
)
```

### Replace All Safety Checks

**Before using replace_all**:

1. **Search first** to see all matches:
```python
Grep(pattern="user_input", path="calculator.py", output_mode="content")
# Review all matches to ensure all should change
```

2. **Check for partial matches**:
```python
# ❌ DANGEROUS:
old_string = "data"  # Matches: data, update, metadata, data_clean
replace_all = True

# ✅ SAFER:
old_string = "data"  # Still risky
# Better to do manually or use word boundaries
```

3. **Consider word boundaries**:
```python
# Instead of replace_all for short strings:
# Manually edit each unique context

Edit(old_string="input_data = data", new_string="input_data = raw_data")
Edit(old_string="process(data)", new_string="process(raw_data)")
Edit(old_string="return data", new_string="return raw_data")
```

### The Verify-Replace Pattern

```python
# 1. Verify with Grep
Grep(pattern="old_var", path="module.py", output_mode="content", -n=True)
# Shows: 5 matches on lines 10, 23, 45, 67, 89

# 2. Review each line in context
Read("module.py", offset=8, limit=5)   # Around line 10
Read("module.py", offset=21, limit=5)  # Around line 23
# ... verify all are correct to change

# 3. Replace all
Edit(
    file_path="module.py",
    old_string="old_var",
    new_string="new_var",
    replace_all=True
)
```

## Strategy 6: Error Recovery

### Error: "old_string not found"

**Diagnosis workflow**:

```python
# 1. Re-read the file
Read("problem.py")

# 2. Copy exact text from output
# (Don't type manually)

# 3. Check for invisible characters
# Use repr() to inspect:
old_string = repr(text_from_read_output)
# Look for \t vs spaces, \r\n vs \n

# 4. Verify file wasn't changed
# (Maybe another process modified it)
Bash("git diff problem.py")

# 5. Try smaller context
# Maybe the surrounding context changed
old_string = "just_the_specific_line"
```

### Error: "old_string not unique"

**Resolution strategies**:

```python
# Strategy 1: Add context above
old_string = """# Previous line
target_line
# Next line"""

# Strategy 2: Include function/class
old_string = """def containing_function():
    target_line"""

# Strategy 3: If truly identical, use replace_all
old_string = "identical_line"
replace_all = True

# Strategy 4: Edit each occurrence separately with different context
Edit(old_string="def foo():\n    target", new_string="...")
Edit(old_string="def bar():\n    target", new_string="...")
```

### Error: "File must be read before editing"

**This means Read wasn't called in the current conversation**:

```python
# ❌ Session resumed after restart - previous Read doesn't count
Edit("file.py", ...)  # Error!

# ✅ Always Read in current session
Read("file.py")
Edit("file.py", ...)  # Success
```

## Strategy 7: Large File Handling

### Targeted Reading with Offset

**Don't read the entire file if you know where to edit**:

```python
# File has 10,000 lines
# Want to edit lines 5000-5010

# ❌ INEFFICIENT:
Read("huge.py")  # Reads lines 1-2000, misses target!

# ✅ EFFICIENT:
# Step 1: Find location
Grep(pattern="def target_function", path="huge.py", output_mode="content", -n=True)
# Shows: Line 5005

# Step 2: Read that section
Read("huge.py", offset=5000, limit=50)

# Step 3: Edit
Edit("huge.py", old_string="...", new_string="...")
```

### Grep-Then-Read Pattern

**Standard workflow for large codebases**:

```python
# 1. Search across files
Grep(pattern="class UserAuth", output_mode="files_with_matches")
# Result: src/auth/user.py

# 2. Find exact location
Grep(pattern="class UserAuth", path="src/auth/user.py", output_mode="content", -n=True)
# Result: Line 45

# 3. Read precise section
Read("src/auth/user.py", offset=40, limit=30)

# 4. Edit
Edit("src/auth/user.py", old_string="...", new_string="...")
```

## Strategy 8: Parallel Editing

### Independent File Edits

**When edits don't depend on each other, do them in parallel**:

```python
# All in single response:
Edit("src/auth.py", old_string="...", new_string="...")
Edit("src/models.py", old_string="...", new_string="...")
Edit("src/views.py", old_string="...", new_string="...")

# All three execute simultaneously
# 3x faster than sequential
```

### Multiple Edits Same File

**Can edit same file multiple times in one response if independent**:

```python
# Add import and fix function in one response:
Read("processor.py")

Edit("processor.py",
     old_string="import os",
     new_string="import os\nimport logging")

Edit("processor.py",
     old_string="def process():\n    return data",
     new_string="def process():\n    logging.info('start')\n    return data")

# Both edits execute (order guaranteed)
```

**Important**: Second edit operates on result of first edit.

## Strategy 9: Refactoring Patterns

### Extract Function

```python
# Before:
Read("app.py")
# Shows:
#     10  def main():
#     11      data = load_data()
#     12      cleaned = [x.strip() for x in data]
#     13      validated = [x for x in cleaned if len(x) > 0]
#     14      return validated

# Step 1: Create new function at end of file
Edit(
    file_path="app.py",
    old_string="def main():",  # Assuming unique
    new_string="""def clean_and_validate(data):
    cleaned = [x.strip() for x in data]
    validated = [x for x in cleaned if len(x) > 0]
    return validated

def main():"""
)

# Step 2: Replace inline code with call
Edit(
    file_path="app.py",
    old_string="""def main():
    data = load_data()
    cleaned = [x.strip() for x in data]
    validated = [x for x in cleaned if len(x) > 0]
    return validated""",
    new_string="""def main():
    data = load_data()
    return clean_and_validate(data)"""
)
```

### Inline Function

```python
# Opposite of extract - replace call with implementation

# Before:
#     def calculate():
#         return helper()
#
#     def helper():
#         return 42

# Step 1: Inline the call
Edit(
    old_string="def calculate():\n    return helper()",
    new_string="def calculate():\n    return 42"
)

# Step 2: Remove helper (if no other uses)
Edit(
    old_string="""def helper():
    return 42

""",
    new_string=""
)
```

### Rename Symbol

```python
# Rename function/variable/class throughout file

# Step 1: Rename definition
Edit(
    old_string="def old_name(x):",
    new_string="def new_name(x):"
)

# Step 2: Rename all calls
Edit(
    old_string="old_name",
    new_string="new_name",
    replace_all=True
)
```

## Strategy 10: Test-Driven Editing

### Edit-Test-Verify Loop

```python
# 1. Make edit
Edit("calculator.py", old_string="...", new_string="...")

# 2. Run tests immediately
Bash("pytest tests/test_calculator.py -v")

# 3. If tests fail, investigate
Read("calculator.py")  # Check what was actually written

# 4. Fix if needed
Edit("calculator.py", old_string="...", new_string="...")

# 5. Re-test
Bash("pytest tests/test_calculator.py -v")
```

**Rule**: Never make multiple edits without testing intermediate states.

### Pre-Edit Validation

```python
# Before making risky edit:

# 1. Ensure tests pass currently
Bash("pytest")

# 2. Make edit
Edit(...)

# 3. Run tests again
Bash("pytest")

# 4. If tests fail, you know the edit caused it
```

## Advanced Techniques

### Technique 1: Regex in Search, Literal in Edit

```python
# Use Grep with regex to find, but Edit with literal strings

# Find all logger calls
Grep(pattern="logger\.(debug|info|warn|error)", path="app.py")

# Edit specific instance with literal match
Edit(
    old_string='logger.info("Starting process")',
    new_string='logger.debug("Starting process")'
)
```

### Technique 2: Two-Phase Refactor

**For complex refactors, use intermediate state**:

```python
# Want to change signature: foo(x) → foo(x, y=None)
# But called in 50 places

# Phase 1: Make parameter optional (backward compatible)
Edit(
    old_string="def foo(x):",
    new_string="def foo(x, y=None):"
)

# Test: Existing calls still work
Bash("pytest")

# Phase 2: Update call sites one by one
Edit(old_string="result = foo(data)", new_string="result = foo(data, mode='strict')")
# ... repeat for each call site

# Phase 3: Make parameter required (after all updated)
Edit(
    old_string="def foo(x, y=None):",
    new_string="def foo(x, y):"
)
```

### Technique 3: Comment Anchors

**Add temporary comments to create unique contexts**:

```python
# Problem: Multiple identical blocks

def process_a():
    return calculate()

def process_b():
    return calculate()  # Identical!

# Solution: Add temporary anchors
Edit(
    old_string="def process_a():",
    new_string="# ANCHOR_A\ndef process_a():"
)

Edit(
    old_string="def process_b():",
    new_string="# ANCHOR_B\ndef process_b():"
)

# Now can edit each uniquely
Edit(
    old_string="# ANCHOR_A\ndef process_a():\n    return calculate()",
    new_string="def process_a():\n    return calculate_fast()"
)

Edit(
    old_string="# ANCHOR_B\ndef process_b():\n    return calculate()",
    new_string="def process_b():\n    return calculate_safe()"
)
```

### Technique 4: Batch Rename with Safety

```python
# Rename variable but avoid false matches

# Step 1: Find all occurrences
Grep(pattern=r"\buser\b", path="app.py", output_mode="content", -n=True)
# Use \b for word boundaries in search

# Step 2: Verify each match
Read("app.py", offset=<line-10>, limit=20) for each match

# Step 3: Rename with context if any false positives
# If all safe:
Edit(old_string="user", new_string="account", replace_all=True)

# If some false positives, edit individually:
Edit(old_string="def login(user):", new_string="def login(account):")
Edit(old_string="user.save()", new_string="account.save()")
# Skip "username" (contains "user" but shouldn't change)
```

## Common Patterns Library

### Pattern: Add Import

```python
Edit(
    old_string="import os",
    new_string="import os\nimport logging"
)
```

### Pattern: Add Parameter

```python
Edit(
    old_string="def process(data):",
    new_string="def process(data, strict=False):"
)
```

### Pattern: Add Error Handling

```python
Edit(
    old_string="""result = risky_operation()
return result""",
    new_string="""try:
    result = risky_operation()
except Exception as e:
    logging.error(f"Operation failed: {e}")
    raise
return result"""
)
```

### Pattern: Add Logging

```python
Edit(
    old_string="""def important_function():
    result = calculate()""",
    new_string="""def important_function():
    logging.info("Starting important_function")
    result = calculate()
    logging.debug(f"Result: {result}")"""
)
```

### Pattern: Add Type Hints

```python
Edit(
    old_string="def process(data):",
    new_string="def process(data: dict) -> dict:"
)
```

### Pattern: Extract Constant

```python
Edit(
    old_string='if timeout > 30:',
    new_string='MAX_TIMEOUT = 30\n\nif timeout > MAX_TIMEOUT:'
)
```

### Pattern: Add Docstring

```python
Edit(
    old_string='def calculate(x, y):',
    new_string='def calculate(x, y):\n    """Calculate sum of x and y."""'
)
```

## Decision Framework

**When to use Edit vs Write:**

| Scenario | Tool | Why |
|----------|------|-----|
| Modify 1-50 lines in large file | Edit | Surgical, clear changes |
| Modify >50% of file | Edit (probably) | Still clearer than rewrite |
| Complete restructure | Write | When Edit would be more complex |
| Create new file | Write | No existing content |
| Generated code (template) | Write | Creating from scratch |
| Add/remove/modify function | Edit | Standard modification |

**Rule**: Default to Edit. Only use Write when Edit is truly more complex.

## Anti-Patterns to Avoid

### ❌ Manual String Construction

```python
# DON'T:
old_string = "def foo():\n    return bar"  # Typed manually

# DO:
# Copy from Read output preserving exact whitespace
```

### ❌ Overly Large Edits

```python
# DON'T:
Edit(old_string="<entire file>", new_string="<entire file with changes>")

# DO:
Edit(old_string="<specific function>", new_string="<modified function>")
```

### ❌ Editing Without Reading

```python
# DON'T:
Edit("file.py", ...)  # Haven't read it!

# DO:
Read("file.py")
Edit("file.py", ...)
```

### ❌ Assuming Uniqueness

```python
# DON'T:
Edit(old_string="return True")  # Might appear many times!

# DO:
Edit(old_string="def is_valid():\n    return True")  # Unique context
```

### ❌ Ignoring Test Failures

```python
# DON'T:
Edit(...)
Bash("pytest")  # Fails
Edit(...)  # Keep going anyway!

# DO:
Edit(...)
Bash("pytest")  # Fails
Read("file.py")  # Investigate
# Fix the issue before continuing
```

## Summary

**The Claude Code editing philosophy**:

1. **Surgical over wholesale** - Change only what needs changing
2. **Explicit over implicit** - Clear before/after, no hidden changes
3. **Verified over assumed** - Test after edits, verify success
4. **Incremental over monolithic** - Small atomic changes
5. **Context-aware** - Include just enough for uniqueness
6. **Whitespace-perfect** - Copy-paste, never type
7. **Test-driven** - Edit, test, verify loop

**Success metrics**:
- ✅ Edit succeeds on first try
- ✅ Tests pass after edit
- ✅ Git diff shows only intended changes
- ✅ No unintended modifications
- ✅ Clear why the change was made

**Key techniques**:
1. Read-Match-Edit pattern
2. Context calibration (Goldilocks principle)
3. Copy-paste discipline for whitespace
4. Incremental atomic edits
5. Grep-then-read for large files
6. Parallel edits for independence
7. Test after every change
8. Natural boundaries for context

---

**See Also:**
- 02-FILE-OPERATIONS.md - Edit tool mechanics and constraints
- 19-BEST-PRACTICES.md - General coding best practices
- 12-VERIFICATION-AND-TESTING.md - Testing workflows
