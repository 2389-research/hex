# File Operations in Claude Code

This document explains how Claude Code reads, edits, and writes files. These are the most fundamental and frequently used operations.

## The Three Core File Tools

```
┌─────────────────────────────────────────────────┐
│              File Operations                     │
│                                                  │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐      │
│  │   READ   │  │   EDIT   │  │  WRITE   │      │
│  │          │  │          │  │          │      │
│  │ View     │  │ Modify   │  │ Create   │      │
│  │ files    │  │ existing │  │ new or   │      │
│  │          │  │ files    │  │ rewrite  │      │
│  └──────────┘  └──────────┘  └──────────┘      │
│                                                  │
│  Must read     Exact string  Overwrites         │
│  before edit   matching      entire file        │
└─────────────────────────────────────────────────┘
```

## Read Tool

### Purpose

The Read tool retrieves file contents so Claude Code can understand current state before making changes.

### Critical Rule: Read Before Edit

**You MUST read a file before editing it.** This is enforced:

```
❌ WRONG:
Edit: myfile.py (without reading first)
→ Error: "File must be read before editing"

✓ CORRECT:
1. Read: myfile.py
2. Edit: myfile.py
```

### Read Tool Signature

```python
Read(
    file_path: str,     # Absolute path to file (required)
    offset: int = 0,    # Line number to start at (optional)
    limit: int = 2000   # Max lines to read (optional)
)
```

### Output Format

Read returns file contents with **line number prefixes**:

```
Format: <spaces><line_number><tab><content>

Example output:
     1	def calculate_total(items):
     2	    """Calculate total price of items."""
     3	    total = 0
     4	    for item in items:
     5	        total += item.price
     6	    return total
```

**Critical**: The prefix format is `<spaces><line_number><tab>`. Everything after the tab is the actual file content.

### Line Number Prefix Details

The prefix structure:
```
"     1\tdef calculate_total(items):\n"
 ^^^^^   ^^^^^^^^^^^^^^^^^^^^^^^^^^^
 spaces  actual content starts here
 and     (after the tab)
 number
 and
 tab
```

Components:
1. **Spaces**: Right-align line numbers (e.g., "     1", "    42", "   123")
2. **Line number**: The actual line number (1-indexed)
3. **Tab character**: Single `\t` separator
4. **Content**: The actual file content with original whitespace

### Why Line Numbers Matter

Line numbers help you:
1. **Navigate**: "The error is on line 42"
2. **Verify**: "I edited the function starting at line 15"
3. **Debug**: "Lines 20-25 are the problematic section"

But for editing, **ignore the prefix entirely**.

### Reading Large Files

Files with >2000 lines are truncated by default:

```python
# Read entire file (default: first 2000 lines)
Read("large_file.py")

# Read starting at line 1000
Read("large_file.py", offset=1000)

# Read 500 lines starting at line 1000
Read("large_file.py", offset=1000, limit=500)
```

**Best practice**: Read only what you need. If searching for a specific function, use Grep first to find its location, then Read with offset.

### Reading Binary Files

Read can handle images, PDFs, and other binary formats:

```python
# Images: Displayed visually to Claude
Read("diagram.png")

# PDFs: Text and visual content extracted
Read("documentation.pdf")

# Jupyter notebooks: All cells with outputs
Read("analysis.ipynb")
```

### Reading Directories

**You cannot Read a directory**. Use Bash with `ls` instead:

```
❌ WRONG:
Read("src/")
→ Error

✓ CORRECT:
Bash("ls -la src/")
```

### Read Examples

#### Example 1: Simple Read

```python
# Request
Read("/Users/harper/project/src/main.py")

# Response
     1	#!/usr/bin/env python3
     2	"""Main entry point for the application."""
     3
     4	import sys
     5	from app.cli import CLI
     6
     7	def main():
     8	    cli = CLI()
     9	    sys.exit(cli.run())
    10
    11	if __name__ == "__main__":
    12	    main()
```

#### Example 2: Read with Offset

