// ABOUTME: Email productivity tools that route to provider implementations
// ABOUTME: Send, read, search, and manage emails through active provider

package tools

import (
	"context"
	"fmt"

	"github.com/harper/jeff/internal/providers"
)

// SendEmailTool sends email messages
type SendEmailTool struct {
	registry *providers.Registry
}

// NewSendEmailTool creates a new send email tool
func NewSendEmailTool(registry *providers.Registry) Tool {
	return &SendEmailTool{registry: registry}
}

// Name returns the tool's identifier
func (t *SendEmailTool) Name() string {
	return "send_email"
}

// Description returns what this tool does
func (t *SendEmailTool) Description() string {
	return "Send an email message with optional attachments"
}

// RequiresApproval returns whether this tool needs user confirmation
func (t *SendEmailTool) RequiresApproval(_ map[string]interface{}) bool {
	// Sending email requires approval
	return true
}

// Execute runs the send email operation
func (t *SendEmailTool) Execute(_ context.Context, params map[string]interface{}) (*Result, error) {
	result, err := t.registry.ExecuteTool("send_email", params)
	if err != nil {
		return &Result{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &Result{
		Success: result.Success,
		Output:  formatProviderResult(result),
		Error:   result.Error,
	}, nil
}

// SearchEmailsTool searches for emails
type SearchEmailsTool struct {
	registry *providers.Registry
}

// NewSearchEmailsTool creates a new search emails tool
func NewSearchEmailsTool(registry *providers.Registry) Tool {
	return &SearchEmailsTool{registry: registry}
}

// Name returns the tool's identifier
func (t *SearchEmailsTool) Name() string {
	return "search_emails"
}

// Description returns what this tool does
func (t *SearchEmailsTool) Description() string {
	return "Search emails with filters (query, from, to, date range, read status)"
}

// RequiresApproval returns whether this tool needs user confirmation
func (t *SearchEmailsTool) RequiresApproval(_ map[string]interface{}) bool {
	// Reading emails requires approval
	return true
}

// Execute runs the search emails operation
func (t *SearchEmailsTool) Execute(_ context.Context, params map[string]interface{}) (*Result, error) {
	result, err := t.registry.ExecuteTool("search_emails", params)
	if err != nil {
		return &Result{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &Result{
		Success: result.Success,
		Output:  formatProviderResult(result),
		Error:   result.Error,
	}, nil
}

// ReadEmailTool reads a specific email
type ReadEmailTool struct {
	registry *providers.Registry
}

// NewReadEmailTool creates a new read email tool
func NewReadEmailTool(registry *providers.Registry) Tool {
	return &ReadEmailTool{registry: registry}
}

// Name returns the tool's identifier
func (t *ReadEmailTool) Name() string {
	return "read_email"
}

// Description returns what this tool does
func (t *ReadEmailTool) Description() string {
	return "Read the full content of a specific email by message ID"
}

// RequiresApproval returns whether this tool needs user confirmation
func (t *ReadEmailTool) RequiresApproval(_ map[string]interface{}) bool {
	return true
}

// Execute runs the read email operation
func (t *ReadEmailTool) Execute(_ context.Context, params map[string]interface{}) (*Result, error) {
	result, err := t.registry.ExecuteTool("read_email", params)
	if err != nil {
		return &Result{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &Result{
		Success: result.Success,
		Output:  formatProviderResult(result),
		Error:   result.Error,
	}, nil
}

// Helper function to format provider results for display
func formatProviderResult(result providers.ToolResult) string {
	if result.Error != "" {
		return fmt.Sprintf("Error: %s", result.Error)
	}

	if result.Data == nil {
		return "Operation completed successfully"
	}

	// Format data based on type
	switch data := result.Data.(type) {
	case string:
		return data
	case map[string]interface{}:
		return fmt.Sprintf("%v", data)
	case []interface{}:
		return fmt.Sprintf("Returned %d items", len(data))
	default:
		return fmt.Sprintf("%v", data)
	}
}
