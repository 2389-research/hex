// ABOUTME: Example usage of the tool system
// ABOUTME: Demonstrates how to register tools, execute them, and handle approvals

package tools_test

import (
	"context"
	"fmt"

	"github.com/harper/pagent/internal/tools"
)

// Example demonstrates the complete tool system workflow
func Example() {
	// 1. Create a registry
	registry := tools.NewRegistry()

	// 2. Create and register a safe tool (no approval needed)
	safeTool := &tools.MockTool{
		NameValue:             "echo",
		DescriptionValue:      "Echoes back the input",
		RequiresApprovalValue: false,
		ExecuteFunc: func(_ context.Context, params map[string]interface{}) (*tools.Result, error) {
			message := params["message"].(string)
			return &tools.Result{
				ToolName: "echo",
				Success:  true,
				Output:   message,
			}, nil
		},
	}
	_ = registry.Register(safeTool)

	// 3. Create and register a dangerous tool (requires approval)
	dangerousTool := &tools.MockTool{
		NameValue:             "delete_file",
		DescriptionValue:      "Deletes a file",
		RequiresApprovalValue: true,
		ExecuteFunc: func(_ context.Context, params map[string]interface{}) (*tools.Result, error) {
			path := params["path"].(string)
			return &tools.Result{
				ToolName: "delete_file",
				Success:  true,
				Output:   fmt.Sprintf("Deleted %s", path),
			}, nil
		},
	}
	_ = registry.Register(dangerousTool)

	// 4. Create an executor with approval function
	approvalFunc := func(toolName string, params map[string]interface{}) bool {
		// In a real application, this would prompt the user
		// For this example, we'll approve delete operations on /tmp
		if toolName == "delete_file" {
			path := params["path"].(string)
			return path == "/tmp/test.txt"
		}
		return true
	}
	executor := tools.NewExecutor(registry, approvalFunc)

	// 5. Execute the safe tool (no approval needed)
	ctx := context.Background()
	result1, _ := executor.Execute(ctx, "echo", map[string]interface{}{
		"message": "Hello, World!",
	})
	fmt.Printf("Echo result: %s\n", result1.Output)

	// 6. Execute dangerous tool with approved path
	result2, _ := executor.Execute(ctx, "delete_file", map[string]interface{}{
		"path": "/tmp/test.txt",
	})
	fmt.Printf("Delete result (approved): %s\n", result2.Output)

	// 7. Execute dangerous tool with denied path
	result3, _ := executor.Execute(ctx, "delete_file", map[string]interface{}{
		"path": "/etc/passwd",
	})
	fmt.Printf("Delete result (denied): success=%v, error=%s\n", result3.Success, result3.Error)

	// 8. List all registered tools
	tools := registry.List()
	fmt.Printf("Registered tools: %v\n", tools)

	// Output:
	// Echo result: Hello, World!
	// Delete result (approved): Deleted /tmp/test.txt
	// Delete result (denied): success=false, error=user denied permission
	// Registered tools: [delete_file echo]
}

// ExampleResultToToolResult demonstrates converting internal results to API format
func ExampleResultToToolResult() {
	// Success result
	successResult := &tools.Result{
		ToolName: "read_file",
		Success:  true,
		Output:   "file contents",
	}
	apiResult1 := tools.ResultToToolResult(successResult, "toolu_123")
	fmt.Printf("API result 1: type=%s, error=%v\n", apiResult1.Type, apiResult1.IsError)

	// Error result
	errorResult := &tools.Result{
		ToolName: "write_file",
		Success:  false,
		Error:    "permission denied",
	}
	apiResult2 := tools.ResultToToolResult(errorResult, "toolu_456")
	fmt.Printf("API result 2: type=%s, error=%v, content=%s\n",
		apiResult2.Type, apiResult2.IsError, apiResult2.Content)

	// Output:
	// API result 1: type=tool_result, error=false
	// API result 2: type=tool_result, error=true, content=Error: permission denied
}
