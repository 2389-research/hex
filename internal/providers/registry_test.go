package providers

import (
	"testing"
	"time"
)

// mockProvider is a test implementation of Provider
type mockProvider struct {
	name            string
	tools           []string
	healthy         bool
	authenticated   bool
	initializeErr   error
	authenticateErr error
	executeResult   ToolResult
	executeErr      error
}

func (m *mockProvider) Name() string {
	return m.name
}

func (m *mockProvider) SupportedTools() []string {
	return m.tools
}

func (m *mockProvider) Initialize(_ map[string]string) error {
	return m.initializeErr
}

func (m *mockProvider) Authenticate() error {
	if m.authenticateErr != nil {
		return m.authenticateErr
	}
	m.authenticated = true
	return nil
}

func (m *mockProvider) Close() error {
	return nil
}

func (m *mockProvider) Status() ProviderStatus {
	msg := "healthy"
	if !m.healthy {
		msg = "unhealthy"
	}
	return ProviderStatus{
		Healthy:   m.healthy,
		Message:   msg,
		LastCheck: time.Now(),
	}
}

func (m *mockProvider) Capabilities() ProviderCapabilities {
	return ProviderCapabilities{
		RateLimits: map[string]int{"test_tool": 100},
		Features:   []string{"test_feature"},
		MaxResults: 500,
	}
}

func (m *mockProvider) ExecuteTool(_ string, _ map[string]interface{}) (ToolResult, error) {
	return m.executeResult, m.executeErr
}

func TestNewRegistry(t *testing.T) {
	registry := NewRegistry()
	if registry == nil {
		t.Fatal("NewRegistry returned nil")
	}
	if registry.providers == nil {
		t.Fatal("providers map not initialized")
	}
	if registry.active != "" {
		t.Fatal("active provider should be empty initially")
	}
}

func TestRegisterProvider(t *testing.T) {
	registry := NewRegistry()
	provider := &mockProvider{
		name:  "test",
		tools: []string{"tool1", "tool2"},
	}

	err := registry.Register(provider)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// Verify provider was registered
	if !registry.HasProvider("test") {
		t.Fatal("provider not registered")
	}

	// Verify it was set as active (first provider)
	if registry.GetActiveName() != "test" {
		t.Fatal("first provider should be set as active")
	}
}

func TestRegisterDuplicateProvider(t *testing.T) {
	registry := NewRegistry()
	provider1 := &mockProvider{name: "test"}
	provider2 := &mockProvider{name: "test"}

	_ = registry.Register(provider1)
	err := registry.Register(provider2)

	if err == nil {
		t.Fatal("expected error when registering duplicate provider")
	}
}

func TestRegisterEmptyName(t *testing.T) {
	registry := NewRegistry()
	provider := &mockProvider{name: ""}

	err := registry.Register(provider)
	if err == nil {
		t.Fatal("expected error when registering provider with empty name")
	}
}

func TestSetActive(t *testing.T) {
	registry := NewRegistry()
	provider1 := &mockProvider{name: "provider1"}
	provider2 := &mockProvider{name: "provider2"}

	_ = registry.Register(provider1)
	_ = registry.Register(provider2)

	// provider1 should be active initially
	if registry.GetActiveName() != "provider1" {
		t.Fatal("first provider should be active")
	}

	// Set provider2 as active
	err := registry.SetActive("provider2")
	if err != nil {
		t.Fatalf("SetActive failed: %v", err)
	}

	if registry.GetActiveName() != "provider2" {
		t.Fatal("active provider not updated")
	}
}

func TestSetActiveNonexistent(t *testing.T) {
	registry := NewRegistry()
	err := registry.SetActive("nonexistent")
	if err == nil {
		t.Fatal("expected error when setting nonexistent provider as active")
	}
}

func TestGetActive(t *testing.T) {
	registry := NewRegistry()
	provider := &mockProvider{name: "test"}
	_ = registry.Register(provider)

	active, err := registry.GetActive()
	if err != nil {
		t.Fatalf("GetActive failed: %v", err)
	}

	if active.Name() != "test" {
		t.Fatalf("expected 'test', got '%s'", active.Name())
	}
}

func TestGetActiveEmpty(t *testing.T) {
	registry := NewRegistry()
	_, err := registry.GetActive()
	if err == nil {
		t.Fatal("expected error when no provider is active")
	}
}

func TestGet(t *testing.T) {
	registry := NewRegistry()
	provider := &mockProvider{name: "test"}
	_ = registry.Register(provider)

	retrieved, err := registry.Get("test")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if retrieved.Name() != "test" {
		t.Fatalf("expected 'test', got '%s'", retrieved.Name())
	}
}

