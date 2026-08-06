package providers

import (
	"context"

	repo "github.com/biisal/bai/internal/db/sqlc"
	broker "github.com/biisal/bai/internal/pubsub"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
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

func buildHistory(history []repo.Message) []openai.ChatCompletionMessageParamUnion {
	var messages []openai.ChatCompletionMessageParamUnion
	for _, m := range history {
		if m.Role == "user" {
			messages = append(messages, openai.UserMessage(m.Content))
		} else {
			messages = append(messages, openai.AssistantMessage(m.Content))
		}
	}
	return messages
}

func (p *ProviderOpenAI) StreamChat(ctx context.Context, modelId string, history []repo.Message) error {
	stream := p.client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
		Messages: buildHistory(history),
		Model:    modelId,
	})

	for stream.Next() {
		current := stream.Current()
		p.broker.Publish(ctx, broker.Message{
			Type: broker.EventAgentMessageChunk,
			// TODO: perse the reasoning content too
			Text: current.Choices[0].Delta.Content,
		})
	}

	if err := stream.Err(); err != nil {
		p.broker.Publish(ctx, broker.Message{Type: broker.EventAgentError, Text: err.Error()})
		return err
	}

	return nil
}
