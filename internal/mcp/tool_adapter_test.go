// ABOUTME: Tests for MCP tool adapter wrapping MCP tools as Hex Tool interface
// ABOUTME: Verifies conversion between MCP and Hex tool execution models

package mcp

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/2389-research/hex/internal/tools"
	muxmcp "github.com/2389-research/mux/mcp"
)

// mockMCPClient implements muxmcp.Client for testing
type mockMCPClient struct {
	callToolFunc func(ctx context.Context, name string, args map[string]any) (*muxmcp.ToolCallResult, error)
}

func (m *mockMCPClient) Start(ctx context.Context) error {
	return nil
}

func (m *mockMCPClient) ListTools(ctx context.Context) ([]muxmcp.ToolInfo, error) {
	return nil, nil
}

func (m *mockMCPClient) CallTool(ctx context.Context, name string, args map[string]any) (*muxmcp.ToolCallResult, error) {
	if m.callToolFunc != nil {
		return m.callToolFunc(ctx, name, args)
	}
	return &muxmcp.ToolCallResult{
		Content: []muxmcp.ContentBlock{
			{Type: "text", Text: "mock response"},
		},
	}, nil
}

func (m *mockMCPClient) Notifications() <-chan muxmcp.Notification {
	return nil
}

func (m *mockMCPClient) Close() error {
	return nil
}

