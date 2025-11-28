# Write Tool Usage Examples

## Basic Usage

### Creating a new file

```go
package main

import (
    "context"
    "fmt"
    "github.com/harper/clem/internal/tools"
)

func main() {
    // Create write tool
    writeTool := tools.NewWriteTool()

    // Execute write operation
    result, err := writeTool.Execute(context.Background(), map[string]interface{}{
        "path":    "/tmp/example.txt",
        "content": "Hello, World!",
        "mode":    "create",  // Optional, default is "create"
    })

    if err != nil {
        panic(err)
    }

    if !result.Success {
        fmt.Printf("Write failed: %s\n", result.Error)
        return
    }

    fmt.Printf("Success! %s\n", result.Output)
    fmt.Printf("Wrote %d bytes to %s\n",
        result.Metadata["bytes_written"],
        result.Metadata["path"])
}
```

### Overwriting an existing file

```go
result, err := writeTool.Execute(context.Background(), map[string]interface{}{
    "path":    "/tmp/existing.txt",
    "content": "New content that replaces the old",
    "mode":    "overwrite",
})
```

### Appending to a file

```go
result, err := writeTool.Execute(context.Background(), map[string]interface{}{
    "path":    "/tmp/log.txt",
    "content": "Another log entry\n",
    "mode":    "append",
})
```

## Using with Registry and Executor

```go
package main

import (
    "context"
    "fmt"
    "github.com/harper/clem/internal/tools"
)

func main() {
    // Create registry
    registry := tools.NewRegistry()

    // Register write tool
    writeTool := tools.NewWriteTool()
    registry.Register(writeTool)

    // Create executor with approval function
    executor := tools.NewExecutor(registry, func(toolName string, params map[string]interface{}) bool {
        // In a real application, this would prompt the user
        fmt.Printf("Tool '%s' wants to write to: %s\n", toolName, params["path"])
        fmt.Print("Approve? (y/n): ")

        var response string
        fmt.Scanln(&response)
        return response == "y"
    })

    // Execute write through executor (will request approval)
    result, err := executor.Execute(context.Background(), "write_file", map[string]interface{}{
        "path":    "/tmp/approved.txt",
        "content": "This write was approved by the user",
    })

    if err != nil {
        panic(err)
    }

    if !result.Success {
        fmt.Printf("Write failed: %s\n", result.Error)
        return
    }

    fmt.Println(result.Output)
}
```

## Write Modes

### Create Mode (default)
- Creates a new file
- Fails if file already exists
- Use when you want to ensure you don't accidentally overwrite

```go
map[string]interface{}{
    "path":    "/tmp/new.txt",
    "content": "content",
    "mode":    "create",  // or omit mode parameter
}
```

### Overwrite Mode
- Replaces existing file content
- Creates file if it doesn't exist
- Use when you want to ensure the file contains only your content

```go
map[string]interface{}{
    "path":    "/tmp/file.txt",
    "content": "new content",
    "mode":    "overwrite",
}
```

### Append Mode
- Adds content to end of existing file
- Creates file if it doesn't exist
- Use for logging or accumulating data

```go
map[string]interface{}{
    "path":    "/tmp/log.txt",
    "content": "log entry\n",
    "mode":    "append",
}
```

## Features

### Automatic Parent Directory Creation

The Write tool automatically creates parent directories if they don't exist:

```go
result, err := writeTool.Execute(context.Background(), map[string]interface{}{
    "path":    "/tmp/deeply/nested/path/file.txt",
    "content": "content",
})
// Creates /tmp/deeply, /tmp/deeply/nested, and /tmp/deeply/nested/path
```

### Path Safety

Paths are automatically cleaned to prevent directory traversal attacks:

```go
result, err := writeTool.Execute(context.Background(), map[string]interface{}{
    "path":    "/tmp/subdir/../../../etc/passwd",  // Dangerous!
    "content": "evil",
})
// Path is cleaned and normalized before writing
```

### Content Size Limits

Content is limited to 10MB by default:

```go
writeTool := tools.NewWriteTool()
writeTool.MaxContentSize = 5 * 1024 * 1024  // Set to 5MB

result, err := writeTool.Execute(context.Background(), map[string]interface{}{
    "path":    "/tmp/large.txt",
    "content": hugeString,  // Will fail if > 5MB
})
```

### Unicode Support

The Write tool correctly handles Unicode content:

```go
result, err := writeTool.Execute(context.Background(), map[string]interface{}{
    "path":    "/tmp/unicode.txt",
    "content": "Hello 世界 🌍",
})
```

## Error Handling

### Missing or Invalid Parameters

```go
result, err := writeTool.Execute(context.Background(), map[string]interface{}{
    // Missing "path" parameter
    "content": "hello",
})
// result.Success == false
// result.Error == "missing or invalid 'path' parameter"
```

### File Already Exists (Create Mode)

```go
result, err := writeTool.Execute(context.Background(), map[string]interface{}{
    "path":    "/tmp/existing.txt",  // File already exists
    "content": "content",
    "mode":    "create",
})
// result.Success == false
// result.Error contains "already exists" and suggests using "overwrite"
```

### Permission Denied

```go
result, err := writeTool.Execute(context.Background(), map[string]interface{}{
    "path":    "/root/protected.txt",  // No permission
    "content": "content",
})
// result.Success == false
// result.Error contains "failed to write file"
```

## Security Considerations

### Always Requires Approval

The Write tool ALWAYS requires approval because writing to disk is dangerous:

```go
writeTool.RequiresApproval(params)  // Always returns true
```

### Metadata Includes Absolute Path

The result metadata always includes the absolute path that was written to:

```go
result, _ := writeTool.Execute(ctx, map[string]interface{}{
    "path":    "relative/path.txt",  // Relative path provided
    "content": "content",
})

absPath := result.Metadata["path"]  // Returns absolute path
// e.g., "/home/user/project/relative/path.txt"
```

### Clean Paths

All paths are cleaned before use to prevent directory traversal:

```go
// These are all cleaned and normalized:
"./file.txt"              -> "/current/dir/file.txt"
"subdir/../file.txt"      -> "/current/dir/file.txt"
"/tmp/dir/./file.txt"     -> "/tmp/dir/file.txt"
```
