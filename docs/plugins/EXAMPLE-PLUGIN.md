# Example Plugin Structure

This document shows how to create a plugin for Clem.

## Directory Structure

```
my-plugin/
├── plugin.json           # Plugin manifest (required)
├── README.md            # Documentation
├── skills/              # Skill files (optional)
│   ├── my-skill.md
│   └── another-skill.md
├── commands/            # Slash commands (optional)
│   └── my-command.md
├── hooks/               # Hook scripts (optional)
│   └── post-edit.sh
└── scripts/             # Lifecycle scripts (optional)
    ├── install.sh
    └── uninstall.sh
```

## Minimal Plugin

The minimal plugin requires only a `plugin.json` file:

```json
{
  "name": "my-plugin",
  "version": "1.0.0",
  "description": "A simple Clem plugin"
}
```

## Complete Plugin Example

### plugin.json

```json
{
  "name": "go-development",
  "version": "1.2.0",
  "description": "Go development tools and workflows",
  "author": "Your Name <you@example.com>",
  "license": "MIT",
  "repository": {
    "type": "git",
    "url": "https://github.com/yourusername/clem-go-plugin.git"
  },
  "keywords": ["go", "golang", "development"],

  "skills": [
    "skills/go-tdd.md",
    "skills/go-testing.md"
  ],

  "commands": [
    "commands/go-test.md",
    "commands/go-build.md"
  ],

  "hooks": {
    "PostToolUse": {
      "Edit": ["gofmt -w ${file}"]
    }
  },

  "scripts": {
    "install": "scripts/install.sh"
  },

  "activation": {
    "languages": ["go"],
    "files": ["go.mod", "go.sum"]
  }
}
```

### skills/go-tdd.md

```markdown
---
name: go-tdd
description: Test-driven development for Go
tags: [go, testing, tdd]
priority: 10
---

# Go Test-Driven Development

When implementing Go features, follow TDD:

## Process

1. Write a failing test in `*_test.go`
2. Run `go test` to see it fail
3. Write minimal code to pass
4. Run `go test` to see it pass
5. Refactor if needed

## Go Testing Best Practices

- Use table-driven tests
- Test exported functions
- Use subtests with `t.Run()`
- Use `testdata/` for fixtures
```

### commands/go-test.md

```markdown
Run Go tests with coverage:

```bash
go test -v -cover ./...
```

Show coverage report and identify untested code.
```

### hooks/post-edit.sh

```bash
#!/bin/bash
# Format Go files after editing

if [[ "${CLAUDE_TOOL_FILE_PATH}" == *.go ]]; then
    gofmt -w "${CLAUDE_TOOL_FILE_PATH}"
    echo "Formatted ${CLAUDE_TOOL_FILE_PATH}" >&2
fi
```

### scripts/install.sh

```bash
#!/bin/bash
# Install plugin dependencies

echo "Installing Go development plugin..." >&2

# Check if gofmt is available
if ! command -v gofmt &> /dev/null; then
    echo "Warning: gofmt not found. Please install Go." >&2
    exit 1
fi

echo "Go development plugin installed successfully!" >&2
```

## Installation

Users can install your plugin in several ways:

### From Git Repository

```bash
clem plugin install https://github.com/yourusername/clem-go-plugin.git
```

### From Local Directory

```bash
clem plugin install ./my-plugin
```

## Plugin Manifest Fields

| Field | Required | Description |
|-------|----------|-------------|
| `name` | Yes | Unique plugin identifier (lowercase, hyphens only) |
| `version` | Yes | Semantic version (e.g., "1.0.0") |
| `description` | Yes | One-line summary |
| `author` | No | Author name and email |
| `license` | No | License identifier (e.g., "MIT") |
| `repository` | No | Source repository info |
| `keywords` | No | Search tags |
| `skills` | No | Skill file paths |
| `commands` | No | Command file paths |
| `hooks` | No | Hook configurations |
| `scripts` | No | Lifecycle script paths |
| `activation` | No | When plugin should activate |

## Activation Rules

Control when your plugin activates:

```json
{
  "activation": {
    "onStartup": true,
    "languages": ["go", "python"],
    "files": ["go.mod", "requirements.txt"],
    "projects": ["django", "flask"]
  }
}
```

- `onStartup`: Always activate
- `languages`: Activate if these languages are detected
- `files`: Activate if these files exist in project
- `projects`: Activate for these project types

## Publishing

1. Create a GitHub repository for your plugin
2. Tag a release: `git tag v1.0.0`
3. Push the tag: `git push origin v1.0.0`
4. Users can install directly from your repository URL

## Testing Your Plugin

```bash
# Install from local directory
clem plugin install ./my-plugin

# List installed plugins
clem plugin list

# Test that skills and commands are loaded
clem

# Uninstall when done
clem plugin uninstall my-plugin
```

## Best Practices

1. **Focused Purpose**: Each plugin should do one thing well
2. **Semantic Versioning**: Follow semver strictly
3. **Documentation**: Include comprehensive README.md
4. **Minimal Dependencies**: Keep plugins lightweight
5. **Graceful Degradation**: Handle missing tools gracefully
6. **Clear Naming**: Use descriptive plugin and skill names
7. **Test Thoroughly**: Test on clean systems before publishing

## Common Patterns

### Language/Framework Plugin

Provides tools for specific tech stack (React, Django, etc.)

### Methodology Plugin

Teaches development practices (TDD, clean code, etc.)

### Tool Integration Plugin

Connects external services (databases, cloud providers, etc.)

### Domain Expertise Plugin

Specialized knowledge (security, performance, accessibility, etc.)

## Support

For questions or issues:
- Check [Clem documentation](../../README.md)
- Open an issue in the plugin repository
- Contact plugin author (see plugin.json)
