package agent

import (
	"context"
	"database/sql"
	"log/slog"

	repo "github.com/biisal/bai/internal/db/sqlc"
	"github.com/biisal/bai/internal/domain"
	"github.com/biisal/bai/internal/files"
)

func (g *Gateway) AddMessageToDB(ctx context.Context, msg domain.Message) error {
	if g.conversation == nil {
		title := "Untitled Conversation"
		for _, part := range msg.Parts {
			if part.Type != domain.PartTextType {
				continue
			}
			if text, ok := part.Data.(domain.TextPartData); ok && text.Text != "" {
				title = text.Text
				break
			}
		}
		conversation, err := g.AddNewConversation(ctx, title)
		if err != nil {
			return err
		}
		g.mu.RLock()
		g.conversation = &conversation
		g.mu.RUnlock()
	}
	parts, err := domain.MarshalParts(msg.Parts)
	if err != nil {
		return err
	}
	_, err = g.db.CreateMessage(ctx, repo.CreateMessageParams{
		ConversationID: g.conversation.ID,
		Role:           msg.Role,
		Parts:          string(parts),
	})
	return err
}

func (g *Gateway) AddNewConversation(ctx context.Context, title string) (repo.Conversation, error) {
	return g.db.CreateConversation(ctx, repo.CreateConversationParams{
		Title:     title,
		Directory: files.CurrentDir(),
	})
}

func (g *Gateway) GetConversationsByCurrentDir(ctx context.Context) ([]repo.Conversation, error) {
	return g.db.GetConversationsByDirectory(ctx, files.CurrentDir())
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
	if err := g.db.AddOrUpdateProvider(ctx, repo.AddOrUpdateProviderParams{
		ProviderName: sql.NullString{Valid: true, String: providerName},
		ModelID:      sql.NullString{Valid: true, String: modelID},
	}); err != nil {
		return err
	}

	return g.SetActive(providerName, modelID)
}

func (g *Gateway) GetProvider(ctx context.Context) (repo.Provider, error) {
	return g.db.GetProvider(ctx)
}

func (G *Gateway) GetMessagesByConversationID(ctx context.Context, conversationID int64) ([]domain.Message, error) {
	messages, err := G.db.GetMessagesByConversation(ctx, conversationID)
	if err != nil {
		return nil, err
	}
	domainMessages := make([]domain.Message, 0, len(messages))
	for _, m := range messages {
		domainMessages = append(domainMessages, toDomainMessage(m))
	}
	return domainMessages, nil
}

func toDomainMessage(m repo.Message) domain.Message {
	parts, err := domain.UnmarshalParts([]byte(m.Parts))
	if err != nil {
		parts = []domain.Part{
			{
				Type: domain.PartTextType,
				Data: domain.TextPartData{Text: m.Parts},
			},
		}
	}
	return domain.Message{
		Role:  m.Role,
		Parts: parts,
	}
}
