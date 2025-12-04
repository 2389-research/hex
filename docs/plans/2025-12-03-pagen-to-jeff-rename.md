# Pagen to Jeff Rename - Full Refactor

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Completely rename the project from "pagen"/"pagent" to "jeff" with zero remaining references

**Architecture:** Systematic file, directory, and content replacement across 383 occurrences

**Tech Stack:** Go 1.24, sed, git mv

**Scope:**
- Module path: `github.com/harper/pagent` → `github.com/harper/jeff`
- Binary name: `pagent` → `jeff`
- Directory: `cmd/pagent` → `cmd/jeff`
- Config directory: `~/.jeff` → `~/.jeff`
- Environment variables: `PAGEN_*` → `JEFF_*`
- All code, docs, comments, user-facing strings

---

## Task 1: Rename Directory Structure

**Files:**
- Rename: `cmd/pagent` → `cmd/jeff`
- Rename: `pagent` (binary) → `jeff`
- Rename: `pagent-architecture.dot` → `jeff-architecture.dot`
- Rename: `pagent-architecture.png` → `jeff-architecture.png`
- Rename: `pagent-architecture.svg` → `jeff-architecture.svg`
- Rename: `docs/plans/2025-12-01-pagen-productivity-agent-design.md` → `docs/plans/2025-12-01-jeff-productivity-agent-design.md`

**Step 1: Rename cmd directory**

```bash
git mv cmd/pagent cmd/jeff
```

**Step 2: Rename binary and architecture files**

```bash
git mv pagent jeff 2>/dev/null || true
git mv pagent-architecture.dot jeff-architecture.dot
git mv pagent-architecture.png jeff-architecture.png
git mv pagent-architecture.svg jeff-architecture.svg
```

**Step 3: Rename design doc**

```bash
git mv docs/plans/2025-12-01-pagen-productivity-agent-design.md docs/plans/2025-12-01-jeff-productivity-agent-design.md
```

**Step 4: Verify renames**

Run: `git status`
Expected: Shows renamed files, no deletions

**Step 5: Commit directory renames**

```bash
git commit -m "refactor: rename directories and files from pagen/pagent to jeff"
```

---

## Task 2: Update Go Module Path

**Files:**
- Modify: `go.mod:1`
- Modify: All `*.go` files with import statements

**Step 1: Update go.mod module declaration**

Replace in `go.mod`:
```go
module github.com/harper/pagent
```

With:
```go
module github.com/harper/jeff
```

**Step 2: Update all import paths in Go files**

Run:
```bash
find . -name "*.go" -type f -exec sed -i '' 's|github.com/harper/pagent|github.com/harper/jeff|g' {} +
```

**Step 3: Verify no old imports remain**

Run: `grep -r "github.com/harper/pagent" --include="*.go" .`
Expected: No output (all imports updated)

**Step 4: Run go mod tidy**

Run: `go mod tidy`
Expected: No errors, go.mod and go.sum updated

**Step 5: Verify compilation**

Run: `go build ./...`
Expected: Successful compilation with no errors

**Step 6: Commit module path changes**

```bash
git add go.mod go.sum
find . -name "*.go" -exec git add {} +
git commit -m "refactor: update module path from pagent to jeff"
```

---

## Task 3: Update Binary Name in Code

**Files:**
- Modify: `cmd/jeff/root.go`
- Modify: All files in `cmd/jeff/*.go` with "pagent" in strings

**Step 1: Update root command Use field**

Find and replace in `cmd/jeff/root.go`:
```go
Use:     "pagent [prompt]",
```

With:
```go
Use:     "jeff [prompt]",
```

**Step 2: Update all command examples in cmd/jeff**

Run:
```bash
find cmd/jeff -name "*.go" -exec sed -i '' 's/pagent /jeff /g' {} +
find cmd/jeff -name "*.go" -exec sed -i '' "s/pagent'/jeff'/g" {} +
find cmd/jeff -name "*.go" -exec sed -i '' 's/pagent"/jeff"/g' {} +
```

**Step 3: Verify changes**

Run: `grep -r "pagent" cmd/jeff/`
Expected: No matches (all updated to "jeff")

**Step 4: Verify compilation**

Run: `go build ./cmd/jeff`
Expected: Successful build, creates `jeff` binary

**Step 5: Commit binary name changes**

```bash
git add cmd/jeff/
git commit -m "refactor: update binary name from pagent to jeff in code"
```

---

## Task 4: Update Configuration Paths

**Files:**
- Modify: `cmd/jeff/provider.go` (multiple occurrences of `~/.jeff`)
- Modify: `cmd/jeff/root.go` (database path references)
- Modify: Any other files with `~/.jeff` or `.jeff`

**Step 1: Update ~/.jeff to ~/.jeff in all files**

Run:
```bash
find . -name "*.go" -exec sed -i '' 's|~/.jeff|~/.jeff|g' {} +
find . -name "*.go" -exec sed -i '' 's|\.jeff|.jeff|g' {} +
find . -name "*.go" -exec sed -i '' 's|/pagen/|/jeff/|g' {} +
```

**Step 2: Update .jeff directory references in other file types**

Run:
```bash
find . -name "*.md" -exec sed -i '' 's|~/.jeff|~/.jeff|g' {} +
find . -name "*.yaml" -exec sed -i '' 's|~/.jeff|~/.jeff|g' {} +
find . -name "*.md" -exec sed -i '' 's|\.jeff|.jeff|g' {} +
```

**Step 3: Verify no .jeff references remain**

Run: `grep -r "\.jeff" --include="*.go" --include="*.md" --include="*.yaml" .`
Expected: No matches (all updated to .jeff)

