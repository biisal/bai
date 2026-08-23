package agent

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/biisal/bai/internal/agent/core/tools"
	"github.com/biisal/bai/internal/agent/providers"
	repo "github.com/biisal/bai/internal/db/sqlc"
	"github.com/biisal/bai/internal/domain"
	broker "github.com/biisal/bai/internal/pubsub"
)

type Gateway struct {
	mu             sync.RWMutex
	broker         broker.Service
	providers      map[string]providers.Provider
	db             repo.Querier
	conversation   *repo.Conversation
	activeProvider providers.Provider
	activeModel    string
}

func NewGateway(db repo.Querier, b broker.Service, providers map[string]providers.Provider, provideId, modelID string) *Gateway {
	g := &Gateway{
		db:        db,
		broker:    b,
		providers: providers,
	}

	if err := g.SetActive(provideId, modelID); err != nil {
		return nil
	}

	return g
}

func (g *Gateway) SetActive(providerID, modelID string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	provider, ok := g.providers[providerID]
	if !ok {
		return fmt.Errorf("unknown provider: %s", providerID)
	}
	g.activeProvider = provider
	g.activeModel = modelID
	return nil
}

func (g *Gateway) Active() (provider providers.Provider, modelID string) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	provider = g.providers[g.activeProvider.ID()]
	return provider, g.activeModel
}

func (g *Gateway) Providers() []string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	ids := make([]string, 0, len(g.providers))
	for id := range g.providers {
		ids = append(ids, id)
	}
	return ids
}

func (g *Gateway) Models(providerID string) []string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return nil
}

func (g *Gateway) SetConversation(conversation *repo.Conversation) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.conversation = conversation
}

func (g *Gateway) StreamChat(ctx context.Context, message string) (*ProviderResponse, error) {
	if err := g.AddMessageToDB(ctx, domain.Message{
		Role: domain.RoleUser,
		Parts: []domain.Part{
			{Type: domain.PartTextType, Data: domain.TextPartData{Text: message}},
		},
	}); err != nil {
		slog.Error("failed to add user message to db", "error", err)
		return nil, err
	}
	messages, err := g.GetMessagesByConversationID(ctx, g.conversation.ID)
	if err != nil {
		slog.Error("failed to get messages", "error", err)
		return nil, err
	}

	for round := 0; ; round++ {
		if round > 0 {
			slog.Debug("agent loop: next round after tool calls", "round", round,
				"history_len", len(messages))
		}
		g.broker.Publish(ctx, broker.Message{Type: broker.EventStreamStarted})
		result, err := g.activeProvider.StreamChat(ctx, g.activeModel, messages)
		if err != nil {
			slog.Error("failed to stream chat", "error", err)
			return nil, err
		}

		slog.Info("stream chat result", "text", result.Text, "tool_calls", len(result.ToolCalls))

		assistantMsg := assistantMessageFromResult(result)
		if err := g.AddMessageToDB(ctx, assistantMsg); err != nil {
			slog.Error("failed to add assistant message to db", "error", err)
			return nil, err
		}
		messages = append(messages, assistantMsg)

		if len(result.ToolCalls) == 0 {
			return &ProviderResponse{Content: result.Text}, nil
		}

		toolMsg := domain.Message{Role: domain.RoleTool}
		for _, tc := range result.ToolCalls {

			slog.Info("tool_call", "tc", tc)

			out, isErr := tools.Execute(ctx, tc, g.broker)
			toolMsg.Parts = append(toolMsg.Parts, domain.Part{
				Type: domain.PartToolResultType,
				Data: domain.ToolResultPartData{ToolCallID: tc.ID, Name: string(tc.Name), Content: out, IsError: isErr},
			})
		}
		if err := g.AddMessageToDB(ctx, toolMsg); err != nil {
			slog.Error("failed to add tool result message to db", "error", err)
			return nil, err
		}
		messages = append(messages, toolMsg)
	}
}

type ProviderResponse struct {
	Content string
}

func assistantMessageFromResult(result providers.StreamResult) domain.Message {
	msg := domain.Message{Role: domain.RoleAssistant}
	if result.ThinkingText != "" {
		msg.Parts = append(msg.Parts, domain.Part{
			Type: domain.PartReasoningType,
			Data: domain.ReasoningPartData{Thinking: result.ThinkingText},
		})
	}
	if result.Text != "" {
		msg.Parts = append(msg.Parts, domain.Part{
			Type: domain.PartTextType,
			Data: domain.TextPartData{Text: result.Text},
		})
	}
	for _, tc := range result.ToolCalls {
		msg.Parts = append(msg.Parts, domain.Part{
			Type: domain.PartToolCallType,
			Data: domain.ToolCallPartData{
				ID:               tc.ID,
				Name:             string(tc.Name),
				Input:            tc.Args,
				ProviderExecuted: false,
				Finished:         true,
			},
		})
	}
	return msg
}
