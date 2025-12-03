# Codex Best Practices → hex Integration Plan

**Based on**: CODEX_AUDIT.md analysis
**Goal**: Adopt high-impact patterns from Codex to improve hex

---

## Priority Matrix

| Practice | Impact | Effort | Priority | Status |
|----------|--------|--------|----------|--------|
| **1. Auto-compaction** | 🔥 HIGH | ⚡ LOW | **P0** | Infrastructure exists, just needs wiring |
| **2. Session resuming** | 🔥 HIGH | 🔨 MEDIUM | **P1** | Would greatly improve UX |
| **3. AGENTS.md directory traversal** | 🔥 HIGH | 🔨 MEDIUM | **P1** | Codex's approach is superior |
| **4. MCP integration** | 🔥 HIGH | 🏗️ HIGH | **P2** | Industry standard, future-proof |
| **5. Execpolicy DSL** | 🌡️ MEDIUM | 🏗️ HIGH | **P3** | Nice-to-have, JSON rules work |
| **6. Image input support** | 🌡️ MEDIUM | 🔨 MEDIUM | **P3** | Valuable for debugging |
| **7. NDJSON rollout format** | ❄️ LOW | 🏗️ HIGH | **P4** | SQLite works fine for us |

---

## P0: Auto-Compaction (IMMEDIATE WIN)

### Current State
- ✅ `internal/context/manager.go` - Has PruneContext() logic
- ✅ `internal/context/summarizer.go` - Has SummarizeMessages()
- ✅ `internal/ui/model.go:570` - Has GetPrunedMessages()
- ❌ **NEVER CALLED** - streamMessage() and sendToolResults() don't use it

### The Fix (5 minutes)
**File**: `internal/ui/update.go:785-796`

**Current code:**
```go
func (m *Model) streamMessage(_ string) tea.Cmd {
    messages := make([]core.Message, 0, len(m.Messages))
    for _, msg := range m.Messages {
        if msg.Role == "tool" {
            continue
        }
        messages = append(messages, core.Message{
            Role:         msg.Role,
            Content:      msg.Content,
            ContentBlock: msg.ContentBlock,
        })
    }
```

**Should be:**
```go
func (m *Model) streamMessage(_ string) tea.Cmd {
    // Use pruned messages instead of raw message list
    messages := m.GetPrunedMessages()
```

**Same fix needed in**: `internal/ui/model.go:983` (sendToolResults)

### Benefits
- ✅ Prevents context overflow errors
- ✅ Reduces token usage (cost savings)
- ✅ Enables longer conversations
- ✅ Infrastructure already written and tested

### Implementation Steps
1. Replace manual message copying with `GetPrunedMessages()` call
2. Test with long conversation (100+ messages)
3. Verify pruning happens when context is near limit
4. Add debug logging to show when compaction occurs

---

## P1: Session Resuming with Picker UI

### What Codex Does Well
```bash
codex resume              # Interactive picker
codex resume --last       # Last session
codex resume {id}         # Specific session
```

**Picker shows:**
- Session ID (truncated)
- Original CWD
- Git branch (if in repo)
- Last activity time
- First user message (as title)

### Current State in hex
- Sessions stored in `~/.hex/hex.db`
- No built-in resume functionality
- User must manually track conversation IDs

### Implementation Plan

#### Phase 1: CLI Resume Command
**File**: `cmd/hex/root.go`

Add new command:
```go
var resumeCmd = &cobra.Command{
    Use:   "resume [conversation-id]",
    Short: "Resume a previous conversation",
    Long:  "Resume an interactive session from history",
    Run:   runResume,
}

func runResume(cmd *cobra.Command, args []string) {
    // If --last flag, get most recent
    // If ID provided, use that
    // Otherwise, show picker
}
```

#### Phase 2: Picker TUI
**New file**: `internal/ui/session_picker.go`

Using Bubbletea list component:
```go
type SessionPickerModel struct {
    sessions []SessionItem
    cursor   int
    selected *SessionItem
}

type SessionItem struct {
    ID        string
    CWD       string
    Branch    string
    Title     string
    UpdatedAt time.Time
}
```

#### Phase 3: Database Queries
**File**: `internal/storage/conversations.go`

```go
func ListRecentConversations(db *sql.DB, limit int) ([]Conversation, error)
func GetConversationByID(db *sql.DB, id string) (*Conversation, error)
func GetMostRecentConversation(db *sql.DB) (*Conversation, error)
```

