// Package components provides reusable UI components for the TUI.
// ABOUTME: Progress bar component using bubbles progress
// ABOUTME: Displays completion percentage with theme colors
package components

import (
	"fmt"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/harper/pagent/internal/ui/themes"
)

// Progress is a themed progress bar component
type Progress struct {
	theme    themes.Theme
	label    string
	value    float64 // 0.0 to 1.0
	progress progress.Model
}

// NewProgress creates a new progress bar
func NewProgress(theme themes.Theme, label string) *Progress {
	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
	)

	return &Progress{
		theme:    theme,
		label:    label,
		value:    0.0,
		progress: p,
	}
}

// Init implements tea.Model
func (p *Progress) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (p *Progress) Update(_ tea.Msg) (tea.Model, tea.Cmd) {
	return p, nil
}

// View implements tea.Model
func (p *Progress) View() string {
	labelStyle := lipgloss.NewStyle().
		Foreground(p.theme.Foreground()).
		Bold(true)

	percentStyle := lipgloss.NewStyle().
		Foreground(p.theme.Primary()).
		Bold(true)

	label := labelStyle.Render(p.label)
	bar := p.progress.ViewAs(p.value)
	percent := percentStyle.Render(fmt.Sprintf(" %.0f%%", p.value*100))

	return lipgloss.JoinHorizontal(lipgloss.Left, label, " ", bar, percent)
}

// SetValue updates the progress value (0.0 to 1.0)
func (p *Progress) SetValue(value float64) {
	if value < 0.0 {
		value = 0.0
	}
	if value > 1.0 {
		value = 1.0
	}
	p.value = value
}

// GetValue returns the current progress value
func (p *Progress) GetValue() float64 {
	return p.value
}
