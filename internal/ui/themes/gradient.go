// Package themes provides theming support for the TUI.
// ABOUTME: Gradient rendering utility for smooth color transitions
// ABOUTME: Character-by-character color interpolation for visual polish
package themes

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// RenderGradient renders text with a gradient effect using the provided colors.
// Colors are interpolated evenly across the text length.
func RenderGradient(text string, colors []lipgloss.Color) string {
	if text == "" {
		return ""
	}

	if len(colors) == 0 {
		return text
	}

	if len(colors) == 1 {
		// Single color - just apply it to all characters
		return lipgloss.NewStyle().Foreground(colors[0]).Render(text)
	}

	var result strings.Builder
	textRunes := []rune(text)
	textLen := len(textRunes)

	// Handle single character case - just use first color
	if textLen == 1 {
		styled := lipgloss.NewStyle().Foreground(colors[0]).Render(text)
		return styled
	}

	for i, char := range textRunes {
		// Calculate position in gradient (0.0 to 1.0)
		position := float64(i) / float64(textLen-1)

		// Find which color segment we're in
		segmentSize := 1.0 / float64(len(colors)-1)
		segmentIndex := int(position / segmentSize)

		// Clamp to valid range
		if segmentIndex >= len(colors)-1 {
			segmentIndex = len(colors) - 2
		}

		// Calculate position within this segment (0.0 to 1.0)
		segmentPosition := (position - float64(segmentIndex)*segmentSize) / segmentSize

		// Interpolate between the two colors in this segment
		startColor := colors[segmentIndex]
		endColor := colors[segmentIndex+1]
		interpolatedColor := InterpolateColor(startColor, endColor, segmentPosition)

		// Apply color to this character
		styled := lipgloss.NewStyle().Foreground(interpolatedColor).Render(string(char))
		result.WriteString(styled)
	}

	return result.String()
}

// InterpolateColor interpolates between two colors at the given position (0.0 to 1.0).
func InterpolateColor(start, end lipgloss.Color, position float64) lipgloss.Color {
	// Clamp position to valid range
	if position < 0.0 {
		position = 0.0
	}
	if position > 1.0 {
		position = 1.0
	}

	// Parse hex colors (lipgloss.Color is a string type)
	startRGB := parseHexColor(string(start))
	endRGB := parseHexColor(string(end))

	// Interpolate each channel
	r := uint8(float64(startRGB[0]) + (float64(endRGB[0])-float64(startRGB[0]))*position)
	g := uint8(float64(startRGB[1]) + (float64(endRGB[1])-float64(startRGB[1]))*position)
	b := uint8(float64(startRGB[2]) + (float64(endRGB[2])-float64(startRGB[2]))*position)

	// Format as hex color
	return lipgloss.Color(fmt.Sprintf("#%02x%02x%02x", r, g, b))
}

// parseHexColor parses a hex color string (#RRGGBB) into RGB values.
func parseHexColor(hex string) [3]uint8 {
	// Remove # prefix if present
	hex = strings.TrimPrefix(hex, "#")

	// Parse RGB components
	var rgb [3]uint8
	if len(hex) >= 6 {
		if val, err := strconv.ParseUint(hex[0:2], 16, 8); err == nil {
			rgb[0] = uint8(val)
		}
		if val, err := strconv.ParseUint(hex[2:4], 16, 8); err == nil {
			rgb[1] = uint8(val)
		}
		if val, err := strconv.ParseUint(hex[4:6], 16, 8); err == nil {
			rgb[2] = uint8(val)
		}
	}

	return rgb
}
