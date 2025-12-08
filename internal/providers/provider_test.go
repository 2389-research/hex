package providers_test

import (
	"context"
	"testing"

	"github.com/2389-research/hex/internal/core"
	"github.com/2389-research/hex/internal/providers"
	"github.com/stretchr/testify/assert"
)

func TestFactoryRegistersProviders(t *testing.T) {
	factory := providers.NewFactory()

	// Should have no providers initially
	_, err := factory.GetProvider("nonexistent")
	assert.Error(t, err)
}

func TestFactoryCreatesProvider(t *testing.T) {
	factory := providers.NewFactory()

	// Register a test provider
	factory.Register("test", &mockProvider{name: "test"})

	provider, err := factory.GetProvider("test")
	assert.NoError(t, err)
	assert.Equal(t, "test", provider.Name())
}

type mockProvider struct {
	name string
}

func (m *mockProvider) CreateMessage(ctx context.Context, req core.MessageRequest) (*core.MessageResponse, error) {
	return nil, nil
}

func (m *mockProvider) CreateMessageStream(ctx context.Context, req core.MessageRequest) (<-chan *core.StreamChunk, error) {
	return nil, nil
}

func (m *mockProvider) Name() string {
	return m.name
}

func (m *mockProvider) ValidateConfig(cfg providers.ProviderConfig) error {
	return nil
}
