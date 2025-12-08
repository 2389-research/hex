package openrouter_test

import (
	"context"
	"testing"

	"github.com/2389-research/hex/internal/core"
	"github.com/2389-research/hex/internal/providers"
	"github.com/2389-research/hex/internal/providers/openrouter"
	"github.com/stretchr/testify/assert"
)

func TestProviderImplementsInterface(t *testing.T) {
	config := providers.ProviderConfig{
		APIKey: "test-key",
	}
	provider := openrouter.NewProvider(config)

	// Verify it implements the Provider interface
	var _ providers.Provider = provider
}

func TestProviderName(t *testing.T) {
	provider := openrouter.NewProvider(providers.ProviderConfig{APIKey: "test"})
	assert.Equal(t, "openrouter", provider.Name())
}

func TestValidateConfig(t *testing.T) {
	provider := openrouter.NewProvider(providers.ProviderConfig{})

	tests := []struct {
		name    string
		config  providers.ProviderConfig
		wantErr bool
	}{
		{
			name:    "valid config",
			config:  providers.ProviderConfig{APIKey: "sk-test123"},
			wantErr: false,
		},
		{
			name:    "missing API key",
			config:  providers.ProviderConfig{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := provider.ValidateConfig(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCreateMessageStream_Interface(t *testing.T) {
	provider := openrouter.NewProvider(providers.ProviderConfig{
		APIKey: "test-key",
	})

	// Verify the method signature matches the interface
	req := core.MessageRequest{
		Model: "anthropic/claude-sonnet-4-5",
		Messages: []core.Message{
			{Role: "user", Content: "Hello"},
		},
	}

	// This will fail to connect, but we're just testing the signature
	ctx := context.Background()
	_, err := provider.CreateMessageStream(ctx, req)
	// We expect an error since we're using a fake API key
	// The important thing is that it compiles and returns the right types
	assert.Error(t, err) // Will fail to connect, but that's expected
}
