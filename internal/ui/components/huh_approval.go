// Package components provides reusable UI components for the Jeff TUI.
// ABOUTME: Huh-based tool approval component using confirm form
// ABOUTME: Provides themed approval dialogs for tool execution
package components

import (
	"fmt"
	"log/slog"
	"runtime/debug"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/harper/jeff/internal/ui/themes"
)

// HuhApproval is a Huh-based approval dialog for tool execution
type HuhApproval struct {
	theme       themes.Theme
	toolName    string
	description string
	approved    bool
	form        *huh.Form
	width       int
	height      int
}

// NewHuhApproval creates a new Huh approval dialog
func NewHuhApproval(theme themes.Theme, toolName, description string) *HuhApproval {
	var approved bool

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Key("approved").
				Title(fmt.Sprintf("Execute %s?", toolName)).
				Description(description).
				Affirmative("Yes!").
				Negative("No.").
				Value(&approved),
		),
	).WithTheme(huhThemeFromJeffTheme(theme)).
		WithWidth(80)

	return &HuhApproval{
		theme:       theme,
		toolName:    toolName,
		description: description,
		approved:    false,
		form:        form,
	}
}

// Init implements tea.Model
func (h *HuhApproval) Init() tea.Cmd {
	return h.form.Init()
}

// Update implements tea.Model with panic recovery
func (h *HuhApproval) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var result tea.Model = h
	var cmd tea.Cmd

	recoverPanic("HuhApproval.Update", func() {
		// Handle form updates
		form, formCmd := h.form.Update(msg)
		if f, ok := form.(*huh.Form); ok {
			h.form = f
			cmd = formCmd
		}

		// Check if form is complete
		if h.form.State == huh.StateCompleted {
			// Extract the approved value from form
			// Huh confirm stores boolean values, not strings
			if val := h.form.GetBool("approved"); val {
				h.approved = true
			} else {
				h.approved = false
			}
		}

		result = h
	})

	return result, cmd
}

// recoverPanic wraps a function with panic recovery for components
func recoverPanic(component string, fn func()) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("Component panic recovered",
				"component", component,
				"panic", r,
				"stack", string(debug.Stack()))
		}
	}()
	fn()
}

// View implements tea.Model
func (h *HuhApproval) View() string {
	return h.form.View()
}

// IsApproved returns whether the tool was approved
func (h *HuhApproval) IsApproved() bool {
	return h.approved
}

// SetApproved sets the approval state (for testing)
func (h *HuhApproval) SetApproved(approved bool) {
	h.approved = approved
}

// IsComplete returns whether the form is complete
func (h *HuhApproval) IsComplete() bool {
	return h.form.State == huh.StateCompleted
}

// SetSize implements the Sizeable interface
func (h *HuhApproval) SetSize(width, height int) tea.Cmd {
	h.width = width
	h.height = height
	if h.form != nil {
		h.form = h.form.WithWidth(width)
	}
	return nil
}

// GetSize implements the Sizeable interface
func (h *HuhApproval) GetSize() (int, int) {
	return h.width, h.height
}

// huhThemeFromJeffTheme converts Jeff theme to Huh theme
func huhThemeFromJeffTheme(theme themes.Theme) *huh.Theme {
	t := huh.ThemeBase()

	// Customize focused field styles
	t.Focused.Base = t.Focused.Base.Foreground(theme.Foreground())
	t.Focused.Title = t.Focused.Title.Foreground(theme.Primary())
	t.Focused.Description = t.Focused.Description.Foreground(theme.Subtle())
	t.Focused.ErrorIndicator = t.Focused.ErrorIndicator.Foreground(theme.Error())
	t.Focused.ErrorMessage = t.Focused.ErrorMessage.Foreground(theme.Error())
	t.Focused.SelectSelector = t.Focused.SelectSelector.Foreground(theme.Primary())
	t.Focused.NextIndicator = t.Focused.NextIndicator.Foreground(theme.Success())
	t.Focused.PrevIndicator = t.Focused.PrevIndicator.Foreground(theme.Warning())
	t.Focused.Option = t.Focused.Option.Foreground(theme.Foreground())
	t.Focused.SelectedOption = t.Focused.SelectedOption.Foreground(theme.Primary())
	t.Focused.FocusedButton = t.Focused.FocusedButton.Foreground(theme.Background()).Background(theme.Primary())
	t.Focused.BlurredButton = t.Focused.BlurredButton.Foreground(theme.Subtle())

	// Customize blurred field styles
	t.Blurred.Base = t.Blurred.Base.Foreground(theme.Subtle())
	t.Blurred.Title = t.Blurred.Title.Foreground(theme.Subtle())
	t.Blurred.Description = t.Blurred.Description.Foreground(theme.Subtle())
	t.Blurred.SelectSelector = t.Blurred.SelectSelector.Foreground(theme.Subtle())
	t.Blurred.NextIndicator = t.Blurred.NextIndicator.Foreground(theme.Subtle())
	t.Blurred.PrevIndicator = t.Blurred.PrevIndicator.Foreground(theme.Subtle())
	t.Blurred.Option = t.Blurred.Option.Foreground(theme.Subtle())
	t.Blurred.SelectedOption = t.Blurred.SelectedOption.Foreground(theme.Subtle())

	return t
}
