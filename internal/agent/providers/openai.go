package providers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/biisal/bai/internal/agent/core/tools"
	"github.com/biisal/bai/internal/agent/providers/variant"
	"github.com/biisal/bai/internal/domain"
	broker "github.com/biisal/bai/internal/pubsub"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

type ProviderOpenAI struct {
	baseUrl       string
	apiKey        string
	broker        broker.Service
	client        *openai.Client
	providerID    string
	systemMessage openai.ChatCompletionMessageParamUnion
}

func NewProviderOpenAI(baseURL, apiKey, providerID string, broker broker.Service, systemPrompt string, extraOpts ...option.RequestOption) Provider {
	opts := append([]option.RequestOption{
		option.WithAPIKey(apiKey), option.WithBaseURL(baseURL),
	}, extraOpts...)
	client := openai.NewClient(opts...)

	systemMessage := openai.SystemMessage(systemPrompt)

	return &ProviderOpenAI{baseURL, apiKey, broker, &client, providerID, systemMessage}
}

func (p *ProviderOpenAI) ID() string {
	return p.providerID
}

func applyVariant(spec *variant.Spec, apiKey string) []option.RequestOption {
	if spec == nil {
		return nil
	}
	return []option.RequestOption{option.WithMiddleware(variantMiddleware(spec, apiKey))}
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

func (p *ProviderOpenAI) buildHistory(history []domain.Message) []openai.ChatCompletionMessageParamUnion {
	messages := []openai.ChatCompletionMessageParamUnion{p.systemMessage}

	for _, m := range history {
		switch m.Role {
		case domain.RoleUser:
			var content []openai.ChatCompletionContentPartUnionParam
			for _, part := range m.Parts {
				if part.Type != domain.PartTextType {
					continue
				}
				text, ok := part.Data.(domain.TextPartData)
				if !ok {
					continue
				}
				content = append(content, openai.TextContentPart(text.Text))
			}
			messages = append(messages, openai.UserMessage(content))
		case domain.RoleAssistant:
			assistant := openai.ChatCompletionAssistantMessageParam{
				ToolCalls: p.ToProviderToolCalls(m.Parts),
			}

			if content := p.ToProviderParts(m.Parts); len(content) > 0 {
				assistant.Content = openai.ChatCompletionAssistantMessageParamContentUnion{
					OfArrayOfContentParts: content,
				}
			}
			messages = append(messages, openai.ChatCompletionMessageParamUnion{OfAssistant: &assistant})
		case domain.RoleTool:
			messages = append(messages, p.ToProviderToolResults(m.Parts)...)
		}
	}
	return messages
}

const streamFlushInterval = 50 * time.Millisecond

func chunkText(current openai.ChatCompletionChunk) string {
	if len(current.Choices) == 0 {
		return ""
	}
	return current.Choices[0].Delta.Content
}

func (p *ProviderOpenAI) StreamChat(ctx context.Context, modelId string, history []domain.Message) (result StreamResult, err error) {
	stream := p.client.Chat.Completions.NewStreaming(
		ctx, openai.ChatCompletionNewParams{
			Messages:      p.buildHistory(history),
			Model:         modelId,
			Tools:         buildOpenAIToolParams(tools.Definitions),
			StreamOptions: openai.ChatCompletionStreamOptionsParam{},
		},
	)
	var acc openai.ChatCompletionAccumulator

	var (
		outputBuilder   strings.Builder
		thinkingBuilder strings.Builder
	)

	flush := func() {
		if thinkingBuilder.Len() > 0 {
			result.ThinkingText += thinkingBuilder.String()
			p.broker.Publish(ctx, broker.Message{
				Type: broker.EventAgentThinking,
				Text: thinkingBuilder.String(),
			})
			thinkingBuilder.Reset()
		}
		if outputBuilder.Len() > 0 {
			p.broker.Publish(ctx, broker.Message{
				Type: broker.EventAgentResponse,
				Text: outputBuilder.String(),
			})
			outputBuilder.Reset()
		}
	}

	flushTicker := time.NewTicker(streamFlushInterval)
	defer flushTicker.Stop()

	for stream.Next() {
		current := stream.Current()
		acc.AddChunk(current)
		if hasReasoning(current.RawJSON()) {
			if reasoning := reasoningDelta(current.RawJSON()); reasoning != "" {
				thinkingBuilder.WriteString(reasoning)
			}
		}

		if content, ok := acc.JustFinishedContent(); ok {
			slog.Debug("content finished", "content", content)
		}

		if tool, ok := acc.JustFinishedToolCall(); ok {
			result.ToolCalls = append(result.ToolCalls, tools.Call{
				ID:   tool.ID,
				Name: tools.ToolType(tool.Name),
				Args: json.RawMessage(tool.Arguments),
			})
		}

		if text := chunkText(current); text != "" {
			outputBuilder.WriteString(text)
		}

		select {
		case <-flushTicker.C:
			flush()
		default:
		}
	}
	if err := stream.Err(); err != nil {
		slog.Error("openai stream error", "error", err)
		return result, err
	}

	flush()

	if len(acc.Choices) > 0 {
		result.Text = acc.Choices[0].Message.Content
		slog.Debug("Full content openai", "content", result.Text)
	}

	return result, nil
}

// hasReasoning reports whether a raw chunk JSON carries reasoning
// content, without unmarshalling into a struct.
func hasReasoning(raw string) bool {
	return strings.Contains(raw, "reasoning")
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

func (p *ProviderOpenAI) ToProviderParts(parts []domain.Part) []openai.ChatCompletionAssistantMessageParamContentArrayOfContentPartUnion {
	content := make([]openai.ChatCompletionAssistantMessageParamContentArrayOfContentPartUnion, 0, len(parts))
	for _, part := range parts {
		switch part.Type {
		case domain.PartTextType:
			text, ok := part.Data.(domain.TextPartData)
			if !ok {
				continue
			}
			content = append(content, openai.ChatCompletionAssistantMessageParamContentArrayOfContentPartUnion{
				OfText: &openai.ChatCompletionContentPartTextParam{Text: text.Text},
			})
		}
	}
	return content
}

func (p *ProviderOpenAI) ToProviderToolCalls(parts []domain.Part) []openai.ChatCompletionMessageToolCallUnionParam {
	var toolCalls []openai.ChatCompletionMessageToolCallUnionParam
	for _, part := range parts {
		if part.Type != domain.PartToolCallType {
			continue
		}
		tc, ok := part.Data.(domain.ToolCallPartData)
		if !ok {
			continue
		}
		toolCalls = append(toolCalls, openai.ChatCompletionMessageToolCallUnionParam{
			OfFunction: &openai.ChatCompletionMessageFunctionToolCallParam{
				ID: tc.ID,
				Function: openai.ChatCompletionMessageFunctionToolCallFunctionParam{
					Name:      tc.Name,
					Arguments: string(tc.Input),
				},
			},
		})
	}
	return toolCalls
}

func (p *ProviderOpenAI) ToProviderToolResults(parts []domain.Part) []openai.ChatCompletionMessageParamUnion {
	var messages []openai.ChatCompletionMessageParamUnion
	for _, part := range parts {
		if part.Type != domain.PartToolResultType {
			continue
		}
		tr, ok := part.Data.(domain.ToolResultPartData)
		if !ok {
			continue
		}
		content := tr.Content
		if tr.Data != "" {
			content += tr.Data
		}
		messages = append(messages, openai.ToolMessage(content, tr.ToolCallID))
	}
	return messages
}
