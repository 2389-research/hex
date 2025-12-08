# Hex - Go Implementation Design

**Date:** 2025-11-25
**Goal:** A powerful CLI for Claude AI inspired by Claude Code, Crush, Codex, and MaKeR
**Approach:** Modern Go implementation with real-world testing

---

## Executive Summary

Hex is a comprehensive command-line interface for Claude AI, drawing inspiration from Claude Code, Crush, Codex, and MaKeR. Built with modern Go practices, leveraging battle-tested libraries, and emphasizing real-world testing over mocks.

**Key Technologies:**
- **CLI:** Cobra for argument parsing
- **Config:** Viper for multi-source configuration
- **TUI:** Bubbletea + Charm.sh ecosystem
- **MCP:** Official go-sdk from Model Context Protocol
- **Plugins:** HashiCorp go-plugin for process isolation
- **Storage:** SQLite with JSON export capability
- **Testing:** VCR cassettes, real components, scenario-based

---

## Architecture Overview

### Project Structure

```
hex/
├── cmd/hex/              # Main entry point
│   └── main.go
├── internal/              # Private implementation
│   ├── core/              # Core types and API client
│   ├── tools/             # Tool implementations
│   ├── mcp/               # MCP runtime
│   ├── storage/           # SQLite + JSON export
│   ├── ui/                # Bubbletea UI
│   └── plugins/           # Plugin loader (go-plugin)
├── pkg/                   # Public APIs
│   └── plugin/            # Plugin interface
├── tests/
│   ├── unit/              # Fast, isolated tests
│   ├── scenarios/         # E2E workflow tests
│   └── golden/            # Output regression tests
├── go.mod
└── go.sum
```

### Design Principles

1. **Hybrid Architecture:** Monolithic binary with plugin extension points
2. **Stdlib First:** Use Go stdlib where excellent (http, json, flags basics)
3. **Best-of-Breed Libraries:** Proven libraries for complex domains (CLI, TUI, plugins)
4. **Real Testing:** VCR cassettes + real components, minimal mocks
5. **Security by Default:** Sandboxed tool execution, permission system

---

## Core Components

### 1. Anthropic API Client (`internal/core/client.go`)

**Responsibilities:**
- HTTP communication with Anthropic API
- Streaming and non-streaming message creation
- Automatic tool use handling
- Error handling and retries

**Interface:**
```go
type Client struct {
    apiKey     string
    httpClient *http.Client
    baseURL    string
}

func (c *Client) CreateMessage(ctx context.Context, req MessageRequest) (*MessageResponse, error)
func (c *Client) StreamMessage(ctx context.Context, req MessageRequest) (<-chan StreamChunk, error)

type MessageRequest struct {
    Model      string
    Messages   []Message
    Tools      []ToolDefinition
    MaxTokens  int
    Stream     bool
}
```

**Key Features:**
- Context-aware (cancellation, timeouts)
- Streaming via Go channels
- Tool definitions automatically injected
- Exponential backoff for rate limits

---

### 2. Tool System (`internal/tools/`)

**Design:** Hybrid approach with safety layer

**Tool Interface:**
```go
type Tool interface {
    Name() string
    Description() string
    InputSchema() map[string]interface{}
    Execute(ctx context.Context, input map[string]interface{}) (ToolResult, error)
}
```

**Built-in Tools:**

1. **ReadTool** - Pure Go file reading
   - Uses `os.ReadFile`
   - Path validation against allowed directories
   - Size limits

2. **WriteTool** - Pure Go file writing
   - Confirmation prompt (respects permission mode)
   - Atomic writes via temp file + rename
   - Creates parent directories

3. **EditTool** - Pure Go file editing
   - Find/replace operations
   - Line-based edits
   - Backup creation

4. **GrepTool** - Pure Go search
   - Uses `regexp` package
   - Recursive directory search
   - Context lines support

5. **BashTool** - Sandboxed shell execution
   - Uses `exec.Command` with restrictions
   - Working directory constraints
   - Environment variable filtering
   - Timeout enforcement