```python
# File has 5000 lines, we want lines 2000-2500
Read("/Users/harper/project/data.py", offset=2000, limit=500)

# Response starts at line 2000
  2000	def process_batch(batch_id):
  2001	    """Process a batch of records."""
  2002	    # ... (500 lines)
  2500	    return results
```

#### Example 3: Reading Multiple Files

```python
# Read files in parallel (independent operations)
Read("/Users/harper/project/src/auth.py")
Read("/Users/harper/project/src/models.py")
Read("/Users/harper/project/tests/test_auth.py")

# All three files returned simultaneously
```

## Edit Tool

The Edit tool modifies existing files using **exact string replacement**. This is the most complex and error-prone tool.

### Edit Tool Signature

```python
Edit(
    file_path: str,      # Absolute path to file (required)
    old_string: str,     # Exact text to find (required)
    new_string: str,     # Replacement text (required)
    replace_all: bool = False  # Replace all occurrences (optional)
)
```

### Critical Rules

#### Rule 1: Must Read First

```
❌ Cannot edit a file you haven't read in this conversation
✓ Must Read, then Edit
```

#### Rule 2: Exact String Matching

The `old_string` must match **exactly**:

```python
# File contains:
def hello():
    print("world")

# ❌ WRONG (extra space):
Edit(
    file_path="test.py",
    old_string="def  hello():",  # Two spaces
    new_string="def hello():"
)
→ Error: "old_string not found"

# ✓ CORRECT (exact match):
Edit(
    file_path="test.py",
    old_string="def hello():",  # One space
    new_string="def greet():"
)
```

#### Rule 3: Ignore Line Number Prefix

When reading file output, **never include the line number prefix** in old_string:

```python
# Read output shows:
     1	def calculate(x, y):
     2	    return x + y

# ❌ WRONG (includes prefix):
Edit(
    old_string="     1\tdef calculate(x, y):\n     2\t    return x + y"
)

# ✓ CORRECT (only actual content):
Edit(
    old_string="def calculate(x, y):\n    return x + y"
)
```

**Remember**: The prefix format is `<spaces><line_number><tab>`. Everything **after the tab** is what you match against.

#### Rule 4: Uniqueness Requirement

The `old_string` must appear **exactly once** in the file (unless using `replace_all=True`):

```python
# File contains:
def foo():
    print("test")

def bar():
    print("test")

# ❌ WRONG (ambiguous - "print("test")" appears twice):
Edit(
    old_string='print("test")',
    new_string='print("production")'
)
→ Error: "old_string not unique"

# ✓ CORRECT (include enough context):
Edit(
    old_string='def foo():\n    print("test")',
    new_string='def foo():\n    print("production")'
)
```

#### Rule 5: Must Be Different

`new_string` must differ from `old_string`:

```python
# ❌ WRONG (no change):
Edit(
    old_string="x = 1",
    new_string="x = 1"
)
→ Error: "new_string must differ from old_string"

# ✓ CORRECT:
Edit(
    old_string="x = 1",
    new_string="x = 2"
)
```

### Whitespace Handling

**Whitespace must match exactly**, including:
- Spaces vs tabs
- Leading/trailing whitespace
- Blank lines

```python
# File uses tabs for indentation
def foo():
\tprint("test")  # \t = tab character

# ❌ WRONG (uses spaces):
Edit(
    old_string="def foo():\n    print(\"test\")",  # 4 spaces
)

# ✓ CORRECT (uses tab):
Edit(
    old_string="def foo():\n\tprint(\"test\")",  # tab
)
```

**Pro tip**: Always copy-paste the exact text from Read output, preserving whitespace.

### Replace All Mode

Set `replace_all=True` to replace every occurrence:

```python
# File contains:
x = oldvar
y = oldvar
z = oldvar

# Replace all occurrences
Edit(
    file_path="test.py",
    old_string="oldvar",
    new_string="newvar",
    replace_all=True
)

# Result:
x = newvar
y = newvar
z = newvar
```

**Use cases for replace_all**:
- Renaming variables throughout a file
- Updating import statements
- Changing string literals

