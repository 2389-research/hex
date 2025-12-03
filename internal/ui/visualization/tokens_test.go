// Package visualization provides real-time visualization of token usage and context windows.
// ABOUTME: Tests for token usage visualization component
// ABOUTME: Validates token calculations, progress rendering, and warning thresholds
package visualization

import (
	"strings"
	"testing"

	"github.com/harper/pagent/internal/ui/themes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTokenVisualization(t *testing.T) {
	theme := themes.NewDracula()
	tv := NewTokenVisualization(theme)

	assert.NotNil(t, tv)
	assert.NotNil(t, tv.progress)
	assert.Equal(t, theme, tv.theme)
	assert.Equal(t, 0, tv.current.TotalTokens)
	assert.False(t, tv.warningShown)
}

func TestTokenVisualization_Update(t *testing.T) {
	theme := themes.NewDracula()
	tv := NewTokenVisualization(theme)

	usage := TokenUsage{
		InputTokens:  60000,
		OutputTokens: 30000,
		TotalTokens:  90000,
		MaxTokens:    200000,
		ModelName:    "claude-3-5-sonnet-20241022",
	}

	tv.Update(usage)

	assert.Equal(t, usage, tv.current)
	assert.Equal(t, 60000, tv.current.InputTokens)
	assert.Equal(t, 30000, tv.current.OutputTokens)
	assert.Equal(t, 90000, tv.current.TotalTokens)
}

func TestTokenVisualization_SetWidth(t *testing.T) {
	theme := themes.NewDracula()
	tv := NewTokenVisualization(theme)

	tv.SetWidth(80)
	assert.Equal(t, 80, tv.width)
	assert.Equal(t, 60, tv.progress.Width) // 80 - 20

	tv.SetWidth(15)
	assert.Equal(t, 15, tv.width)
	assert.Equal(t, 20, tv.progress.Width) // Minimum width
}

func TestTokenVisualization_GetUsagePercentage(t *testing.T) {
	theme := themes.NewDracula()
	tv := NewTokenVisualization(theme)

	tests := []struct {
		name     string
		usage    TokenUsage
		expected float64
	}{
		{
			name: "45% usage",
			usage: TokenUsage{
				TotalTokens: 90000,
				MaxTokens:   200000,
			},
			expected: 0.45,
		},
		{
			name: "80% usage",
			usage: TokenUsage{
				TotalTokens: 160000,
				MaxTokens:   200000,
			},
			expected: 0.80,
		},
		{
			name: "100% usage",
			usage: TokenUsage{
				TotalTokens: 200000,
				MaxTokens:   200000,
			},
			expected: 1.0,
		},
		{
			name: "no max tokens",
			usage: TokenUsage{
				TotalTokens: 50000,
				MaxTokens:   0,
			},
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tv.Update(tt.usage)
			percentage := tv.GetUsagePercentage()
			assert.InDelta(t, tt.expected, percentage, 0.001)
		})
	}
}

func TestTokenVisualization_ShouldWarn(t *testing.T) {
	theme := themes.NewDracula()
	tv := NewTokenVisualization(theme)

	tests := []struct {
		name        string
		usage       TokenUsage
		shouldWarn  bool
		description string
	}{
		{
			name: "below threshold - 50%",
			usage: TokenUsage{
				TotalTokens: 100000,
				MaxTokens:   200000,
			},
			shouldWarn:  false,
			description: "50% usage should not trigger warning",
		},
		{
			name: "below threshold - 79%",
			usage: TokenUsage{
				TotalTokens: 158000,
				MaxTokens:   200000,
			},
			shouldWarn:  false,
			description: "79% usage should not trigger warning",
		},
		{
			name: "at threshold - 80%",
			usage: TokenUsage{
				TotalTokens: 160000,
				MaxTokens:   200000,
			},
			shouldWarn:  true,
			description: "80% usage should trigger warning",
		},
		{
			name: "above threshold - 90%",
			usage: TokenUsage{
				TotalTokens: 180000,
				MaxTokens:   200000,
			},
			shouldWarn:  true,
			description: "90% usage should trigger warning",
		},
		{
			name: "critical - 95%",
			usage: TokenUsage{
				TotalTokens: 190000,
				MaxTokens:   200000,
			},
			shouldWarn:  true,
			description: "95% usage should trigger warning",
		},
		{
			name: "no max tokens",
			usage: TokenUsage{
				TotalTokens: 100000,
				MaxTokens:   0,
			},
			shouldWarn:  false,
			description: "No max tokens should not trigger warning",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tv.Update(tt.usage)
			assert.Equal(t, tt.shouldWarn, tv.ShouldWarn(), tt.description)
		})
	}
}

