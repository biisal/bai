package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/biisal/bai/internal/agent/tools"
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
			slog.Debug("assistant message", "content", m.Content)
			messages = append(messages, openai.AssistantMessage(m.Content))
		}
	}
	return messages
}

func (p *ProviderOpenAI) StreamChat(ctx context.Context, modelId string, history []repo.Message) (finalMessage string, err error) {
	stream := p.client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
		Messages: buildHistory(history),
		Model:    modelId,
		Tools: []openai.ChatCompletionToolUnionParam{
			openai.ChatCompletionFunctionTool(
				openai.FunctionDefinitionParam{
					Name:        "read_file",
					Description: openai.String("Read a file from the filesystem using path, offset, and limit"),
					Parameters: openai.FunctionParameters{
						"type": "object",
						"properties": map[string]any{
							"path": map[string]string{
								"type": "string",
							},
							"offset": map[string]string{
								"type": "integer",
							},
							"limit": map[string]string{
								"type": "integer",
							},
						},
						"required": []string{"path"},
					},
				},
			),
		},
	})
	var acc openai.ChatCompletionAccumulator
	for stream.Next() {
		current := stream.Current()
		acc.AddChunk(current)

		if reasoning := reasoningDelta(current.RawJSON()); reasoning != "" {
			p.broker.Publish(ctx, broker.Message{Type: broker.EventAgentThinking, Text: reasoning})
		}

		if content, ok := acc.JustFinishedContent(); ok {
			slog.Debug("content finished", "content", content)
		}

		if tool, ok := acc.JustFinishedToolCall(); ok {
			var args map[string]any
			if err := json.Unmarshal([]byte(tool.Arguments), &args); err != nil {
				return "", err
			}
			path, _ := args["path"].(string)
			var offset, limit int64
			if v, ok := args["offset"].(float64); ok {
				offset = int64(v)
			}
			if v, ok := args["limit"].(float64); ok {
				limit = int64(v)
			}

			var result string
			content, err := tools.ReadFile(path, offset, limit)
			if err != nil {
				result = fmt.Sprintf("error reading file: %v", err)
			} else {
				result = content
			}
			p.broker.Publish(ctx, broker.Message{
				Type: broker.EventAgentResponse,
				Text: result,
			})

			history = append(history, repo.Message{
				Role: openai.ToolMessage(),
			})
		}

		if len(current.Choices) > 0 && current.Choices[0].Delta.Content != "" {
			p.broker.Publish(ctx, broker.Message{
				Type: broker.EventAgentResponse,
				Text: current.Choices[0].Delta.Content,
			})
		}

	}
	if err := stream.Err(); err != nil {
		slog.Error("openai stream error", "error", err)
		return "", err
	}

	if len(acc.Choices) > 0 {
		finalMessage := acc.Choices[0].Message.Content
		slog.Debug("Full content openai", "content", finalMessage)
		return finalMessage, nil
	}

	return "", nil
}

func reasoningDelta(raw string) string {
	var chunk struct {
		Choices []struct {
			Delta struct {
				Reasoning        string `json:"reasoning"`
				ReasoningContent string `json:"reasoning_content"`
			} `json:"delta"`
		} `json:"choices"`
	}
	if err := json.Unmarshal([]byte(raw), &chunk); err != nil {
		return ""
	}
	if len(chunk.Choices) == 0 {
		return ""
	}
	d := chunk.Choices[0].Delta
	if d.ReasoningContent != "" {
		return d.ReasoningContent
	}
	return d.Reasoning
}
