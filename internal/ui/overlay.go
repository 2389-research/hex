// Package ui provides overlay management for modal/popup interfaces
package ui

import tea "github.com/charmbracelet/bubbletea"

// OverlayType represents different types of overlays
type OverlayType int

const (
	OverlayNone OverlayType = iota
	OverlayToolApproval
	OverlayQuickActions
	OverlaySearch
	OverlayAutocomplete
)

// Overlay represents a modal/popup interface that appears over the main content
type Overlay interface {
	// Type returns the overlay type
	Type() OverlayType

	// IsActive returns whether this overlay is currently shown
	IsActive() bool

	// Render returns the string content to display
	Render() string

	// HandleKey processes a key press and returns whether it was handled
	HandleKey(msg tea.KeyMsg) bool

	// Cancel closes/dismisses the overlay
	Cancel()

	// HandleEscape processes Escape key and returns optional command (nil = just cancel)
	HandleEscape() tea.Cmd

	// HandleCtrlC processes Ctrl+C and returns optional command (nil = just cancel)
	HandleCtrlC() tea.Cmd

	// Priority returns the precedence level (higher = shown on top)
	Priority() int
}

// OverlayManager manages multiple overlays with automatic priority handling
type OverlayManager struct {
	overlays []Overlay
}

// NewOverlayManager creates a new overlay manager
func NewOverlayManager() *OverlayManager {
	return &OverlayManager{
		overlays: make([]Overlay, 0),
	}
}

// Register adds an overlay to be managed
func (om *OverlayManager) Register(overlay Overlay) {
	om.overlays = append(om.overlays, overlay)
}

// GetActive returns the highest priority active overlay, or nil if none
func (om *OverlayManager) GetActive() Overlay {
	var active Overlay
	highestPriority := -1

	for _, overlay := range om.overlays {
		if overlay.IsActive() && overlay.Priority() > highestPriority {
			active = overlay
			highestPriority = overlay.Priority()
		}
	}

	return active
}

// HasActive returns whether any overlay is currently active
func (om *OverlayManager) HasActive() bool {
	return om.GetActive() != nil
}

// HandleKey passes a key event to the active overlay
// Returns true if the key was handled
func (om *OverlayManager) HandleKey(msg tea.KeyMsg) bool {
	active := om.GetActive()
	if active == nil {
		return false
	}
	return active.HandleKey(msg)
}

// HandleEscape cancels the active overlay if Escape is pressed
// Returns a command if the overlay needs to perform an action, or nil
func (om *OverlayManager) HandleEscape() tea.Cmd {
	active := om.GetActive()
	if active != nil {
		return active.HandleEscape()
	}
	return nil
}

// HandleCtrlC cancels the active overlay if Ctrl+C is pressed
// Returns a command if the overlay needs to perform an action, or nil
func (om *OverlayManager) HandleCtrlC() tea.Cmd {
	active := om.GetActive()
	if active != nil {
		return active.HandleCtrlC()
	}
	return nil
}

// Render returns the content of the highest priority active overlay
func (om *OverlayManager) Render() string {
	active := om.GetActive()
	if active == nil {
		return ""
	}
	return active.Render()
}

// CancelAll dismisses all active overlays
func (om *OverlayManager) CancelAll() {
	for _, overlay := range om.overlays {
		if overlay.IsActive() {
			overlay.Cancel()
		}
	}
}
