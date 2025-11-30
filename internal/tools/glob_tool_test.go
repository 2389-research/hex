// ABOUTME: Comprehensive tests for Glob tool - file pattern matching
// ABOUTME: Covers glob patterns, sorting by modification time, subdirectories

package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// ===========================
// Basic Glob Tests
// ===========================

func TestGlobTool_BasicPattern(t *testing.T) {
	tool := NewGlobTool()

	tmpDir := t.TempDir()

	files := []string{"test.go", "main.go", "test.txt", "README.md"}
	for _, name := range files {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte("content"), 0600); err != nil {
			t.Fatalf("Failed to create %s: %v", name, err)
		}
	}

	params := map[string]interface{}{
		"pattern": "*.go",
		"path":    tmpDir,
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
		t.Errorf("Expected both .go files, got: %s", output)
	}
	if strings.Contains(output, "test.txt") || strings.Contains(output, "README.md") {
		t.Errorf("Did not expect non-.go files, got: %s", output)
	}
}

func TestGlobTool_RecursivePattern(t *testing.T) {
	tool := NewGlobTool()

	tmpDir := t.TempDir()

	// Create nested structure
	subdirs := []string{
		"src",
		"src/components",
		"test",
	}
	for _, dir := range subdirs {
		if err := os.MkdirAll(filepath.Join(tmpDir, dir), 0750); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	files := []string{
		"main.go",
		"src/app.go",
		"src/components/button.go",
		"test/test.go",
	}
	for _, name := range files {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte("content"), 0600); err != nil {
			t.Fatalf("Failed to create %s: %v", name, err)
		}
	}

	params := map[string]interface{}{
		"pattern": "**/*.go",
		"path":    tmpDir,
	}

	result, err := tool.Execute(context.Background(), params)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("Expected success, got error: %s", result.Error)
	}

	output := result.Output
	// Should find all 4 .go files
	expected := []string{"main.go", "app.go", "button.go", "test.go"}
	for _, exp := range expected {
		if !strings.Contains(output, exp) {
			t.Errorf("Expected to find %s in output, got: %s", exp, output)
		}
	}
}

func TestGlobTool_MultipleExtensions(t *testing.T) {
	tool := NewGlobTool()

	tmpDir := t.TempDir()

	files := []string{"test.ts", "test.tsx", "test.js", "test.go"}
	for _, name := range files {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte("content"), 0600); err != nil {
			t.Fatalf("Failed to create %s: %v", name, err)
		}
	}

	params := map[string]interface{}{
		"pattern": "*.{ts,tsx}",
		"path":    tmpDir,
	}

	result, err := tool.Execute(context.Background(), params)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("Expected success, got error: %s", result.Error)
	}

	output := result.Output
	if !strings.Contains(output, "test.ts") || !strings.Contains(output, "test.tsx") {
		t.Errorf("Expected both .ts and .tsx files, got: %s", output)
	}
	if strings.Contains(output, "test.js") || strings.Contains(output, "test.go") {
		t.Errorf("Did not expect .js or .go files, got: %s", output)
	}
}

// ===========================
// Sorting Tests
// ===========================

func TestGlobTool_SortByModTime(t *testing.T) {
	tool := NewGlobTool()

	tmpDir := t.TempDir()

	// Create files with different modification times
	oldest := filepath.Join(tmpDir, "oldest.txt")
	middle := filepath.Join(tmpDir, "middle.txt")
	newest := filepath.Join(tmpDir, "newest.txt")

	// Create oldest first
	if err := os.WriteFile(oldest, []byte("old"), 0600); err != nil {
		t.Fatalf("Failed to create oldest: %v", err)
	}
	time.Sleep(10 * time.Millisecond)

	// Create middle
	if err := os.WriteFile(middle, []byte("mid"), 0600); err != nil {
		t.Fatalf("Failed to create middle: %v", err)
	}
	time.Sleep(10 * time.Millisecond)

	// Create newest
	if err := os.WriteFile(newest, []byte("new"), 0600); err != nil {
		t.Fatalf("Failed to create newest: %v", err)
	}

	params := map[string]interface{}{
		"pattern": "*.txt",
		"path":    tmpDir,
	}

	result, err := tool.Execute(context.Background(), params)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("Expected success, got error: %s", result.Error)
	}

	output := result.Output
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) != 3 {
		t.Fatalf("Expected 3 files, got %d", len(lines))
	}

	// Should be sorted newest first (reverse chronological)
	if !strings.Contains(lines[0], "newest.txt") {
		t.Errorf("Expected newest.txt first, got: %s", lines[0])
	}
	if !strings.Contains(lines[2], "oldest.txt") {
		t.Errorf("Expected oldest.txt last, got: %s", lines[2])
	}
}

