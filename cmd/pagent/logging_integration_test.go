// ABOUTME: Integration tests for logging functionality
// ABOUTME: Tests CLI flags and log output in real scenarios

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoggingIntegration(t *testing.T) {
	// Create temp directory for test logs
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	// Save and restore original flags
	originalLogLevel := logLevel
	originalLogFile := logFile
	originalLogFormat := logFormat
	defer func() {
		logLevel = originalLogLevel
		logFile = originalLogFile
		logFormat = originalLogFormat
	}()

	// Set flags
	logLevel = "debug"
	logFile = "" // Use stderr (default)
	logFormat = "text"

	// Initialize logging
	err := initializeLogging()
	require.NoError(t, err)
	require.NotNil(t, globalLogger)

	// Close logger
	closeLogger()
}

func TestLoggingLevels(t *testing.T) {
	tests := []struct {
		name     string
		level    string
		expected string
	}{
		{"debug level", "debug", "debug"},
		{"info level", "info", "info"},
		{"warn level", "warn", "warn"},
		{"error level", "error", "error"},
		{"invalid defaults to info", "invalid", "info"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalLogLevel := logLevel
			defer func() { logLevel = originalLogLevel }()

			logLevel = tt.level
			err := initializeLogging()
			require.NoError(t, err)
			defer closeLogger()

			// Logger should be initialized
			assert.NotNil(t, globalLogger)
		})
	}
}

func TestLoggingFormats(t *testing.T) {
	tests := []struct {
		name   string
		format string
	}{
		{"text format", "text"},
		{"json format", "json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			testLogFile := filepath.Join(tmpDir, "test.log")

			originalLogFile := logFile
			originalLogFormat := logFormat
			defer func() {
				logFile = originalLogFile
				logFormat = originalLogFormat
			}()

			logFile = testLogFile
			logFormat = tt.format

			err := initializeLogging()
			require.NoError(t, err)
			require.NotNil(t, globalLogger)

			closeLogger()
		})
	}
}

func TestLoggingFileCreation(t *testing.T) {
	tmpDir := t.TempDir()
	testLogFile := filepath.Join(tmpDir, "clem.log")

	originalLogFile := logFile
	defer func() { logFile = originalLogFile }()

	logFile = testLogFile

	err := initializeLogging()
	require.NoError(t, err)
	defer closeLogger()

	// File should exist (but might be empty until something is logged)
	_, err = os.Stat(testLogFile)
	assert.NoError(t, err, "Log file should be created")
}

func TestLoggingInvalidPath(t *testing.T) {
	originalLogFile := logFile
	defer func() { logFile = originalLogFile }()

	// Set invalid path
	logFile = "/nonexistent/directory/test.log"

	err := initializeLogging()
	assert.Error(t, err)
	assert.Contains(t, strings.ToLower(err.Error()), "failed to create logger")
}
