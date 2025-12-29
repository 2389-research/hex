package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// AutocompleteProvider provides content for the autocomplete overlay.
// It implements BottomContentProvider, BottomKeyHandler, BottomActivationHandler, and BottomCancelHandler.
type AutocompleteProvider struct {
	model *Model
}

// Header implements BottomContentProvider
func (p *AutocompleteProvider) Header() string {
	// Autocomplete doesn't need a header
	return ""
}

// Content implements BottomContentProvider
func (p *AutocompleteProvider) Content() string {
	if p.model.autocomplete == nil || !p.model.autocomplete.IsActive() {
		return ""
	}

	completions := p.model.autocomplete.GetCompletions()
	if len(completions) == 0 {
		return ""
	}

	var content strings.Builder

	selectedIndex := p.model.autocomplete.GetSelectedIndex()
	selectedStyle := p.model.theme.AutocompleteSelected
	normalStyle := p.model.theme.AutocompleteItem

	// Description style - slightly dimmer than main text but still readable
	descStyle := lipgloss.NewStyle().Foreground(p.model.theme.Colors.Comment)

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

// Footer implements BottomContentProvider
func (p *AutocompleteProvider) Footer() string {
	return "↑↓: navigate • Enter: accept • Esc: cancel"
}

// DesiredHeight implements BottomContentProvider
func (p *AutocompleteProvider) DesiredHeight() int {
	if p.model.autocomplete == nil || !p.model.autocomplete.IsActive() {
		return 0
	}

	completions := p.model.autocomplete.GetCompletions()
	if len(completions) == 0 {
		return 0
	}

	// Calculate height: items + footer line + spacing
	itemHeight := len(completions)
	footerHeight := 2 // blank line + footer text
	totalHeight := itemHeight + footerHeight

	// Cap at 40% of screen height
	if p.model.Height > 0 {
		maxHeight := int(float64(p.model.Height) * 0.4)
		if totalHeight > maxHeight {
			return maxHeight
		}
	}

	return totalHeight
}

// OnActivate implements BottomActivationHandler
func (p *AutocompleteProvider) OnActivate(width, height int) {
	// No special initialization needed
}

// OnDeactivate implements BottomActivationHandler
func (p *AutocompleteProvider) OnDeactivate() {
	// Clear autocomplete state when dismissed
	if p.model.autocomplete != nil {
		p.model.autocomplete.Hide()
	}
}

// HandleKey implements BottomKeyHandler
func (p *AutocompleteProvider) HandleKey(msg tea.KeyMsg) (bool, tea.Cmd) {
	// Handle Escape and Ctrl+C to dismiss
	if msg.Type == tea.KeyEsc || msg.Type == tea.KeyCtrlC {
		return true, nil // Handled - caller should Pop
	}

	// Pass navigation keys to autocomplete model
	if p.model.autocomplete != nil {
		switch msg.Type {
		case tea.KeyUp:
			p.model.autocomplete.Previous()
			return true, nil
		case tea.KeyDown:
			p.model.autocomplete.Next()
			return true, nil
		case tea.KeyEnter, tea.KeyTab:
			// The actual completion logic is handled in update.go
			// Return false to let it through to the main handler
			return false, nil
		}
	}

	// Let typing through to filter autocomplete (letters, numbers, backspace, etc.)
	// Only capture the specific keys we handle above
	return false, nil
}

// Cancel implements BottomCancelHandler
func (p *AutocompleteProvider) Cancel() tea.Cmd {
	if p.model.autocomplete != nil {
		p.model.autocomplete.Hide()
	}
	return nil
}

// AutocompleteOverlay wraps GenericBottomOverlay for autocomplete.
type AutocompleteOverlay struct {
	*GenericBottomOverlay
	provider *AutocompleteProvider
}

// NewAutocompleteOverlay creates a new autocomplete overlay
func NewAutocompleteOverlay(m *Model) *AutocompleteOverlay {
	provider := &AutocompleteProvider{model: m}
	return &AutocompleteOverlay{
		GenericBottomOverlay: NewGenericBottomOverlay(provider, m.theme),
		provider:             provider,
	}
}
