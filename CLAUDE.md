# Claude Instructions for Hex Development

**Project**: Hex - Claude CLI Tool
**Philosophy**: Use hex to develop hex (dogfooding/meta-development)

---

## Recent Work: TUI Bug Fixes (2025-12-07)

Fixed 4 critical/high severity bugs from comprehensive TUI code review:

1. ✅ **Tool Results Visibility** (commit 6009c06)
   - Tool execution results now visible to users with visual indicators
   - Added 🔧 for tool results, 🛠 for tool calls
   - Fixed transparency violation where users couldn't see what tools did

2. ✅ **Stream Cancellation Memory Leaks** (commit 2d23c99)
   - Created `cancelStream()` helper for proper cleanup
   - Fixed 5 locations that leaked contexts and goroutines
   - Prevents memory leaks during streaming operations

3. ✅ **Button Mashing Vulnerability** (commit 9c0f0a2)
   - Added guard in `ApproveToolUse()` to prevent double-execution
   - Protects against accidental multiple tool runs from rapid key presses
   - Critical for destructive operations

4. ✅ **Viewport Throttling** (commit 1e9457b)
   - Throttling now applies to ALL viewport updates, not just streaming
   - Prevents CPU spikes from expensive glamour markdown renders
   - 60fps limit for smooth performance

**Methodology**: These fixes were made using direct file editing (Edit tool) because we were fixing bugs IN the TUI itself. Going forward, use hex for all development work.

---

## Core Principle: Hex First, Always

When working on the hex codebase, **YOU MUST USE HEX ITSELF** for all development tasks. This is non-negotiable.

### Why?

1. **Dogfooding** - If hex can't develop itself, it's not production-ready
2. **Real validation** - Using hex daily exposes bugs and UX issues
3. **Meta-testing** - Every development session is a live integration test
4. **Quality gate** - If a task is too hard with hex, hex needs improvement

---

## Mandatory Workflow

### ❌ WRONG: Direct File Operations

```bash
# DON'T DO THIS:
cat internal/tools/grep_tool.go
vim internal/tools/registry.go
grep -r "getToolSchema" internal/
```

### ✅ RIGHT: Use Hex

```bash
# DO THIS:
source .env  # Load API key
./hex -p --permission-mode=auto --tools=read_file "Read internal/tools/grep_tool.go and analyze the Execute function"

./hex -p --permission-mode=auto --tools=read_file,edit "Read internal/tools/registry.go and add a case for grep tool in getToolSchema function. Change lines 159-165 to add the grep schema before the default case."

./hex -p --permission-mode=auto --tools=grep "Search for getToolSchema in internal/ directory"
```

---

## Required Tools for Development

### Investigation Tasks
```bash
# Code reading and analysis
./hex -p --tools=read_file "..."

# Code searching
./hex -p --tools=grep,glob "..."

# Understanding codebase structure
./hex -p --tools=grep,glob,read_file "..."
```

### Modification Tasks
```bash
# Editing existing files
./hex -p --permission-mode=auto --tools=read_file,edit "..."

# Creating new files
./hex -p --permission-mode=auto --tools=write_file "..."

# Complex multi-file changes
./hex -p --permission-mode=auto --tools=read_file,edit,grep,bash "..."
```

### Testing & Verification
```bash
# Run tests
./hex -p --permission-mode=auto --tools=bash "..."

# Check build status
./hex -p --permission-mode=auto --tools=bash "Run 'make build' and report any errors"

# Verify fix
./hex -p --permission-mode=auto --tools=bash,read_file "..."
```

---

## Development Patterns

### Pattern 1: Bug Investigation

```bash
# 1. Enable debug mode to see what's happening
./hex -p --debug --tools=grep "Find files with 'calculate'"

# 2. Read relevant code
./hex -p --tools=read_file "Read internal/tools/grep_tool.go and explain the Execute function"

# 3. Search for related code
./hex -p --tools=grep,glob "Find all usages of getToolSchema in the codebase"

# 4. Analyze the issue
./hex -p --tools=read_file "Read internal/tools/registry.go starting at line 84 and explain the getToolSchema function"
```

### Pattern 2: Implementing a Fix

```bash
# 1. Understand what needs to change
./hex -p --tools=read_file "Read internal/tools/registry.go and identify where to add grep tool schema"

# 2. Make the change (hex must read the file first for Edit tool)
./hex -p --permission-mode=auto --tools=read_file,edit \
  "Read internal/tools/registry.go. Then use edit tool to add a case for 'grep' tool in the getToolSchema switch statement, before the default case. Add comprehensive schema with pattern, path, output_mode, -i, -A, -B, -C, glob, and type parameters."

# 3. Verify the change
./hex -p --tools=read_file "Read internal/tools/registry.go lines 159-220 and confirm the grep schema was added"

# 4. Test it
./hex -p --permission-mode=auto --tools=bash "Run 'make build' and report success or errors"
```

