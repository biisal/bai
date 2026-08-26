package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

var ErrNoProviders = errors.New("invalid config file: no providers found")

type ProviderFormat string

const (
	FormatOpenAI    ProviderFormat = "openai-compatible"
	FormatAnthropic ProviderFormat = "anthropic"
)

type ProviderConfig struct {
	Name    string         `json:"name"`
	APIKey  string         `json:"api_key"`
	Format  ProviderFormat `json:"format"`
	BaseURL string         `json:"base_url"`
	Variant string         `json:"variant"`
	Models  []ModelConfig  `json:"models"`
}

type ModelConfig struct {
	ID        string `json:"id"`
	Context   int    `json:"context"`
	MaxOutput int    `json:"max_output"`
}

type Config struct {
	DatabasePath string           `json:"database_path"`
	LogFilePath  string           `json:"log_file_path"`
	Providers    []ProviderConfig `json:"providers"`
}

func AppConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "./config.json"
	}
	return filepath.Join(home, ".config", "bai")
}

func DefaultConfigPath() string {
	return filepath.Join(AppConfigDir(), "config.json")
}

func DefaultDatabasePath() string {
	return filepath.Join(AppConfigDir(), "bai.db")
}

func DefaultLogFilePath() string {
	return filepath.Join(AppConfigDir(), "bai.log")
}

func DefaultConfig() *Config {
	return &Config{
		Providers: []ProviderConfig{
			{
				Format:  FormatOpenAI,
				Name:    "OpenAI",
				APIKey:  "sk-...",
				BaseURL: "https://api.openai.com/v1",
				Models: []ModelConfig{
					{
						ID:        "gpt-5.5",
						Context:   4096,
						MaxOutput: 4000,
					},
				},
			},
		},
	}
}

func Load(path string) (*Config, error) {
	finalPath := DefaultConfigPath()
	if path != "" {
		finalPath = path
	}

	if _, err := os.Stat(finalPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file does not exists: %s", finalPath)
	}

	var config Config

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
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		prefix := "invalid config file"
		if errors.Is(err, io.EOF) {
			return nil, fmt.Errorf("%s: empty file", prefix)
		}

		data, err := json.MarshalIndent(DefaultConfig(), "", "  ")
		if err != nil {
			return nil, err
		}

		return nil, fmt.Errorf("invalid config file: make sure your config looks like:\n%v", string(data))
	}

	if len(config.Providers) == 0 {
		return nil, ErrNoProviders
	}

	var errors []string
	metProvides := make(map[string]bool)
	for _, provider := range config.Providers {
		if provider.BaseURL == "" {
			errors = append(errors, fmt.Sprintf("base_url can't be empty for provider: %s", provider.Name))
		}
		if provider.Format == "" {
			errors = append(errors, fmt.Sprintf("format can't be empty for provider: %s", provider.Name))
		}

		if _, ok := metProvides[provider.Name]; ok {
			errors = append(errors, fmt.Sprintf("provider id must be unique for provider: %s", provider.Name))
		}
		metProvides[provider.Name] = true
	}

	if len(errors) > 0 {
		return nil, fmt.Errorf("invalid config:\n%s", strings.Join(errors, "\n"))
	}
	if config.DatabasePath == "" {
		config.DatabasePath = DefaultDatabasePath()
	}
	if config.LogFilePath == "" {
		config.LogFilePath = DefaultLogFilePath()
	}
	return &config, nil
}
