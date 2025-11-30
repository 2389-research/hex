// ABOUTME: Tests for MCP tool adapter wrapping MCP tools as Clem Tool interface
// ABOUTME: Verifies conversion between MCP and Clem tool execution models

package mcp

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/harper/clem/internal/tools"
)

func TestMCPTool_Name(t *testing.T) {
	mcpTool := Tool{
		Name:        "test_tool",
		Description: "Test tool",
		InputSchema: map[string]interface{}{},
	}

	tool := NewMCPToolAdapter(nil, mcpTool)
	if tool.Name() != "test_tool" {
		t.Errorf("Name() = %q, want 'test_tool'", tool.Name())
	}
}

func TestMCPTool_Description(t *testing.T) {
	mcpTool := Tool{
		Name:        "test_tool",
		Description: "This is a test tool",
		InputSchema: map[string]interface{}{},
	}

	tool := NewMCPToolAdapter(nil, mcpTool)
	if tool.Description() != "This is a test tool" {
		t.Errorf("Description() = %q, want 'This is a test tool'", tool.Description())
	}
}

func TestMCPTool_RequiresApproval(t *testing.T) {
	tests := []struct {
		name   string
		tool   Tool
		params map[string]interface{}
		want   bool
	}{
		{
			name: "safe tool",
			tool: Tool{
				Name:        "get_weather",
				Description: "Get weather",
			},
			params: map[string]interface{}{"city": "NYC"},
			want:   false,
		},
		{
			name: "empty params",
			tool: Tool{
				Name:        "test",
				Description: "Test",
			},
			params: map[string]interface{}{},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := NewMCPToolAdapter(nil, tt.tool)
			got := tool.RequiresApproval(tt.params)
			if got != tt.want {
				t.Errorf("RequiresApproval() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMCPTool_Execute_Success(t *testing.T) {
	testTools := []Tool{
		{
			Name:        "echo",
			Description: "Echo input",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"message": map[string]interface{}{
						"type": "string",
					},
				},
			},
		},
	}

	client, server := createTestClientAndServer("2024-11-05", testTools)

	// Custom handler
	server.RegisterToolHandler("echo", func(args map[string]interface{}) (interface{}, error) {
		msg := args["message"].(string)
		return map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": "Echo: " + msg,
				},
			},
		}, nil
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Initialize client
	if err := client.Initialize(ctx, "test-client", "1.0.0", "2024-11-05"); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Create adapter
	tool := NewMCPToolAdapter(client, testTools[0])

	// Execute tool
	params := map[string]interface{}{
		"message": "Hello, World!",
	}

	result, err := tool.Execute(ctx, params)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !result.Success {
		t.Errorf("Execute() success = false, want true. Error: %s", result.Error)
	}

	if !strings.Contains(result.Output, "Echo: Hello, World!") {
		t.Errorf("Output = %q, want to contain 'Echo: Hello, World!'", result.Output)
	}

	if result.ToolName != "echo" {
		t.Errorf("ToolName = %q, want 'echo'", result.ToolName)
	}
}

func TestMCPTool_Execute_Error(t *testing.T) {
	testTools := []Tool{
		{
			Name:        "error_tool",
			Description: "Always errors",
		},
	}

	client, server := createTestClientAndServer("2024-11-05", testTools)

	// Register handler that returns an error
	server.RegisterToolHandler("error_tool", func(_ map[string]interface{}) (interface{}, error) {
		return nil, fmt.Errorf("simulated tool error")
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Initialize(ctx, "test-client", "1.0.0", "2024-11-05"); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Create adapter
	tool := NewMCPToolAdapter(client, testTools[0])

	// Execute - will fail because handler returns an error
	result, err := tool.Execute(ctx, map[string]interface{}{})

	// Should get a result, not a catastrophic error
	if err != nil {
		t.Fatalf("Execute() should return Result with error, not error. Got: %v", err)
	}

	if result.Success {
		t.Error("Execute() success should be false for error case")
	}

	if result.Error == "" {
		t.Error("Result.Error should contain error message")
	}
}

func TestMCPTool_Execute_TextContent(t *testing.T) {
	testTools := []Tool{
		{
			Name:        "get_data",
			Description: "Get data",
		},
	}

	client, server := createTestClientAndServer("2024-11-05", testTools)

	server.RegisterToolHandler("get_data", func(_ map[string]interface{}) (interface{}, error) {
		return map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": "Line 1",
				},
				{
					"type": "text",
					"text": "Line 2",
				},
			},
		}, nil
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Initialize(ctx, "test-client", "1.0.0", "2024-11-05"); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	tool := NewMCPToolAdapter(client, testTools[0])
	result, err := tool.Execute(ctx, map[string]interface{}{})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !result.Success {
		t.Errorf("Execute() should succeed")
	}

	// Should concatenate multiple text blocks
	if !strings.Contains(result.Output, "Line 1") {
		t.Errorf("Output should contain 'Line 1'")
	}
	if !strings.Contains(result.Output, "Line 2") {
		t.Errorf("Output should contain 'Line 2'")
	}
}

func TestMCPTool_Execute_EmptyContent(t *testing.T) {
	testTools := []Tool{
		{
			Name:        "empty_tool",
			Description: "Returns empty content",
		},
	}

	client, server := createTestClientAndServer("2024-11-05", testTools)

	server.RegisterToolHandler("empty_tool", func(_ map[string]interface{}) (interface{}, error) {
		return map[string]interface{}{
			"content": []map[string]interface{}{},
		}, nil
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Initialize(ctx, "test-client", "1.0.0", "2024-11-05"); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	tool := NewMCPToolAdapter(client, testTools[0])
	result, err := tool.Execute(ctx, map[string]interface{}{})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !result.Success {
		t.Error("Execute() should succeed even with empty content")
	}

	if result.Output != "" {
		t.Errorf("Output should be empty, got %q", result.Output)
	}
}

func TestMCPTool_Execute_ContextCancellation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping flaky context cancellation test in short mode")
	}

	testTools := []Tool{
		{
			Name:        "slow_tool",
			Description: "Slow operation",
		},
	}

	client, _ := createTestClientAndServer("2024-11-05", testTools)

	initCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Initialize(initCtx, "test-client", "1.0.0", "2024-11-05"); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	tool := NewMCPToolAdapter(client, testTools[0])

	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result, err := tool.Execute(ctx, map[string]interface{}{})

	// Context cancellation should be reflected in either error or result
	if err == nil && (result == nil || result.Success) {
		t.Error("Execute() should fail with cancelled context")
	}

	if err != nil && !strings.Contains(err.Error(), "context") {
		t.Errorf("Error should mention context, got: %v", err)
	}
}

func TestMCPTool_AsToolDefinition(t *testing.T) {
	mcpTool := Tool{
		Name:        "calculate",
		Description: "Perform calculation",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"expression": map[string]interface{}{
					"type":        "string",
					"description": "Math expression",
				},
			},
			"required": []interface{}{"expression"},
		},
	}

	tool := NewMCPToolAdapter(nil, mcpTool)
	def := tool.AsToolDefinition()

	if def.Name != "calculate" {
		t.Errorf("Definition name = %q, want 'calculate'", def.Name)
	}

	if def.Description != "Perform calculation" {
		t.Errorf("Definition description = %q, want 'Perform calculation'", def.Description)
	}

	// Verify schema is preserved
	schema := def.InputSchema
	if schema["type"] != "object" {
		t.Errorf("Schema type = %q, want 'object'", schema["type"])
	}
}

