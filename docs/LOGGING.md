# Logging in Hex

Hex uses structured logging to help with debugging, monitoring, and understanding what's happening during execution.

## Enabling Logging

By default, logging is set to `info` level and outputs to stderr. You can customize logging behavior using command-line flags.

### Log Levels

Hex supports four log levels:

- **debug**: Verbose output including internal state, API calls, and detailed execution flow
- **info**: General informational messages about application state (default)
- **warn**: Warning messages about potential issues that don't prevent execution
- **error**: Error messages indicating failures or problems

### Command-Line Flags

#### `--log-level`

Set the minimum log level to display:

```bash
# Show only errors
hex --log-level error

# Show debug output for troubleshooting
hex --log-level debug "explain this code"

# Default (info level)
hex "write a function"
```

#### `--log-file`

Write logs to a file instead of (or in addition to) stderr:

```bash
# Log to file only
hex --log-file hex.log "help me debug this"

# Log to file AND stderr when using debug level
hex --log-level debug --log-file debug.log "investigate issue"
```

When `--log-file` is specified:
- Info/warn/error levels: logs only to the file
- Debug level: logs to both file and stderr for convenience

#### `--log-format`

Choose the log output format:

```bash
# Human-readable text format (default)
hex --log-format text

# JSON format for parsing/integration with log aggregators
hex --log-format json --log-file logs.jsonl
```

## Log Output Examples

### Text Format (Default)

```
time=2025-11-28T11:30:00.123-06:00 level=INFO msg="Hex starting" version=0.5.0
time=2025-11-28T11:30:00.150-06:00 level=DEBUG msg="Opening database" path=/Users/harper/.hex/hex.db
time=2025-11-28T11:30:00.200-06:00 level=INFO msg="Database opened successfully" path=/Users/harper/.hex/hex.db
time=2025-11-28T11:30:00.250-06:00 level=DEBUG msg="Loading MCP tools"
time=2025-11-28T11:30:00.300-06:00 level=WARN msg="Failed to load MCP tools" error="config file not found"
time=2025-11-28T11:30:05.500-06:00 level=INFO msg="Hex shutting down"
```

### JSON Format

```json
{"time":"2025-11-28T11:30:00.123-06:00","level":"INFO","msg":"Hex starting","version":"0.5.0"}
{"time":"2025-11-28T11:30:00.150-06:00","level":"DEBUG","msg":"Opening database","path":"/Users/harper/.hex/hex.db"}
{"time":"2025-11-28T11:30:00.200-06:00","level":"INFO","msg":"Database opened successfully","path":"/Users/harper/.hex/hex.db"}
{"time":"2025-11-28T11:30:00.250-06:00","level":"DEBUG","msg":"Loading MCP tools"}
{"time":"2025-11-28T11:30:00.300-06:00","level":"WARN","msg":"Failed to load MCP tools","error":"config file not found"}
{"time":"2025-11-28T11:30:05.500-06:00","level":"INFO","msg":"Hex shutting down"}
```

## Troubleshooting with Logs

### Common Issues

#### Database Connection Errors

Enable debug logging to see database operations:

```bash
hex --log-level debug --log-file debug.log --continue
```

Look for messages like:
- `Opening database` - shows the path being used
- `Database opened successfully` - confirms connection
- `Failed to open database` - shows specific error with suggestions

#### MCP Tool Loading Issues

```bash
hex --log-level debug
```

Look for:
- `Loading MCP tools` - indicates loading attempt
- `Failed to load MCP tools` - shows why tools didn't load
- Tool registration messages with specific tool names

#### API Request/Response Issues

Debug level shows:
- Request IDs for tracking API calls
- Conversation IDs for correlating messages
- Detailed error information from API

### Log File Location

When using `--log-file`, consider these locations:

```bash
# In current directory
hex --log-file ./hex.log

# In home directory
hex --log-file ~/hex.log

# In a dedicated logs directory
mkdir -p ~/.hex/logs
hex --log-file ~/.hex/logs/hex-$(date +%Y%m%d).log
```

### Parsing JSON Logs

JSON logs are ideal for processing with tools like `jq`:

```bash
# Extract only error messages
cat logs.jsonl | jq 'select(.level == "ERROR")'

# Get all messages for a specific conversation
cat logs.jsonl | jq 'select(.conversation_id == "conv-123")'

# Count messages by level
cat logs.jsonl | jq -r .level | sort | uniq -c
```

## Integration Examples

### Systemd Service

```ini
[Service]
ExecStart=/usr/local/bin/hex --log-level info --log-file /var/log/hex/hex.log --log-format json
StandardOutput=journal
StandardError=journal
```

### Shell Script Wrapper

```bash
#!/bin/bash
LOG_DIR="$HOME/.hex/logs"
mkdir -p "$LOG_DIR"
LOG_FILE="$LOG_DIR/hex-$(date +%Y%m%d-%H%M%S).log"

hex --log-level debug --log-file "$LOG_FILE" "$@"
```

## Best Practices

1. **Use appropriate log levels**:
   - `debug` for development and troubleshooting
   - `info` for normal operation
   - `warn` and `error` for production monitoring

2. **Log to files for long sessions**:
   ```bash
   hex --log-file ~/.hex/session.log
   ```

3. **Use JSON format for automated processing**:
   ```bash
   hex --log-format json --log-file logs.jsonl
   ```

4. **Rotate log files** to prevent them from growing too large:
   ```bash
   # Keep last 5 log files
   find ~/.hex/logs -name "hex-*.log" -mtime +5 -delete
   ```

## Performance Impact

Logging is designed to have minimal performance impact:

- Lazy evaluation of log messages
- Thread-safe operations
- Efficient buffering when writing to files
- No blocking on I/O operations

Debug level logging may have a small performance impact due to increased volume, but it's negligible for typical CLI usage.
