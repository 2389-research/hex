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

	"github.com/harper/clem/internal/frontmatter"
	"gopkg.in/yaml.v3"
)

// Command represents a slash command with metadata and template content
type Command struct {
	// Name is the unique identifier for the command (used with /name)
	Name string `yaml:"name"`
	// Description explains what the command does
	Description string `yaml:"description"`
	// Args maps argument names to their descriptions for template expansion
	Args map[string]string `yaml:"args"`

	// Content is the markdown template body after frontmatter
	Content string `yaml:"-"`

	// FilePath is the absolute path to the command file
	FilePath string `yaml:"-"`
	// Source indicates where the command was loaded from: "user", "project", or "builtin"
	Source string `yaml:"-"`
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
	fm, content, err := frontmatter.Split(data)
	if err != nil {
		return nil, fmt.Errorf("split frontmatter: %w", err)
	}

	// Parse YAML frontmatter
	var cmd Command
	if len(fm) > 0 {
		if err := yaml.Unmarshal(fm, &cmd); err != nil {
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
