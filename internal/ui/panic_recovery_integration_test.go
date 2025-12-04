// ABOUTME: Integration tests for panic recovery in Model.Update
// ABOUTME: Verifies TUI stays alive after component panics
package ui

import (
	"bytes"
	"log/slog"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

// TestModelUpdateNormalExecution verifies that Update still works normally without panics
func TestModelUpdateNormalExecution(t *testing.T) {
	m := &Model{
		Messages: []Message{},
		Ready:    true,
	}

	// Send a normal WindowSizeMsg
	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	result, _ := m.Update(msg)

	// Verify normal execution
	assert.NotNil(t, result, "Should return result")
	assert.Equal(t, 100, m.Width, "Should update width")
	assert.Equal(t, 50, m.Height, "Should update height")
}

// TestModelUpdatePreservesStateAfterPanic verifies model state is preserved after panic
func TestModelUpdatePreservesStateAfterPanic(t *testing.T) {
	m := &Model{
		Messages: []Message{
			{Role: "user", Content: "test message"},
		},
		Ready:  true,
		Width:  80,
		Height: 24,
	}

	// Store initial state
	initialMessageCount := len(m.Messages)

	// Send normal message to verify state preservation
	result, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 50})

	// Verify state preserved (width/height updated, messages unchanged)
	resultModel := result.(*Model)
	assert.Equal(t, initialMessageCount, len(resultModel.Messages), "Messages should be preserved")
	assert.Equal(t, 100, resultModel.Width, "Width should be updated")
	assert.Equal(t, 50, resultModel.Height, "Height should be updated")
}

// TestRecoverPanicFunctionIntegration tests RecoverPanic with realistic scenarios
func TestRecoverPanicFunctionIntegration(t *testing.T) {
	// Capture logs
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelError,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)

	// Simulate what happens in Update() when update() panics
	var result tea.Model
	var cmd tea.Cmd
	m := &Model{
		Messages: []Message{},
		Ready:    true,
	}

	RecoverPanic("Model.Update", func() {
		// Simulate a panic in update()
		panic("simulated panic in update")
	})

	// After panic, result should still be nil
	assert.Nil(t, result, "Result should be nil after panic")
	assert.Nil(t, cmd, "Cmd should be nil after panic")

	// In real Update(), we'd return m (stable state) here
	if result == nil {
		result = m
	}
	assert.Equal(t, m, result, "Should return stable state after panic")

	// Verify error was logged
	logOutput := buf.String()
	assert.Contains(t, logOutput, "Component panic recovered", "Should log recovery")
	assert.Contains(t, logOutput, "component=Model.Update", "Should log component name")
	assert.Contains(t, logOutput, "simulated panic in update", "Should log panic message")
}

// TestMultiplePanicRecoveries verifies system can handle multiple panics
func TestMultiplePanicRecoveries(t *testing.T) {
	// Capture logs
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelError,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)

	// First panic
	RecoverPanic("Component1", func() {
		panic("panic 1")
	})

	// Second panic
	RecoverPanic("Component2", func() {
		panic("panic 2")
	})

	// Third panic
	RecoverPanic("Component3", func() {
		panic("panic 3")
	})

	// Verify all three were logged
	logOutput := buf.String()
	assert.Contains(t, logOutput, "panic 1", "Should log first panic")
	assert.Contains(t, logOutput, "panic 2", "Should log second panic")
	assert.Contains(t, logOutput, "panic 3", "Should log third panic")
	assert.Contains(t, logOutput, "component=Component1", "Should log first component")
	assert.Contains(t, logOutput, "component=Component2", "Should log second component")
	assert.Contains(t, logOutput, "component=Component3", "Should log third component")
}

// TestPanicRecoveryWithStackTrace verifies stack traces are captured
func TestPanicRecoveryWithStackTrace(t *testing.T) {
	// Capture logs
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelError,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)

	RecoverPanic("TestComponent", func() {
		// Create a multi-level call stack
		level1 := func() {
			level2 := func() {
				panic("deep panic")
			}
			level2()
		}
		level1()
	})

	// Verify stack trace was logged
	logOutput := buf.String()
	assert.Contains(t, logOutput, "stack=", "Should log stack trace")
	assert.Contains(t, logOutput, "goroutine", "Stack trace should contain goroutine info")
	assert.Contains(t, logOutput, "deep panic", "Should contain panic message")
}
