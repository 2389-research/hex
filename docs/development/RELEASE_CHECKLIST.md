# Release Checklist for v1.0.0

## Pre-Release (1 week before)

### Code Quality
- [ ] All tests passing (`go test ./...`)
- [ ] No race conditions (`go test -race ./...`)
- [ ] Linters clean (`golangci-lint run`)
- [ ] `govulncheck` clean (no known vulnerabilities)
- [ ] Test coverage ≥ 80% (`go test -cover ./...`)
- [ ] Integration tests passing
- [ ] Pre-commit hooks working

### Documentation
- [ ] README.md updated with v1.0.0 features
- [ ] CHANGELOG.md has v1.0.0 entry with release date
- [ ] All docs/ files reviewed for accuracy
- [ ] USER_GUIDE.md complete
- [ ] ARCHITECTURE.md up to date
- [ ] TOOLS.md includes all tools
- [ ] API documentation complete
- [ ] Migration guide (if needed from v0.x)

### Security
- [ ] Security audit complete
- [ ] XSS vulnerability tests passing
- [ ] No secrets in code
- [ ] Dependencies updated
- [ ] MCP tool approval system tested
- [ ] Input validation verified
- [ ] File path sanitization tested

### Performance
- [ ] Benchmarks run
- [ ] No performance regressions from v0.5.0
- [ ] Startup time < 100ms
- [ ] Memory usage acceptable (< 50MB idle)
- [ ] Database queries optimized
- [ ] Token counting accurate

### Features
- [ ] All Phase 1-6 features complete
- [ ] No critical bugs
- [ ] Known issues documented in CHANGELOG
- [ ] Feature flags stable
- [ ] All 13 built-in tools working
- [ ] MCP integration stable
- [ ] Vision/multimodal support working

### Build System
- [ ] `.goreleaser.yml` configuration verified
- [ ] GitHub Actions workflows tested
- [ ] Homebrew tap automation working
- [ ] Docker builds successful
- [ ] All platforms build (linux, darwin, windows)
- [ ] All architectures build (amd64, arm64)

---

## Release Day

### Version Bump
- [ ] Update version in `cmd/hex/root.go` (line 24)
  ```go
  version = "1.0.0"
  ```
- [ ] Update README.md version badge (line 10)
  ```markdown
  **Latest Version**: v1.0.0
  ```
- [ ] Update CHANGELOG.md with release date
- [ ] Verify no other hardcoded versions

### Create Release Branch
```bash
git checkout -b release/v1.0.0
git add cmd/hex/root.go README.md CHANGELOG.md
git commit -m "chore: bump version to v1.0.0"
```

### Final Verification
- [ ] Run full test suite one more time
  ```bash
  go test ./... -race -cover
  ```
- [ ] Build locally for all platforms
  ```bash
  goreleaser build --snapshot --clean
  ```
- [ ] Test binary manually
  ```bash
  ./dist/hex_darwin_amd64_v1/hex --version
  ./dist/hex_darwin_amd64_v1/hex --help
  ```
- [ ] Verify installation scripts work
  ```bash
  ./install.sh  # Test locally with local file
  ```

### Tag and Push
```bash
# Create annotated tag
git tag -a v1.0.0 -m "Release v1.0.0

Hex v1.0.0 is the first production-ready release.

Features:
- Interactive TUI with Bubbletea
- SSE streaming responses
- SQLite conversation storage
- 13 built-in tools
- MCP integration
- Vision/multimodal support
- 80%+ test coverage
- Security audited
- Performance optimized"

# Push release branch and tag
git push origin release/v1.0.0
git push origin v1.0.0
```

### Monitor CI/CD
- [ ] GitHub Actions workflow triggers
  - Visit: https://github.com/2389-research/hex/actions
- [ ] All tests pass in CI
- [ ] GoReleaser builds all artifacts
- [ ] Homebrew tap updates successfully
- [ ] Docker images pushed to GHCR
- [ ] GitHub Release created automatically

### Verify Artifacts
- [ ] Binary archives downloadable from GitHub Releases
- [ ] Checksums.txt present and correct
  ```bash
  curl -sL https://github.com/2389-research/hex/releases/download/v1.0.0/checksums.txt
  ```
- [ ] Linux packages available (.deb, .rpm, .apk)
- [ ] Docker image pullable
  ```bash
  docker pull ghcr.io/harper/hex:1.0.0
  docker pull ghcr.io/harper/hex:latest
  ```
- [ ] Homebrew formula updated
  ```bash
  brew upgrade harper/tap/hex
  ```

### Create/Edit GitHub Release
- [ ] Navigate to: https://github.com/2389-research/hex/releases/tag/v1.0.0
- [ ] Edit release notes (see template below)
- [ ] Mark as "Latest release"
- [ ] Verify artifacts are attached

