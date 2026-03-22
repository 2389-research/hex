// ABOUTME: Tests for stuck detection tracking logic
// ABOUTME: Validates consecutive failure detection and reset behavior
package main

import (
	"strings"
	"testing"

	"github.com/2389-research/hex/internal/core"
)

func TestTurnTracker_NoFailure(t *testing.T) {
	tracker := &turnTracker{}
	results := []core.ContentBlock{
		{Type: "tool_result", ToolUseID: "1", Content: "file contents here"},
	}
	hint := tracker.recordTurnResults(results)
	if hint != "" {
		t.Errorf("expected no hint for successful result, got %q", hint)
	}
}

func TestTurnTracker_SingleFailure(t *testing.T) {
	tracker := &turnTracker{}
	results := []core.ContentBlock{
		{Type: "tool_result", ToolUseID: "1", Content: "Error: file not found"},
	}
	hint := tracker.recordTurnResults(results)
	if hint != "" {
		t.Errorf("expected no hint for first failure, got %q", hint)
	}
}

func TestTurnTracker_ConsecutiveFailures(t *testing.T) {
	tracker := &turnTracker{}
	results := []core.ContentBlock{
		{Type: "tool_result", ToolUseID: "1", Content: "Error: file not found"},
	}
	tracker.recordTurnResults(results)
	hint := tracker.recordTurnResults(results)
	if hint == "" {
		t.Error("expected stuck hint after 2 consecutive failures")
	}
}

func TestTurnTracker_DifferentToolResets(t *testing.T) {
	tracker := &turnTracker{}
	results1 := []core.ContentBlock{
		{Type: "tool_result", ToolUseID: "1", Content: "Error: file not found"},
	}
	tracker.recordTurnResults(results1)
	results2 := []core.ContentBlock{
		{Type: "tool_result", ToolUseID: "2", Content: "Error: command failed"},
	}
	hint := tracker.recordTurnResults(results2)
	if hint != "" {
		t.Errorf("expected no hint when different tool fails, got %q", hint)
	}
}

func TestTurnTracker_SuccessResets(t *testing.T) {
	tracker := &turnTracker{}
	failResults := []core.ContentBlock{
		{Type: "tool_result", ToolUseID: "1", Content: "Error: file not found"},
	}
	tracker.recordTurnResults(failResults)
	successResults := []core.ContentBlock{
		{Type: "tool_result", ToolUseID: "1", Content: "success"},
	}
	tracker.recordTurnResults(successResults)
	hint := tracker.recordTurnResults(failResults)
	if hint != "" {
		t.Errorf("expected no hint after success reset, got %q", hint)
	}
}

func TestTurnTracker_RecordByToolName(t *testing.T) {
	tracker := &turnTracker{}
	toolUses := []core.ToolUse{
		{ID: "tu_1", Name: "bash"},
	}
	results := []core.ContentBlock{
		{Type: "tool_result", ToolUseID: "tu_1", Content: "Error: command not found"},
	}
	tracker.recordByToolName(toolUses, results)
	hint := tracker.recordByToolName(toolUses, results)
	if hint == "" {
		t.Error("expected stuck hint with tool name after 2 consecutive failures")
	}
	if !strings.Contains(hint, "bash") {
		t.Errorf("expected hint to mention tool name 'bash', got %q", hint)
	}
}
