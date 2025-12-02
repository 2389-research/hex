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
