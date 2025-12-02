// ABOUTME: Permission rules for allow/deny lists
// ABOUTME: Handles tool filtering based on allowed and disallowed tool lists

package permissions

import (
	"strings"
)

// Rules defines which tools are allowed or blocked
type Rules struct {
	// AllowedTools is a whitelist of tool names
	// If non-empty, only these tools are allowed
	AllowedTools []string

	// DisallowedTools is a blacklist of tool names
	// These tools are blocked even if in allowed list
	DisallowedTools []string
}

// NewRules creates a new permission rules instance
func NewRules(allowed, disallowed []string) *Rules {
	return &Rules{
		AllowedTools:    normalizeToolNames(allowed),
		DisallowedTools: normalizeToolNames(disallowed),
	}
}

// IsToolAllowed checks if a tool is allowed by the rules
// Logic:
// 1. If tool is in disallowed list, return false
// 2. If allowed list is empty, return true (no whitelist)
// 3. If tool is in allowed list, return true
// 4. Otherwise return false
func (r *Rules) IsToolAllowed(toolName string) bool {
	toolName = normalizeToolName(toolName)

	// Check disallowed list first (blacklist has priority)
	if r.isDisallowed(toolName) {
		return false
	}

	// If no allowed list, all tools are allowed (except disallowed)
	if len(r.AllowedTools) == 0 {
		return true
	}

	// Check if tool is in allowed list
	return r.isAllowed(toolName)
}

// isAllowed checks if a tool is in the allowed list
func (r *Rules) isAllowed(toolName string) bool {
	for _, allowed := range r.AllowedTools {
		if matchToolName(allowed, toolName) {
			return true
		}
	}
	return false
}

// isDisallowed checks if a tool is in the disallowed list
func (r *Rules) isDisallowed(toolName string) bool {
	for _, disallowed := range r.DisallowedTools {
		if matchToolName(disallowed, toolName) {
			return true
		}
	}
	return false
}

// matchToolName checks if two tool names match
// Handles variations like "Read" vs "read_file" vs "ReadFile"
func matchToolName(pattern, toolName string) bool {
	pattern = normalizeToolName(pattern)
	toolName = normalizeToolName(toolName)

	// Exact match
	if pattern == toolName {
		return true
	}

	// Try with/without _file, _tool suffixes
	suffixes := []string{"_file", "_tool", "tool", "file"}
	for _, suffix := range suffixes {
		if pattern+suffix == toolName || pattern == toolName+suffix {
			return true
		}
	}

	return false
}

// normalizeToolName converts a tool name to lowercase and removes underscores
// for flexible matching (e.g., "ReadFile", "read_file", "Read" all match)
func normalizeToolName(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, "_", "")
	return name
}

// normalizeToolNames normalizes a list of tool names
func normalizeToolNames(names []string) []string {
	normalized := make([]string, 0, len(names))
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name != "" {
			normalized = append(normalized, name)
		}
	}
	return normalized
}

// HasAllowList returns true if an allowed tools list is configured
func (r *Rules) HasAllowList() bool {
	return len(r.AllowedTools) > 0
}

// HasDenyList returns true if a disallowed tools list is configured
func (r *Rules) HasDenyList() bool {
	return len(r.DisallowedTools) > 0
}