func TestMCPTool_Name(t *testing.T) {
	mcpTool := muxmcp.ToolInfo{
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
	mcpTool := muxmcp.ToolInfo{
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
		tool   muxmcp.ToolInfo
		params map[string]interface{}
		want   bool
	}{
		{
			name: "safe tool",
			tool: muxmcp.ToolInfo{
				Name:        "get_weather",
				Description: "Get weather",
			},
			params: map[string]interface{}{"city": "NYC"},
			want:   false,
		},
		{
			name: "empty params",
			tool: muxmcp.ToolInfo{
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
	mockClient := &mockMCPClient{
		callToolFunc: func(ctx context.Context, name string, args map[string]any) (*muxmcp.ToolCallResult, error) {
			msg := args["message"].(string)
			return &muxmcp.ToolCallResult{
				Content: []muxmcp.ContentBlock{
					{Type: "text", Text: "Echo: " + msg},
				},
			}, nil
		},
	}

	mcpTool := muxmcp.ToolInfo{
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
	}

	tool := NewMCPToolAdapter(mockClient, mcpTool)

	params := map[string]interface{}{
		"message": "Hello, World!",
	}

	result, err := tool.Execute(context.Background(), params)
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
	mockClient := &mockMCPClient{
		callToolFunc: func(ctx context.Context, name string, args map[string]any) (*muxmcp.ToolCallResult, error) {
			return nil, fmt.Errorf("simulated tool error")
		},
	}

	mcpTool := muxmcp.ToolInfo{
		Name:        "error_tool",
		Description: "Always errors",
	}

	tool := NewMCPToolAdapter(mockClient, mcpTool)

	result, err := tool.Execute(context.Background(), map[string]interface{}{})

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

func TestMCPTool_Execute_IsError(t *testing.T) {
	mockClient := &mockMCPClient{
		callToolFunc: func(ctx context.Context, name string, args map[string]any) (*muxmcp.ToolCallResult, error) {
			return &muxmcp.ToolCallResult{
				Content: []muxmcp.ContentBlock{
					{Type: "text", Text: "Tool error message"},
				},
				IsError: true,
			}, nil
		},
	}

	mcpTool := muxmcp.ToolInfo{
		Name:        "failing_tool",
		Description: "Returns IsError",
	}

	tool := NewMCPToolAdapter(mockClient, mcpTool)

	result, err := tool.Execute(context.Background(), map[string]interface{}{})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result.Success {
		t.Error("Execute() success should be false when IsError is true")
	}

	if result.Error == "" {
		t.Error("Result.Error should contain error message")
	}
}

func TestMCPTool_Execute_TextContent(t *testing.T) {
	mockClient := &mockMCPClient{
		callToolFunc: func(ctx context.Context, name string, args map[string]any) (*muxmcp.ToolCallResult, error) {
			return &muxmcp.ToolCallResult{
				Content: []muxmcp.ContentBlock{
					{Type: "text", Text: "Line 1"},
					{Type: "text", Text: "Line 2"},
				},
			}, nil
		},
	}

	mcpTool := muxmcp.ToolInfo{
		Name:        "get_data",
		Description: "Get data",
	}

	tool := NewMCPToolAdapter(mockClient, mcpTool)
	result, err := tool.Execute(context.Background(), map[string]interface{}{})
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
	mockClient := &mockMCPClient{
		callToolFunc: func(ctx context.Context, name string, args map[string]any) (*muxmcp.ToolCallResult, error) {
			return &muxmcp.ToolCallResult{
				Content: []muxmcp.ContentBlock{},
			}, nil
		},
	}

	mcpTool := muxmcp.ToolInfo{
		Name:        "empty_tool",
		Description: "Returns empty content",
	}

	tool := NewMCPToolAdapter(mockClient, mcpTool)
	result, err := tool.Execute(context.Background(), map[string]interface{}{})
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
	mockClient := &mockMCPClient{
		callToolFunc: func(ctx context.Context, name string, args map[string]any) (*muxmcp.ToolCallResult, error) {
			return nil, ctx.Err()
		},
	}

	mcpTool := muxmcp.ToolInfo{
		Name:        "slow_tool",
		Description: "Slow operation",
	}

	tool := NewMCPToolAdapter(mockClient, mcpTool)

	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result, err := tool.Execute(ctx, map[string]interface{}{})

	// Context cancellation should be reflected
	if err == nil && (result == nil || result.Success) {
		t.Error("Execute() should fail with cancelled context")
	}
}

func TestMCPTool_AsToolDefinition(t *testing.T) {
	mcpTool := muxmcp.ToolInfo{
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

func TestMCPTool_Integration(t *testing.T) {
	// Integration test: create tool, register with Hex's registry, execute
	mockClient := &mockMCPClient{
		callToolFunc: func(ctx context.Context, name string, args map[string]any) (*muxmcp.ToolCallResult, error) {
			city := args["city"].(string)
			return &muxmcp.ToolCallResult{
				Content: []muxmcp.ContentBlock{
					{Type: "text", Text: "Weather in " + city + ": Sunny, 72°F"},
				},
			}, nil
		},
	}

	mcpTool := muxmcp.ToolInfo{
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
	}

	// Create MCP tool adapter
	adapter := NewMCPToolAdapter(mockClient, mcpTool)

	// Use with Hex's tool registry
	registry := tools.NewRegistry()
	_ = registry.Register(adapter)

	// Execute via registry
	foundTool, err := registry.Get("weather")
	if err != nil {
		t.Fatalf("Tool not found in registry: %v", err)
	}

	result, err := foundTool.Execute(context.Background(), map[string]interface{}{
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

func TestMCPTool_ImageContent(t *testing.T) {
	mockClient := &mockMCPClient{
		callToolFunc: func(ctx context.Context, name string, args map[string]any) (*muxmcp.ToolCallResult, error) {
			return &muxmcp.ToolCallResult{
				Content: []muxmcp.ContentBlock{
					{Type: "text", Text: "Image caption"},
					{Type: "image", MimeType: "image/png", Data: "base64data"},
				},
			}, nil
		},
	}

	mcpTool := muxmcp.ToolInfo{
		Name:        "image_tool",
		Description: "Returns image",
	}

	tool := NewMCPToolAdapter(mockClient, mcpTool)
	result, err := tool.Execute(context.Background(), map[string]interface{}{})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !result.Success {
		t.Error("Execute() should succeed")
	}

	if result.Metadata["has_image"] != true {
		t.Error("Metadata should indicate has_image=true")
	}

	if result.Metadata["image_mime_type"] != "image/png" {
		t.Errorf("Metadata should have image_mime_type=image/png, got %v", result.Metadata["image_mime_type"])
	}
}
