package core

import (
	"fmt"
	"time"
)

// Message represents a single message in a conversation
type Message struct {
	ID        string     `json:"id,omitempty"`
	Role      string     `json:"role"` // "user", "assistant", "system"
	Content   string     `json:"content"`
	ToolCalls []ToolUse  `json:"tool_calls,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
}

// Validate checks if the message is valid
func (m *Message) Validate() error {
	switch m.Role {
	case "user", "assistant", "system":
		return nil
	default:
		return fmt.Errorf("invalid role: %s", m.Role)
	}
}

// ToolUse represents a tool invocation from the API
type ToolUse struct {
	Type  string                 `json:"type"`  // "tool_use"
	ID    string                 `json:"id"`
	Name  string                 `json:"name"`
	Input map[string]interface{} `json:"input"`
}

// ToolDefinition defines a tool's schema
type ToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"input_schema"`
}

// MessageRequest is sent to the API
type MessageRequest struct {
	Model     string           `json:"model"`
	Messages  []Message        `json:"messages"`
	MaxTokens int              `json:"max_tokens"`
	Stream    bool             `json:"stream,omitempty"`
	Tools     []ToolDefinition `json:"tools,omitempty"`
	System    string           `json:"system,omitempty"`
}

// MessageResponse is received from the API
type MessageResponse struct {
	ID         string    `json:"id"`
	Type       string    `json:"type"`
	Role       string    `json:"role"`
	Content    []Content `json:"content"`
	Model      string    `json:"model"`
	StopReason string    `json:"stop_reason,omitempty"`
	Usage      Usage     `json:"usage"`
}

// Content represents a content block in the response
type Content struct {
	Type  string                 `json:"type"` // "text" or "tool_use"
	Text  string                 `json:"text,omitempty"`
	ID    string                 `json:"id,omitempty"`
	Name  string                 `json:"name,omitempty"`
	Input map[string]interface{} `json:"input,omitempty"`
}

// Usage tracks token usage
type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// StreamChunk represents a chunk in streaming response
type StreamChunk struct {
	Type         string   `json:"type"`
	Delta        *Delta   `json:"delta,omitempty"`
	Content      *Content `json:"content,omitempty"`
	ContentBlock *Content `json:"content_block,omitempty"` // For content_block_start events
	Index        int      `json:"index,omitempty"`         // Content block index
	Usage        *Usage   `json:"usage,omitempty"`
	Done         bool     `json:"-"`
}

// Delta represents incremental content in streaming
type Delta struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}
