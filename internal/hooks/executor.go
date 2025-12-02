// ABOUTME: Shell command executor for hooks
// ABOUTME: Executes shell commands with timeout, environment variables, and output capture

package hooks

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"
)

// ExecutionResult contains the result of a hook execution
type ExecutionResult struct {
	Success  bool
	Stdout   string
	Stderr   string
	ExitCode int
	Error    error
}

// Executor executes shell commands for hooks
type Executor struct {
	projectPath string
	modelID     string
}

// NewExecutor creates a new hook executor
func NewExecutor(projectPath, modelID string) *Executor {
	return &Executor{
		projectPath: projectPath,
		modelID:     modelID,
	}
}

// Execute runs a hook command with the given event context
func (e *Executor) Execute(hook *HookConfig, event *Event) *ExecutionResult {
	// Create context with timeout
	timeout := time.Duration(hook.GetTimeout()) * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Build environment variables
	env := e.buildEnv(hook, event)

	// Create command
	cmd := exec.CommandContext(ctx, "sh", "-c", hook.Command) //nolint:gosec // G204 - hook commands from trusted config
	cmd.Env = env

	// Set working directory to project path if available and it exists
	if e.projectPath != "" {
		if _, err := os.Stat(e.projectPath); err == nil {
			cmd.Dir = e.projectPath
		}
		// If path doesn't exist, run in current directory
	}

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute command
	err := cmd.Run()

	result := &ExecutionResult{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}

	if err != nil {
		result.Success = false
		result.Error = err
		// Try to get exit code
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = -1
		}
		return result
	}

	result.Success = true
	result.ExitCode = 0
	return result
}

// buildEnv builds the environment variable array for command execution
func (e *Executor) buildEnv(hook *HookConfig, event *Event) []string {
	// Start with current environment
	env := os.Environ()

	// Add base event variables
	baseVars := map[string]string{
		"CLAUDE_EVENT":        string(event.Type),
		"CLAUDE_TIMESTAMP":    event.Timestamp.Format(time.RFC3339),
		"CLAUDE_PROJECT_PATH": e.projectPath,
		"CLAUDE_MODEL_ID":     e.modelID,
	}

	for k, v := range baseVars {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	// Add event-specific variables
	if event.Data != nil {
		for k, v := range event.Data.ToEnvVars() {
			env = append(env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	// Add hook-specific environment variables
	if hook.Env != nil {
		for k, v := range hook.Env {
			env = append(env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	return env
}

// ExecuteAsync runs a hook command asynchronously (fire and forget)
func (e *Executor) ExecuteAsync(hook *HookConfig, event *Event) {
	go func() {
		_ = e.Execute(hook, event)
		// Result is discarded for async execution
	}()
}