**Warning**: Use carefully - it's easy to replace unintended matches.

### Edit Examples

#### Example 1: Simple Function Edit

```python
# Read output:
     5	def calculate_total(items):
     6	    total = 0
     7	    for item in items:
     8	        total += item.price
     9	    return total

# Edit to use sum():
Edit(
    file_path="/Users/harper/project/calc.py",
    old_string="""def calculate_total(items):
    total = 0
    for item in items:
        total += item.price
    return total""",
    new_string="""def calculate_total(items):
    return sum(item.price for item in items)"""
)
```

**Note**: Multi-line strings work fine. Whitespace must match exactly.

#### Example 2: Adding a Line

```python
# Read output:
     1	import sys
     2	import os
     3
     4	def main():

# Add new import:
Edit(
    file_path="/Users/harper/project/main.py",
    old_string="import sys\nimport os",
    new_string="import sys\nimport os\nimport logging"
)

# Result:
import sys
import os
import logging

def main():
```

#### Example 3: Modifying with Context

```python
# Read output (multiple functions):
    10	def foo():
    11	    return "foo"
    12
    13	def bar():
    14	    return "bar"
    15
    16	def baz():
    17	    return "baz"

# Want to change bar() return value
# Include enough context to be unique:
Edit(
    file_path="/Users/harper/project/funcs.py",
    old_string='def bar():\n    return "bar"',
    new_string='def bar():\n    return "updated"'
)
```

**Strategy**: Include surrounding lines if needed for uniqueness.

#### Example 4: Replace All (Renaming)

```python
# Read output:
     1	def process_data(user_data):
     2	    clean_data = sanitize(user_data)
     3	    valid_data = validate(user_data)
     4	    return user_data

# Rename user_data to raw_data throughout:
Edit(
    file_path="/Users/harper/project/process.py",
    old_string="user_data",
    new_string="raw_data",
    replace_all=True
)

# Result:
def process_data(raw_data):
    clean_data = sanitize(raw_data)
    valid_data = validate(raw_data)
    return raw_data
```

### Common Edit Pitfalls

#### Pitfall 1: Including Line Numbers

```python
# ❌ WRONG:
old_string="     5\tdef foo():"

# ✓ CORRECT:
old_string="def foo():"
```

**Fix**: Strip line number prefix (everything before and including the tab).

#### Pitfall 2: Whitespace Mismatch

```python
# File has tabs, you use spaces
# ❌ WRONG:
old_string="    print('test')"  # 4 spaces

# ✓ CORRECT:
old_string="\tprint('test')"  # tab
```

**Fix**: Copy exact text from Read output.

#### Pitfall 3: Too Little Context

```python
# File has this pattern multiple times:
return True

# ❌ WRONG (not unique):
old_string="return True"

# ✓ CORRECT (include function):
old_string="def is_valid():\n    return True"
```

**Fix**: Include enough surrounding code to make match unique.

#### Pitfall 4: Forgetting to Read

```python
# ❌ WRONG:
Edit(file_path="new_file.py", ...)

# ✓ CORRECT:
Read("new_file.py")
Edit("new_file.py", ...)
```

**Fix**: Always Read before Edit.

#### Pitfall 5: Special Characters

```python
# File contains regex pattern:
pattern = r"\d+"

# ❌ WRONG (not escaped):
old_string='pattern = r"\d+"'  # String contains \d+

# ✓ CORRECT (raw string or escape):
old_string=r'pattern = r"\d+"'  # raw string preserves backslash
```

**Fix**: Use raw strings (`r"..."`) when matching text with backslashes.

## Write Tool

The Write tool creates new files or completely rewrites existing files.

### Write Tool Signature

```python
Write(
    file_path: str,  # Absolute path to file (required)
    content: str     # Complete file content (required)
)
```

### When to Use Write

**Use Write for**:
- Creating new files
- Generating files from scratch (configs, templates, etc.)
- Complete file rewrites (rare)

**Don't use Write for**:
- Modifying existing files → Use Edit instead
- Making small changes → Use Edit instead

