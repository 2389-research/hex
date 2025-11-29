// ABOUTME: Template types for YAML-based session templates
// ABOUTME: Defines Template structure with system prompts, initial messages, and tool configurations

package templates

import "time"

// Message represents a template message
type Message struct {
	Role    string `yaml:"role"`    // "user", "assistant", or "system"
	Content string `yaml:"content"` // Message text
}

// Template represents a session template loaded from YAML
type Template struct {
	Name            string    `yaml:"name"`                       // Template name
	Description     string    `yaml:"description"`                // Human-readable description
	SystemPrompt    string    `yaml:"system_prompt,omitempty"`    // System prompt to use
	InitialMessages []Message `yaml:"initial_messages,omitempty"` // Pre-populated messages
	ToolsEnabled    []string  `yaml:"tools_enabled,omitempty"`    // Tools to enable (empty = all)
	Model           string    `yaml:"model,omitempty"`            // Preferred model
	MaxTokens       int       `yaml:"max_tokens,omitempty"`       // Max tokens for responses
	CreatedAt       time.Time `yaml:"-"`                          // When template was created (not in YAML)
}

// Validate checks if a template is valid
func (t *Template) Validate() error {
	if t.Name == "" {
		return &ValidationError{Field: "name", Message: "template name is required"}
	}

	// Validate messages
	for i, msg := range t.InitialMessages {
		if msg.Role != "user" && msg.Role != "assistant" && msg.Role != "system" {
			return &ValidationError{
				Field:   "initial_messages",
				Message: "invalid role at index " + string(rune(i)) + ": must be user, assistant, or system",
			}
		}
		if msg.Content == "" {
			return &ValidationError{
				Field:   "initial_messages",
				Message: "message content cannot be empty at index " + string(rune(i)),
			}
		}
	}

	return nil
}

// ValidationError represents a template validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return "validation error: " + e.Field + ": " + e.Message
}
