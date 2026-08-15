package providers

import (
	"context"
	"fmt"

	"github.com/biisal/bai/internal/config"
	repo "github.com/biisal/bai/internal/db/sqlc"
	broker "github.com/biisal/bai/internal/pubsub"
)

type Provider interface {
	StreamChat(ctx context.Context, modelId string, history []repo.Message) (finalMessage string, err error)
	ID() string
}

func NewFromConfig(cfg config.ProviderConfig, broker broker.Service) (Provider, error) {
	switch cfg.Format {
	case config.FormatOpenAI:
		{
			return NewProviderOpenAI(cfg.BaseURL, cfg.APIKey, cfg.Name, broker), nil
		}
	}

	return nil, fmt.Errorf("unknown provider format: %s, hint use one of: %s",
		cfg.Format, []string{config.FormatOpenAI})
}
