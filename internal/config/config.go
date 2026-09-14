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
	SoundPath    string           `json:"sound_path"`
	SkillsPaths  []string         `json:"skills_paths"`
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

func DefaultSkillsPaths() []string {
	paths := []string{
		filepath.Join(AppConfigDir(), "skills"),
	}

	home, err := os.UserHomeDir()
	if err == nil && home != "" {
		paths = append(
			paths,
			filepath.Join(home, ".agents", "skills"),
			filepath.Join(home, ".claude", "skills"),
		)
	}

	paths = append(
		paths,
		".bai/skills",
		"./skills",
		".agents/skills",
		".claude/skills",
	)

	return paths
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
		if closeErr := file.Close(); closeErr != nil {
			slog.Error("close config file", "error", closeErr)
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

	var errs []string
	metProviders := make(map[string]bool)
	for _, provider := range config.Providers {
		if provider.BaseURL == "" {
			errs = append(errs, fmt.Sprintf("base_url can't be empty for provider: %s", provider.Name))
		}
		if provider.Format == "" {
			errs = append(errs, fmt.Sprintf("format can't be empty for provider: %s", provider.Name))
		}

		if _, ok := metProviders[provider.Name]; ok {
			errs = append(errs, fmt.Sprintf("provider id must be unique for provider: %s", provider.Name))
		}
		metProviders[provider.Name] = true
	}

	if len(errs) > 0 {
		return nil, fmt.Errorf("invalid config:\n%s", strings.Join(errs, "\n"))
	}
	if config.DatabasePath == "" {
		config.DatabasePath = DefaultDatabasePath()
	}
	if config.LogFilePath == "" {
		config.LogFilePath = DefaultLogFilePath()
	}
	return &config, nil
}
