# hexviz - Hex Agent Execution Visualizer

A CLI tool for visualizing multi-agent execution from hex event files.

## Installation

```bash
go build -o bin/hexviz ./cmd/hexviz
```

## Usage

### Tree View (Default)

Shows hierarchical agent structure with costs and durations:

```bash
./bin/hexviz -events hex_events.jsonl
# or explicitly:
./bin/hexviz -events hex_events.jsonl -view tree
```

Example output:
```
root
├── Task: "Analyze codebase and refactor"
├── Cost: $0.09
└── Duration: 2m 15s

├─ root.1
│ ├── Task: "Search for import statements"
│ ├── Cost: $0.04
│ └── Duration: 45s

Total Tree Cost: $0.1576
```

### Timeline View

Shows chronological event sequence:

```bash
./bin/hexviz -events hex_events.jsonl -view timeline
```

Example output:
```
[2025-12-07 10:15:23] root: AgentStart - "Analyze codebase"
[2025-12-07 10:15:25] root: ToolCall - read_file(path="main.go")
[2025-12-07 10:15:26] root: ToolResult - read_file (success)
[2025-12-07 10:15:27] root.1: AgentStart - "Search for imports"
```

### Cost View

Shows detailed cost breakdown per agent:

```bash
./bin/hexviz -events hex_events.jsonl -view cost
```

Example output:
```
Agent Cost Breakdown:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
root
  Model: claude-sonnet-4-5-20250929
  Input:  12,500 tokens ($0.0375)
  Output:  3,200 tokens ($0.0480)
  Cache:   8,000 reads  ($0.0024)
  Total: $0.0879
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Total Cost: $0.1299
```

## Filtering

### Filter by Agent

Show only events for a specific agent and its descendants:

```bash
./bin/hexviz -events hex_events.jsonl -agent root.1
```

### Filter by Event Type

Show only specific event types:

```bash
./bin/hexviz -events hex_events.jsonl -type ToolCall
```

Combine with view modes:
```bash
./bin/hexviz -events hex_events.jsonl -view timeline -type ToolCall
```

## HTML Export

Generate an interactive HTML visualization:

```bash
./bin/hexviz -events hex_events.jsonl -html output.html
```

The HTML file includes:
- Dark theme styling
- Hierarchical tree visualization
- Cost breakdowns
- Standalone file (no external dependencies)

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-events` | `hex_events.jsonl` | Path to event file |
| `-view` | `tree` | View mode: `tree`, `timeline`, or `cost` |
| `-agent` | `` | Filter by agent ID (e.g., `root` or `root.1`) |
| `-type` | `` | Filter by event type (e.g., `ToolCall`) |
| `-html` | `` | Optional HTML output file |

## Event File Format

Events must be in JSON Lines format (one JSON object per line):

```jsonl
{"id":"evt-1","agent_id":"root","parent_id":"","type":"AgentStart","timestamp":"2025-12-07T10:15:23Z","data":{"task":"Main task","model":"claude-sonnet-4-5-20250929"}}
{"id":"evt-2","agent_id":"root.1","parent_id":"root","type":"AgentStart","timestamp":"2025-12-07T10:15:27Z","data":{"task":"Sub task","model":"claude-sonnet-4-5-20250929"}}
{"id":"evt-3","agent_id":"root","type":"AgentStop","timestamp":"2025-12-07T10:17:38Z","data":{"usage":{"input_tokens":12500,"output_tokens":3200}}}
```

Required event types:
- `AgentStart` - Contains `task` and `model` in data
- `AgentStop` - Contains `usage` object with token counts
- `ToolCall` - Tool execution request
- `ToolResult` - Tool execution result
- `StreamStart` - Streaming response started
- `StreamChunk` - Streaming response chunk

## Examples

### View execution of a specific sub-agent
```bash
./bin/hexviz -events hex_events.jsonl -agent root.2 -view tree
```

### See only tool interactions
```bash
./bin/hexviz -events hex_events.jsonl -type ToolCall -view timeline
```

### Generate shareable HTML report
```bash
./bin/hexviz -events hex_events.jsonl -view tree -html execution_report.html
```

### Analyze costs for root agent only
```bash
./bin/hexviz -events hex_events.jsonl -agent root -view cost
```

## Testing

Run the test suite:

```bash
go test ./cmd/hexviz/... -v
```

All tests follow TDD (Test-Driven Development) principles.
