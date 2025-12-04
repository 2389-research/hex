// Package ui provides the Bubble Tea terminal user interface components.
// ABOUTME: Panic recovery utilities for TUI components
// ABOUTME: Prevents component crashes from taking down the entire application
package ui

import (
	"log/slog"
	"runtime/debug"
)

// RecoverPanic wraps a function with panic recovery
// If a panic occurs, it logs the error with component name, panic value, and stack trace
func RecoverPanic(component string, fn func()) {
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
