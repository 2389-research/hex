// Package visualization provides real-time visualization of token usage and context windows.
// ABOUTME: Token usage visualization with real-time tracking and warnings
// ABOUTME: Displays context window fill, token breakdown, and remaining capacity
package visualization

import (
	"fmt"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/lipgloss"
	"github.com/harper/pagent/internal/ui/themes"
)

// TokenUsage tracks token counts for a conversation
type TokenUsage struct {
	InputTokens  int
	OutputTokens int
	TotalTokens  int
	MaxTokens    int
	ModelName    string
}

// TokenVisualization manages token usage visualization
type TokenVisualization struct {
	theme        themes.Theme
	current      TokenUsage
	progress     progress.Model
	width        int
	warningShown bool
}

// NewTokenVisualization creates a new token visualization
func NewTokenVisualization(theme themes.Theme) *TokenVisualization {
	prog := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
		progress.WithoutPercentage(),
	)

	// Style the progress bar with theme colors
	prog.FullColor = string(theme.Primary())
	prog.EmptyColor = string(theme.Subtle())

	return &TokenVisualization{
		theme:    theme,
		progress: prog,
	}
}

// Update updates the token visualization with new usage data
func (tv *TokenVisualization) Update(usage TokenUsage) {
	tv.current = usage

	// Check for warning threshold
	if tv.ShouldWarn() && !tv.warningShown {
		tv.warningShown = true
	} else if !tv.ShouldWarn() {
		tv.warningShown = false
	}
}

// SetWidth updates the visualization width
func (tv *TokenVisualization) SetWidth(width int) {
	tv.width = width
	// Update progress bar width (leave some room for text)
	if width > 20 {
		tv.progress.Width = width - 20
	} else {
		tv.progress.Width = 20
	}
}

// Render renders the full token visualization
func (tv *TokenVisualization) Render() string {
	if tv.current.MaxTokens == 0 {
		return tv.renderEmpty()
	}

	// Calculate usage percentage
	percentage := float64(tv.current.TotalTokens) / float64(tv.current.MaxTokens)

	// Progress bar
	tv.progress.SetPercent(percentage)
	progressBar := tv.progress.View()

	// Token counts with formatting
	countStyle := lipgloss.NewStyle().Foreground(tv.theme.Secondary())
	percentText := fmt.Sprintf("%.0f%%", percentage*100)
	totalText := fmt.Sprintf("(%dK/%dK)",
		tv.current.TotalTokens/1000,
		tv.current.MaxTokens/1000)

	tokensLine := fmt.Sprintf("Tokens: %s %s %s",
		progressBar,
		countStyle.Render(percentText),
		countStyle.Render(totalText))

	// Breakdown line
	inputStyle := lipgloss.NewStyle().Foreground(tv.theme.Success())
	outputStyle := lipgloss.NewStyle().Foreground(tv.theme.Primary())
	subtleStyle := lipgloss.NewStyle().Foreground(tv.theme.Subtle())

	remaining := tv.current.MaxTokens - tv.current.TotalTokens
	breakdownLine := fmt.Sprintf("%s %s | %s %s | %s %s",
		subtleStyle.Render("Input:"),
		inputStyle.Render(fmt.Sprintf("%dK", tv.current.InputTokens/1000)),
		subtleStyle.Render("Output:"),
		outputStyle.Render(fmt.Sprintf("%dK", tv.current.OutputTokens/1000)),
		subtleStyle.Render("Remaining:"),
		subtleStyle.Render(fmt.Sprintf("%dK", remaining/1000)))

	result := tokensLine + "\n" + breakdownLine

	// Warning if above threshold
	if tv.ShouldWarn() {
		warningStyle := lipgloss.NewStyle().
			Foreground(tv.theme.Warning()).
			Bold(true)

		warningIcon := "⚠️"
		if percentage >= 0.95 {
			warningIcon = "🚨"
			warningStyle = warningStyle.Foreground(tv.theme.Error())
		}

		warningMsg := fmt.Sprintf("%s  Approaching limit - consider new session", warningIcon)
		result += "\n" + warningStyle.Render(warningMsg)
	}

	return result
}

// RenderCompact renders a compact single-line token status
func (tv *TokenVisualization) RenderCompact() string {
	if tv.current.MaxTokens == 0 {
		return ""
	}

	percentage := float64(tv.current.TotalTokens) / float64(tv.current.MaxTokens)

	// Choose color based on usage
	var style lipgloss.Style
	if percentage >= 0.95 {
		style = lipgloss.NewStyle().Foreground(tv.theme.Error())
	} else if percentage >= 0.8 {
		style = lipgloss.NewStyle().Foreground(tv.theme.Warning())
	} else {
		style = lipgloss.NewStyle().Foreground(tv.theme.Secondary())
	}

	// Compact format: "Tokens: [███░░░] 45% (90K/200K)"
	barWidth := 10
	filled := int(percentage * float64(barWidth))
	if filled > barWidth {
		filled = barWidth
	}

	bar := "["
	for i := 0; i < barWidth; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}
	bar += "]"

	return style.Render(fmt.Sprintf("Tokens: %s %.0f%% (%dK/%dK)",
		bar,
		percentage*100,
		tv.current.TotalTokens/1000,
		tv.current.MaxTokens/1000))
}

// ShouldWarn returns true if usage is above 80% context used
func (tv *TokenVisualization) ShouldWarn() bool {
	if tv.current.MaxTokens == 0 {
		return false
	}
	percentage := float64(tv.current.TotalTokens) / float64(tv.current.MaxTokens)
	return percentage >= 0.8
}

// GetCurrentUsage returns the current token usage
func (tv *TokenVisualization) GetCurrentUsage() TokenUsage {
	return tv.current
}

// GetUsagePercentage returns the current usage as a percentage
func (tv *TokenVisualization) GetUsagePercentage() float64 {
	if tv.current.MaxTokens == 0 {
		return 0
	}
	return float64(tv.current.TotalTokens) / float64(tv.current.MaxTokens)
}

// renderEmpty renders an empty state
func (tv *TokenVisualization) renderEmpty() string {
	emptyStyle := lipgloss.NewStyle().
		Foreground(tv.theme.Subtle()).
		Italic(true)
	return emptyStyle.Render("No token usage data yet")
}
