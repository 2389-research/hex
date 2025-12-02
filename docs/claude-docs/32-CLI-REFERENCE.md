# Claude Code CLI Reference

## Overview

Claude Code ships as a single binary named `claude`. This document explains how the CLI works, which commands exist, and how options map to runtime behavior. It is a condensed reference distilled from the executable's Commander setup.

## Primary Command

```bash
claude [prompt] [options]
```

- `prompt` (optional) – initial natural-language instruction
- Defaults to interactive TUI mode (React/Ink surface)
- Exit with `Ctrl+C` or `/exit`

### Mode Selection

| Option | Purpose |
|--------|---------|
| `-p, --print` | Non-interactive run that prints the full response and exits |
| `-c, --continue` | Resume the most recent session |
| `-r, --resume [id]` | Resume a specific session (prompts to select if omitted) |
| `--fork-session` | Resume but force a new session ID |

### Output/Input Control

- `--output-format text|json|stream-json`
- `--input-format text|stream-json` (non-interactive only)
- `--include-partial-messages`, `--replay-user-messages` (streaming pipelines)
- `--json-schema <schema>` – enforce structured responses in `--print` mode

### Model Selection and Limits

- `--model <alias>` (e.g., `sonnet`, `opus`)
- `--fallback-model <alias>`
- `--max-thinking-tokens <n>`
- `--max-turns <n>`
- `--max-budget-usd <amount>`

### Debugging and Logging

- `-d, --debug [filter]`
- `--verbose`
- `--mcp-debug` (legacy toggle routed through `--debug`)

### Tool and Permission Control

- `--tools <tool ...>` – restrict tool palette (non-interactive)
- `--allowed-tools <tool ...>` / `--disallowed-tools <tool ...>`
- `--permission-mode <mode>` – e.g., `auto`, `ask`
- `--dangerously-skip-permissions` / `--allow-dangerously-skip-permissions` (sandbox use only)

### Prompt Control

- `--system-prompt <text>`
- `--system-prompt-file <path>`
- `--append-system-prompt*` variants to extend defaults

### Configuration Overrides

- `--settings <json-or-path>` – inject configuration for this run
- `--setting-sources user,project,local` – select which config layers to load
- `--add-dir <path ...>` – expand tool sandbox roots
- `--session-id <uuid>`

### MCP Configuration

- `--mcp-config <json-or-path ...>` – load server definitions
- `--strict-mcp-config` – ignore auto-discovery, use explicit list only

### Advanced Flags

- `--agents <json>` – define ad-hoc subagents for session
- `--plugin-dir <path ...>` – load plugins without global install
- `--ide` – auto-connect to an IDE if available
- `--sdk-url <url>` – stream via SDK transport
- `--teleport [session]`, `--remote <desc>` – remote execution helpers
- `--permission-prompt-tool <tool>` – override how permission prompts are surfaced
- `--resume-session-at <message id>` – cut history when resuming
- `--enable-auth-status`
- `-v, --version`, `-h, --help`

## Subcommands

### `claude mcp`

Manages Model Context Protocol servers:

- `claude mcp serve` – run Claude Code itself as an MCP server
- `claude mcp add <name> <command>` – register server definitions
- `claude mcp add-json <name> <json>`
- `claude mcp list` / `remove` / `update`
- Diagnostic helpers: `servers`, `tools`, `info`, `call`, `grep`, `resources`, `read`

### `claude plugin`

Handles plugin lifecycle:

- `install`, `uninstall`, `enable`, `disable`, `update`, `list`, `show`, `search`
- Works with marketplace catalogs or direct Git/URL references

### Utilities

- `claude doctor` – environment diagnostics
- `claude setup-token` – authentication helper
- `claude migrate-installer` – upgrade legacy installs

## Mode Reference

### Interactive Sessions

- TUI shows chat log, todo tracker, git status, background processes
- Slash commands (`/plan`, `/tests`, etc.) expand to stored prompts
- Tool calls render inline with real-time output

### Non-Interactive Pipelines

Use `--print` for scripting:

```bash
claude "list functions in src/api" --print --output-format json
```

Combine with `jq`, `rg`, or shell scripts for automation.

## Example Workflows

```bash
# Quick question without interactive shell
claude "summarize README" --print

# Resume a previous session and fork history
claude --resume 4af0cf80 --fork-session

# Lock down available tools and capture streaming JSON
claude "run pytest" \
  --print \
  --output-format stream-json \
  --tools Read Edit Bash
```

## Tips

- Prefer `--print` for CI/CD usage; it avoids the TUI but keeps tool access.
- Combine `--settings` with temporary JSON objects to test new configurations safely.
- Use `claude mcp tools` and `claude plugin list` to introspect runtime capabilities before delegating work.
