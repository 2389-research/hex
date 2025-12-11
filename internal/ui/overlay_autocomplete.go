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

// GetHeader returns the header content
func (o *AutocompleteOverlay) GetHeader() string {
	return ""
}

// GetContent returns the main content
func (o *AutocompleteOverlay) GetContent() string {
	return o.Render(0, 0)
}

// GetFooter returns the footer content
func (o *AutocompleteOverlay) GetFooter() string {
	return ""
}

// GetDesiredHeight returns the desired height for this overlay
func (o *AutocompleteOverlay) GetDesiredHeight() int {
	return 10
}

// OnPush is called when the overlay is pushed onto the stack
func (o *AutocompleteOverlay) OnPush(width, height int) {}

// OnPop is called when the overlay is popped from the stack
func (o *AutocompleteOverlay) OnPop() {}

// Render returns the autocomplete dropdown UI
func (o *AutocompleteOverlay) Render(width, height int) string {
	return o.model.renderAutocompleteDropdown()
}

// HandleKey processes key presses for autocomplete
func (o *AutocompleteOverlay) HandleKey(msg tea.KeyMsg) (bool, tea.Cmd) {
	// Autocomplete navigation is already handled in main Update
	return false, nil
}

// Cancel dismisses the autocomplete
func (o *AutocompleteOverlay) Cancel() {
	if o.model.autocomplete != nil {
		o.model.autocomplete.Hide()
	}
}
