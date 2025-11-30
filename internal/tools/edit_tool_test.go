// ABOUTME: Comprehensive tests for Edit tool - exact string replacement
// ABOUTME: Covers basic edits, replace_all, safety checks, error handling

package tools

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// ===========================
// Basic Edit Tests
// ===========================

func TestEditTool_BasicEdit(t *testing.T) {
	tool := NewEditTool()

	// Create temp file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "Hello World\nThis is a test\n"
	if err := os.WriteFile(testFile, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	params := map[string]interface{}{
		"file_path":  testFile,
		"old_string": "World",
		"new_string": "Universe",
	}

	result, err := tool.Execute(context.Background(), params)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("Expected success, got error: %s", result.Error)
	}

	// Verify file was modified
	//nolint:gosec // G304: Test file reads/writes are safe
	newContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read modified file: %v", err)
	}

	expected := "Hello Universe\nThis is a test\n"
	if string(newContent) != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, string(newContent))
	}
}

func TestEditTool_MultilineEdit(t *testing.T) {
	tool := NewEditTool()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "multiline.txt")
	content := "func main() {\n\tfmt.Println(\"old\")\n}\n"
	if err := os.WriteFile(testFile, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	params := map[string]interface{}{
		"file_path":  testFile,
		"old_string": "func main() {\n\tfmt.Println(\"old\")\n}",
		"new_string": "func main() {\n\tfmt.Println(\"new\")\n\tfmt.Println(\"updated\")\n}",
	}

	result, err := tool.Execute(context.Background(), params)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("Expected success, got error: %s", result.Error)
	}
	//nolint:gosec // G304: Test file reads/writes are safe

	newContent, err := os.ReadFile(testFile) //nolint:gosec // G304: Path validated by caller
	if err != nil {
		t.Fatalf("Failed to read modified file: %v", err)
	}

	expected := "func main() {\n\tfmt.Println(\"new\")\n\tfmt.Println(\"updated\")\n}\n"
	if string(newContent) != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, string(newContent))
	}
}

func TestEditTool_PreserveIndentation(t *testing.T) {
	tool := NewEditTool()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "indent.txt")
	content := "func test() {\n\t\tif true {\n\t\t\treturn \"old\"\n\t\t}\n}\n"
	if err := os.WriteFile(testFile, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	params := map[string]interface{}{
		"file_path":  testFile,
		"old_string": "\t\t\treturn \"old\"",
		"new_string": "\t\t\treturn \"new\"",
	}

	result, err := tool.Execute(context.Background(), params)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("Expected success, got error: %s", result.Error)
		//nolint:gosec // G304: Test file reads/writes are safe
	}

	//nolint:gosec // G304: Test file reads/writes are safe
	newContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read modified file: %v", err)
	}

	expected := "func test() {\n\t\tif true {\n\t\t\treturn \"new\"\n\t\t}\n}\n"
	if string(newContent) != expected {
		t.Errorf("Indentation not preserved.\nExpected:\n%s\nGot:\n%s", expected, string(newContent))
	}
}

// ===========================
// Replace All Tests
// ===========================

func TestEditTool_ReplaceAll(t *testing.T) {
	tool := NewEditTool()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "replaceall.txt")
	content := "foo bar foo baz foo"
	if err := os.WriteFile(testFile, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	params := map[string]interface{}{
		"file_path":   testFile,
		"old_string":  "foo",
		"new_string":  "qux",
		"replace_all": true,
	}

	result, err := tool.Execute(context.Background(), params)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !result.Success {
		//nolint:gosec // G304: Test file reads/writes are safe
		t.Fatalf("Expected success, got error: %s", result.Error)
	}

	//nolint:gosec // G304: Test file reads/writes are safe
	newContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read modified file: %v", err)
	}

	expected := "qux bar qux baz qux"
	if string(newContent) != expected {
		t.Errorf("Expected all occurrences replaced.\nExpected: %s\nGot: %s", expected, string(newContent))
	}
}

func TestEditTool_ReplaceAllWithNewlines(t *testing.T) {
	tool := NewEditTool()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "replaceall_newlines.txt")
	content := "log.Println(\"test\")\nlog.Println(\"foo\")\nlog.Println(\"bar\")\n"
	if err := os.WriteFile(testFile, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	params := map[string]interface{}{
		"file_path":   testFile,
		"old_string":  "log.Println",
		"new_string":  "fmt.Println",
		"replace_all": true,
	}

	result, err := tool.Execute(context.Background(), params)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	//nolint:gosec // G304: Test file reads/writes are safe
	if !result.Success {
		t.Fatalf("Expected success, got error: %s", result.Error)
	}

	//nolint:gosec // G304: Test file reads/writes are safe
	newContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read modified file: %v", err)
	}

	expected := "fmt.Println(\"test\")\nfmt.Println(\"foo\")\nfmt.Println(\"bar\")\n"
	if string(newContent) != expected {
		t.Errorf("Expected all log.Println replaced.\nExpected:\n%s\nGot:\n%s", expected, string(newContent))
	}
}

// ===========================
// Error Cases
// ===========================

