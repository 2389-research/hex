# Plugins System

## Overview

Plugins are packaged bundles of extensibility components that add capabilities to Claude Code. They combine skills, hooks, subagents, MCP servers, and slash commands into distributable packages.

## What Are Plugins?

A plugin is a directory structure containing:

- **Skills**: Domain knowledge and methodologies
- **Subagents**: Specialized AI assistants
- **Hooks**: Automated workflow triggers
- **MCP Servers**: External tool integrations
- **Slash Commands**: Prompt shortcuts
- **Documentation**: Usage guides and examples
- **Configuration**: Default settings and templates

Think of plugins as app stores packages - they bundle related functionality into installable units.

## Plugin Architecture

### Directory Structure

```
~/.claude/plugins/my-plugin/
├── manifest.json           # Plugin metadata
├── README.md              # Documentation
├── skills/                # Skill files
│   ├── skill-one.md
│   └── skill-two.md
├── agents/                # Subagent definitions
│   ├── specialist.md
│   └── reviewer.md
├── commands/              # Slash commands
│   ├── custom-cmd.md
│   └── another-cmd.md
├── hooks/                 # Hook scripts
│   ├── post-edit.sh
│   └── pre-commit.sh
├── mcp/                   # MCP server configs
│   └── server-config.json
├── templates/             # File templates
│   ├── component.tsx
│   └── test.spec.ts
└── scripts/               # Utility scripts
    └── setup.sh
```

### Manifest Format

`manifest.json` defines plugin metadata:

```json
{
  "name": "my-plugin",
  "displayName": "My Awesome Plugin",
  "version": "1.2.0",
  "description": "Adds X, Y, and Z capabilities to Claude Code",
  "author": "Your Name <you@example.com>",
  "license": "MIT",
  "homepage": "https://github.com/username/my-plugin",
  "repository": {
    "type": "git",
    "url": "https://github.com/username/my-plugin.git"
  },
  "keywords": [
    "testing",
    "react",
    "typescript"
  ],
  "engines": {
    "claude-code": ">=1.0.0"
  },
  "dependencies": {
    "other-plugin": "^2.0.0"
  },
  "configuration": {
    "defaults": {
      "autoFormat": true,
      "strictMode": false
    },
    "schema": {
      "type": "object",
      "properties": {
        "autoFormat": {
          "type": "boolean",
          "description": "Automatically format code after edits"
        },
        "strictMode": {
          "type": "boolean",
          "description": "Enable strict validation rules"
        }
      }
    }
  },
  "activation": {
    "onStartup": true,
    "languages": ["typescript", "javascript"],
    "projects": ["react", "next"]
  },
  "contributes": {
    "skills": [
      {
        "path": "skills/react-testing.md",
        "default": true
      }
    ],
    "agents": [
      {
        "path": "agents/react-specialist.md",
        "default": false
      }
    ],
    "commands": [
      {
        "path": "commands/create-component.md",
        "alias": "component"
      }
    ],
    "hooks": [
      {
        "event": "PostToolUse",
        "script": "hooks/post-edit.sh",
        "match": {
          "toolName": "Edit",
          "filePattern": ".*\\.(ts|tsx)$"
        }
      }
    ]
  },
  "scripts": {
    "install": "scripts/setup.sh",
    "uninstall": "scripts/cleanup.sh",
    "update": "scripts/update.sh"
  }
}
```

### Manifest Fields

| Field | Required | Description |
|-------|----------|-------------|
| `name` | Yes | Unique plugin identifier |
| `displayName` | Yes | Human-readable name |
| `version` | Yes | Semantic version (semver) |
| `description` | Yes | One-line summary |
| `author` | No | Author name and email |
| `license` | No | SPDX license identifier |
| `homepage` | No | Plugin website URL |
| `repository` | No | Source code repository |
| `keywords` | No | Search tags |
| `engines` | No | Compatible Claude Code versions |
| `dependencies` | No | Required other plugins |
| `configuration` | No | Plugin settings schema |
| `activation` | No | When plugin loads |
| `contributes` | No | What plugin provides |
| `scripts` | No | Lifecycle hooks |

## Marketplace and Distribution

### Official Marketplace

Claude Code plugins are distributed via:

1. **Official Registry**: `registry.claude.dev`
2. **GitHub Repositories**: Tagged releases
3. **Local Installation**: Direct file placement

### Installing Plugins

#### From Marketplace

```bash
# Install by name
claude plugin install superpowers

# Install specific version
claude plugin install superpowers@2.1.0

# Install from GitHub
claude plugin install github:anthropic/superpowers
```

#### From Local Directory

```bash
# Install from path
claude plugin install ./my-plugin

# Symlink for development
claude plugin link ./my-plugin
```

