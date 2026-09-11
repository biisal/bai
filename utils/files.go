package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func ResolvePath(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", fmt.Errorf("path is empty")
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	if strings.HasPrefix(absPath, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		absPath = strings.Replace(absPath, "~", homeDir, 1)
	}
	return absPath, nil
}
