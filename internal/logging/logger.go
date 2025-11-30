// ABOUTME: Structured logging implementation using Go's standard log/slog package
// ABOUTME: Provides thread-safe, context-aware logging with multiple output formats and levels

package logging

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
)

// Level represents log severity levels
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

// Format represents log output format
type Format int

const (
	FormatText Format = iota
	FormatJSON
)

// Context keys for propagating metadata
type contextKey string

const (
	conversationIDKey contextKey = "conversation_id"
	requestIDKey      contextKey = "request_id"
)

// Config holds logger configuration
type Config struct {
	Level     Level
	Format    Format
	Writer    io.Writer
	LogFile   string
	AddSource bool
}

// Logger wraps slog.Logger with additional functionality
type Logger struct {
	slog   *slog.Logger
	level  Level
	format Format
	file   *os.File
	mu     sync.Mutex
}

var (
	globalLogger *Logger
	once         sync.Once
)

// NewLogger creates a new logger with the given configuration
func NewLogger(cfg Config) *Logger {
	var handler slog.Handler

	writer := cfg.Writer
	if writer == nil {
		writer = os.Stderr
	}

	opts := &slog.HandlerOptions{
		Level:     levelToSlog(cfg.Level),
		AddSource: cfg.AddSource,
	}

	switch cfg.Format {
	case FormatJSON:
		handler = slog.NewJSONHandler(writer, opts)
	default:
		handler = slog.NewTextHandler(writer, opts)
	}

	return &Logger{
		slog:   slog.New(handler),
		level:  cfg.Level,
		format: cfg.Format,
	}
}

// NewLoggerWithFile creates a logger that writes to a file and optionally to writer
func NewLoggerWithFile(cfg Config) (*Logger, error) {
	if cfg.LogFile == "" {
		return nil, fmt.Errorf("log file path is required")
	}

	// Create log file
	file, err := os.OpenFile(cfg.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	// Determine writer(s)
	var writer io.Writer
	if cfg.Writer != nil {
		// Write to both file and provided writer
		writer = io.MultiWriter(file, cfg.Writer)
	} else {
		writer = file
	}

	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level:     levelToSlog(cfg.Level),
		AddSource: cfg.AddSource,
	}

	switch cfg.Format {
	case FormatJSON:
		handler = slog.NewJSONHandler(writer, opts)
	default:
		handler = slog.NewTextHandler(writer, opts)
	}

	return &Logger{
		slog:   slog.New(handler),
		level:  cfg.Level,
		format: cfg.Format,
		file:   file,
	}, nil
}

// Close closes any open file handles
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, args ...any) {
	l.slog.Debug(msg, args...)
}

// Info logs an info message
func (l *Logger) Info(msg string, args ...any) {
	l.slog.Info(msg, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, args ...any) {
	l.slog.Warn(msg, args...)
}

// Error logs an error message
func (l *Logger) Error(msg string, args ...any) {
	l.slog.Error(msg, args...)
}

// DebugWith logs a debug message with structured attributes
func (l *Logger) DebugWith(msg string, args ...any) {
	l.slog.Debug(msg, args...)
}

// InfoWith logs an info message with structured attributes
func (l *Logger) InfoWith(msg string, args ...any) {
	l.slog.Info(msg, args...)
}

// WarnWith logs a warning message with structured attributes
func (l *Logger) WarnWith(msg string, args ...any) {
	l.slog.Warn(msg, args...)
}

// ErrorWith logs an error message with structured attributes
func (l *Logger) ErrorWith(msg string, args ...any) {
	l.slog.Error(msg, args...)
}

// ErrorWithErr logs an error message with an error object
func (l *Logger) ErrorWithErr(msg string, err error, args ...any) {
	allArgs := append([]any{"error", err}, args...)
	l.slog.Error(msg, allArgs...)
}

// InfoWithContext logs an info message with context metadata
func (l *Logger) InfoWithContext(ctx context.Context, msg string, args ...any) {
	l.logWithContext(ctx, slog.LevelInfo, msg, args...)
}

// DebugWithContext logs a debug message with context metadata
func (l *Logger) DebugWithContext(ctx context.Context, msg string, args ...any) {
	l.logWithContext(ctx, slog.LevelDebug, msg, args...)
}

// WarnWithContext logs a warning message with context metadata
func (l *Logger) WarnWithContext(ctx context.Context, msg string, args ...any) {
	l.logWithContext(ctx, slog.LevelWarn, msg, args...)
}

// ErrorWithContext logs an error message with context metadata
func (l *Logger) ErrorWithContext(ctx context.Context, msg string, args ...any) {
	l.logWithContext(ctx, slog.LevelError, msg, args...)
}

// logWithContext extracts context values and logs with them
func (l *Logger) logWithContext(ctx context.Context, level slog.Level, msg string, args ...any) {
	// Extract context values
	var ctxArgs []any
	if convID := ctx.Value(conversationIDKey); convID != nil {
		ctxArgs = append(ctxArgs, "conversation_id", convID)
	}
	if reqID := ctx.Value(requestIDKey); reqID != nil {
		ctxArgs = append(ctxArgs, "request_id", reqID)
	}

	// Combine context args with provided args
	allArgs := append(ctxArgs, args...)

	// Log at appropriate level
	l.slog.Log(ctx, level, msg, allArgs...)
}

// WithConversationID adds a conversation ID to the context
func WithConversationID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, conversationIDKey, id)
}

// WithRequestID adds a request ID to the context
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

// LevelFromString converts a string to a log level
func LevelFromString(s string) Level {
	switch strings.ToLower(s) {
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	case "warn", "warning":
		return LevelWarn
	case "error":
		return LevelError
	default:
		return LevelInfo // Default to Info for unknown values
	}
}

// levelToSlog converts our Level to slog.Level
func levelToSlog(level Level) slog.Level {
	switch level {
	case LevelDebug:
		return slog.LevelDebug
	case LevelInfo:
		return slog.LevelInfo
	case LevelWarn:
		return slog.LevelWarn
	case LevelError:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// Global logger functions

// Default returns the global logger instance
func Default() *Logger {
	once.Do(func() {
		globalLogger = NewLogger(Config{
			Level:  LevelInfo,
			Format: FormatText,
			Writer: os.Stderr,
		})
	})
	return globalLogger
}

// SetGlobalLogger sets the global logger instance
func SetGlobalLogger(logger *Logger) {
	globalLogger = logger
}

// Debug logs a debug message using the global logger
func Debug(msg string, args ...any) {
	Default().Debug(msg, args...)
}

// Info logs an info message using the global logger
func Info(msg string, args ...any) {
	Default().Info(msg, args...)
}

// Warn logs a warning message using the global logger
func Warn(msg string, args ...any) {
	Default().Warn(msg, args...)
}

// Error logs an error message using the global logger
func Error(msg string, args ...any) {
	Default().Error(msg, args...)
}

// DebugWith logs a debug message with attributes using the global logger
func DebugWith(msg string, args ...any) {
	Default().DebugWith(msg, args...)
}

// InfoWith logs an info message with attributes using the global logger
func InfoWith(msg string, args ...any) {
	Default().InfoWith(msg, args...)
}

// WarnWith logs a warning message with attributes using the global logger
func WarnWith(msg string, args ...any) {
	Default().WarnWith(msg, args...)
}

// ErrorWith logs an error message with attributes using the global logger
func ErrorWith(msg string, args ...any) {
	Default().ErrorWith(msg, args...)
}

// ErrorWithErr logs an error with an error object using the global logger
func ErrorWithErr(msg string, err error, args ...any) {
	Default().ErrorWithErr(msg, err, args...)
}
