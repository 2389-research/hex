// Package components provides reusable UI components for the Pagen TUI.
// ABOUTME: Huh-based quick actions selector using select form
// ABOUTME: Provides themed action selection menu for common operations
package components

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/harper/pagent/internal/ui/themes"
)

// QuickActionOption represents a quick action option
type QuickActionOption struct {
	Label string
	Value string
}

// HuhQuickActions is a Huh-based quick actions selector
type HuhQuickActions struct {
	theme    themes.Theme
	options  []QuickActionOption
	selected string
	form     *huh.Form
}

// NewHuhQuickActions creates a new Huh quick actions selector
func NewHuhQuickActions(theme themes.Theme, options []QuickActionOption) *HuhQuickActions {
	var selected string

	// Convert options to huh.Option
	huhOptions := make([]huh.Option[string], len(options))
	for i, opt := range options {
		huhOptions[i] = huh.NewOption(opt.Label, opt.Value)
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Key("action").
				Title("Choose action").
				Options(huhOptions...).
				Value(&selected),
		),
	).WithTheme(huhThemeFromPagenTheme(theme)).
		WithWidth(80)

	return &HuhQuickActions{
		theme:    theme,
		options:  options,
		selected: "",
		form:     form,
	}
}

// Init implements tea.Model
func (h *HuhQuickActions) Init() tea.Cmd {
	return h.form.Init()
}

// Update implements tea.Model
func (h *HuhQuickActions) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// Handle form updates
	form, formCmd := h.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		h.form = f
		cmd = formCmd
	}

	// Check if form is complete
	if h.form.State == huh.StateCompleted {
		h.selected = h.form.GetString("action")
	}

	return h, cmd
}

// View implements tea.Model
func (h *HuhQuickActions) View() string {
	return h.form.View()
}

// GetSelected returns the selected action value
func (h *HuhQuickActions) GetSelected() string {
	return h.selected
}

// SetSelected sets the selected action (for testing)
func (h *HuhQuickActions) SetSelected(value string) {
	h.selected = value
}

// IsComplete returns whether the form is complete
func (h *HuhQuickActions) IsComplete() bool {
	return h.form.State == huh.StateCompleted
}
