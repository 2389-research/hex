package openrouter_test

import (
	"testing"

	"github.com/2389-research/hex/internal/providers"
	"github.com/2389-research/hex/internal/providers/openrouter"
	"github.com/stretchr/testify/assert"
)

func TestOpenRouterProviderName(t *testing.T) {
	provider := openrouter.NewProvider(providers.ProviderConfig{
		APIKey: "test-key",
	})

	assert.Equal(t, "openrouter", provider.Name())
}

func TestOpenRouterModelIDHandling(t *testing.T) {
	// OpenRouter uses provider/model format
	req := &providers.MessageRequest{
		Model: "anthropic/claude-sonnet-4-5",
	}

	openaiReq := openrouter.TranslateRequest(req)
	assert.Equal(t, "anthropic/claude-sonnet-4-5", openaiReq.Model)
}
