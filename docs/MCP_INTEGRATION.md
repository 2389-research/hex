# MCP Integration Architecture

Complete technical reference for Clem's Model Context Protocol (MCP) integration.

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Components](#components)
- [Protocol Implementation](#protocol-implementation)
- [Tool Integration](#tool-integration)
- [Server Development](#server-development)
- [Advanced Topics](#advanced-topics)
- [Troubleshooting](#troubleshooting)

## Overview

### What is MCP?

MCP (Model Context Protocol) is an open standard created by Anthropic that enables AI assistants to securely connect to external data sources and tools. It provides:

- **Standardized protocol** for tool discovery and execution
- **Multiple transports** (stdio, HTTP/SSE)
- **Extensible capabilities** (tools, resources, prompts)
- **Language-agnostic** JSON-RPC 2.0 messaging

### Why Integrate MCP with Clem?

**Extensibility**: Add capabilities without modifying Clem source code

**Ecosystem**: Leverage community-built MCP servers

**Standardization**: Compatible with other MCP-enabled tools

**Security**: Sandboxed execution with explicit capability negotiation

### Integration Goals

1. Seamlessly integrate MCP tools alongside built-in tools
2. Support stdio transport (HTTP in future phases)
3. Provide simple CLI for server management
4. Enable graceful error handling and server lifecycle management

## Architecture

### High-Level Overview

```
┌─────────────────────────────────────────────────────────────┐
│                         Clem CLI                            │
│                                                             │
│  ┌──────────────────────────────────────────────────────┐  │
│  │              Tool Registry (Unified)                  │  │
│  │                                                        │  │
│  │  ├─ Built-in Tools (Read, Write, Bash, ...)          │  │
│  │  └─ MCP Tools (via MCPToolAdapter)                    │  │
│  └──────────────────────────────────────────────────────┘  │
│                          ▲                                  │
│                          │                                  │
│  ┌──────────────────────┼────────────────────────────────┐ │
│  │        MCP Tool Manager                               │ │
│  │                                                        │ │
│  │  Fetches tools from MCP Client                        │ │
│  │  Wraps each tool in MCPToolAdapter                    │ │
│  │  Registers with Clem Tool Registry                    │ │
│  └────────────────────────────────────────────────────────┘ │
│                          ▲                                  │
│                          │                                  │
│  ┌──────────────────────┼────────────────────────────────┐ │
│  │         MCP Client (JSON-RPC over stdio)              │ │
│  │                                                        │ │
│  │  ├─ Initialize handshake                              │ │
│  │  ├─ List tools (tools/list)                           │ │
│  │  └─ Call tools (tools/call)                           │ │
│  └────────────────────────────────────────────────────────┘ │
│                          ▲                                  │
│                          │ stdin/stdout                     │
└──────────────────────────┼──────────────────────────────────┘
                           │
                           ▼
              ┌────────────────────────┐
              │   MCP Server Process   │
              │                        │
              │  (Node, Python, etc.)  │
              └────────────────────────┘
```

### Component Responsibilities

**MCP Registry** (`internal/mcp/registry.go`):
- Manage server configurations
- Persist to `.mcp.json`
- CRUD operations (add, remove, list, update)

**MCP Client** (`internal/mcp/client.go`):
- JSON-RPC 2.0 communication
- Process lifecycle (spawn, communicate, cleanup)
- Protocol handshake and version negotiation
- Tool listing and execution

**MCP Tool Adapter** (`internal/mcp/tool_adapter.go`):
- Bridge MCP tools to Clem's Tool interface
- Result format conversion
- Error handling

**MCP Tool Manager** (`internal/mcp/tool_adapter.go`):
- Fetch tools from MCP client
- Manage multiple tool adapters
- Provide unified tool access

## Components

### 1. MCP Registry

**Purpose**: Manage MCP server configurations and persist to disk

**Key Types**:
```go
type ServerConfig struct {
    Name      string   `json:"name"`       // Unique server identifier
    Transport string   `json:"transport"`  // "stdio" (HTTP in future)
    Command   string   `json:"command"`    // Executable to launch
    Args      []string `json:"args"`       // Command-line arguments
}

type MCPConfig struct {
    Version string                  `json:"version"` // Config format version
    Servers map[string]ServerConfig `json:"servers"` // Server name -> config
}
```

**Operations**:
- `AddServer(server)` - Add new server configuration
- `RemoveServer(name)` - Remove server by name
- `GetServer(name)` - Retrieve server configuration
- `UpdateServer(server)` - Update existing server
- `ListServers()` - Get all servers
- `Save()` - Persist to `.mcp.json`
- `Load()` - Load from `.mcp.json`

**Thread Safety**: All operations are protected with `sync.RWMutex`

**Validation**:
- Server name must be non-empty and unique
- Command must be non-empty
- Transport must be "stdio" (in current phase)

### 2. MCP Client

**Purpose**: Communicate with MCP server processes via JSON-RPC 2.0

**Key Types**:
```go
type Client struct {
    cmd       *exec.Cmd          // Server process
    stdin     io.WriteCloser     // Write JSON-RPC requests
    stdout    io.ReadCloser      // Read JSON-RPC responses
    nextID    int32              // Atomic request ID counter
    pending   map[int]chan Response // Request ID -> response channel
}

type Tool struct {
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    InputSchema map[string]interface{} `json:"inputSchema"`
}
```

**Protocol Flow**:
1. **Initialization**:
   ```go
   client := mcp.NewClient("npx", "-y", "@modelcontextprotocol/server-filesystem", "/data")
   err := client.Initialize(ctx, "clem", "1.0.0", "2024-11-05")
   ```

2. **List Tools**:
   ```go
   tools, err := client.ListTools(ctx)
   // Returns: []Tool with name, description, inputSchema
   ```

3. **Call Tool**:
   ```go
   result, err := client.CallTool(ctx, "filesystem_read_file", map[string]interface{}{
       "path": "/data/config.json",
   })
   // Returns: map with "content" array
   ```

4. **Cleanup**:
   ```go
   err := client.Close()
   // Terminates server process
   ```

**Request ID Management**: Atomic counter ensures unique IDs for concurrent requests

**Response Correlation**: Pending map matches responses to waiting goroutines

### 3. MCP Tool Adapter

**Purpose**: Wrap MCP tools to implement Clem's Tool interface

**Key Type**:
```go
type MCPToolAdapter struct {
    client  *Client  // MCP client for tool execution
    mcpTool Tool     // Original MCP tool definition
}
```

**Tool Interface Implementation**:
```go
// Name returns tool identifier
func (a *MCPToolAdapter) Name() string

// Description returns human-readable description
func (a *MCPToolAdapter) Description() string

// RequiresApproval determines if execution needs user approval
func (a *MCPToolAdapter) RequiresApproval(params map[string]interface{}) bool

// Execute runs the tool and returns Result
func (a *MCPToolAdapter) Execute(ctx context.Context, params map[string]interface{}) (*tools.Result, error)

// GetInputSchema returns JSON Schema for parameters
func (a *MCPToolAdapter) GetInputSchema() map[string]interface{}
```

**Result Conversion**:

MCP result format:
```json
{
  "content": [
    {
      "type": "text",
      "text": "File contents here..."
    }
  ]
}
```

Clem result format:
```go
&tools.Result{
    ToolName: "filesystem_read_file",
    Success:  true,
    Output:   "File contents here...",
    Metadata: map[string]interface{}{
        "mcp_result": originalMCPResult,
    },
}
```

### 4. MCP Tool Manager

**Purpose**: Manage multiple MCP tool adapters from a server

**Key Type**:
```go
type MCPToolManager struct {
    client *Client                      // MCP client
    tools  map[string]*MCPToolAdapter  // Tool name -> adapter
    mu     sync.RWMutex                // Concurrent access protection
}
```

**Operations**:
- `RefreshTools(ctx)` - Fetch latest tools from server
- `GetTools()` - Return all tools as `[]tools.Tool`
- `GetTool(name)` - Get specific tool by name
- `Count()` - Number of available tools

**Workflow**:
```go
// Create manager for MCP server
manager := mcp.NewMCPToolManager(client)

// Fetch tools from server
err := manager.RefreshTools(ctx)

// Get all tools for registration
mcpTools := manager.GetTools()

// Register with Clem's tool registry
for _, tool := range mcpTools {
    clemRegistry.Register(tool)
}
```

## Protocol Implementation

### JSON-RPC 2.0 Messaging

**Request Format**:
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "method/name",
  "params": {
    "param": "value"
  }
}
```

**Response Format**:
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "data": "here"
  }
}
```

**Error Response**:
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "error": {
    "code": -32600,
    "message": "Invalid Request"
  }
}
```

**Notification** (no response expected):
```json
{
  "jsonrpc": "2.0",
  "method": "notifications/initialized"
}
```

### Initialize Handshake

**1. Client sends initialize request**:
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "initialize",
  "params": {
    "protocolVersion": "2024-11-05",
    "capabilities": {},
    "clientInfo": {
      "name": "clem",
      "version": "1.0.0"
    }
  }
}
```

**2. Server responds**:
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "protocolVersion": "2024-11-05",
    "capabilities": {
      "tools": {}
    },
    "serverInfo": {
      "name": "filesystem-server",
      "version": "1.0.0"
    }
  }
}
```

**3. Client sends initialized notification**:
```json
{
  "jsonrpc": "2.0",
  "method": "notifications/initialized"
}
```

**4. Connection is ready for operations**

### Tool Listing

**Request**:
```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "tools/list"
}
```

**Response**:
```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "result": {
    "tools": [
      {
        "name": "read_file",
        "description": "Read the complete contents of a file from the file system",
        "inputSchema": {
          "type": "object",
          "properties": {
            "path": {
              "type": "string",
              "description": "Path to the file to read"
            }
          },
          "required": ["path"]
        }
      }
    ]
  }
}
```

### Tool Execution

**Request**:
```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "method": "tools/call",
  "params": {
    "name": "read_file",
    "arguments": {
      "path": "/data/config.json"
    }
  }
}
```

**Response**:
```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "{\n  \"database\": \"postgres\",\n  \"port\": 5432\n}"
      }
    ]
  }
}
```

### Content Block Types

**Text Content**:
```json
{
  "type": "text",
  "text": "Content here..."
}
```

**Image Content** (future):
```json
{
  "type": "image",
  "data": "base64-encoded-image-data",
  "mimeType": "image/png"
}
```

**Resource Content** (future):
```json
{
  "type": "resource",
  "resource": {
    "uri": "file:///path/to/file",
    "mimeType": "application/json"
  }
}
```

## Tool Integration

### Loading MCP Tools at Startup

```go
// 1. Load server configurations
registry := mcp.NewRegistry(".")
if err := registry.Load(); err != nil {
    return err
}

