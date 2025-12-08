package openai_test

import (
	"testing"

	"github.com/2389-research/hex/internal/providers"
	"github.com/2389-research/hex/internal/providers/openai"
	"github.com/stretchr/testify/assert"
)

func TestOpenAIProviderName(t *testing.T) {
	provider := openai.NewProvider(providers.ProviderConfig{
		APIKey: "test-key",
	})

	assert.Equal(t, "openai", provider.Name())
}

func TestOpenAIValidateConfig(t *testing.T) {
	provider := openai.NewProvider(providers.ProviderConfig{})

	tests := []struct {
		name    string
		cfg     providers.ProviderConfig
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: providers.ProviderConfig{
				APIKey: "sk-proj-test",
			},
			wantErr: false,
		},
		{
			name: "missing API key",
			cfg: providers.ProviderConfig{
				APIKey: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := provider.ValidateConfig(tt.cfg)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestOpenAIMessageTranslation(t *testing.T) {
	req := &providers.MessageRequest{
		Model: "gpt-4o",
		Messages: []providers.Message{
			{Role: "user", Content: "hello"},
		},
		MaxTokens: 1000,
	}

	openaiReq := openai.TranslateRequest(req)

	assert.Equal(t, "gpt-4o", openaiReq.Model)
	assert.Len(t, openaiReq.Messages, 1)
	assert.Equal(t, "user", openaiReq.Messages[0].Role)
	assert.Equal(t, "hello", openaiReq.Messages[0].Content)
}