**Safety Executor:**
```go
type SafeExecutor struct {
    allowShell      bool
    allowedPaths    []string
    maxExecTime     time.Duration
    permissionMode  PermissionMode // ask, allow, deny, plan
}

func (e *SafeExecutor) Execute(ctx context.Context, tool Tool, input map[string]interface{}) (ToolResult, error) {
    // Validate permissions
    // Apply sandboxing
    // Execute with timeout
    // Audit logging
}
```

---

### 3. MCP Runtime (`internal/mcp/runtime.go`)

**Integration:** Uses official `github.com/modelcontextprotocol/go-sdk`

**Architecture:**
- One goroutine per MCP server
- Connection pooling for HTTP/SSE transports
- Automatic restart on crash (configurable)

```go
type Runtime struct {
    servers map[string]*MCPServer
    client  *mcpsdk.Client
}

type MCPServer struct {
    Name      string
    Transport string // stdio, http, sse
    Command   string
    Args      []string
    Env       map[string]string
    process   *exec.Cmd
    client    interface{} // StdioClient, HttpClient, or SSEClient
}

func (r *Runtime) StartServer(config ServerConfig) error
func (r *Runtime) StopServer(name string) error
func (r *Runtime) ListTools(serverName string) ([]ToolDefinition, error)
func (r *Runtime) CallTool(serverName, toolName string, input map[string]interface{}) (interface{}, error)
```

**Lifecycle:**
1. Server config loaded from `~/.hex/mcp.json`
2. Servers started on demand or at session start
3. Health checks every 30s
4. Graceful shutdown on SIGTERM
5. Automatic cleanup of zombie processes

---

### 4. Plugin System (`internal/plugins/`)

**Framework:** HashiCorp go-plugin

