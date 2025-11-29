// ABOUTME: Tests for KillShell tool that terminates background bash processes
// ABOUTME: Covers metadata, approval requirements, process killing, and error handling

package tools

import (
	"context"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// TestKillShellTool_Name verifies the tool name
func TestKillShellTool_Name(t *testing.T) {
	tool := NewKillShellTool()
	if tool.Name() != "kill_shell" {
		t.Errorf("Expected name 'kill_shell', got %q", tool.Name())
	}
}

// TestKillShellTool_Description verifies the tool description
func TestKillShellTool_Description(t *testing.T) {
	tool := NewKillShellTool()
	desc := tool.Description()
	if desc == "" {
		t.Error("Description should not be empty")
	}
	if !strings.Contains(desc, "kill") && !strings.Contains(desc, "Kill") {
		t.Errorf("Description should mention killing: %q", desc)
	}
}

// TestKillShellTool_RequiresApproval verifies approval is always required
func TestKillShellTool_RequiresApproval(t *testing.T) {
	tool := NewKillShellTool()

	tests := []struct {
		name   string
		params map[string]interface{}
	}{
		{"with shell_id", map[string]interface{}{"shell_id": "test-123"}},
		{"empty params", map[string]interface{}{}},
		{"nil params", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tool.RequiresApproval(tt.params) {
				t.Error("KillShell should ALWAYS require approval (destructive operation)")
			}
		})
	}
}

// TestKillShellTool_Execute_MissingShellID tests missing shell_id parameter
func TestKillShellTool_Execute_MissingShellID(t *testing.T) {
	tool := NewKillShellTool()
	ctx := context.Background()

	tests := []struct {
		name   string
		params map[string]interface{}
	}{
		{"nil params", nil},
		{"empty params", map[string]interface{}{}},
		{"empty shell_id", map[string]interface{}{"shell_id": ""}},
		{"wrong type", map[string]interface{}{"shell_id": 123}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tool.Execute(ctx, tt.params)
			if err != nil {
				t.Fatalf("Execute should not return error, got %v", err)
			}
			if result.Success {
				t.Error("Result should not be successful with missing/invalid shell_id")
			}
			if !strings.Contains(result.Error, "shell_id") {
				t.Errorf("Error should mention shell_id, got: %q", result.Error)
			}
		})
	}
}

// TestKillShellTool_Execute_NonExistentShellID tests killing non-existent shell
func TestKillShellTool_Execute_NonExistentShellID(t *testing.T) {
	tool := NewKillShellTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"shell_id": "non-existent-shell-12345",
	}

	result, err := tool.Execute(ctx, params)
	if err != nil {
		t.Fatalf("Execute should not return error, got %v", err)
	}

	// Should gracefully handle non-existent shell
	if result.Success {
		t.Error("Result should not be successful for non-existent shell")
	}
	if !strings.Contains(result.Error, "not found") && !strings.Contains(result.Error, "does not exist") {
		t.Errorf("Error should indicate shell not found, got: %q", result.Error)
	}
}

// TestKillShellTool_Execute_KillRunningProcess tests killing an actual running process
func TestKillShellTool_Execute_KillRunningProcess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping process-spawning test in short mode")
	}

	tool := NewKillShellTool()
	ctx := context.Background()

	// Start a long-running background process
	shellID := "test-kill-running-" + time.Now().Format("20060102-150405")
	cmd := exec.Command("sh", "-c", "sleep 300")
	err := cmd.Start()
	if err != nil {
		t.Fatalf("Failed to start test process: %v", err)
	}
	defer cmd.Process.Kill() // Cleanup in case test fails

	// Register it in the background registry
	RegisterBackgroundProcess(shellID, cmd.Process)

	// Kill it via the tool
	params := map[string]interface{}{
		"shell_id": shellID,
	}

	result, err := tool.Execute(ctx, params)
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected success, got error: %q", result.Error)
	}

	// Verify process is actually dead
	time.Sleep(100 * time.Millisecond)  // Give it time to die
	checkErr := cmd.Process.Signal(nil) // Signal(nil) checks if process exists
	if checkErr == nil {
		t.Error("Process should be dead but is still running")
	}

	// Verify it's removed from registry
	if GetBackgroundProcess(shellID) != nil {
		t.Error("Process should be removed from registry after killing")
	}
}

