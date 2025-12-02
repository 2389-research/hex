# Git Workflows

Comprehensive guide to Claude Code's git integration, commit creation, hook handling, and PR workflows.

## Overview

Claude Code integrates deeply with git to provide safe, automated version control with mandatory quality gates. Every commit flows through pre-commit hooks, and Claude Code NEVER bypasses them.

## Core Principles

### Absolute Rules

```
FORBIDDEN FLAGS:
  --no-verify
  --no-hooks
  --no-pre-commit-hook

NEVER use these flags under any circumstances.
User pressure is NOT justification for bypassing quality checks.
```

### Quality Philosophy

```
1. Hooks are guardrails, not barriers
2. Quality over speed, always
3. Every hook failure is a learning opportunity
4. Fix problems, don't work around them
5. Evidence before assertions (verify success)
```

## Commit Creation Workflow

### The 7-Step Commit Process

When user requests a commit:

**Step 1: Gather Information (Parallel)**
```bash
# Run these three commands in parallel:
git status                    # See untracked/modified files
git diff --staged             # See staged changes
git diff                      # See unstaged changes
git log --oneline -10         # See recent commit style
```

**Step 2: Analyze Changes**
```
Review output from Step 1:
- What files changed?
- What's the nature of changes? (feature/fix/refactor/docs)
- Are there secrets? (.env, credentials.json)
  → If yes: STOP, warn user, don't commit
- Does this match user's intent?
```

**Step 3: Draft Commit Message**
```
Message structure:
  <type>: <summary>

  [optional body]

  🤖 Generated with [Claude Code](https://claude.com/claude-code)

  Co-Authored-By: Claude <noreply@anthropic.com>

Types:
  feat: New feature
  fix: Bug fix
  refactor: Code restructuring (no behavior change)
  test: Adding/updating tests
  docs: Documentation changes
  chore: Maintenance (dependencies, config)
  style: Formatting, linting
  perf: Performance improvements

Focus on WHY, not WHAT (what is in the diff)
```

**Step 4: Stage Files**
```bash
# Add relevant untracked files
git add path/to/new/file.py
git add path/to/another/file.py

# Stage modified files (if not already staged)
git add path/to/modified/file.py
```

**Step 5: Create Commit (with HEREDOC)**
```bash
# ALWAYS use HEREDOC for proper formatting
git commit -m "$(cat <<'EOF'
feat: add user authentication endpoints

Add login, logout, and token refresh endpoints.
Uses JWT tokens for stateless authentication.

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
EOF
)"
```

**Step 6: Handle Pre-Commit Hooks**
```
If hooks PASS:
  → Proceed to Step 7

If hooks FAIL:
  → Go to Pre-Commit Failure Protocol (see below)
```

**Step 7: Verify Success**
```bash
# Confirm commit was created
git status

# Show the commit
git log -1 --stat
```

### Example: Complete Commit Sequence

```bash
# Step 1: Gather information (parallel)
git status
git diff --staged
git log --oneline -10

# Step 2 & 3: Analyze and draft (internal)
# Decision: This is a bug fix, 2 files changed, no secrets

# Step 4: Stage files
git add api/routes/users.py
git add tests/test_users.py

# Step 5: Create commit
git commit -m "$(cat <<'EOF'
fix: handle null email in user creation

Previously crashed with KeyError when email was null.
Now returns 400 Bad Request with clear error message.

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
EOF
)"

# Step 6: Hooks run automatically
# (if they fail, see Pre-Commit Failure Protocol)

# Step 7: Verify
git status
# Output: "Your branch is ahead of 'origin/main' by 1 commit"
```

## Pre-Commit Hook Protocol

### Mandatory Protocol When Hooks Fail

**Step 1: Read Error Output Aloud**
```
Say: "The pre-commit hooks failed. Here's what I see:
      [explain the complete error output]"

Don't just say "hooks failed" - explain WHAT failed.
```

