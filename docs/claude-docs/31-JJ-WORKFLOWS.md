# JJ (Jujutsu) Workflows

**ABOUTME: Complete guide to using JJ (Jujutsu) version control system as the preferred VCS in Claude Code**
**ABOUTME: Covers JJ equivalents for git commands, when to use JJ vs git, and JJ workflow patterns**

## Overview

**JJ (Jujutsu)** is a next-generation version control system that improves on git with:
- Automatic local branching (every commit is its own branch)
- Built-in conflict resolution
- Safe, reversible operations
- Cleaner mental model (working copy as a commit)
- Better merge and rebase workflows

**This document provides JJ guidance for Claude Code usage.**

## Installation and Setup

### Check if JJ is Available

```bash
# Check if jj is installed
jj --version

# If not installed, fall back to git
git --version
```

### Initialize JJ in Repository

```bash
# Initialize new JJ repository
jj init

# Or initialize in existing git repo
jj git init --colocate
```

The `--colocate` flag allows JJ and git to coexist, sharing the same .git directory.

## Core Philosophy

### Git vs JJ Mental Model

**Git:**
- Commits are nodes in a graph
- HEAD points to current commit
- Working directory is separate from commits
- Branches are pointers to commits

**JJ:**
- Working copy IS a commit (always)
- Every change creates a new commit automatically
- Branches track changes, not commits
- Operations are logged and reversible

### The Working Copy Commit

In JJ, **your working directory is automatically a commit**:

```bash
# In git, you must explicitly commit
git add .
git commit -m "message"

# In JJ, changes are automatically committed to working copy
jj status  # Shows working copy commit
```

## JJ Command Equivalents

### Basic Operations

| Task | Git | JJ | Notes |
|------|-----|-----|-------|
| **Check status** | `git status` | `jj status` | JJ shows working copy commit |
| **View changes** | `git diff` | `jj diff` | Shows diff of working copy |
| **Commit changes** | `git commit -m "msg"` | `jj commit -m "msg"` | Creates new change from working copy |
| **View log** | `git log` | `jj log` | JJ shows graphical view by default |
| **Undo last change** | `git reset HEAD~1` | `jj undo` | JJ undo is safe and reversible |

### Branching and Merging

| Task | Git | JJ | Notes |
|------|-----|-----|-------|
| **Create branch** | `git checkout -b name` | `jj branch create name` | JJ auto-creates branch for working copy |
| **Switch branch** | `git checkout name` | `jj edit <change>` | Edit specific change |
| **List branches** | `git branch` | `jj branch list` | Shows all branches |
| **Merge branch** | `git merge branch` | `jj merge <changes>` | JJ creates merge commit |
| **Rebase** | `git rebase main` | `jj rebase -d main` | JJ rebase is non-destructive |

### Remote Operations

| Task | Git | JJ | Notes |
|------|-----|-----|-------|
| **Clone repo** | `git clone url` | `jj git clone url` | JJ wraps git for remotes |
| **Fetch** | `git fetch` | `jj git fetch` | Updates remote refs |
| **Pull** | `git pull` | `jj git fetch && jj rebase` | JJ separates fetch and rebase |
| **Push** | `git push` | `jj git push` | JJ wraps git push |
| **Set upstream** | `git push -u origin br` | `jj git push --branch br` | JJ infers branch |

### Advanced Operations

| Task | Git | JJ | Notes |
|------|-----|-----|-------|
| **Amend commit** | `git commit --amend` | `jj commit --amend` | Or just `jj commit` (amends by default) |
| **Interactive rebase** | `git rebase -i` | `jj squash` / `jj split` | JJ provides granular operations |
| **Cherry-pick** | `git cherry-pick hash` | `jj rebase -r hash -d dest` | Rebase specific change |
| **View reflog** | `git reflog` | `jj op log` | JJ operation log |
| **Undo operation** | `git reset --hard hash` | `jj undo` | JJ undo is always safe |

## Common Workflows

### Workflow: Daily Development

```bash
# 1. Start working (working copy is always a commit)
jj status

# 2. Make changes to files
# (use Edit tool to modify code)

# 3. View changes
jj diff

# 4. Commit working copy changes
jj commit -m "feat: add user authentication"

# 5. Working copy is now a new empty commit, ready for next changes
```

**Key difference from git:** You don't need to `git add .` - changes are automatically tracked.

### Workflow: Creating a Feature Branch

```bash
# 1. Create and switch to new change
jj new -m "start authentication feature"

# 2. JJ automatically creates a branch tracking this change
jj branch create feature/auth

# 3. Make changes
# (edit files)

# 4. Commit (amends current change)
jj commit -m "feat: implement OAuth flow"

# 5. Push to remote
jj git push --branch feature/auth
```