### Pattern 3: Creating New Features

```bash
# Use hex to help design and implement
./hex -p --tools=read_file,grep "Analyze the existing tool implementations and design a new XYZ tool"

# Use hex to create the files
./hex -p --permission-mode=auto --tools=write_file,read_file \
  "Create internal/tools/xyz_tool.go with implementation based on the pattern from grep_tool.go"

# Use hex to integrate it
./hex -p --permission-mode=auto --tools=read_file,edit \
  "Add registration for XYZ tool in the appropriate init file"
```

---

## Tool Selection Guide

### Always Available in Print Mode
- `read_file` - Read any file
- `write_file` - Create or overwrite files
- `edit` - Modify existing files (requires reading first)
- `bash` - Run shell commands
- `grep` - Search code with patterns
- `glob` - Find files by pattern

### When to Use Each

| Task | Tools | Example |
|------|-------|---------|
| Read code | `read_file` | "Read internal/tools/grep_tool.go" |
| Search codebase | `grep`, `glob` | "Find all files importing 'core'" |
| Modify code | `read_file`, `edit` | "Change version from 1.0 to 1.1" |
| Create files | `write_file` | "Create new tool implementation" |
| Run tests | `bash` | "Run make test and show results" |
| Multi-step task | `read_file`, `edit`, `bash` | "Add feature, test, commit" |

---

## Best Practices

### 1. Always Use --permission-mode=auto for Edits

```bash
# This avoids manual approval prompts in non-interactive mode
./hex -p --permission-mode=auto --tools=read_file,edit "..."
```

### 2. Read Before Edit

The Edit tool requires you to read the file first (it does exact string replacement):

```bash
# ✅ CORRECT
./hex -p --permission-mode=auto --tools=read_file,edit \
  "Read file.go. Then edit it to change X to Y."

# ❌ WRONG - will fail
./hex -p --permission-mode=auto --tools=edit \
  "Edit file.go to change X to Y"
```

### 3. Use Debug Mode for Investigations

```bash
# See exactly what hex is doing
./hex -p --debug --tools=grep "..."

# Logs go to /tmp/hex-debug.log
tail -f /tmp/hex-debug.log
```

### 4. Be Specific with File Paths

```bash
# ✅ GOOD - absolute or relative from project root
./hex -p --tools=read_file "Read internal/tools/grep_tool.go"

# ❌ BAD - vague path
./hex -p --tools=read_file "Read the grep tool file"
```

### 5. Combine Tools for Complex Tasks

```bash
# Multi-step workflow
./hex -p --permission-mode=auto \
  --tools=read_file,grep,edit,bash \
  "1) Find all files with 'TODO' comments using grep
   2) Read the most important one
   3) Fix the TODO by editing the file
   4) Run tests to verify the fix"
```

---

## Common Tasks

### Add a New Tool

```bash
# 1. Design it
./hex -p --tools=read_file,grep \
  "Read examples of existing tools like grep_tool.go and bash_tool.go. Design a new_tool.go based on these patterns."

# 2. Implement it
./hex -p --permission-mode=auto --tools=write_file \
  "Create internal/tools/new_tool.go with implementation"

# 3. Add schema
./hex -p --permission-mode=auto --tools=read_file,edit \
  "Read internal/tools/registry.go and add schema for new_tool in getToolSchema function"

# 4. Register it
./hex -p --permission-mode=auto --tools=read_file,edit \
  "Read cmd/hex/root.go and add registration for new_tool"

# 5. Test it
./hex -p --permission-mode=auto --tools=bash \
  "Run 'make build && ./hex -p --tools=new_tool \"test it\"'"
```

### Fix a Bug

```bash
# 1. Reproduce with debug
./hex -p --debug --tools=... "Trigger the bug"

# 2. Investigate
./hex -p --tools=read_file,grep "Find and analyze the buggy code"

# 3. Fix
./hex -p --permission-mode=auto --tools=read_file,edit "Apply the fix"

# 4. Verify
./hex -p --permission-mode=auto --tools=bash "Run tests"
```

### Add Documentation

```bash
./hex -p --permission-mode=auto --tools=read_file,write_file \
  "Read the current README.md. Create FEATURE.md documenting the XYZ feature based on the code in internal/tools/xyz_tool.go"
```

### Run Tests

```bash
# Unit tests
./hex -p --permission-mode=auto --tools=bash "Run 'make test' and report results"

# Integration tests
./hex -p --permission-mode=auto --tools=bash "Run scenario tests in .scratch/"

# Specific test
./hex -p --permission-mode=auto --tools=bash "Run 'go test -v internal/tools/grep_tool_test.go'"
```

---

## When NOT to Use Hex

### Exceptions (Use Direct Commands)