### Critical Rules

#### Rule 1: Read Before Rewriting

If the file exists, **must Read before Write**:

```python
# Rewriting existing file
# ❌ WRONG:
Write(file_path="existing.py", content="...")

# ✓ CORRECT:
Read("existing.py")
Write("existing.py", content="...")
```

**Why**: Write overwrites entirely. Reading first proves you understand current state.

#### Rule 2: Prefer Edit Over Write

For existing files, Edit is almost always better:

```python
# File has 200 lines, want to change 1 function

# ❌ WRONG (rewrite everything):
Read("large.py")
Write("large.py", content="<entire 200 lines with change>")

# ✓ CORRECT (change only what's needed):
Read("large.py")
Edit("large.py", old_string="old_function", new_string="new_function")
```

**Why**: Edit is safer, clearer, and respects the existing code.

#### Rule 3: Never Use Write for New Files Unless Necessary

**Prefer editing existing files over creating new ones**:

```python
# Want to add a new function

# ❌ WRONG (create new file):
Write("new_utils.py", content="def my_function(): ...")

# ✓ CORRECT (add to existing):
Read("utils.py")
Edit("utils.py", old_string="# end of file", new_string="def my_function(): ...\n# end of file")
```

**Why**: Consolidates code, reduces file sprawl.

### Write Examples

#### Example 1: Creating a New File

```python
Write(
    file_path="/Users/harper/project/config.json",
    content="""{
  "database": {
    "host": "localhost",
    "port": 5432
  },
  "debug": true
}"""
)
```

**Result**: Creates `config.json` with exact content.

#### Example 2: Generating from Template

```python
# Generate README
Write(
    file_path="/Users/harper/project/README.md",
    content="""# Project Name

## Installation

```bash
pip install -r requirements.txt
```

## Usage

```python
from app import run
run()
```
"""
)
```

#### Example 3: Rewriting After Major Refactor

```python
# Rare case: Complete restructure of small file
Read("old_structure.py")  # Understand current state

Write(
    file_path="old_structure.py",
    content="""# Completely new structure
class NewDesign:
    def __init__(self):
        self.data = {}

    def process(self):
        return self.data
"""
)
```

**Note**: Only do this when Edit would be more complex than rewriting.

### Write Pitfalls

#### Pitfall 1: Unnecessary Rewrites

```python
# Change one line in a 100-line file

# ❌ WRONG:
Write(file_path="big.py", content="<all 100 lines>")

# ✓ CORRECT:
Edit(file_path="big.py", old_string="old_line", new_string="new_line")
```

**Fix**: Use Edit for modifications.

#### Pitfall 2: Forgetting to Read

```python
# ❌ WRONG (rewriting without understanding):
Write(file_path="existing.py", content="...")

# ✓ CORRECT:
Read("existing.py")
Write("existing.py", content="...")
```

**Fix**: Always Read before Write for existing files.

#### Pitfall 3: Creating Unnecessary Files

```python
# ❌ WRONG (file sprawl):
Write("utils_new.py", content="...")
Write("helpers_extra.py", content="...")

# ✓ CORRECT (consolidate):
Edit("utils.py", ...)  # Add to existing
```

**Fix**: Prefer modifying existing files.

## File Path Requirements

### Absolute Paths

**All tools require absolute paths**:

```python
# ❌ WRONG (relative):
Read("src/main.py")

# ✓ CORRECT (absolute):
Read("/Users/harper/project/src/main.py")
```

### Getting Absolute Paths

Use working directory:

```python
# If working directory is /Users/harper/project
working_dir = "/Users/harper/project"

# Construct absolute paths
file_path = f"{working_dir}/src/main.py"
Read(file_path)
```

Or use Bash:

```bash
# Get absolute path
pwd  # → /Users/harper/project
readlink -f src/main.py  # → /Users/harper/project/src/main.py
```

## Multi-File Operations

### Editing Multiple Files (Parallel)

If edits are independent:

