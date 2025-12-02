// ABOUTME: Command struct and YAML frontmatter parsing
// ABOUTME: Represents slash command files with metadata and template content

// Package commands provides slash command loading and execution.
package commands

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

// Command represents a slash command with metadata and template content
type Command struct {
	// Frontmatter fields
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Args        map[string]string `yaml:"args"` // arg name -> description

	// Command content (markdown body after frontmatter)
	Content string `yaml:"-"`

	// File location metadata
	FilePath string `yaml:"-"`
	Source   string `yaml:"-"` // "user", "project", or "builtin"
}

// Parse reads and parses a command file from disk
func Parse(path string) (*Command, error) {
	data, err := os.ReadFile(path) //nolint:gosec // G304 - file paths from trusted config
	if err != nil {
		return nil, fmt.Errorf("read command file: %w", err)
	}

	return ParseBytes(path, data)
}

// ParseBytes parses command data from bytes
func ParseBytes(path string, data []byte) (*Command, error) {
	// Split frontmatter and content
	frontmatter, content, err := splitFrontmatter(data)
	if err != nil {
		return nil, fmt.Errorf("split frontmatter: %w", err)
	}

	// Parse YAML frontmatter
	var cmd Command
	if len(frontmatter) > 0 {
		if err := yaml.Unmarshal(frontmatter, &cmd); err != nil {
			return nil, fmt.Errorf("parse YAML frontmatter: %w", err)
		}
	}

	// Validate required fields
	if cmd.Name == "" {
		return nil, fmt.Errorf("command missing required 'name' field")
	}
	if cmd.Description == "" {
		return nil, fmt.Errorf("command missing required 'description' field")
	}

	// Store content and metadata
	cmd.Content = string(content)
	cmd.FilePath = path

	return &cmd, nil
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

// Expand expands the command template with provided arguments
func (c *Command) Expand(args map[string]interface{}) (string, error) {
	// Create template
	tmpl, err := template.New(c.Name).Parse(c.Content)
	if err != nil {
		return "", fmt.Errorf("parse command template: %w", err)
	}

	// Execute template with args
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, args); err != nil {
		return "", fmt.Errorf("execute command template: %w", err)
	}

	return buf.String(), nil
}

// String returns a formatted representation for display
func (c *Command) String() string {
	return fmt.Sprintf("Command{name=%s, source=%s}", c.Name, c.Source)
}

// HasArgs returns true if the command expects arguments
func (c *Command) HasArgs() bool {
	return len(c.Args) > 0
}

// ArgNames returns the names of expected arguments
func (c *Command) ArgNames() []string {
	names := make([]string, 0, len(c.Args))
	for name := range c.Args {
		names = append(names, name)
	}
	return names
}

// UsageString returns a formatted usage string for the command
func (c *Command) UsageString() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("/%s", c.Name))

	if c.HasArgs() {
		for _, argName := range c.ArgNames() {
			sb.WriteString(fmt.Sprintf(" <%s>", argName))
		}
	}

	return sb.String()
}