func TestTokenVisualization_Render(t *testing.T) {
	theme := themes.NewDracula()
	tv := NewTokenVisualization(theme)
	tv.SetWidth(80)

	t.Run("empty state", func(t *testing.T) {
		rendered := tv.Render()
		assert.Contains(t, rendered, "No token usage data yet")
	})

	t.Run("normal usage - 45%", func(t *testing.T) {
		usage := TokenUsage{
			InputTokens:  60000,
			OutputTokens: 30000,
			TotalTokens:  90000,
			MaxTokens:    200000,
			ModelName:    "claude-3-5-sonnet-20241022",
		}
		tv.Update(usage)
		rendered := tv.Render()

		// Should contain token information
		assert.Contains(t, rendered, "Tokens:")
		assert.Contains(t, rendered, "Input:")
		assert.Contains(t, rendered, "Output:")
		assert.Contains(t, rendered, "Remaining:")

		// Should contain values
		assert.Contains(t, rendered, "60K")  // Input
		assert.Contains(t, rendered, "30K")  // Output
		assert.Contains(t, rendered, "110K") // Remaining

		// Should NOT contain warning at 45%
		assert.NotContains(t, rendered, "⚠️")
		assert.NotContains(t, rendered, "Approaching limit")
	})

	t.Run("warning usage - 85%", func(t *testing.T) {
		usage := TokenUsage{
			InputTokens:  120000,
			OutputTokens: 50000,
			TotalTokens:  170000,
			MaxTokens:    200000,
			ModelName:    "claude-3-5-sonnet-20241022",
		}
		tv.Update(usage)
		rendered := tv.Render()

		// Should contain warning icon and message
		assert.Contains(t, rendered, "⚠️")
		assert.Contains(t, rendered, "Approaching limit")
	})

	t.Run("critical usage - 96%", func(t *testing.T) {
		usage := TokenUsage{
			InputTokens:  140000,
			OutputTokens: 52000,
			TotalTokens:  192000,
			MaxTokens:    200000,
			ModelName:    "claude-3-5-sonnet-20241022",
		}
		tv.Update(usage)
		rendered := tv.Render()

		// Should contain critical warning icon
		assert.Contains(t, rendered, "🚨")
		assert.Contains(t, rendered, "Approaching limit")
	})
}

func TestTokenVisualization_RenderCompact(t *testing.T) {
	theme := themes.NewDracula()
	tv := NewTokenVisualization(theme)

	t.Run("empty state", func(t *testing.T) {
		rendered := tv.RenderCompact()
		assert.Empty(t, rendered)
	})

	t.Run("normal usage - 45%", func(t *testing.T) {
		usage := TokenUsage{
			InputTokens:  60000,
			OutputTokens: 30000,
			TotalTokens:  90000,
			MaxTokens:    200000,
			ModelName:    "claude-3-5-sonnet-20241022",
		}
		tv.Update(usage)
		rendered := tv.RenderCompact()

		// Should be single line
		lines := strings.Split(rendered, "\n")
		assert.Equal(t, 1, len(lines))

		// Should contain compact bar and percentages
		assert.Contains(t, rendered, "Tokens:")
		assert.Contains(t, rendered, "[")
		assert.Contains(t, rendered, "]")
		assert.Contains(t, rendered, "45%")
		assert.Contains(t, rendered, "90K/200K")

		// Should have some filled bars (█) and some empty (░)
		assert.Contains(t, rendered, "█")
		assert.Contains(t, rendered, "░")
	})

	t.Run("progress bar accuracy", func(t *testing.T) {
		tests := []struct {
			name         string
			totalTokens  int
			maxTokens    int
			expectedFill int // out of 10
		}{
			{"0% filled", 0, 200000, 0},
			{"10% filled", 20000, 200000, 1},
			{"50% filled", 100000, 200000, 5},
			{"80% filled", 160000, 200000, 8},
			{"100% filled", 200000, 200000, 10},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				usage := TokenUsage{
					TotalTokens: tt.totalTokens,
					MaxTokens:   tt.maxTokens,
				}
				tv.Update(usage)
				rendered := tv.RenderCompact()

				// Count filled bars
				filledCount := strings.Count(rendered, "█")
				assert.Equal(t, tt.expectedFill, filledCount,
					"Expected %d filled bars for %d%% usage", tt.expectedFill, tt.totalTokens*100/tt.maxTokens)
			})
		}
	})
}

