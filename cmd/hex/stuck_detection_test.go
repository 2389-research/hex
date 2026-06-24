// ABOUTME: Tests for stuck detection tracking logic
// ABOUTME: Validates consecutive failure detection and reset behavior
package main

import (
	"bytes"
	"encoding/json"
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

func TestFormatOutputStreamJSONWritesSingleLine(t *testing.T) {
	resp := &core.MessageResponse{
		ID:         "msg_123",
		Type:       "message",
		Role:       "assistant",
		Model:      "claude-test",
		StopReason: "end_turn",
		Content: []core.Content{
			{Type: "text", Text: "hello"},
		},
		Usage: core.Usage{InputTokens: 3, OutputTokens: 2},
	}

	var out bytes.Buffer
	err := formatOutputTo(&out, resp, "stream-json")
	if err != nil {
		t.Fatalf("formatOutputTo stream-json returned error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected one JSON line, got %d lines: %q", len(lines), out.String())
	}

	var decoded core.MessageResponse
	if err := json.Unmarshal([]byte(lines[0]), &decoded); err != nil {
		t.Fatalf("stream-json output is not valid JSON: %v\n%s", err, lines[0])
	}
	if decoded.ID != resp.ID || decoded.Content[0].Text != "hello" {
		t.Fatalf("stream-json output did not preserve response: %#v", decoded)
	}
}

func TestEffectiveMaxTurnsUsesConfiguredFlag(t *testing.T) {
	originalMaxTurns := maxTurns
	defer func() { maxTurns = originalMaxTurns }()

	maxTurns = 7
	if got := effectiveMaxTurns(); got != 7 {
		t.Fatalf("effectiveMaxTurns() = %d, want 7", got)
	}

	maxTurns = 0
	if got := effectiveMaxTurns(); got != 50 {
		t.Fatalf("effectiveMaxTurns() with zero = %d, want default 50", got)
	}
}
