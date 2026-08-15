package config

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	test_utils "github.com/biisal/bai/utils/tests"
)

func createTempFile(t *testing.T, path string, fileContent ...any) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	if len(fileContent) > 0 {
		if err := json.NewEncoder(file).Encode(fileContent[0]); err != nil {
			t.Fatal(err)
		}
	}

	t.Cleanup(func() {
		defer func() {
			if err := os.Remove(path); err != nil {
				t.Fatal(err)
			}
		}()
	})
}

func defaultConfigString() string {
	data, _ := json.MarshalIndent(DefaultConfig(), "", "  ")
	return string(data)
}

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr error
		setupFN func(t *testing.T)
		checkFN func(t *testing.T, config *Config)
	}{
		{
			name:    "invalid path",
			path:    "/tmp/tempfile.json",
			wantErr: fmt.Errorf("config file does not exists: /tmp/tempfile.json"),
		},
		{
			name:    "valid path with no providers",
			path:    "/tmp/tempfile.json",
			wantErr: ErrNoProviders,
			setupFN: func(t *testing.T) {
				path := "/tmp/tempfile.json"
				createTempFile(t, path, Config{
					Providers: []ProviderConfig{},
				})
			},
		},
		{
			name:    "valid path with empty file",
			path:    "/tmp/tempfile.json",
			wantErr: fmt.Errorf("invalid config file: empty file"),
			setupFN: func(t *testing.T) {
				path := "/tmp/tempfile.json"
				createTempFile(t, path)
			},
		},
		{
			name:    "valid path with invalid config",
			path:    "/tmp/tempfile.json",
			wantErr: fmt.Errorf("invalid config file: make sure your config looks like:\n%s", defaultConfigString()),
			setupFN: func(t *testing.T) {
				path := "/tmp/tempfile.json"
				createTempFile(t, path, "invalid content here")
			},
		},
		{
			name:    "should return error if  base_url is empty",
			path:    "/tmp/tempfile.json",
			wantErr: fmt.Errorf("invalid config:\nbase_url can't be empty for provider: test"),
			setupFN: func(t *testing.T) {
				path := "/tmp/tempfile.json"
				createTempFile(t, path, Config{
					Providers: []ProviderConfig{
						{
							BaseURL: "",
							Format:  FormatOpenAI,
							Name:    "test",
						},
					},
				})
			},
		},
		{
			name:    "valid path without format",
			path:    "/tmp/tempfile.json",
			wantErr: fmt.Errorf("invalid config:\nformat can't be empty for provider: test"),
			setupFN: func(t *testing.T) {
				path := "/tmp/tempfile.json"
				createTempFile(t, path, Config{
					Providers: []ProviderConfig{
						{
							BaseURL: "https://api.openai.com/v1",
							Format:  "",
							Name:    "test",
						},
					},
				})
			},
		},
		{
			name:    "should throw error of list of missing base_url and format",
			path:    "/tmp/tempfile.json",
			wantErr: fmt.Errorf("invalid config:\nbase_url can't be empty for provider: test\nformat can't be empty for provider: test"),
			setupFN: func(t *testing.T) {
				path := "/tmp/tempfile.json"
				createTempFile(t, path, Config{
					Providers: []ProviderConfig{
						{
							Name: "test",
						},
					},
				})
			},
		},
		{
			name:    "should throw error of list of missing base_url and format for different config file",
			path:    "/tmp/tempfile.json",
			wantErr: fmt.Errorf("invalid config:\nformat can't be empty for provider: test-2\nbase_url can't be empty for provider: test-3"),
			setupFN: func(t *testing.T) {
				path := "/tmp/tempfile.json"
				createTempFile(t, path, Config{
					Providers: []ProviderConfig{
						{
							Name:    "test",
							BaseURL: "https://api.openai.com/v1",
							Format:  FormatOpenAI,
						},
						{
							Name:    "test-2",
							BaseURL: "https://api.openai.com/v1",
						},
						{
							Name:   "test-3",
							Format: FormatOpenAI,
						},
					},
				})
			},
		},
		{
			name:    "should throw error when name is duplicated",
			path:    "/tmp/tempfile.json",
			wantErr: fmt.Errorf("invalid config:\nprovider id must be unique for provider: test"),
			setupFN: func(t *testing.T) {
				path := "/tmp/tempfile.json"
				createTempFile(t, path, Config{
					Providers: []ProviderConfig{
						{
							Name:    "test",
							BaseURL: "https://api.openai.com/v1",
							Format:  FormatOpenAI,
						},
						{
							Name:    "test",
							BaseURL: "https://api.openai.com/v1",
							Format:  FormatOpenAI,
						},
					},
				})
			},
		},
		{
			name:    "set defalut database_path if not set in config",
			path:    "/tmp/tempfile.json",
			wantErr: nil,
			checkFN: func(t *testing.T, config *Config) {
				dbPath := DefaultDatabasePath()
				if config.DatabasePath != dbPath {
					t.Errorf("config.DatabasePath = %v, want %v", config.DatabasePath, dbPath)
				}
			},
			setupFN: func(t *testing.T) {
				path := "/tmp/tempfile.json"
				createTempFile(t, path, Config{
					Providers: []ProviderConfig{
						{
							Name:    "test",
							BaseURL: "https://api.openai.com/v1",
							Format:  FormatOpenAI,
						},
					},
				})
			},
		},
		{
			name:    "set config database_path if set",
			path:    "/tmp/tempfile.json",
			wantErr: nil,
			checkFN: func(t *testing.T, config *Config) {
				dbPath := "/tmp/test.db"
				if config.DatabasePath != dbPath {
					t.Errorf("config.DatabasePath = %v, want %v", config.DatabasePath, dbPath)
				}
			},
			setupFN: func(t *testing.T) {
				path := "/tmp/tempfile.json"
				createTempFile(t, path, Config{
					DatabasePath: "/tmp/test.db",
					Providers: []ProviderConfig{
						{
							Name:    "test",
							BaseURL: "https://api.openai.com/v1",
							Format:  FormatOpenAI,
						},
					},
				})
			},
		},
		{
			name:    "set default log_file_path if not set in config",
			path:    "/tmp/tempfile.json",
			wantErr: nil,
			checkFN: func(t *testing.T, config *Config) {
				logPath := DefaultLogFilePath()
				if config.LogFilePath != logPath {
					t.Errorf("config.LogFilePath = %v, want %v", config.LogFilePath, logPath)
				}
			},
			setupFN: func(t *testing.T) {
				path := "/tmp/tempfile.json"
				createTempFile(t, path, Config{
					DatabasePath: "/tmp/test.db",
					Providers: []ProviderConfig{
						{
							Name:    "test",
							BaseURL: "https://api.openai.com/v1",
							Format:  FormatOpenAI,
						},
					},
				})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupFN != nil {
				tt.setupFN(t)
			}

			config, err := Load(tt.path)
			if tt.checkFN != nil {
				tt.checkFN(t, config)
			}
			test_utils.AssertError(t, err, tt.wantErr)
		})
	}
}
