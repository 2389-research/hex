// ABOUTME: Task tool implementation for spawning sub-agent processes
// ABOUTME: Launches Clem subprocesses to handle complex multi-step tasks autonomously

package tools

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

const (
	// DefaultTaskTimeout is the default timeout for task execution
	DefaultTaskTimeout = 5 * time.Minute

	// MaxTaskTimeout is the maximum allowed timeout (30 minutes)
	MaxTaskTimeout = 30 * time.Minute
)

// TaskTool implements sub-agent task delegation functionality
type TaskTool struct {
	DefaultTimeout time.Duration // Default timeout for tasks
	ClembinPath    string        // Path to clem binary (empty = search PATH)
}

// NewTaskTool creates a new task tool with default settings
func NewTaskTool() *TaskTool {
	return &TaskTool{
		DefaultTimeout: DefaultTaskTimeout,
		ClembinPath:    "", // Will search PATH
	}
}

// Name returns the tool name
func (t *TaskTool) Name() string {
	return "task"
}

// Description returns the tool description
func (t *TaskTool) Description() string {
	return "Launch a sub-agent to handle complex, multi-step tasks autonomously. Parameters: prompt (required), description (required), subagent_type (required), model (optional), resume (optional)"
}

// RequiresApproval always returns true for task execution
func (t *TaskTool) RequiresApproval(params map[string]interface{}) bool {
	// ALWAYS require approval for task execution
	// - Spawns new processes (resource usage)
	// - Uses API (costs money)
	// - Can perform arbitrary actions through sub-agent
	return true
}

// Execute launches a sub-agent and returns its output
func (t *TaskTool) Execute(ctx context.Context, params map[string]interface{}) (*Result, error) {
	// Validate required parameters
	prompt, ok := params["prompt"].(string)
	if !ok || prompt == "" {
		return &Result{
			ToolName: "task",
			Success:  false,
			Error:    "missing or invalid 'prompt' parameter (must be non-empty string)",
		}, nil
	}

	description, ok := params["description"].(string)
	if !ok || description == "" {
		return &Result{
			ToolName: "task",
			Success:  false,
			Error:    "missing or invalid 'description' parameter (must be non-empty string)",
		}, nil
	}

	subagentType, ok := params["subagent_type"].(string)
	if !ok || subagentType == "" {
		return &Result{
			ToolName: "task",
			Success:  false,
			Error:    "missing or invalid 'subagent_type' parameter (must be non-empty string)",
		}, nil
	}

	// Optional parameters with type validation
	var modelName string
	if modelParam, exists := params["model"]; exists {
		var ok bool
		modelName, ok = modelParam.(string)
		if !ok {
			return &Result{
				ToolName: "task",
				Success:  false,
				Error:    "invalid 'model' parameter (must be string)",
			}, nil
		}
	}

	var resumeID string
	if resumeParam, exists := params["resume"]; exists {
		var ok bool
		resumeID, ok = resumeParam.(string)
		if !ok {
			return &Result{
				ToolName: "task",
				Success:  false,
				Error:    "invalid 'resume' parameter (must be string)",
			}, nil
		}
	}

	// Validate streaming parameter (optional bool)
	if streamingParam, exists := params["streaming"]; exists {
		if _, ok := streamingParam.(bool); !ok {
			return &Result{
				ToolName: "task",
				Success:  false,
				Error:    "invalid 'streaming' parameter (must be bool)",
			}, nil
		}
	}

	// Get timeout (default: 5 minutes)
	timeout := t.DefaultTimeout
	if timeoutParam, ok := params["timeout"].(float64); ok && timeoutParam > 0 {
		timeout = time.Duration(timeoutParam) * time.Second
	}

	// Cap timeout at 30 minutes for safety
	if timeout > MaxTaskTimeout {
		timeout = MaxTaskTimeout
	}

	// Create command with timeout context
	cmdCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Find clem binary
	clembinPath := t.ClembinPath
	if clembinPath == "" {
		var err error
		clembinPath, err = exec.LookPath("clem")
		if err != nil {
			// Try building from current project
			clembinPath, err = t.buildClem(cmdCtx)
			if err != nil {
				return &Result{
					ToolName: "task",
					Success:  false,
					Error:    fmt.Sprintf("clem binary not found: %v", err),
				}, nil
			}
		}
	}

	// Build command: clem --print "<prompt>"
	args := []string{"--print", prompt}

	// Add model flag if specified
	if modelName != "" {
		args = append([]string{"--model", modelName}, args...)
	}

	// Add resume flag if specified
	if resumeID != "" {
		args = append([]string{"--resume", resumeID}, args...)
	}

	//nolint:gosec // G204: Args constructed from validated parameters for subprocess tool execution
	cmd := exec.CommandContext(cmdCtx, clembinPath, args...)

	// Inherit environment variables (API key, config, etc.)
	cmd.Env = os.Environ()

	// Set working directory to current directory
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "" // Will use process's current directory
	}
	cmd.Dir = cwd

	// Capture combined output
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output

	// Record start time
	startTime := time.Now()

	// Execute command
	execErr := cmd.Run()
	duration := time.Since(startTime)

	// Check if timeout occurred first (before checking other errors)
	if cmdCtx.Err() == context.DeadlineExceeded {
		return &Result{
			ToolName: "task",
			Success:  false,
			Error:    fmt.Sprintf("task timed out after %v", timeout),
			Metadata: map[string]interface{}{
				"timeout":       timeout.Seconds(),
				"duration":      duration.Seconds(),
				"prompt":        prompt,
				"description":   description,
				"subagent_type": subagentType,
			},
		}, nil
	}

	// Check for context cancellation
	if cmdCtx.Err() == context.Canceled {
		return &Result{
			ToolName: "task",
			Success:  false,
			Error:    "task cancelled",
			Metadata: map[string]interface{}{
				"duration":      duration.Seconds(),
				"prompt":        prompt,
				"description":   description,
				"subagent_type": subagentType,
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
				ToolName: "task",
				Success:  false,
				Error:    fmt.Sprintf("failed to execute sub-agent: %v", execErr),
				Metadata: map[string]interface{}{
					"prompt":        prompt,
					"description":   description,
					"subagent_type": subagentType,
				},
			}, nil
		}
	}

	// Get output
	outputStr := output.String()
	if outputStr == "" {
		outputStr = "(no output)"
	}

	// Determine success based on exit code
	success := exitCode == 0
	errorMsg := ""
	if !success {
		errorMsg = fmt.Sprintf("sub-agent exited with code %d", exitCode)
	}

	// Build metadata
	metadata := map[string]interface{}{
		"exit_code":     exitCode,
		"duration":      duration.Seconds(),
		"prompt":        prompt,
		"description":   description,
		"subagent_type": subagentType,
	}

	if modelName != "" {
		metadata["model"] = modelName
	}

	if resumeID != "" {
		metadata["resume_id"] = resumeID
	}

	// Return result
	return &Result{
		ToolName: "task",
		Success:  success,
		Output:   outputStr,
		Error:    errorMsg,
		Metadata: metadata,
	}, nil
}

