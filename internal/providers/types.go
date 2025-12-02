// ABOUTME: Core types and interfaces for the provider system
// ABOUTME: Defines Provider interface and related types for productivity tool backends

package providers

import (
	"time"
)

// Provider is the interface that all productivity providers must implement
// Providers handle authentication and execution of productivity tools (email, calendar, tasks)
type Provider interface {
	// Metadata
	Name() string             // Provider name (e.g., "gmail", "outlook")
	SupportedTools() []string // List of tool names this provider implements

	// Lifecycle
	Initialize(config map[string]string) error // Setup and configuration
	Authenticate() error                       // Perform OAuth or other auth flow
	Close() error                              // Cleanup resources

	// Health (for future multi-provider routing)
	Status() ProviderStatus             // Current provider health status
	Capabilities() ProviderCapabilities // Provider capabilities and limits

	// Tool execution
	ExecuteTool(toolName string, params map[string]interface{}) (ToolResult, error)
}

// ProviderStatus represents the current health status of a provider
type ProviderStatus struct {
	Healthy   bool      // Whether provider is operational
	Message   string    // Status message or error description
	LastCheck time.Time // When status was last checked
}

// ProviderCapabilities describes what a provider can do
type ProviderCapabilities struct {
	RateLimits map[string]int // tool_name -> calls per hour
	Features   []string       // Feature flags (e.g., "attachments", "calendar_sharing")
	MaxResults int            // Maximum results for list/search operations
}

// ToolResult is returned from tool execution
type ToolResult struct {
	Success  bool                   `json:"success"`            // Whether operation succeeded
	Data     interface{}            `json:"data,omitempty"`     // Result data (format varies by tool)
	Error    string                 `json:"error,omitempty"`    // Error message if failed
	Metadata map[string]interface{} `json:"metadata,omitempty"` // Additional provider-specific info
}

// ProviderInfo contains metadata about a registered provider
type ProviderInfo struct {
	Name           string   // Provider name
	Authenticated  bool     // Whether provider has valid credentials
	Active         bool     // Whether this is the active provider
	SupportedTools []string // Tools this provider implements
}

// Common error messages
const (
	ErrNotImplemented      = "tool not implemented by this provider"
	ErrNotAuthenticated    = "provider not authenticated"
	ErrRateLimitExceeded   = "rate limit exceeded"
	ErrInvalidParameters   = "invalid tool parameters"
	ErrProviderUnavailable = "provider temporarily unavailable"
)
