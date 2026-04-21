package logger

import (
	"io"
	"log"
	"log/slog"
	"os"
)

const logFile = "./app.log"

func New(env string) *slog.Logger {
	var level slog.Level
	if env == "production" {
		level = slog.LevelInfo
	} else {
		level = slog.LevelDebug
	}

	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Failed to open log file", logFile, ":", err)
	}

	return slog.New(slog.NewJSONHandler(io.MultiWriter(file, os.Stdout), &slog.HandlerOptions{
		Level: level,
	}))
}