**Step 4: Verify compilation**

Run: `go build ./...`
Expected: Successful compilation

**Step 5: Commit config path changes**

```bash
git add .
git commit -m "refactor: update config paths from ~/.jeff to ~/.jeff"
```

---

## Task 5: Update Environment Variable Names

**Files:**
- Modify: `cmd/jeff/provider.go` (PAGEN_GMAIL_CLIENT_ID, PAGEN_GMAIL_CLIENT_SECRET)
- Modify: Any other files with PAGEN_ prefix

**Step 1: Update PAGEN_ to JEFF_ in all files**

Run:
```bash
find . -name "*.go" -exec sed -i '' 's/PAGEN_/JEFF_/g' {} +
find . -name "*.md" -exec sed -i '' 's/PAGEN_/JEFF_/g' {} +
```

**Step 2: Verify no PAGEN_ references remain**

Run: `grep -r "PAGEN_" --include="*.go" --include="*.md" .`
Expected: No matches (all updated to JEFF_)

**Step 3: Verify compilation**

Run: `go build ./...`
Expected: Successful compilation

**Step 4: Commit environment variable changes**

```bash
git add .
git commit -m "refactor: update environment variables from PAGEN_ to JEFF_"
```

---

## Task 6: Update Documentation and Comments

**Files:**
- Modify: All `*.md` files in `docs/`
- Modify: All ABOUTME comments in `*.go` files
- Modify: All other comments and strings

**Step 1: Update "Pagen" to "Jeff" in markdown files (case-sensitive)**

Run:
```bash
find docs -name "*.md" -exec sed -i '' 's/Pagen/Jeff/g' {} +
find docs -name "*.md" -exec sed -i '' 's/pagen/jeff/g' {} +
find . -name "README.md" -exec sed -i '' 's/Pagen/Jeff/g' {} +
find . -name "README.md" -exec sed -i '' 's/pagen/jeff/g' {} +
```

**Step 2: Update "pagen" and "pagent" in all Go comments**

Run:
```bash
find . -name "*.go" -exec sed -i '' 's/pagen/jeff/g' {} +
find . -name "*.go" -exec sed -i '' 's/Pagen/Jeff/g' {} +
```

Note: "pagent" → "jeff" already done in previous tasks

**Step 3: Verify no pagen/pagent references remain**

Run: `grep -ri "pagen\|pagent" --include="*.go" --include="*.md" . | grep -v "Phase 1\|Task [0-9]"`
Expected: No matches except in this plan file and git history

**Step 4: Verify compilation and tests**

Run:
```bash
go build ./...
go test -short ./... -v
```

Expected: All builds and tests pass

**Step 5: Commit documentation changes**

```bash
git add .
git commit -m "docs: update all references from pagen/pagent to jeff"
```

---

## Task 7: Update This Plan File

**Files:**
- Modify: `docs/plans/2025-12-03-pagen-to-jeff-rename.md` (this file)

**Step 1: Update header and content to use past tense**

Replace header "Pagen to Jeff Rename" with "Jeff Rename Complete"

Update goal to: "Completed full rename from pagen/pagent to jeff with zero remaining references"

**Step 2: Commit plan update**

```bash
git add docs/plans/2025-12-03-pagen-to-jeff-rename.md
git commit -m "docs: mark rename plan as complete"
```

---

## Task 8: Final Verification and Build

**Files:**
- Build: `jeff` binary
- Test: All packages

**Step 1: Clean and rebuild from scratch**

Run:
```bash
go clean -cache
go clean -modcache
go mod download
go build -o jeff ./cmd/jeff
```

Expected: Successful build, creates `jeff` binary

**Step 2: Run full test suite**

Run: `go test ./... -v`
Expected: All tests pass (skip slow integration tests with -short if needed)

**Step 3: Verify binary works**

Run: `./jeff --version`
Expected: Shows version information

**Step 4: Test basic functionality**

Run: `./jeff --help`
Expected: Shows help with "jeff" in usage, not "pagent"

**Step 5: Search for any remaining references**

Run: `grep -ri "pagen\|pagent" --include="*.go" --include="*.md" --include="*.yaml" . | grep -v ".git" | grep -v "docs/plans/2025-12-03"`
Expected: No matches (clean!)

**Step 6: Commit any final changes**

```bash
git add .
git commit -m "build: final verification and build of jeff binary"
```

---

## Task 9: Update Git Repository Metadata (if needed)

**Note:** This task is optional and depends on whether you want to rename the git repository itself.

**Step 1: Add tag for the rename**

```bash
git tag -a v0.1.0-jeff -m "Complete rename from pagen to jeff"
```

**Step 2: Push all changes**

```bash
git push origin main
git push origin --tags
```

**Step 3: Update GitHub repository name (manual)**

If you want to rename the GitHub repository:
1. Go to repository Settings
2. Rename from `pagen-agent` to `jeff`
3. Update local remote: `git remote set-url origin git@github.com:harper/jeff.git`

---

## Success Criteria

- ✅ Zero occurrences of "pagen" or "pagent" in code/docs (except this plan and git history)
- ✅ Module path is `github.com/harper/jeff`
- ✅ Binary is named `jeff`
- ✅ Config directory is `~/.jeff`
- ✅ Environment variables use `JEFF_` prefix
- ✅ All tests pass
- ✅ Binary builds and runs successfully
- ✅ Help text shows "jeff" not "pagent"

---

## Rollback Plan

If something breaks:

```bash
git reset --hard HEAD~9  # Reset to before rename
git clean -fd
```

Then investigate what broke and fix before retrying.
