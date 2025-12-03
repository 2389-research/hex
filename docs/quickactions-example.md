# Quick Actions Menu - Visual Guide

## What is it?

The Quick Actions Menu is a Vim-style command palette that appears when you press `:` in Hex's interactive mode. It provides fuzzy-searchable shortcuts to tools and common actions.

## How to Use

### Opening the Menu

Press `:` when the textarea is not focused (similar to Vim's command mode).

```
┌────────────────────────────────────────────────────────────┐
│                                                            │
│                      Quick Actions                         │
│                                                            │
│  :_                                                        │
│                                                            │
│  ▸ read <file> - Read a file                              │
│    grep <pattern> - Search files with grep                │
│    web <url> - Fetch web page                             │
│    attach <file> - Attach an image                        │
│    save - Save conversation                               │
│                                                            │
│  ... and 1 more                                           │
│                                                            │
│  Enter: execute • Esc: cancel                             │
│                                                            │
└────────────────────────────────────────────────────────────┘
```

### Searching Actions

Type to filter actions with fuzzy search:

```
┌────────────────────────────────────────────────────────────┐
│                                                            │
│                      Quick Actions                         │
│                                                            │
│  :re_                                                      │
│                                                            │
│  ▸ read <file> - Read a file                              │
│                                                            │
│  Enter: execute • Esc: cancel                             │
│                                                            │
└────────────────────────────────────────────────────────────┘
```

### Executing with Arguments

Type the full command with arguments:

```
┌────────────────────────────────────────────────────────────┐
│                                                            │
│                      Quick Actions                         │
│                                                            │
│  :read src/main.go_                                        │
│                                                            │
│  No matching actions                                       │
│                                                            │
│  Enter: execute • Esc: cancel                             │
│                                                            │
└────────────────────────────────────────────────────────────┘
```

When you press Enter, it will:
1. Parse "read" as the command
2. Parse "src/main.go" as the argument
3. Execute `read` action with arg "src/main.go"

## Available Actions

| Action | Usage | Description |
|--------|-------|-------------|
| `read` | `:read <file>` | Read a file and add to conversation |
| `grep` | `:grep <pattern>` | Search files with grep |
| `web` | `:web <url>` | Fetch and parse a web page |
| `attach` | `:attach <file>` | Attach an image to conversation |
| `save` | `:save` | Save current conversation |
| `export` | `:export` | Export conversation as markdown |

## Fuzzy Search Examples

The fuzzy search is smart about matching:

| Query | Matches | Reason |
|-------|---------|--------|
| `read` | `read` | Exact match |
| `re` | `read` | Prefix match |
| `rd` | `read` | Fuzzy match (r...d) |
| `gre` | `grep` | Prefix match |
| `wb` | `web` | Fuzzy match (w...b) |
| `exp` | `export` | Prefix match |

## Keyboard Shortcuts

When the Quick Actions modal is open:

- **Type**: Search/filter actions
- **Enter**: Execute first matched action (or parsed command)
- **Backspace**: Delete character
- **Esc**: Close modal without executing

## Integration with Tools

The quick actions system is designed to integrate with Hex's tool execution system:

1. User types `:read config.yaml`
2. Quick Actions parses command: `read` + `config.yaml`
3. Executes `read` action handler
4. Handler triggers tool system with parsed args
5. Tool approval flow (if needed)
6. Result appears in conversation

## Design Philosophy

### Vim-Inspired
- `:` prefix feels natural to developers
- Quick and keyboard-friendly
- No mouse required

### Fuzzy-First
- Type what you remember
- Don't need exact spelling
- Instant filtering as you type

### Command + Autocomplete Hybrid
- Shows available actions
- Accepts full commands with args
- Best of both worlds

## Future Enhancements

Potential improvements for future tasks:

1. **Arrow Navigation**: ↑/↓ to select from filtered list
2. **Action History**: Recently used actions first
3. **More Actions**:
   - `:help` - Show help
   - `:clear` - Clear conversation
   - `:history` - Show history
   - `:favorite` - Toggle favorite
4. **Custom Actions**: User-defined shortcuts
5. **Action Preview**: Show what will happen before executing

## Technical Details

### Files
- `internal/ui/quickactions.go` - Core implementation
- `internal/ui/quickactions_test.go` - Unit tests
- `internal/ui/quickactions_integration_test.go` - Integration tests
- Model integration in `model.go`, `update.go`, `view.go`

### Testing
- 23 comprehensive tests
- Unit tests for registry, search, parsing
- Integration tests for model interaction
- All tests passing ✅

### Performance
- Custom fuzzy algorithm (no external deps)
- O(n) search where n = number of actions
- Fast enough for interactive use
- Registry uses sync.RWMutex for thread safety
