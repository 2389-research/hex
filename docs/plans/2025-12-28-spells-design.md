# Spells: Switchable Agent Personalities for Hex

**Date:** 2025-12-28
**Status:** Approved
**Author:** Doctor Biz + Claude

## Summary

Spells let hex emulate other codegen agents by swapping system prompts and session configuration. Users can borrow useful behaviors from tools like Cursor or Codex, or switch between personality modes (terse, teacher, pair programmer) based on the task.

## Goals

1. **Borrow behaviors** - Adopt useful patterns from other agents without switching tools
2. **Personality switching** - Different modes for different tasks (terse for quick fixes, teacher for learning)
3. **Consistent with hex patterns** - Follow existing cascade (builtin → user → project)

## Non-Goals

- Full agent emulation (tool implementations, model switching)
- Automatic context-aware spell selection
- Spell marketplace/registry (v1 is local-only)

## Design

### Spell Structure

Each spell is a directory containing:

```
~/.hex/spells/codex/
├── system.md       # System prompt (markdown with YAML frontmatter)
├── config.yaml     # Session configuration
└── tools/          # Optional tool overrides
    └── bash.yaml
```

### system.md Format

```markdown
---
name: codex
description: OpenAI Codex-style agent
author: hex-team
version: 1.0.0
---

You are an AI coding assistant. You write clean, efficient code...
[rest of system prompt]
```

### config.yaml Format

```yaml
# Layering behavior (user can override with --replace/--layer)
mode: replace  # or "layer"

# Tool configuration
tools:
  enabled: [bash, read_file, write_file, edit]  # empty = all
  disabled: [web_search]

# Reasoning configuration (modern models)
reasoning:
  effort: medium        # none, low, medium, high
  show_thinking: false  # expose thinking blocks to user

# Response preferences
response:
  max_tokens: 4096
  format: text          # text, json, markdown
  style: concise        # concise, detailed, code-first

# Legacy sampling (optional, often ignored by reasoning models)
sampling:
  temperature: 1.0
```

### Tool Overrides (tools/*.yaml)

```yaml
# Override bash tool behavior for this spell
schema:
  properties:
    command:
      description: "Run shell command (prefer simple commands)"
defaults:
  timeout: 30000
restrictions:
  - no_sudo
  - no_rm_rf
```

### Layering Modes

- **replace** - Spell completely overrides hex's system prompt
- **layer** - Spell adds to hex's base prompt (keeps CLAUDE.md)

Spell author sets default in config.yaml. User can override at cast time.

## CLI Interface

### Interactive Mode

```bash
# Cast a spell (sticky until changed)
/spell codex

# Cast with override
/spell codex --replace    # pure emulation
/spell codex --layer      # add to existing context

# Check current spell
/spell                    # shows "Active spell: codex"

# Reset to default
/spell reset
/spell off

# List available
/spell list
```

### Print Mode

```bash
# Single invocation
./hex -p --spell codex "fix this bug"

# With layering override
./hex -p --spell codex --spell-mode=replace "fix this bug"

# Combine with other flags
./hex -p --spell terse --tools=bash,edit "refactor main.go"
```

### Management Commands

```bash
./hex spell list              # List all available spells
./hex spell info codex        # Show spell details
./hex spell validate <path>   # Validate a spell directory
```

## Curated Spells (v1)

### Agent Emulations

| Spell | Description | Key Traits |
|-------|-------------|------------|
| `codex` | OpenAI Codex style | Terse, code-focused, minimal explanation |
| `cursor` | Cursor editor style | Context-aware, refactor-friendly |

### Personality Modes

| Spell | Description | Key Traits |
|-------|-------------|------------|
| `terse` | Minimal output | Code only, no explanations unless asked |
| `teacher` | Educational mode | Explains reasoning, teaches concepts |

Start with 3-4 well-tested spells. Let community grow the library.

## Implementation

### New Package: internal/spells/

```
internal/spells/
├── spell.go           # Spell type definition
├── loader.go          # Directory-based loading
├── registry.go        # Spell registry
├── applicator.go      # Applies spell to session
└── tool_override.go   # Tool behavior modifications
```

### Key Types

```go
type Spell struct {
    Name          string
    Description   string
    SystemPrompt  string
    Config        SpellConfig
    ToolOverrides map[string]ToolOverride
    Mode          LayerMode  // replace or layer
    Source        string     // builtin, user, project
}

type SpellConfig struct {
    Tools     ToolsConfig
    Reasoning ReasoningConfig
    Response  ResponseConfig
}

type LayerMode string
const (
    LayerModeReplace LayerMode = "replace"
    LayerModeLayer   LayerMode = "layer"
)
```

### Loading Cascade

```
builtin (internal/spells/builtin/)
  → user (~/.hex/spells/)
  → project (./.hex/spells/)
```

Project spells override user spells override builtin spells (by name).

### Integration Points

1. **Session initialization** - Spell applied before first message
2. **System prompt builder** - Merge/replace based on mode
3. **Tool registry** - Apply tool overrides when spell active
4. **Config cascade** - Spell config merges with session defaults
5. **State tracking** - Active spell in session state

## Future Considerations (not v1)

- Spell composition (`/spell codex+terse`)
- Spell inheritance (base spell + modifications)
- Community spell registry
- Auto-suggestion based on file types
- Spell analytics (which spells used most)

## Open Questions

None - design validated through brainstorming session.

## References

- Hex skills system: `internal/skills/`
- Hex commands system: `internal/commands/`
- Hex templates: `internal/templates/`
