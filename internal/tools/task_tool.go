// ABOUTME: Task tool implementation for spawning sub-agent processes
// ABOUTME: Launches Hex subprocesses to handle complex multi-step tasks autonomously

package tools

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/2389-research/hex/internal/coordinator"
	"github.com/2389-research/hex/internal/registry"
	"github.com/2389-research/hex/internal/subagents"
)

const (
	// DefaultTaskTimeout is the default timeout for task execution
	DefaultTaskTimeout = 5 * time.Minute

	// MaxTaskTimeout is the maximum allowed timeout (30 minutes)
	MaxTaskTimeout = 30 * time.Minute

	// DefaultMaxAgentDepth is the default maximum recursion depth for sub-agents
	DefaultMaxAgentDepth = 5
)

// Global counter for generating unique sub-agent IDs
var subagentCounter atomic.Uint64

// getAgentDepth reads the current agent depth from environment variable
// Returns 0 if not set (root agent) or if value is invalid
func getAgentDepth() int {
	depthStr := os.Getenv("HEX_AGENT_DEPTH")
	if depthStr == "" {
		return 0 // Root agent
	}
	depth, err := strconv.Atoi(depthStr)
	if err != nil {
		return 0
	}
	return depth
}

// getMaxAgentDepth reads the maximum agent depth from environment variable
// Returns DefaultMaxAgentDepth if not set or if value is invalid
func getMaxAgentDepth() int {
	maxDepthStr := os.Getenv("HEX_MAX_AGENT_DEPTH")
	if maxDepthStr == "" {
		return DefaultMaxAgentDepth
	}
	maxDepth, err := strconv.Atoi(maxDepthStr)
	if err != nil || maxDepth < 1 {
		return DefaultMaxAgentDepth
	}
	return maxDepth
}

// MuxAgentRunner is an interface for running a mux agent.
// This interface breaks the import cycle between tools and adapter packages.
type MuxAgentRunner interface {
	// RunAgent executes a subagent with the given configuration and returns the output.
	RunAgent(ctx context.Context, agentID, prompt, systemPrompt string, allowedTools []string) (output string, err error)
}

// MuxConfig holds configuration for spawning mux agents
type MuxConfig struct {
	AgentRunner MuxAgentRunner // Runner that creates and executes mux agents
}

// TaskTool implements sub-agent task delegation functionality
type TaskTool struct {
	DefaultTimeout time.Duration       // Default timeout for tasks
	HexBinPath     string              // Path to hex binary (empty = search PATH)
	Executor       *subagents.Executor // Legacy subagent executor
	UseFramework   bool                // If true, use new subagent framework instead of direct subprocess
	UseMux         bool                // If true, use mux agents (Phase 2)
	MuxConfig      *MuxConfig          // Configuration for mux agents
}

// generateAgentID creates a hierarchical agent ID based on parent ID
// Format: root.1, root.2, root.1.1, root.1.2, etc.
func generateAgentID() string {
	parentID := os.Getenv("HEX_AGENT_ID")
	if parentID == "" {
		parentID = "root"
	}

	// Get next sequential ID for this parent
	nextID := subagentCounter.Add(1)
	return fmt.Sprintf("%s.%d", parentID, nextID)
}

// NewTaskTool creates a new task tool with default settings
func NewTaskTool() *TaskTool {
	return &TaskTool{
		DefaultTimeout: DefaultTaskTimeout,
		HexBinPath:     "", // Will search PATH
		Executor:       nil,
		UseFramework:   false, // Backward compatible: use old implementation by default
	}
}

// NewTaskToolWithFramework creates a task tool that uses the subagent framework
func NewTaskToolWithFramework() *TaskTool {
	executor := subagents.NewExecutor()
	return &TaskTool{
		DefaultTimeout: DefaultTaskTimeout,
		HexBinPath:     "",
		Executor:       executor,
		UseFramework:   true,
	}
}

