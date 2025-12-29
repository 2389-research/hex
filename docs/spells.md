# Spells

Spells let you switch hex's personality to emulate other coding agents or adopt different behavior modes.

## Quick Start

```bash
# Use the terse spell (minimal output)
hex -p --spell terse "read main.go"

# Use the teacher spell (educational mode)
hex -p --spell teacher "explain this regex: ^[a-z]+$"

# In interactive mode
/spell terse
/spell list
/spell reset
```

## Available Spells

### Builtin Spells

| Spell | Description |
|-------|-------------|
| `terse` | Minimal output - code only, no explanations |
| `teacher` | Educational mode - explains concepts thoroughly |

### Custom Spells

Create your own spells in `~/.hex/spells/<name>/`:

```
~/.hex/spells/my-spell/
├── system.md       # Required: System prompt
├── config.yaml     # Optional: Configuration
└── tools/          # Optional: Tool overrides
    └── bash.yaml
```

## Creating a Spell

### system.md

```markdown
---
name: my-spell
description: My custom spell
author: yourname
version: 1.0.0
---

Your system prompt content here.
Instructions for how the AI should behave.
```

### config.yaml

```yaml
# How the spell interacts with existing prompts
mode: layer  # or "replace"

# Tool configuration
tools:
  enabled: [bash, read_file, edit]
  disabled: [web_search]

# Reasoning behavior
reasoning:
  effort: medium  # none, low, medium, high
  show_thinking: false

# Response preferences
response:
  max_tokens: 4096
  style: concise  # concise, detailed, code-first
```

### Tool Overrides

Create `tools/<tool-name>.yaml` to customize tool behavior:

```yaml
# tools/bash.yaml
defaults:
  timeout: 30000
restrictions:
  - no_sudo
```

## Layer vs Replace Mode

- **layer** (default): Spell adds to your existing CLAUDE.md and hex defaults
- **replace**: Spell completely replaces the system prompt

Override at runtime:
```bash
hex -p --spell codex --spell-mode=replace "..."
hex -p --spell codex --spell-mode=layer "..."
```

## Spell Precedence

Spells are loaded from multiple locations with later sources overriding earlier:

1. Builtin (`internal/spells/builtin/`)
2. User (`~/.hex/spells/`)
3. Project (`.hex/spells/`)

A project spell with the same name as a user spell will take precedence.