### Workflow: Syncing with Remote

```bash
# 1. Fetch updates from remote
jj git fetch

# 2. View what changed
jj log

# 3. Rebase your work onto latest main
jj rebase -d main

# 4. If conflicts, resolve them
# (JJ shows conflict markers in files)
jj status  # Shows conflicts

# 5. After resolving, continue
jj rebase --continue

# 6. Push your changes
jj git push --branch feature/auth
```

### Workflow: Fixing Mistakes

```bash
# Made a mistake in last commit
jj undo

# Made a mistake 3 operations ago
jj op log  # Find operation ID
jj undo --op <operation-id>

# Want to edit an older change
jj log  # Find change ID
jj edit <change-id>
# Make edits
jj commit --amend
```

### Workflow: Squashing Commits

```bash
# Git: git rebase -i HEAD~3 (then pick/squash)

# JJ: Squash current change into parent
jj squash

# JJ: Squash specific change
jj squash -r <change-id>

# JJ: Squash with message
jj squash -m "Combined feature changes"
```

### Workflow: Splitting a Change

```bash
# Git: git reset HEAD~1, then git add -p, git commit (multiple times)

# JJ: Split current change
jj split

# JJ opens editor showing diff
# Delete lines you DON'T want in first commit
# Save and exit

# Now you have two changes:
# - First change: lines you kept
# - Second change: lines you removed from first
```

## JJ in Claude Code Context

### When to Use JJ vs Git

**Use JJ for:**
- ✅ Daily development workflow
- ✅ Creating and managing changes
- ✅ Rebasing and merging
- ✅ Undoing mistakes
- ✅ Splitting and squashing commits

**Use git for:**
- ⚠️ When JJ not available (`jj --version` fails)
- ⚠️ Pre-commit hooks (some hook systems expect git)
- ⚠️ CI/CD systems that expect git commands
- ⚠️ When explicitly requested by user

### Checking JJ Availability

```bash
# Before using JJ, verify it's available
if command -v jj &> /dev/null; then
    jj status
else
    git status
fi
```

### Commit Message Format (Same as Git)

```bash
# JJ follows conventional commit format (same as git)
jj commit -m "feat: add user authentication"
jj commit -m "fix: resolve login race condition"
jj commit -m "docs: update API documentation"
```

### Pre-commit Hooks with JJ

**Important:** Some pre-commit hook systems expect git. If hooks fail with JJ:

```bash
# Try with git instead
git commit -m "message"

# Or configure pre-commit to work with JJ
# (depends on hook configuration)
```

## JJ Cheat Sheet

### Quick Reference

```bash
# Status and changes
jj status              # Show working copy status
jj diff                # Show changes in working copy
jj log                 # Show change log (graphical)

# Committing
jj commit -m "msg"     # Commit working copy changes
jj commit --amend      # Amend current change
jj new                 # Create new empty change

# Branching
jj branch create name  # Create branch
jj branch list         # List branches
jj branch delete name  # Delete branch

# Syncing
jj git fetch           # Fetch from remote
jj git push            # Push to remote
jj rebase -d main      # Rebase onto main

# Editing history
jj edit <change>       # Edit specific change
jj squash              # Squash into parent
jj split               # Split current change
jj undo                # Undo last operation

# Resolving conflicts
jj status              # Shows conflicts
jj resolve             # Start conflict resolution
# (edit files to resolve)
jj rebase --continue   # Continue after resolution
```

### Common Patterns

```bash
# Pattern: Quick commit
jj commit -m "fix: typo"

# Pattern: Amend last commit
# (just make changes and commit again)
jj commit --amend

# Pattern: Undo mistake
jj undo

# Pattern: Rebase onto latest main
jj git fetch && jj rebase -d main

# Pattern: Push current branch
jj git push

# Pattern: View full history
jj log --all

# Pattern: Show change details
jj show <change-id>
```

## Conflict Resolution

### JJ Conflict Resolution Workflow

JJ handles conflicts differently from git:

```bash
# 1. Conflicts detected during rebase
jj rebase -d main
# Output: Conflict in file.txt

# 2. View conflicts
jj status
# Shows: file.txt has conflicts

# 3. JJ shows conflicts in file with markers
# <<<<<<<
# |||||||
# =======
# >>>>>>>

# 4. Edit file to resolve conflicts
# (use Edit tool)

# 5. Mark as resolved (automatic when you edit file)
jj status
# Should show no conflicts

# 6. Continue rebase
jj rebase --continue
```

### Conflict Markers in JJ

JJ uses 3-way conflict markers:

```
<<<<<<< Destination (main)
Code from main branch
||||||| Base
Original code before changes
=======
Your changes
>>>>>>> Source (feature/auth)
```

