// ABOUTME: Factory creates and manages provider instances
// ABOUTME: Provides registry for available providers and instantiation logic
package providers

import (
	"fmt"
	"sync"
)

// Factory creates provider instances
type Factory struct {
	mu        sync.RWMutex
	providers map[string]Provider
}

// NewFactory creates a new provider factory
func NewFactory() *Factory {
	return &Factory{
		providers: make(map[string]Provider),
	}
}

// Register adds a provider to the factory
func (f *Factory) Register(name string, provider Provider) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.providers[name] = provider
}

// GetProvider returns a provider by name
func (f *Factory) GetProvider(name string) (Provider, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	provider, ok := f.providers[name]
	if !ok {
		return nil, fmt.Errorf("provider not found: %s", name)
	}
	return provider, nil
}

// ListProviders returns all registered provider names
func (f *Factory) ListProviders() []string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	names := make([]string, 0, len(f.providers))
	for name := range f.providers {
		names = append(names, name)
	}
	return names
}
