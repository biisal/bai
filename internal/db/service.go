package db

import (
	"context"
	"database/sql"

	repo "github.com/biisal/bai/internal/db/sqlc"
)

type ServiceInterface interface {
	CreateConversation(ctx context.Context, title string, dir string) (repo.Conversation, error)
	CreateMessage(ctx context.Context, conversationID int64, content string, role Role) (int64, error)
	GetConversatonsByDir(ctx context.Context, dir string) ([]repo.Conversation, error)
	GetConversation(ctx context.Context, id int64) (repo.Conversation, error)
	AddOrUpdateProvider(ctx context.Context, name, providerID, modelID string) error
	GetProvider(ctx context.Context) (repo.Provider, error)

	GetMessagesByConversationID(ctx context.Context, conversationID int64) ([]repo.Message, error)
	WithTx(tx *sql.Tx) ServiceInterface
}

type Service struct {
	q *repo.Queries
}

func New(db *sql.DB) ServiceInterface {
	return &Service{q: repo.New(db)}
}

func (s *Service) WithTx(tx *sql.Tx) ServiceInterface {
	return &Service{q: repo.New(tx)}
}

func (s *Service) CreateConversation(ctx context.Context, title, dir string) (repo.Conversation, error) {
	return s.q.CreateConversation(ctx, repo.CreateConversationParams{
		Title:     title,
		Directory: dir,
	})
}

func (s *Service) GetConversation(ctx context.Context, id int64) (repo.Conversation, error) {
	return s.q.GetConversation(ctx, id)
}

func (s *Service) ListConversations(ctx context.Context) ([]repo.ListConversationsRow, error) {
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

type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

func (s *Service) CreateMessage(ctx context.Context, conversationID int64, content string, role Role) (int64, error) {
	return s.q.CreateMessage(ctx, repo.CreateMessageParams{
		ConversationID: conversationID,
		Role:           string(role),
		Content:        content,
	})
}

func (s *Service) GetConversatonsByDir(ctx context.Context, dir string) ([]repo.Conversation, error) {
	return s.q.GetConversationsByDirectory(ctx, dir)
}

func (s *Service) AddOrUpdateProvider(ctx context.Context, name, providerID, modelID string) error {
	return s.q.AddOrUpdateProvider(ctx, repo.AddOrUpdateProviderParams{
		Name: name, ProviderID: sql.NullString{String: providerID, Valid: true}, ModelID: sql.NullString{String: modelID, Valid: true},
	})
}

func (s *Service) GetProvider(ctx context.Context) (repo.Provider, error) {
	return s.q.GetProvider(ctx)
}

func (s *Service) GetMessagesByConversationID(ctx context.Context, conversationID int64) ([]repo.Message, error) {
	return s.q.GetMessagesByConversation(ctx, conversationID)
}
