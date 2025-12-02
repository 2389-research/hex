// ABOUTME: Integration example showing how real tools would use this architecture
// ABOUTME: Demonstrates the complete lifecycle from registration to API integration

package tools_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/harper/pagent/internal/tools"
)

// Example_completeWorkflow demonstrates a complete tool system workflow
func Example_completeWorkflow() {
	// Simulate a simple file system tool
	type FileSystemTool struct {
		name string
	}

	readTool := &FileSystemTool{name: "read_file"}
	writeTool := &FileSystemTool{name: "write_file"}

	// Implement Tool interface for FileSystemTool
	getName := func(t *FileSystemTool) string { return t.name }
	getDesc := func(t *FileSystemTool) string { return fmt.Sprintf("Tool: %s", t.name) }
	requiresApproval := func(t *FileSystemTool, _ map[string]interface{}) bool {
		// Write operations require approval
		return t.name == "write_file"
	}
	execute := func(t *FileSystemTool, _ context.Context, params map[string]interface{}) (*tools.Result, error) {
		path := params["path"].(string)
		if t.name == "read_file" {
			return &tools.Result{
				ToolName: t.name,
				Success:  true,
				Output:   fmt.Sprintf("Contents of %s", path),
				Metadata: map[string]interface{}{"path": path},
			}, nil
		}
		// write_file
		content := params["content"].(string)
		return &tools.Result{
			ToolName: t.name,
			Success:  true,
			Output:   fmt.Sprintf("Wrote %d bytes to %s", len(content), path),
			Metadata: map[string]interface{}{"path": path, "bytes": len(content)},
		}, nil
	}

	// Wrap in MockTool for demonstration
	readMock := &tools.MockTool{
		NameValue:        getName(readTool),
		DescriptionValue: getDesc(readTool),
		ExecuteFunc: func(ctx context.Context, params map[string]interface{}) (*tools.Result, error) {
			return execute(readTool, ctx, params)
		},
	}

	writeMock := &tools.MockTool{
		NameValue:             getName(writeTool),
		DescriptionValue:      getDesc(writeTool),
		RequiresApprovalValue: requiresApproval(writeTool, nil),
		ExecuteFunc: func(ctx context.Context, params map[string]interface{}) (*tools.Result, error) {
			return execute(writeTool, ctx, params)
		},
	}

	// 1. Setup: Create registry and register tools
	registry := tools.NewRegistry()
	_ = registry.Register(readMock)
	_ = registry.Register(writeMock)

	// 2. Create executor with approval logic
	approvalFunc := func(toolName string, params map[string]interface{}) bool {
		// Approve writes to /tmp, deny others
		if toolName == "write_file" {
			path := params["path"].(string)
			return strings.HasPrefix(path, "/tmp/")
		}
		return true
	}

	executor := tools.NewExecutor(registry, approvalFunc)

	// 3. Simulate API request (read_file)
	apiRequest1 := `{
		"type": "tool_use",
		"id": "toolu_read_123",
		"name": "read_file",
		"input": {
			"path": "/tmp/config.json"
		}
	}`

	var toolUse1 tools.ToolUse
	_ = json.Unmarshal([]byte(apiRequest1), &toolUse1)

	// 4. Execute tool
	ctx := context.Background()
	result1, _ := executor.Execute(ctx, toolUse1.Name, toolUse1.Input)

	// 5. Convert to API response
	toolResult1 := tools.ResultToToolResult(result1, toolUse1.ID)
	fmt.Printf("Read operation: success=%v\n", !toolResult1.IsError)

	// 6. Simulate API request (write_file with approval)
	apiRequest2 := `{
		"type": "tool_use",
		"id": "toolu_write_456",
		"name": "write_file",
		"input": {
			"path": "/tmp/output.txt",
			"content": "Hello, World!"
		}
	}`

	var toolUse2 tools.ToolUse
	_ = json.Unmarshal([]byte(apiRequest2), &toolUse2)

	result2, _ := executor.Execute(ctx, toolUse2.Name, toolUse2.Input)
	toolResult2 := tools.ResultToToolResult(result2, toolUse2.ID)
	fmt.Printf("Write operation (approved): success=%v\n", !toolResult2.IsError)

	// 7. Simulate denied write operation
	apiRequest3 := `{
		"type": "tool_use",
		"id": "toolu_write_789",
		"name": "write_file",
		"input": {
			"path": "/etc/passwd",
			"content": "malicious"
		}
	}`

	var toolUse3 tools.ToolUse
	_ = json.Unmarshal([]byte(apiRequest3), &toolUse3)

	result3, _ := executor.Execute(ctx, toolUse3.Name, toolUse3.Input)
	toolResult3 := tools.ResultToToolResult(result3, toolUse3.ID)
	fmt.Printf("Write operation (denied): success=%v\n", !toolResult3.IsError)

	// 8. List all available tools
	availableTools := registry.List()
	fmt.Printf("Available tools: %v\n", availableTools)

	// Output:
	// Read operation: success=true
	// Write operation (approved): success=true
	// Write operation (denied): success=false
	// Available tools: [read_file write_file]
}
