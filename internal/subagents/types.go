// ABOUTME: Subagent type definitions and configuration structures
// ABOUTME: Defines the four predefined subagent types and their characteristics
//
// Example usage:
//
//	// Create an executor
//	executor := subagents.NewExecutor()
//
//	// Execute an Explore subagent
//	req := &subagents.ExecutionRequest{
//		Type:        subagents.TypeExplore,
//		Prompt:      "Find all authentication code",
//		Description: "Explore authentication system",
//	}
//	result, err := executor.Execute(ctx, req)
//
//	// Use parallel dispatch for multiple independent tasks
//	dispatcher := subagents.NewDispatcher(executor)
//	requests := []*subagents.DispatchRequest{
//		{ID: "task1", Request: &subagents.ExecutionRequest{...}},
//		{ID: "task2", Request: &subagents.ExecutionRequest{...}},
//	}
//	results := dispatcher.DispatchParallel(ctx, requests)
//
//	// Execute with hooks
//	result, err := executor.ExecuteWithHooks(ctx, req, hookEngine)

package subagents

import (
	"time"
)

// SubagentType represents a predefined subagent type with specialized behavior
type SubagentType string

const (
	// TypeGeneralPurpose is the default subagent for general tasks
	TypeGeneralPurpose SubagentType = "general-purpose"

	// TypeExplore is optimized for fast codebase exploration and research
	TypeExplore SubagentType = "Explore"

	// TypePlan is specialized for design and planning work
	TypePlan SubagentType = "Plan"

	// TypeCodeReviewer performs code review and quality checks
	TypeCodeReviewer SubagentType = "code-reviewer"
)

// ValidSubagentTypes returns all valid subagent type strings
func ValidSubagentTypes() []string {
	return []string{
		string(TypeGeneralPurpose),
		string(TypeExplore),
		string(TypePlan),
		string(TypeCodeReviewer),
	}
}

// IsValid checks if a subagent type string is valid
func IsValid(t string) bool {
	switch SubagentType(t) {
	case TypeGeneralPurpose, TypeExplore, TypePlan, TypeCodeReviewer:
		return true
	default:
		return false
	}
}

// Config holds configuration for a subagent instance
type Config struct {
	// Type is the subagent type (general-purpose, Explore, Plan, code-reviewer)
	Type SubagentType

	// Model is the Claude model to use (optional, inherits from parent if empty)
	Model string

	// Timeout is the maximum execution time (0 = use default)
	Timeout time.Duration

	// MaxTokens is the maximum response length (0 = use default)
	MaxTokens int

	// Temperature controls randomness (0.0-1.0, 0 = use default)
	Temperature float64

	// AllowedTools restricts which tools this subagent can use (nil = all tools)
	AllowedTools []string

	// SystemPrompt overrides the default system prompt for this type (optional)
	SystemPrompt string
}

// DefaultConfig returns the default configuration for a subagent type
func DefaultConfig(t SubagentType) *Config {
	config := &Config{
		Type:         t,
		Timeout:      5 * time.Minute,
		MaxTokens:    4096,
		Temperature:  1.0,
		AllowedTools: nil, // All tools allowed by default
	}

	// Customize based on type
	switch t {
	case TypeExplore:
		// Explorer is read-only and thorough
		config.AllowedTools = []string{"Read", "Grep", "Glob", "Bash"}
		config.Temperature = 0.7
		config.MaxTokens = 8192 // Longer responses for detailed analysis

	case TypePlan:
		// Planner is read-only and strategic
		config.AllowedTools = []string{"Read", "Grep", "Glob"}
		config.Temperature = 0.6
		config.MaxTokens = 6144

	case TypeCodeReviewer:
		// Reviewer is read-only and critical
		config.AllowedTools = []string{"Read", "Grep", "Glob"}
		config.Temperature = 0.3
		config.MaxTokens = 6144

	case TypeGeneralPurpose:
		// General purpose gets all tools
		config.AllowedTools = nil
		config.Temperature = 1.0
		config.MaxTokens = 4096
	}

	return config
}

// Result contains the outcome of a subagent execution
type Result struct {
	// Success indicates if the subagent completed successfully
	Success bool

	// Output is the subagent's response text
	Output string

	// Error contains error message if Success is false
	Error string

	// Type is the subagent type that was executed
	Type SubagentType

	// Metadata contains execution details (duration, tokens, etc.)
	Metadata map[string]interface{}

	// StartTime is when execution began
	StartTime time.Time

	// EndTime is when execution completed
	EndTime time.Time
}

// Duration returns how long the subagent took to execute
func (r *Result) Duration() time.Duration {
	if r.EndTime.IsZero() || r.StartTime.IsZero() {
		return 0
	}
	return r.EndTime.Sub(r.StartTime)
}

// ExecutionRequest contains all information needed to execute a subagent
type ExecutionRequest struct {
	// Type is the subagent type to execute
	Type SubagentType

	// Prompt is the task description for the subagent
	Prompt string

	// Description is a human-readable summary of what this subagent will do
	Description string

	// Config overrides default configuration (optional)
	Config *Config

	// Context contains additional context to inject (optional)
	Context map[string]string
}

// Validate checks if the execution request is valid
func (r *ExecutionRequest) Validate() error {
	if r.Type == "" {
		return &ValidationError{Field: "Type", Message: "subagent type is required"}
	}

	if !IsValid(string(r.Type)) {
		return &ValidationError{
			Field:   "Type",
			Message: "invalid subagent type",
			Details: map[string]interface{}{"type": r.Type, "valid_types": ValidSubagentTypes()},
		}
	}

	if r.Prompt == "" {
		return &ValidationError{Field: "Prompt", Message: "prompt is required"}
	}

	if r.Description == "" {
		return &ValidationError{Field: "Description", Message: "description is required"}
	}

	return nil
}

// ValidationError represents a validation failure
type ValidationError struct {
	Field   string
	Message string
	Details map[string]interface{}
}

// Error implements the error interface
func (e *ValidationError) Error() string {
	if len(e.Details) == 0 {
		return e.Field + ": " + e.Message
	}
	return e.Field + ": " + e.Message
}
