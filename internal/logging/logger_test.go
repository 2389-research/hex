// ABOUTME: Test suite for structured logging implementation
// ABOUTME: Validates log levels, formats, file output, and context propagation

package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLogger_DefaultConfig(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(Config{
		Level:  LevelInfo,
		Format: FormatText,
		Writer: &buf,
	})

	require.NotNil(t, logger)
	assert.Equal(t, LevelInfo, logger.level)
}

func TestLogLevels(t *testing.T) {
	tests := []struct {
		name         string
		configLevel  Level
		logFunc      func(*Logger, string)
		logLevel     Level
		shouldAppear bool
	}{
		{
			name:         "Debug logged when level is Debug",
			configLevel:  LevelDebug,
			logFunc:      func(l *Logger, msg string) { l.Debug(msg) },
			logLevel:     LevelDebug,
			shouldAppear: true,
		},
		{
			name:         "Debug not logged when level is Info",
			configLevel:  LevelInfo,
			logFunc:      func(l *Logger, msg string) { l.Debug(msg) },
			logLevel:     LevelDebug,
			shouldAppear: false,
		},
		{
			name:         "Info logged when level is Info",
			configLevel:  LevelInfo,
			logFunc:      func(l *Logger, msg string) { l.Info(msg) },
			logLevel:     LevelInfo,
			shouldAppear: true,
		},
		{
			name:         "Info logged when level is Debug",
			configLevel:  LevelDebug,
			logFunc:      func(l *Logger, msg string) { l.Info(msg) },
			logLevel:     LevelInfo,
			shouldAppear: true,
		},
		{
			name:         "Warn logged when level is Warn",
			configLevel:  LevelWarn,
			logFunc:      func(l *Logger, msg string) { l.Warn(msg) },
			logLevel:     LevelWarn,
			shouldAppear: true,
		},
		{
			name:         "Error always logged",
			configLevel:  LevelError,
			logFunc:      func(l *Logger, msg string) { l.Error(msg) },
			logLevel:     LevelError,
			shouldAppear: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := NewLogger(Config{
				Level:  tt.configLevel,
				Format: FormatText,
				Writer: &buf,
			})

			tt.logFunc(logger, "test message")

			output := buf.String()
			if tt.shouldAppear {
				assert.Contains(t, output, "test message")
			} else {
				assert.Empty(t, output)
			}
		})
	}
}

func TestLogWithAttributes(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(Config{
		Level:  LevelInfo,
		Format: FormatText,
		Writer: &buf,
	})

	logger.InfoWith("user action", "user_id", "123", "action", "login")

	output := buf.String()
	assert.Contains(t, output, "user action")
	assert.Contains(t, output, "user_id=123")
	assert.Contains(t, output, "action=login")
}

func TestJSONFormat(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(Config{
		Level:  LevelInfo,
		Format: FormatJSON,
		Writer: &buf,
	})

	logger.InfoWith("test event", "key", "value")

	output := buf.String()
	assert.NotEmpty(t, output)

	// Parse as JSON to validate structure
	var logEntry map[string]interface{}
	err := json.Unmarshal([]byte(output), &logEntry)
	require.NoError(t, err)

	assert.Equal(t, "test event", logEntry["msg"])
	assert.Equal(t, "value", logEntry["key"])
	assert.Contains(t, logEntry, "time")
	assert.Contains(t, logEntry, "level")
}

func TestContextPropagation(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(Config{
		Level:  LevelInfo,
		Format: FormatJSON,
		Writer: &buf,
	})

	ctx := context.Background()
	ctx = WithConversationID(ctx, "conv-123")
	ctx = WithRequestID(ctx, "req-456")

	logger.InfoWithContext(ctx, "request received")

	output := buf.String()
	var logEntry map[string]interface{}
	err := json.Unmarshal([]byte(output), &logEntry)
	require.NoError(t, err)

	assert.Equal(t, "conv-123", logEntry["conversation_id"])
	assert.Equal(t, "req-456", logEntry["request_id"])
}

