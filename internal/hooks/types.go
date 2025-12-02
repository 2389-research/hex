// ABOUTME: Type definitions for the hooks system
// ABOUTME: Defines hook events, configuration, and execution context

// Package hooks implements the lifecycle hooks system for automation.
package hooks

import "fmt"

// HookEvent represents a lifecycle event that can trigger hooks
type HookEvent string

// Lifecycle hook events that can trigger hook execution
const (
	SessionStart      HookEvent = "SessionStart"
	SessionEnd        HookEvent = "SessionEnd"
	UserPromptSubmit  HookEvent = "UserPromptSubmit"
	ModelResponseDone HookEvent = "ModelResponseDone"
	PreToolUse        HookEvent = "PreToolUse"
	PostToolUse       HookEvent = "PostToolUse"
	PreCommit         HookEvent = "PreCommit"
	PostCommit        HookEvent = "PostCommit"
	OnError           HookEvent = "OnError"
	PlanModeEnter     HookEvent = "PlanModeEnter"
)

// HookConfig defines a single hook's configuration
type HookConfig struct {
	Command string            `yaml:"command" json:"command"`
	Matcher map[string]string `yaml:"matcher,omitempty" json:"matcher,omitempty"`
	Timeout int               `yaml:"timeout,omitempty" json:"timeout,omitempty"` // seconds
}

// Validate checks if the hook configuration is valid
func (h *HookConfig) Validate() error {
	if h.Command == "" {
		return fmt.Errorf("hook command cannot be empty")
	}
	if h.Timeout < 0 {
		return fmt.Errorf("hook timeout cannot be negative")
	}
	return nil
}

// HooksConfig maps hook events to their configurations
//
//nolint:revive // Name matches Claude Code convention
type HooksConfig map[HookEvent][]HookConfig

// EventData contains context passed to hooks
type EventData map[string]interface{}