func TestGetNonexistent(t *testing.T) {
	registry := NewRegistry()
	_, err := registry.Get("nonexistent")
	if err == nil {
		t.Fatal("expected error when getting nonexistent provider")
	}
}

func TestList(t *testing.T) {
	registry := NewRegistry()
	provider1 := &mockProvider{
		name:    "provider1",
		tools:   []string{"tool1"},
		healthy: true,
	}
	provider2 := &mockProvider{
		name:    "provider2",
		tools:   []string{"tool2"},
		healthy: false,
	}

	_ = registry.Register(provider1)
	_ = registry.Register(provider2)

	infos := registry.List()
	if len(infos) != 2 {
		t.Fatalf("expected 2 providers, got %d", len(infos))
	}

	// Verify sorting (alphabetical)
	if infos[0].Name != "provider1" {
		t.Fatalf("expected first provider to be 'provider1', got '%s'", infos[0].Name)
	}

	// Verify active marker
	if !infos[0].Active {
		t.Fatal("first provider should be marked as active")
	}
	if infos[1].Active {
		t.Fatal("second provider should not be marked as active")
	}
}

func TestListNames(t *testing.T) {
	registry := NewRegistry()
	_ = registry.Register(&mockProvider{name: "beta"})
	_ = registry.Register(&mockProvider{name: "alpha"})
	_ = registry.Register(&mockProvider{name: "gamma"})

	names := registry.ListNames()
	if len(names) != 3 {
		t.Fatalf("expected 3 names, got %d", len(names))
	}

	// Verify alphabetical sorting
	expected := []string{"alpha", "beta", "gamma"}
	for i, name := range names {
		if name != expected[i] {
			t.Fatalf("expected names[%d] = '%s', got '%s'", i, expected[i], name)
		}
	}
}

func TestExecuteTool(t *testing.T) {
	registry := NewRegistry()
	provider := &mockProvider{
		name:    "test",
		tools:   []string{"test_tool"},
		healthy: true,
		executeResult: ToolResult{
			Success: true,
			Data:    "test data",
		},
	}
	_ = registry.Register(provider)

	result, err := registry.ExecuteTool("test_tool", nil)
	if err != nil {
		t.Fatalf("ExecuteTool failed: %v", err)
	}

	if !result.Success {
		t.Fatal("expected successful result")
	}
	if result.Data != "test data" {
		t.Fatalf("expected 'test data', got '%v'", result.Data)
	}
}

func TestExecuteToolNoActive(t *testing.T) {
	registry := NewRegistry()
	_, err := registry.ExecuteTool("test_tool", nil)
	if err == nil {
		t.Fatal("expected error when no active provider")
	}
}

func TestExecuteToolUnsupported(t *testing.T) {
	registry := NewRegistry()
	provider := &mockProvider{
		name:    "test",
		tools:   []string{"other_tool"},
		healthy: true,
	}
	_ = registry.Register(provider)

	_, err := registry.ExecuteTool("unsupported_tool", nil)
	if err == nil {
		t.Fatal("expected error when tool not supported")
	}
}

func TestExecuteToolUnhealthyProvider(t *testing.T) {
	registry := NewRegistry()
	provider := &mockProvider{
		name:    "test",
		tools:   []string{"test_tool"},
		healthy: false,
	}
	_ = registry.Register(provider)

	_, err := registry.ExecuteTool("test_tool", nil)
	if err == nil {
		t.Fatal("expected error when provider is unhealthy")
	}
}

func TestCount(t *testing.T) {
	registry := NewRegistry()
	if registry.Count() != 0 {
		t.Fatal("expected 0 providers initially")
	}

	_ = registry.Register(&mockProvider{name: "test1"})
	if registry.Count() != 1 {
		t.Fatal("expected 1 provider after registration")
	}

	_ = registry.Register(&mockProvider{name: "test2"})
	if registry.Count() != 2 {
		t.Fatal("expected 2 providers after second registration")
	}
}

func TestHasProvider(t *testing.T) {
	registry := NewRegistry()
	_ = registry.Register(&mockProvider{name: "test"})

	if !registry.HasProvider("test") {
		t.Fatal("expected HasProvider to return true for registered provider")
	}

	if registry.HasProvider("nonexistent") {
		t.Fatal("expected HasProvider to return false for nonexistent provider")
	}
}

func TestConcurrentAccess(_ *testing.T) {
	registry := NewRegistry()

	// Register initial provider
	_ = registry.Register(&mockProvider{name: "test"})

	// Concurrent reads
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			_ = registry.List()
			_ = registry.GetActiveName()
			_, _ = registry.GetActive()
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}
