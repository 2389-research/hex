// ABOUTME: Permission checker combining mode and rules
// ABOUTME: Core permission logic that determines if a tool should be allowed to execute

// Package permissions provides tool execution permission checking.
package permissions

import (
	"fmt"
)

// Checker evaluates whether a tool execution should be allowed
type Checker struct {
	mode  Mode
	rules *Rules
}

// NewChecker creates a new permission checker
func NewChecker(mode Mode, rules *Rules) *Checker {
	if rules == nil {
		rules = NewRules(nil, nil)
	}
	return &Checker{
		mode:  mode,
		rules: rules,
	}
}

// CheckResult contains the result of a permission check
type CheckResult struct {
	// Allowed indicates if the tool is allowed to execute
	Allowed bool

	// RequiresPrompt indicates if user should be prompted for approval
	RequiresPrompt bool

	// Reason explains why the tool was allowed/denied
	Reason string

	// ToolName is the name of the tool being checked
	ToolName string
}

// Check evaluates if a tool should be allowed to execute
func (c *Checker) Check(toolName string, _ map[string]interface{}) CheckResult {
	result := CheckResult{
		ToolName: toolName,
	}

	// First check if tool is allowed by rules
	if !c.rules.IsToolAllowed(toolName) {
		result.Allowed = false
		result.RequiresPrompt = false

		if c.rules.HasAllowList() {
			result.Reason = fmt.Sprintf("tool %q not in allowed list", toolName)
		} else {
			result.Reason = fmt.Sprintf("tool %q is disallowed", toolName)
		}
		return result
	}

	// Tool is allowed by rules, now check mode
	switch c.mode {
	case ModeAuto:
		result.Allowed = true
		result.RequiresPrompt = false
		result.Reason = "auto-approved (permission mode: auto)"

	case ModeDeny:
		result.Allowed = false
		result.RequiresPrompt = false
		result.Reason = "denied (permission mode: deny)"

	case ModeAsk:
		result.Allowed = false // Will be determined by user prompt
		result.RequiresPrompt = true
		result.Reason = "user approval required (permission mode: ask)"

	default:
		result.Allowed = false
		result.RequiresPrompt = true
		result.Reason = "unknown permission mode, defaulting to ask"
	}

	return result
}

// GetMode returns the current permission mode
func (c *Checker) GetMode() Mode {
	return c.mode
}

// GetRules returns the current permission rules
func (c *Checker) GetRules() *Rules {
	return c.rules
}

// ShouldAutoApprove returns true if the checker auto-approves all tools
func (c *Checker) ShouldAutoApprove() bool {
	return c.mode == ModeAuto
}

// ShouldDenyAll returns true if the checker denies all tools
func (c *Checker) ShouldDenyAll() bool {
	return c.mode == ModeDeny
}
