package agent

import (
	"fmt"
	"testing"

	"github.com/biisal/bai/internal/config"
	test_utils "github.com/biisal/bai/utils/tests"
)

func TestBuildProvider(t *testing.T) {
	t.Run("Returns error for unknown provider format", func(t *testing.T) {
		cfg := config.ProviderConfig{
			Name:   "test",
			Format: "unknown-format",
		}
		_, err := buildProvider(cfg)
		test_utils.AssertError(t, err, fmt.Errorf("unknown provider format: unknown-format, hint use one of: [openai-compatible anthropic]"))
	})
}