---

## GitHub Release Notes Template

```markdown
# 🎉 Hex v1.0.0 - Production Release

We're excited to announce the first production-ready release of Hex, a powerful CLI for Claude AI inspired by Claude Code, Crush, Codex, and MaKeR!

## ✨ Highlights

### Interactive Terminal UI
- Beautiful TUI built with Bubbletea
- Real-time streaming responses
- Markdown rendering with syntax highlighting
- Vim-style navigation (j/k, gg/G, /)
- Token tracking and metrics

### Conversation Management
- SQLite persistence (never lose your work)
- Resume with `--continue` or `--resume <ID>`
- Full-text search with FTS5
- Favorites system
- Export to Markdown/JSON/HTML

### Powerful Tool System
13 built-in tools for file operations, code search, web access, and more:
- Read, Write, Edit - File operations
- Bash, BashOutput, KillShell - Command execution
- Grep, Glob - Code search
- WebFetch, WebSearch - Web access
- AskUserQuestion, TodoWrite, Task - Interactive workflows

### MCP Integration
- Full Model Context Protocol support
- Connect to external tool servers
- stdio transport
- JSON-RPC 2.0 messaging

### Vision & Multimodal
- Send images with `--image` flag
- Analyze screenshots and diagrams
- Multimodal content blocks

### Production Ready
- 80%+ test coverage
- Security audited (XSS fixes)
- Performance optimized (< 100ms startup)
- Structured logging
- Comprehensive documentation

## 📦 Installation

### Quick Install (Recommended)

**macOS/Linux:**
```bash
curl -sSL https://raw.githubusercontent.com/harper/hex/main/install.sh | bash
```

**Windows (PowerShell as Admin):**
```powershell
iwr -useb https://raw.githubusercontent.com/harper/hex/main/install.ps1 | iex
```

### Package Managers

**Homebrew (macOS/Linux):**
```bash
brew install harper/tap/hex
```

**Debian/Ubuntu (.deb):**
```bash
wget https://github.com/2389-research/hex/releases/download/v1.0.0/hex_1.0.0_Linux_x86_64.deb
sudo dpkg -i hex_1.0.0_Linux_x86_64.deb
```

**RedHat/Fedora (.rpm):**
```bash
wget https://github.com/2389-research/hex/releases/download/v1.0.0/hex_1.0.0_Linux_x86_64.rpm
sudo rpm -i hex_1.0.0_Linux_x86_64.rpm
```

**Alpine Linux (.apk):**
```bash
wget https://github.com/2389-research/hex/releases/download/v1.0.0/hex_1.0.0_Linux_x86_64.apk
sudo apk add --allow-untrusted hex_1.0.0_Linux_x86_64.apk
```

**Docker:**
```bash
docker pull ghcr.io/harper/hex:1.0.0
docker run --rm -e ANTHROPIC_API_KEY=$ANTHROPIC_API_KEY ghcr.io/harper/hex:1.0.0
```

**Go Install:**
```bash
go install github.com/2389-research/hex/cmd/hex@v1.0.0
```

## 🏁 Quick Start

```bash
# Set API key
export ANTHROPIC_API_KEY='your-api-key-here'

# Start interactive session
hex

# One-shot query
hex --print "Hello, world!"

# Resume last conversation
hex --continue

# Use with images
hex --image screenshot.png "What's in this image?"
```

## 📚 Documentation

- [User Guide](https://github.com/2389-research/hex/blob/main/docs/USER_GUIDE.md) - Complete usage guide
- [Architecture](https://github.com/2389-research/hex/blob/main/docs/ARCHITECTURE.md) - System design
- [Tools Reference](https://github.com/2389-research/hex/blob/main/docs/TOOLS.md) - All 13 tools
- [MCP Integration](https://github.com/2389-research/hex/blob/main/docs/MCP_INTEGRATION.md) - MCP guide

## 🔒 Security

This release includes:
- XSS vulnerability fixes in HTML export
- MCP tool approval heuristics
- Input validation throughout
- Secrets never logged
- Secure file path handling

## 📊 What's New in v1.0.0

See [CHANGELOG.md](https://github.com/2389-research/hex/blob/main/CHANGELOG.md) for full details.

### Added
- Interactive TUI with Bubbletea
- SSE streaming responses
- SQLite conversation storage
- 13 built-in tools
- MCP integration (JSON-RPC 2.0)
- Vision/multimodal support
- History search with FTS5
- Favorites system
- Autocomplete (3 providers)
- Quick actions command palette
- Export (Markdown/JSON/HTML)
- Templates (YAML-based)
- Smart suggestions with learning
- Structured logging
- GitHub Actions CI/CD
- Multi-platform builds
- Docker images
- Homebrew support

### Security
- Fixed XSS in HTML export
- MCP tool approval system
- Input validation
- Secrets protection

### Performance
- Startup time < 100ms
- Parallel MCP connections
- Database indexes optimized
- Memory usage < 50MB

## 🙏 Acknowledgments

Built with these excellent libraries:
- [Bubbletea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Styling
- [Glamour](https://github.com/charmbracelet/glamour) - Markdown rendering
- [Cobra](https://github.com/spf13/cobra) - CLI framework

## 🐛 Known Issues

None! Please report any issues on GitHub.

## 📝 Full Changelog

**Full Changelog**: https://github.com/2389-research/hex/compare/v0.5.0...v1.0.0
```