// 2. Connect to each server
servers := registry.ListServers()
for _, serverConfig := range servers {
    // 3. Create client
    client, err := mcp.NewClient(serverConfig.Command, serverConfig.Args...)
    if err != nil {
        log.Printf("Failed to connect to %s: %v", serverConfig.Name, err)
        continue
    }

    // 4. Initialize handshake
    err = client.Initialize(ctx, "clem", "1.0.0", "2024-11-05")
    if err != nil {
        log.Printf("Failed to initialize %s: %v", serverConfig.Name, err)
        client.Close()
        continue
    }

    // 5. Create tool manager
    toolManager := mcp.NewMCPToolManager(client)
    if err := toolManager.RefreshTools(ctx); err != nil {
        log.Printf("Failed to load tools from %s: %v", serverConfig.Name, err)
        client.Close()
        continue
    }

    // 6. Register tools
    mcpTools := toolManager.GetTools()
    for _, tool := range mcpTools {
        // Add server name prefix to avoid collisions
        prefixedTool := mcp.WithPrefix(tool, serverConfig.Name)
        clemRegistry.Register(prefixedTool)
    }

    log.Printf("Loaded %d tools from %s", len(mcpTools), serverConfig.Name)
}
```

### Tool Naming Convention

To avoid name collisions between servers and built-in tools, MCP tools are prefixed with their server name:

**Example**:
- Server name: `filesystem`
- Tool name from server: `read_file`
- Registered name in Clem: `filesystem_read_file`

**Implementation**:
```go
func WithPrefix(tool tools.Tool, prefix string) tools.Tool {
    return &PrefixedToolAdapter{
        tool:   tool,
        prefix: prefix,
    }
}

