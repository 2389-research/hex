# MCP Servers

## Overview

Model Context Protocol (MCP) is an extensibility protocol that allows Claude Code to interact with external systems, data sources, and services through standardized server implementations. MCP servers extend Claude Code's capabilities beyond its built-in tools.

## What is Model Context Protocol

MCP is an open protocol developed by Anthropic that provides a standardized way for AI assistants to:

- **Access Resources**: Read files, query databases, fetch web content
- **Execute Tools**: Run functions with defined inputs and outputs
- **Provide Context**: Supply structured data to enhance responses
- **Integrate Systems**: Connect to external services and APIs

MCP servers act as middleware between Claude Code and external systems, exposing capabilities through well-defined interfaces.

## How MCP Servers Extend Claude Code

MCP servers appear as additional tools in Claude Code's tool palette. When an MCP server is configured:

1. **Tool Discovery**: Claude Code queries the server for available tools
2. **Schema Registration**: Tool definitions (parameters, types, descriptions) are loaded
3. **Tool Invocation**: Claude can call these tools like built-in capabilities
4. **Response Handling**: Results are formatted and integrated into responses

### Tool Naming Convention

MCP tools follow the pattern: `mcp__<server-name>__<tool-name>`

Example: `mcp__playwright__browser_navigate`

## Configuration Concepts

MCP servers are defined through configuration objects that describe:

- **Transport** – stdio, HTTP, or SSE
- **Command** – the executable and arguments to launch the server
- **Environment** – key-value pairs for credentials or runtime toggles
- **Scopes** – whether the definition is global, user-specific, or project-specific

Claude Code merges these layers at startup, then negotiates capabilities with each enabled server. Availability therefore depends entirely on the current configuration; documentation should treat MCP access as dynamic.

## Capability Categories

Common MCP server families include:

- **Knowledge & Logging**: Chronicle-style activity logs, personal journals, status dashboards
- **Productivity**: Todo managers, calendar tools, documentation repositories
- **Browser Automation**: Playwright or Chrome controllers for UI testing
- **External APIs**: GitHub, Asana, Zendesk, Salesforce, etc.
- **Local Utilities**: Filesystem browsers, code search services, database shells

Each server exposes one or more tools following the `mcp__server__tool` naming pattern. Claude Code treats them like native tools, so workflows can mix built-in operations with MCP-powered capabilities seamlessly.

## Creating Custom MCP Servers

### Server Structure

An MCP server must implement the MCP protocol specification:

```javascript
// Basic Node.js MCP server structure
import { Server } from "@modelcontextprotocol/sdk/server/index.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";

const server = new Server({
  name: "my-server",
  version: "1.0.0",
}, {
  capabilities: {
    tools: {},
  },
});

// List available tools
server.setRequestHandler("tools/list", async () => {
  return {
    tools: [
      {
        name: "my_tool",
        description: "Description of what the tool does",
        inputSchema: {
          type: "object",
          properties: {
            param1: {
              type: "string",
              description: "First parameter",
            },
          },
          required: ["param1"],
        },
      },
    ],
  };
});

// Handle tool calls
server.setRequestHandler("tools/call", async (request) => {
  if (request.params.name === "my_tool") {
    const result = await executeMyTool(request.params.arguments);
    return {
      content: [
        {
          type: "text",
          text: JSON.stringify(result),
        },
      ],
    };
  }
  throw new Error(`Unknown tool: ${request.params.name}`);
});

// Start server
const transport = new StdioServerTransport();
await server.connect(transport);
```

### Python MCP Server

```python
from mcp.server import Server
from mcp.server.stdio import stdio_server
from mcp.types import Tool, TextContent

app = Server("my-python-server")

@app.list_tools()
async def list_tools() -> list[Tool]:
    return [
        Tool(
            name="my_tool",
            description="Description of the tool",
            inputSchema={
                "type": "object",
                "properties": {
                    "param1": {"type": "string"},
                },
                "required": ["param1"],
            },
        )
    ]

@app.call_tool()
async def call_tool(name: str, arguments: dict) -> list[TextContent]:
    if name == "my_tool":
        result = execute_my_tool(arguments["param1"])
        return [TextContent(type="text", text=str(result))]
    raise ValueError(f"Unknown tool: {name}")

async def main():
    async with stdio_server() as streams:
        await app.run(
            streams[0], streams[1], app.create_initialization_options()
        )

if __name__ == "__main__":
    import asyncio
    asyncio.run(main())
```

## Tool Definitions and Schemas

### Input Schema Format

Tool parameters are defined using JSON Schema:

