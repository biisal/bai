package providers

import (
	"context"
	"log/slog"

	broker "github.com/biisal/bai/internal/pubsub"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/responses"
)

type ProviderOpenAI struct {
	baseUrl    string
	apiKey     string
	broker     broker.Service
	client     *openai.Client
	providerID string
}

func NewProviderOpenAI(baseURL, apiKey, providerID string, broker broker.Service) Provider {
	client := openai.NewClient(option.WithAPIKey(apiKey), option.WithBaseURL(baseURL))

	return &ProviderOpenAI{baseURL, apiKey, broker, &client, providerID}
}

func (p *ProviderOpenAI) ID() string {
	return p.providerID
}

func (p *ProviderOpenAI) StreamChat(ctx context.Context, modelId string, content string) error {
	stream := p.client.Responses.NewStreaming(ctx, responses.ResponseNewParams{
		Input: responses.ResponseNewParamsInputUnion{OfString: openai.String(content)},
		Model: modelId,
	})

	for stream.Next() {
		data := stream.Current()
		p.broker.Publish(ctx, broker.Message{Type: broker.EventAgentMessageChunk, Text: data.Text})
		slog.Info("openai response", "text", data.JSON.Text.Raw())
		if data.JSON.Text.Valid() {
			// p.broker.Publish(ctx, broker.Message{Type: broker.EventAgentStopThinking, Text: data.Text})
			break
		}
	}

	if err := stream.Err(); err != nil {
		p.broker.Publish(ctx, broker.Message{Type: broker.EventAgentError, Text: err.Error()})
		return err
	}

	return nil
}
