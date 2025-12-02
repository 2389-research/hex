// ABOUTME: Project directory discovery utilities
// ABOUTME: Finds .clem directories by searching upward from current directory

// Package project provides utilities for finding project directories.
package project

import (
	"os"
	"path/filepath"
)

const (
	// MaxDirSearchDepth limits how many parent directories to search
	// to prevent infinite loops and excessive filesystem traversal
	MaxDirSearchDepth = 10
)

// FindDir searches for a .clem subdirectory by walking up the directory tree
// from the current working directory. Returns the full path to the found subdirectory
// or an empty string if not found within MaxDirSearchDepth levels.
//
// Example: If cwd is /home/user/projects/myapp/src and .clem/skills exists at
// /home/user/projects/myapp/.clem/skills, calling FindDir("skills") returns
// "/home/user/projects/myapp/.clem/skills"
func FindDir(subdir string) string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}

	// Search upwards for .clem directory
	searchDir := cwd
	for i := 0; i < MaxDirSearchDepth; i++ {
		clemDir := filepath.Join(searchDir, ".clem", subdir)
		if info, err := os.Stat(clemDir); err == nil && info.IsDir() {
			return clemDir
		}

		parent := filepath.Dir(searchDir)
		if parent == searchDir {
			break // Reached filesystem root
		}
		searchDir = parent
	}

	return ""
}
