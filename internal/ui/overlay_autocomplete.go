package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"
)

// AutocompleteOverlay implements the Overlay interface for autocomplete dropdown
type AutocompleteOverlay struct {
	model *Model
}

// NewAutocompleteOverlay creates a new autocomplete overlay
func NewAutocompleteOverlay(m *Model) *AutocompleteOverlay {
	return &AutocompleteOverlay{model: m}
}

// GetHeader returns the header content
func (o *AutocompleteOverlay) GetHeader() string {
	// Autocomplete doesn't need a header
	return ""
}

// GetContent returns the main content
func (o *AutocompleteOverlay) GetContent() string {
	if o.model.autocomplete == nil || !o.model.autocomplete.IsActive() {
		return ""
	}

	completions := o.model.autocomplete.GetCompletions()
	if len(completions) == 0 {
		return ""
	}

	var content strings.Builder

	selectedIndex := o.model.autocomplete.GetSelectedIndex()
	selectedStyle := o.model.theme.AutocompleteSelected
	normalStyle := o.model.theme.AutocompleteItem

	// Description style - slightly dimmer than main text but still readable
	descStyle := lipgloss.NewStyle().Foreground(o.model.theme.Colors.Comment)

	for i, completion := range completions {
		var line strings.Builder

		// Selection indicator and styling
		if i == selectedIndex {
			line.WriteString("▸ ")
			line.WriteString(completion.Display)

			// Add description if available
			if completion.Description != "" {
				line.WriteString(" (" + completion.Description + ")")
			}

			content.WriteString(selectedStyle.Render(line.String()))
		} else {
			line.WriteString("  ")
			line.WriteString(completion.Display)

			// Add description if available (dimmed but readable)
			if completion.Description != "" {
				line.WriteString(" ")
				line.WriteString(descStyle.Render("(" + completion.Description + ")"))
			}

			content.WriteString(normalStyle.Render(line.String()))
		}

		content.WriteString("\n")
	}

	return content.String()
}

// GetFooter returns the footer content
func (o *AutocompleteOverlay) GetFooter() string {
	return "↑↓: navigate • Enter: accept • Esc: cancel"
}

// GetDesiredHeight returns the desired height for this overlay
func (o *AutocompleteOverlay) GetDesiredHeight() int {
	if o.model.autocomplete == nil || !o.model.autocomplete.IsActive() {
		return 0
	}

	completions := o.model.autocomplete.GetCompletions()
	if len(completions) == 0 {
		return 0
	}

	// Calculate height: items + footer line + spacing
	itemHeight := len(completions)
	footerHeight := 2 // blank line + footer text
	totalHeight := itemHeight + footerHeight

	// Cap at 40% of screen height
	if o.model.Height > 0 {
		maxHeight := int(float64(o.model.Height) * 0.4)
		if totalHeight > maxHeight {
			return maxHeight
		}
	}

	return totalHeight
}

// OnPush is called when the overlay is pushed onto the stack
func (o *AutocompleteOverlay) OnPush(width, height int) {
	// No special initialization needed
}

// OnPop is called when the overlay is popped from the stack
func (o *AutocompleteOverlay) OnPop() {
	// Clear autocomplete state when dismissed
	if o.model.autocomplete != nil {
		o.model.autocomplete.Hide()
	}
}

// Render returns the complete overlay rendering
func (o *AutocompleteOverlay) Render(width, height int) string {
	if o.model.autocomplete == nil || !o.model.autocomplete.IsActive() {
		return ""
	}

	completions := o.model.autocomplete.GetCompletions()
	if len(completions) == 0 {
		return ""
	}

	var b strings.Builder

	// Use full width minus some padding for the dropdown
	dropdownWidth := width - 4
	if dropdownWidth < 40 {
		dropdownWidth = 40
	}
	boxStyle := o.model.theme.AutocompleteDropdown.Width(dropdownWidth)
	helpStyle := o.model.theme.AutocompleteHelp

	// Header (autocomplete doesn't use header, but keep pattern consistent)
	if header := o.GetHeader(); header != "" {
		b.WriteString(header)
		b.WriteString("\n")
	}

	// Content
	b.WriteString(o.GetContent())

	// Footer
	b.WriteString("\n")
	if footer := o.GetFooter(); footer != "" {
		b.WriteString(helpStyle.Render(footer))
	}

	return boxStyle.Render(b.String())
}

// HandleKey processes key presses for autocomplete
func (o *AutocompleteOverlay) HandleKey(msg tea.KeyMsg) (bool, tea.Cmd) {
	// Handle Escape and Ctrl+C to dismiss
	if msg.Type == tea.KeyEsc || msg.Type == tea.KeyCtrlC {
		return true, nil // Handled - caller should Pop
	}

	// Pass navigation keys to autocomplete model
	if o.model.autocomplete != nil {
		switch msg.Type {
		case tea.KeyUp:
			o.model.autocomplete.Previous()
			return true, nil
		case tea.KeyDown:
			o.model.autocomplete.Next()
			return true, nil
		case tea.KeyEnter, tea.KeyTab:
			// The actual completion logic is handled in update.go
			// Return false to let it through to the main handler
			return false, nil
		}
	}

	// Modal: capture all other input to prevent leakage
	return true, nil
}

// Cancel dismisses the autocomplete
func (o *AutocompleteOverlay) Cancel() {
	if o.model.autocomplete != nil {
		o.model.autocomplete.Hide()
	}
}
