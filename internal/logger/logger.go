package logger

import "log/slog"

func Error(source, message string, reason error, attrs ...any) {
	args := append([]any{"reason", reason}, attrs...)
	slog.Default().With("source", source).Error(message, args...)
}

func Warn(source, message string, reason error, attrs ...any) {
	if reason == nil {
		slog.Default().With("source", source).Warn(message, attrs...)
	} else {
		args := append([]any{"reason", reason}, attrs...)
		slog.Default().With("source", source).Warn(message, args...)
	}
}

func Info(source, message string, attrs ...any) {
	slog.Default().With("source", source).Info(message, attrs...)
}