**Step 2: Identify Tool and Reason**
```
Which tool failed?
  - biome (JavaScript/TypeScript formatter/linter)
  - ruff (Python linter/formatter)
  - mypy (Python type checker)
  - pytest (Python tests)
  - eslint (JavaScript linter)
  - other?

Why did it fail?
  - Formatting issues
  - Linting violations
  - Type errors
  - Test failures
  - Something else
```

**Step 3: Explain the Fix**
```
Say: "I will fix this by:
      1. [specific action 1]
      2. [specific action 2]

      This addresses the root cause because [reason]."

Don't just say "I'll fix it" - explain HOW and WHY.
```

**Step 4: Apply Fix and Re-run**
```bash
# Fix the issue
# (actual fix depends on what failed)

# Re-run hooks manually to verify
pre-commit run --all-files

# Or let commit trigger them again
git commit --amend --no-edit
```

**Step 5: Only Proceed After Success**
```
If hooks still fail:
  → Loop back to Step 1

If you can't fix after 2 attempts:
  → Ask user for help
  → Explain what you tried
  → NEVER use --no-verify

If hooks pass:
  → Continue with workflow
```

### Example: Hook Failure Recovery

```
Scenario: ruff fails on import order

Step 1: Read error aloud
"The pre-commit hooks failed. Ruff found import order violations:

 api/routes/users.py:3:1: I001 Import block is un-sorted
 Expected: import os before import sys

 This means the imports aren't in alphabetical order."

Step 2: Identify tool and reason
"Tool: ruff
 Issue: Import statements not in alphabetical order
 File: api/routes/users.py"

Step 3: Explain fix
"I will fix this by reordering the imports in api/routes/users.py
 to be alphabetical. This satisfies ruff's import sorting rules."

Step 4: Apply fix
```bash
# Edit the file to fix imports
# (use Edit tool)

# Re-run hooks
pre-commit run --all-files
```
Output: All checks passed!

Step 5: Proceed
"Hooks now pass. Continuing with commit."
```

### Example: Hook Modified Files

```
Scenario: biome auto-formats files during pre-commit

# Original commit attempt
git commit -m "$(cat <<'EOF'
feat: add new API endpoint
...
EOF
)"

# Hook output:
biome format...........................Modified
Files were modified by this hook. Additional output:

api/routes/new_endpoint.py

# Protocol:
Step 1: Recognize what happened
"The pre-commit hook succeeded but modified files.
 Biome auto-formatted api/routes/new_endpoint.py."

Step 2: Check if safe to amend
```bash
# Check authorship
git log -1 --format='%an %ae'
```
Output: Claude <noreply@anthropic.com>

```bash
# Check not pushed
git status
```
Output: Your branch is ahead of 'origin/main' by 1 commit.

Step 3: Amend the commit
"This is safe to amend - I'm the author and it hasn't been pushed."

```bash
git add api/routes/new_endpoint.py
git commit --amend --no-edit
```

Step 4: Verify
```bash
git log -1 --stat
```
Shows: api/routes/new_endpoint.py with formatting changes included.
```

### When NOT to Amend

```
NEVER amend if:
1. Author is someone else (git log -1 --format='%an %ae')
   → Create new commit instead

2. Commit has been pushed (git status doesn't show "ahead")
   → Create new commit instead

3. User explicitly said not to
   → Create new commit

Always create NEW commit when in doubt.
```

## PR Creation Workflow

### When User Requests PR

**Step 1: Understand Branch State (Parallel)**
```bash
git status                          # Current branch, tracked remote
git diff --staged                   # Staged changes
git diff                            # Unstaged changes
git log main..HEAD --oneline        # Commits since divergence
git diff main...HEAD                # All changes in branch
```

**Step 2: Analyze Changes**
```
Review ALL commits, not just latest:
- What's the complete scope of changes?
- How many files affected?
- What's the overall purpose?
- Any breaking changes?
```

**Step 3: Draft PR Summary**
```markdown
## Summary
- [Bullet 1: High-level change]
- [Bullet 2: Another major change]
- [Bullet 3: Third change if applicable]

## Test plan
- [ ] All unit tests pass
- [ ] Integration tests pass
- [ ] Manually tested [specific scenario]
- [ ] [Other verification steps]

🤖 Generated with [Claude Code](https://claude.com/claude-code)
```

