// ABOUTME: Tests for autocomplete system
// ABOUTME: Validates fuzzy matching, providers, and navigation

package ui_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/harper/clem/internal/ui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAutocomplete_NewAutocomplete(t *testing.T) {
	ac := ui.NewAutocomplete()

	assert.NotNil(t, ac)
	assert.False(t, ac.IsActive())
	assert.Equal(t, 0, len(ac.GetCompletions()))
}

func TestAutocomplete_ShowHide(t *testing.T) {
	ac := ui.NewAutocomplete()

	// Initially inactive
	assert.False(t, ac.IsActive())

	// Show with tool provider
	provider := ui.NewToolProvider([]string{"read", "write", "execute"})
	ac.RegisterProvider("test", provider)
	ac.Show("r", "test")

	assert.True(t, ac.IsActive())
	assert.Greater(t, len(ac.GetCompletions()), 0)

	// Hide
	ac.Hide()
	assert.False(t, ac.IsActive())
	assert.Equal(t, 0, len(ac.GetCompletions()))
}

func TestAutocomplete_Navigation(t *testing.T) {
	ac := ui.NewAutocomplete()
	provider := ui.NewToolProvider([]string{"alpha", "beta", "gamma"})
	ac.RegisterProvider("test", provider)
	ac.Show("", "test")

	// Should start at index 0
	assert.Equal(t, 0, ac.GetSelectedIndex())

	// Next
	ac.Next()
	assert.Equal(t, 1, ac.GetSelectedIndex())

	// Next again
	ac.Next()
	assert.Equal(t, 2, ac.GetSelectedIndex())

	// Next wraps around
	ac.Next()
	assert.Equal(t, 0, ac.GetSelectedIndex())

	// Previous
	ac.Previous()
	assert.Equal(t, 2, ac.GetSelectedIndex())

	// Previous again
	ac.Previous()
	assert.Equal(t, 1, ac.GetSelectedIndex())
}

func TestAutocomplete_GetSelected(t *testing.T) {
	ac := ui.NewAutocomplete()
	provider := ui.NewToolProvider([]string{"alpha", "beta", "gamma"})
	ac.RegisterProvider("test", provider)
	ac.Show("", "test")

	// Get first item
	selected := ac.GetSelected()
	require.NotNil(t, selected)
	assert.Equal(t, "alpha", selected.Value)

	// Navigate and get second item
	ac.Next()
	selected = ac.GetSelected()
	require.NotNil(t, selected)
	assert.Equal(t, "beta", selected.Value)
}

func TestAutocomplete_Update(t *testing.T) {
	ac := ui.NewAutocomplete()
	provider := ui.NewToolProvider([]string{"read", "write", "execute", "grep"})
	ac.RegisterProvider("test", provider)
	ac.Show("", "test")

	// Should have all completions initially
	assert.Equal(t, 4, len(ac.GetCompletions()))

	// Update with filter
	ac.Update("re")
	completions := ac.GetCompletions()

	// Should only have "read" (fuzzy match)
	assert.Greater(t, len(completions), 0)
	found := false
	for _, c := range completions {
		if c.Value == "read" {
			found = true
			break
		}
	}
	assert.True(t, found, "Expected 'read' in fuzzy matches for 're'")

	// Update with no matches
	ac.Update("xyz123")
	assert.False(t, ac.IsActive(), "Should hide when no matches")
}

