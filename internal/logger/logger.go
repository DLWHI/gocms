package logger

import "log/slog"

func Error(source, message string, reason error) {
	slog.Default().With("source", source).Error(message, "reason", reason)
}

func Warn(source, message string, reason error) {
	slog.Default().With("source", source).Warn(message, "reason", reason)
}

func Info(source, message string, attrs ...any) {
	slog.Default().With("source", source).Info(message, attrs...)
}
