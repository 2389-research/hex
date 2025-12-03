// Package visualization provides real-time visualization of token usage and context windows.
// ABOUTME: Token usage visualization with real-time tracking and warnings
// ABOUTME: Displays context window fill, token breakdown, and historical usage graphs
package visualization

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/harper/hex/internal/ui/theme"
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
	theme         *theme.Theme
	current       TokenUsage
	history       []TokenUsage
	progress      progress.Model
	width         int
	height        int
	detailedView  bool
	warningShown  bool
	maxHistoryLen int
}

// NewTokenVisualization creates a new token visualization
func NewTokenVisualization(t *theme.Theme) *TokenVisualization {
	prog := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
		progress.WithoutPercentage(),
	)

	// Style the progress bar with Dracula colors
	prog.FullColor = string(t.Colors.Purple)
	prog.EmptyColor = string(t.Colors.CurrentLine)

	return &TokenVisualization{
		theme:         t,
		progress:      prog,
		history:       []TokenUsage{},
		maxHistoryLen: 50,
		detailedView:  false,
	}
}

// Init initializes the visualization
func (tv *TokenVisualization) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (tv *TokenVisualization) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		tv.width = msg.Width
		tv.height = msg.Height
		tv.progress.Width = msg.Width - 20
		return tv, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return tv, tea.Quit

		case "d":
			// Toggle detailed view
			tv.detailedView = !tv.detailedView
			return tv, nil
		}

	case TokenUpdateMsg:
		tv.updateUsage(msg.Usage)
		return tv, nil
	}

	return tv, nil
}

// View renders the token visualization
func (tv *TokenVisualization) View() string {
	if tv.width == 0 {
		return "Loading..."
	}

	if tv.detailedView {
		return tv.renderDetailed()
	}
	return tv.renderCompact()
}

// renderCompact renders a compact token status
func (tv *TokenVisualization) renderCompact() string {
	if tv.current.MaxTokens == 0 {
		return tv.renderEmpty()
	}

	// Calculate usage percentage
	percentage := float64(tv.current.TotalTokens) / float64(tv.current.MaxTokens)

	// Progress bar
	tv.progress.SetPercent(percentage)
	progressBar := tv.progress.View()

	// Token counts
	countStyle := lipgloss.NewStyle().Foreground(tv.theme.Colors.Cyan)
	counts := countStyle.Render(fmt.Sprintf("%d / %d tokens",
		tv.current.TotalTokens, tv.current.MaxTokens))

	// Warning if above 80%
	var warning string
	if percentage >= 0.8 {
		warningStyle := lipgloss.NewStyle().
			Foreground(tv.theme.Colors.Red).
			Bold(true)
		warningIcon := "⚠️"
		if percentage >= 0.95 {
			warningIcon = "🚨"
		}
		warning = "\n" + warningStyle.Render(fmt.Sprintf("%s Context window at %.0f%%!", warningIcon, percentage*100))
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		progressBar,
		counts,
		warning,
	)
}

// renderDetailed renders a detailed view with breakdown and history
func (tv *TokenVisualization) renderDetailed() string {
	var sections []string

	// Title
	titleStyle := lipgloss.NewStyle().
		Foreground(tv.theme.Colors.Pink).
		Bold(true).
		Padding(1, 0)
	sections = append(sections, titleStyle.Render("Token Usage Dashboard"))

	// Current usage section
	sections = append(sections, tv.renderCurrentUsage())

	// Breakdown section
	sections = append(sections, tv.renderBreakdown())

	// History sparkline
	if len(tv.history) > 0 {
		sections = append(sections, tv.renderHistory())
	}

	// Help
	helpStyle := lipgloss.NewStyle().
		Foreground(tv.theme.Colors.Comment).
		Padding(1, 0)
	sections = append(sections, helpStyle.Render("[d] Toggle View • [q] Quit"))

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// renderCurrentUsage renders the current usage with progress bar
func (tv *TokenVisualization) renderCurrentUsage() string {
	if tv.current.MaxTokens == 0 {
		return tv.renderEmpty()
	}

	percentage := float64(tv.current.TotalTokens) / float64(tv.current.MaxTokens)

	// Model info
	modelStyle := lipgloss.NewStyle().Foreground(tv.theme.Colors.Purple)
	modelInfo := modelStyle.Render(fmt.Sprintf("Model: %s", tv.current.ModelName))

	// Progress bar with label
	tv.progress.SetPercent(percentage)
	progressBar := tv.progress.View()

	// Percentage and counts
	countStyle := lipgloss.NewStyle().Foreground(tv.theme.Colors.Cyan)
	counts := countStyle.Render(fmt.Sprintf("%.1f%% - %d / %d tokens",
		percentage*100, tv.current.TotalTokens, tv.current.MaxTokens))

	// Warning
	var warning string
	if percentage >= 0.8 {
		warningStyle := lipgloss.NewStyle().
			Foreground(tv.theme.Colors.Yellow).
			Bold(true)
		if percentage >= 0.95 {
			warningStyle = warningStyle.Foreground(tv.theme.Colors.Red)
		}
		warning = "\n" + warningStyle.Render(fmt.Sprintf("⚠️  Warning: Context window %.0f%% full", percentage*100))
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		modelInfo,
		progressBar,
		counts,
		warning,
	)
}

// renderBreakdown renders input vs output token breakdown
func (tv *TokenVisualization) renderBreakdown() string {
	if tv.current.TotalTokens == 0 {
		return ""
	}

	titleStyle := lipgloss.NewStyle().
		Foreground(tv.theme.Colors.Yellow).
		Bold(true).
		Padding(1, 0, 0, 0)

	title := titleStyle.Render("Token Breakdown:")

	inputPct := float64(tv.current.InputTokens) / float64(tv.current.TotalTokens) * 100
	outputPct := float64(tv.current.OutputTokens) / float64(tv.current.TotalTokens) * 100

	inputStyle := lipgloss.NewStyle().Foreground(tv.theme.Colors.Green)
	outputStyle := lipgloss.NewStyle().Foreground(tv.theme.Colors.Cyan)

	input := inputStyle.Render(fmt.Sprintf("  Input:  %6d (%.1f%%)", tv.current.InputTokens, inputPct))
	output := outputStyle.Render(fmt.Sprintf("  Output: %6d (%.1f%%)", tv.current.OutputTokens, outputPct))

	return lipgloss.JoinVertical(lipgloss.Left, title, input, output)
}

// renderHistory renders a sparkline of historical token usage
func (tv *TokenVisualization) renderHistory() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(tv.theme.Colors.Yellow).
		Bold(true).
		Padding(1, 0, 0, 0)

	title := titleStyle.Render("Usage History:")

	// Generate sparkline
	sparkline := tv.generateSparkline(tv.history, tv.width-4)

	sparkStyle := lipgloss.NewStyle().Foreground(tv.theme.Colors.Purple)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		sparkStyle.Render(sparkline),
	)
}