1. **Building**: `make build` - faster than through hex
2. **Git operations**: `git commit`, `git push` - direct is fine
3. **Installing dependencies**: `go mod tidy` - system operation
4. **Running hex**: Obviously can't use hex to run hex!

### But Consider Using Hex For:

1. **Git investigation**: "What changed in this commit?"
2. **Dependency analysis**: "Which files import package X?"
3. **Complex git**: "Create a commit message based on the changes"

---

## Testing New Features

### Always Test with Hex Itself

When you add a new feature, test it by using hex:

```bash
# If you add a new tool:
./hex -p --permission-mode=auto --tools=new_tool "Test the new tool"

# If you add a new flag:
./hex --new-flag -p "Test the flag"

# If you fix a bug:
./hex -p "Try the scenario that was broken"
```

### Scenario Testing

Create `.scratch/scenario_N_feature_name.sh` tests:

```bash
#!/usr/bin/env bash
# Test new feature using real hex binary

source .env
./hex -p --permission-mode=auto --tools=... "Test new feature"

# Verify it worked
if [ $? -eq 0 ]; then
    echo "✅ PASS"
else
    echo "❌ FAIL"
    exit 1
fi
```

---

## Meta-Development Benefits

### What We Learn By Using Hex

1. **UX issues** - "This prompt is confusing" → improve tool descriptions
2. **Missing features** - "I need to do X" → add new tool
3. **Performance** - "This is slow" → optimize
4. **Bugs** - "This failed" → fix immediately
5. **Documentation** - "How do I..." → improve docs

### Tracking Improvements

When you find an issue while using hex:

```bash
# Document it
echo "Issue: Can't easily XYZ" >> .scratch/hex_ux_issues.txt

# Fix it
./hex -p --permission-mode=auto --tools=read_file,edit "Implement solution"

# Test it
./hex -p "Try the improved workflow"
```

---

## Emergency Escape Hatch

### If Hex is Completely Broken

If hex is so broken you can't use it to fix itself:

1. **Fix minimally by hand** - Just enough to make hex work again
2. **Immediately use hex** - Use the working hex to make further fixes
3. **Document why** - Explain in commit message why hand-editing was needed

```bash
# Emergency fix
vim internal/tools/broken.go  # Fix critical syntax error

# Rebuild
make build

# NOW USE HEX for everything else
./hex -p --tools=read_file "Read broken.go and verify the fix"
./hex -p --tools=read_file,edit "Improve the fix properly"
```

---

## Commit Messages

### When Committing Hex-Developed Changes

Include that hex was used:

```
fix: add JSON schemas for grep and glob tools

Previously grep and glob tools had empty input schemas.
Fixed using hex to debug hex itself.

Investigation:
- Used ./hex -p --debug to discover empty schemas
- Used ./hex -p --tools=read_file to analyze code
- Used ./hex -p to generate schemas

Verified with scenario tests.
```

---

## TL;DR - The Rules

1. ✅ **USE HEX FOR**: Reading code, searching code, editing code, creating files
2. ✅ **USE HEX FOR**: Investigation, analysis, understanding codebase
3. ✅ **USE HEX FOR**: Testing new features, verifying fixes
4. ❌ **DON'T USE HEX FOR**: Building (use `make`), git commits (use `git`)
5. 🎯 **ALWAYS**: Use `--permission-mode=auto` for non-interactive edits
6. 🎯 **ALWAYS**: Read files before editing them
7. 🎯 **ALWAYS**: Use `--debug` when investigating issues
8. 🎯 **PRINCIPLE**: If it's too hard with hex, that's a hex bug - fix hex!

---

## Examples from Real Development

### How We Fixed the Grep/Glob Bug

```bash
# 1. Discovery (using hex with --debug)
./hex -p --debug --tools=grep "Find files with calculate"
# Saw: "input_schema": {"properties": {}} ← EMPTY!

# 2. Investigation (using hex to read code)
./hex -p --tools=read_file \
  "Read internal/tools/grep_tool.go and explain Execute function"

./hex -p --tools=grep \
  "Find getToolSchema function in internal/tools/"

./hex -p --tools=read_file \
  "Read internal/tools/registry.go and show getToolSchema function"

# 3. Solution (using hex to generate schemas)
./hex -p --tools=read_file,write_file \
  "Generate JSON schema for grep tool parameters"

# 4. Implementation (manual - but based on hex output)
# Applied the schemas to registry.go

# 5. Verification (using hex to test)
./hex -p --tools=bash "Run 'make build'"
./hex -p --tools=grep,glob "Test if grep works now"

# Result: ✅ BUG FIXED using hex to debug hex!
```

---

**Remember**: Every time you use hex to develop hex, you're validating that hex is useful. If hex can't help you, improve hex!

🔄 **Meta-development is the ultimate integration test.**
