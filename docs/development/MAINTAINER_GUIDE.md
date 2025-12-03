# Maintainer Guide

Quick reference for Hex maintainers.

## Daily Operations

### Running Tests

```bash
# Quick test
make test-short

# Full test with coverage
make test-coverage

# All checks (recommended before commit)
make verify
```

### Building Locally

```bash
# Current platform
make build

# Test release build
make snapshot

# Full release test
make release
```

### Code Quality

```bash
# Format code
make fmt

# Run linter
make lint

# Static analysis
make vet
```

## Release Process

### 1. Pre-release Checklist

- [ ] All tests pass: `make test`
- [ ] All checks pass: `make verify`
- [ ] CHANGELOG.md updated with changes
- [ ] Version bumped in relevant files
- [ ] Documentation updated if needed
- [ ] Breaking changes documented

### 2. Create Release

```bash
# Update version
VERSION="1.2.3"

# Commit changes
git add .
git commit -m "chore: prepare v${VERSION} release"

# Create and push tag
git tag -a "v${VERSION}" -m "Release v${VERSION}"
git push origin main
git push origin "v${VERSION}"
```

### 3. Monitor Release

1. Go to GitHub Actions → Release workflow
2. Wait for completion (~8-10 minutes)
3. Verify artifacts on Releases page
4. Test installation methods

### 4. Post-release

```bash
# Test install script
curl -sSL https://raw.githubusercontent.com/harper/hex/main/install.sh | bash

# Test Homebrew (after tap updates)
brew upgrade harper/tap/hex

# Test Docker
docker pull ghcr.io/harper/hex:latest
```

## Common Tasks

### Testing Install Script

```bash
# Syntax check
bash -n install.sh

# Dry run with debugging
bash -x install.sh
```

### Testing Workflows Locally

```bash
# Install act (if not installed)
brew install act

# Run test workflow
act -j test

# Run specific job
act -j lint
```

### Updating Dependencies

```bash
# Update all
go get -u ./...
go mod tidy

# Update specific package
go get -u github.com/some/package

# Verify
make test
```

### Docker Operations

```bash
# Build locally
docker build -t hex:local .

# Run locally
docker run -it --rm hex:local --help

# Test multi-platform
docker buildx build --platform linux/amd64,linux/arm64 -t hex:test .
```

## CI/CD Troubleshooting

### Test Failures

**Symptom**: CI tests fail but local tests pass

**Solutions**:
1. Check Go version matches CI (1.24.x)
2. Run with race detector: `go test -race ./...`
3. Check for timing/platform issues
4. Review CI logs for specific errors

### Release Failures

**Symptom**: Release workflow fails

**Solutions**:
1. Test locally: `make snapshot`
2. Check GoReleaser logs
3. Verify tag format: `v*.*.*`
4. Ensure all tests pass

### Homebrew Update Failures

**Symptom**: Homebrew tap doesn't update

**Solutions**:
1. Check `HOMEBREW_TAP_TOKEN` secret exists
2. Verify tap repository exists
3. Check repository dispatch webhook
4. Manual update if needed

## Secrets Management

### Required Secrets

| Secret | Purpose | How to Generate |
|--------|---------|-----------------|
| `HOMEBREW_TAP_TOKEN` | Update Homebrew tap | GitHub PAT with `repo` scope |

### Updating Secrets

1. Go to Settings → Secrets and variables → Actions
2. Update or add secret
3. Trigger workflow to test

## Monitoring

### Key Metrics

- **CI duration**: Should be < 10 minutes
- **Release duration**: Should be < 15 minutes
- **Binary size**: Should be ~20-25MB
- **Test coverage**: Aim for > 70%

### Dashboards

- **GitHub Actions**: Repository → Actions tab
- **Releases**: Repository → Releases
- **Issues**: Repository → Issues
- **PRs**: Repository → Pull requests

## Emergency Procedures

### Rollback Release

```bash
# Delete tag locally and remotely
git tag -d v1.2.3
git push origin :refs/tags/v1.2.3

# Delete GitHub release
# Go to Releases → Edit → Delete release

# Restore previous version tag (if needed)
git tag -a v1.2.2 -m "Restore v1.2.2"
git push origin v1.2.2
```

### Hotfix Process

```bash
# Create hotfix from tag
git checkout -b hotfix/issue-123 v1.2.3

# Make fix
# ... edit files ...

# Commit
git commit -am "fix: critical issue"

# Tag new version
git tag -a v1.2.4 -m "Hotfix: critical issue"

# Push
git push origin hotfix/issue-123
git push origin v1.2.4
```

## Best Practices

### Commits

- Use conventional commits: `feat:`, `fix:`, `docs:`, etc.
- Write descriptive messages
- Reference issues: `Fixes #123`

### PRs

- Fill out PR template completely
- Ensure CI passes before merge
- Get review for significant changes
- Squash merge to keep history clean

### Releases

- Follow semantic versioning
- Document breaking changes
- Test installation methods
- Announce in community channels

### Code Review

- Check for test coverage
- Review error handling
- Verify documentation updates
- Test locally if complex

## Resources

- **CI/CD Docs**: [docs/CI_CD.md](docs/CI_CD.md)
- **Contributing**: [.github/CONTRIBUTING.md](.github/CONTRIBUTING.md)
- **Release Checklist**: [.github/RELEASE_CHECKLIST.md](.github/RELEASE_CHECKLIST.md)
- **Architecture**: [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)

## Quick Reference

### Makefile Targets

```bash
make help              # Show all targets
make build             # Build binary
make test              # Run tests
make test-coverage     # Coverage report
make lint              # Run linter
make fmt               # Format code
make vet               # Static analysis
make verify            # All checks
make release           # Test release build
make snapshot          # Build snapshot
make clean             # Clean artifacts
```

### Git Workflow

```bash
# Feature branch
git checkout -b feature/name
# ... make changes ...
git commit -am "feat: description"
git push origin feature/name
# ... create PR ...

# Release
git checkout main
git pull
# ... update CHANGELOG ...
git tag -a v1.2.3 -m "Release v1.2.3"
git push origin main v1.2.3
```

### Testing Commands

```bash
# Unit tests
go test ./...

# With race detection
go test -race ./...

# With coverage
go test -cover ./...

# Specific package
go test ./internal/core/...

# Verbose
go test -v ./...
```

## Support

For questions or issues:
1. Check documentation first
2. Search existing issues
3. Ask in discussions
4. Create new issue if needed

---

**Last Updated**: 2025-11-28
**Maintainer**: Harper
