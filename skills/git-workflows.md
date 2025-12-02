---
name: git-workflows
description: Git best practices for commits, branches, and pull requests
tags:
  - git
  - version-control
  - workflow
  - commits
activationPatterns:
  - "git.*commit"
  - "create.*branch"
  - "merge.*pr"
  - "pull request"
priority: 6
version: 1.0.0
---

# Git Workflows Best Practices

## Commit Guidelines

### Atomic Commits

Each commit should be a single logical change:

✅ **Good - One Thing**:
```bash
git add internal/auth/handler.go internal/auth/handler_test.go
git commit -m "feat: add password reset endpoint"
```

❌ **Bad - Multiple Things**:
```bash
git add internal/auth/ internal/database/ cmd/clem/
git commit -m "various fixes and improvements"
```

### Commit Messages

Follow conventional commit format:

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types**:
- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation only
- `test:` - Adding or updating tests
- `refactor:` - Code change that neither fixes a bug nor adds a feature
- `perf:` - Performance improvement
- `chore:` - Maintenance tasks (dependencies, build)

**Examples**:

```bash
# Feature
git commit -m "feat(auth): add two-factor authentication support"

# Bug fix
git commit -m "fix(api): handle null user in session validation"

# Documentation
git commit -m "docs: add API usage examples to README"

# Refactoring
git commit -m "refactor(db): extract connection pooling to separate package"

# Breaking change
git commit -m "feat(api): change user endpoint response format

BREAKING CHANGE: /api/users now returns {users: [...]} instead of [...]"
```

### What Makes a Good Commit Message?

✅ **Good Messages**:
```
fix: prevent race condition in session cleanup

The session cleanup goroutine was accessing the sessions map
without holding the mutex, causing occasional panics under load.
Added proper locking around all map access.
```

❌ **Bad Messages**:
```
fix stuff
wip
updates
fix bug
```

**Rules**:
- Use imperative mood ("add feature" not "added feature")
- First line max 50 characters
- Body wraps at 72 characters
- Separate subject from body with blank line
- Explain what and why, not how

## Branching Strategy

### Branch Naming

```
<type>/<short-description>
```

**Examples**:
- `feature/user-authentication`
- `fix/database-connection-leak`
- `docs/api-documentation`
- `refactor/error-handling`

### Main Branch Protection

The `main` branch should always be deployable:

```bash
# ❌ Never commit directly to main
git checkout main
git commit -m "quick fix"

# ✅ Always use feature branches
git checkout -b fix/typo-in-readme
git commit -m "docs: fix typo in installation instructions"
git push origin fix/typo-in-readme
# Then create pull request
```

### Feature Branch Workflow

1. **Create branch from main**:
```bash
git checkout main
git pull origin main
git checkout -b feature/new-feature
```

2. **Make commits**:
```bash
# Make changes
git add .
git commit -m "feat: implement new feature"
```

3. **Keep branch updated**:
```bash
# Regularly sync with main
git checkout main
git pull origin main
git checkout feature/new-feature
git rebase main  # Or merge if preferred
```

4. **Push and create PR**:
```bash
git push origin feature/new-feature
# Create pull request via GitHub/GitLab
```

5. **After PR merged, cleanup**:
```bash
git checkout main
git pull origin main
git branch -d feature/new-feature
```

## Pull Request Best Practices

### PR Title

Should be descriptive and follow same format as commits:

```
feat(auth): add OAuth2 provider support
fix(api): resolve timeout in user listing endpoint
docs: add deployment guide
```

### PR Description

Use a template:

```markdown
## Summary
Brief description of what this PR does

## Changes
- Bullet point list of changes
- Each major change gets a bullet
- Focus on what, not how

## Testing
How this was tested:
- [ ] Unit tests added/updated
- [ ] Integration tests pass
- [ ] Manually tested in dev environment
- [ ] Tested edge cases: X, Y, Z

## Breaking Changes
List any breaking changes and migration steps

## Related Issues
Fixes #123
Related to #456
```

### PR Size

Keep PRs small and focused:

✅ **Good**: 100-300 lines, single feature
❌ **Bad**: 2000+ lines, multiple features

If PR is large, consider splitting:
```bash
# Instead of one large PR
feature/complete-auth-system

# Split into smaller PRs
feature/auth-database-schema
feature/auth-api-endpoints
feature/auth-ui-components
```

### PR Readiness Checklist

Before creating PR:

- [ ] All tests pass locally
- [ ] Code follows project style guidelines
- [ ] Documentation updated
- [ ] No debug code or comments left
- [ ] Ran linter/formatter
- [ ] Rebased on latest main
- [ ] Commit history is clean

## Common Git Commands

### Checking Status

```bash
# See what's changed
git status

# See diff of changes
git diff

# See diff of staged changes
git diff --staged
```

### Staging Changes

```bash
# Stage specific files
git add file1.go file2.go

# Stage all changes
git add .

# Stage parts of a file interactively
git add -p file.go

# Unstage files
git reset HEAD file.go
```

### Committing

```bash
# Commit staged changes
git commit -m "feat: add new feature"

# Commit with detailed message (opens editor)
git commit

# Amend last commit (add forgotten changes)
git add forgotten_file.go
git commit --amend --no-edit

# Amend and change message
git commit --amend
```

### Branching

```bash
# Create new branch
git checkout -b feature/new-feature

# Switch branches
git checkout main

# List branches
git branch -a

# Delete local branch
git branch -d feature/old-feature

# Delete remote branch
git push origin --delete feature/old-feature
```

### Syncing

```bash
# Fetch latest from remote
git fetch origin

# Pull latest changes
git pull origin main

# Push changes
git push origin feature/my-feature

# Force push (use with caution!)
git push --force-with-lease origin feature/my-feature
```

### Undoing Changes

```bash
# Discard changes in working directory
git checkout -- file.go

# Unstage file (keep changes)
git reset HEAD file.go

# Undo last commit (keep changes)
git reset --soft HEAD~1

# Undo last commit (discard changes) - DANGEROUS
git reset --hard HEAD~1

# Revert a commit (creates new commit)
git revert abc123
```

### Viewing History

```bash
# Show commit log
git log

# One line per commit
git log --oneline

# Show specific file history
git log -- file.go

# Show changes in commit
git show abc123
```

## Advanced Workflows

### Rebasing

Keep commit history clean:

```bash
# Rebase feature branch on main
git checkout feature/my-feature
git rebase main

# Interactive rebase (squash commits)
git rebase -i HEAD~3
```

### Cherry-picking

Apply specific commit to another branch:

```bash
git checkout target-branch
git cherry-pick abc123
```

### Stashing

Save work in progress:

```bash
# Stash changes
git stash

# List stashes
git stash list

# Apply most recent stash
git stash pop

# Apply specific stash
git stash apply stash@{1}
```

## Merge vs Rebase

### Use Merge When:
- Preserving exact history is important
- Working on public/shared branches
- Team prefers merge commits

```bash
git checkout main
git merge feature/my-feature
```

### Use Rebase When:
- Want clean linear history
- Working on private feature branch
- Before creating pull request

```bash
git checkout feature/my-feature
git rebase main
```

## Common Mistakes to Avoid

### ❌ Committing Secrets

```bash
# Never commit
- API keys
- Passwords
- Private keys
- .env files
```

Add to `.gitignore`:
```
.env
*.key
secrets/
config/local.yaml
```

### ❌ Large Binary Files

Git is not designed for large binaries. Use Git LFS or S3.

### ❌ Fixing Public History

```bash
# ❌ Don't rebase or force-push public branches
git push --force origin main  # NEVER

# ✅ Use revert for public branches
git revert abc123
```

### ❌ Committing Debug Code

```go
// ❌ Remove before committing
fmt.Println("DEBUG: here")
log.Fatal("TODO: implement this")
```

### ❌ Vague Commit Messages

```bash
# ❌ Bad
git commit -m "fix"
git commit -m "updates"
git commit -m "wip"

# ✅ Good
git commit -m "fix: prevent nil pointer in user handler"
```

## Collaboration Tips

### Before Starting Work

```bash
# Always sync first
git checkout main
git pull origin main
git checkout -b feature/new-work
```

### Before Pushing

```bash
# Make sure tests pass
go test ./...

# Lint code
golangci-lint run

# Rebase on latest main
git fetch origin
git rebase origin/main
```

### Resolving Conflicts

```bash
# During rebase
git rebase main
# CONFLICT...

# Fix conflicts in files
# Then:
git add .
git rebase --continue

# Or abort
git rebase --abort
```

## Summary

Good git workflow:
- **Commit atomically** - one logical change per commit
- **Write clear messages** - explain what and why
- **Branch frequently** - keep main stable
- **Review before merge** - use pull requests
- **Keep history clean** - rebase private branches
- **Sync regularly** - avoid large divergence

**Remember**: Git is a tool for collaboration. Clear history and good messages help your future self and teammates.