// buildClem attempts to build clem in a temporary location
func (t *TaskTool) buildClem(ctx context.Context) (string, error) {
	// Find go.mod to locate project root
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}

	// Search for go.mod up to 5 levels up
	projectRoot := ""
	searchDir := cwd
	for i := 0; i < 5; i++ {
		goModPath := filepath.Join(searchDir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			projectRoot = searchDir
			break
		}
		parent := filepath.Dir(searchDir)
		if parent == searchDir {
			break // Reached filesystem root
		}
		searchDir = parent
	}

	if projectRoot == "" {
		return "", fmt.Errorf("could not find project root (no go.mod)")
	}

	// Build to temporary location
	tempBin := filepath.Join(os.TempDir(), "clem-"+fmt.Sprintf("%d", time.Now().Unix()))

	//nolint:gosec // G204: Args constructed from validated parameters for subprocess tool execution
	buildCmd := exec.CommandContext(ctx, "go", "build", "-o", tempBin, "./cmd/clem")
	buildCmd.Dir = projectRoot

	var buildOutput bytes.Buffer
	buildCmd.Stdout = &buildOutput
	buildCmd.Stderr = &buildOutput

	if err := buildCmd.Run(); err != nil {
		return "", fmt.Errorf("build clem: %w (output: %s)", err, buildOutput.String())
	}

	return tempBin, nil
}