#### From URL

```bash
# Install from Git URL
claude plugin install https://github.com/user/plugin.git

# Install from tarball
claude plugin install https://example.com/plugin.tar.gz
```

### Managing Plugins

```bash
# List installed plugins
claude plugin list

# Show plugin info
claude plugin show superpowers

# Update plugin
claude plugin update superpowers

# Update all plugins
claude plugin update --all

# Uninstall plugin
claude plugin uninstall superpowers

# Enable/disable without uninstalling
claude plugin disable superpowers
claude plugin enable superpowers

# Search marketplace
claude plugin search testing
```

## Creating Plugins

### Step 1: Initialize Structure

```bash
# Create plugin directory
mkdir -p ~/.claude/plugins/my-plugin
cd ~/.claude/plugins/my-plugin

# Create basic structure
mkdir -p skills agents commands hooks templates scripts
```

### Step 2: Create Manifest

`manifest.json`:
```json
{
  "name": "my-plugin",
  "displayName": "My Plugin",
  "version": "0.1.0",
  "description": "My first Claude Code plugin",
  "author": "Your Name",
  "license": "MIT"
}
```

### Step 3: Add Components

**Skill** (`skills/my-skill.md`):
```markdown
---
name: my-skill
description: Does something useful
---

# My Skill

Instructions for Claude...
```

**Subagent** (`agents/my-agent.md`):
```markdown
---
name: my-agent
description: Specialized assistant
tools:
  - Read
  - Edit
---

You are a specialist in...
```

**Slash Command** (`commands/my-cmd.md`):
```markdown
Perform a specific task with these requirements:
- Requirement 1
- Requirement 2
```

**Hook Script** (`hooks/post-edit.sh`):
```bash
#!/bin/bash
echo "File edited: ${CLAUDE_TOOL_FILE_PATH}"
```

### Step 4: Update Manifest

```json
{
  "name": "my-plugin",
  "version": "0.1.0",
  "contributes": {
    "skills": [
      { "path": "skills/my-skill.md" }
    ],
    "agents": [
      { "path": "agents/my-agent.md" }
    ],
    "commands": [
      { "path": "commands/my-cmd.md", "alias": "mycmd" }
    ],
    "hooks": [
      {
        "event": "PostToolUse",
        "script": "hooks/post-edit.sh",
        "match": { "toolName": "Edit" }
      }
    ]
  }
}
```

### Step 5: Test Locally

```bash
# Link for development
claude plugin link ~/.claude/plugins/my-plugin

# Test the plugin
claude

# Use the skill, agent, or command
```

### Step 6: Document

`README.md`:
```markdown
# My Plugin

One-line description of what it does.

## Installation

```bash
claude plugin install my-plugin
```

## Features

- Feature 1: Description
- Feature 2: Description

## Usage

### Skill: my-skill

Explain when and how to use...

### Agent: my-agent

Describe what the agent specializes in...

### Command: /mycmd

Show example usage...

## Configuration

Available settings:
- `setting1`: Description (default: value)
- `setting2`: Description (default: value)

## Examples

Provide real-world examples...

## License

MIT
```

### Step 7: Publish

#### To GitHub

```bash
git init
git add .
git commit -m "Initial commit"
git tag v0.1.0
git push origin main --tags
```

Users install with:
```bash
claude plugin install github:yourusername/my-plugin
```

#### To Marketplace

```bash
# Login to marketplace
claude plugin login

# Publish
claude plugin publish
```

## Plugin Patterns

### 1. Language/Framework Plugin

Provides tools for specific tech stack:

```
react-toolkit/
├── skills/
│   ├── react-hooks.md
│   ├── react-testing.md
│   └── react-performance.md
├── agents/
│   └── react-specialist.md
├── templates/
│   ├── component.tsx
│   ├── hook.ts
│   └── test.spec.tsx
└── commands/
    ├── create-component.md
    └── create-hook.md
```

**Activation**:
```json
{
  "activation": {
    "languages": ["typescript", "javascript"],
    "projects": ["react", "next"]
  }
}
```

### 2. Methodology Plugin

Teaches development practices:

```
tdd-toolkit/
├── skills/
│   ├── test-driven-development.md
│   ├── test-patterns.md
│   └── mocking-strategies.md
├── agents/
│   └── test-writer.md
├── hooks/
│   └── pre-commit-tests.sh
└── commands/
    └── generate-tests.md
```

**Activation**:
```json
{
  "activation": {
    "onStartup": true
  }
}
```

### 3. Tool Integration Plugin

Connects external services:

```
deployment-toolkit/
├── mcp/
│   └── kubernetes-server.json
├── skills/
│   ├── k8s-deployment.md
│   └── docker-optimization.md
├── hooks/
│   ├── post-commit-deploy.sh
│   └── pre-deploy-checks.sh
└── scripts/
    └── setup-kubectl.sh
```

**Activation**:
```json
{
  "activation": {
    "onStartup": false
  },
  "scripts": {
    "install": "scripts/setup-kubectl.sh"
  }
}
```

### 4. Domain Expertise Plugin

Specialized knowledge:

```
security-toolkit/
├── skills/
│   ├── owasp-top-10.md
│   ├── secure-coding.md
│   └── crypto-patterns.md
├── agents/
│   ├── security-auditor.md
│   └── penetration-tester.md
└── templates/
    ├── security-checklist.md
    └── threat-model.md
```

**Activation**:
```json
{
  "activation": {
    "onStartup": false
  }
}
```

## Advanced Features

### Plugin Dependencies

Plugins can depend on other plugins:

```json
{
  "name": "advanced-react",
  "dependencies": {
    "react-toolkit": "^1.0.0",
    "testing-toolkit": "^2.0.0"
  }
}
```

Claude Code automatically installs dependencies.

### Configuration Schema

Define plugin settings:

```json
{
  "configuration": {
    "defaults": {
      "componentStyle": "functional",
      "testFramework": "jest",
      "strictMode": true
    },
    "schema": {
      "type": "object",
      "properties": {
        "componentStyle": {
          "type": "string",
          "enum": ["functional", "class"],
          "description": "Preferred React component style"
        },
        "testFramework": {
          "type": "string",
          "enum": ["jest", "vitest", "mocha"],
          "description": "Testing framework to use"
        },
        "strictMode": {
          "type": "boolean",
          "description": "Enable strict validation"
        }
      }
    }
  }
}
```

Users configure in `.claude/settings.json`:
```json
{
  "plugins": {
    "react-toolkit": {
      "componentStyle": "functional",
      "testFramework": "vitest"
    }
  }
}
```

Skills/agents access via environment variables:
```bash
${PLUGIN_REACT_TOOLKIT_COMPONENT_STYLE}
${PLUGIN_REACT_TOOLKIT_TEST_FRAMEWORK}
```

### Conditional Activation

Plugins can auto-activate based on context:

```json
{
  "activation": {
    "onStartup": false,
    "languages": ["python"],
    "files": ["requirements.txt", "setup.py"],
    "projects": ["django", "flask"]
  }
}
```

Activates when:
- Project contains Python files
- Has `requirements.txt` or `setup.py`
- Is a Django or Flask project

### Lifecycle Scripts

Run code at plugin lifecycle events:

```json
{
  "scripts": {
    "install": "scripts/install.sh",
    "uninstall": "scripts/uninstall.sh",
    "update": "scripts/update.sh",
    "activate": "scripts/activate.sh",
    "deactivate": "scripts/deactivate.sh"
  }
}
```

**install.sh**:
```bash
#!/bin/bash
# Download additional resources
curl -O https://example.com/data.json

# Install system dependencies
npm install -g some-tool

# Setup configuration
cp templates/config.json ~/.config/
```

**uninstall.sh**:
```bash
#!/bin/bash
# Clean up resources
rm -rf ~/.config/my-plugin

# Remove system dependencies
npm uninstall -g some-tool
```

### MCP Server Integration

Plugins can bundle MCP servers:

`mcp/database-server.json`:
```json
{
  "name": "database-tools",
  "command": "npx",
  "args": ["-y", "@myorg/database-mcp-server"],
  "env": {
    "DB_CONNECTION_STRING": "${DATABASE_URL}"
  }
}
```

Referenced in manifest:
```json
{
  "contributes": {
    "mcpServers": [
      {
        "path": "mcp/database-server.json",
        "autoStart": true
      }
    ]
  }
}
```

### Template System

Provide file templates:

`templates/component.tsx`:
```typescript
import React from 'react';

interface {{ComponentName}}Props {
  // Props go here
}

export const {{ComponentName}}: React.FC<{{ComponentName}}Props> = (props) => {
  return (
    <div>
      {/* Component content */}
    </div>
  );
};
```

`commands/create-component.md`:
```markdown
Create a new React component with the following:

1. Ask user for component name
2. Create file: `src/components/{{ComponentName}}.tsx`
3. Use template from `templates/component.tsx`
4. Replace `{{ComponentName}}` with actual name
5. Create test file: `src/components/{{ComponentName}}.test.tsx`
```

## Best Practices

### 1. Focused Purpose

Each plugin should have a clear, focused purpose:

**Good**: `react-hooks-toolkit` - React hooks expertise
**Bad**: `web-dev-everything` - Too broad

### 2. Semantic Versioning

