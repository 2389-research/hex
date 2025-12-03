// Package components provides reusable Bubbles components with Dracula theme styling.
// ABOUTME: Progress bar component for visual feedback during operations
// ABOUTME: Wraps bubbles.Progress with Dracula theme for streaming, tool execution, and token usage
package components

import (
	"fmt"

	"github.com/2389-research/hex/internal/ui/theme"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
)

// ProgressType represents different types of progress indicators
type ProgressType int

const (
	// ProgressTypeStreaming indicates API streaming progress
	ProgressTypeStreaming ProgressType = iota
	// ProgressTypeToolExecution indicates tool execution progress
	ProgressTypeToolExecution
	// ProgressTypeTokenUsage indicates token/context usage
	ProgressTypeTokenUsage
	// ProgressTypeBatch indicates batch operation progress
	ProgressTypeBatch
)

// Progress wraps bubbles.Progress with Dracula styling
type Progress struct {
	progress     progress.Model
	theme        *theme.Theme
	progressType ProgressType
	label        string
	value        float64
	total        float64
	width        int
}

// NewProgress creates a new progress bar with Dracula styling
func NewProgress(progressType ProgressType, width int) *Progress {
	draculaTheme := theme.DraculaTheme()

	// Set custom gradient based on progress type
	var prog progress.Model
	switch progressType {
	case ProgressTypeStreaming:
		// Cyan gradient for streaming
		prog = progress.New(
			progress.WithGradient(
				string(draculaTheme.Colors.Cyan),
				string(draculaTheme.Colors.Purple),
			),
			progress.WithWidth(width),
		)
	case ProgressTypeToolExecution:
		// Orange/Yellow gradient for tool execution
		prog = progress.New(
			progress.WithGradient(
				string(draculaTheme.Colors.Orange),
				string(draculaTheme.Colors.Yellow),
			),
			progress.WithWidth(width),
		)
	case ProgressTypeTokenUsage:
		// Green to Yellow gradient for token usage (green = low, yellow/red = high)
		prog = progress.New(
			progress.WithGradient(
				string(draculaTheme.Colors.Green),
				string(draculaTheme.Colors.Yellow),
			),
			progress.WithWidth(width),
		)
	case ProgressTypeBatch:
		// Purple/Pink gradient for batch operations
		prog = progress.New(
			progress.WithGradient(
				string(draculaTheme.Colors.Purple),
				string(draculaTheme.Colors.Pink),
			),
			progress.WithWidth(width),
		)
	default:
		prog = progress.New(progress.WithDefaultGradient(), progress.WithWidth(width))
	}

	return &Progress{
		progress:     prog,
		theme:        draculaTheme,
		progressType: progressType,
		width:        width,
		value:        0,
		total:        100,
	}
}

// SetLabel sets the progress bar label
func (p *Progress) SetLabel(label string) {
	p.label = label
}

// SetProgress updates the progress value (0.0 to 1.0)
func (p *Progress) SetProgress(value float64) {
	if value < 0 {
		value = 0
	}
	if value > 1 {
		value = 1
	}
	p.value = value
}

// SetProgressValues sets current and total values (calculates percentage)
func (p *Progress) SetProgressValues(current, total float64) {
	p.total = total
	if total > 0 {
		p.SetProgress(current / total)
	}
}

// IncrementProgress increments the progress by a delta
func (p *Progress) IncrementProgress(delta float64) {
	p.SetProgress(p.value + delta)
}

// SetWidth sets the progress bar width
func (p *Progress) SetWidth(width int) {
	p.width = width
	p.progress.Width = width
}

// GetProgress returns the current progress value
func (p *Progress) GetProgress() float64 {
	return p.value
}

// IsComplete returns true if progress is at 100%
func (p *Progress) IsComplete() bool {
	return p.value >= 1.0
}

// Reset resets the progress to 0
func (p *Progress) Reset() {
	p.value = 0
	p.label = ""
}

// Update handles progress bar updates
func (p *Progress) Update(msg tea.Msg) tea.Cmd {
	model, cmd := p.progress.Update(msg)
	p.progress = model.(progress.Model)
	return cmd
}

// View renders the progress bar
func (p *Progress) View() string {
	var output string

	// Render label if present
	if p.label != "" {
		labelStyle := p.theme.Body.Width(p.width)
		output = labelStyle.Render(p.label) + "\n"
	}

	// Render progress bar
	output += p.progress.ViewAs(p.value)

	// Add percentage or count indicator
	if p.progressType == ProgressTypeTokenUsage {
		// Show token count and percentage
		percent := int(p.value * 100)
		indicator := p.theme.Muted.Render(fmt.Sprintf(" %d%% (%.0f/%.0f tokens)", percent, p.value*p.total, p.total))
		output += indicator
	} else if p.total > 1 {
		// Show count for batch operations
		indicator := p.theme.Muted.Render(fmt.Sprintf(" %.0f/%.0f", p.value*p.total, p.total))
		output += indicator
	} else {
		// Show percentage
		percent := int(p.value * 100)
		indicator := p.theme.Muted.Render(fmt.Sprintf(" %d%%", percent))
		output += indicator
	}

	return output
}

// ViewCompact renders a compact version without label
func (p *Progress) ViewCompact() string {
	return p.progress.ViewAs(p.value)
}

// StreamingProgress creates a progress bar for API streaming
func StreamingProgress(chunksReceived, estimatedTotal int, width int) *Progress {
	prog := NewProgress(ProgressTypeStreaming, width)
	prog.SetLabel("Streaming response...")
	if estimatedTotal > 0 {
		prog.SetProgressValues(float64(chunksReceived), float64(estimatedTotal))
	} else {
		// Indeterminate progress for unknown total
		prog.SetProgress(0.5)
	}
	return prog
}

// ToolExecutionProgress creates a progress bar for tool execution
func ToolExecutionProgress(toolName string, width int) *Progress {
	prog := NewProgress(ProgressTypeToolExecution, width)
	prog.SetLabel(fmt.Sprintf("Executing %s...", toolName))
	prog.SetProgress(0.5) // Indeterminate for unknown duration
	return prog
}

// TokenUsageProgress creates a progress bar for token/context usage
func TokenUsageProgress(currentTokens, maxTokens int, width int) *Progress {
	prog := NewProgress(ProgressTypeTokenUsage, width)
	prog.SetLabel("Context window")
	prog.SetProgressValues(float64(currentTokens), float64(maxTokens))
	return prog
}

// BatchProgress creates a progress bar for batch operations
func BatchProgress(current, total int, operation string, width int) *Progress {
	prog := NewProgress(ProgressTypeBatch, width)
	prog.SetLabel(fmt.Sprintf("%s (%d/%d)", operation, current, total))
	prog.SetProgressValues(float64(current), float64(total))
	return prog
}

// ProgressBar is a simple wrapper for backward compatibility
type ProgressBar struct {
	*Progress
}

// NewProgressBar creates a new progress bar (backward compatible)
func NewProgressBar(width int) *ProgressBar {
	return &ProgressBar{
		Progress: NewProgress(ProgressTypeStreaming, width),
	}
}

// ProgressBarMsg is a message for updating progress
type ProgressBarMsg struct {
	Progress float64
	Label    string
}

// ProgressCompleteMsg is sent when progress reaches 100%
type ProgressCompleteMsg struct{}
