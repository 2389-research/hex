// Package forms provides beautiful huh-based forms for the hex TUI.
// ABOUTME: Huh-based quick actions menu for keyboard-driven command palette
// ABOUTME: Provides beautiful select-based action UI with fuzzy search
package forms

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/harper/hex/internal/ui/theme"
)

// QuickAction represents a single quick action
type QuickAction struct {
	Name        string
	Description string
	Category    string // Tools, Navigation, Settings
	KeyBinding  string // Optional keyboard shortcut
	Handler     func(args string) error
}

// QuickActionCategory represents a category of actions
type QuickActionCategory string

const (
	// CategoryTools for tool-related actions
	CategoryTools QuickActionCategory = "Tools"
	// CategoryNavigation for navigation actions
	CategoryNavigation QuickActionCategory = "Navigation"
	// CategorySettings for settings and configuration
	CategorySettings QuickActionCategory = "Settings"
)

// QuickActionsForm creates a beautiful huh form for quick actions
type QuickActionsForm struct {
	actions      []*QuickAction
	selectedName string
	theme        *theme.Theme
}

// NewQuickActionsForm creates a new quick actions form
func NewQuickActionsForm(actions []*QuickAction) *QuickActionsForm {
	return &QuickActionsForm{
		actions: actions,
		theme:   theme.DraculaTheme(),
	}
}

// Run displays the form and returns the selected action name
func (f *QuickActionsForm) Run() (string, error) {
	if len(f.actions) == 0 {
		return "", fmt.Errorf("no actions available")
	}

	// Group actions by category
	categorized := f.categorizeActions()

	// Build options grouped by category
	options := make([]huh.Option[string], 0)

	// Add each category with a separator
	for _, category := range []QuickActionCategory{CategoryTools, CategoryNavigation, CategorySettings} {
		if acts, ok := categorized[category]; ok && len(acts) > 0 {
			// Add category header (disabled option for visual separation)
			options = append(options, huh.NewOption(
				f.theme.Muted.Render(fmt.Sprintf("─── %s ───", category)),
				"",
			).Selected(false))

			// Add actions in this category
			for _, action := range acts {
				label := f.formatActionLabel(action)
				options = append(options, huh.NewOption(label, action.Name))
			}
		}
	}

	// Create the form with Dracula theme
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("⚡ Quick Actions").
				Description("Select an action or start typing to filter").
				Options(options...).
				Value(&f.selectedName).
				Height(15).
				Filtering(true),
		),
	).WithTheme(f.getDraculaTheme())

	// Run the form
	err := form.Run()
	if err != nil {
		return "", err
	}

	// If empty string was selected (category header), return error
	if f.selectedName == "" {
		return "", fmt.Errorf("no action selected")
	}

	return f.selectedName, nil
}

// categorizeActions groups actions by category
func (f *QuickActionsForm) categorizeActions() map[QuickActionCategory][]*QuickAction {
	categorized := make(map[QuickActionCategory][]*QuickAction)

	for _, action := range f.actions {
		category := QuickActionCategory(action.Category)
		if category == "" {
			category = CategoryTools // Default category
		}
		categorized[category] = append(categorized[category], action)
	}

	return categorized
}

// formatActionLabel formats an action for display with description and key binding
func (f *QuickActionsForm) formatActionLabel(action *QuickAction) string {
	var parts []string

	// Action name (cyan)
	parts = append(parts, f.theme.Emphasized.Render(action.Name))

	// Description (muted)
	if action.Description != "" {
		parts = append(parts, f.theme.Muted.Render("- "+action.Description))
	}

	// Key binding (pink, in parentheses)
	if action.KeyBinding != "" {
		parts = append(parts, f.theme.Subtitle.Render(fmt.Sprintf("[%s]", action.KeyBinding)))
	}

	return strings.Join(parts, " ")
}

// getDraculaTheme returns a huh theme configured with Dracula colors
func (f *QuickActionsForm) getDraculaTheme() *huh.Theme {
	t := huh.ThemeBase()

	// Use the theme instance colors
	colors := f.theme.Colors

	// Configure with Dracula colors
	t.Focused.Base = t.Focused.Base.
		BorderForeground(colors.Purple)

	t.Focused.Title = t.Focused.Title.
		Foreground(colors.Purple).
		Bold(true)

	t.Focused.Description = t.Focused.Description.
		Foreground(colors.Foreground)

	t.Focused.SelectSelector = t.Focused.SelectSelector.
		Foreground(colors.Pink)

	t.Focused.SelectedOption = t.Focused.SelectedOption.
		Foreground(colors.Cyan).
		Bold(true)

	t.Focused.UnselectedOption = t.Focused.UnselectedOption.
		Foreground(colors.Comment)

	t.Focused.FocusedButton = t.Focused.FocusedButton.
		Foreground(colors.Background).
		Background(colors.Purple).
		Bold(true)

	t.Focused.BlurredButton = t.Focused.BlurredButton.
		Foreground(colors.Foreground).
		Background(colors.CurrentLine)

	// Filter/search input styling
	t.Focused.TextInput.Cursor = t.Focused.TextInput.Cursor.
		Foreground(colors.Pink)

	t.Focused.TextInput.Placeholder = t.Focused.TextInput.Placeholder.
		Foreground(colors.Comment)

	t.Focused.TextInput.Prompt = t.Focused.TextInput.Prompt.
		Foreground(colors.Purple)

	return t
}
