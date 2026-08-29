package logger

import (
	"log/slog"
	"os"
	"strings"
)

// New returns a JSON slog logger writing to stdout (PaaS-friendly).
// LOG_LEVEL: debug|info|warn|error (default info).
func New() *slog.Logger {
	level := slog.LevelInfo
	switch strings.ToLower(strings.TrimSpace(os.Getenv("LOG_LEVEL"))) {
	case "debug":
		level = slog.LevelDebug
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	case "info", "":
		level = slog.LevelInfo
	}
	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	return slog.New(h)
}

// SetDefault installs logger as the process default.
func SetDefault(log *slog.Logger) {
	slog.SetDefault(log)
}
