---
name: commit
description: Review changes and create a well-crafted commit
args:
  message: Custom commit message (optional)
---

# Commit Changes Workflow

You are preparing to commit the current changes{{if .message}} with message: "{{.message}}"{{end}}.

## Pre-Commit Review

### 1. Verify Current State

Run these commands to understand what will be committed:

```bash
git status          # See modified, staged, and untracked files
git diff            # See unstaged changes
git diff --staged   # See staged changes
```

### 2. Pre-Commit Checklist

Verify before committing:

**Code Quality**
- [ ] All tests pass locally
- [ ] No debug code (console.log, print statements, etc.)
- [ ] No commented-out code
- [ ] No TODO/FIXME comments (unless intentional)
- [ ] Code follows project style

**Testing**
- [ ] New code has tests
- [ ] Tests cover edge cases
- [ ] All tests are passing
- [ ] No skipped or disabled tests (unless documented)

**Documentation**
- [ ] Code comments are accurate
- [ ] Public APIs documented
- [ ] README updated if needed
- [ ] CHANGELOG updated (if applicable)

**Safety**
- [ ] No secrets or credentials
- [ ] No hardcoded passwords or API keys
- [ ] No sensitive data
- [ ] .gitignore updated for new generated files

**Dependencies**
- [ ] New dependencies justified and documented
- [ ] Lock files updated
- [ ] No unnecessary dependencies added

### 3. Review Log for Context

Check recent commits to match style:

```bash
git log --oneline -10
```

Note the commit message format used in the project.

## Commit Message Guidelines

### Format

Follow conventional commits format:

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Type** (required):
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation only
- `style`: Formatting, missing semicolons, etc.
- `refactor`: Code change that neither fixes a bug nor adds a feature
- `perf`: Performance improvement
- `test`: Adding or updating tests
- `chore`: Maintenance tasks, dependency updates

**Scope** (optional): Component or area affected

**Subject** (required):
- Short summary (50 chars or less)
- Imperative mood ("add" not "added" or "adds")
- No period at the end
- Lowercase first letter

**Body** (optional):
- Explain what and why (not how)
- Wrap at 72 characters
- Separate from subject with blank line

**Footer** (optional):
- Breaking changes: `BREAKING CHANGE: description`
- Issue references: `Fixes #123`, `Closes #456`

### Examples

**Simple commit:**
```
fix: prevent racing of requests

Introduce a request id and a reference to latest request. Dismiss
incoming responses other than from latest request.
```

**With scope:**
```
feat(auth): add OAuth2 support

Implement OAuth2 authentication flow for third-party providers.
Supports Google, GitHub, and Microsoft providers.
```

**Breaking change:**
```
refactor(api)!: drop support for Node 12

BREAKING CHANGE: Node 12 is no longer supported. Minimum version is now Node 14.

Refs #234
```

## Staging Changes

If not already staged, stage the files to commit:

```bash
# Stage specific files
git add path/to/file1.go path/to/file2.go

# Stage all modified files
git add -u

# Stage all files (including untracked)
git add .
```

**Be selective**: Only commit related changes together.

## Creating the Commit

{{if .message}}
### Using Provided Message

Create commit with the provided message:
```bash
git commit -m "{{.message}}"
```
{{else}}
### Crafting the Message

Based on the changes, create an appropriate commit message following the guidelines above.

```bash
git commit
```

This will open an editor for a detailed message, or use `-m` for a short one:

```bash
git commit -m "type(scope): short description"
```
{{end}}

## Post-Commit Verification

After committing:

1. **Verify Commit**
   ```bash
   git show              # See the commit you just made
   git log -1 --stat     # See commit with file stats
   ```

2. **Check Status**
   ```bash
   git status           # Should show clean working tree
   ```

3. **Test Again** (if critical)
   ```bash
   # Run tests to verify committed state works
   go test ./...
   ```

## Common Issues

**Forgot to add a file?**
```bash
git add forgotten-file.go
git commit --amend --no-edit
```

**Typo in commit message?**
```bash
git commit --amend
# Edit message in editor
```

**Need to uncommit?**
```bash
git reset --soft HEAD~1   # Keep changes staged
git reset HEAD~1          # Keep changes unstaged
```

## Output Format

Provide a summary:

```markdown
## Commit Summary

### Changes Included
- [List of modified files]
- [Brief description of changes]

### Commit Message
```
[Your crafted commit message]
```

### Verification
- [ ] Pre-commit checklist complete
- [ ] Tests passing
- [ ] Message follows format
- [ ] Ready to commit

### Command to Run
```bash
git commit -m "your message here"
```
```

## Important Notes

- **Never use `--no-verify`** - Pre-commit hooks are there to help
- **Commit related changes together** - Each commit should be atomic
- **Write for humans** - Someone will read this in 6 months
- **Be honest** - Commit message should match actual changes

Review the changes and prepare the commit now.
