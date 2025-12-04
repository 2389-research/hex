package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/harper/jeff/internal/core"
	"github.com/harper/jeff/internal/hooks"
)

func TestHooksIntegration_ConfigLoading(t *testing.T) {
	// Create temporary config
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	logFile := filepath.Join(tmpDir, "session.log")

	// Note: YAML keys must match the exact case of HookEvent constants
	configContent := fmt.Sprintf(`hooks:
  SessionStart:
    - command: "echo 'started' > %s"
  SessionEnd:
    - command: "echo 'ended' >> %s"
`, logFile, logFile)

	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Set config path
	t.Setenv("PAGEN_CONFIG_PATH", configPath)

	// Load config
	cfg, err := core.LoadConfig()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Initialize hook manager
	hookManager := hooks.NewManager(cfg.Hooks)

	// Trigger SessionStart
	startResults, err := hookManager.Trigger(context.Background(), hooks.SessionStart, hooks.EventData{
		"timestamp": time.Now().Format(time.RFC3339),
		"model":     "test-model",
	})
	if err != nil {
		t.Fatalf("SessionStart failed: %v", err)
	}
	for i, r := range startResults {
		if !r.Success {
			t.Errorf("SessionStart result %d failed: %s", i, r.Error)
		}
	}

	// Trigger SessionEnd
	endResults, err := hookManager.Trigger(context.Background(), hooks.SessionEnd, hooks.EventData{
		"timestamp": time.Now().Format(time.RFC3339),
	})
	if err != nil {
		t.Fatalf("SessionEnd failed: %v", err)
	}
	for i, r := range endResults {
		if !r.Success {
			t.Errorf("SessionEnd result %d failed: %s", i, r.Error)
		}
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
