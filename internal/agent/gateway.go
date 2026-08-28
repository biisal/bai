package agent

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	fantasy "charm.land/fantasy"
	"github.com/biisal/bai/internal/agent/core/instruction"
	"github.com/biisal/bai/internal/agent/core/tools"
	repo "github.com/biisal/bai/internal/db/sqlc"
	broker "github.com/biisal/bai/internal/pubsub"
)

type Gateway struct {
	mu             sync.RWMutex
	broker         broker.Service
	providers      map[string]fantasy.Provider
	db             repo.Querier
	conversation   *repo.Conversation
	activeProvider fantasy.Provider
	activeModel    string
}

func NewGateway(
	db repo.Querier,
	b broker.Service,
	providers map[string]fantasy.Provider,
	providerID, modelID string,
) *Gateway {
	g := &Gateway{
		db:        db,
		broker:    b,
		providers: providers,
	}
	if err := g.SetActive(providerID, modelID); err != nil {
		return nil
	}
	return g
}

func (g *Gateway) ActiveConversationTitle() string {
	if g.conversation == nil {
		return "bai | Start a new conversation"
	}
	return fmt.Sprintf("bai | %s", g.conversation.Title)
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

func (g *Gateway) Active() (fantasy.Provider, string) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.activeProvider, g.activeModel
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

func (g *Gateway) SetConversation(conversation *repo.Conversation) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.conversation = conversation
}

type ProviderResponse struct {
	Content string
}

func (g *Gateway) trySavingMsgToDB(partialReasoning, partialText *strings.Builder) {
	var parts []fantasy.MessagePart
	if partialReasoning.Len() > 0 {
		parts = append(parts, fantasy.ReasoningPart{Text: partialReasoning.String()})
	}
	if partialText.Len() > 0 {
		parts = append(parts, fantasy.TextPart{Text: partialText.String()})
	}

	if len(parts) == 0 {
		return
	}

	saveCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if saveErr := g.AddMessageToDB(saveCtx, fantasy.Message{
		Role:    fantasy.MessageRoleAssistant,
		Content: parts,
	}); saveErr != nil {
		slog.Error("failed to save interrupted message", "error", saveErr)
	}
}

func (g *Gateway) StreamChat(ctx context.Context, message string) (*ProviderResponse, error) {
	// 1. Save user message.
	if err := g.AddMessageToDB(ctx, fantasy.Message{
		Role:    fantasy.MessageRoleUser,
		Content: []fantasy.MessagePart{fantasy.TextPart{Text: message}},
	}); err != nil {
		slog.Error("failed to add user message to db", "error", err)
		return nil, err
	}

	// 2. Load conversation history.
	history, err := g.GetMessagesByConversationID(ctx, g.conversation.ID)
	if err != nil {
		slog.Error("failed to get messages", "error", err)
		return nil, err
	}

	provider, modelID := g.Active()
	model, err := provider.LanguageModel(ctx, modelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get language model: %w", err)
	}

	agentTools := tools.NewTools(g.broker)
	ag := fantasy.NewAgent(
		model,
		fantasy.WithSystemPrompt(instruction.BuildSystemPrompt()),
		fantasy.WithTools(agentTools...),
		fantasy.WithMaxRetries(3),
	)

	var partialReasoning strings.Builder
	var partialText strings.Builder

	result, err := ag.Stream(ctx, fantasy.AgentStreamCall{
		Messages: history,

		OnRetry: fantasy.DefaultRetryOptions().OnRetry,

		OnToolCall: func(toolCall fantasy.ToolCallContent) error {
			slog.Debug("tool call", "input", toolCall.Input, "name", toolCall.ToolName)
			return nil
		},

		OnAgentStart: func() {
			g.broker.Publish(ctx, broker.Message{Type: broker.EventStreamStarted, IsComplete: true})
		},

		OnTextDelta: func(_ string, text string) error {
			partialText.WriteString(text)
			g.broker.Publish(ctx, broker.Message{Type: broker.EventAgentResponse, Text: text})
			return nil
		},

		OnReasoningDelta: func(_ string, text string) error {
			partialReasoning.WriteString(text)
			g.broker.Publish(ctx, broker.Message{Type: broker.EventAgentThinking, Text: text})
			return nil
		},

		OnStepFinish: func(step fantasy.StepResult) error {
			partialReasoning.Reset()
			partialText.Reset()
			for _, fm := range step.Messages {
				if saveErr := g.AddMessageToDB(ctx, fm); saveErr != nil {
					slog.Error("failed to save step message", "error", saveErr)
				}
			}
			return nil
		},
	})
	if err != nil {
		g.trySavingMsgToDB(&partialReasoning, &partialText)
		return nil, err
	}

	return &ProviderResponse{Content: result.Response.Content.Text()}, nil
}