```python
# All in one response (parallel execution)
Edit("/Users/harper/project/file1.py", old_string="...", new_string="...")
Edit("/Users/harper/project/file2.py", old_string="...", new_string="...")
Edit("/Users/harper/project/file3.py", old_string="...", new_string="...")
```

**All three edits execute simultaneously**.

### Editing Multiple Files (Sequential)

If later edits depend on earlier results:

```python
# Response 1: Read and edit first file
Read("/Users/harper/project/base.py")
Edit("/Users/harper/project/base.py", ...)

# Response 2: Use results to edit second file
Read("/Users/harper/project/dependent.py")
Edit("/Users/harper/project/dependent.py", ...)
```

**Edits happen in separate responses when dependent**.

## Best Practices

### 1. Always Read Before Edit

```python
✓ Read → Edit
❌ Edit (without reading)
```

### 2. Match Exact Whitespace

```python
✓ Copy-paste from Read output
❌ Type manually (easy to get wrong)
```

### 3. Include Sufficient Context

```python
✓ old_string = "def foo():\n    return bar"  # Unique
❌ old_string = "return bar"  # Might appear multiple times
```

### 4. Prefer Edit Over Write

```python
✓ Edit existing files (surgical changes)
❌ Write to rewrite entire files (risky)
```

### 5. Use replace_all Carefully

```python
✓ replace_all for intentional global changes (variable rename)
❌ replace_all without verifying all matches are desired
```

### 6. Test After Editing

```python
# After editing code:
Bash("pytest")  # Verify changes work
```

### 7. Group Related Reads

```python
✓ Read all needed files in parallel (one response)
❌ Read files one at a time (multiple responses)
```

## Debugging Edit Failures

### Error: "old_string not found"

**Cause**: Text doesn't match exactly.

**Fix**:
1. Re-read the file
2. Copy exact text (including whitespace)
3. Check for tabs vs spaces
4. Verify no typos

### Error: "old_string not unique"

**Cause**: Text appears multiple times.

**Fix**:
1. Include more context (surrounding lines)
2. Make the match more specific
3. Or use `replace_all=True` if intentional

### Error: "File must be read before editing"

**Cause**: Didn't read file in current conversation.

**Fix**:
1. Read the file
2. Then edit

### Error: "new_string must differ"

**Cause**: Tried to replace text with identical text.

**Fix**:
1. Actually change something
2. Or don't edit if no change needed

## Real-World Workflow Example

### Task: Add logging to a function

```python
# Step 1: Find the file
Grep(pattern="def process_data", output_mode="files_with_matches")
# Result: src/processor.py

# Step 2: Read the file
Read("/Users/harper/project/src/processor.py")
# Result:
#      1	import json
#      2
#      3	def process_data(data):
#      4	    result = json.loads(data)
#      5	    return result

# Step 3: Add logging import
Edit(
    file_path="/Users/harper/project/src/processor.py",
    old_string="import json",
    new_string="import json\nimport logging"
)

# Step 4: Add logging to function
Edit(
    file_path="/Users/harper/project/src/processor.py",
    old_string="""def process_data(data):
    result = json.loads(data)
    return result""",
    new_string="""def process_data(data):
    logging.debug(f"Processing data: {data}")
    result = json.loads(data)
    logging.debug(f"Parsed result: {result}")
    return result"""
)

# Step 5: Verify
Bash("python -m pytest tests/test_processor.py")
```

**Note**: Two separate edits (import, then function) to keep each change focused.

## Summary

| Tool | Purpose | Must Read First? | Overwrites File? |
|------|---------|------------------|------------------|
| **Read** | View file contents | N/A | No |
| **Edit** | Modify existing files | Yes | No (surgical) |
| **Write** | Create or rewrite files | Yes (if exists) | Yes (complete) |

**Golden rules**:
1. Read before Edit (always)
2. Match exact whitespace
3. Ignore line number prefixes
4. Include enough context for uniqueness
5. Prefer Edit over Write for changes
6. Use absolute paths

---

*Next: [03-TOOL-SYSTEM.md](./03-TOOL-SYSTEM.md) for complete tool inventory and selection logic.*
