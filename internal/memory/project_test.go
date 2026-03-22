// ABOUTME: Tests for project memory scanner and loader
// ABOUTME: Validates detection of project type, build commands, and persistence
package memory

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDetectProject_GoProject(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/test\ngo 1.21\n"), 0644)
	os.WriteFile(filepath.Join(dir, "Makefile"), []byte("build:\n\tgo build ./...\n"), 0644)
	os.MkdirAll(filepath.Join(dir, "cmd"), 0755)
	os.MkdirAll(filepath.Join(dir, "internal"), 0755)

	proj, err := DetectProject(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if proj.Language != "go" {
		t.Errorf("expected language 'go', got %q", proj.Language)
	}
	if proj.TestCommand != "go test ./..." {
		t.Errorf("expected test command 'go test ./...', got %q", proj.TestCommand)
	}
}

func TestDetectProject_NodeProject(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"name":"test"}`), 0644)

	proj, err := DetectProject(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if proj.Language != "javascript" {
		t.Errorf("expected language 'javascript', got %q", proj.Language)
	}
}

func TestDetectProject_PythonProject(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pyproject.toml"), []byte("[project]\nname = \"test\"\n"), 0644)

	proj, err := DetectProject(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if proj.Language != "python" {
		t.Errorf("expected language 'python', got %q", proj.Language)
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	hexDir := filepath.Join(dir, ".hex")

	proj := &ProjectInfo{
		Language:     "go",
		BuildCommand: "make build",
		TestCommand:  "go test ./...",
		Structure:    []string{"cmd/", "internal/"},
	}

	err := Save(hexDir, proj)
	if err != nil {
		t.Fatalf("save error: %v", err)
	}

	loaded, err := Load(hexDir)
	if err != nil {
		t.Fatalf("load error: %v", err)
	}
	if loaded.Language != "go" {
		t.Errorf("expected language 'go', got %q", loaded.Language)
	}
	if loaded.BuildCommand != "make build" {
		t.Errorf("expected build command 'make build', got %q", loaded.BuildCommand)
	}
}

func TestToPromptContext(t *testing.T) {
	proj := &ProjectInfo{
		Language:     "go",
		BuildCommand: "make build",
		TestCommand:  "go test ./...",
		Structure:    []string{"cmd/", "internal/", "docs/"},
	}

	ctx := proj.ToPromptContext()
	if ctx == "" {
		t.Error("expected non-empty prompt context")
	}
	if !strings.Contains(ctx, "go") {
		t.Error("prompt context should mention language")
	}
	if !strings.Contains(ctx, "make build") {
		t.Error("prompt context should mention build command")
	}
}
