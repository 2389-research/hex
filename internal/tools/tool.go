// ABOUTME: Tool interface definition and core abstractions
// ABOUTME: Defines the contract that all tools must implement

package tools

import "context"

// Tool represents an executable tool that can be invoked by the assistant.
// Implementations must:
//   - Validate their own parameters in Execute() and return appropriate errors
//   - Handle type assertions safely (params may have incorrect types)
//   - Be safe for concurrent execution from multiple goroutines
type Tool interface {
	// Name returns the unique identifier for this tool (e.g., "read_file", "bash")
	Name() string

	// Description returns a human-readable description of what this tool does
	Description() string

	// RequiresApproval returns true if this specific tool invocation needs user approval.
	// The decision can be based on the parameters (e.g., writing to /etc requires approval).
	// params is guaranteed to be non-nil but may be empty or have unexpected types.
	RequiresApproval(params map[string]interface{}) bool

	// Execute runs the tool with the given parameters.
	// Implementations must:
	//   - Validate required parameters exist
	//   - Type assert parameters safely (return error on wrong type)
	//   - Respect context cancellation
	//   - Return a Result (not error) for expected failures like "file not found"
	//   - Return an error only for unexpected/catastrophic failures
	// params is guaranteed to be non-nil but may be empty.
	Execute(ctx context.Context, params map[string]interface{}) (*Result, error)
}
