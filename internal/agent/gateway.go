package agent

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/biisal/bai/internal/agent/providers"
	"github.com/biisal/bai/internal/db"
	repo "github.com/biisal/bai/internal/db/sqlc"
)

type Gateway struct {
	mu             sync.RWMutex
	providers      map[string]providers.Provider
	db             db.ServiceInterface
	conversation   *repo.Conversation
	activeProvider providers.Provider
	activeModel    string
}

func NewGateway(db db.ServiceInterface, providers map[string]providers.Provider, provideId, modelID string) *Gateway {
	g := &Gateway{
		db:        db,
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

	// TODO : can produce bug if provider is not found
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

func (g *Gateway) StreamChat(ctx context.Context, meessage string) (*ProviderResponse, error) {
	if err := g.AddMessageToDB(ctx, meessage, db.RoleUser); err != nil {
		slog.Error("failed to add user message to db", "error", err)
		return nil, err
	}
	messages, err := g.GetMessagesByConversationID(ctx, g.conversation.ID)
	if err != nil {
		slog.Error("failed to get messages", "error", err)
		return nil, err
	}

	messages = append(messages, repo.Message{
		Role:    "user",
		Content: meessage,
	})

	finalResp, err := g.activeProvider.StreamChat(ctx, g.activeModel, messages)
	if err != nil {
		slog.Error("failed to stream chat", "error", err)
		return nil, err
	}

	if err := g.AddMessageToDB(ctx, finalResp, db.RoleAssistant); err != nil {
		slog.Error("failed to add user message to db", "error", err)
		return nil, err
	}
	return &ProviderResponse{
		Content: finalResp,
	}, nil
}
