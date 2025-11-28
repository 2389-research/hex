// ABOUTME: Adapter that wraps MCP tools to implement Clem's Tool interface
// ABOUTME: Bridges MCP protocol tools with Clem's tool execution system

package mcp

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/harper/clem/internal/core"
	"github.com/harper/clem/internal/tools"
)

// MCPToolAdapter wraps an MCP tool to implement Clem's Tool interface
type MCPToolAdapter struct {
	client  *Client
	mcpTool Tool
}

// NewMCPToolAdapter creates a new adapter for an MCP tool
func NewMCPToolAdapter(client *Client, mcpTool Tool) *MCPToolAdapter {
	return &MCPToolAdapter{
		client:  client,
		mcpTool: mcpTool,
	}
}

// Name returns the tool's unique identifier
func (a *MCPToolAdapter) Name() string {
	return a.mcpTool.Name
}

// Description returns the tool's human-readable description
func (a *MCPToolAdapter) Description() string {
	return a.mcpTool.Description
}

// RequiresApproval determines if this tool execution needs user approval
// For MCP tools, we default to not requiring approval unless specific
// patterns are detected (can be enhanced later)
func (a *MCPToolAdapter) RequiresApproval(params map[string]interface{}) bool {
	// Future enhancement: check tool name patterns, params, etc.
	// For now, MCP tools don't require approval by default
	return false
}

// Execute runs the MCP tool and converts the result to Clem's Result format
func (a *MCPToolAdapter) Execute(ctx context.Context, params map[string]interface{}) (*tools.Result, error) {
	// Ensure params is non-nil
	if params == nil {
		params = make(map[string]interface{})
	}

	// Call the MCP tool via the client
	mcpResult, err := a.client.CallTool(ctx, a.mcpTool.Name, params)
	if err != nil {
		// Check if it's a context cancellation
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		// MCP tool execution failed - return as a failed Result
		return &tools.Result{
			ToolName: a.mcpTool.Name,
			Success:  false,
			Output:   "",
			Error:    err.Error(),
		}, nil
	}

	// Convert MCP result to Clem Result
	result := a.convertMCPResult(mcpResult)
	return result, nil
}

// convertMCPResult converts an MCP tool response to a Clem Result
func (a *MCPToolAdapter) convertMCPResult(mcpResult map[string]interface{}) *tools.Result {
	result := &tools.Result{
		ToolName: a.mcpTool.Name,
		Success:  true,
		Metadata: make(map[string]interface{}),
	}

	// Extract content array
	content, ok := mcpResult["content"].([]interface{})
	if !ok {
		// No content or invalid format
		result.Success = false
		result.Error = "Invalid MCP response: missing or invalid content"
		return result
	}

	// Process content blocks
	var outputParts []string
	for _, item := range content {
		contentBlock, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		contentType, _ := contentBlock["type"].(string)
		switch contentType {
		case "text":
			if text, ok := contentBlock["text"].(string); ok {
				outputParts = append(outputParts, text)
			}
		case "image":
			// Future: handle image content
			result.Metadata["has_image"] = true
		case "resource":
			// Future: handle resource content
			result.Metadata["has_resource"] = true
		}
	}

	result.Output = strings.Join(outputParts, "\n")

	// Store original MCP result in metadata for debugging
	result.Metadata["mcp_result"] = mcpResult

	return result
}

// AsToolDefinition converts the MCP tool to Clem's ToolDefinition format
// This is used when registering the tool with the API
func (a *MCPToolAdapter) AsToolDefinition() core.ToolDefinition {
	return core.ToolDefinition{
		Name:        a.mcpTool.Name,
		Description: a.mcpTool.Description,
		InputSchema: a.mcpTool.InputSchema,
	}
}

// GetInputSchema returns the tool's input schema
func (a *MCPToolAdapter) GetInputSchema() map[string]interface{} {
	return a.mcpTool.InputSchema
}

// MCPToolManager manages MCP tool adapters for a server
type MCPToolManager struct {
	client  *Client
	tools   map[string]*MCPToolAdapter
	mu      sync.RWMutex
}

// NewMCPToolManager creates a new tool manager for an MCP server
func NewMCPToolManager(client *Client) *MCPToolManager {
	return &MCPToolManager{
		client: client,
		tools:  make(map[string]*MCPToolAdapter),
	}
}

// RefreshTools fetches the latest tool list from the MCP server
func (m *MCPToolManager) RefreshTools(ctx context.Context) error {
	mcpTools, err := m.client.ListTools(ctx)
	if err != nil {
		return fmt.Errorf("failed to list tools: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Clear existing tools
	m.tools = make(map[string]*MCPToolAdapter)

	// Create adapters for each tool
	for _, mcpTool := range mcpTools {
		adapter := NewMCPToolAdapter(m.client, mcpTool)
		m.tools[mcpTool.Name] = adapter
	}

	return nil
}

// GetTools returns all available MCP tools as Clem tools
func (m *MCPToolManager) GetTools() []tools.Tool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]tools.Tool, 0, len(m.tools))
	for _, adapter := range m.tools {
		result = append(result, adapter)
	}

	return result
}

// GetTool returns a specific MCP tool by name
func (m *MCPToolManager) GetTool(name string) tools.Tool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.tools[name]
}

// Count returns the number of available tools
func (m *MCPToolManager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.tools)
}