func TestEditTool_AmbiguousMatch(t *testing.T) {
	tool := NewEditTool()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "ambiguous.txt")
	content := "foo bar foo baz foo"
	if err := os.WriteFile(testFile, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	params := map[string]interface{}{
		"file_path":  testFile,
		"old_string": "foo",
		"new_string": "qux",
		// replace_all not set - should fail on ambiguous match
	}

	result, err := tool.Execute(context.Background(), params)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.Success {
		t.Fatal("Expected failure for ambiguous match without replace_all")
	}

	if result.Error == "" {
		t.Error("Expected error message for ambiguous match")
	}
}

func TestEditTool_StringNotFound(t *testing.T) {
	tool := NewEditTool()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "notfound.txt")
	content := "Hello World"
	if err := os.WriteFile(testFile, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	params := map[string]interface{}{
		"file_path":  testFile,
		"old_string": "NonExistent",
		"new_string": "Something",
	}

	result, err := tool.Execute(context.Background(), params)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.Success {
		t.Fatal("Expected failure when old_string not found")
	}

	if result.Error == "" {
		t.Error("Expected error message when string not found")
	}
}

func TestEditTool_FileNotFound(t *testing.T) {
	tool := NewEditTool()

	params := map[string]interface{}{
		"file_path":  "/nonexistent/file.txt",
		"old_string": "foo",
		"new_string": "bar",
	}

	result, err := tool.Execute(context.Background(), params)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.Success {
		t.Fatal("Expected failure when file doesn't exist")
	}
}

func TestEditTool_MissingParameters(t *testing.T) {
	tool := NewEditTool()

	tests := []struct {
		name   string
		params map[string]interface{}
	}{
		{
			name:   "missing file_path",
			params: map[string]interface{}{"old_string": "foo", "new_string": "bar"},
		},
		{
			name:   "missing old_string",
			params: map[string]interface{}{"file_path": "/tmp/test.txt", "new_string": "bar"},
		},
		{
			name:   "missing new_string",
			params: map[string]interface{}{"file_path": "/tmp/test.txt", "old_string": "foo"},
		},
		{
			name:   "empty old_string",
			params: map[string]interface{}{"file_path": "/tmp/test.txt", "old_string": "", "new_string": "bar"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tool.Execute(context.Background(), tt.params)
			if err != nil {
				t.Fatalf("Execute failed: %v", err)
			}

			if result.Success {
				t.Errorf("%s: expected failure", tt.name)
			}
		})
	}
}

func TestEditTool_InvalidParameterTypes(t *testing.T) {
	tool := NewEditTool()

	params := map[string]interface{}{
		"file_path":   123, // Should be string
		"old_string":  "foo",
		"new_string":  "bar",
		"replace_all": "yes", // Should be bool
	}

	result, err := tool.Execute(context.Background(), params)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.Success {
		t.Error("Expected failure with invalid parameter types")
	}
}

// ===========================
// Safety Tests
// ===========================

func TestEditTool_AlwaysRequiresApproval(t *testing.T) {
	tool := NewEditTool()

	// Any edit should require approval
	params := map[string]interface{}{
		"file_path":  "/tmp/test.txt",
		"old_string": "foo",
		"new_string": "bar",
	}

	if !tool.RequiresApproval(params) {
		t.Error("Edit tool should ALWAYS require approval")
	}

	// Even with empty params
	if !tool.RequiresApproval(map[string]interface{}{}) {
		t.Error("Edit tool should require approval even with empty params")
	}
}

// ===========================
// Metadata Tests
// ===========================

func TestEditTool_Metadata(t *testing.T) {
	tool := NewEditTool()

	if tool.Name() != "edit" {
		t.Errorf("Expected name 'edit', got '%s'", tool.Name())
	}

	if tool.Description() == "" {
		t.Error("Expected non-empty description")
	}
}

// ===========================
// Edge Cases
// ===========================

func TestEditTool_EmptyFile(t *testing.T) {
	tool := NewEditTool()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "empty.txt")
	if err := os.WriteFile(testFile, []byte(""), 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	params := map[string]interface{}{
		"file_path":  testFile,
		"old_string": "foo",
		"new_string": "bar",
	}

	result, err := tool.Execute(context.Background(), params)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.Success {
		t.Fatal("Expected failure when searching empty file")
	}
}

func TestEditTool_SameOldAndNew(t *testing.T) {
	tool := NewEditTool()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "same.txt")
	content := "Hello World"
	if err := os.WriteFile(testFile, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	params := map[string]interface{}{
		"file_path":  testFile,
		"old_string": "World",
		"new_string": "World", // Same as old
	}

	result, err := tool.Execute(context.Background(), params)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.Success {
		t.Fatal("Expected failure when old_string and new_string are identical")
	}
}

func TestEditTool_UnicodeContent(t *testing.T) {
	tool := NewEditTool()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "unicode.txt")
	content := "Hello 世界 🌍"
	if err := os.WriteFile(testFile, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	params := map[string]interface{}{
		"file_path":  testFile,
		"old_string": "世界",
		"new_string": "World",
	}

	result, err := tool.Execute(context.Background(), params)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	//nolint:gosec // G304: Test file reads/writes are safe

	if !result.Success {
		t.Fatalf("Expected success with unicode, got error: %s", result.Error)
	}

	//nolint:gosec // G304: Test file reads/writes are safe
	newContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read modified file: %v", err)
	}

	expected := "Hello World 🌍"
	if string(newContent) != expected {
		t.Errorf("Expected: %s\nGot: %s", expected, string(newContent))
	}
}
