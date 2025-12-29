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
# Current Spell

Show the currently active spell (if any).
{{end}}
