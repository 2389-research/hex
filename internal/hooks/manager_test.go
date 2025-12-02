package hooks

import (
	"context"
	"testing"
)

func TestManager_Trigger(t *testing.T) {
	config := HooksConfig{
		SessionStart: []HookConfig{
			{Command: "echo 'session started'", Timeout: 5},
		},
	}

	mgr := NewManager(config)

	results, err := mgr.Trigger(context.Background(), SessionStart, EventData{
		"timestamp": "2025-12-02T10:00:00Z",
	})

	if err != nil {
		t.Fatalf("Trigger() error = %v", err)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}

	if !results[0].Success {
		t.Errorf("expected success, got failure: %s", results[0].Error)
	}
}

func TestManager_Trigger_NoHooks(t *testing.T) {
	mgr := NewManager(HooksConfig{})

	results, err := mgr.Trigger(context.Background(), SessionStart, EventData{})

	if err != nil {
		t.Fatalf("Trigger() error = %v", err)
	}

	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestManager_TriggerAsync(t *testing.T) {
	t.Parallel()
	config := HooksConfig{
		SessionStart: []HookConfig{
			{Command: "echo 'async hook 1'", Timeout: 5},
			{Command: "echo 'async hook 2'", Timeout: 5},
		},
	}

	mgr := NewManager(config)

	// Trigger async hooks
	mgr.TriggerAsync(SessionStart, EventData{
		"timestamp": "2025-12-02T10:00:00Z",
	})

	// Wait should block until all hooks complete
	mgr.Wait()

	// If we get here, all hooks completed successfully
	// Test passes by not hanging or panicking
}

func TestManager_Wait_NoHooks(t *testing.T) {
	t.Parallel()
	mgr := NewManager(HooksConfig{})

	// Wait should return immediately when no hooks are running
	mgr.Wait()
	// Test passes by not hanging
}

func TestManager_TriggerAsync_MultipleEvents(t *testing.T) {
	t.Parallel()
	config := HooksConfig{
		SessionStart: []HookConfig{
			{Command: "echo 'start'", Timeout: 5},
		},
		SessionEnd: []HookConfig{
			{Command: "echo 'end'", Timeout: 5},
		},
	}

	mgr := NewManager(config)

	// Trigger multiple async events
	mgr.TriggerAsync(SessionStart, EventData{"timestamp": "1"})
	mgr.TriggerAsync(SessionEnd, EventData{"timestamp": "2"})

	// Wait should block until all hooks from both events complete
	mgr.Wait()
	// Test passes by completing without hanging
}
