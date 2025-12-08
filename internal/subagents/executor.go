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
	"strings"
	"time"

	"github.com/2389-research/hex/internal/events"
	"github.com/google/uuid"
)

// Executor manages the execution of isolated subagent instances
type Executor struct {
	// HexBinPath is the path to the hex binary (empty = search PATH)
	HexBinPath string

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
		HexBinPath:     "",
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

	// Record AgentStarted event
	if store := events.Global(); store != nil {
		_ = store.Record(events.Event{
			ID:        uuid.New().String(),
			AgentID:   isolatedCtx.ID,
			ParentID:  parentID,
			Type:      events.EventAgentStarted,
			Timestamp: startTime,
			Data: map[string]interface{}{
				"type":        string(req.Type),
				"description": req.Description,
				"prompt":      req.Prompt,
			},
		})
	}

	// Execute the subagent
	output, execErr := e.executeSubprocess(ctx, req, config, isolatedCtx)

	// Record end time
	endTime := time.Now()

	// Record AgentStopped event
	if store := events.Global(); store != nil {
		stopData := map[string]interface{}{
			"type":        string(req.Type),
			"description": req.Description,
			"duration":    endTime.Sub(startTime).Seconds(),
			"success":     execErr == nil,
		}
		if execErr != nil {
			stopData["error"] = execErr.Error()
		}

		_ = store.Record(events.Event{
			ID:        uuid.New().String(),
			AgentID:   isolatedCtx.ID,
			ParentID:  parentID,
			Type:      events.EventAgentStopped,
			Timestamp: endTime,
			Data:      stopData,
		})
	}

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

// executeSubprocess spawns a hex subprocess for the subagent
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

	// Find hex binary
	hexbinPath := e.HexBinPath
	if hexbinPath == "" {
		var err error
		hexbinPath, err = exec.LookPath("hex")
		if err != nil {
			// Try building from current project
			hexbinPath, err = e.buildHex(cmdCtx)
			if err != nil {
				return "", fmt.Errorf("hex binary not found: %w", err)
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
	cmd := exec.CommandContext(cmdCtx, hexbinPath, args...) //nolint:gosec // G204: Args constructed from validated parameters

	// Inherit environment variables
	cmd.Env = os.Environ()

	// Ensure API key is available to subagent
	// Load config to get API key if not already in environment
	if os.Getenv("HEX_API_KEY") == "" {
		// Try to load from config file
		cfg, err := e.loadConfigForSubagent()
		if err == nil && cfg.APIKey != "" {
			cmd.Env = append(cmd.Env, fmt.Sprintf("HEX_API_KEY=%s", cfg.APIKey))
		}
	}

	// Generate hierarchical agent ID for this subagent
	parentAgentID := os.Getenv("HEX_AGENT_ID")
	if parentAgentID == "" {
		parentAgentID = "root"
	}
	agentID := events.GenerateAgentID(parentAgentID)

	// Add subagent-specific environment variables
	cmd.Env = append(cmd.Env,
		fmt.Sprintf("HEX_SUBAGENT_TYPE=%s", req.Type),
		fmt.Sprintf("HEX_SUBAGENT_CONTEXT_ID=%s", isolatedCtx.ID),
		fmt.Sprintf("HEX_AGENT_ID=%s", agentID),
		fmt.Sprintf("HEX_PARENT_AGENT_ID=%s", parentAgentID),
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

// buildHex attempts to build hex in a temporary location
func (e *Executor) buildHex(ctx context.Context) (string, error) {
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
	tempBin := filepath.Join(os.TempDir(), fmt.Sprintf("hex-subagent-%d", time.Now().Unix()))

	buildCmd := exec.CommandContext(ctx, "go", "build", "-o", tempBin, "./cmd/hex") //nolint:gosec // G204: Static args for build command
	buildCmd.Dir = projectRoot

	var buildOutput bytes.Buffer
	buildCmd.Stdout = &buildOutput
	buildCmd.Stderr = &buildOutput

	if err := buildCmd.Run(); err != nil {
		return "", fmt.Errorf("build hex: %w (output: %s)", err, buildOutput.String())
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

// loadConfigForSubagent loads the config to extract API key for subagents
// This avoids importing core package by using basic config loading
func (e *Executor) loadConfigForSubagent() (*subagentConfig, error) {
	// Try reading HEX_API_KEY from environment first
	if apiKey := os.Getenv("HEX_API_KEY"); apiKey != "" {
		return &subagentConfig{APIKey: apiKey}, nil
	}

	// Also check for ANTHROPIC_API_KEY (standard Anthropic env var)
	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		return &subagentConfig{APIKey: apiKey}, nil
	}

	// Try reading from ~/.hex/config.yaml
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(home, ".hex", "config.yaml")
	data, err := os.ReadFile(configPath) //nolint:gosec // G304: Path is constructed from UserHomeDir, not user input
	if err != nil {
		return nil, err
	}

	// Very simple YAML parsing for api_key
	// This avoids circular dependency on core package
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "api_key:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				apiKey := strings.TrimSpace(parts[1])
				apiKey = strings.Trim(apiKey, "\"'") // Remove quotes if present
				return &subagentConfig{APIKey: apiKey}, nil
			}
		}
	}

	return nil, fmt.Errorf("api_key not found in config")
}

// subagentConfig is a minimal config structure for subagent needs
type subagentConfig struct {
	APIKey string
}
