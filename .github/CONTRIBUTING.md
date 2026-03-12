# Contributing to Hex

Thank you for your interest in contributing to Hex! This document provides guidelines and instructions for contributing.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Testing](#testing)
- [Pull Request Process](#pull-request-process)
- [Coding Standards](#coding-standards)
- [Release Process](#release-process)

## Code of Conduct

This project follows the [Contributor Covenant Code of Conduct](https://www.contributor-covenant.org/version/2/1/code_of_conduct/). By participating, you are expected to uphold this code.

## Getting Started

### Prerequisites

- Go 1.24 or later
- Git
- Make
- golangci-lint (optional, for linting)
- goreleaser (optional, for testing releases)

### Fork and Clone

1. Fork the repository on GitHub
2. Clone your fork:

```bash
git clone https://github.com/YOUR_USERNAME/hex.git
cd hex
```

3. Add upstream remote:

```bash
git remote add upstream https://github.com/2389-research/hex.git
```

### Build and Test

```bash
# Install dependencies
make deps

# Build the binary
make build

# Run tests
make test

# Run all verification steps
make verify
```

## Development Workflow

### Creating a Feature Branch

```bash
# Fetch latest changes
git fetch upstream
git checkout main
git merge upstream/main

# Create feature branch
git checkout -b feature/your-feature-name
```

### Making Changes

1. Write code following our [coding standards](#coding-standards)
2. Add tests for new functionality
3. Update documentation as needed
4. Run verification steps:

```bash
make verify
```

### Committing Changes

We follow [Conventional Commits](https://www.conventionalcommits.org/):

```bash
# Format: <type>(<scope>): <description>

git commit -m "feat(ui): add keyboard shortcuts for navigation"
git commit -m "fix(storage): resolve database locking issue"
git commit -m "docs: update installation instructions"
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `test`: Test additions or updates
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `ci`: CI/CD changes
- `chore`: Maintenance tasks

## Testing

### Test Philosophy

- **No mocks**: We use real components and avoid mocks
- **Integration tests**: Test end-to-end workflows
- **Example-based**: Documentation doubles as validation

### Running Tests

```bash
# All tests
make test

# Short tests only (skip integration)
make test-short

# With coverage report
make test-coverage
```

### Writing Tests

```go
func TestFeature(t *testing.T) {
    // Arrange
    // ... setup

    // Act
    // ... execute

    // Assert
    require.NoError(t, err)
    assert.Equal(t, expected, actual)
}
```

## Pull Request Process

### Before Submitting

1. Ensure all tests pass: `make test`
2. Run linting: `make lint`
3. Update documentation if needed
4. Add entry to CHANGELOG.md
5. Rebase on latest main:

```bash
git fetch upstream
git rebase upstream/main
```

### Submitting PR

1. Push to your fork:

```bash
git push origin feature/your-feature-name
```

2. Create pull request on GitHub
3. Fill out the PR template completely
4. Link related issues
5. Wait for CI to pass
6. Address review feedback

### PR Title Format

Follow conventional commit format:

```
feat(ui): add keyboard shortcuts for navigation
fix(storage): resolve database locking issue
```

### PR Description

Include:
- **What**: What changes does this PR introduce?
- **Why**: Why are these changes needed?
- **How**: How were the changes implemented?
- **Testing**: How were the changes tested?
- **Screenshots**: If UI changes, include before/after screenshots

## Coding Standards

### Go Style

Follow the official [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments).

**Key points:**
- Use `gofmt` for formatting (run `make fmt`)
- Keep functions small and focused
- Write descriptive variable names
- Add comments for exported functions
- Handle errors explicitly

### File Headers

All code files should start with:

```go
// ABOUTME: Brief description of what this file does
// ABOUTME: Second line if needed
package packagename
```

### Project Structure

```
hex/
├── cmd/hex/           # CLI entry point
├── internal/           # Private implementation
│   ├── core/          # API client, types, config
│   ├── ui/            # Bubbletea TUI
│   ├── storage/       # SQLite persistence
│   └── tools/         # Tool system
├── docs/              # Documentation
└── test/              # Integration tests
```

### Imports

Group imports as:

```go
import (
    // Standard library
    "context"
    "fmt"

    // External dependencies
    "github.com/charmbracelet/bubbletea"
    "github.com/spf13/cobra"

    // Internal packages
    "github.com/2389-research/hex/internal/core"
    "github.com/2389-research/hex/internal/ui"
)
```

### Error Handling

```go
// Good: Wrap errors with context
if err != nil {
    return fmt.Errorf("failed to read config: %w", err)
}

// Bad: Generic error messages
if err != nil {
    return err
}
```

### Testing

```go
// Use testify for assertions
import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestExample(t *testing.T) {
    // Use require for fatal errors
    result, err := doSomething()
    require.NoError(t, err)

    // Use assert for non-fatal checks
    assert.Equal(t, expected, result)
}
```

## Release Process

Releases are automated via GitHub Actions when a tag is pushed.

### Creating a Release

1. Update version in relevant files
2. Update CHANGELOG.md
3. Create and push tag:

```bash
git tag -a v1.2.3 -m "Release v1.2.3"
git push upstream v1.2.3
```

4. GitHub Actions will:
   - Run tests
   - Build binaries for all platforms
   - Create GitHub release
   - Update Homebrew tap

See [.github/RELEASE_CHECKLIST.md](.github/RELEASE_CHECKLIST.md) for detailed steps.

## Documentation

### Where to Document

- **Code comments**: Exported functions and complex logic
- **README.md**: Quick start and overview
- **docs/USER_GUIDE.md**: Comprehensive usage guide
- **docs/ARCHITECTURE.md**: System design and internals
- **docs/TOOLS.md**: Tool system reference
- **CHANGELOG.md**: Version history

### Documentation Style

- Use clear, concise language
- Include code examples
- Add screenshots for UI features
- Keep examples up-to-date

## Getting Help

- **Questions**: Open a discussion on GitHub
- **Bugs**: Open an issue with reproduction steps
- **Features**: Open an issue to discuss before implementing

## Attribution

Contributors are recognized in:
- GitHub contributors page
- Release notes for their contributions

Thank you for contributing to Hex!
