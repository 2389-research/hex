// ABOUTME: Panic recovery tests for HuhApproval component
// ABOUTME: Ensures approval form stays stable after panics
package components

import (
	"bytes"
	"log/slog"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/harper/pagent/internal/ui/themes"
	"github.com/stretchr/testify/assert"
)

// TestHuhApprovalUpdatePanicRecovery verifies panic recovery in Update
func TestHuhApprovalUpdatePanicRecovery(t *testing.T) {
	// Capture logs
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelError,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)

	theme := themes.NewDracula()
	approval := NewHuhApproval(theme, "bash", "Run: echo test")

	// Create a message that would cause a panic if form.Update panics
	// We'll simulate this by setting form to nil
	approval.form = nil

	// Attempt update - this would normally panic when calling nil.Update()
	result, cmd := approval.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Verify stable state returned
	assert.NotNil(t, result, "Should return stable state after panic")
	assert.Nil(t, cmd, "Command should be nil after panic")

	// Verify error was logged
	logOutput := buf.String()
	assert.Contains(t, logOutput, "Component panic recovered", "Should log recovery")
	assert.Contains(t, logOutput, "component=HuhApproval.Update", "Should log component name")
}

// TestHuhApprovalUpdateNormalExecution verifies normal operation still works
func TestHuhApprovalUpdateNormalExecution(t *testing.T) {
	theme := themes.NewDracula()
	approval := NewHuhApproval(theme, "bash", "Run: echo test")
	approval.Init()

	// Send a normal key message
	result, _ := approval.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})

	// Verify normal execution
	assert.NotNil(t, result, "Should return result")
	assert.IsType(t, &HuhApproval{}, result, "Should return HuhApproval")
}

// TestHuhApprovalUpdatePreservesStateAfterPanic verifies state preservation
func TestHuhApprovalUpdatePreservesStateAfterPanic(t *testing.T) {
	theme := themes.NewDracula()
	approval := NewHuhApproval(theme, "bash", "Run: echo test")

	// Set approval state
	approval.SetApproved(true)
	initialApproved := approval.IsApproved()

	// Store initial values
	initialToolName := approval.toolName
	initialDescription := approval.description

	// Induce panic by setting form to nil
	approval.form = nil

	// Trigger panic
	result, _ := approval.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Verify state preserved
	resultApproval := result.(*HuhApproval)
	assert.Equal(t, initialApproved, resultApproval.IsApproved(), "Approval state should be preserved")
	assert.Equal(t, initialToolName, resultApproval.toolName, "Tool name should be preserved")
	assert.Equal(t, initialDescription, resultApproval.description, "Description should be preserved")
}

// TestHuhApprovalMultiplePanics verifies component can recover from multiple panics
func TestHuhApprovalMultiplePanics(t *testing.T) {
	// Capture logs
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelError,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)

	theme := themes.NewDracula()
	approval := NewHuhApproval(theme, "bash", "Run: echo test")

	// Induce first panic
	approval.form = nil
	result1, _ := approval.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.NotNil(t, result1, "Should recover from first panic")

	// Induce second panic
	result2, _ := approval.Update(tea.KeyMsg{Type: tea.KeySpace})
	assert.NotNil(t, result2, "Should recover from second panic")

	// Verify both panics were logged
	logOutput := buf.String()
	// Count occurrences of "Component panic recovered"
	count := 0
	for i := 0; i < len(logOutput); i++ {
		if i+25 <= len(logOutput) && logOutput[i:i+25] == "Component panic recovered" {
			count++
		}
	}
	assert.GreaterOrEqual(t, count, 2, "Should log multiple panic recoveries")
}

// TestRecoverPanicFunction tests the local recoverPanic function
func TestRecoverPanicFunction(t *testing.T) {
	// Capture logs
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelError,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)

	// Test that recoverPanic catches panics
	didExecute := false

	recoverPanic("TestComponent", func() {
		didExecute = true
		panic("test panic")
	})

	assert.True(t, didExecute, "Function should execute before panic")

	// Verify error was logged
	logOutput := buf.String()
	assert.Contains(t, logOutput, "Component panic recovered", "Should log recovery")
	assert.Contains(t, logOutput, "component=TestComponent", "Should log component name")
	assert.Contains(t, logOutput, "test panic", "Should log panic value")
}

// TestRecoverPanicNoPanic verifies recoverPanic doesn't interfere with normal execution
func TestRecoverPanicNoPanic(t *testing.T) {
	executed := false

	recoverPanic("TestComponent", func() {
		executed = true
	})

	assert.True(t, executed, "Function should execute normally")
}