Follow semver strictly:
- `1.0.0` → `1.0.1`: Bug fixes
- `1.0.0` → `1.1.0`: New features (backward compatible)
- `1.0.0` → `2.0.0`: Breaking changes

### 3. Minimal Dependencies

Only depend on plugins you truly need:
- Each dependency is another thing to break
- Users must install all dependencies
- Circular dependencies are forbidden

### 4. Graceful Degradation

Plugin should work even if optional features fail:

```bash
#!/bin/bash
# Try to use fancy tool, fall back to basic
if command -v fancy-tool &> /dev/null; then
  fancy-tool process
else
  echo "Warning: fancy-tool not found, using basic mode"
  cat file | basic-process
fi
```

### 5. User Configuration

Provide sensible defaults, allow customization:

```json
{
  "configuration": {
    "defaults": {
      "enabled": true,
      "verbose": false
    }
  }
}
```

### 6. Documentation

Include comprehensive docs:
- README with examples
- Inline comments in skills
- Change log
- Migration guides for breaking changes

### 7. Testing

Test your plugin:
- Verify skills load correctly
- Test agents work as expected
- Check hooks trigger properly
- Validate slash commands expand correctly

### 8. Versioning Components

If components can update independently:

```
skills/
├── skill-v1.md
└── skill-v2.md
```

```json
{
  "contributes": {
    "skills": [
      {
        "path": "skills/skill-v2.md",
        "default": true
      },
      {
        "path": "skills/skill-v1.md",
        "default": false
      }
    ]
  }
}
```

### 9. Backward Compatibility

When updating plugins:
- Don't break existing functionality
- Deprecate before removing
- Provide migration paths
- Document breaking changes

### 10. Security

Plugins run with user privileges:
- Validate all inputs
- Don't hardcode secrets
- Use secure defaults
- Warn on dangerous operations

## Distribution Checklist

Before publishing:

- [ ] Manifest is complete and valid
- [ ] Version follows semver
- [ ] README has installation instructions
- [ ] Examples are included
- [ ] License is specified
- [ ] All skills/agents load correctly
- [ ] Hooks work as expected
- [ ] Slash commands expand properly
- [ ] No hardcoded paths or secrets
- [ ] Tested on clean installation
- [ ] CHANGELOG is updated
- [ ] Git tags match version
- [ ] Repository is public (if open source)

## Marketplace Submission

To submit to official marketplace:

1. **Prepare repository**:
   - Public GitHub repo
   - Valid `manifest.json`
   - Comprehensive `README.md`
   - Open source license (MIT, Apache, etc.)

2. **Tag release**:
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```

3. **Submit**:
   ```bash
   claude plugin login
   claude plugin publish
   ```

4. **Review process**:
   - Automated validation
   - Security scan
   - Manual review (24-48 hours)
   - Approval or feedback

5. **Published**:
   - Listed in marketplace
   - Users can install by name
   - Updates via `claude plugin update`

## Plugin Examples

### Minimal Plugin

```
minimal-plugin/
├── manifest.json
├── README.md
└── skills/
    └── my-skill.md
```

`manifest.json`:
```json
{
  "name": "minimal-plugin",
  "version": "1.0.0",
  "description": "Minimal example plugin",
  "contributes": {
    "skills": [
      { "path": "skills/my-skill.md" }
    ]
  }
}
```

### Full-Featured Plugin

```
full-plugin/
├── manifest.json
├── README.md
├── CHANGELOG.md
├── LICENSE
├── skills/
│   ├── skill-one.md
│   └── skill-two.md
├── agents/
│   ├── agent-one.md
│   └── agent-two.md
├── commands/
│   ├── cmd-one.md
│   └── cmd-two.md
├── hooks/
│   ├── post-edit.sh
│   └── pre-commit.sh
├── templates/
│   ├── template.tsx
│   └── config.json
├── scripts/
│   ├── install.sh
│   ├── uninstall.sh
│   └── utils.sh
├── mcp/
│   └── server.json
└── docs/
    ├── getting-started.md
    └── advanced.md
```

## Summary

Plugins are powerful distribution mechanisms that:

- **Bundle related components** into installable packages
- **Extend Claude Code** with new capabilities
- **Share via marketplace** or direct installation
- **Version and update** independently
- **Configure** per-user or per-project

Key plugin components:
- **Manifest**: Metadata and configuration
- **Skills**: Knowledge and processes
- **Agents**: Specialized assistants
- **Commands**: Prompt shortcuts
- **Hooks**: Workflow automation
- **MCP Servers**: External tool integration
- **Templates**: File scaffolding

With thoughtful design, plugins become force multipliers - packaging expertise and automation into shareable units that make every Claude Code user more productive.
