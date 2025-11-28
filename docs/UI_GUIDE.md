# Clem UI Guide

This guide covers the interactive UI features, keyboard shortcuts, and customization options for Clem's TUI (Terminal User Interface).

## Overview

Clem features a modern, interactive terminal UI built with Bubbletea that provides:
- Real-time streaming responses with token rate display
- Interactive tool approval with risk assessment
- Multiple view modes (Chat, History, Tools)
- Comprehensive keyboard shortcuts
- Visual indicators for all operations

## UI Components

### Title Bar
```
Clem • claude-3-5-sonnet-20241022 ● ⚡ Streaming (42 tok/s)
```
- Model name
- Connection status indicator (○ disconnected, ● connected, ◉ streaming/error)
- Streaming indicator with token rate (when streaming)
- Tool execution spinner (when tools running)

### Main View Area

The main area shows different content based on the current view mode:

**Chat View**: Shows conversation history with:
- User messages prefixed with "You:"
- Assistant messages with markdown rendering
- Streaming responses rendered in real-time

**History View**: Browse past conversations (coming soon)

**Tools View**: Inspect available tools and execution history (coming soon)

### Input Area

```
┌────────────────────────────────────────┐
│ Send a message...                      │
│                                        │
└────────────────────────────────────────┘
```

Multi-line textarea for composing messages. Press Enter to send, Alt+Enter for newline.

### Status Bar

```
claude-3-5-sonnet ● 12k↓ 8k↑ [chat]  ?:help ^C:quit
```

Shows:
- Model name
- Connection status
- Token usage (input ↓ / output ↑)
- Context usage indicator (bar graph when >50% full)
- Current mode [chat/history/tools]
- Quick help text

## Keyboard Shortcuts

### Global Shortcuts

| Shortcut | Action |
|----------|--------|
| `Ctrl+C` | Quit Clem |
| `Esc` | Exit current mode / Quit |
| `?` | Toggle help panel |
| `Tab` | Switch view (Chat → History → Tools → Chat) |

### Conversation Management

| Shortcut | Action |
|----------|--------|
| `Ctrl+L` | Clear screen (clears viewport display) |
| `Ctrl+K` | Clear conversation (deletes all messages) |
| `Ctrl+S` | Save conversation to database |
| `Ctrl+E` | Export conversation to markdown format |

### Input & Messaging

| Shortcut | Action |
|----------|--------|
| `Enter` | Send message |
| `Alt+Enter` | Insert newline in message |

### Navigation (Vim-style)

These work when the input area is not focused:

| Shortcut | Action |
|----------|--------|
| `j` | Scroll down one line |
| `k` | Scroll up one line |
| `gg` | Go to top of conversation |
| `G` | Go to bottom of conversation |
| `/` | Enter search mode |

### Streaming Controls

| Shortcut | Action |
|----------|--------|
| `Ctrl+T` | Toggle typewriter mode (progressive reveal) |

### Tool Approval

When a tool requests approval, the following keys are available:

| Shortcut | Action |
|----------|--------|
| `y` or `a` | Approve and execute tool |
| `n` or `d` | Deny tool execution |
| `v` | View details (working directory, tool ID, etc.) |
| `Esc` | Deny and exit approval mode |

## Tool Approval Interface

When Claude wants to run a tool, you'll see a detailed approval prompt:

```
┌─ Tool Approval Required ───────────────────────┐
│ Tool: bash                                      │
│ Risk: Caution ⚠                                │
│                                                 │
│ Parameters:                                     │
│   command: git status                           │
│                                                 │
│ [A]pprove  [D]eny  [V]iew Details              │
└─────────────────────────────────────────────────┘
```

### Risk Levels

Tools are automatically assessed for risk:

- **Safe ✓** (Green): Read-only operations (list, get, read, search, find)
- **Caution ⚠** (Yellow): Write operations or shell commands
- **Danger ⚠⚠** (Red): Dangerous commands detected (rm, delete, format, curl|sh, etc.)

### Tool Execution Indicators

While a tool is running, you'll see:
```
⣾ Running bash... (1.2s)
```

The spinner animates and shows elapsed time.

## Streaming Features

### Token Rate Display

When receiving a streaming response:
```
⚡ Streaming (42 tok/s)
```

Shows real-time token generation rate using an exponential moving average.

### Typewriter Mode

Press `Ctrl+T` to enable typewriter mode, which progressively reveals streaming text character-by-character for a more dramatic effect.

When enabled, you'll see:
```
⚡ Streaming (42 tok/s) [typewriter 67%]
```

## Status Indicators

### Connection Status

- `○` Gray: Disconnected from API
- `●` Green: Connected and idle
- `◉` Cyan: Actively streaming
- `◉` Red: Error state

### Context Usage

When your conversation uses >50% of the context window, a progress bar appears:
```
[████████░░]
```

- Green bars: Normal usage (<80%)
- Yellow bars: High usage (>80%)
- Warning message appears when approaching limit

## Search Mode

Press `/` to enter search mode:
```
Search: my query_
```

Type your search query and press Enter to execute (coming soon).

Press Esc to exit search mode.

## Help Panel

Press `?` to toggle the full help panel:

```
┌─────────────────────────────────────────┐
│ Keyboard Shortcuts                      │
│                                         │
│ Ctrl+C         Quit Clem               │
│ Ctrl+L         Clear screen            │
│ Ctrl+K         Clear conversation      │
│ ...                                    │
└─────────────────────────────────────────┘
```

Press `?` again or `Esc` to hide it.

## Visual Examples

### Normal Chat
```
Clem • claude-3-5-sonnet-20241022 ●

You: Hello!
Assistant: Hi\! How can I help you today?

You: Can you help me analyze this code?

Assistant: Of course\! Please share the code you would like...
```

### Streaming Response
```
Clem • claude-3-5-sonnet-20241022 ◉ ⚡ Streaming (38 tok/s)

You: What is recursion?

Assistant:
Recursion is a programming technique where a function calls itself...
```

### Tool Execution
```
Clem • claude-3-5-sonnet-20241022 ● ⣾ Running bash... (0.8s)

┌─ Tool Approval Required ───────────────────────┐
│ Tool: bash                                      │
│ Risk: Safe ✓                                   │
│                                                 │
│ Parameters:                                     │
│   command: git status                           │
│                                                 │
│ [A]pprove  [D]eny  [V]iew Details              │
└─────────────────────────────────────────────────┘
```

## Tips and Tricks

1. **Vim Navigation**: Unfocus the input (press Esc) to use vim-style navigation (j/k/gg/G)
2. **Quick Save**: Press Ctrl+S to save your conversation periodically
3. **Context Management**: Watch the context usage bar and export/clear when getting full
4. **Typewriter Mode**: Great for presentations or following along with long responses
5. **Tool Approval**: Press "v" to see full details before approving dangerous operations

## Troubleshooting

### Input Not Working
- Make sure the input area is focused (click on it or press Esc to unfocus, then click)
- Check that you are not in search mode (press Esc to exit)

### Vim Keys Not Working
- Vim navigation only works when input is NOT focused
- Press Esc to unfocus the input area

### Streaming Stopped
- Check connection indicator (should be ● or ◉)
- Look for error messages in the chat
- Press Ctrl+C and restart if needed

## Customization

Current version uses fixed color scheme. Future versions will support:
- Custom color themes
- Configurable keybindings
- Layout customization

## Accessibility

- All colors have been chosen for readability
- Status indicators use both color AND symbols
- Keyboard-only navigation is fully supported
- Screen reader support planned for future releases
