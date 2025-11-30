// ABOUTME: Tests for quick actions menu system
// ABOUTME: Validates action registration, fuzzy search, and execution
package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewQuickActionsRegistry(t *testing.T) {
	registry := NewQuickActionsRegistry()
	assert.NotNil(t, registry)
	assert.NotNil(t, registry.actions)

	// Should have built-in actions
	assert.Greater(t, len(registry.actions), 0, "Should have built-in actions registered")
}

func TestRegisterAction(t *testing.T) {
	registry := NewQuickActionsRegistry()

	// Register a custom action
	handler := func(_ string) error {
		return nil
	}

	err := registry.RegisterAction("test", "Test action", "test <arg>", handler)
	require.NoError(t, err)

	// Check action was registered
	action, err := registry.GetAction("test")
	require.NoError(t, err)
	assert.Equal(t, "test", action.Name)
	assert.Equal(t, "Test action", action.Description)
	assert.Equal(t, "test <arg>", action.Usage)
	assert.NotNil(t, action.Handler)
}

func TestRegisterActionDuplicate(t *testing.T) {
	registry := NewQuickActionsRegistry()

	handler := func(_ string) error {
		return nil
	}

	// Register first time - should succeed
	err := registry.RegisterAction("test", "Test action", "test", handler)
	require.NoError(t, err)

	// Register again - should fail
	err = registry.RegisterAction("test", "Different description", "test", handler)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")
}

func TestGetActionNotFound(t *testing.T) {
	registry := NewQuickActionsRegistry()

	_, err := registry.GetAction("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestListActions(t *testing.T) {
	registry := NewQuickActionsRegistry()

	actions := registry.ListActions()
	assert.Greater(t, len(actions), 0, "Should return built-in actions")

	// Verify actions have required fields
	for _, action := range actions {
		assert.NotEmpty(t, action.Name)
		assert.NotEmpty(t, action.Description)
		assert.NotEmpty(t, action.Usage)
		assert.NotNil(t, action.Handler)
	}
}

func TestFuzzySearchActions(t *testing.T) {
	registry := NewQuickActionsRegistry()

	tests := []struct {
		name     string
		query    string
		minMatch int // Minimum expected matches
	}{
		{
			name:     "exact match",
			query:    "read",
			minMatch: 1,
		},
		{
			name:     "partial match",
			query:    "gre",
			minMatch: 1,
		},
		{
			name:     "fuzzy match",
			query:    "wb",
			minMatch: 1, // Should match "web"
		},
		{
			name:     "empty query returns all",
			query:    "",
			minMatch: 6, // All built-in actions
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := registry.FuzzySearch(tt.query)
			assert.GreaterOrEqual(t, len(results), tt.minMatch,
				"Query '%s' should match at least %d actions", tt.query, tt.minMatch)
		})
	}
}

func TestFuzzySearchOrdering(t *testing.T) {
	registry := NewQuickActionsRegistry()

	// Search for "read"
	results := registry.FuzzySearch("read")

	// First result should be exact match
	require.Greater(t, len(results), 0)
	assert.Equal(t, "read", results[0].Name)
}

func TestParseActionCommand(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedCmd string
		expectedArg string
	}{
		{
			name:        "command only",
			input:       "read",
			expectedCmd: "read",
			expectedArg: "",
		},
		{
			name:        "command with argument",
			input:       "read /path/to/file",
			expectedCmd: "read",
			expectedArg: "/path/to/file",
		},
		{
			name:        "command with multiple arguments",
			input:       "grep pattern *.go",
			expectedCmd: "grep",
			expectedArg: "pattern *.go",
		},
		{
			name:        "command with leading/trailing spaces",
			input:       "  save  ",
			expectedCmd: "save",
			expectedArg: "",
		},
		{
			name:        "empty input",
			input:       "",
			expectedCmd: "",
			expectedArg: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, arg := ParseActionCommand(tt.input)
			assert.Equal(t, tt.expectedCmd, cmd)
			assert.Equal(t, tt.expectedArg, arg)
		})
	}
}

func TestBuiltInActions(t *testing.T) {
	registry := NewQuickActionsRegistry()

	// Test that all built-in actions are registered
	builtins := []string{"read", "grep", "web", "attach", "save", "export"}

	for _, name := range builtins {
		t.Run(name, func(t *testing.T) {
			action, err := registry.GetAction(name)
			require.NoError(t, err, "Built-in action %s should be registered", name)
			assert.Equal(t, name, action.Name)
			assert.NotEmpty(t, action.Description)
			assert.NotNil(t, action.Handler)
		})
	}
}

func TestExecuteAction(t *testing.T) {
	registry := NewQuickActionsRegistry()

	// Register a test action that records execution
	executed := false
	var receivedArgs string

	handler := func(args string) error {
		executed = true
		receivedArgs = args
		return nil
	}

	err := registry.RegisterAction("test", "Test", "test", handler)
	require.NoError(t, err)

	// Execute the action
	err = registry.Execute("test", "some args")
	require.NoError(t, err)

	assert.True(t, executed, "Handler should have been called")
	assert.Equal(t, "some args", receivedArgs)
}

func TestExecuteActionNotFound(t *testing.T) {
	registry := NewQuickActionsRegistry()

	err := registry.Execute("nonexistent", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestActionHandlerError(t *testing.T) {
	registry := NewQuickActionsRegistry()

	// Register action that returns error
	handler := func(_ string) error {
		return assert.AnError
	}

	err := registry.RegisterAction("test", "Test", "test", handler)
	require.NoError(t, err)

	// Execute should return the handler error
	err = registry.Execute("test", "")
	assert.Error(t, err)
}

func TestFuzzySearchCaseInsensitive(t *testing.T) {
	registry := NewQuickActionsRegistry()

	// Search with different cases
	lowerResults := registry.FuzzySearch("read")
	upperResults := registry.FuzzySearch("READ")
	mixedResults := registry.FuzzySearch("ReAd")

	// All should return same results
	assert.Equal(t, len(lowerResults), len(upperResults))
	assert.Equal(t, len(lowerResults), len(mixedResults))
}

func TestFuzzySearchNoMatch(t *testing.T) {
	registry := NewQuickActionsRegistry()

	// Search for something that definitely won't match
	results := registry.FuzzySearch("zzzzzzzzz")

	// Should return empty list, not error
	assert.NotNil(t, results)
	assert.Len(t, results, 0)
}
