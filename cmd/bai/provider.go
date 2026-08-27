package main

import (
	"fmt"
	"net/http"

	fantasy "charm.land/fantasy"
	"charm.land/fantasy/providers/anthropic"
	"charm.land/fantasy/providers/openaicompat"
	"github.com/openai/openai-go/v3/option"

	"github.com/biisal/bai/internal/agent/providers/variant"
	"github.com/biisal/bai/internal/config"
)

func buildProvider(cfg config.ProviderConfig) (fantasy.Provider, error) {
	switch cfg.Format {
	case config.FormatOpenAI:
		return buildOpenAIProvider(cfg)
	case config.FormatAnthropic:
		return buildAnthropicProvider(cfg)
	default:
		return nil, fmt.Errorf("unknown provider format: %s, hint use one of: %v", cfg.Format, []string{string(config.FormatOpenAI), string(config.FormatAnthropic)})
	}
}

func buildOpenAIProvider(cfg config.ProviderConfig) (fantasy.Provider, error) {
	opts := []openaicompat.Option{
		openaicompat.WithAPIKey(cfg.APIKey),
		openaicompat.WithBaseURL(cfg.BaseURL),
		openaicompat.WithName(cfg.Name),
	}

	if cfg.Variant != "" {
		factory, ok := variant.Get(cfg.Variant)
		if !ok {
			return nil, fmt.Errorf("unknown variant %q, available: %v", cfg.Variant, variant.Names())
		}
		spec, err := factory(cfg)
		if err != nil {
			return nil, fmt.Errorf("variant %q: %w", cfg.Variant, err)
		}
		opts = append(opts, openaicompat.WithSDKOptions(
			option.WithMiddleware(variantMiddleware(spec, cfg.APIKey)),
		))
	}

	return openaicompat.New(opts...)
}

func buildAnthropicProvider(cfg config.ProviderConfig) (fantasy.Provider, error) {
	opts := []anthropic.Option{
		anthropic.WithAPIKey(cfg.APIKey),
		anthropic.WithBaseURL(cfg.BaseURL),
		anthropic.WithName(cfg.Name),
	}

	if cfg.Variant != "" {
		factory, ok := variant.Get(cfg.Variant)
		if !ok {
			return nil, fmt.Errorf("unknown variant %q, available: %v", cfg.Variant, variant.Names())
		}
		spec, err := factory(cfg)
		if err != nil {
			return nil, fmt.Errorf("variant %q: %w", cfg.Variant, err)
		}

		headers := make(map[string]string, len(spec.Headers))
		for _, h := range spec.Headers {
			headers[h.Key] = h.Value()
		}
		if cfg.APIKey == "" && spec.AuthFallback != "" {
			headers["Authorization"] = spec.AuthScheme + " " + spec.AuthFallback
		}
		if len(headers) > 0 {
			opts = append(opts, anthropic.WithHeaders(headers))
		}
	}

	return anthropic.New(opts...)
}

func variantMiddleware(spec *variant.Spec, apiKey string) option.Middleware {
	return func(req *http.Request, next option.MiddlewareNext) (*http.Response, error) {
		for _, h := range spec.Headers {
			req.Header.Set(h.Key, h.Value())
		}
		if apiKey == "" && spec.AuthFallback != "" {
			req.Header.Set("Authorization", spec.AuthScheme+" "+spec.AuthFallback)
		}
		return next(req)
	}
}
