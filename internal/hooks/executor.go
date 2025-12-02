// ABOUTME: Hook execution engine for running shell commands
// ABOUTME: Handles timeouts, environment setup, and result collection

package hooks

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

// Executor runs hook commands
type Executor struct{}

// NewExecutor creates a new hook executor
func NewExecutor() *Executor {
	return &Executor{}
}

// HookResult contains the result of hook execution
type HookResult struct {
	Success  bool
	Output   string
	Error    string
	Duration time.Duration
}

// Run executes a hook command with timeout
func (e *Executor) Run(ctx context.Context, event HookEvent, config HookConfig, data EventData) (*HookResult, error) {
	timeout := time.Duration(config.Timeout) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second // Default 30s
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Prepare environment with event data
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("marshal event data: %w", err)
	}

	start := time.Now()

	// #nosec G204 -- Hook executor intentionally runs user-configured commands
	cmd := exec.CommandContext(ctx, "sh", "-c", config.Command)
	cmd.Env = append(cmd.Env,
		fmt.Sprintf("CLAUDE_HOOK_EVENT=%s", event),
		fmt.Sprintf("CLAUDE_HOOK_DATA=%s", jsonData),
	)

	output, err := cmd.CombinedOutput()
	duration := time.Since(start)

	result := &HookResult{
		Success:  err == nil,
		Output:   string(output),
		Duration: duration,
	}

	if err != nil {
		result.Error = err.Error()
		return result, err
	}

	return result, nil
}
