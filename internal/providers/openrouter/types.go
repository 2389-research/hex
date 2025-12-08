// ABOUTME: OpenRouter uses OpenAI-compatible format
// ABOUTME: Only difference is model ID format (provider/model)
package openrouter

import (
	"github.com/2389-research/hex/internal/providers"
	"github.com/2389-research/hex/internal/providers/openai"
)

// TranslateRequest reuses OpenAI translation
// OpenRouter is OpenAI-compatible
func TranslateRequest(req *providers.MessageRequest) *openai.OpenAIRequest {
	return openai.TranslateRequest(req)
}