**Benefits:**
- Process isolation (plugin crashes don't crash Hex)
- Language agnostic (plugins can be in any language)
- gRPC communication
- Protocol versioning
- Battle-tested (used by Terraform, Vault, etc.)

**Plugin Interface:**
```go
// pkg/plugin/interface.go
type HexentPlugin interface {
    Name() string
    Version() string
    GetTools() []ToolDefinition
    GetCommands() []CommandDefinition
    ExecuteTool(name string, input map[string]interface{}) (interface{}, error)
    OnSessionStart(sessionID string) error
    OnSessionEnd(sessionID string) error
}
```

**Plugin Loading:**
```go
var Handshake = plugin.HandshakeConfig{
    ProtocolVersion:  1,
    MagicCookieKey:   "HEX_PLUGIN",
    MagicCookieValue: "hex",
}

type PluginLoader struct {
    client *plugin.Client
}

func (l *PluginLoader) LoadPlugin(path string) (HexentPlugin, error) {
    client := plugin.NewClient(&plugin.ClientConfig{
        HandshakeConfig: Handshake,
        Plugins:         PluginMap,
        Cmd:             exec.Command(path),
        AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
    })

    rpcClient, err := client.Client()
    raw, err := rpcClient.Dispense("hex")
    return raw.(HexentPlugin), nil
}
```

**Plugin Discovery:**
- Scan `~/.hex/plugins/` directory
- Load plugins marked as enabled in `~/.hex/config.yaml`
- Hot reload support (restart plugin process)

---

### 5. Storage Layer (`internal/storage/`)

**Database:** SQLite with JSON export

**Schema:**
```sql
-- conversations table
CREATE TABLE conversations (
    id TEXT PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    title TEXT,
    model TEXT
);

-- messages table
CREATE TABLE messages (
    id TEXT PRIMARY KEY,
    conversation_id TEXT NOT NULL,
    role TEXT NOT NULL, -- user, assistant
    content TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (conversation_id) REFERENCES conversations(id)
);

-- tool_calls table
CREATE TABLE tool_calls (
    id TEXT PRIMARY KEY,
    message_id TEXT NOT NULL,
    tool_name TEXT NOT NULL,
    input TEXT NOT NULL, -- JSON
    output TEXT, -- JSON
    status TEXT, -- pending, success, error
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (message_id) REFERENCES messages(id)
);
```

**Interface:**
```go
type Store struct {
    db *sql.DB
}

func (s *Store) CreateConversation() (string, error)
func (s *Store) SaveMessage(convID string, msg core.Message) error
func (s *Store) LoadConversation(convID string) ([]core.Message, error)
func (s *Store) ListConversations(limit int) ([]ConversationInfo, error)
func (s *Store) ExportJSON(convID string, path string) error
func (s *Store) ImportJSON(path string) error
func (s *Store) Migrate() error
```

**JSON Export Format:**
```json
{
  "id": "conv-abc123",
  "created_at": "2025-11-25T10:00:00Z",
  "title": "Conversation about Go",
  "model": "claude-sonnet-4-5-20250929",
  "messages": [
    {
      "id": "msg-1",
      "role": "user",
      "content": "Hello",
      "created_at": "2025-11-25T10:00:00Z"
    },
    {
      "id": "msg-2",
      "role": "assistant",
      "content": "Hi! How can I help?",
      "tool_calls": [],
      "created_at": "2025-11-25T10:00:01Z"
    }
  ]
}
```

---

### 6. Terminal UI (`internal/ui/`)

**Framework:** Bubbletea + Charm.sh ecosystem

**Libraries:**
- `bubbletea` - Main TUI framework
- `lipgloss` - Styling and layout
- `bubbles` - Reusable components (textinput, viewport)
- `glamour` - Markdown rendering

**Interactive Mode:**
```go
type InteractiveModel struct {
    client      *core.Client
    messages    []core.Message
    input       textinput.Model
    viewport    viewport.Model
    loading     bool
    streaming   bool
    streamBuf   strings.Builder
    width       int
    height      int
}

func (m InteractiveModel) Init() tea.Cmd {
    return textinput.Blink
}

func (m InteractiveModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+c":
            return m, tea.Quit
        case "enter":
            return m, m.sendMessage()
        }
    case StreamChunkMsg:
        m.streamBuf.WriteString(msg.Content)
        m.viewport.SetContent(m.renderMessages())
        if msg.Done {
            m.loading = false
            m.streaming = false
        }
        return m, nil
    }
    return m, nil
}

func (m InteractiveModel) View() string {
    // Render with lipgloss
    return lipgloss.JoinVertical(
        lipgloss.Left,
        m.renderHeader(),
        m.viewport.View(),
        m.renderInput(),
    )
}
```

**Print Mode:**
```go
type PrintMode struct {
    format     string // text, json, stream-json
    writer     io.Writer
}

func (p *PrintMode) WriteResponse(resp *core.MessageResponse) error {
    switch p.format {
    case "text":
        fmt.Fprintln(p.writer, resp.Content)
    case "json":
        json.NewEncoder(p.writer).Encode(resp)
    case "stream-json":
        // SSE format
        for chunk := range resp.Stream {
            fmt.Fprintf(p.writer, "data: %s\n\n", chunk.ToJSON())
            if f, ok := p.writer.(http.Flusher); ok {
                f.Flush()
            }
        }
    }
    return nil
}
```

**Streaming:**
- **Interactive mode:** Progressive rendering via bubbletea updates
- **Print mode:** SSE format (`data: {...}\n\n`) to stdout

---

### 7. CLI Interface (`cmd/hex/main.go`)

**Framework:** Cobra

**Root Command:**
```go
var rootCmd = &cobra.Command{
    Use:   "hex [prompt]",
    Short: "Hex - AI assistant CLI",
    Long:  "Hex is a powerful AI assistant for your terminal",
    Args:  cobra.ArbitraryArgs,
    Run:   runInteractive,
}

func init() {
    // Global flags
    rootCmd.PersistentFlags().BoolP("print", "p", false, "Print mode")
    rootCmd.PersistentFlags().String("output-format", "text", "Output format (text, json, stream-json)")
    rootCmd.PersistentFlags().StringP("model", "m", "claude-sonnet-4-5-20250929", "Model to use")
    rootCmd.PersistentFlags().BoolP("continue", "c", false, "Continue most recent conversation")
    rootCmd.PersistentFlags().StringP("resume", "r", "", "Resume specific conversation")
    rootCmd.PersistentFlags().String("system-prompt", "", "System prompt")
    rootCmd.PersistentFlags().String("json-schema", "", "JSON schema for structured output")
    rootCmd.PersistentFlags().StringSlice("tools", nil, "Available tools")
    rootCmd.PersistentFlags().StringSlice("allowed-tools", nil, "Allowed tools")
    rootCmd.PersistentFlags().StringSlice("disallowed-tools", nil, "Disallowed tools")
    rootCmd.PersistentFlags().String("permission-mode", "default", "Permission mode")
    rootCmd.PersistentFlags().Bool("verbose", false, "Verbose output")
    rootCmd.PersistentFlags().String("debug", "", "Debug categories")

    // Subcommands
    rootCmd.AddCommand(mcpCmd)
    rootCmd.AddCommand(pluginCmd)
    rootCmd.AddCommand(doctorCmd)
    rootCmd.AddCommand(setupCmd)
    rootCmd.AddCommand(exportCmd)
}
```

**Subcommands:**
```go
// hex mcp
var mcpCmd = &cobra.Command{
    Use:   "mcp",
    Short: "Manage MCP servers",
}
mcpCmd.AddCommand(mcpListCmd, mcpAddCmd, mcpRemoveCmd, mcpGetCmd)

// hex plugin
var pluginCmd = &cobra.Command{
    Use:   "plugin",
    Short: "Manage plugins",
}
pluginCmd.AddCommand(pluginListCmd, pluginInstallCmd, pluginEnableCmd, pluginDisableCmd)

// hex doctor
var doctorCmd = &cobra.Command{
    Use:   "doctor",
    Short: "Check installation health",
    Run:   runDoctor,
}

// hex setup-token
var setupCmd = &cobra.Command{
    Use:   "setup-token [token]",
    Short: "Configure API token",
    Run:   runSetup,
}

// hex export
var exportCmd = &cobra.Command{
    Use:   "export",
    Short: "Export conversation to JSON",
    Run:   runExport,
}
```

---

### 8. Configuration (`internal/core/config.go`)

**Framework:** Viper

**Config Structure:**
```go
type Config struct {
    APIKey         string                    `mapstructure:"api_key"`
    Model          string                    `mapstructure:"model"`
    DefaultTools   []string                  `mapstructure:"default_tools"`
    PermissionMode string                    `mapstructure:"permission_mode"`
    MCPServers     map[string]MCPServerConfig `mapstructure:"mcp_servers"`
    Plugins        map[string]PluginConfig    `mapstructure:"plugins"`
}

type MCPServerConfig struct {
    Transport string            `mapstructure:"transport"`
    Command   string            `mapstructure:"command"`
    Args      []string          `mapstructure:"args"`
    Env       map[string]string `mapstructure:"env"`
    Enabled   bool              `mapstructure:"enabled"`
}

type PluginConfig struct {
    Path    string `mapstructure:"path"`
    Enabled bool   `mapstructure:"enabled"`
}
```

**Configuration Precedence (highest to lowest):**
1. Command-line flags
2. Environment variables (HEX_*)
3. `.env` file (current directory)
4. `~/.hex/config.yaml`
5. Defaults

**Example config.yaml:**
```yaml
api_key: sk-ant-api03-...
model: claude-sonnet-4-5-20250929
default_tools:
  - Bash
  - Read
  - Write
  - Edit
  - Grep
permission_mode: ask

mcp_servers:
  filesystem:
    transport: stdio
    command: npx
    args: [-y, @modelcontextprotocol/server-filesystem, /Users/harper]
    enabled: true

  github:
    transport: http
    url: https://api.github.com/mcp
    enabled: false

plugins:
  myplugin:
    path: ~/.hex/plugins/myplugin
    enabled: true
```

---

## Testing Strategy

### Philosophy: Real Components, Minimal Mocks

**Guiding Principles:**
1. Mocks are lies that hide bugs
2. Test real behavior with real data
3. Use VCR cassettes for API calls
4. Only mock external services we don't control

### 1. Unit Tests (`tests/unit/`)

**Coverage:**
- Individual functions and methods
- Pure logic (no I/O)
- Error handling paths

**Example:**
```go
func TestToolValidation(t *testing.T) {
    tests := []struct {
        name    string
        input   map[string]interface{}
        wantErr bool
    }{
        {"valid", map[string]interface{}{"file_path": "test.txt"}, false},
        {"missing_path", map[string]interface{}{}, true},
        {"invalid_path", map[string]interface{}{"file_path": "../../../etc/passwd"}, true},
    }

    tool := tools.NewReadTool()
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tool.Validate(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### 2. Scenario Tests (`tests/scenarios/`)

**Structure:**
- Each scenario is a markdown file describing the workflow
- Corresponding `_test.go` file implements the test
- Uses real components (DB, filesystem)
- Uses VCR cassettes for API calls

**Example Scenario:**
```markdown
# Scenario: Basic Tool Execution

## Given
- User starts a new conversation
- User has Read tool enabled

## When
- User sends: "Read the file test.txt"
- Hex uses Read tool to read test.txt
- Hex responds with file contents

## Then
- File is read successfully
- Response contains file contents
- Tool call is logged in database
```

**Test Implementation:**
```go
// tests/scenarios/02-tool-execution_test.go
func TestScenario_ToolExecution(t *testing.T) {
    // Setup: Real filesystem
    tmpDir := t.TempDir()
    testFile := filepath.Join(tmpDir, "test.txt")
    os.WriteFile(testFile, []byte("Hello, Hex!"), 0644)

    // Setup: VCR for API calls
    r, err := recorder.New("fixtures/tool_execution")
    require.NoError(t, err)
    defer r.Stop()

    // Setup: Real database
    dbPath := filepath.Join(tmpDir, "test.db")
    store, err := storage.NewStore(dbPath)
    require.NoError(t, err)

    // Setup: Real client with VCR
    httpClient := &http.Client{Transport: r}
    client := core.NewClient(testAPIKey, core.WithHTTPClient(httpClient))

    // Setup: Real tool with test directory
    readTool := tools.NewReadTool(tools.WithBasePath(tmpDir))

    // Execute: Real conversation flow
    convID, err := store.CreateConversation()
    require.NoError(t, err)

    req := core.MessageRequest{
        Model:    "claude-sonnet-4-5-20250929",
        Messages: []core.Message{{Role: "user", Content: "Read test.txt"}},
        Tools:    []core.ToolDefinition{readTool.Definition()},
    }

    resp, err := client.CreateMessage(context.Background(), req)
    require.NoError(t, err)

    // Verify: Tool was called
    assert.Len(t, resp.ToolCalls, 1)
    assert.Equal(t, "Read", resp.ToolCalls[0].Name)

    // Execute tool
    result, err := readTool.Execute(context.Background(), resp.ToolCalls[0].Input)
    require.NoError(t, err)
    assert.Contains(t, result.Content, "Hello, Hex!")

    // Verify: Conversation saved to real DB
    messages, err := store.LoadConversation(convID)
    require.NoError(t, err)
    assert.Len(t, messages, 2) // user message + assistant response
}
```

### 3. Golden Tests (`tests/golden/`)

**Purpose:** Detect regressions in output format

**Approach:**
- Run command with known inputs
- Compare output to golden file
- Update golden files with `--update` flag

**Example:**
```go
func TestGolden_HelpOutput(t *testing.T) {
    cmd := exec.Command("hex", "--help")
    output, err := cmd.CombinedOutput()
    require.NoError(t, err)

    goldenFile := "testdata/help.golden"

    if *update {
        os.WriteFile(goldenFile, output, 0644)
        t.Log("Updated golden file")
        return
    }

    expected, err := os.ReadFile(goldenFile)
    require.NoError(t, err)

    assert.Equal(t, string(expected), string(output))
}

func TestGolden_ConversationExport(t *testing.T) {
    // Setup: Real conversation in real DB
    store := setupTestStore(t)
    convID := createTestConversation(store, t)

    // Export to JSON
    var buf bytes.Buffer
    err := store.ExportJSON(convID, &buf)
    require.NoError(t, err)

    // Compare to golden
    goldentest.Assert(t, "conversation_export.golden", buf.Bytes())
}
```

### 4. Integration Tests

**Purpose:** Test full system with real components

**When to run:**
- Not in short mode (`go test -short` skips these)
- Requires environment setup
- May use real API (with rate limiting)

**Example:**
```go
func TestIntegration_FullWorkflow(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    // Real everything
    tmpDir := t.TempDir()

    // Real MCP server (test implementation)
    mcpServer := startTestMCPServer(t)
    defer mcpServer.Stop()

    // Real database
    store, _ := storage.NewStore(filepath.Join(tmpDir, "test.db"))

    // Real API (with VCR or actual if ANTHROPIC_API_KEY set)
    var httpClient *http.Client
    if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
        httpClient = http.DefaultClient
    } else {
        r, _ := recorder.New("fixtures/full_workflow")
        defer r.Stop()
        httpClient = &http.Client{Transport: r}
    }

    client := core.NewClient(testAPIKey, core.WithHTTPClient(httpClient))

    // Real MCP runtime
    runtime := mcp.NewRuntime()
    err := runtime.StartServer(mcp.ServerConfig{
        Name:      "test",
        Transport: "stdio",
        Command:   mcpServer.Path(),
    })
    require.NoError(t, err)
    defer runtime.StopServer("test")

    // Execute full conversation with tools and MCP
    convID, _ := store.CreateConversation()

    // Message 1: Use MCP tool
    resp1, err := sendMessageWithMCP(client, runtime, "List files")
    require.NoError(t, err)
    store.SaveMessage(convID, resp1)

    // Message 2: Use built-in tool
    resp2, err := sendMessageWithTool(client, "Read file.txt")
    require.NoError(t, err)
    store.SaveMessage(convID, resp2)

    // Verify complete workflow
    messages, _ := store.LoadConversation(convID)
    assert.Len(t, messages, 4) // 2 user + 2 assistant
}
```

### Test Matrix Summary

```
┌─────────────────┬──────────────┬──────────────┬─────────────┬──────────┐
│ Test Type       │ API          │ Filesystem   │ Database    │ MCP      │
├─────────────────┼──────────────┼──────────────┼─────────────┼──────────┤
│ Unit            │ N/A          │ Real (tmpdir)│ N/A         │ N/A      │
│ Scenario        │ VCR Cassette │ Real (tmpdir)│ Real (temp) │ Mock     │
│ Golden          │ VCR Cassette │ Real (tmpdir)│ Real (temp) │ N/A      │
│ Integration     │ Real/VCR     │ Real         │ Real        │ Real     │
└─────────────────┴──────────────┴──────────────┴─────────────┴──────────┘
```

---

## Dependencies

### Core Dependencies

```go
require (
    github.com/spf13/cobra v1.8.0              // CLI framework
    github.com/spf13/viper v1.18.2             // Configuration
    github.com/charmbracelet/bubbletea v0.25.0 // TUI framework
    github.com/charmbracelet/lipgloss v0.9.1   // TUI styling
    github.com/charmbracelet/bubbles v0.18.0   // TUI components
    github.com/charmbracelet/glamour v0.6.0    // Markdown rendering
    github.com/hashicorp/go-plugin v1.6.0      // Plugin system
    github.com/modelcontextprotocol/go-sdk     // MCP integration
    github.com/mattn/go-sqlite3 v1.14.19       // SQLite driver
    github.com/joho/godotenv v1.5.1            // .env file support
)
```

### Testing Dependencies

```go
require (
    github.com/stretchr/testify v1.8.4         // Assertions
    github.com/dnaeon/go-vcr/v2 v2.3.0         // API recording/replay
    github.com/rogpeppe/go-internal v1.12.0    // Golden file testing
)
```

---

## Implementation Phases

### Phase 1: Foundation (Week 1)
- Project setup (go.mod, structure)
- CLI framework (Cobra)
- Configuration (Viper)
- API client (basic, non-streaming)
- Print mode (text output)

**Deliverable:** `hex --print "hello"` works

### Phase 2: Core Features (Week 2-3)
- Interactive UI (Bubbletea)
- Streaming support
- Storage layer (SQLite)
- Basic tool system (Read, Write)
- Conversation history (--continue, --resume)

**Deliverable:** Full interactive chat with history

### Phase 3: Tools & Safety (Week 4)
- Complete tool suite (Bash, Edit, Grep)
- Safety executor
- Permission system
- Tool testing (scenario-based)

**Deliverable:** Full tool execution with safety

### Phase 4: MCP Integration (Week 5)
- MCP runtime
- Server lifecycle management
- Transport implementations (stdio, http, sse)
- MCP commands (list, add, remove)

**Deliverable:** MCP server integration working

### Phase 5: Plugin System (Week 6)
- go-plugin integration
- Plugin interface
- Plugin loader
- Plugin commands

**Deliverable:** Can load and use plugins

### Phase 6: Advanced Features (Week 7-8)
- Structured output (--json-schema)
- Advanced CLI options
- Export/import
- Polish and optimization

**Deliverable:** Feature parity with original

### Phase 7: Testing & Documentation (Week 9-10)
- Comprehensive test suite
- Golden tests
- Documentation
- Examples

**Deliverable:** Production-ready

---

## Success Criteria

### Functional Parity
- [ ] All CLI flags supported
- [ ] All subcommands implemented
- [ ] Tool execution matches original
- [ ] MCP integration works
- [ ] Plugin system functional
- [ ] Streaming responses
- [ ] Conversation history
- [ ] Structured output

### Quality Metrics
- [ ] >80% test coverage
- [ ] All scenario tests passing
- [ ] Golden tests validate output
- [ ] No data races (go test -race)
- [ ] Performance comparable to original
- [ ] Binary size <50MB

### User Experience
- [ ] Fast startup (<100ms)
- [ ] Responsive UI (60fps in interactive)
- [ ] Clear error messages
- [ ] Help documentation complete
- [ ] Works on macOS, Linux, Windows

---

## Risk Mitigation

### Technical Risks

**Risk:** MCP Go SDK is immature
- **Mitigation:** Abstract SDK behind interface, can swap implementation
- **Fallback:** Implement MCP protocol directly

**Risk:** Bubbletea performance with large conversations
- **Mitigation:** Virtualized scrolling, message pagination
- **Fallback:** Fall back to simple text output

**Risk:** SQLite concurrency issues
- **Mitigation:** Single writer goroutine, proper locking
- **Fallback:** Use write-ahead log (WAL) mode

**Risk:** Plugin crashes affect main process
- **Mitigation:** go-plugin isolates in separate process
- **Fallback:** Automatic restart, circuit breaker

### Schedule Risks

**Risk:** Underestimating complexity
- **Mitigation:** Iterative development, MVP per phase
- **Response:** Cut advanced features if needed

**Risk:** Testing takes longer than expected
- **Mitigation:** Write tests alongside code, use subagents
- **Response:** Prioritize scenario tests over unit tests

---

## Conclusion

This design provides a solid foundation for a production-quality Go implementation of Claude Code CLI. Key strengths:

1. **Modern Go practices:** Leverages stdlib + best-of-breed libraries
2. **Battle-tested patterns:** go-plugin, Cobra, Bubbletea
3. **Real-world testing:** VCR cassettes, real components
4. **Extensible architecture:** Plugin system, MCP integration
5. **Security first:** Sandboxed execution, permission system

The phased approach allows incremental delivery of value while managing complexity.
