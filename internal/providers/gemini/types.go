// ABOUTME: Type translation between hex universal format and Gemini API format
// ABOUTME: Handles contents/parts structure and function declarations
package gemini

import (
	"github.com/2389-research/hex/internal/providers"
)

// GeminiRequest is the Gemini-specific request format
type GeminiRequest struct {
	Contents         []GeminiContent   `json:"contents"`
	GenerationConfig *GenerationConfig `json:"generationConfig,omitempty"`
	Tools            []GeminiTool      `json:"tools,omitempty"`
}

type GeminiContent struct {
	Role  string       `json:"role"`
	Parts []GeminiPart `json:"parts"`
}

type GeminiPart struct {
	Text string `json:"text"`
}

type GenerationConfig struct {
	MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
	Temperature     float64 `json:"temperature,omitempty"`
}

type GeminiTool struct {
	FunctionDeclarations []FunctionDeclaration `json:"functionDeclarations"`
}

type FunctionDeclaration struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// TranslateRequest converts universal format to Gemini format
func TranslateRequest(req *providers.MessageRequest) *GeminiRequest {
	contents := make([]GeminiContent, len(req.Messages))
	for i, msg := range req.Messages {
		contents[i] = GeminiContent{
			Role: msg.Role,
			Parts: []GeminiPart{
				{Text: msg.Content},
			},
		}
	}

	var tools []GeminiTool
	if len(req.Tools) > 0 {
		functionDecls := make([]FunctionDeclaration, len(req.Tools))
		for i, tool := range req.Tools {
			functionDecls[i] = FunctionDeclaration{
				Name:        tool.Name,
				Description: tool.Description,
				Parameters:  tool.InputSchema,
			}
		}
		tools = []GeminiTool{{FunctionDeclarations: functionDecls}}
	}

	return &GeminiRequest{
		Contents: contents,
		GenerationConfig: &GenerationConfig{
			MaxOutputTokens: req.MaxTokens,
			Temperature:     req.Temperature,
		},
		Tools: tools,
	}
}
