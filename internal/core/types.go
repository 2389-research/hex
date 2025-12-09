package core

import (
	"encoding/json"
	"fmt"
	"time"
)

// ContentBlock represents a single block of content (text, image, tool_use, or tool_result)
type ContentBlock struct {
	Type      string                 `json:"type"`                  // "text", "image", "tool_use", or "tool_result"
	Text      string                 `json:"text,omitempty"`        // For text blocks
	Source    *ImageSource           `json:"source,omitempty"`      // For image blocks
	ID        string                 `json:"id,omitempty"`          // For tool_use and tool_result blocks
	Name      string                 `json:"name,omitempty"`        // For tool_use blocks
	Input     map[string]interface{} `json:"input,omitempty"`       // For tool_use blocks
	ToolUseID string                 `json:"tool_use_id,omitempty"` // For tool_result blocks (ID field above is for tool_use)
	Content   string                 `json:"content,omitempty"`     // For tool_result blocks
}

// NewTextBlock creates a text content block
func NewTextBlock(text string) ContentBlock {
	return ContentBlock{
		Type: "text",
		Text: text,
	}
}

// NewImageBlock creates an image content block
func NewImageBlock(source *ImageSource) ContentBlock {
	return ContentBlock{
		Type:   "image",
		Source: source,
	}
}

// NewToolResultBlock creates a tool_result content block
func NewToolResultBlock(toolUseID string, content string) ContentBlock {
	return ContentBlock{
		Type:      "tool_result",
		ToolUseID: toolUseID,
		Content:   content,
	}
}

// Message represents a single message in a conversation
type Message struct {
	ID           string         `json:"id,omitempty"`
	Role         string         `json:"role"` // "user", "assistant", "system"
	Content      string         `json:"-"`    // Internal field, not directly serialized
	ContentBlock []ContentBlock `json:"-"`    // Internal field, not directly serialized
	ToolCalls    []ToolUse      `json:"tool_calls,omitempty"`
	CreatedAt    *time.Time     `json:"created_at,omitempty"`
}

// MarshalJSON implements custom JSON marshalling for Message
// This handles the Claude API requirement that 'content' can be either a string or an array
func (m Message) MarshalJSON() ([]byte, error) {
	type Alias Message
	aux := struct {
		Content interface{} `json:"content"` // Required field - no omitempty
		*Alias
	}{
		Alias: (*Alias)(&m),
	}

	// If ContentBlock is set, use it as an array
	// Otherwise, use the string Content (even if empty - API requires content field)
	if len(m.ContentBlock) > 0 {
		aux.Content = m.ContentBlock
	} else {
		aux.Content = m.Content // Will be "" if not set, but field is required
	}

	return json.Marshal(aux)
}

// UnmarshalJSON implements custom JSON unmarshalling for Message
func (m *Message) UnmarshalJSON(data []byte) error {
	type Alias Message
	aux := &struct {
		Content json.RawMessage `json:"content,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(m),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Try to unmarshal content as array first
	var blocks []ContentBlock
	if err := json.Unmarshal(aux.Content, &blocks); err == nil {
		m.ContentBlock = blocks
		return nil
	}

	// Otherwise, treat as string
	var str string
	if err := json.Unmarshal(aux.Content, &str); err == nil {
		m.Content = str
		return nil
	}

	return nil
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
	Type  string                 `json:"type"` // "tool_use"
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
	InputTokens      int `json:"input_tokens"`
	OutputTokens     int `json:"output_tokens"`
	CacheReadTokens  int `json:"cache_read_input_tokens,omitempty"`
	CacheWriteTokens int `json:"cache_creation_input_tokens,omitempty"`
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
	Type         string `json:"type"`
	Text         string `json:"text,omitempty"`
	PartialJSON  string `json:"partial_json,omitempty"`  // For input_json_delta
	StopReason   string `json:"stop_reason,omitempty"`   // For message_delta
	StopSequence string `json:"stop_sequence,omitempty"` // For message_delta
}
