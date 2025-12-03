// Package components provides reusable UI components for the TUI.
// ABOUTME: Smooth state transition utilities for UI animations
// ABOUTME: Provides fade-in, slide-in effects for components
package components

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// TransitionState represents animation state
type TransitionState int

const (
	// TransitionIdle indicates no animation is active
	TransitionIdle TransitionState = iota
	// TransitionFadingIn indicates fade-in animation is in progress
	TransitionFadingIn
	// TransitionFadingOut indicates fade-out animation is in progress
	TransitionFadingOut
	// TransitionComplete indicates animation has finished
	TransitionComplete
)

// FadeTransition handles fade-in/out animations
type FadeTransition struct {
	state     TransitionState
	opacity   float64 // 0.0 to 1.0
	duration  time.Duration
	startTime time.Time
}

// NewFadeTransition creates a new fade transition
func NewFadeTransition(duration time.Duration) *FadeTransition {
	return &FadeTransition{
		state:    TransitionIdle,
		opacity:  0.0,
		duration: duration,
	}
}

// FadeIn starts fade-in animation
func (f *FadeTransition) FadeIn() tea.Cmd {
	f.state = TransitionFadingIn
	f.startTime = time.Now()
	f.opacity = 0.0
	return f.tick()
}

// tick updates animation frame
func (f *FadeTransition) tick() tea.Cmd {
	return tea.Tick(16*time.Millisecond, func(t time.Time) tea.Msg {
		return transitionTickMsg{time: t}
	})
}

type transitionTickMsg struct {
	time time.Time
}

// Update updates the transition state
func (f *FadeTransition) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case transitionTickMsg:
		if f.state == TransitionFadingIn {
			elapsed := msg.time.Sub(f.startTime)
			progress := float64(elapsed) / float64(f.duration)

			if progress >= 1.0 {
				f.opacity = 1.0
				f.state = TransitionComplete
				return nil
			}

			f.opacity = progress
			return f.tick()
		}
	}
	return nil
}

// GetOpacity returns current opacity (0.0 to 1.0)
func (f *FadeTransition) GetOpacity() float64 {
	return f.opacity
}

// IsComplete returns whether transition is done
func (f *FadeTransition) IsComplete() bool {
	return f.state == TransitionComplete
}
