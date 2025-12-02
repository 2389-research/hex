// ABOUTME: Provider registry for managing productivity service providers
// ABOUTME: Thread-safe registration, activation, and routing of provider instances

// Package providers defines the provider interface and registry for productivity tools
package providers

import (
	"fmt"
	"sort"
	"sync"
)

// Registry manages available productivity providers
type Registry struct {
	providers map[string]Provider // name -> provider instance
	active    string              // currently active provider name
	mu        sync.RWMutex
}

// NewRegistry creates a new provider registry
func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]Provider),
		active:    "",
	}
}

// Register adds a provider to the registry
func (r *Registry) Register(p Provider) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := p.Name()
	if name == "" {
		return fmt.Errorf("provider name cannot be empty")
	}

	if _, exists := r.providers[name]; exists {
		return fmt.Errorf("provider %s already registered", name)
	}

	r.providers[name] = p

	// If this is the first provider, make it active
	if r.active == "" {
		r.active = name
	}

	return nil
}

// SetActive sets the currently active provider
func (r *Registry) SetActive(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.providers[name]; !exists {
		return fmt.Errorf("provider %s not found", name)
	}

	r.active = name
	return nil
}

// GetActive returns the currently active provider
func (r *Registry) GetActive() (Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.active == "" {
		return nil, fmt.Errorf("no active provider set")
	}

	p, exists := r.providers[r.active]
	if !exists {
		return nil, fmt.Errorf("active provider %s not found", r.active)
	}

	return p, nil
}

// GetActiveName returns the name of the currently active provider
func (r *Registry) GetActiveName() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.active
}

// Get retrieves a provider by name
func (r *Registry) Get(name string) (Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, exists := r.providers[name]
	if !exists {
		return nil, fmt.Errorf("provider %s not found", name)
	}

	return p, nil
}

// List returns information about all registered providers
func (r *Registry) List() []ProviderInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	infos := make([]ProviderInfo, 0, len(r.providers))
	for name, provider := range r.providers {
		status := provider.Status()
		info := ProviderInfo{
			Name:           name,
			Authenticated:  status.Healthy, // TODO: separate auth check
			Active:         name == r.active,
			SupportedTools: provider.SupportedTools(),
		}
		infos = append(infos, info)
	}

	// Sort by name for consistent output
	sort.Slice(infos, func(i, j int) bool {
		return infos[i].Name < infos[j].Name
	})

	return infos
}

// ListNames returns all registered provider names sorted alphabetically
func (r *Registry) ListNames() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// ExecuteTool routes a tool call to the active provider
func (r *Registry) ExecuteTool(toolName string, params map[string]interface{}) (ToolResult, error) {
	activeProvider, err := r.GetActive()
	if err != nil {
		return ToolResult{
			Success: false,
			Error:   fmt.Sprintf("no active provider: %v", err),
		}, err
	}

	// Check if provider supports this tool
	supportedTools := activeProvider.SupportedTools()
	supported := false
	for _, t := range supportedTools {
		if t == toolName {
			supported = true
			break
		}
	}

	if !supported {
		err := fmt.Errorf("provider %s does not support tool %s",
			activeProvider.Name(), toolName)
		return ToolResult{
			Success: false,
			Error:   err.Error(),
		}, err
	}

	// Check provider health
	status := activeProvider.Status()
	if !status.Healthy {
		err := fmt.Errorf("provider %s is unhealthy: %s",
			activeProvider.Name(), status.Message)
		return ToolResult{
			Success: false,
			Error:   err.Error(),
		}, err
	}

	// Execute tool
	return activeProvider.ExecuteTool(toolName, params)
}

// Count returns the number of registered providers
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.providers)
}

// HasProvider checks if a provider with the given name exists
func (r *Registry) HasProvider(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.providers[name]
	return exists
}
