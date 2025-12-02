package hooks

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestIntegration_SessionLifecycle(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "session.log")

	config := HooksConfig{
		SessionStart: []HookConfig{
			{Command: fmt.Sprintf("echo 'started' > %s", logFile), Timeout: 5},
		},
		SessionEnd: []HookConfig{
			{Command: fmt.Sprintf("echo 'ended' >> %s", logFile), Timeout: 5},
		},
	}

	mgr := NewManager(config)

	// Trigger session start
	_, err := mgr.Trigger(context.Background(), SessionStart, EventData{})
	if err != nil {
		t.Fatalf("SessionStart failed: %v", err)
	}

	// Trigger session end
	_, err = mgr.Trigger(context.Background(), SessionEnd, EventData{})
	if err != nil {
		t.Fatalf("SessionEnd failed: %v", err)
	}

	// Verify log file
	content, err := os.ReadFile(logFile) //nolint:gosec // Test file with controlled path
	if err != nil {
		t.Fatalf("failed to read log: %v", err)
	}

	expected := "started\nended\n"
	if string(content) != expected {
		t.Errorf("expected %q, got %q", expected, string(content))
	}
}
