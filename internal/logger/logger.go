package logger

import (
	"log/slog"
	"os"
)

func SetUpLogger(filePath string, level slog.Level) (*os.File, error) {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}
	lg := slog.New(slog.NewJSONHandler(file, &slog.HandlerOptions{Level: level, AddSource: true}))
	slog.SetDefault(lg)
	return file, nil
}
