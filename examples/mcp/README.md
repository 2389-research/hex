# MCP Integration Examples

Practical examples for using MCP (Model Context Protocol) servers with Hex.

## Table of Contents

- [Quick Start](#quick-start)
- [Example 1: Filesystem Server](#example-1-filesystem-server)
- [Example 2: Fetch Server](#example-2-fetch-server)
- [Example 3: SQLite Database](#example-3-sqlite-database)
- [Example 4: Custom Weather Server](#example-4-custom-weather-server)
- [Common Use Cases](#common-use-cases)
- [Configuration Templates](#configuration-templates)

## Quick Start

### 1. Install an MCP Server

```bash
# Official filesystem server
npm install -g @modelcontextprotocol/server-filesystem

# Official fetch server (for web requests)
npm install -g @modelcontextprotocol/server-fetch

# Official SQLite server
npm install -g @modelcontextprotocol/server-sqlite
```

### 2. Configure Hex

```bash
# Add server to Hex
hex mcp add filesystem npx -y @modelcontextprotocol/server-filesystem ~/Documents

# Verify it's configured
hex mcp list
```

### 3. Use in Conversation

```bash
# Start Hex
hex

# MCP tools are automatically loaded
# Ask Claude to use them:
> "List all markdown files in my Documents directory"
```

## Example 1: Filesystem Server

**Purpose**: Safe file operations within allowed directories

### Setup

```bash
# Install
npm install -g @modelcontextprotocol/server-filesystem

# Add to Hex (allow access to project directory)
hex mcp add filesystem npx -y @modelcontextprotocol/server-filesystem $(pwd)

# Or allow multiple directories
hex mcp add filesystem npx -y @modelcontextprotocol/server-filesystem ~/Documents ~/Projects
```

### Available Tools

- `filesystem_read_file` - Read file contents
- `filesystem_write_file` - Write file contents
- `filesystem_list_directory` - List directory contents
- `filesystem_create_directory` - Create directories
- `filesystem_move_file` - Move or rename files
- `filesystem_search_files` - Search for files by pattern
- `filesystem_get_file_info` - Get file metadata

### Example Conversations

**Read a configuration file**:
```
You: "Read the package.json and tell me what dependencies we use"

Claude: [Uses filesystem_read_file]

Claude: "Based on package.json, your project uses:
- React 18.2.0 for UI
- TypeScript 5.0.4 for type safety
- Vite 4.3.9 for building
..."
```

**Find files by pattern**:
```
You: "Find all TypeScript test files"

Claude: [Uses filesystem_search_files]

Claude: "I found 23 TypeScript test files:
- src/components/Button.test.tsx
- src/utils/validation.test.ts
- src/api/client.test.ts
..."
```

**Organize files**:
```
You: "Move all .log files to a 'logs' directory"

Claude: [Uses filesystem_create_directory, filesystem_move_file]

Claude: "I've created a 'logs' directory and moved 8 log files:
- Moved app.log -> logs/app.log
- Moved error.log -> logs/error.log
..."
```

### Security Notes

- Server only accesses directories specified in configuration
- Cannot access files outside allowed directories
- No privilege escalation or system file access
- Read/write operations are sandboxed

## Example 2: Fetch Server

**Purpose**: Make HTTP requests to web APIs and websites

### Setup

```bash
# Install
npm install -g @modelcontextprotocol/server-fetch

# Add to Hex
hex mcp add fetch npx -y @modelcontextprotocol/server-fetch
```

### Available Tools

- `fetch_fetch` - GET request to URL
- `fetch_post` - POST request with body
- `fetch_get_json` - Fetch and parse JSON

### Example Conversations

**Check API status**:
```
You: "What's the latest version of the Anthropic Python SDK?"

Claude: [Uses fetch_get_json]

Claude: "The latest version is 0.18.1, released on March 15, 2024.
It includes:
- New streaming helpers
- Improved error handling
- Updated types for Claude 3 models
..."
```

**Fetch documentation**:
```
You: "Fetch the React hooks documentation and summarize the useState hook"

Claude: [Uses fetch_fetch]

Claude: "Based on the official React documentation, useState is a Hook that:

1. Lets you add state to functional components
2. Returns an array with [currentState, updateFunction]
3. Can be called multiple times for multiple state variables
..."
```

**Monitor service health**:
```
You: "Check if GitHub is up"

Claude: [Uses fetch_fetch]

Claude: "GitHub is operational. The status page shows:
- API: All systems operational
- Git Operations: All systems operational
- Last incident: None in the past 7 days
"
```

### Configuration Options

**With custom headers**:
```json
{
  "name": "fetch",
  "transport": "stdio",
  "command": "npx",
  "args": [
    "-y",
    "@modelcontextprotocol/server-fetch",
    "--user-agent",
    "Hex/1.0"
  ]
}
```

**Note**: Header configuration support varies by server version. Check server documentation.

## Example 3: SQLite Database

**Purpose**: Query and modify SQLite databases

### Setup

```bash
# Install
npm install -g @modelcontextprotocol/server-sqlite

# Add to Hex (specify database file)
hex mcp add database npx -y @modelcontextprotocol/server-sqlite ~/myapp.db
```

### Available Tools

- `database_query` - Execute SELECT queries
- `database_execute` - Execute INSERT/UPDATE/DELETE
- `database_list_tables` - List all tables
- `database_describe_table` - Show table schema

### Example Conversations

**Explore database schema**:
```
You: "What tables are in the database?"

Claude: [Uses database_list_tables]

Claude: "The database contains 5 tables:
- users (4 columns)
- posts (6 columns)
- comments (5 columns)
- tags (2 columns)
- post_tags (2 columns)
"
```

**Query data**:
```
You: "Show me the 5 most recent posts"

Claude: [Uses database_query]

Claude: "Here are the 5 most recent posts:

1. 'Introducing MCP Support' by Alice (2024-03-15)
2. 'Building CLI Tools in Go' by Bob (2024-03-14)
3. 'TypeScript Best Practices' by Charlie (2024-03-13)
..."
```

**Analyze data**:
```
You: "Which user has posted the most comments?"

Claude: [Uses database_query]

Claude: "Based on the database:

Alice has posted the most comments with 47 total:
- 12 on 'Introducing MCP Support'
- 8 on 'Building CLI Tools in Go'
- 27 on other posts

This is 3x more than the next most active commenter (Bob with 15).
"
```

### Security Notes

- Read-only mode available: pass `--readonly` flag
- Transactions are supported
- No access to other databases on system
- SQL injection protection via parameterized queries

## Example 4: Custom Weather Server

**Purpose**: Demonstrate building a custom MCP server

### Server Implementation (Node.js)

**weather-server.js**:
```javascript
#!/usr/bin/env node
import { Server } from "@modelcontextprotocol/sdk/server/index.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";

const server = new Server(
  {
    name: "weather-server",
    version: "1.0.0",
  },
  {
    capabilities: {
      tools: {},
    },
  }
);

// Mock weather data (in production, call real API)
const weatherData = {
  "San Francisco": { temp: 68, condition: "Partly Cloudy" },
  "New York": { temp: 72, condition: "Sunny" },
  "London": { temp: 55, condition: "Rainy" },
};

server.setRequestHandler("tools/list", async () => ({
  tools: [
    {
      name: "get_weather",
      description: "Get current weather for a city",
      inputSchema: {
        type: "object",
        properties: {
          city: {
            type: "string",
            description: "City name (e.g., 'San Francisco')",
          },
        },
        required: ["city"],
      },
    },
  ],
}));

server.setRequestHandler("tools/call", async (request) => {
  const { name, arguments: args } = request.params;

  if (name === "get_weather") {
    const city = args.city;
    const weather = weatherData[city];

    if (!weather) {
      return {
        content: [
          {
            type: "text",
            text: `Weather data not available for ${city}`,
          },
        ],
      };
    }

    return {
      content: [
        {
          type: "text",
          text: `Weather in ${city}: ${weather.temp}°F, ${weather.condition}`,
        },
      ],
    };
  }

  throw new Error(`Unknown tool: ${name}`);
});

async function main() {
  const transport = new StdioServerTransport();
  await server.connect(transport);
  console.error("Weather server started");
}

main().catch(console.error);
```

### Setup

```bash
# Make executable
chmod +x weather-server.js

# Add to Hex
hex mcp add weather node weather-server.js
```

### Usage

```
You: "What's the weather in San Francisco?"

Claude: [Uses weather_get_weather]

Claude: "The current weather in San Francisco is 68°F and Partly Cloudy."
```

## Common Use Cases

### Use Case 1: Code Analysis Across Files

**Setup**: Filesystem server

```bash
hex mcp add filesystem npx -y @modelcontextprotocol/server-filesystem $(pwd)
```

**Conversation**:
```
You: "Analyze all TypeScript files and identify common patterns in error handling"

Claude: [Uses filesystem_search_files to find .ts files]
Claude: [Uses filesystem_read_file on each file]

Claude: "I've analyzed 45 TypeScript files. Here are the error handling patterns:

1. Try-catch blocks (78% of files)
   - Most common in async functions
   - Usually log and re-throw

2. Result<T, E> types (23% of files)
   - Used in functional programming modules
   - Avoids exceptions

3. Error boundaries (12% of files)
   - React components only
   - Fallback UI provided
..."
```

### Use Case 2: API Integration Testing

**Setup**: Fetch server

```bash
hex mcp add fetch npx -y @modelcontextprotocol/server-fetch
```

**Conversation**:
```
You: "Test our API endpoints:
- GET /api/health
- GET /api/users/1
- POST /api/users with sample data"

Claude: [Uses fetch_fetch for GET requests]
Claude: [Uses fetch_post for POST request]

Claude: "API test results:

✅ GET /api/health
   Status: 200
   Response: {'status': 'ok', 'version': '1.2.3'}

✅ GET /api/users/1
   Status: 200
   Response: User 'alice' found

❌ POST /api/users
   Status: 400
   Error: Missing required field 'email'

Recommendation: Update API documentation to clearly mark 'email' as required.
"
```

### Use Case 3: Database Migration Verification

**Setup**: SQLite server

```bash
hex mcp add db npx -y @modelcontextprotocol/server-sqlite app.db
```

**Conversation**:
```
You: "Verify the database migration to v5 completed successfully:
1. Check 'users' table has 'email_verified' column
2. Ensure all existing users have email_verified=false
3. Verify index on email_verified exists"

Claude: [Uses database_describe_table]
Claude: [Uses database_query to check data]
Claude: [Uses database_query to verify index]

Claude: "Migration verification results:

✅ 'email_verified' column exists (type: BOOLEAN, default: false)
✅ All 1,247 existing users have email_verified=false
✅ Index 'idx_users_email_verified' exists and is valid

Migration to v5 completed successfully!
"
```

### Use Case 4: Documentation Generation

**Setup**: Filesystem + Fetch servers

```bash
hex mcp add filesystem npx -y @modelcontextprotocol/server-filesystem $(pwd)
hex mcp add fetch npx -y @modelcontextprotocol/server-fetch
```

**Conversation**:
```
You: "Generate API documentation:
1. Read all route files in src/routes/
2. Extract endpoint definitions, parameters, responses
3. Fetch OpenAPI spec template from our wiki
4. Create docs/API.md with complete documentation"

Claude: [Uses filesystem_search_files to find routes]
Claude: [Uses filesystem_read_file on each route]
Claude: [Uses fetch_fetch to get template]
Claude: [Uses filesystem_write_file to create docs]

Claude: "I've generated comprehensive API documentation in docs/API.md:

- Documented 23 endpoints across 8 route files
- Included request/response examples for each
- Added authentication requirements
- Listed all query parameters and their types

The documentation follows your OpenAPI template format.
"
```

## Configuration Templates

### Minimal .mcp.json

```json
{
  "version": "1.0",
  "servers": {
    "filesystem": {
      "name": "filesystem",
      "transport": "stdio",
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/home/user/project"]
    }
  }
}
```

### Multi-Server Configuration

```json
{
  "version": "1.0",
  "servers": {
    "filesystem": {
      "name": "filesystem",
      "transport": "stdio",
      "command": "npx",
      "args": [
        "-y",
        "@modelcontextprotocol/server-filesystem",
        "/home/user/Documents",
        "/home/user/Projects"
      ]
    },
    "fetch": {
      "name": "fetch",
      "transport": "stdio",
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-fetch"]
    },
    "database": {
      "name": "database",
      "transport": "stdio",
      "command": "npx",
      "args": [
        "-y",
        "@modelcontextprotocol/server-sqlite",
        "/home/user/app.db"
      ]
    }
  }
}
```

### Development Setup

```json
{
  "version": "1.0",
  "servers": {
    "local-files": {
      "name": "local-files",
      "transport": "stdio",
      "command": "npx",
      "args": [
        "-y",
        "@modelcontextprotocol/server-filesystem",
        "/home/user/project/src",
        "/home/user/project/test"
      ]
    },
    "api-fetch": {
      "name": "api-fetch",
      "transport": "stdio",
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-fetch"]
    },
    "test-db": {
      "name": "test-db",
      "transport": "stdio",
      "command": "npx",
      "args": [
        "-y",
        "@modelcontextprotocol/server-sqlite",
        "/home/user/project/test.db",
        "--readonly"
      ]
    },
    "custom-tools": {
      "name": "custom-tools",
      "transport": "stdio",
      "command": "node",
      "args": ["/home/user/project/tools/mcp-server.js"]
    }
  }
}
```

### Production Setup

```json
{
  "version": "1.0",
  "servers": {
    "docs": {
      "name": "docs",
      "transport": "stdio",
      "command": "npx",
      "args": [
        "-y",
        "@modelcontextprotocol/server-filesystem",
        "/var/www/docs"
      ]
    },
    "prod-db": {
      "name": "prod-db",
      "transport": "stdio",
      "command": "npx",
      "args": [
        "-y",
        "@modelcontextprotocol/server-postgres",
        "--connection-string",
        "postgresql://user:pass@localhost:5432/prod"
      ]
    }
  }
}
```

**Security Note**: Never commit credentials to version control. Use environment variables:

```json
{
  "servers": {
    "prod-db": {
      "name": "prod-db",
      "transport": "stdio",
      "command": "sh",
      "args": [
        "-c",
        "npx -y @modelcontextprotocol/server-postgres --connection-string \"$DATABASE_URL\""
      ]
    }
  }
}
```

## Tips and Best Practices

### 1. Start Small

Begin with one server (filesystem is recommended) and learn how it works before adding more.

### 2. Use Descriptive Names

Choose server names that clearly indicate their purpose:
- ✅ `project-files`, `api-fetch`, `test-db`
- ❌ `server1`, `mcp`, `s`

### 3. Limit Scope

Configure servers with minimal necessary permissions:
- Filesystem: only directories that are needed
- Database: read-only when possible
- Fetch: consider using allowlist/blocklist

### 4. Test Servers Manually

Before adding to Hex, test server works standalone:

```bash
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}' | npx -y @modelcontextprotocol/server-filesystem /path
```

### 5. Monitor Resource Usage

MCP servers are separate processes. Monitor their resource usage:

```bash
# While Hex is running
ps aux | grep npx
```

### 6. Version Pin in Production

For production setups, pin specific server versions:

```bash
hex mcp add filesystem npx -y @modelcontextprotocol/server-filesystem@1.0.2 /data
```

### 7. Keep .mcp.json in Version Control

Commit `.mcp.json` to share server configurations across your team:

```bash
git add .mcp.json
git commit -m "Add MCP server configurations"
```

## Troubleshooting

### Server Not Starting

```bash
# Test server manually
npx -y @modelcontextprotocol/server-filesystem /path/to/dir

# Check npm global bin directory is in PATH
npm bin -g

# Reinstall server
npm uninstall -g @modelcontextprotocol/server-filesystem
npm install -g @modelcontextprotocol/server-filesystem
```

### Tools Not Appearing

```bash
# List configured servers
hex mcp list

# Verify .mcp.json exists and is valid
cat .mcp.json | jq

# Check Hex output for errors
hex 2>&1 | grep -i mcp
```

### Permission Errors

```bash
# For filesystem server, verify directory exists and is accessible
ls -la /path/to/allowed/directory

# For database server, verify file exists and is readable
ls -la /path/to/database.db
```

## Next Steps

- **Read** [TOOLS.md](../../docs/TOOLS.md) for complete MCP tools documentation
- **Explore** [MCP_INTEGRATION.md](../../docs/MCP_INTEGRATION.md) for architecture details
- **Build** your own MCP server using the examples above
- **Share** your servers with the community

## See Also

- [Official MCP Specification](https://spec.modelcontextprotocol.io/)
- [MCP TypeScript SDK](https://github.com/modelcontextprotocol/typescript-sdk)
- [MCP Python SDK](https://github.com/modelcontextprotocol/python-sdk)
- [MCP Server Examples](https://github.com/modelcontextprotocol/servers)
