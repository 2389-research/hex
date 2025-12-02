// ABOUTME: Hook event types and data structures
// ABOUTME: Defines the 10 official hook events and their context data

package hooks

import (
	"fmt"
	"time"
)

// EventType represents the type of hook event
type EventType string

const (
	// EventSessionStart fires when a Claude Code session begins (before first user message)
	EventSessionStart EventType = "SessionStart"

	// EventSessionEnd fires when Claude Code session terminates (clean shutdown)
	EventSessionEnd EventType = "SessionEnd"

	// EventUserPromptSubmit fires when user sends a message (before Claude processes it)
	EventUserPromptSubmit EventType = "UserPromptSubmit"

	// EventPreToolUse fires after Claude creates tool parameters, before tool execution
	EventPreToolUse EventType = "PreToolUse"

	// EventPostToolUse fires immediately after a tool completes successfully
	EventPostToolUse EventType = "PostToolUse"

	// EventPermissionRequest fires when user is shown a permission dialog for a tool
	EventPermissionRequest EventType = "PermissionRequest"

	// EventNotification fires when Claude Code sends notifications to the user
	EventNotification EventType = "Notification"

	// EventStop fires when main Claude Code agent finishes responding to user
	EventStop EventType = "Stop"

	// EventSubagentStop fires when a subagent finishes responding
	EventSubagentStop EventType = "SubagentStop"

	// EventPreCompact fires before Claude Code runs a compact operation
	EventPreCompact EventType = "PreCompact"
)

// Event represents a hook event with associated context data
type Event struct {
	Type      EventType
	Timestamp time.Time
	Data      EventData
}

// EventData contains context-specific data for each event type
type EventData interface {
	// ToEnvVars converts the event data to environment variables for shell execution
	ToEnvVars() map[string]string
}

// SessionStartData contains data for SessionStart event
type SessionStartData struct {
	ProjectPath string
	ModelID     string
}

// ToEnvVars converts SessionStartData to environment variables
func (d SessionStartData) ToEnvVars() map[string]string {
	return map[string]string{
		"CLAUDE_PROJECT_PATH": d.ProjectPath,
		"CLAUDE_MODEL_ID":     d.ModelID,
	}
}

// SessionEndData contains data for SessionEnd event
type SessionEndData struct {
	ProjectPath  string
	MessageCount int
}

// ToEnvVars converts SessionEndData to environment variables
func (d SessionEndData) ToEnvVars() map[string]string {
	return map[string]string{
		"CLAUDE_PROJECT_PATH":  d.ProjectPath,
		"CLAUDE_MESSAGE_COUNT": intToString(d.MessageCount),
	}
}

// UserPromptSubmitData contains data for UserPromptSubmit event
type UserPromptSubmitData struct {
	MessageText   string
	MessageLength int
}

// ToEnvVars converts UserPromptSubmitData to environment variables
func (d UserPromptSubmitData) ToEnvVars() map[string]string {
	return map[string]string{
		"CLAUDE_MESSAGE_TEXT":   d.MessageText,
		"CLAUDE_MESSAGE_LENGTH": intToString(d.MessageLength),
	}
}

// ToolUseData contains data for PreToolUse and PostToolUse events
type ToolUseData struct {
	ToolName   string
	FilePath   string // extracted from parameters if available
	Success    bool   // only for PostToolUse
	Error      string // only for PostToolUse (if failed)
	IsSubagent bool
}

// ToEnvVars converts ToolUseData to environment variables
func (d ToolUseData) ToEnvVars() map[string]string {
	env := map[string]string{
		"CLAUDE_TOOL_NAME":      d.ToolName,
		"CLAUDE_TOOL_FILE_PATH": d.FilePath,
		"CLAUDE_IS_SUBAGENT":    boolToString(d.IsSubagent),
	}
	// Only include success/error for PostToolUse
	if d.Success || d.Error != "" {
		env["CLAUDE_TOOL_SUCCESS"] = boolToString(d.Success)
		if d.Error != "" {
			env["CLAUDE_ERROR"] = d.Error
		}
	}
	return env
}

// PermissionRequestData contains data for PermissionRequest event
type PermissionRequestData struct {
	ToolName    string
	Action      string
	Description string
	IsSubagent  bool
}

// ToEnvVars converts PermissionRequestData to environment variables
func (d PermissionRequestData) ToEnvVars() map[string]string {
	return map[string]string{
		"CLAUDE_TOOL_NAME":   d.ToolName,
		"CLAUDE_ACTION":      d.Action,
		"CLAUDE_DESCRIPTION": d.Description,
		"CLAUDE_IS_SUBAGENT": boolToString(d.IsSubagent),
	}
}

// NotificationData contains data for Notification event
type NotificationData struct {
	Level   string
	Message string
	Source  string
}

// ToEnvVars converts NotificationData to environment variables
func (d NotificationData) ToEnvVars() map[string]string {
	return map[string]string{
		"CLAUDE_NOTIFICATION_LEVEL":   d.Level,
		"CLAUDE_NOTIFICATION_MESSAGE": d.Message,
		"CLAUDE_NOTIFICATION_SOURCE":  d.Source,
	}
}

// StopData contains data for Stop event
type StopData struct {
	ResponseLength int
	TokensUsed     int
	ToolsUsed      []string
	IsSubagent     bool
}

// ToEnvVars converts StopData to environment variables
func (d StopData) ToEnvVars() map[string]string {
	env := map[string]string{
		"CLAUDE_RESPONSE_LENGTH": intToString(d.ResponseLength),
		"CLAUDE_TOKENS_USED":     intToString(d.TokensUsed),
		"CLAUDE_IS_SUBAGENT":     boolToString(d.IsSubagent),
	}
	if len(d.ToolsUsed) > 0 {
		// Join tools with commas
		tools := ""
		for i, tool := range d.ToolsUsed {
			if i > 0 {
				tools += ","
			}
			tools += tool
		}
		env["CLAUDE_TOOLS_USED"] = tools
	}
	return env
}

// SubagentStopData contains data for SubagentStop event
type SubagentStopData struct {
	TaskDescription string
	ResponseLength  int
	TokensUsed      int
}

// ToEnvVars converts SubagentStopData to environment variables
func (d SubagentStopData) ToEnvVars() map[string]string {
	return map[string]string{
		"CLAUDE_TASK_DESCRIPTION": d.TaskDescription,
		"CLAUDE_RESPONSE_LENGTH":  intToString(d.ResponseLength),
		"CLAUDE_TOKENS_USED":      intToString(d.TokensUsed),
		"CLAUDE_IS_SUBAGENT":      "true",
	}
}

// PreCompactData contains data for PreCompact event
type PreCompactData struct {
	CompactionType   string
	CurrentSize      int
	EstimatedNewSize int
}

// ToEnvVars converts PreCompactData to environment variables
func (d PreCompactData) ToEnvVars() map[string]string {
	return map[string]string{
		"CLAUDE_COMPACTION_TYPE":    d.CompactionType,
		"CLAUDE_CURRENT_SIZE":       intToString(d.CurrentSize),
		"CLAUDE_ESTIMATED_NEW_SIZE": intToString(d.EstimatedNewSize),
	}
}

// Helper functions for type conversion
func intToString(i int) string {
	// Import fmt at the top if needed
	return fmt.Sprintf("%d", i)
}

func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
