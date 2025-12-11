package ui

import tea "github.com/charmbracelet/bubbletea"

// AutocompleteOverlay implements the Overlay interface for autocomplete dropdown
type AutocompleteOverlay struct {
	model *Model
}

// NewAutocompleteOverlay creates a new autocomplete overlay
func NewAutocompleteOverlay(m *Model) *AutocompleteOverlay {
	return &AutocompleteOverlay{model: m}
}

// Type returns the overlay type
func (o *AutocompleteOverlay) Type() OverlayType {
	return OverlayAutocomplete
}

// IsActive returns whether autocomplete is currently shown
func (o *AutocompleteOverlay) IsActive() bool {
	return o.model.autocomplete != nil && o.model.autocomplete.IsActive()
}

// Render returns the autocomplete dropdown UI
func (o *AutocompleteOverlay) Render() string {
	return o.model.renderAutocompleteDropdown()
}

// HandleKey processes key presses for autocomplete
func (o *AutocompleteOverlay) HandleKey(msg tea.KeyMsg) bool {
	// Autocomplete navigation is already handled in main Update
	return false
}

// Cancel dismisses the autocomplete
func (o *AutocompleteOverlay) Cancel() {
	if o.model.autocomplete != nil {
		o.model.autocomplete.Hide()
	}
}

// HandleEscape dismisses autocomplete (no command needed)
func (o *AutocompleteOverlay) HandleEscape() tea.Cmd {
	o.Cancel()
	return nil
}

// HandleCtrlC dismisses autocomplete (no command needed)
func (o *AutocompleteOverlay) HandleCtrlC() tea.Cmd {
	o.Cancel()
	return nil
}

// Priority returns the precedence level
func (o *AutocompleteOverlay) Priority() int {
	return 50 // Medium priority - helpful but not critical
}
