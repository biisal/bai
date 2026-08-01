package main

import (
	"fmt"
	"testing"

	test_utils "github.com/biisal/bai/utils/tests"
)

func TestStart(t *testing.T) {
	t.Run("Should throw error when invalid config path provided",
		func(t *testing.T) {
			err := start("invalid path")
			test_utils.AssertError(t, err, fmt.Errorf("failed to load config: config file does not exists: invalid path"))
		})

	t.Run("Should throw error when provider format is unknown",
		func(t *testing.T) {
			path := test_utils.WriteTempConfig(t, `{
				"providers": [
					{"id": "openai", "name": "OpenAI", "format": "unknown-format", "base_url": "https://api.openai.com/v1"}
				]
			}`)
			err := start(path)
			test_utils.AssertError(t, err, fmt.Errorf("failed to create provider: unknown provider format: unknown-format, hint use one of: [openai-compatible]"))
		})
}
