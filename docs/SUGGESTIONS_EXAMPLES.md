# Smart Suggestions - Visual Examples

## Overview
The smart suggestion system analyzes your input in real-time and suggests relevant tools based on detected patterns.

## Example Screenshots (Text Representation)

### Example 1: File Path Detection
```
┌──────────────────────────────────────────────────────────────────┐
│ Hex • claude-3-5-sonnet-20241022 ●                             │
│                                                                  │
│ [Chat Messages Appear Here]                                     │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────────┐
│ Can you read /etc/hosts for me?                                 │
└──────────────────────────────────────────────────────────────────┘

╭──────────────────────────────────────────────────────────────────╮
│ 💡 Suggestions                                                   │
│                                                                  │
│ → read_file                                                      │
│   Detected absolute file path                                   │
│   Action: :read /etc/hosts                                      │
│                                                                  │
│ Tab: accept • Esc: dismiss                                      │
╰──────────────────────────────────────────────────────────────────╯

Tokens: 0 in / 0 out [Chat] ctrl+c: quit • enter: send • tab: accept
```

**What happens:**
- User types path in their message
- Suggestion appears below input
- Press **Tab** → Input changes to `:read /etc/hosts`
- Press **Esc** → Suggestion disappears

---

### Example 2: Multiple Suggestions
```
┌──────────────────────────────────────────────────────────────────┐
│ search for TODO in /src/main.go                                 │
└──────────────────────────────────────────────────────────────────┘

╭──────────────────────────────────────────────────────────────────╮
│ 💡 Suggestions                                                   │
│                                                                  │
│ → read_file                                                      │
│   Detected absolute file path                                   │
│   Action: :read /src/main.go                                    │
│                                                                  │
│ Other suggestions:                                              │
│   • grep (75% confident)                                        │
│                                                                  │
│ Tab: accept • Esc: dismiss                                      │
╰──────────────────────────────────────────────────────────────────╯
```

**Explanation:**
- Input matches multiple patterns (file path + search intent)
- Top suggestion (highest confidence) shown prominently
- Other suggestions listed below
- Tab accepts top suggestion only

---

### Example 3: URL Detection
```
┌──────────────────────────────────────────────────────────────────┐
│ Fetch https://api.github.com/repos/harper/hex                  │
└──────────────────────────────────────────────────────────────────┘

╭──────────────────────────────────────────────────────────────────╮
│ 💡 Suggestions                                                   │
│                                                                  │
│ → web_fetch                                                      │
│   Detected HTTP/HTTPS URL                                       │
│   Action: :web_fetch https://api.github.com/repos/harper/hex   │
│                                                                  │
│ Tab: accept • Esc: dismiss                                      │
╰──────────────────────────────────────────────────────────────────╯
```

---

### Example 4: Shell Command
```
┌──────────────────────────────────────────────────────────────────┐
│ git status                                                       │
└──────────────────────────────────────────────────────────────────┘

╭──────────────────────────────────────────────────────────────────╮
│ 💡 Suggestions                                                   │
│                                                                  │
│ → bash                                                           │
│   Input looks like a shell command                              │
│   Action: :bash                                                 │
│                                                                  │
│ Tab: accept • Esc: dismiss                                      │
╰──────────────────────────────────────────────────────────────────╯
```

---

### Example 5: Glob Pattern
```
┌──────────────────────────────────────────────────────────────────┐
│ Find all **/*.go files in the project                           │
└──────────────────────────────────────────────────────────────────┘

╭──────────────────────────────────────────────────────────────────╮
│ 💡 Suggestions                                                   │
│                                                                  │
│ → glob                                                           │
│   Detected glob pattern                                         │
│   Action: :glob **/*.go                                         │
│                                                                  │
│ Tab: accept • Esc: dismiss                                      │
╰──────────────────────────────────────────────────────────────────╯
```

---

## Pattern Detection Cheat Sheet

