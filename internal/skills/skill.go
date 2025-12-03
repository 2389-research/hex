// ABOUTME: Skill struct and YAML frontmatter parsing
// ABOUTME: Represents individual skill files with metadata and content

package skills

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/2389-research/hex/internal/frontmatter"
	"gopkg.in/yaml.v3"
)

const (
	// DefaultSkillPriority is the priority assigned to skills without an explicit priority
	DefaultSkillPriority = 5
)

// Skill represents a reusable knowledge module with metadata
type Skill struct {
	// Name is the unique identifier for the skill
	Name string `yaml:"name"`
	// Description explains what the skill does and when to use it
	Description string `yaml:"description"`
	// Tags are labels for categorizing and filtering skills
	Tags []string `yaml:"tags"`
	// ActivationPatterns are regex patterns that trigger the skill
	ActivationPatterns []string `yaml:"activationPatterns"`
	// Model specifies the AI model to use for this skill
	Model string `yaml:"model"`
	// Priority determines skill precedence (higher values win)
	Priority int `yaml:"priority"`
	// Dependencies lists other skills this skill depends on
	Dependencies []string `yaml:"dependencies"`
	// Version tracks the skill version for compatibility
	Version string `yaml:"version"`

	// Content is the markdown body after frontmatter
	Content string `yaml:"-"`

	// FilePath is the absolute path to the skill file
	FilePath string `yaml:"-"`
	// Source indicates where the skill was loaded from: "user", "project", or "builtin"
	Source string `yaml:"-"`

	// Cached compiled regex patterns for activation matching
	compiledPatterns []*regexp.Regexp `yaml:"-"`
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
	fm, content, err := frontmatter.Split(data)
	if err != nil {
		return nil, fmt.Errorf("split frontmatter: %w", err)
	}

	// Parse YAML frontmatter
	var skill Skill
	if len(fm) > 0 {
		if err := yaml.Unmarshal(fm, &skill); err != nil {
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
		skill.Priority = DefaultSkillPriority
	}

	// Store content and metadata
	skill.Content = string(content)
	skill.FilePath = path

	// Pre-compile all activation pattern regexes for performance
	// Compile failures are silently ignored (pattern will be skipped during matching)
	skill.compiledPatterns = make([]*regexp.Regexp, 0, len(skill.ActivationPatterns))
	for _, pattern := range skill.ActivationPatterns {
		re, err := regexp.Compile("(?i)" + pattern)
		if err == nil {
			skill.compiledPatterns = append(skill.compiledPatterns, re)
		}
	}

	return &skill, nil
}

// MatchesPattern checks if a user message matches any activation pattern
// Uses pre-compiled regex patterns for optimal performance
// Falls back to on-demand compilation for backward compatibility with tests
func (s *Skill) MatchesPattern(message string) bool {
	// If patterns were pre-compiled during ParseBytes, use them
	if len(s.compiledPatterns) > 0 {
		lowerMsg := strings.ToLower(message)
		for _, re := range s.compiledPatterns {
			if re.MatchString(lowerMsg) {
				return true
			}
		}
		return false
	}

	// Fallback for skills created without ParseBytes (e.g., tests)
	if len(s.ActivationPatterns) == 0 {
		return false
	}

	lowerMsg := strings.ToLower(message)
	for _, pattern := range s.ActivationPatterns {
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
