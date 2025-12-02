// ABOUTME: Project directory discovery utilities
// ABOUTME: Finds .claude directories by searching upward from current directory

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

// FindDir searches for a .claude subdirectory by walking up the directory tree
// from the current working directory. Returns the full path to the found subdirectory
// or an empty string if not found within MaxDirSearchDepth levels.
//
// Example: If cwd is /home/user/projects/myapp/src and .claude/skills exists at
// /home/user/projects/myapp/.claude/skills, calling FindDir("skills") returns
// "/home/user/projects/myapp/.claude/skills"
func FindDir(subdir string) string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}

	// Search upwards for .claude directory
	searchDir := cwd
	for i := 0; i < MaxDirSearchDepth; i++ {
		claudeDir := filepath.Join(searchDir, ".claude", subdir)
		if info, err := os.Stat(claudeDir); err == nil && info.IsDir() {
			return claudeDir
		}

		parent := filepath.Dir(searchDir)
		if parent == searchDir {
			break // Reached filesystem root
		}
		searchDir = parent
	}

	return ""
}
