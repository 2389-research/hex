// ABOUTME: Bash tool implementation for shell command execution
// ABOUTME: Executes shell commands with timeout, output capture, and safety features

package tools

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	// DefaultCommandTimeout is the default timeout for command execution
	DefaultCommandTimeout = 30 * time.Second

	// MaxCommandTimeout is the maximum allowed timeout (5 minutes)
	MaxCommandTimeout = 5 * time.Minute

	// MaxOutputSize is the maximum combined output size (1MB)
	MaxOutputSize = 1024 * 1024
)

// BashTool implements shell command execution functionality
type BashTool struct {
	DefaultTimeout time.Duration // Default timeout for commands
	MaxOutputSize  int           // Maximum output size (bytes)
}

// NewBashTool creates a new bash tool with default settings
func NewBashTool() *BashTool {
	return &BashTool{
		DefaultTimeout: DefaultCommandTimeout,
		MaxOutputSize:  MaxOutputSize,
	}
}

// Name returns the tool name
func (t *BashTool) Name() string {
	return "bash"
}

// Description returns the tool description
func (t *BashTool) Description() string {
	return "Executes a shell command in a subprocess. Parameters: command (required), timeout (optional, seconds), working_dir (optional), run_in_background (optional, boolean)"
}

// RequiresApproval always returns true for command execution
func (t *BashTool) RequiresApproval(params map[string]interface{}) bool {
	// ALWAYS require approval for command execution
	// Running arbitrary shell commands is extremely dangerous
	return true
}

// Execute runs a shell command and returns its output
func (t *BashTool) Execute(ctx context.Context, params map[string]interface{}) (*Result, error) {
	// Validate and extract command parameter
	command, ok := params["command"].(string)
	if !ok || command == "" {
		return &Result{
			ToolName: "bash",
			Success:  false,
			Error:    "missing or invalid 'command' parameter",
		}, nil
	}

	// Check if run_in_background is specified
	runInBackground := false
	if bgParam, exists := params["run_in_background"]; exists {
		bgBool, ok := bgParam.(bool)
		if !ok {
			return &Result{
				ToolName: "bash",
				Success:  false,
				Error:    "invalid 'run_in_background' parameter: must be a boolean",
			}, nil
		}
		runInBackground = bgBool
	}

	// If background execution is requested, launch background process
	if runInBackground {
		return t.executeBackground(ctx, command, params)
	}

	// Otherwise, execute synchronously (existing behavior)
	return t.executeSynchronous(ctx, command, params)
}

// executeBackground launches a command as a background process
func (t *BashTool) executeBackground(ctx context.Context, command string, params map[string]interface{}) (*Result, error) {
	// Generate unique ID for this background process
	bashID := uuid.New().String()

	// Get working directory (same as synchronous)
	workingDir := ""
	if dir, ok := params["working_dir"].(string); ok && dir != "" {
		// Clean and validate working directory
		cleanDir := filepath.Clean(dir)
		absDir, err := filepath.Abs(cleanDir)
		if err != nil {
			return &Result{
				ToolName: "bash",
				Success:  false,
				Error:    fmt.Sprintf("invalid working_dir: %v", err),
			}, nil
		}
		workingDir = absDir
	}

	// Create BackgroundProcess
	proc := &BackgroundProcess{
		ID:        bashID,
		Command:   command,
		StartTime: time.Now(),
		Stdout:    []string{},
		Stderr:    []string{},
		Done:      false,
		ExitCode:  0,
	}

	// Register the process
	err := GetBackgroundRegistry().Register(proc)
	if err != nil {
		return &Result{
			ToolName: "bash",
			Success:  false,
			Error:    fmt.Sprintf("failed to register background process: %v", err),
		}, nil
	}

	// Launch the command in a goroutine
	go func() {
		// Create command
		cmd := exec.Command("sh", "-c", command)

		// Set working directory if specified
		if workingDir != "" {
			cmd.Dir = workingDir
		}

		// Create pipes for stdout and stderr
		stdoutPipe, err := cmd.StdoutPipe()
		if err != nil {
			proc.MarkDone(-1)
			return
		}

		stderrPipe, err := cmd.StderrPipe()
		if err != nil {
			proc.MarkDone(-1)
			return
		}

		// Start the command
		if err := cmd.Start(); err != nil {
			proc.MarkDone(-1)
			return
		}

		// Store the OS process
		proc.mu.Lock()
		proc.Process = cmd.Process
		proc.mu.Unlock()

		// Use WaitGroup to synchronize output readers
		var wg sync.WaitGroup

		// Read stdout in a goroutine
		wg.Add(1)
		go func() {
			defer wg.Done()
			scanner := bufio.NewScanner(stdoutPipe)
			for scanner.Scan() {
				proc.AppendStdout(scanner.Text())
			}
		}()

		// Read stderr in a goroutine
		wg.Add(1)
		go func() {
			defer wg.Done()
			scanner := bufio.NewScanner(stderrPipe)
			for scanner.Scan() {
				proc.AppendStderr(scanner.Text())
			}
		}()

		// Wait for command to complete
		err = cmd.Wait()

		// Wait for all output to be read before marking done
		wg.Wait()

		exitCode := 0
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				exitCode = exitError.ExitCode()
			} else {
				exitCode = -1
			}
		}

		// Mark as done with exit code
		proc.MarkDone(exitCode)
	}()

	// Return immediately with bash_id
	return &Result{
		ToolName: "bash",
		Success:  true,
		Output:   fmt.Sprintf("Background process started with ID: %s", bashID),
		Metadata: map[string]interface{}{
			"bash_id": bashID,
			"command": command,
			"status":  "running",
		},
	}, nil
}