func (p *PrefixedToolAdapter) Name() string {
    return p.prefix + "_" + p.tool.Name()
}
```

### Error Handling

**Connection Errors**:
```go
client, err := mcp.NewClient(command, args...)
if err != nil {
    // Server binary not found, failed to start process
    // Gracefully skip this server, log warning
    return fmt.Errorf("failed to connect: %w", err)
}
```

**Initialize Errors**:
```go
err := client.Initialize(ctx, "clem", "1.0.0", "2024-11-05")
if err != nil {
    // Protocol version mismatch, server error
    // Close client, skip server
    client.Close()
    return fmt.Errorf("handshake failed: %w", err)
}
```

**Tool Execution Errors**:
```go
result, err := client.CallTool(ctx, name, params)
if err != nil {
    // Network error, server crash, timeout
    // Return error result to Claude
    return &tools.Result{
        Success: false,
        Error:   err.Error(),
    }, nil
}
```

**Graceful Degradation**: If MCP servers fail to load, Clem continues with built-in tools only

## Server Development

### Creating a Custom MCP Server

#### Node.js Example

**1. Install MCP SDK**:
```bash
npm install @modelcontextprotocol/sdk
```

**2. Create server** (`my-server.js`):
```javascript
#!/usr/bin/env node
import { Server } from "@modelcontextprotocol/sdk/server/index.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";

