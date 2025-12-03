// ABOUTME: Project directory discovery utilities
// ABOUTME: Finds .hex directories by searching upward from current directory

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

// FindDir searches for a .hex subdirectory by walking up the directory tree
// from the current working directory. Returns the full path to the found subdirectory
// or an empty string if not found within MaxDirSearchDepth levels.
//
// Example: If cwd is /home/user/projects/myapp/src and .hex/skills exists at
// /home/user/projects/myapp/.hex/skills, calling FindDir("skills") returns
// "/home/user/projects/myapp/.hex/skills"
func FindDir(subdir string) string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}

	// Search upwards for .hex directory
	searchDir := cwd
	for i := 0; i < MaxDirSearchDepth; i++ {
		hexDir := filepath.Join(searchDir, ".hex", subdir)
		if info, err := os.Stat(hexDir); err == nil && info.IsDir() {
			return hexDir
		}

		parent := filepath.Dir(searchDir)
		if parent == searchDir {
			break // Reached filesystem root
		}
		searchDir = parent
	}

	return ""
}
