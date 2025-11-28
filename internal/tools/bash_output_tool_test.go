// ABOUTME: Tests for BashOutput tool that retrieves output from background processes
// ABOUTME: Tests metadata, output retrieval, filtering, incremental reads, and error cases

package tools

import (
	"context"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestBashOutputTool_Name(t *testing.T) {
	tool := NewBashOutputTool()
	expected := "bash_output"
	if tool.Name() != expected {
		t.Errorf("Name() = %q, want %q", tool.Name(), expected)
	}
}

func TestBashOutputTool_Description(t *testing.T) {
	tool := NewBashOutputTool()
	desc := tool.Description()
	if desc == "" {
		t.Error("Description() returned empty string")
	}
	if !strings.Contains(desc, "bash_id") {
		t.Error("Description() should mention bash_id parameter")
	}
}

func TestBashOutputTool_RequiresApproval(t *testing.T) {
	tool := NewBashOutputTool()
	params := map[string]interface{}{
		"bash_id": "test-id",
	}

	// BashOutput should never require approval (read-only)
	if tool.RequiresApproval(params) {
		t.Error("RequiresApproval() should return false for read-only tool")
	}
}

func TestBashOutputTool_Execute_MissingBashID(t *testing.T) {
	tool := NewBashOutputTool()
	ctx := context.Background()

	// Test missing bash_id
	result, err := tool.Execute(ctx, map[string]interface{}{})
	if err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	if result.Success {
		t.Error("Execute() should fail with missing bash_id")
	}

	if !strings.Contains(result.Error, "bash_id") {
		t.Errorf("Error should mention bash_id, got: %s", result.Error)
	}
}

func TestBashOutputTool_Execute_InvalidBashID(t *testing.T) {
	tool := NewBashOutputTool()
	ctx := context.Background()

	// Test non-existent bash_id
	params := map[string]interface{}{
		"bash_id": "non-existent-id",
	}

	result, err := tool.Execute(ctx, params)
	if err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	if result.Success {
		t.Error("Execute() should fail with non-existent bash_id")
	}

	if !strings.Contains(result.Error, "not found") && !strings.Contains(result.Error, "does not exist") {
		t.Errorf("Error should indicate ID not found, got: %s", result.Error)
	}
}

func TestBashOutputTool_Execute_EmptyOutput(t *testing.T) {
	// Register a background process with no output yet
	bashID := "test-empty-123"
	bgProc := &BackgroundProcess{
		ID:         bashID,
		Command:    "sleep 10",
		StartTime:  time.Now(),
		Stdout:     []string{},
		Stderr:     []string{},
		ReadOffset: 0,
	}

	GetBackgroundRegistry().Register(bgProc)
	defer GetBackgroundRegistry().Remove(bashID)

	tool := NewBashOutputTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"bash_id": bashID,
	}

	result, err := tool.Execute(ctx, params)
	if err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	if !result.Success {
		t.Errorf("Execute() should succeed with empty output, got error: %s", result.Error)
	}

	if result.Output != "" && !strings.Contains(result.Output, "no new output") {
		t.Errorf("Output should indicate no output or be empty, got: %s", result.Output)
	}
}

func TestBashOutputTool_Execute_BasicOutput(t *testing.T) {
	// Register a background process with some output
	bashID := "test-basic-456"
	bgProc := &BackgroundProcess{
		ID:        bashID,
		Command:   "echo hello",
		StartTime: time.Now(),
		Stdout: []string{
			"hello",
			"world",
		},
		Stderr:     []string{},
		ReadOffset: 0,
		Done:       true,
		ExitCode:   0,
	}

	GetBackgroundRegistry().Register(bgProc)
	defer GetBackgroundRegistry().Remove(bashID)

	tool := NewBashOutputTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"bash_id": bashID,
	}

	result, err := tool.Execute(ctx, params)
	if err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	if !result.Success {
		t.Errorf("Execute() should succeed, got error: %s", result.Error)
	}

	if !strings.Contains(result.Output, "hello") {
		t.Errorf("Output should contain 'hello', got: %s", result.Output)
	}

	if !strings.Contains(result.Output, "world") {
		t.Errorf("Output should contain 'world', got: %s", result.Output)
	}
}

func TestBashOutputTool_Execute_IncrementalRead(t *testing.T) {
	// Register a background process
	bashID := "test-incremental-789"
	bgProc := &BackgroundProcess{
		ID:        bashID,
		Command:   "test command",
		StartTime: time.Now(),
		Stdout: []string{
			"line1",
			"line2",
			"line3",
		},
		Stderr:     []string{},
		ReadOffset: 0,
	}

	GetBackgroundRegistry().Register(bgProc)
	defer GetBackgroundRegistry().Remove(bashID)

	tool := NewBashOutputTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"bash_id": bashID,
	}

	// First read - should get all 3 lines
	result1, err := tool.Execute(ctx, params)
	if err != nil {
		t.Fatalf("First Execute() returned error: %v", err)
	}

	if !result1.Success {
		t.Errorf("First Execute() should succeed, got error: %s", result1.Error)
	}

	lineCount := strings.Count(result1.Output, "line")
	if lineCount != 3 {
		t.Errorf("First read should contain 3 lines, got %d", lineCount)
	}

	// Add more output
	proc, _ := GetBackgroundRegistry().Get(bashID)
	proc.Stdout = append(proc.Stdout, "line4", "line5")

	// Second read - should only get new lines (line4, line5)
	result2, err := tool.Execute(ctx, params)
	if err != nil {
		t.Fatalf("Second Execute() returned error: %v", err)
	}

	if !result2.Success {
		t.Errorf("Second Execute() should succeed, got error: %s", result2.Error)
	}

	if strings.Contains(result2.Output, "line1") || strings.Contains(result2.Output, "line2") {
		t.Error("Second read should not contain previously read lines")
	}

	if !strings.Contains(result2.Output, "line4") || !strings.Contains(result2.Output, "line5") {
		t.Error("Second read should contain new lines (line4, line5)")
	}
}

