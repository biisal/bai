package agent

import (
	"context"
	"log/slog"

	"github.com/biisal/bai/internal/db"
	repo "github.com/biisal/bai/internal/db/sqlc"
	"github.com/biisal/bai/internal/files"
)

func (g *Gateway) AddMessageToDB(ctx context.Context, message string, role db.Role) error {
	if g.conversation == nil {
		title := "Untitled Conversation"
		if message != "" {
			title = message
		}
		conversation, err := g.AddNewConversation(ctx, title)
		if err != nil {
			return err
		}
		g.mu.RLock()
		g.conversation = &conversation
		g.mu.RUnlock()
	}
	_, err := g.db.CreateMessage(ctx, g.conversation.ID, message, role)
	return err
}

func (g *Gateway) AddNewConversation(ctx context.Context, title string) (repo.Conversation, error) {
	return g.db.CreateConversation(ctx, title, files.CurrentDir())
}

func (g *Gateway) GetConversationsByCurrentDir(ctx context.Context) ([]repo.Conversation, error) {
	return g.db.GetConversatonsByDir(ctx, files.CurrentDir())
}

func (g *Gateway) SetActiveConversation(ctx context.Context, conversationID int64, conversation *repo.Conversation) error {
	if conversation == nil {
		conv, err := g.db.GetConversation(ctx, conversationID)
		if err != nil {
			slog.Error("set_active_conversation", "conversationID", conversationID, "err", err)
			return err
		}
		conversation = &conv
	}

	g.mu.Lock()
	g.conversation = conversation
	g.mu.Unlock()
	return nil
}

func (g *Gateway) AddOrUpdateProvider(ctx context.Context, providerName, modelID string) error {
	if err := g.db.AddOrUpdateProvider(ctx, providerName, modelID); err != nil {
		return err
	}

	return g.SetActive(providerName, modelID)
}

func (g *Gateway) GetProvider(ctx context.Context) (repo.Provider, error) {
	return g.db.GetProvider(ctx)
}

func (G *Gateway) GetMessagesByConversationID(ctx context.Context, conversationID int64) ([]repo.Message, error) {
	return G.db.GetMessagesByConversationID(ctx, conversationID)
}
