package providers

import (
	"fmt"

	"github.com/biisal/bai/internal/config"
)

type Provider interface {
	StreamChat(modelId string)
}

func NewFromConfig(cfg config.ProviderConfig) (Provider, error) {
	switch cfg.Format {
	case config.FormatOpenAI:
		{
			return NewProviderOpenAI(cfg.BaseURL, cfg.APIKey), nil
		}
	}

	return nil, fmt.Errorf("unknown provider format: %s, hint use one of: %s",
		cfg.Format, []string{config.FormatOpenAI})
}
