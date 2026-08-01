package db

import (
	"context"
	"database/sql"

	repo "github.com/biisal/bai/internal/db/sqlc"
)

type ServiceInterface interface {
	CreateConversation(ctx context.Context, title string) (int64, error)
}

type Service struct {
	q *repo.Queries
}

func New(db *sql.DB) ServiceInterface {
	return &Service{q: repo.New(db)}
}

func (s *Service) CreateConversation(ctx context.Context, title string) (int64, error) {
	return s.q.CreateConversation(ctx, title)
}

func (s *Service) GetConversation(ctx context.Context, id int64) (repo.Conversation, error) {
	return s.q.GetConversation(ctx, id)
}

func (s *Service) ListConversations(ctx context.Context) ([]repo.Conversation, error) {
	return s.q.ListConversations(ctx)
}

func (s *Service) UpdateConversation(ctx context.Context, id int64, title string) error {
	return s.q.UpdateConversation(ctx, repo.UpdateConversationParams{
		ID:    id,
		Title: title,
	})
}

func (s *Service) DeleteConversation(ctx context.Context, id int64) error {
	return s.q.DeleteConversation(ctx, id)
}

func (s *Service) SaveUserMessage(ctx context.Context, conversationID int64, content string) (int64, error) {
	return s.q.CreateMessage(ctx, repo.CreateMessageParams{
		ConversationID: conversationID,
		Role:           "user",
		Content:        content,
	})
}
