// ABOUTME: Tracks consecutive tool failures to detect when the agent is stuck
// ABOUTME: Injects hints into the conversation when repeated failures are detected
package main

import (
	"fmt"
	"strings"

	"github.com/2389-research/hex/internal/core"
)

// turnTracker monitors tool execution patterns to detect stuck loops
type turnTracker struct {
	lastFailedTool string
	failCount      int
}

// recordTurnResults analyzes tool results from a turn and updates tracking state.
// Returns a hint string if the agent appears stuck, empty string otherwise.
func (t *turnTracker) recordTurnResults(toolResults []core.ContentBlock) string {
	var failedTool string
	hasFailure := false

	for _, result := range toolResults {
		if result.Type == "tool_result" && strings.Contains(result.Content, "Error:") {
			hasFailure = true
			failedTool = result.ToolUseID
		}
	}

	if !hasFailure {
		t.lastFailedTool = ""
		t.failCount = 0
		return ""
	}

	if failedTool == t.lastFailedTool {
		t.failCount++
	} else {
		t.lastFailedTool = failedTool
		t.failCount = 1
	}

	if t.failCount >= 2 {
		return fmt.Sprintf("[hex: The same operation has failed %d times consecutively. Consider trying a different approach rather than retrying the same action.]", t.failCount)
	}

	return ""
}

// recordByToolName takes tool uses and results, tracking by tool name for better detection.
func (t *turnTracker) recordByToolName(toolUses []core.ToolUse, toolResults []core.ContentBlock) string {
	idToName := make(map[string]string)
	for _, tu := range toolUses {
		idToName[tu.ID] = tu.Name
	}

	var failedToolName string
	hasFailure := false

	for _, result := range toolResults {
		if result.Type == "tool_result" && strings.Contains(result.Content, "Error:") {
			hasFailure = true
			if name, ok := idToName[result.ToolUseID]; ok {
				failedToolName = name
			}
		}
	}

	if !hasFailure {
		t.lastFailedTool = ""
		t.failCount = 0
		return ""
	}

	if failedToolName == t.lastFailedTool && failedToolName != "" {
		t.failCount++
	} else {
		t.lastFailedTool = failedToolName
		t.failCount = 1
	}

	if t.failCount >= 2 {
		return fmt.Sprintf("[hex: Tool '%s' has failed %d times consecutively. Try a different approach — read the error carefully, check your assumptions, or try an alternative strategy.]", failedToolName, t.failCount)
	}

	return ""
}
