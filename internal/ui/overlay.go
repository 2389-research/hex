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

// OverlayManager manages a stack of overlays
type OverlayManager struct {
	stack []Overlay
}

// NewOverlayManager creates a new overlay manager
func NewOverlayManager() *OverlayManager {
	return &OverlayManager{
		stack: make([]Overlay, 0, 4), // Pre-allocate for common case
	}
}

// Push adds an overlay to the top of the stack
func (om *OverlayManager) Push(overlay Overlay) {
	om.stack = append(om.stack, overlay)
	// OnPush will be called by Model with width/height
}

// Pop removes and returns the top overlay from the stack
func (om *OverlayManager) Pop() Overlay {
	if len(om.stack) == 0 {
		return nil
	}
	overlay := om.stack[len(om.stack)-1]
	om.stack = om.stack[:len(om.stack)-1]
	overlay.OnPop()
	return overlay
}

// Peek returns the top overlay without removing it
func (om *OverlayManager) Peek() Overlay {
	if len(om.stack) == 0 {
		return nil
	}
	return om.stack[len(om.stack)-1]
}

// Clear removes all overlays from the stack
func (om *OverlayManager) Clear() {
	for len(om.stack) > 0 {
		om.Pop()
	}
}

// GetActive returns the top overlay on the stack, or nil if empty
func (om *OverlayManager) GetActive() Overlay {
	return om.Peek()
}

// HasActive returns whether any overlay is currently active
func (om *OverlayManager) HasActive() bool {
	return om.GetActive() != nil
}

// HandleKey passes a key event to the active overlay
// Returns (true, cmd) if key was handled, (false, nil) if no active overlay
func (om *OverlayManager) HandleKey(msg tea.KeyMsg) (bool, tea.Cmd) {
	active := om.GetActive()
	if active == nil {
		return false, nil
	}
	// Modal behavior: overlay always captures input
	return active.HandleKey(msg)
}

// HandleEscape is deprecated - use HandleKey instead
// Kept for backward compatibility during migration
func (om *OverlayManager) HandleEscape() tea.Cmd {
	active := om.GetActive()
	if active != nil {
		handled, cmd := active.HandleKey(tea.KeyMsg{Type: tea.KeyEsc})
		if handled {
			return cmd
		}
	}
	return nil
}

// HandleCtrlC is deprecated - use HandleKey instead
// Kept for backward compatibility during migration
func (om *OverlayManager) HandleCtrlC() tea.Cmd {
	active := om.GetActive()
	if active != nil {
		handled, cmd := active.HandleKey(tea.KeyMsg{Type: tea.KeyCtrlC})
		if handled {
			return cmd
		}
	}
	return nil
}

// Render returns the content of the top overlay
// Returns empty string if no active overlay
func (om *OverlayManager) Render(width, height int) string {
	active := om.GetActive()
	if active == nil {
		return ""
	}
	return active.Render(width, height)
}

// IsFullscreen returns true if the active overlay is fullscreen
func (om *OverlayManager) IsFullscreen() bool {
	active := om.GetActive()
	if active == nil {
		return false
	}
	if fs, ok := active.(FullscreenOverlay); ok {
		return fs.IsFullscreen()
	}
	return false
}

// CancelAll dismisses all active overlays
func (om *OverlayManager) CancelAll() {
	om.Clear()
}
