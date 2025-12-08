// ABOUTME: Type translation between hex universal format and Anthropic API format
// ABOUTME: Handles message structure, tool definitions, and response parsing
package anthropic

import (
	"github.com/2389-research/hex/internal/providers"
)

// anthropicRequest is the Anthropic-specific request format
type anthropicRequest struct {
	Model       string             `json:"model"`
	Messages    []anthropicMessage `json:"messages"`
	MaxTokens   int                `json:"max_tokens"`
	Stream      bool               `json:"stream"`
	Temperature float64            `json:"temperature,omitempty"`
	Tools       []anthropicTool    `json:"tools,omitempty"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"input_schema"`
}

// translateRequest converts universal format to Anthropic format
func translateRequest(req *providers.MessageRequest) *anthropicRequest {
	messages := make([]anthropicMessage, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = anthropicMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	tools := make([]anthropicTool, len(req.Tools))
	for i, tool := range req.Tools {
		tools[i] = anthropicTool{
			Name:        tool.Name,
			Description: tool.Description,
			InputSchema: tool.InputSchema,
		}
	}

	return &anthropicRequest{
		Model:       req.Model,
		Messages:    messages,
		MaxTokens:   req.MaxTokens,
		Stream:      req.Stream,
		Temperature: req.Temperature,
		Tools:       tools,
	}
}