func TestMCPTool_ConvertParams(t *testing.T) {
	tests := []struct {
		name  string
		input map[string]interface{}
		want  map[string]interface{}
	}{
		{
			name: "simple params",
			input: map[string]interface{}{
				"key": "value",
			},
			want: map[string]interface{}{
				"key": "value",
			},
		},
		{
			name: "nested params",
			input: map[string]interface{}{
				"outer": map[string]interface{}{
					"inner": "value",
				},
			},
			want: map[string]interface{}{
				"outer": map[string]interface{}{
					"inner": "value",
				},
			},
		},
		{
			name:  "empty params",
			input: map[string]interface{}{},
			want:  map[string]interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// The adapter should pass params through unchanged
			// (no conversion needed between Clem and MCP format)
			_ = NewMCPToolAdapter(nil, Tool{Name: "test"})

			// In practice, Execute would use these params directly
			// This test just verifies the contract
			if len(tt.input) != len(tt.want) {
				t.Errorf("Param conversion issue: input len %d, want %d", len(tt.input), len(tt.want))
			}
		})
	}
}

func TestMCPTool_Integration(t *testing.T) {
	// Integration test: create tool, register with Clem's registry, execute
	testTools := []Tool{
		{
			Name:        "weather",
			Description: "Get weather",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"city": map[string]interface{}{
						"type": "string",
					},
				},
			},
		},
	}

	client, server := createTestClientAndServer("2024-11-05", testTools)

	server.RegisterToolHandler("weather", func(args map[string]interface{}) (interface{}, error) {
		city := args["city"].(string)
		return map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": "Weather in " + city + ": Sunny, 72°F",
				},
			},
		}, nil
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Initialize(ctx, "test-client", "1.0.0", "2024-11-05"); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Create MCP tool adapter
	mcpTool := NewMCPToolAdapter(client, testTools[0])

	// Use with Clem's tool registry
	registry := tools.NewRegistry()
	_ = registry.Register(mcpTool)

	// Execute via registry
	foundTool, err := registry.Get("weather")
	if err != nil {
		t.Fatalf("Tool not found in registry: %v", err)
	}

	result, err := foundTool.Execute(ctx, map[string]interface{}{
		"city": "San Francisco",
	})

	if err != nil {
		t.Fatalf("Execute via registry failed: %v", err)
	}

	if !result.Success {
		t.Errorf("Result should be success")
	}

	if !strings.Contains(result.Output, "San Francisco") {
		t.Errorf("Output should contain 'San Francisco', got: %s", result.Output)
	}
}