// Define server
const server = new Server(
  {
    name: "my-custom-server",
    version: "1.0.0",
  },
  {
    capabilities: {
      tools: {},
    },
  }
);

// Register tools/list handler
server.setRequestHandler("tools/list", async () => ({
  tools: [
    {
      name: "greet",
      description: "Generate a greeting message",
      inputSchema: {
        type: "object",
        properties: {
          name: {
            type: "string",
            description: "Name to greet",
          },
        },
        required: ["name"],
      },
    },
  ],
}));

// Register tools/call handler
server.setRequestHandler("tools/call", async (request) => {
  const { name, arguments: args } = request.params;

  if (name === "greet") {
    return {
      content: [
        {
          type: "text",
          text: `Hello, ${args.name}! Welcome to MCP.`,
        },
      ],
    };
  }

  throw new Error(`Unknown tool: ${name}`);
});

// Start server
async function main() {
  const transport = new StdioServerTransport();
  await server.connect(transport);
  console.error("Server started successfully");
}

main().catch((error) => {
  console.error("Server error:", error);
  process.exit(1);
});
```

**3. Make executable**:
```bash
chmod +x my-server.js
```

**4. Test manually**:
```bash
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0.0"}}}' | node my-server.js
```

**5. Add to Clem**:
```bash
clem mcp add myserver node my-server.js
```

#### Python Example

**1. Install MCP SDK**:
```bash
pip install mcp
```

**2. Create server** (`my_server.py`):
```python
#!/usr/bin/env python3
import asyncio
import sys
from mcp.server import Server
from mcp.server.stdio import stdio_server

# Create server instance
server = Server("my-python-server")

@server.list_tools()
async def list_tools():
    return [
        {
            "name": "calculate",
            "description": "Perform basic arithmetic",
            "inputSchema": {
                "type": "object",
                "properties": {
                    "operation": {
                        "type": "string",
                        "enum": ["add", "subtract", "multiply", "divide"],
                    },
                    "a": {"type": "number"},
                    "b": {"type": "number"},
                },
                "required": ["operation", "a", "b"],
            },
        }
    ]

@server.call_tool()
async def call_tool(name: str, arguments: dict):
    if name == "calculate":
        a = arguments["a"]
        b = arguments["b"]
        op = arguments["operation"]

        if op == "add":
            result = a + b
        elif op == "subtract":
            result = a - b
        elif op == "multiply":
            result = a * b
        elif op == "divide":
            result = a / b if b != 0 else "Error: Division by zero"

        return {
            "content": [
                {
                    "type": "text",
                    "text": f"Result: {result}",
                }
            ]
        }

    raise ValueError(f"Unknown tool: {name}")

async def main():
    async with stdio_server() as (read_stream, write_stream):
        await server.run(
            read_stream,
            write_stream,
            server.create_initialization_options()
        )

if __name__ == "__main__":
    asyncio.run(main())
