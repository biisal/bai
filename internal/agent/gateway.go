package agent

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

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
	activeProvider string
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

func (g *Gateway) SetActive(providerID, modelID string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if _, ok := g.providers[providerID]; !ok {
		return fmt.Errorf("unknown provider: %s", providerID)
	}
	g.activeProvider = providerID
	g.activeModel = modelID
	return nil
}

func (g *Gateway) Active() (providerID, modelID string) {
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

	// 3. Resolve the active language model.
	g.mu.RLock()
	provider := g.providers[g.activeProvider]
	modelID := g.activeModel
	g.mu.RUnlock()

	model, err := provider.LanguageModel(ctx, modelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get language model: %w", err)
	}

	// 4. Build the agent with tools and system prompt.
	agentTools := tools.NewTools(g.broker)
	ag := fantasy.NewAgent(
		model,
		fantasy.WithSystemPrompt(instruction.BuildSystemPrompt()),
		fantasy.WithTools(agentTools...),
	)

	// 5. Stream the response, publishing to the broker and saving to DB.
	g.broker.Publish(ctx, broker.Message{Type: broker.EventStreamStarted, IsComplete: true})

	savedLen := len(history) // skip re-saving history messages

	result, err := ag.Stream(ctx, fantasy.AgentStreamCall{
		Messages: history,

		OnTextDelta: func(_ string, text string) error {
			g.broker.Publish(ctx, broker.Message{Type: broker.EventAgentResponse, Text: text})
			return nil
		},

		OnReasoningDelta: func(_ string, text string) error {
			g.broker.Publish(ctx, broker.Message{Type: broker.EventAgentThinking, Text: text})
			return nil
		},

		OnStepFinish: func(step fantasy.StepResult) error {
			// Save only the messages generated in this step.
			newMsgs := step.Messages
			if len(newMsgs) > savedLen {
				newMsgs = newMsgs[savedLen:]
			}
			for _, fm := range newMsgs {
				if saveErr := g.AddMessageToDB(ctx, fm); saveErr != nil {
					slog.Error("failed to save step message", "error", saveErr)
				}
			}
			savedLen = len(step.Messages)
			return nil
		},
	})
	if err != nil {
		return nil, err
	}

	return &ProviderResponse{Content: result.Response.Content.Text()}, nil
}
