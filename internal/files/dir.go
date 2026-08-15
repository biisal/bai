package files

import (
	"log/slog"
	"os"
)

func CurrentDir() string {
	currentWd, err := os.Getwd()
	if err != nil {
		slog.Error("failed to get current dir", "error", err)
	}
	return currentWd
}
