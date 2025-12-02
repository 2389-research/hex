package themes_test

import (
	"testing"

	"github.com/harper/pagent/internal/ui/themes"
	"github.com/stretchr/testify/assert"
)

func TestGetTheme(t *testing.T) {
	tests := []struct {
		name          string
		themeName     string
		expectedTheme string
	}{
		{
			name:          "dracula theme",
			themeName:     "dracula",
			expectedTheme: "Dracula",
		},
		{
			name:          "gruvbox theme",
			themeName:     "gruvbox",
			expectedTheme: "Gruvbox Dark",
		},
		{
			name:          "nord theme",
			themeName:     "nord",
			expectedTheme: "Nord",
		},
		{
			name:          "unknown theme defaults to dracula",
			themeName:     "unknown",
			expectedTheme: "Dracula",
		},
		{
			name:          "empty string defaults to dracula",
			themeName:     "",
			expectedTheme: "Dracula",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			theme := themes.GetTheme(tt.themeName)
			assert.NotNil(t, theme)
			assert.Equal(t, tt.expectedTheme, theme.Name())
		})
	}
}

func TestAvailableThemes(t *testing.T) {
	available := themes.AvailableThemes()
	assert.Len(t, available, 3)
	assert.Contains(t, available, "dracula")
	assert.Contains(t, available, "gruvbox")
	assert.Contains(t, available, "nord")
}

func TestThemeInterface(t *testing.T) {
	// Test that all themes implement the Theme interface properly
	themeNames := themes.AvailableThemes()

	for _, name := range themeNames {
		t.Run(name, func(t *testing.T) {
			theme := themes.GetTheme(name)

			// Test all methods return non-empty values
			assert.NotEmpty(t, theme.Name())
			assert.NotEmpty(t, theme.Background())
			assert.NotEmpty(t, theme.Foreground())
			assert.NotEmpty(t, theme.Primary())
			assert.NotEmpty(t, theme.Secondary())
			assert.NotEmpty(t, theme.Success())
			assert.NotEmpty(t, theme.Warning())
			assert.NotEmpty(t, theme.Error())
			assert.NotEmpty(t, theme.Border())
			assert.NotEmpty(t, theme.BorderFocus())
			assert.NotEmpty(t, theme.Subtle())

			// Test gradient has at least 2 colors
			gradient := theme.TitleGradient()
			assert.GreaterOrEqual(t, len(gradient), 2, "gradient should have at least 2 colors")
		})
	}
}
