// ABOUTME: Tests for project directory discovery
// ABOUTME: Verifies upward directory search and depth limits

package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindDir(t *testing.T) {
	// Create a temporary directory structure
	tmpDir := t.TempDir()

	// Create .clem/skills directory
	skillsDir := filepath.Join(tmpDir, ".clem", "skills")
	if err := os.MkdirAll(skillsDir, 0o750); err != nil { //nolint:gosec // G301 - test directory
		t.Fatal(err)
	}

	// Create a subdirectory
	subDir := filepath.Join(tmpDir, "src", "nested")
	if err := os.MkdirAll(subDir, 0o750); err != nil { //nolint:gosec // G301 - test directory
		t.Fatal(err)
	}

	// Save original cwd and restore after test
	origCwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(origCwd); err != nil {
			t.Error(err)
		}
	}()

	// Change to subdirectory
	if err := os.Chdir(subDir); err != nil {
		t.Fatal(err)
	}

	// Should find .clem/skills by walking up
	found := FindDir("skills")

	// Resolve symlinks for comparison (macOS /var -> /private/var)
	wantPath, err := filepath.EvalSymlinks(skillsDir)
	if err != nil {
		wantPath = skillsDir
	}
	foundPath, err := filepath.EvalSymlinks(found)
	if err != nil {
		foundPath = found
	}

	if foundPath != wantPath {
		t.Errorf("FindDir(skills) = %q, want %q", foundPath, wantPath)
	}
}

func TestFindDirNotFound(t *testing.T) {
	// Create a temporary directory without .clem
	tmpDir := t.TempDir()

	// Save original cwd and restore after test
	origCwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(origCwd); err != nil {
			t.Error(err)
		}
	}()

	// Change to temp directory
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	// Should return empty string when not found
	found := FindDir("skills")
	if found != "" {
		t.Errorf("FindDir(skills) = %q, want empty string", found)
	}
}