// ExecuteStreaming launches a sub-agent and streams incremental output updates.
// Returns a channel that emits Result objects as output becomes available.
// The channel is closed when execution completes or fails.
func (t *TaskTool) ExecuteStreaming(ctx context.Context, params map[string]interface{}) (<-chan *Result, error) {
	// Validate required parameters (same as Execute)
	prompt, ok := params["prompt"].(string)
	if !ok || prompt == "" {
		resultChan := make(chan *Result, 1)
		resultChan <- &Result{
			ToolName: "task",
			Success:  false,
			Error:    "missing or invalid 'prompt' parameter (must be non-empty string)",
		}
		close(resultChan)
		return resultChan, nil
	}

	description, ok := params["description"].(string)
	if !ok || description == "" {
		resultChan := make(chan *Result, 1)
		resultChan <- &Result{
			ToolName: "task",
			Success:  false,
			Error:    "missing or invalid 'description' parameter (must be non-empty string)",
		}
		close(resultChan)
		return resultChan, nil
	}

	subagentType, ok := params["subagent_type"].(string)
	if !ok || subagentType == "" {
		resultChan := make(chan *Result, 1)
		resultChan <- &Result{
			ToolName: "task",
			Success:  false,
			Error:    "missing or invalid 'subagent_type' parameter (must be non-empty string)",
		}
		close(resultChan)
		return resultChan, nil
	}

	// Optional parameters with type validation
	var modelName string
	if modelParam, exists := params["model"]; exists {
		var ok bool
		modelName, ok = modelParam.(string)
		if !ok {
			resultChan := make(chan *Result, 1)
			resultChan <- &Result{
				ToolName: "task",
				Success:  false,
				Error:    "invalid 'model' parameter (must be string)",
			}
			close(resultChan)
			return resultChan, nil
		}
	}

	var resumeID string
	if resumeParam, exists := params["resume"]; exists {
		var ok bool
		resumeID, ok = resumeParam.(string)
		if !ok {
			resultChan := make(chan *Result, 1)
			resultChan <- &Result{
				ToolName: "task",
				Success:  false,
				Error:    "invalid 'resume' parameter (must be string)",
			}
			close(resultChan)
			return resultChan, nil
		}
	}

	// Get timeout (default: 5 minutes)
	timeout := t.DefaultTimeout
	if timeoutParam, ok := params["timeout"].(float64); ok && timeoutParam > 0 {
		timeout = time.Duration(timeoutParam) * time.Second
	}

	// Cap timeout at 30 minutes for safety
	if timeout > MaxTaskTimeout {
		timeout = MaxTaskTimeout
	}

	// Create command with timeout context
	cmdCtx, cancel := context.WithTimeout(ctx, timeout)

	// Find clem binary
	clembinPath := t.ClembinPath
	if clembinPath == "" {
		var err error
		clembinPath, err = exec.LookPath("clem")
		if err != nil {
			// Try building from current project
			clembinPath, err = t.buildClem(cmdCtx)
			if err != nil {
				cancel()
				resultChan := make(chan *Result, 1)
				resultChan <- &Result{
					ToolName: "task",
					Success:  false,
					Error:    fmt.Sprintf("clem binary not found: %v", err),
				}
				close(resultChan)
				return resultChan, nil
			}
		}
	}

	// Build command: clem --print "<prompt>"
	args := []string{"--print", prompt}

	// Add model flag if specified
	if modelName != "" {
		args = append([]string{"--model", modelName}, args...)
	}

	// Add resume flag if specified
	if resumeID != "" {
		args = append([]string{"--resume", resumeID}, args...)
	}

	//nolint:gosec // G204: Args constructed from validated parameters for subprocess tool execution
	cmd := exec.CommandContext(cmdCtx, clembinPath, args...)

	// Inherit environment variables (API key, config, etc.)
	cmd.Env = os.Environ()

	// Set working directory to current directory
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "" // Will use process's current directory
	}
	cmd.Dir = cwd

	// Create pipes for stdout and stderr
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		resultChan := make(chan *Result, 1)
		resultChan <- &Result{
			ToolName: "task",
			Success:  false,
			Error:    fmt.Sprintf("failed to create stdout pipe: %v", err),
		}
		close(resultChan)
		return resultChan, nil
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		cancel()
		resultChan := make(chan *Result, 1)
		resultChan <- &Result{
			ToolName: "task",
			Success:  false,
			Error:    fmt.Sprintf("failed to create stderr pipe: %v", err),
		}
		close(resultChan)
		return resultChan, nil
	}

	// Start command
	startTime := time.Now()
	if err := cmd.Start(); err != nil {
		cancel()
		resultChan := make(chan *Result, 1)
		resultChan <- &Result{
			ToolName: "task",
			Success:  false,
			Error:    fmt.Sprintf("failed to start sub-agent: %v", err),
			Metadata: map[string]interface{}{
				"prompt":        prompt,
				"description":   description,
				"subagent_type": subagentType,
			},
		}
		close(resultChan)
		return resultChan, nil
	}

	// Create result channel (buffered to avoid blocking)
	resultChan := make(chan *Result, 10)

	// Launch goroutine to handle streaming
	go func() {
		defer close(resultChan)
		defer cancel()

		var outputBuf bytes.Buffer
		var bytesRead int64
		var wg sync.WaitGroup

		// Mutex to protect outputBuf from concurrent writes
		var mu sync.Mutex

		// Read from stdout
		wg.Add(1)
		go func() {
			defer wg.Done()
			scanner := bufio.NewScanner(stdoutPipe)
			for scanner.Scan() {
				line := scanner.Text() + "\n"
				mu.Lock()
				outputBuf.WriteString(line)
				bytesRead += int64(len(line))
				currentOutput := outputBuf.String()
				currentBytes := bytesRead
				mu.Unlock()

				// Send incremental update
				resultChan <- &Result{
					ToolName: "task",
					Success:  true,
					Output:   currentOutput,
					Metadata: map[string]interface{}{
						"bytes_read":    currentBytes,
						"streaming":     true,
						"prompt":        prompt,
						"description":   description,
						"subagent_type": subagentType,
					},
				}
			}
		}()

		// Read from stderr
		wg.Add(1)
		go func() {
			defer wg.Done()
			scanner := bufio.NewScanner(stderrPipe)
			for scanner.Scan() {
				line := scanner.Text() + "\n"
				mu.Lock()
				outputBuf.WriteString(line)
				bytesRead += int64(len(line))
				currentOutput := outputBuf.String()
				currentBytes := bytesRead
				mu.Unlock()

				// Send incremental update
				resultChan <- &Result{
					ToolName: "task",
					Success:  true,
					Output:   currentOutput,
					Metadata: map[string]interface{}{
						"bytes_read":    currentBytes,
						"streaming":     true,
						"prompt":        prompt,
						"description":   description,
						"subagent_type": subagentType,
					},
				}
			}
		}()

		// Wait for all output to be read
		wg.Wait()

		// Wait for command to finish
		execErr := cmd.Wait()
		duration := time.Since(startTime)

		// Get final output
		mu.Lock()
		finalOutput := outputBuf.String()
		finalBytes := bytesRead
		mu.Unlock()

		if finalOutput == "" {
			finalOutput = "(no output)"
		}

		// Check if timeout occurred
		if cmdCtx.Err() == context.DeadlineExceeded {
			resultChan <- &Result{
				ToolName: "task",
				Success:  false,
				Output:   finalOutput,
				Error:    fmt.Sprintf("task timed out after %v", timeout),
				Metadata: map[string]interface{}{
					"timeout":       timeout.Seconds(),
					"duration":      duration.Seconds(),
					"bytes_read":    finalBytes,
					"prompt":        prompt,
					"description":   description,
					"subagent_type": subagentType,
				},
			}
			return
		}

		// Check for context cancellation
		if cmdCtx.Err() == context.Canceled {
			resultChan <- &Result{
				ToolName: "task",
				Success:  false,
				Output:   finalOutput,
				Error:    "task cancelled",
				Metadata: map[string]interface{}{
					"duration":      duration.Seconds(),
					"bytes_read":    finalBytes,
					"prompt":        prompt,
					"description":   description,
					"subagent_type": subagentType,
				},
			}
			return
		}

		// Get exit code
		exitCode := 0
		if execErr != nil {
			if exitError, ok := execErr.(*exec.ExitError); ok {
				exitCode = exitError.ExitCode()
			} else {
				// Other error (command not found, etc.)
				resultChan <- &Result{
					ToolName: "task",
					Success:  false,
					Output:   finalOutput,
					Error:    fmt.Sprintf("failed to execute sub-agent: %v", execErr),
					Metadata: map[string]interface{}{
						"bytes_read":    finalBytes,
						"prompt":        prompt,
						"description":   description,
						"subagent_type": subagentType,
					},
				}
				return
			}
		}

		// Determine success based on exit code
		success := exitCode == 0
		errorMsg := ""
		if !success {
			errorMsg = fmt.Sprintf("sub-agent exited with code %d", exitCode)
		}

		// Build metadata
		metadata := map[string]interface{}{
			"exit_code":     exitCode,
			"duration":      duration.Seconds(),
			"bytes_read":    finalBytes,
			"prompt":        prompt,
			"description":   description,
			"subagent_type": subagentType,
		}

		if modelName != "" {
			metadata["model"] = modelName
		}

		if resumeID != "" {
			metadata["resume_id"] = resumeID
		}

		// Send final result
		resultChan <- &Result{
			ToolName: "task",
			Success:  success,
			Output:   finalOutput,
			Error:    errorMsg,
			Metadata: metadata,
		}
	}()

	return resultChan, nil
}
