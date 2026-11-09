package agent

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/biisal/bai/internal/config"
	"github.com/biisal/bai/internal/db"
	repo "github.com/biisal/bai/internal/db/sqlc"
	"github.com/biisal/bai/internal/files"
)

func TestResolveConversationAndModels(t *testing.T) {
	conn, err := db.Connect(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	ctx := context.Background()
	if err := db.Migrate(ctx, conn); err != nil {
		t.Fatal(err)
	}
	q := repo.New(conn)
	g := &Gateway{db: q, cfg: &config.Config{Providers: []config.ProviderConfig{
		{Name: "p", Models: []config.ModelConfig{{ID: "m"}}},
	}}}

	// No id: creates a conversation with a truncated title.
	long := strings.Repeat("é", maxTitleRunes+10)
	created, isNew, err := g.ResolveConversation(ctx, 0, "  "+long+"  ")
	if err != nil || !isNew || len([]rune(created.Title)) != maxTitleRunes {
		t.Fatalf("new: %+v isNew=%v err=%v", created, isNew, err)
	}

	// Existing id in this directory: reused.
	got, isNew, err := g.ResolveConversation(ctx, created.ID, "ignored")
	if err != nil || isNew || got.ID != created.ID {
		t.Fatalf("existing: %+v isNew=%v err=%v", got, isNew, err)
	}

	// Conversation from another directory: ErrWrongDirectory, and resolve starts a new one.
	other, err := q.CreateConversation(ctx, repo.CreateConversationParams{Title: "x", Directory: files.CurrentDir() + "-other"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := g.ConversationInDir(ctx, other.ID); !errors.Is(err, ErrWrongDirectory) {
		t.Fatalf("want ErrWrongDirectory, got %v", err)
	}
	if got, isNew, err := g.ResolveConversation(ctx, other.ID, "t"); err != nil || !isNew || got.ID == other.ID {
		t.Fatalf("other dir: %+v isNew=%v err=%v", got, isNew, err)
	}

	// Unknown id: new conversation.
	if _, isNew, err := g.ResolveConversation(ctx, 9999, "t"); err != nil || !isNew {
		t.Fatalf("missing: isNew=%v err=%v", isNew, err)
	}

	if err := g.AddOrUpdateProvider(ctx, "p", "nope"); !errors.Is(err, ErrUnknownModel) {
		t.Fatalf("want ErrUnknownModel, got %v", err)
	}
}