// ===========================
// Error Cases
// ===========================

func TestGlobTool_MissingPattern(t *testing.T) {
	tool := NewGlobTool()

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

func TestGlobTool_InvalidPath(t *testing.T) {
	tool := NewGlobTool()

	params := map[string]interface{}{
		"pattern": "*.go",
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

func TestGlobTool_NoMatches(t *testing.T) {
	tool := NewGlobTool()

	tmpDir := t.TempDir()

	// Create some files that won't match
	files := []string{"test.txt", "README.md"}
	for _, name := range files {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte("content"), 0600); err != nil {
			t.Fatalf("Failed to create %s: %v", name, err)
		}
	}

	params := map[string]interface{}{
		"pattern": "*.go",
		"path":    tmpDir,
	}

	result, err := tool.Execute(context.Background(), params)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// No matches is still success, just empty output
	if !result.Success {
		t.Fatalf("Expected success with no matches, got error: %s", result.Error)
	}

	if result.Output != "" {
		t.Errorf("Expected empty output for no matches, got: %s", result.Output)
	}
}

// ===========================
// Metadata Tests
// ===========================

func TestGlobTool_Metadata(t *testing.T) {
	tool := NewGlobTool()

	if tool.Name() != "glob" {
		t.Errorf("Expected name 'glob', got '%s'", tool.Name())
	}

	if tool.Description() == "" {
		t.Error("Expected non-empty description")
	}
}

func TestGlobTool_RequiresApproval(t *testing.T) {
	tool := NewGlobTool()

	params := map[string]interface{}{
		"pattern": "*.go",
		"path":    "/tmp",
	}

	// Glob is read-only, should not require approval
	if tool.RequiresApproval(params) {
		t.Error("Glob should not require approval")
	}
}

// ===========================
// Default Path Tests
// ===========================

func TestGlobTool_DefaultPath(t *testing.T) {
	tool := NewGlobTool()

	// When path is not provided, should use current directory
	params := map[string]interface{}{
		"pattern": "*.go",
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

// ===========================
// Specific Pattern Tests
// ===========================

func TestGlobTool_DirectoryPattern(t *testing.T) {
	tool := NewGlobTool()

	tmpDir := t.TempDir()

	// Create nested structure
	subdirs := []string{
		"src/components",
		"test/unit",
	}
	for _, dir := range subdirs {
		if err := os.MkdirAll(filepath.Join(tmpDir, dir), 0750); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	files := []string{
		"src/components/button.tsx",
		"src/components/input.tsx",
		"test/unit/button.test.tsx",
	}
	for _, name := range files {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte("content"), 0600); err != nil {
			t.Fatalf("Failed to create %s: %v", name, err)
		}
	}

	params := map[string]interface{}{
		"pattern": "src/**/*.tsx",
		"path":    tmpDir,
	}

	result, err := tool.Execute(context.Background(), params)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("Expected success, got error: %s", result.Error)
	}

	output := result.Output
	if !strings.Contains(output, "button.tsx") || !strings.Contains(output, "input.tsx") {
		t.Errorf("Expected both component files, got: %s", output)
	}
	if strings.Contains(output, "button.test.tsx") {
		t.Errorf("Did not expect test file, got: %s", output)
	}
}
