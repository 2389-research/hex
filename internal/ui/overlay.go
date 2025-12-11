// Package ui provides overlay management for modal/popup interfaces
package ui

import tea "github.com/charmbracelet/bubbletea"

// Overlay represents a modal interface that appears over the main content
type Overlay interface {
	// Structured rendering
	GetHeader() string
	GetContent() string
	GetFooter() string
	Render(width, height int) string

	// Input handling
	HandleKey(msg tea.KeyMsg) (handled bool, cmd tea.Cmd)

	// Lifecycle
	OnPush(width, height int)
	OnPop()

	// Height management
	GetDesiredHeight() int
}

// Scrollable adds viewport scrolling capability to any overlay
type Scrollable interface {
	Overlay
	Update(msg tea.Msg) tea.Cmd
}

// FullscreenOverlay represents a fullscreen modal with viewport
type FullscreenOverlay interface {
	Scrollable
	SetHeight(height int)
	IsFullscreen() bool
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
