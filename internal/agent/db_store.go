package agent

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"

	"charm.land/fantasy"
	repo "github.com/biisal/bai/internal/db/sqlc"
	"github.com/biisal/bai/internal/domain"
	"github.com/biisal/bai/internal/files"
)

func (g *Gateway) AddMessageToDB(ctx context.Context, msg fantasy.Message) error {
	if g.conversation == nil {
		title := "Untitled Conversation"
		for _, part := range msg.Content {
			if text, ok := part.(fantasy.TextPart); ok && text.Text != "" {
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
	partsBytes, err := json.Marshal(msg.Content)
	if err != nil {
		return err
	}
	_, err = g.db.CreateMessage(ctx, repo.CreateMessageParams{
		ConversationID: g.conversation.ID,
		Role:           domain.Role(msg.Role),
		Parts:          string(partsBytes),
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

func (G *Gateway) GetMessagesByConversationID(ctx context.Context, conversationID int64) ([]fantasy.Message, error) {
	messages, err := G.db.GetMessagesByConversation(ctx, conversationID)
	if err != nil {
		return nil, err
	}
	fantasyMessages := make([]fantasy.Message, 0, len(messages))
	for _, m := range messages {
		fantasyMessages = append(fantasyMessages, toFantasyMessage(m))
	}
	return fantasyMessages, nil
}

func toFantasyMessage(m repo.Message) fantasy.Message {
	var rawParts []json.RawMessage
	var content []fantasy.MessagePart
	
	err := json.Unmarshal([]byte(m.Parts), &rawParts)
	if err == nil {
		for _, raw := range rawParts {
			part, err := fantasy.UnmarshalMessagePart(raw)
			if err == nil {
				content = append(content, part)
			}
		}
	} else {
		// Fallback for older messages
		content = []fantasy.MessagePart{
			fantasy.TextPart{Text: m.Parts},
		}
	}
	return fantasy.Message{
		Role:    fantasy.MessageRole(m.Role),
		Content: content,
	}
}
