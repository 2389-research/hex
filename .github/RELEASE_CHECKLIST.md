# Release Checklist

Checklist for preparing and publishing Hex releases.

## Pre-Release Checklist

### Code Quality

- [ ] All tests pass (`make test`)
- [ ] Integration tests pass (`go test ./test/integration/...`)
- [ ] No compiler warnings
- [ ] Code coverage meets minimum threshold (>80%)
- [ ] All linters pass (if configured)
- [ ] No TODO comments for critical items

### Documentation

- [ ] README.md updated with new features
- [ ] CHANGELOG.md updated with version and date
- [ ] Release notes written (RELEASE_NOTES.md)
- [ ] User guide updated for new features
- [ ] Architecture docs reflect current design
- [ ] API documentation current (if applicable)
- [ ] All code examples tested

### Version Management

- [ ] Version number decided (semver: MAJOR.MINOR.PATCH)
- [ ] Version updated in relevant files
- [ ] git tag ready (format: `vX.Y.Z`)
- [ ] go.mod version matches

### Testing

- [ ] Manual testing on macOS
- [ ] Manual testing on Linux (if available)
- [ ] Manual testing on Windows (if available)
- [ ] Fresh install test (`go install github.com/2389-research/hex/cmd/hex@latest`)
- [ ] Config migration test (if schema changed)
- [ ] Database migration test (if schema changed)
- [ ] Upgrade test from previous version

### Features Verification

Core functionality:

- [ ] Interactive mode launches successfully
- [ ] Streaming responses work
- [ ] All tools execute correctly
- [ ] Tool approval prompts appear
- [ ] Conversation persistence works
- [ ] `--continue` flag works
- [ ] `--resume` flag works
- [ ] Print mode works
- [ ] TUI renders correctly

## Release Process

### 1. Final Commit

```bash
# Ensure all changes committed
git status

# Run tests one final time
make test
go test ./...
```

### 2. Create Git Tag

```bash
# Create annotated tag
git tag -a vX.Y.Z -m "Release vX.Y.Z: description"

# Verify tag
git tag -l
git show vX.Y.Z
```

### 3. Push Tag

```bash
# Push tag to remote
git push origin vX.Y.Z

# Verify on GitHub
# Visit: https://github.com/2389-research/hex/tags
```

### 4. Create GitHub Release

On GitHub:

1. Go to Releases page
2. Click "Draft a new release"
3. Select tag: `vX.Y.Z`
4. Title: "Hex vX.Y.Z - Description"
5. Copy content from RELEASE_NOTES.md
6. Check "Set as latest release"
7. Publish release

### 5. Verify Installation

```bash
# Clean install test
go install github.com/2389-research/hex/cmd/hex@latest

# Verify version
hex --version

# Quick functionality test
hex doctor
hex --print "Hello"
```

### 6. Update Documentation Site (if applicable)

- [ ] Update website with new version
- [ ] Update getting started guide
- [ ] Update API reference
- [ ] Publish blog post (optional)

## Post-Release Checklist

### Announcement

- [ ] Update README badge (if applicable)
- [ ] Announce on social media (optional)
- [ ] Update project homepage
- [ ] Notify users on mailing list (if exists)

### Monitoring

- [ ] Monitor GitHub issues for bug reports
- [ ] Monitor installation stats
- [ ] Check for common user questions
- [ ] Watch for CI/CD failures on user machines

### Next Version Prep

- [ ] Create next milestone (e.g., v0.3.0)
- [ ] Update ROADMAP.md with next goals
- [ ] File issues for known bugs/improvements
- [ ] Plan next sprint

## Hotfix Process (if needed)

If critical bug found after release:

### 1. Create Hotfix Branch

```bash
git checkout -b hotfix/vX.Y.Z vX.Y.Z
```

### 2. Fix Bug

```bash
# Make minimal fix
# Add regression test
# Update CHANGELOG.md
git commit -m "fix: critical bug description"
```

### 3. Tag Hotfix

```bash
git tag -a vX.Y.Z -m "Hotfix vX.Y.Z: Critical bug fix"
git push origin hotfix/vX.Y.Z
git push origin vX.Y.Z
```

### 4. Merge Back

```bash
git checkout main
git merge hotfix/vX.Y.Z
git push origin main
```

## Version Numbering Guide

Following [Semantic Versioning 2.0.0](https://semver.org/):

### MAJOR version (X.0.0)
- Incompatible API changes
- Breaking changes to CLI interface
- Database schema changes requiring manual migration

### MINOR version (0.X.0)
- New features (backward compatible)
- New tools
- New commands
- Performance improvements

### PATCH version (0.0.X)
- Bug fixes
- Documentation updates
- Minor improvements

## Release Schedule

Suggested cadence:

- **Major releases**: Every 6-12 months
- **Minor releases**: Every 4-8 weeks
- **Patch releases**: As needed for critical bugs

## Communication Channels

Where to announce releases:

- [ ] GitHub Releases page
- [ ] Project README
- [ ] Documentation site
- [ ] Twitter/X (if applicable)
- [ ] Reddit (if applicable)
- [ ] Discord/Slack (if applicable)

## Rollback Plan

If release has critical issues:

### 1. Immediate Actions

```bash
# Yank release on GitHub (mark as pre-release or draft)
# Update release notes with warning
# Pin users to previous version in README
```

### 2. Communicate

- Post issue on GitHub
- Update release notes
- Notify users via all channels
- Provide workaround if possible

### 3. Fix Forward

- Create hotfix branch
- Fix critical issue
- Release patch version
- Update all documentation

## Checklist Archive

### v1.0.0 (2026-03-12)

Hex v1.0.0 release:
- ✅ Interactive mode with TUI
- ✅ Full tool system (Read, Write, Edit, Bash, Grep, Glob)
- ✅ Storage layer with SQLite
- ✅ Streaming API support
- ✅ GoReleaser integration
- ✅ Comprehensive documentation
- ✅ All tests passing

---

**Last Updated**: 2026-03-12
**Next Review**: Before next release
