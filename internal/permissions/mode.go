// ABOUTME: Permission mode definitions and parsing
// ABOUTME: Defines how tool execution permissions are handled (auto, ask, deny)

package permissions

import (
	"fmt"
	"strings"
)

// Mode represents how permission requests are handled
type Mode int

const (
	// ModeAsk prompts the user for each tool execution (default)
	ModeAsk Mode = iota
	// ModeAuto automatically approves all tool executions
	ModeAuto
	// ModeDeny blocks all tool executions
	ModeDeny
)

// String returns the string representation of the permission mode
func (m Mode) String() string {
	switch m {
	case ModeAsk:
		return "ask"
	case ModeAuto:
		return "auto"
	case ModeDeny:
		return "deny"
	default:
		return "unknown"
	}
}

// ParseMode parses a string into a permission mode
func ParseMode(s string) (Mode, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "ask":
		return ModeAsk, nil
	case "auto":
		return ModeAuto, nil
	case "deny":
		return ModeDeny, nil
	default:
		return ModeAsk, fmt.Errorf("invalid permission mode %q: must be 'auto', 'ask', or 'deny'", s)
	}
}
