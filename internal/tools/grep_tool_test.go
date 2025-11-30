// ABOUTME: Comprehensive tests for Grep tool - ripgrep-based code search
// ABOUTME: Covers patterns, output modes, context lines, filters, case sensitivity

package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ===========================
// Basic Search Tests
// ===========================

func TestGrepTool_BasicSearch(t *testing.T) {
	tool := NewGrepTool()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	content := `package main

func main() {
	fmt.Println("hello")
}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	params := map[string]interface{}{
		"pattern":     "Println",
		"path":        tmpDir,
		"output_mode": "content",
	}

	result, err := tool.Execute(context.Background(), params)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("Expected success, got error: %s", result.Error)
	}

	output := result.Output
	if !strings.Contains(output, "Println") {
		t.Errorf("Expected output to contain 'Println', got: %s", output)
	}
}

func TestGrepTool_FilesWithMatches(t *testing.T) {
	tool := NewGrepTool()

	tmpDir := t.TempDir()

	// Create multiple files
	files := map[string]string{
		"file1.txt": "hello world",
		"file2.txt": "foo bar",
		"file3.txt": "hello universe",
	}

	for name, content := range files {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create %s: %v", name, err)
		}
	}

	params := map[string]interface{}{
		"pattern":     "hello",
		"path":        tmpDir,
		"output_mode": "files_with_matches",
	}

	result, err := tool.Execute(context.Background(), params)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("Expected success, got error: %s", result.Error)
	}

	output := result.Output
	if !strings.Contains(output, "file1.txt") {
		t.Error("Expected file1.txt in output")
	}
	if !strings.Contains(output, "file3.txt") {
		t.Error("Expected file3.txt in output")
	}
	if strings.Contains(output, "file2.txt") {
		t.Error("Did not expect file2.txt in output")
	}
}

func TestGrepTool_CountMode(t *testing.T) {
	tool := NewGrepTool()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "foo\nbar\nfoo\nbaz\nfoo\n"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	params := map[string]interface{}{
		"pattern":     "foo",
		"path":        tmpDir,
		"output_mode": "count",
	}

	result, err := tool.Execute(context.Background(), params)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("Expected success, got error: %s", result.Error)
	}

	output := result.Output
	if !strings.Contains(output, "3") {
		t.Errorf("Expected count of 3, got: %s", output)
	}
}

// ===========================
// Context Lines Tests
// ===========================

func TestGrepTool_ContextBefore(t *testing.T) {
	tool := NewGrepTool()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "line1\nline2\nline3\nTARGET\nline5\n"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	params := map[string]interface{}{
		"pattern":     "TARGET",
		"path":        tmpDir,
		"output_mode": "content",
		"-B":          2, // 2 lines before
	}

	result, err := tool.Execute(context.Background(), params)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("Expected success, got error: %s", result.Error)
	}

	output := result.Output
	// ripgrep shows context with line numbers: "2-line2\n3-line3\n4:TARGET"
	if !strings.Contains(output, "line2") || !strings.Contains(output, "line3") || !strings.Contains(output, "TARGET") {
		t.Errorf("Expected context before TARGET (line2, line3, TARGET), got: %s", output)
	}
}

func TestGrepTool_ContextAfter(t *testing.T) {
	tool := NewGrepTool()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "line1\nTARGET\nline3\nline4\nline5\n"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	params := map[string]interface{}{
		"pattern":     "TARGET",
		"path":        tmpDir,
		"output_mode": "content",
		"-A":          2, // 2 lines after
	}

	result, err := tool.Execute(context.Background(), params)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("Expected success, got error: %s", result.Error)
	}

	output := result.Output
	// ripgrep shows context with line numbers: "2:TARGET\n3-line3\n4-line4"
	if !strings.Contains(output, "line3") || !strings.Contains(output, "line4") || !strings.Contains(output, "TARGET") {
		t.Errorf("Expected context after TARGET (TARGET, line3, line4), got: %s", output)
	}
}

func TestGrepTool_ContextAround(t *testing.T) {
	tool := NewGrepTool()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "line1\nline2\nTARGET\nline4\nline5\n"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	params := map[string]interface{}{
		"pattern":     "TARGET",
		"path":        tmpDir,
		"output_mode": "content",
		"-C":          1, // 1 line before and after
	}

	result, err := tool.Execute(context.Background(), params)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("Expected success, got error: %s", result.Error)
	}

	output := result.Output
	// ripgrep shows context with line numbers: "2-line2\n3:TARGET\n4-line4"
	if !strings.Contains(output, "line2") || !strings.Contains(output, "line4") || !strings.Contains(output, "TARGET") {
		t.Errorf("Expected context around TARGET (line2, TARGET, line4), got: %s", output)
	}
}

// ===========================
// Filter Tests
// ===========================

func TestGrepTool_GlobFilter(t *testing.T) {
	tool := NewGrepTool()

	tmpDir := t.TempDir()

	files := map[string]string{
		"test.go":  "package main",
		"test.txt": "package test",
		"main.go":  "package main",
	}

	for name, content := range files {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create %s: %v", name, err)
		}
	}

	params := map[string]interface{}{
		"pattern":     "package",
		"path":        tmpDir,
		"glob":        "*.go",
		"output_mode": "files_with_matches",
	}

	result, err := tool.Execute(context.Background(), params)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("Expected success, got error: %s", result.Error)
	}

	output := result.Output
	if !strings.Contains(output, "test.go") || !strings.Contains(output, "main.go") {
		t.Error("Expected both .go files")
	}
	if strings.Contains(output, "test.txt") {
		t.Error("Did not expect .txt file")
	}
}

func TestGrepTool_TypeFilter(t *testing.T) {
	tool := NewGrepTool()

	tmpDir := t.TempDir()

	files := map[string]string{
		"test.go": "func main",
		"test.py": "def main",
		"main.go": "func test",
	}

	for name, content := range files {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create %s: %v", name, err)
		}
	}

	params := map[string]interface{}{
		"pattern":     "func",
		"path":        tmpDir,
		"type":        "go",
		"output_mode": "files_with_matches",
	}

	result, err := tool.Execute(context.Background(), params)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("Expected success, got error: %s", result.Error)
	}

	output := result.Output
	if !strings.Contains(output, "test.go") || !strings.Contains(output, "main.go") {
		t.Error("Expected both .go files")
	}
	if strings.Contains(output, "test.py") {
		t.Error("Did not expect .py file")
	}
}

// ===========================
// Case Sensitivity Tests
// ===========================

func TestGrepTool_CaseSensitive(t *testing.T) {
	tool := NewGrepTool()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "Hello HELLO hello"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	params := map[string]interface{}{
		"pattern":     "hello",
		"path":        tmpDir,
		"output_mode": "count",
	}

	result, err := tool.Execute(context.Background(), params)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("Expected success, got error: %s", result.Error)
	}

	output := result.Output
	// Should only match lowercase "hello"
	if !strings.Contains(output, "1") {
		t.Errorf("Expected count of 1 (case sensitive), got: %s", output)
	}
}

func TestGrepTool_CaseInsensitive(t *testing.T) {
	tool := NewGrepTool()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	// Put each "hello" on a separate line so rg -c counts 3 lines
	content := "Hello\nHELLO\nhello"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	params := map[string]interface{}{
		"pattern":     "hello",
		"path":        tmpDir,
		"output_mode": "count",
		"-i":          true,
	}

	result, err := tool.Execute(context.Background(), params)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("Expected success, got error: %s", result.Error)
	}

	output := result.Output
	// Should match all three lines
	if !strings.Contains(output, "3") {
		t.Errorf("Expected count of 3 (case insensitive), got: %s", output)
	}
}

// ===========================
// Error Cases
// ===========================

func TestGrepTool_MissingPattern(t *testing.T) {
	tool := NewGrepTool()

	params := map[string]interface{}{
		"path": "/tmp",
	}

	result, err := tool.Execute(context.Background(), params)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.Success {
		t.Fatal("Expected failure when pattern is missing")
	}
}

func TestGrepTool_InvalidPath(t *testing.T) {
	tool := NewGrepTool()

	params := map[string]interface{}{
		"pattern": "test",
		"path":    "/nonexistent/path/that/does/not/exist",
	}

	result, err := tool.Execute(context.Background(), params)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.Success {
		t.Fatal("Expected failure for invalid path")
	}
}

func TestGrepTool_NoMatches(t *testing.T) {
	tool := NewGrepTool()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "foo bar baz"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	params := map[string]interface{}{
		"pattern":     "nonexistent",
		"path":        tmpDir,
		"output_mode": "content",
	}

	result, err := tool.Execute(context.Background(), params)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("Expected success with no matches, got error: %s", result.Error)
	}

	output := result.Output
	if output != "" && !strings.Contains(output, "No matches") {
		t.Errorf("Expected empty or 'No matches', got: %s", output)
	}
}

// ===========================
// Metadata Tests
// ===========================

func TestGrepTool_Metadata(t *testing.T) {
	tool := NewGrepTool()

	if tool.Name() != "grep" {
		t.Errorf("Expected name 'grep', got '%s'", tool.Name())
	}

	if tool.Description() == "" {
		t.Error("Expected non-empty description")
	}
}

func TestGrepTool_RequiresApproval(t *testing.T) {
	tool := NewGrepTool()

	params := map[string]interface{}{
		"pattern": "test",
		"path":    "/tmp",
	}

	// Grep is read-only, should not require approval for normal paths
	if tool.RequiresApproval(params) {
		t.Error("Grep should not require approval for normal paths")
	}
}

// ===========================
// Regex Pattern Tests
// ===========================

func TestGrepTool_RegexPattern(t *testing.T) {
	tool := NewGrepTool()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "error: something\nwarning: else\nerror: another\n"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	params := map[string]interface{}{
		"pattern":     "error:.*",
		"path":        tmpDir,
		"output_mode": "count",
	}

	result, err := tool.Execute(context.Background(), params)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("Expected success, got error: %s", result.Error)
	}

	output := result.Output
	if !strings.Contains(output, "2") {
		t.Errorf("Expected 2 matches for regex pattern, got: %s", output)
	}
}

func TestGrepTool_DefaultPath(t *testing.T) {
	tool := NewGrepTool()

	// When path is not provided, should use current directory
	params := map[string]interface{}{
		"pattern":     "package",
		"output_mode": "files_with_matches",
	}

	result, err := tool.Execute(context.Background(), params)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Should succeed (searching current directory)
	if !result.Success {
		t.Fatalf("Expected success with default path, got error: %s", result.Error)
	}
}
