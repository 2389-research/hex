// ABOUTME: Integration tests for hex-mux adapter.
// ABOUTME: Verifies end-to-end agent creation and tool execution.
package adapter

import (
	"context"
	"testing"

	"github.com/2389-research/hex/internal/tools"
	muxtool "github.com/2389-research/mux/tool"
)

func TestRootAgentToolExecution(t *testing.T) {
	// Create a simple mock tool
	mockTool := &mockHexTool{
		name:        "echo",
		description: "Echo back input",
		result: &tools.Result{
			ToolName: "echo",
			Success:  true,
			Output:   "echoed: hello",
		},
	}

	// Create registry with adapted tool
	registry := muxtool.NewRegistry()
	registry.Register(AdaptTool(mockTool))

	// Verify tool is registered
	tool, ok := registry.Get("echo")
	if !ok {
		t.Fatal("expected echo tool to be registered")
	}

	// Execute tool through mux
	result, err := tool.Execute(context.Background(), map[string]any{"input": "hello"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Output != "echoed: hello" {
		t.Errorf("expected output 'echoed: hello', got %s", result.Output)
	}
}

func TestSubagentToolFiltering(t *testing.T) {
	// Create multiple mock tools
	readTool := &mockHexTool{name: "Read", description: "Read files"}
	writeTool := &mockHexTool{name: "Write", description: "Write files"}
	bashTool := &mockHexTool{name: "Bash", description: "Execute bash"}

	// Create registry with all tools
	registry := muxtool.NewRegistry()
	registry.Register(AdaptTool(readTool))
	registry.Register(AdaptTool(writeTool))
	registry.Register(AdaptTool(bashTool))

	// Create filtered registry (simulating Explore subagent)
	filtered := muxtool.NewFilteredRegistry(registry, []string{"Read", "Bash"}, nil)

	// Verify filtering
	if _, ok := filtered.Get("Read"); !ok {
		t.Error("expected Read to be allowed")
	}
	if _, ok := filtered.Get("Bash"); !ok {
		t.Error("expected Bash to be allowed")
	}
	if _, ok := filtered.Get("Write"); ok {
		t.Error("expected Write to be denied")
	}

	// Verify count
	if filtered.Count() != 2 {
		t.Errorf("expected 2 tools, got %d", filtered.Count())
	}
}