```

**3. Add to Clem**:
```bash
clem mcp add calc python my_server.py
```

### Best Practices for Server Development

**1. Input Validation**:
```javascript
server.setRequestHandler("tools/call", async (request) => {
  const { name, arguments: args } = request.params;

  // Validate required parameters
  if (!args.path) {
    throw new Error("Missing required parameter: path");
  }

  // Validate parameter types
  if (typeof args.path !== "string") {
    throw new Error("path must be a string");
  }

  // Proceed with execution
  // ...
});
```

**2. Error Handling**:
```javascript
server.setRequestHandler("tools/call", async (request) => {
  try {
    const result = await performOperation(request.params.arguments);
    return {
      content: [{ type: "text", text: result }]
    };
  } catch (error) {
    // Return error as content, not throw
    return {
      content: [
        {
          type: "text",
          text: `Error: ${error.message}`
        }
      ],
      isError: true
    };
  }
});
```

**3. Logging**:
```javascript
// Use stderr for logging (stdout is for JSON-RPC)
console.error("Tool executed:", name);
console.error("Parameters:", args);
```

**4. Resource Cleanup**:
```javascript
process.on("SIGTERM", async () => {
  console.error("Received SIGTERM, cleaning up...");
  await cleanup();
  process.exit(0);
});
```

**5. Schema Documentation**:
```javascript
{
  name: "read_file",
  description: "Read the complete contents of a file",
  inputSchema: {
    type: "object",
    properties: {
      path: {
        type: "string",
        description: "Absolute or relative path to the file"
      },
      encoding: {
        type: "string",
        description: "File encoding (default: utf-8)",
        default: "utf-8"
      }
    },
    required: ["path"]
  }
}
```

## Advanced Topics

### Concurrent Request Handling

The MCP client supports concurrent tool calls:

```go
// Each request gets a unique ID
id := atomic.AddInt32(&c.nextID, 1)

// Create response channel
respChan := make(chan Response, 1)

// Register in pending map
c.mu.Lock()
c.pending[id] = respChan
c.mu.Unlock()

// Send request and wait for response
// Multiple goroutines can do this simultaneously
```

### Process Lifecycle Management

**Startup**:
```go
cmd := exec.Command(command, args...)
cmd.Stdin = stdin
cmd.Stdout = stdout
cmd.Stderr = os.Stderr  // Logging visible to user

if err := cmd.Start(); err != nil {
    return nil, fmt.Errorf("failed to start process: %w", err)
}
```

**Graceful Shutdown**:
```go
func (c *Client) Close() error {
    // Close stdin to signal server to exit
    c.stdin.Close()

    // Wait for process to exit
    err := c.cmd.Wait()

    // Clean up resources
    c.mu.Lock()
    defer c.mu.Unlock()
    for _, ch := range c.pending {
        close(ch)
    }
    c.pending = nil

    return err
}
```

**Crash Detection**:
```go
// Monitor process in background goroutine
go func() {
    err := c.cmd.Wait()
    if err != nil {
        log.Printf("MCP server exited: %v", err)
        // Notify all pending requests
        c.mu.Lock()
        for _, ch := range c.pending {
            ch <- Response{Error: "server crashed"}
        }
        c.mu.Unlock()
    }
}()
```

### Context Cancellation

All operations respect context cancellation:

```go
func (c *Client) CallTool(ctx context.Context, name string, args map[string]interface{}) (map[string]interface{}, error) {
    // Send request
    respChan := c.sendRequest(...)

    // Wait for response or cancellation
    select {
    case resp := <-respChan:
        return resp.Result, resp.Error
    case <-ctx.Done():
        return nil, ctx.Err()
    }
}
```

### Future: HTTP Transport

**Configuration**:
```json
{
  "servers": {
    "api": {
      "name": "api",
      "transport": "http",
      "url": "http://localhost:8080/mcp",
      "headers": {
        "Authorization": "Bearer ${API_TOKEN}"
      }
    }
  }
}
```

**Implementation** (planned):
```go
type HTTPClient struct {
    baseURL    string
    httpClient *http.Client
    headers    map[string]string
}

func (c *HTTPClient) CallTool(ctx context.Context, name string, args map[string]interface{}) {
    // POST to baseURL/tools/call
    // Include headers
    // Return response
}
```

### Future: Resources Support

MCP resources provide read access to file-like data:

**List resources**:
```json
{
  "jsonrpc": "2.0",
  "id": 4,
  "method": "resources/list"
}
```

**Read resource**:
```json
{
  "jsonrpc": "2.0",
  "id": 5,
  "method": "resources/read",
  "params": {
    "uri": "file:///path/to/resource"
  }
}
```

### Future: Prompts Support

MCP prompts are pre-defined conversation templates:

**List prompts**:
```json
{
  "jsonrpc": "2.0",
  "id": 6,
  "method": "prompts/list"
}
```

**Get prompt**:
```json
{
  "jsonrpc": "2.0",
  "id": 7,
  "method": "prompts/get",
  "params": {
    "name": "code-review",
    "arguments": {
      "file": "main.go"
    }
  }
}
```

## Troubleshooting

### Debugging MCP Communication

**Enable verbose logging**:
```bash
PAGEN_DEBUG=1 clem
```

**Check server stderr**:
Server logs go to stderr, which Clem forwards to terminal:
```
[MCP:filesystem] Server started successfully
[MCP:filesystem] Tool called: read_file
[MCP:filesystem] Reading: /data/config.json
```

**Inspect .mcp.json**:
```bash
cat .mcp.json | jq
```

**Test server manually**:
```bash
# Send initialize request
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}' | npx -y @modelcontextprotocol/server-filesystem /data