func TestBashOutputTool_Execute_WithStderr(t *testing.T) {
	// Register a background process with both stdout and stderr
	bashID := "test-stderr-321"
	bgProc := &BackgroundProcess{
		ID:        bashID,
		Command:   "test command",
		StartTime: time.Now(),
		Stdout: []string{
			"normal output",
		},
		Stderr: []string{
			"error output",
		},
		ReadOffset: 0,
	}

	GetBackgroundRegistry().Register(bgProc)
	defer GetBackgroundRegistry().Remove(bashID)

	tool := NewBashOutputTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"bash_id": bashID,
	}

	result, err := tool.Execute(ctx, params)
	if err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	if !result.Success {
		t.Errorf("Execute() should succeed, got error: %s", result.Error)
	}

	if !strings.Contains(result.Output, "normal output") {
		t.Error("Output should contain stdout")
	}

	if !strings.Contains(result.Output, "error output") {
		t.Error("Output should contain stderr")
	}
}

func TestBashOutputTool_Execute_WithFilter(t *testing.T) {
	// Register a background process with various output
	bashID := "test-filter-654"
	bgProc := &BackgroundProcess{
		ID:        bashID,
		Command:   "test command",
		StartTime: time.Now(),
		Stdout: []string{
			"ERROR: something went wrong",
			"INFO: processing item 1",
			"ERROR: another error",
			"INFO: processing item 2",
		},
		Stderr:     []string{},
		ReadOffset: 0,
	}

	GetBackgroundRegistry().Register(bgProc)
	defer GetBackgroundRegistry().Remove(bashID)

	tool := NewBashOutputTool()
	ctx := context.Background()

	// Test with regex filter to only show ERROR lines
	params := map[string]interface{}{
		"bash_id": bashID,
		"filter":  "ERROR:",
	}

	result, err := tool.Execute(ctx, params)
	if err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	if !result.Success {
		t.Errorf("Execute() should succeed, got error: %s", result.Error)
	}

	errorCount := strings.Count(result.Output, "ERROR:")
	if errorCount != 2 {
		t.Errorf("Filtered output should contain 2 ERROR lines, got %d", errorCount)
	}

	if strings.Contains(result.Output, "INFO:") {
		t.Error("Filtered output should not contain INFO lines")
	}
}

func TestBashOutputTool_Execute_InvalidRegexFilter(t *testing.T) {
	// Register a background process
	bashID := "test-invalid-regex-987"
	bgProc := &BackgroundProcess{
		ID:         bashID,
		Command:    "test command",
		StartTime:  time.Now(),
		Stdout:     []string{"test"},
		Stderr:     []string{},
		ReadOffset: 0,
	}

	GetBackgroundRegistry().Register(bgProc)
	defer GetBackgroundRegistry().Remove(bashID)

	tool := NewBashOutputTool()
	ctx := context.Background()

	// Test with invalid regex
	params := map[string]interface{}{
		"bash_id": bashID,
		"filter":  "[invalid(regex",
	}

	result, err := tool.Execute(ctx, params)
	if err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	if result.Success {
		t.Error("Execute() should fail with invalid regex")
	}

	if !strings.Contains(result.Error, "regex") && !strings.Contains(result.Error, "filter") {
		t.Errorf("Error should mention regex/filter issue, got: %s", result.Error)
	}
}

func TestBashOutputTool_Execute_ProcessMetadata(t *testing.T) {
	// Register a completed background process
	bashID := "test-metadata-111"
	startTime := time.Now().Add(-5 * time.Second)
	bgProc := &BackgroundProcess{
		ID:        bashID,
		Command:   "echo done",
		StartTime: startTime,
		Stdout: []string{
			"done",
		},
		Stderr:     []string{},
		ReadOffset: 0,
		Done:       true,
		ExitCode:   0,
	}

	GetBackgroundRegistry().Register(bgProc)
	defer GetBackgroundRegistry().Remove(bashID)

	tool := NewBashOutputTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"bash_id": bashID,
	}

	result, err := tool.Execute(ctx, params)
	if err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	if !result.Success {
		t.Errorf("Execute() should succeed, got error: %s", result.Error)
	}

	// Check metadata
	if result.Metadata == nil {
		t.Fatal("Metadata should not be nil")
	}

	if _, ok := result.Metadata["done"]; !ok {
		t.Error("Metadata should include 'done' field")
	}

	if _, ok := result.Metadata["exit_code"]; !ok {
		t.Error("Metadata should include 'exit_code' field")
	}
}

// Test helper to compile regex - used in implementation
func TestRegexCompilation(t *testing.T) {
	validPatterns := []string{
		"ERROR",
		"^INFO:",
		".*warning.*",
		"[0-9]+",
	}

	for _, pattern := range validPatterns {
		_, err := regexp.Compile(pattern)
		if err != nil {
			t.Errorf("Valid pattern %q failed to compile: %v", pattern, err)
		}
	}

	invalidPatterns := []string{
		"[invalid",
		"(?P<unclosed",
		"*invalid",
	}

	for _, pattern := range invalidPatterns {
		_, err := regexp.Compile(pattern)
		if err == nil {
			t.Errorf("Invalid pattern %q should have failed to compile", pattern)
		}
	}
}
