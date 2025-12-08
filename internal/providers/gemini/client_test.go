package gemini_test

import (
	"testing"

	"github.com/2389-research/hex/internal/providers"
	"github.com/2389-research/hex/internal/providers/gemini"
	"github.com/stretchr/testify/assert"
)

func TestGeminiProviderName(t *testing.T) {
	provider := gemini.NewProvider(providers.ProviderConfig{
		APIKey: "test-key",
	})

	assert.Equal(t, "gemini", provider.Name())
}

func TestGeminiValidateConfig(t *testing.T) {
	provider := gemini.NewProvider(providers.ProviderConfig{})

	err := provider.ValidateConfig(providers.ProviderConfig{
		APIKey: "",
	})
	assert.Error(t, err)

	err = provider.ValidateConfig(providers.ProviderConfig{
		APIKey: "AIza-test",
	})
	assert.NoError(t, err)
}