# Send tools/list
echo '{"jsonrpc":"2.0","id":2,"method":"tools/list"}' | npx -y @modelcontextprotocol/server-filesystem /data
```

### Common Errors

#### "server process failed to start"

**Cause**: Command not found or not executable

**Solutions**:
- Verify command exists: `which node`
- Test command manually: `node server.js`
- Check file permissions: `chmod +x server.js`
- Use full path: `/usr/local/bin/node server.js`

#### "protocol version mismatch"

**Cause**: Server doesn't support requested MCP version

**Solutions**:
- Update server to latest version
- Check server documentation for supported versions
- Try older protocol version in handshake

#### "tool not found"

**Cause**: Tool name doesn't match server's tools

**Solutions**:
- List tools with `clem mcp list`
- Check for typos in tool name
- Verify tool is registered (check server logs)
- Remember server name prefix: `filesystem_read_file`, not `read_file`

#### "invalid JSON-RPC response"

**Cause**: Server returned malformed JSON or non-JSON output

**Solutions**:
- Check server stderr for errors
- Ensure server only writes JSON-RPC to stdout
- Verify server uses correct JSON-RPC 2.0 format
- Test server with manual JSON-RPC requests

#### "context deadline exceeded"

**Cause**: Tool execution took too long

**Solutions**:
- Check server is running and responsive
- Verify network connectivity (for HTTP transport)
- Investigate why tool is slow (check server logs)
- Increase timeout if operation legitimately takes long

## Reference

### Official MCP Specification

https://spec.modelcontextprotocol.io/

### Official MCP SDK

**TypeScript/JavaScript**:
```bash
npm install @modelcontextprotocol/sdk
```
https://github.com/modelcontextprotocol/typescript-sdk

**Python**:
```bash
pip install mcp
```
https://github.com/modelcontextprotocol/python-sdk

### Example Servers

**Filesystem**:
https://github.com/modelcontextprotocol/servers/tree/main/src/filesystem

**Fetch**:
https://github.com/modelcontextprotocol/servers/tree/main/src/fetch

**SQLite**:
https://github.com/modelcontextprotocol/servers/tree/main/src/sqlite

**PostgreSQL**:
https://github.com/modelcontextprotocol/servers/tree/main/src/postgres

## Appendix

### JSON-RPC 2.0 Error Codes

| Code | Message | Meaning |
|------|---------|---------|
| -32700 | Parse error | Invalid JSON |
| -32600 | Invalid Request | Invalid JSON-RPC format |
| -32601 | Method not found | Method doesn't exist |
| -32602 | Invalid params | Invalid method parameters |
| -32603 | Internal error | Server internal error |

### MCP Protocol Versions

| Version | Released | Notes |
|---------|----------|-------|
| 2024-11-05 | Nov 2024 | Initial stable release |

### Clem MCP Implementation Status

| Feature | Status | Version |
|---------|--------|---------|
| stdio transport | ✅ Complete | v0.3.0 |
| HTTP/SSE transport | 📋 Planned | v0.4.0 |
| Tools support | ✅ Complete | v0.3.0 |
| Resources support | 📋 Planned | v0.4.0 |
| Prompts support | 📋 Planned | v0.4.0 |
| Server lifecycle | 🔄 Basic | v0.3.0 |
| Auto-reconnect | 📋 Planned | v0.5.0 |
| Health checks | 📋 Planned | v0.5.0 |

---

**See Also**:
- [TOOLS.md](TOOLS.md) - User-facing tool documentation
- [ARCHITECTURE.md](ARCHITECTURE.md) - Overall Clem architecture
- [examples/mcp/](../examples/mcp/) - Working examples
