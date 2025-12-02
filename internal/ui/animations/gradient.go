// Package animations provides gradient generation and color interpolation for visual polish.
// ABOUTME: Gradient generation and color interpolation for visual polish
// ABOUTME: Provides smooth color transitions and animated effects using Dracula theme
package animations

import (
	"fmt"
	"math"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// GradientStyle represents a horizontal gradient configuration
type GradientStyle struct {
	StartColor lipgloss.Color
	EndColor   lipgloss.Color
	Steps      int
}

// NewGradient creates a new gradient style with given colors and steps
func NewGradient(start, end lipgloss.Color, steps int) *GradientStyle {
	if steps < 2 {
		steps = 2
	}
	return &GradientStyle{
		StartColor: start,
		EndColor:   end,
		Steps:      steps,
	}
}

// parseHexColor converts a hex color string to RGB values
func parseHexColor(hex string) (r, g, b int) {
	// Remove # prefix if present
	if len(hex) > 0 && hex[0] == '#' {
		hex = hex[1:]
	}

	// Parse hex string
	if len(hex) == 6 {
		_, _ = fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
	}
	return
}

// interpolateColor creates an intermediate color between two colors
func interpolateColor(start, end lipgloss.Color, ratio float64) lipgloss.Color {
	r1, g1, b1 := parseHexColor(string(start))
	r2, g2, b2 := parseHexColor(string(end))

	r := int(float64(r1) + (float64(r2)-float64(r1))*ratio)
	g := int(float64(g1) + (float64(g2)-float64(g1))*ratio)
	b := int(float64(b1) + (float64(b2)-float64(b1))*ratio)

	return lipgloss.Color(fmt.Sprintf("#%02x%02x%02x", r, g, b))
}

// Render creates a gradient string with the given text
func (g *GradientStyle) Render(text string) string {
	if len(text) == 0 {
		return ""
	}

	var result string
	textLen := len([]rune(text))

	for i, char := range text {
		ratio := float64(i) / float64(textLen-1)
		if textLen == 1 {
			ratio = 0
		}

		color := interpolateColor(g.StartColor, g.EndColor, ratio)
		style := lipgloss.NewStyle().Foreground(color)
		result += style.Render(string(char))
	}

	return result
}

// RenderBar creates a gradient color bar of specified width
func (g *GradientStyle) RenderBar(width int) string {
	if width <= 0 {
		return ""
	}

	var result string
	for i := 0; i < width; i++ {
		ratio := float64(i) / float64(width-1)
		if width == 1 {
			ratio = 0
		}

		color := interpolateColor(g.StartColor, g.EndColor, ratio)
		style := lipgloss.NewStyle().Foreground(color)
		result += style.Render("█")
	}

	return result
}

// TransitionState tracks the state of an animated transition
type TransitionState struct {
	StartTime time.Time
	Duration  time.Duration
	StartVal  float64
	EndVal    float64
}

// NewTransition creates a new transition state
func NewTransition(duration time.Duration, startVal, endVal float64) *TransitionState {
	return &TransitionState{
		StartTime: time.Now(),
		Duration:  duration,
		StartVal:  startVal,
		EndVal:    endVal,
	}
}

// Value returns the current interpolated value based on elapsed time
func (t *TransitionState) Value() float64 {
	elapsed := time.Since(t.StartTime)
	if elapsed >= t.Duration {
		return t.EndVal
	}

	progress := float64(elapsed) / float64(t.Duration)
	// Ease-in-out cubic function
	progress = easeInOutCubic(progress)

	return t.StartVal + (t.EndVal-t.StartVal)*progress
}

// IsComplete returns true if the transition has finished
func (t *TransitionState) IsComplete() bool {
	return time.Since(t.StartTime) >= t.Duration
}

// easeInOutCubic applies cubic easing to progress value
func easeInOutCubic(t float64) float64 {
	if t < 0.5 {
		return 4 * t * t * t
	}
	v := -2*t + 2
	return 1 - v*v*v/2
}

// PulseEffect creates a pulsing animation effect
type PulseEffect struct {
	BaseColor      lipgloss.Color
	HighlightColor lipgloss.Color
	Period         time.Duration
	StartTime      time.Time
}

// NewPulseEffect creates a new pulse effect
func NewPulseEffect(base, highlight lipgloss.Color, period time.Duration) *PulseEffect {
	return &PulseEffect{
		BaseColor:      base,
		HighlightColor: highlight,
		Period:         period,
		StartTime:      time.Now(),
	}
}

// CurrentColor returns the current color in the pulse cycle
func (p *PulseEffect) CurrentColor() lipgloss.Color {
	elapsed := time.Since(p.StartTime)
	progress := float64(elapsed%p.Period) / float64(p.Period)

	// Use sine wave for smooth pulsing
	ratio := (math.Sin(progress*2*math.Pi) + 1) / 2

	return interpolateColor(p.BaseColor, p.HighlightColor, ratio)
}

// RenderWithPulse renders text with the current pulse color
func (p *PulseEffect) RenderWithPulse(text string) string {
	color := p.CurrentColor()
	style := lipgloss.NewStyle().Foreground(color)
	return style.Render(text)
}

// ShimmerEffect creates a shimmer animation that moves across text
type ShimmerEffect struct {
	BaseColor    lipgloss.Color
	ShimmerColor lipgloss.Color
	Period       time.Duration
	StartTime    time.Time
	ShimmerWidth float64 // Width of shimmer as fraction of text length
}

// NewShimmerEffect creates a new shimmer effect
func NewShimmerEffect(base, shimmer lipgloss.Color, period time.Duration) *ShimmerEffect {
	return &ShimmerEffect{
		BaseColor:    base,
		ShimmerColor: shimmer,
		Period:       period,
		StartTime:    time.Now(),
		ShimmerWidth: 0.3, // 30% of text length
	}
}

// RenderWithShimmer renders text with moving shimmer effect
func (s *ShimmerEffect) RenderWithShimmer(text string) string {
	if len(text) == 0 {
		return ""
	}

	elapsed := time.Since(s.StartTime)
	progress := float64(elapsed%s.Period) / float64(s.Period)

	var result string
	textLen := len([]rune(text))

	for i, char := range text {
		pos := float64(i) / float64(textLen)

		// Calculate shimmer intensity at this position
		shimmerCenter := progress
		distance := math.Abs(pos - shimmerCenter)

		var color lipgloss.Color
		if distance < s.ShimmerWidth/2 {
			// Within shimmer region
			intensity := 1 - (distance / (s.ShimmerWidth / 2))
			color = interpolateColor(s.BaseColor, s.ShimmerColor, intensity)
		} else {
			color = s.BaseColor
		}

		style := lipgloss.NewStyle().Foreground(color)
		result += style.Render(string(char))
	}

	return result
}

// FadeEffect handles fade-in and fade-out transitions
type FadeEffect struct {
	Color     lipgloss.Color
	Duration  time.Duration
	StartTime time.Time
	FadeIn    bool
}

// NewFadeIn creates a fade-in effect
func NewFadeIn(color lipgloss.Color, duration time.Duration) *FadeEffect {
	return &FadeEffect{
		Color:     color,
		Duration:  duration,
		StartTime: time.Now(),
		FadeIn:    true,
	}
}

// NewFadeOut creates a fade-out effect
func NewFadeOut(color lipgloss.Color, duration time.Duration) *FadeEffect {
	return &FadeEffect{
		Color:     color,
		Duration:  duration,
		StartTime: time.Now(),
		FadeIn:    false,
	}
}

// CurrentOpacity returns the current opacity (0.0 to 1.0)
func (f *FadeEffect) CurrentOpacity() float64 {
	elapsed := time.Since(f.StartTime)
	if elapsed >= f.Duration {
		if f.FadeIn {
			return 1.0
		}
		return 0.0
	}

	progress := float64(elapsed) / float64(f.Duration)
	if !f.FadeIn {
		progress = 1 - progress
	}

	return progress
}

// IsComplete returns true if fade is complete
func (f *FadeEffect) IsComplete() bool {
	return time.Since(f.StartTime) >= f.Duration
}
