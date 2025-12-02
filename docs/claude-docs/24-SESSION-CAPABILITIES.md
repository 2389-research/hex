# Session Capabilities

## Overview

Claude Code automatically discovers every enabled MCP server at session start. This document explains how those capabilities appear, how to inspect them, and what kinds of tools typically show up. Treat this as a mental model rather than a hardcoded list—actual availability depends on project configuration.

## Capability Discovery Workflow

1. **Startup** – Claude Code loads MCP definitions from user/project configuration and launches each server.
2. **Handshake** – Each server reports its tools via `tools/list`.
3. **Tool Palette** – Claude merges MCP tools with native tools, exposing them through the same planner/selector.
4. **Inspection** – Use `claude mcp tools`, `claude mcp info <tool>`, or ask Claude to enumerate the current palette when you need the exact set.

## Tool Naming

```
mcp__<server>__<tool>
```

- Prefix keeps names unique
- Middle segment is the configured server identifier
- Final segment is the tool within that server

Example:

```text
mcp__playwright__browser_navigate
```

## Typical Capability Families

| Family | Examples | Common Uses |
|--------|----------|-------------|
| Knowledge & Logging | Chronicle, Private Journal | Session resumption, decision history, personal learning |
| Productivity | Todo or project trackers | Progress reporting, lightweight workflow automation |
| Browser Automation | Playwright, Chrome remoting | UI testing, visual inspection, recording repro steps |
| External APIs | GitHub, Asana, Sentry, custom SaaS | Triage issues, sync tickets, fetch metrics |
| Local Utilities | Filesystem explorers, database shells | Cross-repo search, structured data access |

Each family can expose multiple tools (readers, writers, search endpoints, etc.). Claude Code picks tools dynamically based on the task at hand.

## Example Tool Signature

```typescript
mcp__chronicle__remember_this({
  activity: string,
  context?: string
})
```

This pattern mirrors built-in tools: structured parameters, documented semantics, and deterministic output.

## Best Practices

- **Ask before using** unfamiliar MCP tools; some may interact with production systems.
- **Inspect schema** with `claude mcp info <tool>` to understand required parameters.
- **Model fallback**: always plan for a server to be unavailable. Code defensively and provide alternatives (e.g., skip logging if Chronicle is disabled).
- **Document dependencies**: when drafting workflows, note which MCP servers are required so teammates can enable them if needed.
