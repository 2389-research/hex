// ABOUTME: Examples demonstrating tool usage
// ABOUTME: Shows how to use Read tool and other tools in practice

package tools_test

import (
	"context"
	"fmt"
	"os"

	"github.com/harper/jeff/internal/tools"
)

// ExampleReadTool demonstrates basic file reading
func ExampleReadTool() {
	// Create a test file
	tmpFile, _ := os.CreateTemp("", "example-*.txt")
	defer func() { _ = os.Remove(tmpFile.Name()) }()
	_, _ = tmpFile.WriteString("Hello, World!")
	_ = tmpFile.Close()

	// Create and use the Read tool
	tool := tools.NewReadTool()
	result, _ := tool.Execute(context.Background(), map[string]interface{}{
		"path": tmpFile.Name(),
	})

	fmt.Println("Success:", result.Success)
	fmt.Println("Output:", result.Output)
	// Output:
	// Success: true
	// Output: Hello, World!
}

// ExampleReadTool_withOffset demonstrates reading with an offset
func ExampleReadTool_withOffset() {
	tmpFile, _ := os.CreateTemp("", "example-*.txt")
	defer func() { _ = os.Remove(tmpFile.Name()) }()
	_, _ = tmpFile.WriteString("0123456789")
	_ = tmpFile.Close()

	tool := tools.NewReadTool()
	result, _ := tool.Execute(context.Background(), map[string]interface{}{
		"path":   tmpFile.Name(),
		"offset": float64(5),
	})

	fmt.Println(result.Output)
	// Output: 56789
}

// ExampleReadTool_withLimit demonstrates reading with a limit
func ExampleReadTool_withLimit() {
	tmpFile, _ := os.CreateTemp("", "example-*.txt")
	defer func() { _ = os.Remove(tmpFile.Name()) }()
	_, _ = tmpFile.WriteString("0123456789")
	_ = tmpFile.Close()

	tool := tools.NewReadTool()
	result, _ := tool.Execute(context.Background(), map[string]interface{}{
		"path":  tmpFile.Name(),
		"limit": float64(5),
	})

	fmt.Println(result.Output)
	// Output: 01234
}

// ExampleReadTool_requiresApproval demonstrates approval checking
func ExampleReadTool_requiresApproval() {
	tool := tools.NewReadTool()

	// Sensitive path requires approval
	sensitive := tool.RequiresApproval(map[string]interface{}{
		"path": "/etc/passwd",
	})
	fmt.Println("Sensitive path:", sensitive)

	// Regular path does not require approval
	regular := tool.RequiresApproval(map[string]interface{}{
		"path": "/tmp/test.txt",
	})
	fmt.Println("Regular path:", regular)

	// Output:
	// Sensitive path: true
	// Regular path: false
}

// ExampleReadTool_withExecutor demonstrates using the Read tool with Executor
func ExampleReadTool_withExecutor() {
	tmpFile, _ := os.CreateTemp("", "example-*.txt")
	defer func() { _ = os.Remove(tmpFile.Name()) }()
	_, _ = tmpFile.WriteString("Hello from Executor!")
	_ = tmpFile.Close()

	// Create registry and register the Read tool
	registry := tools.NewRegistry()
	_ = registry.Register(tools.NewReadTool())

	// Create executor with auto-approval
	executor := tools.NewExecutor(registry, func(_ string, _ map[string]interface{}) bool {
		return true // Auto-approve for this example
	})

	// Execute the tool through the executor
	result, _ := executor.Execute(context.Background(), "read_file", map[string]interface{}{
		"path": tmpFile.Name(),
	})

	fmt.Println("Success:", result.Success)
	fmt.Println("Output:", result.Output)
	// Output:
	// Success: true
	// Output: Hello from Executor!
}
