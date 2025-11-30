// ABOUTME: KillShell tool terminates background bash processes
// ABOUTME: Sends SIGTERM for graceful shutdown, then SIGKILL if needed

package tools

import (
	"context"
	"fmt"
	"syscall"
	"time"
)

// KillShellTool implements background process termination
type KillShellTool struct{}

// NewKillShellTool creates a new kill shell tool
func NewKillShellTool() *KillShellTool {
	return &KillShellTool{}
}

// Name returns the tool name
func (t *KillShellTool) Name() string {
	return "kill_shell"
}

// Description returns the tool description
func (t *KillShellTool) Description() string {
	return "Kill a running background bash shell by its ID. Parameters: shell_id (required)"
}

// RequiresApproval always returns true (killing processes is destructive)
func (t *KillShellTool) RequiresApproval(_ map[string]interface{}) bool {
	return true
}

// Execute terminates a background process
func (t *KillShellTool) Execute(_ context.Context, params map[string]interface{}) (*Result, error) {
	// Validate and extract shell_id parameter
	shellID, ok := params["shell_id"].(string)
	if !ok || shellID == "" {
		return &Result{
			ToolName: "kill_shell",
			Success:  false,
			Error:    "missing or invalid 'shell_id' parameter",
		}, nil
	}

	// Get the background process from registry
	bgProc, err := GetBackgroundRegistry().Get(shellID)
	if err != nil {
		return &Result{
			ToolName: "kill_shell",
			Success:  false,
			Error:    fmt.Sprintf("shell '%s' not found", shellID),
		}, nil
	}

	// Get the OS process
	process := bgProc.Process
	if process == nil {
		// Process struct exists but OS process is nil - clean up registry
		_ = GetBackgroundRegistry().Remove(shellID)
		return &Result{
			ToolName: "kill_shell",
			Success:  true,
			Output:   fmt.Sprintf("Shell %s process was already terminated. Cleaned up registry.", shellID),
			Metadata: map[string]interface{}{
				"shell_id": shellID,
			},
		}, nil
	}

	// Attempt graceful shutdown with SIGTERM
	killErr := process.Signal(syscall.SIGTERM)
	if killErr != nil {
		// Process might already be dead - try to remove from registry
		_ = GetBackgroundRegistry().Remove(shellID)
		return &Result{
			ToolName: "kill_shell",
			Success:  true,
			Output:   fmt.Sprintf("Shell %s was already terminated. Cleaned up registry.", shellID),
			Metadata: map[string]interface{}{
				"shell_id": shellID,
			},
		}, nil
	}

	// Wait briefly for graceful shutdown
	time.Sleep(100 * time.Millisecond)

	// Check if process is still alive
	checkErr := process.Signal(syscall.Signal(0)) // Signal 0 checks if process exists
	if checkErr == nil {
		// Still alive - force kill with SIGKILL
		_ = process.Kill()
		time.Sleep(50 * time.Millisecond)
	}

	// Remove from registry
	_ = GetBackgroundRegistry().Remove(shellID)

	return &Result{
		ToolName: "kill_shell",
		Success:  true,
		Output:   fmt.Sprintf("Successfully killed shell %s", shellID),
		Metadata: map[string]interface{}{
			"shell_id": shellID,
		},
	}, nil
}