### Benefits
- 🎯 Users can easily continue work
- 🎯 No need to remember conversation IDs
- 🎯 Git branch awareness helps context switching
- 🎯 Better for iterative development workflows

---

## P1: AGENTS.md Directory Traversal

### What Codex Does Better

**Codex lookup order (merged top-down):**
1. `~/.codex/AGENTS.md` - Global
2. Repo root `AGENTS.md`
3. Each parent dir → CWD
4. CWD `AGENTS.md` or `AGENTS.override.md` (override replaces)

**Key insight**: Override files **replace** instead of merge

### Current State in hex
**File**: `internal/ui/model.go:204-210`

```go
// Load CLAUDE.md from current directory or ~/.hex/CLAUDE.md
claudeMdPath := filepath.Join(".", "CLAUDE.md")
if _, err := os.Stat(claudeMdPath); os.IsNotExist(err) {
    home, _ := os.UserHomeDir()
    claudeMdPath = filepath.Join(home, ".hex", "CLAUDE.md")
}
```

**Problems:**
- ❌ Only checks CWD and global
- ❌ Doesn't traverse parent directories
- ❌ No override mechanism
- ❌ Doesn't merge multiple files

### Implementation Plan

#### New Approach
**File**: `internal/config/agents_md.go`

```go
// LoadAgentsMd traverses from repo root to CWD, merging AGENTS.md files
func LoadAgentsMd(cwd string) (string, error) {
    var sections []string

    // 1. Global ~/.hex/AGENTS.md
    if global := loadGlobal(); global != "" {
        sections = append(sections, global)
    }

    // 2. Find repo root
    repoRoot := findRepoRoot(cwd)
    if repoRoot == "" {
        repoRoot = cwd  // No git repo, use CWD as root
    }

    // 3. Traverse from repo root → CWD
    path := repoRoot
    for {
        // Check for override first
        if override := loadFile(filepath.Join(path, "AGENTS.override.md")); override != "" {
            sections = append(sections, override)
        } else if regular := loadFile(filepath.Join(path, "AGENTS.md")); regular != "" {
            sections = append(sections, regular)
        }

        // Stop at CWD
        if path == cwd {
            break
        }

        // Move to next subdirectory toward CWD
        path = nextPathSegment(path, cwd)
    }

    return strings.Join(sections, "\n\n---\n\n"), nil
}
```

### Benefits
- 🎯 Monorepo support (different rules per package)
- 🎯 Override mechanism for special directories
- 🎯 Respects project structure
- 🎯 Merges global + project + local instructions

---

## P2: MCP (Model Context Protocol) Integration

### What is MCP?
- **Standard protocol** for AI agents to access external tools
- Created by Anthropic
- Used by Codex, Claude Desktop, others
- Server/client architecture over stdio or SSE

### Why MCP Matters
- 🌐 Access to **any** MCP server (databases, APIs, services)
- 🔌 Pluggable tool ecosystem
- 🏭 Industry standard (not proprietary)
- 🚀 Future-proof extensibility

### Examples of MCP Servers
```bash
# Filesystem access
npx @modelcontextprotocol/server-filesystem ~/projects

# GitHub integration
npx @modelcontextprotocol/server-github

# Database queries
npx @modelcontextprotocol/server-postgres postgres://localhost

# Web browsing
npx @modelcontextprotocol/server-puppeteer
```

### Implementation Plan

#### Phase 1: MCP Client Library
**New package**: `internal/mcp/`

```go
package mcp

type Client struct {
    conn   io.ReadWriter
    tools  []Tool
}

func NewClient(command string, args []string) (*Client, error)
func (c *Client) ListTools() ([]Tool, error)
func (c *Client) CallTool(name string, input map[string]any) (string, error)
```

#### Phase 2: Configuration
**File**: `~/.hex/config.yaml`

```yaml
mcp_servers:
  filesystem:
    command: "npx"
    args: ["@modelcontextprotocol/server-filesystem", "/Users/harper/projects"]

  github:
    command: "npx"
    args: ["@modelcontextprotocol/server-github"]
    env:
      GITHUB_TOKEN: "${GITHUB_TOKEN}"
```

#### Phase 3: Tool Registry Integration
**File**: `internal/tools/registry.go`

```go
// Register MCP tools alongside built-in tools
func (r *Registry) LoadMCPServers(config MCPConfig) error {
    for name, serverConfig := range config.Servers {
        client := mcp.NewClient(serverConfig.Command, serverConfig.Args)
        tools := client.ListTools()

        for _, tool := range tools {
            r.Register(tool.Name, &MCPToolHandler{
                client: client,
                tool:   tool,
            })
        }
    }
}
```