// executeSynchronous executes a command synchronously (original behavior)
func (t *BashTool) executeSynchronous(ctx context.Context, command string, params map[string]interface{}) (*Result, error) {
	// Get timeout (default: 30s)
	timeout := t.DefaultTimeout
	if timeoutParam, ok := params["timeout"].(float64); ok && timeoutParam > 0 {
		timeout = time.Duration(timeoutParam) * time.Second
	}

	// Cap timeout at 5 minutes for safety
	if timeout > MaxCommandTimeout {
		timeout = MaxCommandTimeout
	}

	// Get working directory (default: current directory)
	workingDir := ""
	if dir, ok := params["working_dir"].(string); ok && dir != "" {
		// Clean and validate working directory
		cleanDir := filepath.Clean(dir)
		absDir, err := filepath.Abs(cleanDir)
		if err != nil {
			return &Result{
				ToolName: "bash",
				Success:  false,
				Error:    fmt.Sprintf("invalid working_dir: %v", err),
			}, nil
		}
		workingDir = absDir
	}

	// Create command with timeout context
	cmdCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Execute command using sh -c (more portable than bash -c)
	cmd := exec.CommandContext(cmdCtx, "sh", "-c", command)

	// Set working directory if specified
	if workingDir != "" {
		cmd.Dir = workingDir
	}

	// Capture stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Record start time
	startTime := time.Now()

	// Execute command
	execErr := cmd.Run()
	duration := time.Since(startTime)

	// Check if timeout occurred first (before checking other errors)
	if cmdCtx.Err() == context.DeadlineExceeded {
		// Timeout
		return &Result{
			ToolName: "bash",
			Success:  false,
			Error:    fmt.Sprintf("command timed out after %v", timeout),
			Metadata: map[string]interface{}{
				"timeout":  timeout.Seconds(),
				"duration": duration.Seconds(),
			},
		}, nil
	}

	// Get exit code
	exitCode := 0
	if execErr != nil {
		if exitError, ok := execErr.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			// Other error (command not found, etc.)
			return &Result{
				ToolName: "bash",
				Success:  false,
				Error:    fmt.Sprintf("failed to execute command: %v", execErr),
			}, nil
		}
	}

	// Check output size
	stdoutStr := stdout.String()
	stderrStr := stderr.String()
	totalOutput := len(stdoutStr) + len(stderrStr)

	if totalOutput > t.MaxOutputSize {
		return &Result{
			ToolName: "bash",
			Success:  false,
			Error:    fmt.Sprintf("output too large: %d bytes (max: %d bytes)", totalOutput, t.MaxOutputSize),
			Metadata: map[string]interface{}{
				"stdout_size": len(stdoutStr),
				"stderr_size": len(stderrStr),
			},
		}, nil
	}

	// Build output string
	var output strings.Builder
	if stdoutStr != "" {
		output.WriteString("STDOUT:\n")
		output.WriteString(stdoutStr)
	}
	if stderrStr != "" {
		if output.Len() > 0 {
			output.WriteString("\n\n")
		}
		output.WriteString("STDERR:\n")
		output.WriteString(stderrStr)
	}
	if output.Len() == 0 {
		output.WriteString("(no output)")
	}

	// Determine success based on exit code
	success := exitCode == 0
	errorMsg := ""
	if !success {
		errorMsg = fmt.Sprintf("command exited with code %d", exitCode)
	}

	// Calculate line counts (count newlines + 1 for each stream that has content)
	stdoutLines := 0
	if stdoutStr != "" {
		stdoutLines = strings.Count(stdoutStr, "\n") + 1
		// Adjust for trailing newline
		if strings.HasSuffix(stdoutStr, "\n") {
			stdoutLines--
		}
		if stdoutLines < 1 {
			stdoutLines = 1
		}
	}

	stderrLines := 0
	if stderrStr != "" {
		stderrLines = strings.Count(stderrStr, "\n") + 1
		// Adjust for trailing newline
		if strings.HasSuffix(stderrStr, "\n") {
			stderrLines--
		}
		if stderrLines < 1 {
			stderrLines = 1
		}
	}

	// Return result
	return &Result{
		ToolName: "bash",
		Success:  success,
		Output:   output.String(),
		Error:    errorMsg,
		Metadata: map[string]interface{}{
			"exit_code":    exitCode,
			"duration":     duration.Seconds(),
			"command":      command,
			"working_dir":  workingDir,
			"stdout_lines": stdoutLines,
			"stderr_lines": stderrLines,
		},
	}, nil
}