```json
{
  "inputSchema": {
    "type": "object",
    "properties": {
      "url": {
        "type": "string",
        "description": "URL to fetch",
        "format": "uri"
      },
      "method": {
        "type": "string",
        "enum": ["GET", "POST", "PUT", "DELETE"],
        "description": "HTTP method"
      },
      "headers": {
        "type": "object",
        "additionalProperties": {
          "type": "string"
        },
        "description": "Request headers"
      }
    },
    "required": ["url"]
  }
}
```

### Schema Best Practices

1. **Descriptive Names**: Use clear, action-oriented tool names
2. **Detailed Descriptions**: Explain what the tool does and when to use it
3. **Required Fields**: Mark essential parameters as required
4. **Type Validation**: Use appropriate JSON Schema types
5. **Enums for Options**: Use enums for limited choice parameters
6. **Default Values**: Provide sensible defaults where applicable

### Return Value Format

Tools should return structured content:

```javascript
return {
  content: [
    {
      type: "text",
      text: "Result text or JSON string",
    },
  ],
  isError: false, // Optional: mark as error
};
```

## Resource Providers

MCP servers can also provide resources (not just tools):

```javascript
server.setRequestHandler("resources/list", async () => {
  return {
    resources: [
      {
        uri: "file:///path/to/resource",
        name: "Resource Name",
        description: "What this resource contains",
        mimeType: "text/plain",
      },
    ],
  };
});

server.setRequestHandler("resources/read", async (request) => {
  const uri = request.params.uri;
  const content = await readResource(uri);
  return {
    contents: [
      {
        uri: uri,
        mimeType: "text/plain",
        text: content,
      },
    ],
  };
});
```

## Real Examples from Current Session

### Example 1: Journal Writing

```javascript
// Tool invocation
await mcp__private_journal__process_thoughts({
  technical_insights: "Discovered that MCP servers use stdio transport",
  feelings: "Excited to build custom tools for automation",
  project_notes: "Need to document MCP integration patterns"
});
```

### Example 2: Browser Automation

```javascript
// Navigate to page
await mcp__playwright__browser_navigate({
  url: "https://example.com"
});

// Take snapshot
const snapshot = await mcp__playwright__browser_snapshot();

// Click element
await mcp__playwright__browser_click({
  element: "Submit button",
  ref: "button[type=submit]"
});
```

### Example 3: Todo Management

```javascript
// Create todo
await mcp__toki__add_todo({
  description: "Implement custom MCP server",
  priority: "high",
  tags: ["development", "mcp"],
  due_date: "2025-12-15T15:04:05Z"
});

// List todos
const todos = await mcp__toki__list_todos({
  done: false,
  priority: "high"
});
```

## Permission Management

MCP tool permissions are configured in `~/.claude/settings.json`:

```json
{
  "permissions": {
    "allow": [
      "mcp__private-journal__process_thoughts",
      "mcp__socialmedia__create_post",
      "mcp__chronicle__add_entry"
    ],
    "deny": [
      "mcp__dangerous_tool__*"
    ]
  }
}
```

## Debugging MCP Servers

### Server Logs

Most MCP servers write logs to stderr:

```bash
# Run server manually to see logs
node /path/to/server/index.js 2> server.log
```

### Testing Server Locally

```bash
# Test with MCP inspector
npx @modelcontextprotocol/inspector node /path/to/server/index.js
```

### Common Issues

1. **Server Not Starting**: Check command path and args in mcp.json
2. **Tools Not Appearing**: Verify server implements tools/list handler
3. **Permission Denied**: Check permissions in settings.json
4. **Tool Errors**: Review server logs for runtime errors

## Official MCP Servers

Anthropic provides reference implementations:

- **@modelcontextprotocol/server-filesystem**: File system access
- **@modelcontextprotocol/server-github**: GitHub API integration
- **@modelcontextprotocol/server-postgres**: PostgreSQL database
- **@modelcontextprotocol/server-puppeteer**: Browser automation
- **@modelcontextprotocol/server-sqlite**: SQLite database

Install via npm:

```bash
npm install -g @modelcontextprotocol/server-filesystem
```

Configure in mcp.json:

```json
{
  "servers": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "--root", "/workspace"]
    }
  }
}
```

## Best Practices

1. **Idempotent Operations**: Design tools to be safely retried
2. **Clear Error Messages**: Return helpful error descriptions
3. **Parameter Validation**: Validate inputs before execution
4. **Resource Cleanup**: Clean up resources on server shutdown
5. **Security**: Never expose sensitive data in tool responses
6. **Documentation**: Include usage examples in tool descriptions
7. **Versioning**: Use semantic versioning for server releases

## See Also

- [Slash Commands](10-SLASH-COMMANDS.md)
- [Configuration](11-CONFIGURATION.md)
- [Official MCP Documentation](https://modelcontextprotocol.io)