// generateSparkline creates a simple text sparkline
func (tv *TokenVisualization) generateSparkline(history []TokenUsage, width int) string {
	if len(history) == 0 {
		return ""
	}

	// Use last N points that fit in width
	points := history
	if len(points) > width {
		points = points[len(points)-width:]
	}

	// Find max for normalization
	maxVal := 0.0
	for _, usage := range points {
		val := float64(usage.TotalTokens) / float64(usage.MaxTokens)
		if val > maxVal {
			maxVal = val
		}
	}
	if maxVal == 0 {
		maxVal = 1
	}

	// Sparkline characters (from empty to full)
	chars := []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

	var sparkline strings.Builder
	for _, usage := range points {
		val := float64(usage.TotalTokens) / float64(usage.MaxTokens)
		normalized := val / maxVal
		idx := int(math.Floor(normalized * float64(len(chars)-1)))
		if idx >= len(chars) {
			idx = len(chars) - 1
		}
		sparkline.WriteRune(chars[idx])
	}

	return sparkline.String()
}

// renderEmpty renders an empty state
func (tv *TokenVisualization) renderEmpty() string {
	emptyStyle := lipgloss.NewStyle().
		Foreground(tv.theme.Colors.Comment).
		Italic(true)
	return emptyStyle.Render("No token usage data yet")
}

// updateUsage updates the current usage and adds to history
func (tv *TokenVisualization) updateUsage(usage TokenUsage) {
	tv.current = usage

	// Add to history
	tv.history = append(tv.history, usage)

	// Trim history if too long
	if len(tv.history) > tv.maxHistoryLen {
		tv.history = tv.history[1:]
	}

	// Check for warning threshold
	percentage := float64(usage.TotalTokens) / float64(usage.MaxTokens)
	if percentage >= 0.8 && !tv.warningShown {
		tv.warningShown = true
	} else if percentage < 0.8 {
		tv.warningShown = false
	}
}

// UpdateTokens updates the token visualization with new usage data
func (tv *TokenVisualization) UpdateTokens(usage TokenUsage) {
	tv.updateUsage(usage)
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

// IsNearCapacity returns true if usage is above 80%
func (tv *TokenVisualization) IsNearCapacity() bool {
	return tv.GetUsagePercentage() >= 0.8
}

// RenderStatusBar renders a compact status for the status bar
func (tv *TokenVisualization) RenderStatusBar() string {
	if tv.current.MaxTokens == 0 {
		return ""
	}

	percentage := tv.GetUsagePercentage()

	// Choose color based on usage
	var style lipgloss.Style
	if percentage >= 0.95 {
		style = lipgloss.NewStyle().Foreground(tv.theme.Colors.Red)
	} else if percentage >= 0.8 {
		style = lipgloss.NewStyle().Foreground(tv.theme.Colors.Yellow)
	} else {
		style = lipgloss.NewStyle().Foreground(tv.theme.Colors.Cyan)
	}

	return style.Render(fmt.Sprintf("Tokens: %d/%d (%.0f%%)",
		tv.current.TotalTokens, tv.current.MaxTokens, percentage*100))
}

// Messages

// TokenUpdateMsg notifies of token usage updates
type TokenUpdateMsg struct {
	Usage TokenUsage
}
