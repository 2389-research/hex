// ABOUTME: Example usage of the Bash tool
// ABOUTME: Demonstrates executing commands with different parameters

package tools

import (
	"context"
	"fmt"
	"log"
)

func ExampleBashTool() {
	// Create a new Bash tool
	bashTool := NewBashTool()

	// Execute a simple command
	result, err := bashTool.Execute(context.Background(), map[string]interface{}{
		"command": "echo 'Hello from Bash tool!'",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Success:", result.Success)
	fmt.Println("Exit Code:", result.Metadata["exit_code"])
	// Output:
	// Success: true
	// Exit Code: 0
}

func ExampleBashTool_withTimeout() {
	bashTool := NewBashTool()

	// Execute command with custom timeout (2 seconds)
	result, err := bashTool.Execute(context.Background(), map[string]interface{}{
		"command": "sleep 1; echo 'Done sleeping'",
		"timeout": float64(2),
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Success:", result.Success)
	// Output:
	// Success: true
}

func ExampleBashTool_withWorkingDirectory() {
	bashTool := NewBashTool()

	// Execute command in a specific directory
	result, err := bashTool.Execute(context.Background(), map[string]interface{}{
		"command":     "pwd",
		"working_dir": "/tmp",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Working directory set:", result.Metadata["working_dir"])
	// Output:
	// Working directory set: /tmp
}

func ExampleBashTool_requiresApproval() {
	bashTool := NewBashTool()

	// ALL bash commands require approval for safety
	params := map[string]interface{}{
		"command": "ls /",
	}

	requiresApproval := bashTool.RequiresApproval(params)
	fmt.Println("Requires approval:", requiresApproval)
	// Output:
	// Requires approval: true
}

func ExampleBashTool_withExecutor() {
	// Create a registry and register the bash tool
	registry := NewRegistry()
	bashTool := NewBashTool()
	registry.Register(bashTool)

	// Create an executor with approval function
	approvalFunc := func(toolName string, params map[string]interface{}) bool {
		// In real use, prompt the user for approval
		command := params["command"].(string)
		fmt.Printf("Approve command '%s'? ", command)
		return true // Auto-approve for demo
	}

	executor := NewExecutor(registry, approvalFunc)

	// Execute the tool through the executor
	result, err := executor.Execute(context.Background(), "bash", map[string]interface{}{
		"command": "echo 'Executed via executor'",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Success:", result.Success)
	// Output:
	// Approve command 'echo 'Executed via executor''? Success: true
}
