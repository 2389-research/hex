// ABOUTME: YAML frontmatter parsing utilities
// ABOUTME: Separates YAML frontmatter from markdown content with safety limits

// Package frontmatter provides utilities for parsing YAML frontmatter from markdown files.
package frontmatter

import (
	"bytes"
	"fmt"
)

// MaxFrontmatterLines limits how far we search for frontmatter closing delimiter
// to prevent frontmatter bomb attacks (malicious files with unclosed frontmatter)
const MaxFrontmatterLines = 100

// Split separates YAML frontmatter from markdown content
// Returns frontmatter bytes, content bytes, and any error
//
// The function expects frontmatter to be delimited by "---" lines:
//
//	---
//	key: value
//	---
//	Content here...
//
// If no frontmatter is present, returns nil frontmatter and all data as content.
// If frontmatter is not closed within MaxFrontmatterLines, returns an error.
func Split(data []byte) (frontmatter, content []byte, err error) {
	// Check for frontmatter delimiter (---)
	if !bytes.HasPrefix(data, []byte("---\n")) && !bytes.HasPrefix(data, []byte("---\r\n")) {
		// No frontmatter, entire file is content
		return nil, data, nil
	}

	// Find closing delimiter
	lines := bytes.Split(data, []byte("\n"))
	endIdx := -1

	// Search for closing delimiter, but limit to prevent frontmatter bombs
	searchLimit := len(lines)
	if searchLimit > MaxFrontmatterLines {
		searchLimit = MaxFrontmatterLines
	}

	for i := 1; i < searchLimit; i++ {
		line := bytes.TrimSpace(lines[i])
		if bytes.Equal(line, []byte("---")) {
			endIdx = i
			break
		}
	}

	if endIdx == -1 {
		return nil, nil, fmt.Errorf("unclosed frontmatter: missing closing '---' within first %d lines", MaxFrontmatterLines)
	}

	// Extract frontmatter (between delimiters)
	frontmatterLines := lines[1:endIdx]
	frontmatter = bytes.Join(frontmatterLines, []byte("\n"))

	// Extract content (after closing delimiter)
	if endIdx+1 < len(lines) {
		contentLines := lines[endIdx+1:]
		content = bytes.Join(contentLines, []byte("\n"))
	}

	return frontmatter, content, nil
}
