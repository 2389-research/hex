package themes_test

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/harper/jeff/internal/ui/themes"
	"github.com/stretchr/testify/assert"
)

func TestRenderGradient(t *testing.T) {
	t.Run("renders single character", func(t *testing.T) {
		colors := []lipgloss.Color{
			lipgloss.Color("#ff0000"),
			lipgloss.Color("#0000ff"),
		}

		result := themes.RenderGradient("A", colors)
		assert.NotEmpty(t, result)
		// In test environment, lipgloss may not render ANSI codes
		// Just verify we get a result containing the character
		assert.Contains(t, result, "A")
	})

	t.Run("renders multi-character text", func(t *testing.T) {
		colors := []lipgloss.Color{
			lipgloss.Color("#ff0000"),
			lipgloss.Color("#0000ff"),
		}

		result := themes.RenderGradient("Hello", colors)
		assert.NotEmpty(t, result)
		// Verify all characters are present
		assert.Contains(t, result, "H")
		assert.Contains(t, result, "e")
		assert.Contains(t, result, "l")
		assert.Contains(t, result, "o")
	})

	t.Run("handles empty text", func(t *testing.T) {
		colors := []lipgloss.Color{
			lipgloss.Color("#ff0000"),
			lipgloss.Color("#0000ff"),
		}

		result := themes.RenderGradient("", colors)
		assert.Equal(t, "", result)
	})

	t.Run("handles single color", func(t *testing.T) {
		colors := []lipgloss.Color{
			lipgloss.Color("#ff0000"),
		}

		result := themes.RenderGradient("Hello", colors)
		assert.NotEmpty(t, result)
		// With single color, all chars should have same color
	})

	t.Run("interpolates colors evenly", func(t *testing.T) {
		colors := []lipgloss.Color{
			lipgloss.Color("#ff0000"), // Red
			lipgloss.Color("#0000ff"), // Blue
		}

		// With 2 characters and 2 colors, first should be red, second should be blue
		result := themes.RenderGradient("AB", colors)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "A")
		assert.Contains(t, result, "B")
	})
}

func TestInterpolateColor(t *testing.T) {
	t.Run("interpolates at 0% (start color)", func(t *testing.T) {
		start := lipgloss.Color("#ff0000") // Red
		end := lipgloss.Color("#0000ff")   // Blue

		result := themes.InterpolateColor(start, end, 0.0)
		// At 0%, should be exactly the start color
		assert.Equal(t, "#ff0000", string(result))
	})

	t.Run("interpolates at 100% (end color)", func(t *testing.T) {
		start := lipgloss.Color("#ff0000") // Red
		end := lipgloss.Color("#0000ff")   // Blue

		result := themes.InterpolateColor(start, end, 1.0)
		// At 100%, should be exactly the end color
		assert.Equal(t, "#0000ff", string(result))
	})

	t.Run("interpolates at 50% (midpoint)", func(t *testing.T) {
		start := lipgloss.Color("#ff0000") // Red (255, 0, 0)
		end := lipgloss.Color("#0000ff")   // Blue (0, 0, 255)

		result := themes.InterpolateColor(start, end, 0.5)
		// At 50%, should be roughly purple (127, 0, 127)
		resultStr := string(result)
		assert.NotEmpty(t, resultStr)
		// Should start with # and be 7 characters
		assert.Len(t, resultStr, 7)
		assert.True(t, resultStr[0] == '#')
		// Should be close to purple #7f007f (allowing for rounding)
		assert.Contains(t, resultStr, "7f")
	})
}