func TestToolProvider_GetCompletions(t *testing.T) {
	provider := ui.NewToolProvider([]string{"read_file", "write_file", "grep", "bash"})

	tests := []struct {
		name          string
		input         string
		expectMatches []string
	}{
		{
			name:          "empty input returns all",
			input:         "",
			expectMatches: []string{"read_file", "write_file", "grep", "bash"},
		},
		{
			name:          "exact match",
			input:         "grep",
			expectMatches: []string{"grep"},
		},
		{
			name:          "prefix match",
			input:         "read",
			expectMatches: []string{"read_file"},
		},
		{
			name:          "fuzzy match",
			input:         "rf",
			expectMatches: []string{"read_file"},
		},
		{
			name:          "multiple fuzzy matches",
			input:         "file",
			expectMatches: []string{"read_file", "write_file"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			completions := provider.GetCompletions(tt.input)

			// Check that all expected matches are present
			for _, expected := range tt.expectMatches {
				found := false
				for _, c := range completions {
					if c.Value == expected {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected to find %s in completions", expected)
			}
		})
	}
}

func TestToolProvider_SetTools(t *testing.T) {
	provider := ui.NewToolProvider(nil)

	// Initially empty
	completions := provider.GetCompletions("")
	assert.Equal(t, 0, len(completions))

	// Set tools
	provider.SetTools([]string{"alpha", "beta"})
	completions = provider.GetCompletions("")
	assert.Equal(t, 2, len(completions))
}

func TestFileProvider_GetCompletions(t *testing.T) {
	// Create a temporary directory with test files
	tmpDir := t.TempDir()

	testFiles := []string{
		"alpha.txt",
		"beta.go",
		"gamma.md",
		".hidden",
	}

	for _, file := range testFiles {
		f, err := os.Create(filepath.Join(tmpDir, file))
		require.NoError(t, err)
		f.Close()
	}

	// Create a subdirectory
	subDir := filepath.Join(tmpDir, "subdir")
	err := os.Mkdir(subDir, 0755)
	require.NoError(t, err)

	provider := ui.NewFileProvider()
	provider.SetBasePath(tmpDir)

	t.Run("empty input returns all visible files", func(t *testing.T) {
		completions := provider.GetCompletions("")

		// Should have 4 items: 3 files + 1 dir (hidden file excluded)
		assert.Equal(t, 4, len(completions))

		// Check that subdirectory has trailing slash
		found := false
		for _, c := range completions {
			if c.Display == "subdir/" {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected subdir/ in completions")
	})

	t.Run("fuzzy match on filename", func(t *testing.T) {
		completions := provider.GetCompletions("alp")

		assert.Greater(t, len(completions), 0)
		assert.Equal(t, "alpha.txt", filepath.Base(completions[0].Value))
	})

	t.Run("hidden files shown when explicitly requested", func(t *testing.T) {
		completions := provider.GetCompletions(".")

		found := false
		for _, c := range completions {
			if filepath.Base(c.Value) == ".hidden" {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected .hidden file when searching with .")
	})
}

func TestHistoryProvider_GetCompletions(t *testing.T) {
	provider := ui.NewHistoryProvider()

	// Add some history
	provider.AddToHistory("first command")
	provider.AddToHistory("second command")
	provider.AddToHistory("third command")

	t.Run("empty input returns recent history", func(t *testing.T) {
		completions := provider.GetCompletions("")

		assert.Equal(t, 3, len(completions))
		// Most recent should be first
		assert.Equal(t, "third command", completions[0].Value)
	})

	t.Run("fuzzy match on history", func(t *testing.T) {
		completions := provider.GetCompletions("first")

		assert.Greater(t, len(completions), 0)
		assert.Equal(t, "first command", completions[0].Value)
	})

	t.Run("duplicate handling", func(t *testing.T) {
		provider.AddToHistory("first command")
		completions := provider.GetCompletions("")

		// Should still have 3 items, with "first command" at the front
		assert.Equal(t, 3, len(completions))
		assert.Equal(t, "first command", completions[0].Value)
	})

	t.Run("history size limit", func(t *testing.T) {
		// Add 101 commands
		for i := 0; i < 101; i++ {
			provider.AddToHistory("command " + string(rune(i)))
		}

		completions := provider.GetCompletions("")
		// Should be limited to 10 in display, but history stores 100
		assert.LessOrEqual(t, len(completions), 10)
	})
}

func TestDetectProvider(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "tool command",
			input:    ":tool read",
			expected: "tool",
		},
		{
			name:     "file path with slash",
			input:    "/home/user/file.txt",
			expected: "file",
		},
		{
			name:     "relative path",
			input:    "./file.txt",
			expected: "file",
		},
		{
			name:     "home directory",
			input:    "~/documents",
			expected: "file",
		},
		{
			name:     "default to history",
			input:    "some command",
			expected: "history",
		},
		{
			name:     "empty defaults to history",
			input:    "",
			expected: "history",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ui.DetectProvider(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAutocomplete_MaxCompletions(t *testing.T) {
	ac := ui.NewAutocomplete()

	// Create provider with 20 items
	tools := make([]string, 20)
	for i := 0; i < 20; i++ {
		tools[i] = string(rune('a' + i))
	}

	provider := ui.NewToolProvider(tools)
	ac.RegisterProvider("test", provider)
	ac.Show("", "test")

	// Should limit to 10
	completions := ac.GetCompletions()
	assert.LessOrEqual(t, len(completions), 10)
}

func TestAutocomplete_EmptyCompletions(t *testing.T) {
	ac := ui.NewAutocomplete()
	provider := ui.NewToolProvider([]string{"alpha", "beta"})
	ac.RegisterProvider("test", provider)

	// Show with query that has no matches
	ac.Show("xyz", "test")

	// Should not be active when no completions
	assert.False(t, ac.IsActive())
}

func TestCompletion_Struct(t *testing.T) {
	c := ui.Completion{
		Value:       "test_value",
		Display:     "Test Value",
		Description: "A test completion",
		Score:       100,
	}

	assert.Equal(t, "test_value", c.Value)
	assert.Equal(t, "Test Value", c.Display)
	assert.Equal(t, "A test completion", c.Description)
	assert.Equal(t, 100, c.Score)
}
