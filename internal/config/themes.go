package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

type Theme struct {
	Background string `json:"background"`
	Foreground string `json:"foreground"`

	Muted           string `json:"muted"`
	MutedForeground string `json:"mutedForeground"`

	Primary           string `json:"primary"`
	PrimaryForeground string `json:"primaryForeground"`

	Secondary           string `json:"secondary"`
	SecondaryForeground string `json:"secondaryForeground"`

	Accent           string `json:"accent"`
	AccentForeground string `json:"accentForeground"`

	Border string `json:"border"`

	Destructive           string `json:"destructive"`
	DestructiveForeground string `json:"destructiveForeground"`
	Success               string `json:"success"`
	SuccessForeground     string `json:"successForeground"`
	Warning               string `json:"warning"`
	WarningForeground     string `json:"warningForeground"`
}

func DefaultTheme() *Theme {
	return &Theme{
		Background: "#ffffff",
		Foreground: "7",

		Muted:           "236",
		MutedForeground: "244",

		Primary:           "6",
		PrimaryForeground: "0",

		Secondary:           "240",
		SecondaryForeground: "7",

		Accent:           "14",
		AccentForeground: "0",

		Border: "6",

		Destructive:           "1",
		DestructiveForeground: "7",
		Success:               "10",
		SuccessForeground:     "0",
		Warning:               "11",
		WarningForeground:     "0",
	}
}

func ThemeConfigDir() string {
	return filepath.Join(AppConfigDir(), "themes")
}

func ThemeConfigPath() string {
	return filepath.Join(ThemeConfigDir(), "default.json")
}

func NewTheme(path string) (*Theme, error) {
	finalPath := ThemeConfigPath()
	if path != "" {
		finalPath = path
	}

	if _, err := os.Stat(finalPath); os.IsNotExist(err) {
		return DefaultTheme(), nil
	}

	var theme Theme

	file, err := os.Open(finalPath)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := file.Close(); err != nil {
			slog.Error(err.Error())
			return
		}
	}()
	if err := json.NewDecoder(file).Decode(&theme); err != nil {
		prefix := "invalid config file"
		if errors.Is(err, io.EOF) {
			slog.Warn("found empty config file, using default theme")
			return DefaultTheme(), nil
		}
		return nil, fmt.Errorf("%s: %w", prefix, err)
	}
	return &theme, nil
}