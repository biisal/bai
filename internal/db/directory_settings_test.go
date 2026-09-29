package db

import (
	"context"
	"testing"

	repo "github.com/biisal/bai/internal/db/sqlc"
)

func TestUpsertDirectorySettings(t *testing.T) {
	conn, err := Connect("file::memory:?_fk=1")
	if err != nil {
		t.Fatalf("Connect() failed: %v", err)
	}
	defer func() { _ = conn.Close() }()

	ctx := context.Background()
	if err := Migrate(ctx, conn); err != nil {
		t.Fatalf("Migrate() failed: %v", err)
	}
	q := repo.New(conn)

	for _, enabled := range []bool{true, false} {
		if err := q.UpsertDirectorySettings(ctx, repo.UpsertDirectorySettingsParams{Directory: "/x", AutoGitInit: enabled}); err != nil {
			t.Fatalf("upsert(%v) failed: %v", enabled, err)
		}
	}
	got, err := q.GetDirectorySettings(ctx, "/x")
	if err != nil {
		t.Fatalf("GetDirectorySettings() failed: %v", err)
	}
	if got.AutoGitInit {
		t.Errorf("AutoGitInit = true, want false after second upsert")
	}
}