**Step 4: Execute PR Creation (Parallel)**
```bash
# Create branch if needed (only if not on one)
git checkout -b feature/user-auth

# Push with upstream tracking if needed
git push -u origin feature/user-auth

# Create PR using gh CLI with HEREDOC
gh pr create --title "Add user authentication" --body "$(cat <<'EOF'
## Summary
- Add JWT-based authentication system
- Implement login, logout, and token refresh endpoints
- Add auth middleware for protecting routes

## Test plan
- [ ] All unit tests pass (pytest)
- [ ] Integration tests for auth flow pass
- [ ] Manually tested login/logout cycle
- [ ] Verified token expiration handling
- [ ] Tested protected endpoints require auth

🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)"
```

**Step 5: Return PR URL**
```
Report: "PR created: https://github.com/user/repo/pull/123"
```

### Example: Complete PR Creation

```bash
# Step 1: Understand state (parallel)
git status
# On branch feature/auth-endpoints
# Your branch is up to date with 'origin/feature/auth-endpoints'.

git log main..HEAD --oneline
# a1b2c3d Add token refresh endpoint
# d4e5f6g Add logout endpoint
# g7h8i9j Add login endpoint
# j0k1l2m Set up JWT configuration

git diff main...HEAD --stat
# api/auth.py           | 120 ++++++++++++++++
# api/middleware.py     |  45 ++++++
# api/routes/auth.py    |  89 ++++++++++++
# tests/test_auth.py    | 156 ++++++++++++++++++++
# pyproject.toml        |   2 +
# 5 files changed, 412 insertions(+)

# Step 2 & 3: Analyze and draft (internal)
# 4 commits, all related to auth
# New files: auth.py, middleware.py, routes/auth.py, test_auth.py
# Added dependency: pyjwt

# Step 4: Create PR
git push -u origin feature/auth-endpoints  # If not already pushed

gh pr create --title "Add user authentication system" --body "$(cat <<'EOF'
## Summary
- Implement JWT-based authentication
- Add login endpoint (POST /auth/login)
- Add logout endpoint (POST /auth/logout)
- Add token refresh endpoint (POST /auth/refresh)
- Add authentication middleware for protected routes

## Test plan
- [ ] All unit tests pass (156 new tests added)
- [ ] Integration tests cover full auth flow
- [ ] Manually tested login → access protected route → logout
- [ ] Verified token expiration after 1 hour
- [ ] Verified refresh token rotation
- [ ] Tested invalid credentials return 401
- [ ] Tested missing token returns 401
- [ ] Tested expired token returns 401

🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)"

# Output: https://github.com/user/repo/pull/42

# Step 5: Report
"PR created successfully: https://github.com/user/repo/pull/42"
```

## Branch Management

### Default: Work on Main

Decision tree:
1. User specifies branch? → Use that branch
2. User wants isolation? → Create feature branch
3. Otherwise → Work on main for simplicity

### When to Create Branches

```
Create branch if:
- User explicitly requests it
- Large feature that will take multiple sessions
- Experimental work that might be discarded
- Working with a team (shared repo)

Stay on main if:
- Personal project
- Small changes
- Rapid iteration
- User prefers main (default)
```

### Branch Creation

```bash
# Create and switch to new branch
git checkout -b feature/new-feature

# Or using git worktree (for isolation)
git worktree add ../repo-feature feature/new-feature
cd ../repo-feature
# Work here without affecting main repo
```

### Branch Hygiene

```bash
# Before creating branch, ensure clean state
git status

# If uncommitted changes exist
git stash push -m "WIP: before branch"
git checkout -b feature/new-feature
git stash pop

# Delete merged branches
git branch -d feature/old-feature

# Force delete unmerged (only if sure)
git branch -D feature/abandoned
```

## Merge Strategies

### Fast-Forward (Preferred)

```bash
# When possible, use fast-forward
git checkout main
git merge --ff-only feature/auth

# If fast-forward not possible
# Either rebase first or use merge commit
```

### Merge Commit (Teams)