// TestKillShellTool_Execute_KillAlreadyExitedProcess tests killing a process that already exited
func TestKillShellTool_Execute_KillAlreadyExitedProcess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping process-spawning test in short mode")
	}

	tool := NewKillShellTool()
	ctx := context.Background()

	// Start a process that exits immediately
	shellID := "test-kill-exited-" + time.Now().Format("20060102-150405")
	cmd := exec.Command("sh", "-c", "exit 0")
	err := cmd.Start()
	if err != nil {
		t.Fatalf("Failed to start test process: %v", err)
	}

	// Register it
	RegisterBackgroundProcess(shellID, cmd.Process)

	// Wait for it to exit
	cmd.Wait()
	time.Sleep(50 * time.Millisecond)

	// Try to kill it (should still clean up registry)
	params := map[string]interface{}{
		"shell_id": shellID,
	}

	result, err := tool.Execute(ctx, params)
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	// Should succeed (cleaning up is success even if process already dead)
	if !result.Success {
		t.Errorf("Expected success for cleaning up dead process, got: %q", result.Error)
	}

	// Verify it's removed from registry
	if GetBackgroundProcess(shellID) != nil {
		t.Error("Dead process should be removed from registry")
	}
}

// TestKillShellTool_Execute_Metadata verifies metadata in result
func TestKillShellTool_Execute_Metadata(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping process-spawning test in short mode")
	}

	tool := NewKillShellTool()
	ctx := context.Background()

	// Start a background process
	shellID := "test-metadata-" + time.Now().Format("20060102-150405")
	cmd := exec.Command("sh", "-c", "sleep 300")
	err := cmd.Start()
	if err != nil {
		t.Fatalf("Failed to start test process: %v", err)
	}
	defer cmd.Process.Kill() // Cleanup

	RegisterBackgroundProcess(shellID, cmd.Process)

	params := map[string]interface{}{
		"shell_id": shellID,
	}

	result, err := tool.Execute(ctx, params)
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	// Check metadata exists
	if result.Metadata == nil {
		t.Error("Metadata should not be nil")
	}

	// Check shell_id is in metadata
	if shellIDMeta, ok := result.Metadata["shell_id"].(string); !ok || shellIDMeta != shellID {
		t.Errorf("Metadata should contain shell_id=%q, got: %v", shellID, result.Metadata["shell_id"])
	}
}

// TestKillShellTool_Execute_OutputMessage verifies informative output message
func TestKillShellTool_Execute_OutputMessage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping process-spawning test in short mode")
	}

	tool := NewKillShellTool()
	ctx := context.Background()

	// Start a background process
	shellID := "test-output-" + time.Now().Format("20060102-150405")
	cmd := exec.Command("sh", "-c", "sleep 300")
	err := cmd.Start()
	if err != nil {
		t.Fatalf("Failed to start test process: %v", err)
	}
	defer cmd.Process.Kill() // Cleanup

	RegisterBackgroundProcess(shellID, cmd.Process)

	params := map[string]interface{}{
		"shell_id": shellID,
	}

	result, err := tool.Execute(ctx, params)
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	// Check output is informative
	if result.Output == "" {
		t.Error("Output should not be empty")
	}

	if !strings.Contains(result.Output, shellID) {
		t.Errorf("Output should mention shell_id, got: %q", result.Output)
	}
}

// TestKillShellTool_Execute_ContextCancellation tests context cancellation handling
func TestKillShellTool_Execute_ContextCancellation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping process-spawning test in short mode")
	}

	tool := NewKillShellTool()

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Start a background process
	shellID := "test-context-" + time.Now().Format("20060102-150405")
	cmd := exec.Command("sh", "-c", "sleep 300")
	err := cmd.Start()
	if err != nil {
		t.Fatalf("Failed to start test process: %v", err)
	}
	defer cmd.Process.Kill() // Cleanup

	RegisterBackgroundProcess(shellID, cmd.Process)

	params := map[string]interface{}{
		"shell_id": shellID,
	}

	result, err := tool.Execute(ctx, params)

	// Tool should still complete even with cancelled context
	// (killing is fast and important for cleanup)
	if err != nil {
		t.Fatalf("Execute should handle cancelled context gracefully, got error: %v", err)
	}

	if !result.Success {
		t.Logf("Result failed (acceptable): %q", result.Error)
	}

	// Clean up
	cmd.Process.Kill()
	UnregisterBackgroundProcess(shellID)
}
