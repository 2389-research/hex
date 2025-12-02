// ABOUTME: Skill struct and YAML frontmatter parsing
// ABOUTME: Represents individual skill files with metadata and content

package skills

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// Skill represents a reusable knowledge module with metadata
type Skill struct {
	// Frontmatter fields
	Name               string   `yaml:"name"`
	Description        string   `yaml:"description"`
	Tags               []string `yaml:"tags"`
	ActivationPatterns []string `yaml:"activationPatterns"`
	Model              string   `yaml:"model"`
	Priority           int      `yaml:"priority"`
	Dependencies       []string `yaml:"dependencies"`
	Version            string   `yaml:"version"`

	// Skill content (markdown body after frontmatter)
	Content string `yaml:"-"`

	// File location metadata
	FilePath string `yaml:"-"`
	Source   string `yaml:"-"` // "user", "project", or "builtin"
}

// Parse reads and parses a skill file from disk
func Parse(path string) (*Skill, error) {
	data, err := os.ReadFile(path) //nolint:gosec // G304 - file paths from trusted config
	if err != nil {
		return nil, fmt.Errorf("read skill file: %w", err)
	}

	return ParseBytes(path, data)
}

// ParseBytes parses skill data from bytes
func ParseBytes(path string, data []byte) (*Skill, error) {
	// Split frontmatter and content
	frontmatter, content, err := splitFrontmatter(data)
	if err != nil {
		return nil, fmt.Errorf("split frontmatter: %w", err)
	}

	// Parse YAML frontmatter
	var skill Skill
	if len(frontmatter) > 0 {
		if err := yaml.Unmarshal(frontmatter, &skill); err != nil {
			return nil, fmt.Errorf("parse YAML frontmatter: %w", err)
		}
	}

	// Validate required fields
	if skill.Name == "" {
		return nil, fmt.Errorf("skill missing required 'name' field")
	}
	if skill.Description == "" {
		return nil, fmt.Errorf("skill missing required 'description' field")
	}

	// Set defaults
	if skill.Priority == 0 {
		skill.Priority = 5 // Default priority
	}

	// Store content and metadata
	skill.Content = string(content)
	skill.FilePath = path

	return &skill, nil
}

// splitFrontmatter separates YAML frontmatter from markdown content
func splitFrontmatter(data []byte) (frontmatter, content []byte, err error) {
	// Check for frontmatter delimiter (---)
	if !bytes.HasPrefix(data, []byte("---\n")) && !bytes.HasPrefix(data, []byte("---\r\n")) {
		// No frontmatter, entire file is content
		return nil, data, nil
	}

	// Find closing delimiter
	lines := bytes.Split(data, []byte("\n"))
	endIdx := -1
	for i := 1; i < len(lines); i++ {
		line := bytes.TrimSpace(lines[i])
		if bytes.Equal(line, []byte("---")) {
			endIdx = i
			break
		}
	}

	if endIdx == -1 {
		return nil, nil, fmt.Errorf("unclosed frontmatter: missing closing '---'")
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

// MatchesPattern checks if a user message matches any activation pattern
func (s *Skill) MatchesPattern(message string) bool {
	if len(s.ActivationPatterns) == 0 {
		return false
	}

	// Case-insensitive matching
	lowerMsg := strings.ToLower(message)

	for _, pattern := range s.ActivationPatterns {
		// Compile regex pattern (case-insensitive)
		re, err := regexp.Compile("(?i)" + pattern)
		if err != nil {
			// Invalid regex pattern, skip
			continue
		}

		if re.MatchString(lowerMsg) {
			return true
		}
	}

	return false
}

// String returns a formatted representation for display
func (s *Skill) String() string {
	return fmt.Sprintf("Skill{name=%s, source=%s, priority=%d}", s.Name, s.Source, s.Priority)
}
