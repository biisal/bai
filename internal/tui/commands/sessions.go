package commands

import (
	"context"

	"charm.land/bubbles/v2/list"
	repo "github.com/biisal/bai/internal/db/sqlc"
)

type ConversationItem struct {
	CreatedAt    string
	Conversation repo.Conversation
}

func (ls ConversationItem) Title() string {
	return ls.Conversation.Title
}

func (ls ConversationItem) Description() string {
	return ls.CreatedAt
}

func (ls ConversationItem) FilterValue() string {
	return ls.Conversation.Title
}

func parseConversations(ctx context.Context, getconversations func(ctx context.Context) ([]repo.Conversation, error)) []list.Item {
	var items []list.Item
	conversations, err := getconversations(ctx)
	if err != nil {
		return items
	}
	for _, conv := range conversations {
		items = append(items, ConversationItem{
			CreatedAt:    conv.CreatedAt.Format("Monday, Jan 2, 2006 at 03:04 PM"),
			Conversation: conv,
		})
	}
	return items
}