func TestTokenVisualization_WarningShown(t *testing.T) {
	theme := themes.NewDracula()
	tv := NewTokenVisualization(theme)

	// Below threshold
	usage := TokenUsage{
		TotalTokens: 100000,
		MaxTokens:   200000,
	}
	tv.Update(usage)
	assert.False(t, tv.warningShown)

	// Cross threshold
	usage.TotalTokens = 170000
	tv.Update(usage)
	assert.True(t, tv.warningShown)

	// Drop below threshold
	usage.TotalTokens = 100000
	tv.Update(usage)
	assert.False(t, tv.warningShown)
}

func TestTokenVisualization_GetCurrentUsage(t *testing.T) {
	theme := themes.NewDracula()
	tv := NewTokenVisualization(theme)

	usage := TokenUsage{
		InputTokens:  60000,
		OutputTokens: 30000,
		TotalTokens:  90000,
		MaxTokens:    200000,
		ModelName:    "claude-3-5-sonnet-20241022",
	}
	tv.Update(usage)

	current := tv.GetCurrentUsage()
	assert.Equal(t, usage, current)
}

func TestTokenVisualization_IntegrationScenario(t *testing.T) {
	// Simulate a real conversation flow
	theme := themes.NewDracula()
	tv := NewTokenVisualization(theme)
	tv.SetWidth(80)

	// Start with empty state
	rendered := tv.Render()
	assert.Contains(t, rendered, "No token usage")

	// First message - low usage
	tv.Update(TokenUsage{
		InputTokens:  5000,
		OutputTokens: 2000,
		TotalTokens:  7000,
		MaxTokens:    200000,
		ModelName:    "claude-3-5-sonnet-20241022",
	})
	assert.False(t, tv.ShouldWarn())
	percentage := tv.GetUsagePercentage()
	assert.Less(t, percentage, 0.1) // Less than 10%

	// Simulate conversation progress - moderate usage
	tv.Update(TokenUsage{
		InputTokens:  40000,
		OutputTokens: 20000,
		TotalTokens:  60000,
		MaxTokens:    200000,
		ModelName:    "claude-3-5-sonnet-20241022",
	})
	assert.False(t, tv.ShouldWarn())

	// Approaching limit - should warn
	tv.Update(TokenUsage{
		InputTokens:  120000,
		OutputTokens: 50000,
		TotalTokens:  170000,
		MaxTokens:    200000,
		ModelName:    "claude-3-5-sonnet-20241022",
	})
	assert.True(t, tv.ShouldWarn())
	rendered = tv.Render()
	assert.Contains(t, rendered, "⚠️")

	// Critical - should show urgent warning
	tv.Update(TokenUsage{
		InputTokens:  140000,
		OutputTokens: 52000,
		TotalTokens:  192000,
		MaxTokens:    200000,
		ModelName:    "claude-3-5-sonnet-20241022",
	})
	assert.True(t, tv.ShouldWarn())
	rendered = tv.Render()
	assert.Contains(t, rendered, "🚨")
}

func TestTokenVisualization_DifferentThemes(t *testing.T) {
	// Test that visualization works with different themes
	themes := []struct {
		name  string
		theme themes.Theme
	}{
		{"dracula", themes.NewDracula()},
		{"gruvbox", themes.NewGruvbox()},
		{"nord", themes.NewNord()},
	}

	usage := TokenUsage{
		InputTokens:  60000,
		OutputTokens: 30000,
		TotalTokens:  90000,
		MaxTokens:    200000,
		ModelName:    "claude-3-5-sonnet-20241022",
	}

	for _, tt := range themes {
		t.Run(tt.name, func(t *testing.T) {
			tv := NewTokenVisualization(tt.theme)
			tv.SetWidth(80)
			tv.Update(usage)

			rendered := tv.Render()
			require.NotEmpty(t, rendered)
			assert.Contains(t, rendered, "Tokens:")
			assert.Contains(t, rendered, "Input:")
			assert.Contains(t, rendered, "Output:")

			compact := tv.RenderCompact()
			require.NotEmpty(t, compact)
			assert.Contains(t, compact, "Tokens:")
		})
	}
}
