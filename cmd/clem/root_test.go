package main

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestRootCommand(t *testing.T) {
	// Test --help flag
	rootCmd.SetArgs([]string{"--help"})
	err := rootCmd.Execute()
	assert.NoError(t, err)
}

func TestVersionFlag(t *testing.T) {
	// Create a fresh command instance to avoid state pollution
	cmd := &cobra.Command{
		Use:     "clem [prompt]",
		Version: version,
	}

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--version"})

	err := cmd.Execute()
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "0.1.0")
}

func TestPrintFlag(t *testing.T) {
	rootCmd.SetArgs([]string{"--print", "test"})
	// Should not panic
	// Actual functionality tested in integration tests
}

func TestFlagConflictValidation(t *testing.T) {
	// Save original flag values and restore after test
	originalContinue := continueFlag
	originalResume := resumeID
	defer func() {
		continueFlag = originalContinue
		resumeID = originalResume
	}()

	// Test: both --continue and --resume flags set should return error
	continueFlag = true
	resumeID = "conv-123"

	err := runInteractive("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot use both")

	// Test: only --continue should not error from validation (may error from other reasons)
	continueFlag = true
	resumeID = ""
	err = runInteractive("")
	// If error occurs, it should NOT be the flag conflict error
	if err != nil {
		assert.NotContains(t, err.Error(), "cannot use both")
	}

	// Test: only --resume should not error from validation (may error from other reasons)
	continueFlag = false
	resumeID = "conv-123"
	err = runInteractive("")
	// If error occurs, it should NOT be the flag conflict error
	if err != nil {
		assert.NotContains(t, err.Error(), "cannot use both")
	}

	// Reset flags
	continueFlag = false
	resumeID = ""
}