| Input Pattern | Tool Suggested | Confidence | Example |
|--------------|----------------|------------|---------|
| `/path/to/file` | read_file | 85% | `/etc/hosts` |
| `./relative/path` | read_file | 80% | `./config.yaml` |
| `~/home/path` | read_file | 85% | `~/Documents/notes.txt` |
| `https://...` | web_fetch | 90% | `https://example.com` |
| `http://...` | web_fetch | 90% | `http://api.example.com` |
| "search for X" | grep | 75% | "search for TODO" |
| "find X" | grep | 75% | "find all functions" |
| "run X" | bash | 80% | "run npm test" |
| `git status` | bash | 85% | Any common shell command |
| `**/*.ext` | glob | 75% | `**/*.js` |
| `*.ext` | glob | 75% | `*.txt` |
| "write to X" | write_file | 80% | "write to config.json" |
| "edit X" | edit | 75% | "edit main.go" |
| "google X" | web_search | 80% | "google React hooks" |

---

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| **Tab** | Accept top suggestion (applies action to input) |
| **Esc** | Dismiss all suggestions |
| *(typing)* | Auto-analyzes and updates suggestions |

---

## Smart Behaviors

### 1. No Suggestions for Short Input
```
Input: "hi"
Result: No suggestions (too short)
```

### 2. No Suggestions for Tool Commands
```
Input: ":read /etc/hosts"
Result: No suggestions (already a tool command)
```

### 3. No Suggestions Mid-Sentence
```
Input: "how "
Result: No suggestions (trailing space, clearly mid-sentence)
```

### 4. High Confidence Only
```
Input: "maybe check something"
Result: No suggestions (no pattern matched with >=70% confidence)
```

### 5. Deduplicated Suggestions
```
Input: "Compare /etc/hosts and ./config.yaml"
Result: Only ONE read_file suggestion (not two)
```

---

## Learning System in Action

### Scenario 1: Accepting Suggestions
```
Day 1: User types "/etc/hosts" → suggests read_file (85%)
        User presses Tab → Accepts

Day 2: User types "/var/log/app.log" → suggests read_file (87%)
        Learning system increased confidence!
```

### Scenario 2: Rejecting Suggestions
```
Day 1: User types "ls -la" → suggests bash (85%)
        User presses Esc → Rejects

Day 2: User types "git status" → suggests bash (80%)
        Still suggests but lower confidence

Day 5: User types "docker ps" → No bash suggestion
        Confidence dropped below 70% threshold
```

### Scenario 3: Ignoring Suggestions
```
User types "/etc/hosts"
Suggestion appears but user just keeps typing
Eventually presses Enter without Tab or Esc
→ Counted as "Ignored" (mild negative feedback)
```

---

## Tips for Best Experience

1. **Accept Useful Suggestions** - Press Tab when suggestion is helpful
2. **Dismiss Bad Suggestions** - Press Esc when not useful (helps learning)
3. **Just Ignore** - Keep typing if you don't care (mild feedback)
4. **Be Specific** - More context = better suggestions
5. **Use Natural Language** - System understands phrases like "search for", "run", "edit"

---

## Color Coding (Terminal)

- **Tool Name**: Bold Cyan
- **Reason**: Italic Gray
- **Action**: Light Gray
- **Help Text**: Darker Gray
- **Border**: Light Blue

---

## Advanced: Viewing Statistics

While not currently exposed in UI, the learning system tracks:

```go
// Internal statistics available
{
  "read_file": {
    Total: 25,
    Accepted: 20,
    Rejected: 3,
    Ignored: 2,
    AcceptanceRate: 0.80,
    ConfidenceAdjustment: +0.12
  },
  "bash": {
    Total: 15,
    Accepted: 5,
    Rejected: 8,
    Ignored: 2,
    AcceptanceRate: 0.33,
    ConfidenceAdjustment: -0.18
  }
}
```

Future enhancement: Show these stats in a Tools view or help panel.
