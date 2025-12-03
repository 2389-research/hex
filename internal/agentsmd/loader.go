// Package agentsmd provides hierarchical AGENTS.md file loading with directory traversal.
// ABOUTME: Loads AGENTS.md files from directory traversal (git root → CWD)
// ABOUTME: Supports override and merge semantics inspired by OpenAI Codex
package agentsmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// LoadContext loads AGENTS.md files from the directory hierarchy
// Strategy:
//  1. Load global ~/.clem/AGENTS.md (if exists)
//  2. Find git repository root
//  3. Traverse from repo root → current working directory
//  4. At each level, check for:
//     - AGENTS.override.md (replaces previous REPOSITORY context, preserves global)
//     - AGENTS.md (merges with previous context)
//
// Returns the combined context string and any error
func LoadContext() (string, error) {
	var globalContext string
	var repoContext strings.Builder

	// 1. Load global AGENTS.md from ~/.clem/
	// Global context is NEVER overridden by repository files
	homeDir, err := os.UserHomeDir()
	if err == nil {
		globalPath := filepath.Join(homeDir, ".clem", "AGENTS.md")
		//nolint:gosec // Intentional file read from user's home directory
		if content, err := os.ReadFile(globalPath); err == nil {
			globalContext = fmt.Sprintf("# Global Context\n\n%s\n\n", string(content))
		}
	}

	// 2. Find git repository root
	repoRoot, err := findGitRoot()
	if err != nil {
		// Not in a git repo - just return global context if any
		return globalContext, nil
	}

	// 3. Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return globalContext, fmt.Errorf("get working directory: %w", err)
	}

	// 4. Build path from repo root to CWD
	relPath, err := filepath.Rel(repoRoot, cwd)
	if err != nil {
		return globalContext, fmt.Errorf("compute relative path: %w", err)
	}

	// 5. Traverse each directory level from repo root to CWD
	// Start with the repo root itself
	var pathsToCheck []string
	pathsToCheck = append(pathsToCheck, repoRoot)

	// Add each subdirectory in the path from root to CWD
	if relPath != "." {
		pathParts := strings.Split(relPath, string(filepath.Separator))
		currentPath := repoRoot
		for _, part := range pathParts {
			if part == "" || part == "." {
				continue
			}
			currentPath = filepath.Join(currentPath, part)
			pathsToCheck = append(pathsToCheck, currentPath)
		}
	}

	for _, currentPath := range pathsToCheck {

		// Check for AGENTS.override.md first (replaces repository context only)
		overridePath := filepath.Join(currentPath, "AGENTS.override.md")
		//nolint:gosec // Intentional file read from repository directories
		if content, err := os.ReadFile(overridePath); err == nil {
			// Override found - reset REPOSITORY context only (preserves global)
			repoContext.Reset()
			repoContext.WriteString(fmt.Sprintf("# Context from %s\n\n", overridePath))
			repoContext.Write(content)
			repoContext.WriteString("\n\n")
			continue
		}

		// Check for AGENTS.md (merges with existing context)
		agentsPath := filepath.Join(currentPath, "AGENTS.md")
		//nolint:gosec // Intentional file read from repository directories
		if content, err := os.ReadFile(agentsPath); err == nil {
			repoContext.WriteString(fmt.Sprintf("# Context from %s\n\n", agentsPath))
			repoContext.Write(content)
			repoContext.WriteString("\n\n")
		}
	}

	// Combine global context (always first) with repository context
	return globalContext + repoContext.String(), nil
}

// findGitRoot finds the root of the git repository
// Returns the absolute path to the repo root, or error if not in a git repo
func findGitRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("not in a git repository")
	}

	root := strings.TrimSpace(string(output))
	return root, nil
}