**Resolve by editing file to final desired state**, then save.

## Advanced JJ Features

### Operation Log

JJ maintains a log of ALL operations (like git reflog but better):

```bash
# View operation log
jj op log

# Output shows all operations with IDs:
# @ 5a3b2c1 2025-12-01 14:30:00 commit: "feat: add auth"
# o 9f8e7d6 2025-12-01 14:25:00 rebase
# o 3c4b5a6 2025-12-01 14:20:00 commit: "fix: bug"

# Undo to specific operation
jj undo --op 9f8e7d6
```

### Change IDs vs Commit IDs

**JJ uses change IDs** (stable across rebases):

```bash
# View change ID
jj log
# Shows: qpvuntsm user@host 2025-12-01 14:30:00 feat: add auth

# Git commit ID changes on rebase
# JJ change ID stays the same

# Reference change by prefix
jj show qpvu  # Short prefix is enough
```

### Revsets (Powerful Query Language)

```bash
# Show all changes on current branch
jj log -r 'mine()'

# Show changes that will be pushed
jj log -r 'remote..@'

# Show changes with specific message
jj log -r 'description(auth)'

# Show changes by author
jj log -r 'author(harper)'
```

## Integration with Claude Code Workflows

### Workflow: Feature Development

```bash
# 1. Check status (prefer JJ)
jj status

# 2. Create feature branch
jj branch create feature/user-auth

# 3. Make changes
# (Claude edits files using Edit tool)

# 4. Commit changes
jj commit -m "feat: implement user authentication

- Add OAuth 2.0 flow
- Integrate with Google/GitHub
- Add JWT token generation
- All tests passing

🤖 Generated with Claude Code"

# 5. Push to remote
jj git push --branch feature/user-auth
```

### Workflow: Pre-commit Hook Integration

```bash
# Attempt commit with JJ
jj commit -m "feat: add feature"

# If using colocated git+jj repo and git hooks fail:
# 1. The git hooks will run automatically
# 2. If they fail, JJ commit is blocked

# Fix issues and retry
jj commit -m "feat: add feature"
```

### Workflow: Creating Pull Request

```bash
# 1. Ensure changes committed
jj status

# 2. Push branch
jj git push --branch feature/auth

# 3. Create PR using gh CLI (works same as with git)
gh pr create --title "Add authentication" --body "..."
```

## Troubleshooting

### Issue: JJ not found

```
Error: command not found: jj
```

**Solution:** Fall back to git:

```bash
git status
git commit -m "message"
```

### Issue: Colocated repo not initialized

```
Error: not a jj repository
```

**Solution:** Initialize JJ in existing git repo:

```bash
jj git init --colocate
```

### Issue: Conflicts during rebase

```
Error: Conflicts in 3 files
```

**Solution:** Resolve conflicts:

```bash
# 1. View conflicted files
jj status

# 2. Edit files to resolve conflicts
# (use Edit tool)

# 3. Continue rebase
jj rebase --continue
```

### Issue: Need to undo multiple operations

```bash
# View operation history
jj op log

# Undo to specific point
jj undo --op <operation-id>
```

## Best Practices

### ✅ Do:

- Prefer JJ over git when available
- Use `jj commit` frequently (it's cheap)
- Leverage `jj undo` for safe experimentation
- Use `jj squash` to clean up history
- Check `jj status` regularly
- Use change IDs for referencing changes

### ❌ Don't:

- Force JJ when it's not available (fall back to git gracefully)
- Bypass pre-commit hooks
- Forget to `jj git fetch` before rebasing
- Use JJ if CI/CD explicitly requires git
- Delete the .git directory (needed for git interop)

## Summary

**JJ Workflow Quick Start:**

```bash
# Daily workflow
jj status                          # Check status
jj commit -m "message"             # Commit changes
jj git fetch                       # Fetch updates
jj rebase -d main                  # Rebase onto main
jj git push --branch feature/name  # Push to remote

# Fixing mistakes
jj undo                            # Undo last operation
jj squash                          # Squash into parent
jj split                           # Split change

# Advanced
jj op log                          # View operation log
jj log -r 'mine()'                 # View my changes
jj edit <change>                   # Edit specific change
```

**When to use:**
- ✅ JJ as primary VCS (when available)
- ⚠️ Git as fallback (when JJ unavailable or CI requires it)

## See Also

- [15-GIT-WORKFLOWS.md](./15-GIT-WORKFLOWS.md) - Git workflows (use as fallback)
- [04-BASH-AND-COMMAND-EXECUTION.md](./04-BASH-AND-COMMAND-EXECUTION.md) - Running VCS commands
- [JJ Official Docs](https://martinvonz.github.io/jj/) - Complete JJ documentation
