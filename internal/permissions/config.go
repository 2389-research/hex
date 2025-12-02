// ABOUTME: Permission configuration loading and management
// ABOUTME: Handles configuration from CLI flags and config files

package permissions

import (
	"fmt"
)

// Config holds permission configuration
type Config struct {
	// Mode is the permission mode (auto, ask, deny)
	Mode Mode

	// AllowedTools is a list of allowed tool names
	AllowedTools []string

	// DisallowedTools is a list of disallowed tool names
	DisallowedTools []string
}

// DefaultConfig returns the default permission configuration
func DefaultConfig() *Config {
	return &Config{
		Mode:            ModeAsk,
		AllowedTools:    nil,
		DisallowedTools: nil,
	}
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Check for conflicting flags
	// Note: having both allowed and disallowed lists is valid - disallowed takes precedence

	// Auto mode with deny list is valid (deny list has no effect)
	// Deny mode with allow list is valid (deny mode overrides)

	return nil
}

// ToChecker converts config to a permission checker
func (c *Config) ToChecker() (*Checker, error) {
	if err := c.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	rules := NewRules(c.AllowedTools, c.DisallowedTools)
	return NewChecker(c.Mode, rules), nil
}

// String returns a human-readable representation of the config
func (c *Config) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("mode=%s", c.Mode))

	if len(c.AllowedTools) > 0 {
		parts = append(parts, fmt.Sprintf("allowed=%v", c.AllowedTools))
	}

	if len(c.DisallowedTools) > 0 {
		parts = append(parts, fmt.Sprintf("disallowed=%v", c.DisallowedTools))
	}

	result := "Permissions("
	for i, part := range parts {
		if i > 0 {
			result += ", "
		}
		result += part
	}
	result += ")"

	return result
}
