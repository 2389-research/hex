package anthropic_test

import (
	"testing"

	"github.com/2389-research/hex/internal/providers"
	"github.com/2389-research/hex/internal/providers/anthropic"
	"github.com/stretchr/testify/assert"
)

func TestAnthropicProviderName(t *testing.T) {
	provider := anthropic.NewProvider(providers.ProviderConfig{
		APIKey: "test-key",
	})

	assert.Equal(t, "anthropic", provider.Name())
}

func TestAnthropicValidateConfig(t *testing.T) {
	provider := anthropic.NewProvider(providers.ProviderConfig{})

	tests := []struct {
		name    string
		cfg     providers.ProviderConfig
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: providers.ProviderConfig{
				APIKey: "sk-ant-test",
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