// NewTaskToolWithMux creates a task tool that uses mux agents (in-process)
func NewTaskToolWithMux(runner MuxAgentRunner) *TaskTool {
	return &TaskTool{
		DefaultTimeout: DefaultTaskTimeout,
		HexBinPath:     "",
		UseMux:         true,
		MuxConfig:      &MuxConfig{AgentRunner: runner},
	}
}

// SetMuxRunner enables mux agent execution with the given runner.
// This allows setting mux config after construction (useful when tools have circular dependencies).
func (t *TaskTool) SetMuxRunner(runner MuxAgentRunner) {
	if runner != nil {
		t.MuxConfig = &MuxConfig{AgentRunner: runner}
		t.UseMux = true
	} else {
		t.MuxConfig = nil
		t.UseMux = false
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
func (t *TaskTool) RequiresApproval(_ map[string]interface{}) bool {
	// ALWAYS require approval for task execution
	// - Spawns new processes (resource usage)
	// - Uses API (costs money)
	// - Can perform arbitrary actions through sub-agent
	return true
}

// Execute launches a sub-agent and returns its output
func (t *TaskTool) Execute(ctx context.Context, params map[string]interface{}) (*Result, error) {
	// If using mux agents (Phase 2), delegate to mux executor
	if t.UseMux && t.MuxConfig != nil {
		return t.executeWithMux(ctx, params)
	}

	// If using the old subagent framework, delegate to it
	if t.UseFramework && t.Executor != nil {
		return t.executeWithFramework(ctx, params)
	}

	// Otherwise, use the legacy subprocess implementation
	return t.executeLegacy(ctx, params)
}

// executeWithFramework uses the new subagent framework
func (t *TaskTool) executeWithFramework(ctx context.Context, params map[string]interface{}) (*Result, error) {
	// Check recursion depth BEFORE doing anything else
	currentDepth := getAgentDepth()
	maxDepth := getMaxAgentDepth()

	if currentDepth >= maxDepth {
		return &Result{
			ToolName: "task",
			Success:  false,
			Error: fmt.Sprintf(
				"max agent depth (%d) exceeded - this usually means:\n"+
					"1. The task is too complex for recursive decomposition\n"+
					"2. The agent is stuck in a loop\n"+
					"3. You need to break down the task differently\n"+
					"Set HEX_MAX_AGENT_DEPTH to increase limit (use with caution)",
				maxDepth),
		}, nil
	}

	if os.Getenv("HEX_DEBUG") != "" {
		fmt.Fprintf(os.Stderr, "[TASK] Agent depth: %d/%d (framework)\n", currentDepth, maxDepth)
	}

	// Validate and extract parameters
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

	subagentTypeStr, ok := params["subagent_type"].(string)
	if !ok || subagentTypeStr == "" {
		return &Result{
			ToolName: "task",
			Success:  false,
			Error:    "missing or invalid 'subagent_type' parameter (must be non-empty string)",
		}, nil
	}

	// Validate subagent type
	if !subagents.IsValid(subagentTypeStr) {
		return &Result{
			ToolName: "task",
			Success:  false,
			Error:    fmt.Sprintf("invalid subagent_type '%s', must be one of: %v", subagentTypeStr, subagents.ValidSubagentTypes()),
		}, nil
	}

	// Create execution request
	req := &subagents.ExecutionRequest{
		Type:        subagents.SubagentType(subagentTypeStr),
		Prompt:      prompt,
		Description: description,
	}

	// Add optional model parameter
	if modelParam, exists := params["model"]; exists {
		if modelName, ok := modelParam.(string); ok {
			config := subagents.DefaultConfig(req.Type)
			config.Model = modelName
			req.Config = config
		}
	}

	// Execute subagent
	result, err := t.Executor.Execute(ctx, req)
	if err != nil {
		return &Result{
			ToolName: "task",
			Success:  false,
			Error:    err.Error(),
		}, nil
	}

	// Convert subagent result to tool result
	return &Result{
		ToolName: "task",
		Success:  result.Success,
		Output:   result.Output,
		Error:    result.Error,
		Metadata: result.Metadata,
	}, nil
}

// executeWithMux uses mux agents for in-process subagent execution (Phase 2)
func (t *TaskTool) executeWithMux(ctx context.Context, params map[string]interface{}) (*Result, error) {
	// Check recursion depth BEFORE doing anything else
	currentDepth := getAgentDepth()
	maxDepth := getMaxAgentDepth()

	if currentDepth >= maxDepth {
		return &Result{
			ToolName: "task",
			Success:  false,
			Error: fmt.Sprintf(
				"max agent depth (%d) exceeded - this usually means:\n"+
					"1. The task is too complex for recursive decomposition\n"+
					"2. The agent is stuck in a loop\n"+
					"3. You need to break down the task differently\n"+
					"Set HEX_MAX_AGENT_DEPTH to increase limit (use with caution)",
				maxDepth),
		}, nil
	}

	if os.Getenv("HEX_DEBUG") != "" {
		fmt.Fprintf(os.Stderr, "[TASK] Agent depth: %d/%d (mux)\n", currentDepth, maxDepth)
	}

	// Validate and extract parameters
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

	subagentTypeStr, ok := params["subagent_type"].(string)
	if !ok || subagentTypeStr == "" {
		return &Result{
			ToolName: "task",
			Success:  false,
			Error:    "missing or invalid 'subagent_type' parameter (must be non-empty string)",
		}, nil
	}

	// Validate subagent type
	if !subagents.IsValid(subagentTypeStr) {
		return &Result{
			ToolName: "task",
			Success:  false,
			Error:    fmt.Sprintf("invalid subagent_type '%s', must be one of: %v", subagentTypeStr, subagents.ValidSubagentTypes()),
		}, nil
	}

	// Generate agent ID
	agentID := generateAgentID()
	startTime := time.Now()

	// Ensure coordinator cleanup on exit
	defer coordinator.ReleaseAll(agentID)

	// Get configuration for this subagent type
	config := subagents.DefaultConfig(subagents.SubagentType(subagentTypeStr))

	// Set environment for depth tracking
	_ = os.Setenv("HEX_AGENT_DEPTH", strconv.Itoa(currentDepth+1))
	defer func() {
		_ = os.Setenv("HEX_AGENT_DEPTH", strconv.Itoa(currentDepth))
	}()

	// Run the agent via the runner interface
	output, err := t.MuxConfig.AgentRunner.RunAgent(ctx, agentID, prompt, config.SystemPrompt, config.AllowedTools)

	duration := time.Since(startTime)

	if output == "" {
		output = "(no output)"
	}

	errorMsg := ""
	success := err == nil
	if err != nil {
		errorMsg = err.Error()
	}

	return &Result{
		ToolName: "task",
		Success:  success,
		Output:   output,
		Error:    errorMsg,
		Metadata: map[string]interface{}{
			"duration":      duration.Seconds(),
			"agent_id":      agentID,
			"subagent_type": subagentTypeStr,
			"description":   description,
			"executor":      "mux",
		},
	}, nil
}

// executeLegacy uses the original subprocess implementation
func (t *TaskTool) executeLegacy(ctx context.Context, params map[string]interface{}) (*Result, error) {
	// Check recursion depth BEFORE doing anything else
	currentDepth := getAgentDepth()
	maxDepth := getMaxAgentDepth()

	if currentDepth >= maxDepth {
		return &Result{
			ToolName: "task",
			Success:  false,
			Error: fmt.Sprintf(
				"max agent depth (%d) exceeded - this usually means:\n"+
					"1. The task is too complex for recursive decomposition\n"+
					"2. The agent is stuck in a loop\n"+
					"3. You need to break down the task differently\n"+
					"Set HEX_MAX_AGENT_DEPTH to increase limit (use with caution)",
				maxDepth),
		}, nil
	}

	if os.Getenv("HEX_DEBUG") != "" {
		fmt.Fprintf(os.Stderr, "[TASK] Agent depth: %d/%d\n", currentDepth, maxDepth)
	}

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

	// Find hex binary
	hexbinPath := t.HexBinPath
	if hexbinPath == "" {
		var err error
		hexbinPath, err = exec.LookPath("hex")
		if err != nil {
			// Try building from current project
			hexbinPath, err = t.buildHex(cmdCtx)
			if err != nil {
				return &Result{
					ToolName: "task",
					Success:  false,
					Error:    fmt.Sprintf("hex binary not found: %v", err),
				}, nil
			}
		}
	}

	// Build command: hex --print "<prompt>"
	args := []string{"--print", prompt}

	// Add model flag if specified
	if modelName != "" {
		args = append([]string{"--model", modelName}, args...)
	}

	// Add resume flag if specified
	if resumeID != "" {
		args = append([]string{"--resume", resumeID}, args...)
	}

	cmd := exec.CommandContext(cmdCtx, hexbinPath, args...) //nolint:gosec // G204: Args constructed from validated parameters for subprocess tool execution

	// Generate hierarchical agent ID for this sub-agent
	agentID := generateAgentID()
	parentID := os.Getenv("HEX_AGENT_ID")
	if parentID == "" {
		parentID = "root"
	}

	// Inherit environment variables (API key, config, etc.)
	cmd.Env = os.Environ()

	// Set agent ID environment variables for process hierarchy tracking
	cmd.Env = append(cmd.Env, fmt.Sprintf("HEX_AGENT_ID=%s", agentID))
	cmd.Env = append(cmd.Env, fmt.Sprintf("HEX_PARENT_AGENT_ID=%s", parentID))

	// Pass depth+1 to child agent
	cmd.Env = append(cmd.Env, fmt.Sprintf("HEX_AGENT_DEPTH=%d", currentDepth+1))

	// Get default tools for subagent type and pass tool restrictions
	config := subagents.DefaultConfig(subagents.SubagentType(subagentType))
	if len(config.AllowedTools) > 0 {
		cmd.Env = append(cmd.Env,
			fmt.Sprintf("HEX_ALLOWED_TOOLS=%s", strings.Join(config.AllowedTools, ",")),
		)
	}
	// Note: hex's Config struct doesn't have DeniedTools yet, but we set the env var
	// for future compatibility when DeniedTools is added to Config
	cmd.Env = append(cmd.Env, "HEX_DENIED_TOOLS=")

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

	// Start the command so we can register the process
	if err := cmd.Start(); err != nil {
		return &Result{
			ToolName: "task",
			Success:  false,
			Error:    fmt.Sprintf("failed to start sub-agent: %v", err),
			Metadata: map[string]interface{}{
				"prompt":        prompt,
				"description":   description,
				"subagent_type": subagentType,
			},
		}, nil
	}

	// Register process immediately after start
	if err := registry.Global().Register(agentID, parentID, cmd.Process); err != nil {
		// Log error but continue - registration failure shouldn't block execution
		fmt.Fprintf(os.Stderr, "Warning: failed to register sub-agent %s: %v\n", agentID, err)
	}

	// Ensure cleanup on exit - deregister from process registry and release coordinator locks
	defer registry.Global().Deregister(agentID)
	defer coordinator.ReleaseAll(agentID)

	// Wait for command to complete
	execErr := cmd.Wait()
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

// buildHex attempts to build hex in a temporary location
func (t *TaskTool) buildHex(ctx context.Context) (string, error) {
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
	tempBin := filepath.Join(os.TempDir(), "hex-"+fmt.Sprintf("%d", time.Now().Unix()))

	buildCmd := exec.CommandContext(ctx, "go", "build", "-o", tempBin, "./cmd/hex") //nolint:gosec // G204: Args constructed from validated parameters for subprocess tool execution
	buildCmd.Dir = projectRoot

	var buildOutput bytes.Buffer
	buildCmd.Stdout = &buildOutput
	buildCmd.Stderr = &buildOutput

	if err := buildCmd.Run(); err != nil {
		return "", fmt.Errorf("build hex: %w (output: %s)", err, buildOutput.String())
	}

	return tempBin, nil
}

// ExecuteStreaming launches a sub-agent and streams incremental output updates.
// Returns a channel that emits Result objects as output becomes available.
// The channel is closed when execution completes or fails.
func (t *TaskTool) ExecuteStreaming(ctx context.Context, params map[string]interface{}) (<-chan *Result, error) {
	// Check recursion depth BEFORE doing anything else
	currentDepth := getAgentDepth()
	maxDepth := getMaxAgentDepth()

	if currentDepth >= maxDepth {
		resultChan := make(chan *Result, 1)
		resultChan <- &Result{
			ToolName: "task",
			Success:  false,
			Error: fmt.Sprintf(
				"max agent depth (%d) exceeded - this usually means:\n"+
					"1. The task is too complex for recursive decomposition\n"+
					"2. The agent is stuck in a loop\n"+
					"3. You need to break down the task differently\n"+
					"Set HEX_MAX_AGENT_DEPTH to increase limit (use with caution)",
				maxDepth),
		}
		close(resultChan)
		return resultChan, nil
	}

	if os.Getenv("HEX_DEBUG") != "" {
		fmt.Fprintf(os.Stderr, "[TASK] Agent depth: %d/%d (streaming)\n", currentDepth, maxDepth)
	}

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

	// Find hex binary
	hexbinPath := t.HexBinPath
	if hexbinPath == "" {
		var err error
		hexbinPath, err = exec.LookPath("hex")
		if err != nil {
			// Try building from current project
			hexbinPath, err = t.buildHex(cmdCtx)
			if err != nil {
				cancel()
				resultChan := make(chan *Result, 1)
				resultChan <- &Result{
					ToolName: "task",
					Success:  false,
					Error:    fmt.Sprintf("hex binary not found: %v", err),
				}
				close(resultChan)
				return resultChan, nil
			}
		}
	}

	// Build command: hex --print "<prompt>"
	args := []string{"--print", prompt}

	// Add model flag if specified
	if modelName != "" {
		args = append([]string{"--model", modelName}, args...)
	}

	// Add resume flag if specified
	if resumeID != "" {
		args = append([]string{"--resume", resumeID}, args...)
	}

	cmd := exec.CommandContext(cmdCtx, hexbinPath, args...) //nolint:gosec // G204: Args constructed from validated parameters for subprocess tool execution

	// Generate hierarchical agent ID for this sub-agent
	agentID := generateAgentID()
	parentID := os.Getenv("HEX_AGENT_ID")
	if parentID == "" {
		parentID = "root"
	}

	// Inherit environment variables (API key, config, etc.)
	cmd.Env = os.Environ()

	// Set agent ID environment variables for process hierarchy tracking
	cmd.Env = append(cmd.Env, fmt.Sprintf("HEX_AGENT_ID=%s", agentID))
	cmd.Env = append(cmd.Env, fmt.Sprintf("HEX_PARENT_AGENT_ID=%s", parentID))

	// Pass depth+1 to child agent
	cmd.Env = append(cmd.Env, fmt.Sprintf("HEX_AGENT_DEPTH=%d", currentDepth+1))

	// Get default tools for subagent type and pass tool restrictions
	streamConfig := subagents.DefaultConfig(subagents.SubagentType(subagentType))
	if len(streamConfig.AllowedTools) > 0 {
		cmd.Env = append(cmd.Env,
			fmt.Sprintf("HEX_ALLOWED_TOOLS=%s", strings.Join(streamConfig.AllowedTools, ",")),
		)
	}
	// Note: hex's Config struct doesn't have DeniedTools yet, but we set the env var
	// for future compatibility when DeniedTools is added to Config
	cmd.Env = append(cmd.Env, "HEX_DENIED_TOOLS=")

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

	// Register process with registry for cascading stop protocol
	if err := registry.Global().Register(agentID, parentID, cmd.Process); err != nil {
		// Log error but continue - registration failure shouldn't block execution
		fmt.Fprintf(os.Stderr, "Warning: failed to register sub-agent %s: %v\n", agentID, err)
	}

	// Create result channel (buffered to avoid blocking)
	resultChan := make(chan *Result, 10)

	// Launch goroutine to handle streaming
	go func() {
		defer close(resultChan)
		defer cancel()
		defer registry.Global().Deregister(agentID)
		defer coordinator.ReleaseAll(agentID)

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
