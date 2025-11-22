// Package logger provides logger interface for srpc and few basic implementations.
package logger

import "log/slog"

// Logger accepts even amount of args:
// logger.Info("something", "key1", "val1", "key2", "val2", "key3", "val3")
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

// NoopLogger does nothing.
type NoopLogger struct{}

func (no NoopLogger) Debug(msg string, args ...any) {}
func (no NoopLogger) Info(msg string, args ...any)  {}
func (no NoopLogger) Warn(msg string, args ...any)  {}
func (no NoopLogger) Error(msg string, args ...any) {}

// DefaultSLogger uses standard [slog] backage for logging.
type DefaultSLogger struct{}

func (DefaultSLogger) Debug(msg string, args ...any) {
	slog.Debug(msg, args...)
}

func (DefaultSLogger) Info(msg string, args ...any) {
	slog.Info(msg, args...)
}

func (DefaultSLogger) Warn(msg string, args ...any) {
	slog.Warn(msg, args...)
}

func (DefaultSLogger) Error(msg string, args ...any) {
	slog.Error(msg, args...)
}