---

## Post-Release Verification

### Install and Test (All Platforms)

**macOS (Homebrew):**
```bash
brew install harper/tap/hex
hex --version
hex --print "Test query"
```

**Linux (Install Script):**
```bash
curl -sSL https://raw.githubusercontent.com/harper/hex/main/install.sh | bash
hex --version
```

**Ubuntu (Debian Package):**
```bash
wget https://github.com/2389-research/hex/releases/download/v1.0.0/hex_1.0.0_Linux_x86_64.deb
sudo dpkg -i hex_1.0.0_Linux_x86_64.deb
hex --version
```

**Docker:**
```bash
docker pull ghcr.io/harper/hex:1.0.0
docker run --rm ghcr.io/harper/hex:1.0.0 --version
```

**Windows (PowerShell):**
```powershell
iwr -useb https://raw.githubusercontent.com/harper/hex/main/install.ps1 | iex
hex --version
```

### Functional Testing
- [ ] API key setup works
- [ ] Interactive mode starts
- [ ] Streaming responses work
- [ ] Tool execution works
- [ ] Conversation persistence works
- [ ] Resume conversation works
- [ ] History search works
- [ ] Export works
- [ ] MCP integration works (if .mcp.json configured)

### Documentation Verification
- [ ] All links in README work
- [ ] Documentation renders correctly on GitHub
- [ ] Installation instructions accurate
- [ ] Code examples work

---

## Post-Release Communication

### GitHub Discussions
- [ ] Create announcement post
- [ ] Link to release notes
- [ ] Invite feedback

### Social Media (Optional)
- [ ] Twitter/X announcement
- [ ] Reddit r/golang post
- [ ] Hacker News submission
- [ ] Dev.to blog post

### Blog Post (Optional)
Topics to cover:
- Journey to v1.0
- Key features
- Lessons learned
- Future roadmap

---

## Monitoring (First Week)

### Metrics to Track
- [ ] Download count (GitHub Insights)
- [ ] Star count growth
- [ ] Issue reports
- [ ] Homebrew install analytics (if available)
- [ ] Docker pull count (GHCR metrics)

### Issue Triage
- [ ] Monitor GitHub Issues daily
- [ ] Respond to questions within 24h
- [ ] Fix critical bugs immediately
- [ ] Document common questions in FAQ

### Feedback Collection
- [ ] Create feedback discussion thread
- [ ] Survey users about missing features
- [ ] Note feature requests for v1.1

---

## Rollback Plan

If critical issue discovered after release:

### 1. Immediate Triage
```bash
# Assess severity
# - Does it affect all users?
# - Is data at risk?
# - Is there a workaround?
```

### 2. Quick Fix Path (Minor Issue)
```bash
# Fix in patch release v1.0.1
git checkout main
git checkout -b hotfix/v1.0.1
# Fix issue
git commit -m "fix: critical issue description"
git tag -a v1.0.1 -m "Hotfix for critical issue"
git push origin hotfix/v1.0.1
git push origin v1.0.1
```

### 3. Full Rollback (Critical Issue)
```bash
# Delete tag and release
git tag -d v1.0.0
git push origin :refs/tags/v1.0.0

# Delete GitHub Release
# Visit: https://github.com/2389-research/hex/releases
# Delete v1.0.0 release

# Update README to point to v0.5.0
# Fix issue
# Re-release as v1.0.1
```

### 4. Communication
- [ ] Post issue notice on GitHub
- [ ] Update README with warning
- [ ] Notify users via Discussions
- [ ] Document in CHANGELOG

---

## Success Criteria

Release is successful when:
- ✅ All artifacts available
- ✅ All installation methods work
- ✅ No critical bugs reported in first 48h
- ✅ Downloads > 10 in first week
- ✅ At least 3 positive user reports
- ✅ Documentation feedback positive

---

## Notes

- Don't rush - verify everything
- Test on fresh systems (VM or Docker)
- Have rollback plan ready
- Communicate clearly with users
- Celebrate the achievement! 🎉

---

**Last Updated**: 2025-11-28
