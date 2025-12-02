package hooks

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestExecutor_Run(t *testing.T) {
	exec := NewExecutor()

	ctx := context.Background()
	event := SessionStart
	data := EventData{"test": "value"}

	config := HookConfig{
		Command: "echo 'test'",
		Timeout: 5,
	}

	result, err := exec.Run(ctx, event, config, data)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if !result.Success {
		t.Errorf("expected success, got failure: %s", result.Error)
	}
}

func TestExecutor_Run_Timeout(t *testing.T) {
	exec := NewExecutor()

	ctx := context.Background()
	config := HookConfig{
		Command: "sleep 10",
		Timeout: 1,
	}

	start := time.Now()
	_, err := exec.Run(ctx, SessionStart, config, EventData{})
	duration := time.Since(start)

	if err == nil {
		t.Error("expected timeout error, got nil")
	}

	// Verify timeout occurred within reasonable time (should be ~1s, not 10s)
	if duration > 2*time.Second {
		t.Errorf("timeout took too long: %v", duration)
	}
}

func TestExecutor_Run_EnvironmentVariables(t *testing.T) {
	exec := NewExecutor()

	config := HookConfig{
		Command: "echo $CLAUDE_HOOK_EVENT $CLAUDE_HOOK_DATA",
		Timeout: 5,
	}

	event := SessionStart
	data := EventData{"key": "value"}

	result, err := exec.Run(context.Background(), event, config, data)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if !strings.Contains(result.Output, "SessionStart") {
		t.Errorf("output missing event name: %s", result.Output)
	}

	if !strings.Contains(result.Output, "key") {
		t.Errorf("output missing event data: %s", result.Output)
	}
}
