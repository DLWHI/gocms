package logger

import "log/slog"

func Error(source string, message string, reason error, attrs ...any) {
	args := append([]any{"source", source, "reason", reason}, attrs...)
	slog.Default().Error(message, args...)
}

func Warn(source string, message string, reason error, attrs ...any) {
	args := []any{"source", source}
	if reason != nil {
		args = append(args, "reason", reason)
	}
	args = append(args, attrs...)
	slog.Default().Warn(message, args...)
}

func Info(source string, message string, attrs ...any) {
	args := append([]any{"source", source}, attrs...)
	slog.Default().Info(message, args...)
}
