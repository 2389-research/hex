// ABOUTME: Spell struct and configuration types for agent personality switching
// ABOUTME: Defines Spell, SpellConfig, and LayerMode types with validation

package spells

import "fmt"

// LayerMode determines how a spell interacts with existing system prompt
type LayerMode string

const (
	// LayerModeReplace completely replaces the existing system prompt
	LayerModeReplace LayerMode = "replace"
	// LayerModeLayer adds to the existing system prompt
	LayerModeLayer LayerMode = "layer"
)

// ReasoningEffort levels for modern AI models
type ReasoningEffort string

const (
	ReasoningEffortNone   ReasoningEffort = "none"
	ReasoningEffortLow    ReasoningEffort = "low"
	ReasoningEffortMedium ReasoningEffort = "medium"
	ReasoningEffortHigh   ReasoningEffort = "high"
)

// ToolsConfig defines tool availability for a spell
type ToolsConfig struct {
	Enabled  []string `yaml:"enabled,omitempty"`  // Tools to enable (empty = all)
	Disabled []string `yaml:"disabled,omitempty"` // Tools to disable
}

// ReasoningConfig defines reasoning behavior for modern models
type ReasoningConfig struct {
	Effort       ReasoningEffort `yaml:"effort,omitempty"`        // none, low, medium, high
	ShowThinking bool            `yaml:"show_thinking,omitempty"` // expose thinking blocks
}

// ResponseConfig defines response preferences
type ResponseConfig struct {
	MaxTokens int    `yaml:"max_tokens,omitempty"` // max response tokens
	Format    string `yaml:"format,omitempty"`     // text, json, markdown
	Style     string `yaml:"style,omitempty"`      // concise, detailed, code-first
}

// SamplingConfig defines legacy sampling parameters
type SamplingConfig struct {
	Temperature float64 `yaml:"temperature,omitempty"`
}

// SpellConfig holds all configuration options for a spell
type SpellConfig struct {
	Mode      LayerMode       `yaml:"mode,omitempty"`
	Tools     ToolsConfig     `yaml:"tools,omitempty"`
	Reasoning ReasoningConfig `yaml:"reasoning,omitempty"`
	Response  ResponseConfig  `yaml:"response,omitempty"`
	Sampling  SamplingConfig  `yaml:"sampling,omitempty"`
}

// ToolOverride customizes a specific tool's behavior
type ToolOverride struct {
	Schema       map[string]interface{} `yaml:"schema,omitempty"`
	Defaults     map[string]interface{} `yaml:"defaults,omitempty"`
	Restrictions []string               `yaml:"restrictions,omitempty"`
}

// Spell represents a switchable agent personality
type Spell struct {
	// Metadata from system.md frontmatter
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Author      string `yaml:"author,omitempty"`
	Version     string `yaml:"version,omitempty"`

	// System prompt content (body of system.md)
	SystemPrompt string `yaml:"-"`

	// Configuration from config.yaml
	Config SpellConfig `yaml:"-"`

	// Tool overrides from tools/*.yaml
	ToolOverrides map[string]ToolOverride `yaml:"-"`

	// Runtime metadata
	Mode     LayerMode `yaml:"-"` // Effective mode (from config or override)
	Source   string    `yaml:"-"` // builtin, user, project
	FilePath string    `yaml:"-"` // Path to spell directory
}

// Validate checks if a spell has required fields
func (s *Spell) Validate() error {
	if s.Name == "" {
		return fmt.Errorf("spell missing required 'name' field")
	}
	if s.Description == "" {
		return fmt.Errorf("spell missing required 'description' field")
	}
	return nil
}

// String returns a formatted representation for display
func (s *Spell) String() string {
	return fmt.Sprintf("Spell{name=%s, source=%s, mode=%s}", s.Name, s.Source, s.Mode)
}
