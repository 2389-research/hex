// ABOUTME: BashOutput tool for retrieving output from background bash processes
// ABOUTME: Supports incremental reading and regex filtering of process output

package tools

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

// BashOutputTool retrieves output from background bash processes
type BashOutputTool struct{}

// NewBashOutputTool creates a new bash output tool
func NewBashOutputTool() *BashOutputTool {
	return &BashOutputTool{}
}

// Name returns the tool name
func (t *BashOutputTool) Name() string {
	return "bash_output"
}

// Description returns the tool description
func (t *BashOutputTool) Description() string {
	return "Retrieve output from a background bash shell. Parameters: bash_id (required), filter (optional regex pattern)"
}

// RequiresApproval always returns false - this is a read-only tool
func (t *BashOutputTool) RequiresApproval(_ map[string]interface{}) bool {
	return false
}

// Execute retrieves output from a background process
func (t *BashOutputTool) Execute(_ context.Context, params map[string]interface{}) (*Result, error) {
	// Validate and extract bash_id parameter
	bashID, ok := params["bash_id"].(string)
	if !ok || bashID == "" {
		return &Result{
			ToolName: "bash_output",
			Success:  false,
			Error:    "missing or invalid 'bash_id' parameter",
		}, nil
	}

	// Get the background process
	proc, err := GetBackgroundRegistry().Get(bashID)
	if err != nil {
		return &Result{
			ToolName: "bash_output",
			Success:  false,
			Error:    fmt.Sprintf("background process '%s' not found", bashID),
		}, nil
	}

	// Get optional filter parameter
	var filterRegex *regexp.Regexp
	if filterParam, ok := params["filter"].(string); ok && filterParam != "" {
		var err error
		filterRegex, err = regexp.Compile(filterParam)
		if err != nil {
			return &Result{
				ToolName: "bash_output",
				Success:  false,
				Error:    fmt.Sprintf("invalid regex filter: %v", err),
			}, nil
		}
	}

	// Get new output since last read
	stdout, stderr := proc.GetNewOutput()

	// Apply filter if provided
	if filterRegex != nil {
		stdout = filterLines(stdout, filterRegex)
		stderr = filterLines(stderr, filterRegex)
	}

	// Build output string
	var output strings.Builder
	hasOutput := false

	if len(stdout) > 0 {
		output.WriteString("STDOUT:\n")
		output.WriteString(strings.Join(stdout, "\n"))
		hasOutput = true
	}

	if len(stderr) > 0 {
		if output.Len() > 0 {
			output.WriteString("\n\n")
		}
		output.WriteString("STDERR:\n")
		output.WriteString(strings.Join(stderr, "\n"))
		hasOutput = true
	}

	if !hasOutput {
		output.WriteString("(no new output)")
	}

	// Build metadata
	metadata := map[string]interface{}{
		"bash_id":      bashID,
		"command":      proc.Command,
		"done":         proc.IsDone(),
		"stdout_lines": len(stdout),
		"stderr_lines": len(stderr),
	}

	if proc.IsDone() {
		metadata["exit_code"] = proc.GetExitCode()
	}

	if filterRegex != nil {
		metadata["filter"] = filterRegex.String()
	}

	return &Result{
		ToolName: "bash_output",
		Success:  true,
		Output:   output.String(),
		Metadata: metadata,
	}, nil
}

// filterLines returns only lines matching the regex
func filterLines(lines []string, regex *regexp.Regexp) []string {
	filtered := make([]string, 0)
	for _, line := range lines {
		if regex.MatchString(line) {
			filtered = append(filtered, line)
		}
	}
	return filtered
}
