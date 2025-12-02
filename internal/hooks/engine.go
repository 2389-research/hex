// ABOUTME: Hook execution engine and orchestration
// ABOUTME: Coordinates hook execution, filtering, and error handling

package hooks

import (
	"errors"
	"fmt"
	"time"
)

// Engine manages hook execution across the application lifecycle
type Engine struct {
	config   *Config
	executor *Executor
	enabled  bool
}

// NewEngine creates a new hook engine
func NewEngine(projectPath, modelID string) (*Engine, error) {
	config, err := LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("load hook config: %w", err)
	}

	executor := NewExecutor(projectPath, modelID)

	return &Engine{
		config:   config,
		executor: executor,
		enabled:  true,
	}, nil
}

// NewEngineWithConfig creates an engine with a specific configuration
// Useful for testing
func NewEngineWithConfig(config *Config, projectPath, modelID string) *Engine {
	executor := NewExecutor(projectPath, modelID)
	return &Engine{
		config:   config,
		executor: executor,
		enabled:  true,
	}
}

// Disable turns off hook execution
func (e *Engine) Disable() {
	e.enabled = false
}

// Enable turns on hook execution
func (e *Engine) Enable() {
	e.enabled = true
}

// Fire executes all hooks for a given event
// Collects and returns all errors from failed hooks as a multi-error
func (e *Engine) Fire(eventType EventType, data EventData) error {
	if !e.enabled {
		return nil
	}

	event := &Event{
		Type:      eventType,
		Timestamp: time.Now(),
		Data:      data,
	}

	hooks := e.config.GetHooks(eventType)
	if len(hooks) == 0 {
		return nil
	}

	// Collect all errors from failed hooks
	var errs []error
	for _, hook := range hooks {
		// Check if hook should execute based on matchers
		if !hook.ShouldExecute(event) {
			continue
		}

		// Execute hook
		if hook.Async {
			e.executor.ExecuteAsync(&hook, event)
		} else {
			result := e.executor.Execute(&hook, event)
			if !result.Success && !hook.IgnoreFailure {
				// Include hook command in error for debugging
				err := fmt.Errorf("hook failed (command: %s): %w (stderr: %s)", hook.Command, result.Error, result.Stderr)
				errs = append(errs, err)
			}
		}
	}

	// Return combined error if any hooks failed
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// FireSessionStart is a convenience method for SessionStart event
func (e *Engine) FireSessionStart(projectPath, modelID string) error {
	return e.Fire(EventSessionStart, SessionStartData{
		ProjectPath: projectPath,
		ModelID:     modelID,
	})
}

// FireSessionEnd is a convenience method for SessionEnd event
func (e *Engine) FireSessionEnd(projectPath string, messageCount int) error {
	return e.Fire(EventSessionEnd, SessionEndData{
		ProjectPath:  projectPath,
		MessageCount: messageCount,
	})
}

// FireUserPromptSubmit is a convenience method for UserPromptSubmit event
func (e *Engine) FireUserPromptSubmit(messageText string) error {
	return e.Fire(EventUserPromptSubmit, UserPromptSubmitData{
		MessageText:   messageText,
		MessageLength: len(messageText),
	})
}

// FirePreToolUse is a convenience method for PreToolUse event
func (e *Engine) FirePreToolUse(toolName, filePath string, isSubagent bool) error {
	return e.Fire(EventPreToolUse, ToolUseData{
		ToolName:   toolName,
		FilePath:   filePath,
		IsSubagent: isSubagent,
	})
}

// FirePostToolUse is a convenience method for PostToolUse event
func (e *Engine) FirePostToolUse(toolName, filePath string, success bool, errMsg string, isSubagent bool) error {
	return e.Fire(EventPostToolUse, ToolUseData{
		ToolName:   toolName,
		FilePath:   filePath,
		Success:    success,
		Error:      errMsg,
		IsSubagent: isSubagent,
	})
}

// FirePermissionRequest is a convenience method for PermissionRequest event
func (e *Engine) FirePermissionRequest(toolName, action, description string, isSubagent bool) error {
	return e.Fire(EventPermissionRequest, PermissionRequestData{
		ToolName:    toolName,
		Action:      action,
		Description: description,
		IsSubagent:  isSubagent,
	})
}

// FireNotification is a convenience method for Notification event
func (e *Engine) FireNotification(level, message, source string) error {
	return e.Fire(EventNotification, NotificationData{
		Level:   level,
		Message: message,
		Source:  source,
	})
}

// FireStop is a convenience method for Stop event
func (e *Engine) FireStop(responseLength, tokensUsed int, toolsUsed []string, isSubagent bool) error {
	return e.Fire(EventStop, StopData{
		ResponseLength: responseLength,
		TokensUsed:     tokensUsed,
		ToolsUsed:      toolsUsed,
		IsSubagent:     isSubagent,
	})
}

// FireSubagentStop is a convenience method for SubagentStop event
func (e *Engine) FireSubagentStop(taskDescription string, responseLength, tokensUsed int) error {
	return e.Fire(EventSubagentStop, SubagentStopData{
		TaskDescription: taskDescription,
		ResponseLength:  responseLength,
		TokensUsed:      tokensUsed,
	})
}

// FirePreCompact is a convenience method for PreCompact event
func (e *Engine) FirePreCompact(compactionType string, currentSize, estimatedNewSize int) error {
	return e.Fire(EventPreCompact, PreCompactData{
		CompactionType:   compactionType,
		CurrentSize:      currentSize,
		EstimatedNewSize: estimatedNewSize,
	})
}
