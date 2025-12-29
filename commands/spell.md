---
name: spell
description: Switch agent personality (spell)
args:
  action: "Spell name, 'list', 'reset', or 'off'"
---

{{if eq .action "list"}}
# Available Spells

List all available spells that can be cast.
{{else if eq .action "reset"}}
# Reset Spell

Reset to default hex behavior (no spell active).
{{else if eq .action "off"}}
# Spell Off

Disable the current spell and return to default behavior.
{{else if .action}}
# Cast Spell: {{.action}}

Activate the {{.action}} spell for this session.
{{else}}
# Available Spells

**Builtin spells you can cast:**

| Spell | Description | Mode |
|-------|-------------|------|
| `terse` | Minimal output, code-first | layer |
| `teacher` | Educational, explains concepts | layer |
| `codex` | OpenAI Codex style, minimal prompting | layer |
| `claude-code` | Extreme brevity, direct answers | replace |
| `devin` | Methodical planning-first | layer |
| `cursor` | Context-aware, refactor-friendly | layer |
| `antigravity` | Design-focused, aesthetics-first | layer |

**Usage:**
- `/spell <name>` - Cast a spell (e.g., `/spell terse`)
- `/spell reset` or `/spell off` - Return to default hex behavior

**Modes:**
- **layer** - Adds personality while keeping Hex identity
- **replace** - Full identity override (e.g., claude-code identifies as Claude)

**Print mode:** `hex -p --spell <name> "prompt"`
{{end}}
