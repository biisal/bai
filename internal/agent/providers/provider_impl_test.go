package providers

import (
	"fmt"
	"testing"

	"github.com/biisal/bai/internal/config"
	test_utils "github.com/biisal/bai/utils/tests"
)

func TestNewFromConfig(t *testing.T) {
	tests := []struct {
		name    string
		wantErr error
		input   config.ProviderConfig
	}{
		{
			input: config.ProviderConfig{
				Format:  "invalid-provider",
				APIKey:  "sk-...",
				BaseURL: "https://api.openai.com/v1",
			},
			name:    "should throw error when invalid format provided",
			wantErr: fmt.Errorf("unknown provider format: invalid-provider, hint use one of: %s", []string{config.FormatOpenAI}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewFromConfig(tt.input)
			test_utils.AssertError(t, err, tt.wantErr)
			if tt.wantErr == nil && provider == nil {
				t.Errorf("got nil, want provider")
			}
		})
	}
}
