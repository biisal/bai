package agent

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"unicode/utf8"

	"charm.land/fantasy"
	"github.com/biisal/bai/internal/config"
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
		conversation, err := g.AddNewConversation(ctx, titleFor(title))
		if err != nil {
			return err
		}
		g.mu.Lock()
		g.conversation = &conversation
		g.mu.Unlock()
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

var ErrWrongDirectory = errors.New("conversation belongs to another directory")

const maxTitleRunes = 80

func titleFor(text string) string {
	text = strings.TrimSpace(text)
	if utf8.RuneCountInString(text) <= maxTitleRunes {
		return text
	}
	return string([]rune(text)[:maxTitleRunes])
}

// ConversationInDir returns the conversation only if it belongs to the current directory.
func (g *Gateway) ConversationInDir(ctx context.Context, id int64) (repo.Conversation, error) {
	conversation, err := g.db.GetConversation(ctx, id)
	if err != nil {
		return repo.Conversation{}, err
	}
	if conversation.Directory != files.CurrentDir() {
		return repo.Conversation{}, ErrWrongDirectory
	}
	return conversation, nil
}

// ResolveConversation returns the conversation for id, or creates one titled
// from text when id is unset, missing, or belongs to another directory.
func (g *Gateway) ResolveConversation(ctx context.Context, id int64, text string) (repo.Conversation, bool, error) {
	if id > 0 {
		conversation, err := g.ConversationInDir(ctx, id)
		if err == nil {
			return conversation, false, nil
		}
		if !errors.Is(err, sql.ErrNoRows) && !errors.Is(err, ErrWrongDirectory) {
			return repo.Conversation{}, false, err
		}
	}
	conversation, err := g.AddNewConversation(ctx, titleFor(text))
	return conversation, true, err
}

func (g *Gateway) Models() []config.ProviderConfig { return g.cfg.Providers }

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

func (g *Gateway) GetProvider(ctx context.Context) (repo.Provider, error) {
	return g.db.GetProvider(ctx)
}

func (g *Gateway) GetConversation(ctx context.Context, id int64) (repo.Conversation, error) {
	return g.db.GetConversation(ctx, id)
}

func (g *Gateway) GetMessagesByConversationID(ctx context.Context, conversationID int64) ([]fantasy.Message, error) {
	messages, err := g.db.GetMessagesByConversation(ctx, conversationID)
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

func (g *Gateway) SetThemeToDB(ctx context.Context, theme string) error {
	return g.db.AddOrUpdateSettings(ctx, sql.NullString{Valid: true, String: theme})
}

func (g *Gateway) GetMessageRowsByConversationID(ctx context.Context, conversationID int64) ([]repo.Message, error) {
	return g.db.GetMessagesByConversation(ctx, conversationID)
}

func (g *Gateway) RenameConversation(ctx context.Context, id int64, title string) (repo.Conversation, error) {
	return g.db.UpdateConversationTitle(ctx, repo.UpdateConversationTitleParams{ID: id, Title: title})
}

func (g *Gateway) DeleteConversation(ctx context.Context, id int64) error {
	if _, err := g.db.DeleteMessagesByConversation(ctx, id); err != nil {
		return err
	}
	_, err := g.db.DeleteConversation(ctx, id)
	return err
}
