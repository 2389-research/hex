// ABOUTME: Test suite for panic recovery utilities
// ABOUTME: Verifies panics are caught and logged without crashing TUI
package ui

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRecoverPanic_NoPanic(t *testing.T) {
	// Test that RecoverPanic doesn't interfere with normal execution
	executed := false

	RecoverPanic("test", func() {
		executed = true
	})

	assert.True(t, executed, "Function should execute normally")
}

func TestRecoverPanic_CatchesPanic(t *testing.T) {
	// Capture logs
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelError,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)

	// Test that RecoverPanic catches panics
	didPanic := false

	RecoverPanic("test", func() {
		didPanic = true
		panic("test panic")
	})

	assert.True(t, didPanic, "Function should execute before panic")

	// Verify error was logged
	logOutput := buf.String()
	assert.Contains(t, logOutput, "Component panic recovered", "Should log recovery message")
	assert.Contains(t, logOutput, "component=test", "Should log component name")
	assert.Contains(t, logOutput, "test panic", "Should log panic value")
}

func TestRecoverPanic_LogsStackTrace(t *testing.T) {
	// Capture logs
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelError,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)

	RecoverPanic("test", func() {
		panic("test panic")
	})

	// Verify stack trace was logged
	logOutput := buf.String()
	assert.Contains(t, logOutput, "stack=", "Should log stack trace")
	assert.Contains(t, logOutput, "goroutine", "Stack trace should contain goroutine info")
}

func TestRecoverPanic_MultipleComponentNames(t *testing.T) {
	// Capture logs
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelError,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)

	// Test with different component names
	RecoverPanic("Component1", func() {
		panic("error 1")
	})

	logOutput := buf.String()
	assert.Contains(t, logOutput, "component=Component1", "Should log correct component name")

	buf.Reset()

	RecoverPanic("Component2", func() {
		panic("error 2")
	})

	logOutput = buf.String()
	assert.Contains(t, logOutput, "component=Component2", "Should log correct component name")
}

func TestRecoverPanic_DifferentPanicTypes(t *testing.T) {
	// Test with string panic
	RecoverPanic("test", func() {
		panic("string panic")
	})

	// Test with error panic
	RecoverPanic("test", func() {
		panic(assert.AnError)
	})

	// Test with int panic
	RecoverPanic("test", func() {
		panic(42)
	})

	// Test with struct panic
	RecoverPanic("test", func() {
		panic(struct{ msg string }{"structured panic"})
	})

	// If we get here, all panic types were recovered
	assert.True(t, true, "All panic types should be recovered")
}
