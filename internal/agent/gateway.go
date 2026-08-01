package agent

import (
	"context"
	"fmt"
	"sync"

	"github.com/biisal/bai/internal/agent/providers"
	"github.com/biisal/bai/internal/db"
)

type Gateway struct {
	mu        sync.RWMutex
	providers map[string]providers.Provider
	model     string
	active    providers.Provider
	MsgChan   chan Message
	db        db.ServiceInterface
}

func NewGateway(db db.ServiceInterface, providers map[string]providers.Provider) *Gateway {
	g := &Gateway{
		db:        db,
		providers: providers,
		MsgChan:   make(chan Message),
	}
	for _, provider := range providers {
		g.active = provider
		break
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
	g.active = provider
	g.model = modelID
	return nil
}

func (g *Gateway) Active() (provider providers.Provider, modelID string) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.active, g.model
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

func (g *Gateway) StreamChat(ctx context.Context, system string, history []ReplayMessage) (*ProviderResponse, error) {
	g.active.StreamChat(g.model)
	return nil, nil
}
