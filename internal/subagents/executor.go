// ABOUTME: Isolated execution engine for subagents
// ABOUTME: Spawns and manages subagent processes with isolated contexts

package subagents

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// Executor manages the execution of isolated subagent instances
type Executor struct {
	// ClembinPath is the path to the clem binary (empty = search PATH)
	ClembinPath string

	// ContextManager manages isolated contexts
	ContextManager *ContextManager

	// DefaultTimeout is the default execution timeout
	DefaultTimeout time.Duration

	// MaxTimeout is the maximum allowed timeout
	MaxTimeout time.Duration
}

// NewExecutor creates a new subagent executor
func NewExecutor() *Executor {
	return &Executor{
		ClembinPath:    "",
		ContextManager: NewContextManager(),
		DefaultTimeout: 5 * time.Minute,
		MaxTimeout:     30 * time.Minute,
	}
}

// Execute runs a subagent with isolated context and returns the result
func (e *Executor) Execute(ctx context.Context, req *ExecutionRequest) (*Result, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid execution request: %w", err)
	}

	// Get or create configuration
	config := req.Config
	if config == nil {
		config = DefaultConfig(req.Type)
	}

	// Create isolated context
	parentID := "main" // In a real implementation, this would come from the parent agent
	isolatedCtx := e.ContextManager.CreateContext(parentID, req.Type)
	defer func() {
		// Clean up context after execution
		_ = e.ContextManager.DeleteContext(isolatedCtx.ID)
	}()

	// Record start time
	startTime := time.Now()

	// Execute the subagent
	output, execErr := e.executeSubprocess(ctx, req, config, isolatedCtx)

	// Record end time
	endTime := time.Now()

	// Build result
	result := &Result{
		Success:   execErr == nil,
		Output:    output,
		Type:      req.Type,
		StartTime: startTime,
		EndTime:   endTime,
		Metadata: map[string]interface{}{
			"type":        string(req.Type),
			"description": req.Description,
			"context_id":  isolatedCtx.ID,
			"duration":    endTime.Sub(startTime).Seconds(),
		},
	}

	if execErr != nil {
		result.Error = execErr.Error()
	}

	return result, nil
}

// executeSubprocess spawns a clem subprocess for the subagent
func (e *Executor) executeSubprocess(ctx context.Context, req *ExecutionRequest, config *Config, isolatedCtx *IsolatedContext) (string, error) {
	// Get timeout
	timeout := config.Timeout
	if timeout == 0 {
		timeout = e.DefaultTimeout
	}
	if timeout > e.MaxTimeout {
		timeout = e.MaxTimeout
	}

	// Create command context with timeout
	cmdCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Find clem binary
	clembinPath := e.ClembinPath
	if clembinPath == "" {
		var err error
		clembinPath, err = exec.LookPath("clem")
		if err != nil {
			// Try building from current project
			clembinPath, err = e.buildClem(cmdCtx)
			if err != nil {
				return "", fmt.Errorf("clem binary not found: %w", err)
			}
		}
	}

	// Build command arguments
	args := []string{"--print", req.Prompt}

	// Add model flag if specified
	if config.Model != "" {
		args = append([]string{"--model", config.Model}, args...)
	}

	// Add max tokens if specified
	if config.MaxTokens > 0 {
		args = append([]string{"--max-tokens", fmt.Sprintf("%d", config.MaxTokens)}, args...)
	}

	// Create command
	cmd := exec.CommandContext(cmdCtx, clembinPath, args...) //nolint:gosec // G204: Args constructed from validated parameters

	// Inherit environment variables
	cmd.Env = os.Environ()

	// Add subagent-specific environment variables
	cmd.Env = append(cmd.Env,
		fmt.Sprintf("CLEM_SUBAGENT_TYPE=%s", req.Type),
		fmt.Sprintf("CLEM_SUBAGENT_CONTEXT_ID=%s", isolatedCtx.ID),
	)

	// Set working directory
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "" // Will use process's current directory
	}
	cmd.Dir = cwd

	// Capture combined output
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output

	// Execute command
	execErr := cmd.Run()

	// Check for timeout
	if cmdCtx.Err() == context.DeadlineExceeded {
		return output.String(), fmt.Errorf("subagent timed out after %v", timeout)
	}

	// Check for cancellation
	if cmdCtx.Err() == context.Canceled {
		return output.String(), fmt.Errorf("subagent cancelled")
	}

	// Check for execution error
	if execErr != nil {
		if exitError, ok := execErr.(*exec.ExitError); ok {
			return output.String(), fmt.Errorf("subagent exited with code %d", exitError.ExitCode())
		}
		return output.String(), fmt.Errorf("failed to execute subagent: %w", execErr)
	}

	return output.String(), nil
}

// buildClem attempts to build clem in a temporary location
func (e *Executor) buildClem(ctx context.Context) (string, error) {
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
	tempBin := filepath.Join(os.TempDir(), fmt.Sprintf("clem-subagent-%d", time.Now().Unix()))

	buildCmd := exec.CommandContext(ctx, "go", "build", "-o", tempBin, "./cmd/clem") //nolint:gosec // G204: Static args for build command
	buildCmd.Dir = projectRoot

	var buildOutput bytes.Buffer
	buildCmd.Stdout = &buildOutput
	buildCmd.Stderr = &buildOutput

	if err := buildCmd.Run(); err != nil {
		return "", fmt.Errorf("build clem: %w (output: %s)", err, buildOutput.String())
	}

	return tempBin, nil
}

// HookEngine is an interface for firing hooks
// This allows us to avoid a circular dependency on the hooks package
type HookEngine interface {
	FireSubagentStop(taskDescription, subagentType string, responseLength, tokensUsed int, success bool, executionTime float64) error
}

// ExecuteWithHooks runs a subagent and fires hooks at appropriate times
// This method integrates with the hooks system
func (e *Executor) ExecuteWithHooks(ctx context.Context, req *ExecutionRequest, hookEngine HookEngine) (*Result, error) {
	// Execute the subagent
	result, err := e.Execute(ctx, req)

	// Fire SubagentStop hook if we have a hook engine
	if hookEngine != nil {
		// Calculate metrics for the hook
		responseLength := len(result.Output)
		tokensUsed := 0 // We don't have token info yet, would come from API response
		executionTime := result.Duration().Seconds()

		// Fire the hook (ignore errors - hooks shouldn't break execution)
		_ = hookEngine.FireSubagentStop(
			req.Description,
			string(req.Type),
			responseLength,
			tokensUsed,
			result.Success,
			executionTime,
		)

		result.Metadata["hook_fired"] = true
	}

	return result, err
}