```bash
# Create merge commit (preserves history)
git checkout main
git merge --no-ff feature/auth -m "$(cat <<'EOF'
Merge feature/auth: Add authentication system

Implements JWT-based auth with login/logout/refresh.

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
EOF
)"
```

### Rebase (Clean History)

```bash
# Rebase feature branch onto main
git checkout feature/auth
git rebase main

# If conflicts, resolve and continue
git add <resolved-files>
git rebase --continue

# Then fast-forward merge
git checkout main
git merge --ff-only feature/auth
```

**⚠️ NEVER rebase with -i (interactive)**
```
-i flag requires interactive input, which Claude Code can't provide.
Use non-interactive rebasing only.
```

## Git Safety Rules

### Forbidden Actions

```bash
# NEVER force push to main/master
git push --force origin main        # ❌ NEVER

# NEVER use --no-verify
git commit --no-verify              # ❌ NEVER

# NEVER amend other people's commits
# Check authorship first:
git log -1 --format='%an %ae'
# If not you → don't amend

# NEVER use interactive rebase (requires manual input)
git rebase -i                       # ❌ NEVER
git add -i                          # ❌ NEVER
```

### Safe Actions

```bash
# Safe to force push feature branches (if you own them)
git push --force origin feature/my-branch    # ✅ OK

# Safe to amend your own commits (if not pushed)
git log -1 --format='%an %ae'  # Check: is it you?
git status                      # Check: not pushed yet?
git commit --amend --no-edit    # ✅ OK if both true

# Safe to use non-interactive git
git rebase main                 # ✅ OK (non-interactive)
git add <files>                 # ✅ OK (non-interactive)
```

### Pre-Push Verification

```bash
# Before pushing, verify:
1. Tests pass
git status                      # Clean state?
uv run pytest                   # Tests pass?

2. Hooks passed
git log -1                      # Commit has Claude signature?

3. Right branch
git branch --show-current       # On intended branch?

4. Right remote
git remote -v                   # Pushing to right repo?

# Then push
git push origin feature/my-branch
```

## Change Verification

### After Every Commit

```bash
# Verify commit was created
git status
# Expected: "Your branch is ahead..."

# Show what was committed
git log -1 --stat
# Review: Does this match intent?

# Show the changes
git show HEAD
# Review: Are changes correct?
```

### Before Push

```bash
# See what will be pushed
git log origin/main..HEAD --oneline

# See the diff
git diff origin/main..HEAD

# Verify tests pass
uv run pytest

# Then push
git push
```

### After Merge

```bash
# Verify merge succeeded
git log --oneline --graph -10

# Verify expected files changed
git diff HEAD~1 --stat

# Run full test suite
uv run pytest
```

## Handling Git Conflicts

### When Conflicts Occur

```bash
# Git will report conflicts
git merge feature/auth
# Auto-merging api/routes.py
# CONFLICT (content): Merge conflict in api/routes.py

# Step 1: See what conflicts
git status
# Shows: both modified: api/routes.py

# Step 2: Read the conflicted file
# (use Read tool on api/routes.py)

# Step 3: Resolve conflicts
# (use Edit tool to resolve)
# Remove conflict markers: <<<<<<<, =======, >>>>>>>
# Keep correct version or combine both

# Step 4: Stage resolved files
git add api/routes.py

# Step 5: Complete merge
git merge --continue
# (will open commit message, already provided)

# Step 6: Verify
git log --oneline --graph -5
uv run pytest
```

### Conflict Resolution Strategy

```
1. Read both versions (ours vs theirs)
2. Understand intent of both changes
3. Decide resolution:
   - Keep ours
   - Keep theirs
   - Combine both
   - Write new solution
4. Test the resolution
5. Never leave conflict markers in code
```

## Stash Management

### When to Use Stash

```
Use git stash when:
- Need to switch branches with uncommitted work
- Want to test something on clean state
- Need to pull updates
```

### Stash Workflow

```bash
# Save current work
git stash push -m "WIP: working on auth"

# List stashes
git stash list
# stash@{0}: On main: WIP: working on auth

# Apply stash
git stash pop              # Apply and remove
git stash apply            # Apply and keep

# Show stash contents
git stash show -p stash@{0}

# Drop stash
git stash drop stash@{0}

# Clear all stashes
git stash clear
```

