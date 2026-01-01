// ABOUTME: Adapter that wraps MCP tools to implement Hex's Tool interface
// ABOUTME: Bridges mux MCP protocol tools with Hex's tool execution system

package mcp

import (
	"context"
	"strings"

	"github.com/2389-research/hex/internal/core"
	"github.com/2389-research/hex/internal/tools"
	muxmcp "github.com/2389-research/mux/mcp"
)

// MCPToolAdapter wraps an MCP tool to implement Hex's Tool interface
//
//nolint:revive // MCP prefix clarifies this adapts MCP tools vs other tool types
type MCPToolAdapter struct {
	client  muxmcp.Client
	mcpTool muxmcp.ToolInfo
}

// NewMCPToolAdapter creates a new adapter for an MCP tool
func NewMCPToolAdapter(client muxmcp.Client, mcpTool muxmcp.ToolInfo) *MCPToolAdapter {
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
func (a *MCPToolAdapter) RequiresApproval(_ map[string]interface{}) bool {
	// Future enhancement: check tool name patterns, params, etc.
	// For now, MCP tools don't require approval by default
	return false
}

// Execute runs the MCP tool and converts the result to Hex's Result format
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

	// Convert MCP result to Hex Result
	result := a.convertMCPResult(mcpResult)
	return result, nil
}

// convertMCPResult converts an MCP ToolCallResult to a Hex Result
func (a *MCPToolAdapter) convertMCPResult(mcpResult *muxmcp.ToolCallResult) *tools.Result {
	result := &tools.Result{
		ToolName: a.mcpTool.Name,
		Success:  !mcpResult.IsError,
		Metadata: make(map[string]interface{}),
	}

	// Process content blocks
	var outputParts []string
	for _, block := range mcpResult.Content {
		switch block.Type {
		case "text":
			outputParts = append(outputParts, block.Text)
		case "image":
			result.Metadata["has_image"] = true
			if block.MimeType != "" {
				result.Metadata["image_mime_type"] = block.MimeType
			}
		case "resource":
			result.Metadata["has_resource"] = true
		}
	}

	result.Output = strings.Join(outputParts, "\n")

	if mcpResult.IsError && result.Output != "" {
		result.Error = result.Output
	}

	return result
}

// AsToolDefinition converts the MCP tool to Hex's ToolDefinition format
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
