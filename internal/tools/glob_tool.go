// ABOUTME: Glob tool implementation - file pattern matching with doublestar support
// ABOUTME: Sorts results by modification time (newest first)

package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// GlobTool finds files matching glob patterns
type GlobTool struct{}

// NewGlobTool creates a new Glob tool instance
func NewGlobTool() *GlobTool {
	return &GlobTool{}
}

// Name returns the tool identifier
func (t *GlobTool) Name() string {
	return "glob"
}

// Description returns the tool's purpose
func (t *GlobTool) Description() string {
	return "Find files by glob pattern. Supports ** for recursive matching. Sorted by modification time (newest first)."
}

// RequiresApproval returns false (glob is read-only)
func (t *GlobTool) RequiresApproval(_ map[string]interface{}) bool {
	// Glob is read-only, no approval needed
	return false
}

// Execute performs the glob search
func (t *GlobTool) Execute(_ context.Context, params map[string]interface{}) (*Result, error) {
	// Extract pattern (required)
	pattern, ok := params["pattern"].(string)
	if !ok || pattern == "" {
		return &Result{
			ToolName: t.Name(),
			Success:  false,
			Error:    "missing or invalid 'pattern' parameter",
		}, nil
	}

	// Extract path (default to current directory)
	searchPath := "."
	if pathParam, ok := params["path"].(string); ok && pathParam != "" {
		searchPath = pathParam
	}

	// Verify path exists
	if _, err := os.Stat(searchPath); err != nil {
		return &Result{
			ToolName: t.Name(),
			Success:  false,
			Error:    fmt.Sprintf("path does not exist: %s", searchPath),
		}, nil
	}

	// Find matching files
	matchSet := make(map[string]bool) // Deduplicate

	// Expand brace patterns first (e.g., *.{ts,tsx} -> [*.ts, *.tsx])
	patterns := expandBraces(pattern)

	for _, pat := range patterns {
		// Combine path and pattern
		fullPattern := filepath.Join(searchPath, pat)

		// Handle ** (doublestar) for recursive matching
		if strings.Contains(pat, "**") {
			// Walk the directory tree manually
			err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return nil // Skip inaccessible paths
				}
				if info.IsDir() {
					return nil // Skip directories themselves
				}

				// Get relative path from searchPath
				relPath, err := filepath.Rel(searchPath, path)
				if err != nil {
					return nil
				}

				// Check if it matches the pattern
				matched, err := matchPattern(pat, relPath)
				if err != nil {
					return nil
				}

				if matched {
					matchSet[path] = true
				}

				return nil
			})

			if err != nil {
				return &Result{
					ToolName: t.Name(),
					Success:  false,
					Error:    fmt.Sprintf("failed to walk directory: %v", err),
				}, nil
			}
		} else {
			// Use standard filepath.Glob for simple patterns
			foundMatches, err := filepath.Glob(fullPattern)
			if err != nil {
				return &Result{
					ToolName: t.Name(),
					Success:  false,
					Error:    fmt.Sprintf("invalid pattern: %v", err),
				}, nil
			}

			// Filter out directories and add to set
			for _, match := range foundMatches {
				info, err := os.Stat(match)
				if err != nil {
					continue
				}
				if !info.IsDir() {
					matchSet[match] = true
				}
			}
		}
	}

	// Convert set to slice
	matches := make([]string, 0, len(matchSet))
	for match := range matchSet {
		matches = append(matches, match)
	}

	// No matches is still success
	if len(matches) == 0 {
		return &Result{
			ToolName: t.Name(),
			Success:  true,
			Output:   "",
		}, nil
	}

	// Sort by modification time (newest first)
	sort.Slice(matches, func(i, j int) bool {
		infoI, err1 := os.Stat(matches[i])
		infoJ, err2 := os.Stat(matches[j])

		if err1 != nil || err2 != nil {
			return false
		}

		return infoI.ModTime().After(infoJ.ModTime())
	})

	// Format output
	output := strings.Join(matches, "\n")

	return &Result{
		ToolName: t.Name(),
		Success:  true,
		Output:   output,
	}, nil
}

// matchPattern matches a file path against a glob pattern with ** support
func matchPattern(pattern, path string) (bool, error) {
	// Normalize path separators
	pattern = filepath.ToSlash(pattern)
	path = filepath.ToSlash(path)

	// Handle ** by converting to a regex-like pattern
	// src/**/*.tsx should match src/anything/file.tsx
	if strings.Contains(pattern, "**") {
		// Split pattern around **
		parts := strings.Split(pattern, "**")
		if len(parts) == 2 {
			prefix := parts[0]
			suffix := strings.TrimPrefix(parts[1], "/")

			// Check if path starts with prefix (if any)
			if prefix != "" {
				prefix = strings.TrimSuffix(prefix, "/")
				if !strings.HasPrefix(path, prefix) {
					return false, nil
				}
				// Remove prefix from path
				path = strings.TrimPrefix(path, prefix)
				path = strings.TrimPrefix(path, "/")
			}

			// Check if remaining path matches suffix
			if suffix != "" {
				matched, err := filepath.Match(suffix, filepath.Base(path))
				if err != nil {
					return false, err
				}
				return matched, nil
			}

			return true, nil
		}
	}

	// Simple pattern without **
	matched, err := filepath.Match(pattern, path)
	if err != nil {
		return false, err
	}
	if matched {
		return true, nil
	}

	// Try matching just the basename
	matched, err = filepath.Match(pattern, filepath.Base(path))
	return matched, err
}

// expandBraces expands {a,b,c} patterns into multiple patterns
func expandBraces(pattern string) []string {
	// Simple brace expansion for patterns like *.{ts,tsx}
	if !strings.Contains(pattern, "{") {
		return []string{pattern}
	}

	start := strings.Index(pattern, "{")
	end := strings.Index(pattern, "}")

	if start == -1 || end == -1 || end < start {
		return []string{pattern}
	}

	prefix := pattern[:start]
	suffix := pattern[end+1:]
	options := strings.Split(pattern[start+1:end], ",")

	results := make([]string, 0, len(options))
	for _, opt := range options {
		results = append(results, prefix+opt+suffix)
	}

	return results
}