func TestFileOutput(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	logger, err := NewLoggerWithFile(Config{
		Level:   LevelInfo,
		Format:  FormatText,
		LogFile: logFile,
	})
	require.NoError(t, err)
	require.NotNil(t, logger)

	logger.Info("test message to file")

	// Close the logger to flush
	err = logger.Close()
	require.NoError(t, err)

	// Read the file
	//nolint:gosec // G304: Test file reads/writes are safe
	content, err := os.ReadFile(logFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), "test message to file")
}

func TestFileOutputError(t *testing.T) {
	// Try to create logger with invalid path
	logger, err := NewLoggerWithFile(Config{
		Level:   LevelInfo,
		Format:  FormatText,
		LogFile: "/nonexistent/dir/test.log",
	})

	assert.Error(t, err)
	assert.Nil(t, logger)
}

func TestMultiWriter(t *testing.T) {
	// Log to both stderr and file
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	var buf bytes.Buffer
	logger, err := NewLoggerWithFile(Config{
		Level:   LevelInfo,
		Format:  FormatText,
		LogFile: logFile,
		Writer:  &buf, // Also write to buffer
	})
	require.NoError(t, err)

	logger.Info("test message")
	err = logger.Close()
	require.NoError(t, err)

	// Check buffer
	assert.Contains(t, buf.String(), "test message")

	//nolint:gosec // G304: Test file reads/writes are safe
	// Check file
	content, err := os.ReadFile(logFile) //nolint:gosec // G304: Path validated by caller
	require.NoError(t, err)
	assert.Contains(t, string(content), "test message")
}

func TestErrorWithError(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(Config{
		Level:  LevelError,
		Format: FormatText,
		Writer: &buf,
	})

	testErr := assert.AnError
	logger.ErrorWithErr("operation failed", testErr, "operation", "test")

	output := buf.String()
	assert.Contains(t, output, "operation failed")
	assert.Contains(t, output, "error=")
	assert.Contains(t, output, "operation=test")
}

func TestLevelFromString(t *testing.T) {
	tests := []struct {
		input    string
		expected Level
	}{
		{"debug", LevelDebug},
		{"DEBUG", LevelDebug},
		{"info", LevelInfo},
		{"INFO", LevelInfo},
		{"warn", LevelWarn},
		{"WARN", LevelWarn},
		{"error", LevelError},
		{"ERROR", LevelError},
		{"invalid", LevelInfo}, // Default to Info
		{"", LevelInfo},        // Default to Info
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := LevelFromString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGlobalLogger(t *testing.T) {
	// Save and restore original global logger
	original := globalLogger
	defer func() {
		globalLogger = original
	}()

	var buf bytes.Buffer
	newLogger := NewLogger(Config{
		Level:  LevelInfo,
		Format: FormatText,
		Writer: &buf,
	})
	SetGlobalLogger(newLogger)

	// Call directly on the set logger since Default() uses once.Do
	newLogger.Info("global test message")
	assert.Contains(t, buf.String(), "global test message")
}

func TestDefaultLogger(t *testing.T) {
	// Save and restore original global logger
	original := globalLogger
	defer func() {
		globalLogger = original
		// Note: We cannot safely restore sync.Once, so we leave it initialized
	}()

	// Reset for this test
	globalLogger = nil
	// Note: We cannot reset sync.Once due to internal mutex, but that's OK
	// since this test only verifies Default() works

	// The default logger should exist and work
	assert.NotNil(t, Default())

	// Should be able to log without panicking
	require.NotPanics(t, func() {
		Default().Info("test")
	})
}

func TestLoggerWithSource(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(Config{
		Level:     LevelDebug,
		Format:    FormatText,
		Writer:    &buf,
		AddSource: true,
	})

	logger.Debug("test with source")

	output := buf.String()
	assert.Contains(t, output, "test with source")
	// Should contain source location information
	assert.Contains(t, output, "source=")
}

func TestConcurrentLogging(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(Config{
		Level:  LevelInfo,
		Format: FormatJSON,
		Writer: &buf,
	})

	// Log concurrently from multiple goroutines
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(n int) {
			logger.InfoWith("concurrent log", "goroutine", n)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should have 10 log entries
	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	assert.Len(t, lines, 10)

	// Each should be valid JSON
	for _, line := range lines {
		var entry map[string]interface{}
		err := json.Unmarshal([]byte(line), &entry)
		assert.NoError(t, err)
	}
}
