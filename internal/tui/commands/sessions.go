package commands

import (
	"context"
	"log/slog"

	"charm.land/bubbles/v2/list"
	repo "github.com/biisal/bai/internal/db/sqlc"
)

type ConversationItem struct {
	ID           int64
	Name         string
	Desc         string
	Conversation repo.Conversation
}

func (ls ConversationItem) Title() string {
	return ls.Name
}

func (ls ConversationItem) Description() string {
	return ls.Desc
}

func (ls ConversationItem) FilterValue() string {
	return ls.Name
}

func parseConversations(ctx context.Context, getconversations func(ctx context.Context) ([]repo.Conversation, error)) []list.Item {
	slog.Info("parseConversations", "ctx", ctx)
	var items []list.Item
	conversations, err := getconversations(ctx)
	if err != nil {
		return items
	}
	for _, conv := range conversations {
		items = append(items, ConversationItem{
			ID:           conv.ID,
			Name:         conv.Title,
			Desc:         conv.CreatedAt.Format("Monday, Jan 2, 2006 at 03:04 PM"),
			Conversation: conv,
		})
	}
	slog.Info("conversatons", "items", len(items))
	return items
}