### Benefits
- 🌐 Access external systems without custom code
- 🔌 Community-built tool servers
- 📦 Easy to add new capabilities
- 🎯 Standard protocol (not custom)

---

## P3: Execpolicy DSL vs JSON Rules

### Current Approach (Just Implemented)
**File**: `internal/approval/rules.go`

```json
{
  "rules": {
    "Read": "always_allow",
    "Write": "never_allow"
  }
}
```

**Limitations:**
- Per-tool only (not per-command pattern)
- No wildcard matching
- No conditional logic
- Binary allow/deny (no "ask")

### Codex's Execpolicy
**File**: `.execpolicy.toml`

```toml
# Allow all read operations
[[rules]]
pattern = "cat *"
action = "allow"

[[rules]]
pattern = "ls *"
action = "allow"

# Require approval for destructive ops
[[rules]]
pattern = "rm -rf *"
action = "ask"
justification_required = true

# Block dangerous commands
[[rules]]
pattern = "dd if=*"
action = "deny"
reason = "Direct disk access is dangerous"

# Wildcard support
[[rules]]
pattern = "git *"
action = "allow"
sandbox = true
```

### Pros/Cons

| Approach | Pros | Cons |
|----------|------|------|
| **JSON (ours)** | Simple, works for basic use | Limited expressiveness |
| **Execpolicy DSL** | Powerful, granular control | More complex to implement |

### Recommendation
- **Keep JSON for now** (P3 priority)
- Consider execpolicy DSL if users request pattern matching
- Could add `"pattern"` field to our JSON rules as middle ground

---

## P3: Image Input Support

### What Codex Does
```bash
# CLI flag
codex -i screenshot.png "Explain this error"
codex --image img1.png,img2.jpg "Summarize these"

# TUI paste
Ctrl+V / Cmd+V to paste images directly
```

### Implementation Plan

#### Phase 1: CLI Input
**File**: `cmd/hex/root.go`

```go
var imageFiles []string

rootCmd.Flags().StringSliceVarP(&imageFiles, "image", "i", []string{}, "Image files to include")
```

#### Phase 2: Base64 Encoding
**File**: `internal/core/message.go`

```go
type ImageContent struct {
    Type   string `json:"type"`   // "image"
    Source struct {
        Type      string `json:"type"`       // "base64"
        MediaType string `json:"media_type"` // "image/png"
        Data      string `json:"data"`       // base64 encoded
    } `json:"source"`
}

func LoadImage(path string) (*ImageContent, error) {
    data, _ := os.ReadFile(path)
    encoded := base64.StdEncoding.EncodeToString(data)
    mediaType := detectMediaType(path)

    return &ImageContent{
        Type: "image",
        Source: ImageSource{
            Type:      "base64",
            MediaType: mediaType,
            Data:      encoded,
        },
    }
}
```

#### Phase 3: TUI Clipboard (macOS)
**File**: `internal/ui/clipboard.go`

```go
func ReadClipboardImage() ([]byte, error) {
    cmd := exec.Command("osascript", "-e", "the clipboard as «class PNGf»")
    output, err := cmd.Output()
    // Parse and decode AppleScript output
}
```

### Benefits
- 📷 Explain screenshots
- 🐛 Debug UI errors visually
- 📊 Analyze diagrams/charts
- 🎨 Design feedback

---

## Quick Wins Summary

### Do This Week
1. **Enable auto-compaction** (5 min) - Just call `GetPrunedMessages()`
2. **Add debug logging** for context usage (15 min)
3. **Test long conversations** to verify compaction works

### Do This Month
1. **Session resuming CLI** - Basic `hex resume --last`
2. **AGENTS.md directory traversal** - Better than current approach
3. **Session picker TUI** - Nice UX improvement

### Do This Quarter
1. **MCP integration** - Future-proof tool system
2. **Image input** - Valuable for debugging
3. **Consider execpolicy DSL** - If users request it

---

## Testing Checklist

For each adopted practice:

- [ ] Unit tests pass
- [ ] Integration test with real API
- [ ] Manual TUI testing
- [ ] Works in tmux
- [ ] Works in raw terminal
- [ ] Error handling tested
- [ ] Documentation updated
- [ ] Example in README

---

## References

- Codex source: `/Users/harper/workspace/2389/agent-class/agents/codex`
- Audit doc: `CODEX_AUDIT.md`
- hex source: `/Users/harper/Public/src/2389/hex`
