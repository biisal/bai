package providers

import (
	"context"
	"fmt"

	"github.com/biisal/bai/internal/agent/core/instruction"
	"github.com/biisal/bai/internal/agent/core/tools"
	"github.com/biisal/bai/internal/config"
	"github.com/biisal/bai/internal/domain"
	broker "github.com/biisal/bai/internal/pubsub"
)

type StreamResult struct {
	Text      string
	ToolCalls []tools.Call
}

type Provider interface {
	StreamChat(ctx context.Context, modelId string, history []domain.Message) (StreamResult, error)
	ID() string
}

func NewFromConfig(cfg config.ProviderConfig, broker broker.Service) (Provider, error) {
	systemPrompt := instruction.BuildSystemPrompt()
	switch cfg.Format {
	case config.FormatOpenAI:
		{
			return NewProviderOpenAI(cfg.BaseURL, cfg.APIKey, cfg.Name, broker, systemPrompt), nil
		}
	}

	return nil, fmt.Errorf("unknown provider format: %s, hint use one of: %s",
		cfg.Format, []string{config.FormatOpenAI})
}
