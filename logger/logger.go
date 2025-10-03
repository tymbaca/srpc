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

type NoopLogger struct{}

func (no NoopLogger) Debug(msg string, args ...any) {}
func (no NoopLogger) Info(msg string, args ...any)  {}
func (no NoopLogger) Warn(msg string, args ...any)  {}
func (no NoopLogger) Error(msg string, args ...any) {}

type DefaulSLogger struct{}

func (DefaulSLogger) Debug(msg string, args ...any) {
	slog.Debug(msg, args...)
}

func (DefaulSLogger) Info(msg string, args ...any) {
	slog.Info(msg, args...)
}

func (DefaulSLogger) Warn(msg string, args ...any) {
	slog.Warn(msg, args...)
}

func (DefaulSLogger) Error(msg string, args ...any) {
	slog.Error(msg, args...)
}