## Git Configuration

### Checking Config

```bash
# See current config
git config --list

# Check specific values
git config user.name
git config user.email
```

### Claude Code NEVER Changes Config

- Don't change `user.name`
- Don't change `user.email`
- Don't add aliases
- Don't modify core settings

If config seems wrong, ask the user to fix it manually.

## Advanced Git Operations

### Cherry-Pick

```bash
# Pick specific commit from another branch
git cherry-pick a1b2c3d

# If conflicts, resolve and continue
git add <resolved-files>
git cherry-pick --continue
```

### Revert

```bash
# Undo a commit by creating inverse commit
git revert a1b2c3d

# Revert without committing (to modify)
git revert --no-commit a1b2c3d
git commit -m "$(cat <<'EOF'
revert: undo feature X

Feature X caused issues in production.
Reverting until fix is ready.

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
EOF
)"
```

### Reset (Careful!)

```bash
# Soft reset (keep changes staged)
git reset --soft HEAD~1

# Mixed reset (keep changes unstaged) [DEFAULT]
git reset HEAD~1

# Hard reset (DISCARD changes)
# Only do this if user explicitly requests
git reset --hard HEAD~1
```

**⚠️ Hard reset warning:**
```
NEVER use --hard without explicit user permission.
Data loss is permanent.

Ask: "This will permanently discard changes. Are you sure?"
```

## GitHub CLI Integration

### Using gh for PRs

```bash
# Create PR
gh pr create --title "..." --body "..."

# View PR
gh pr view 123

# List PRs
gh pr list

# Check PR status
gh pr status

# Merge PR
gh pr merge 123 --squash
gh pr merge 123 --merge
gh pr merge 123 --rebase

# Close PR
gh pr close 123
```

### Using gh for Issues

```bash
# Create issue
gh issue create --title "Bug: ..." --body "..."

# List issues
gh issue list

# View issue
gh issue view 456

# Close issue
gh issue close 456
```

### Using gh for Reviews

```bash
# View PR comments
gh pr view 123 --comments

# Add review comment
gh pr review 123 --comment -b "Looks good!"

# Approve PR
gh pr review 123 --approve

# Request changes
gh pr review 123 --request-changes -b "Please fix X"
```

## Git Status Interpretation

### Clean State

```bash
git status
```
```
On branch main
Your branch is up to date with 'origin/main'.

nothing to commit, working tree clean
```
Meaning: No changes, safe to switch branches

### Uncommitted Changes

```bash
git status
```
```
On branch main
Changes not staged for commit:
  modified:   api/routes.py

Untracked files:
  api/new_file.py
```
Meaning: Have changes, need to commit or stash

### Ahead of Remote

```bash
git status
```
```
On branch main
Your branch is ahead of 'origin/main' by 2 commits.

nothing to commit, working tree clean
```
Meaning: Have local commits, need to push

### Diverged from Remote

```bash
git status
```
```
On branch main
Your branch and 'origin/main' have diverged,
and have 1 and 2 different commits each, respectively.
```
Meaning: Both local and remote have unique commits, need to merge or rebase

## Summary of Git Workflows

### Commit Workflow
1. Gather info (status, diff, log)
2. Analyze changes
3. Draft message
4. Stage files
5. Commit with HEREDOC
6. Handle hooks (NEVER bypass)
7. Verify success

### PR Workflow
1. Understand branch state
2. Analyze ALL commits
3. Draft summary and test plan
4. Push branch
5. Create PR with gh
6. Return URL

### Hook Protocol
1. Read error aloud
2. Identify tool and reason
3. Explain fix
4. Apply fix
5. Re-run until pass

### Safety Rules
- NEVER --no-verify
- NEVER force push to main
- NEVER amend others' commits
- NEVER use -i (interactive)
- ALWAYS verify before push
- ALWAYS fix hooks, not bypass

These workflows ensure quality, safety, and team collaboration while maintaining Claude Code's commitment to never bypass quality gates.
