package providers

import (
	"context"
	"fmt"

	"github.com/biisal/bai/internal/agent/core/instruction"
	"github.com/biisal/bai/internal/agent/core/tools"
	"github.com/biisal/bai/internal/agent/providers/variant"
	"github.com/biisal/bai/internal/config"
	"github.com/biisal/bai/internal/domain"
	broker "github.com/biisal/bai/internal/pubsub"
)

type StreamResult struct {
	Text         string
	ThinkingText string
	ToolCalls    []tools.Call
}

type Provider interface {
	StreamChat(ctx context.Context, modelId string, history []domain.Message) (StreamResult, error)
	ID() string
}

func resolveVariant(cfg config.ProviderConfig) (*variant.Spec, error) {
	if cfg.Variant == "" {
		return nil, nil
	}
	factory, ok := variant.Get(cfg.Variant)
	if !ok {
		return nil, fmt.Errorf("unknown provider variant: %s, hint use one of: %s",
			cfg.Variant, variant.Names())
	}
	return factory(cfg)
}

func NewFromConfig(cfg config.ProviderConfig, broker broker.Service) (Provider, error) {
	systemPrompt := instruction.BuildSystemPrompt()
	switch cfg.Format {
	case config.FormatOpenAI:
		{
			spec, err := resolveVariant(cfg)
			if err != nil {
				return nil, err
			}
			return NewProviderOpenAI(cfg.BaseURL, cfg.APIKey, cfg.Name, broker,
				systemPrompt, applyVariant(spec, cfg.APIKey)...), nil
		}
	}
	return nil, fmt.Errorf("unknown provider format: %s, hint use one of: %s",
		cfg.Format, []string{config.FormatOpenAI})
}
