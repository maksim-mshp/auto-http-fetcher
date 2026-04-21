package logger

import (
	"log/slog"
	"os"
)

func New(env string) *slog.Logger {
	var level slog.Level
	if env == "production" {
		level = slog.LevelInfo
	} else {
		level = slog.LevelDebug
	}

	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))
}
