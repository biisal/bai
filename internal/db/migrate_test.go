package db

import (
	"context"
	"testing"
)

func TestMigrate(t *testing.T) {
	conn, err := Connect("file::memory:?_fk=1")
	if err != nil {
		t.Fatalf("Connect() failed: %v", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			t.Fatalf("Close() failed: %v", err)
		}
	}()

	ctx := context.Background()

	if err := Migrate(ctx, conn); err != nil {
		t.Fatalf("Migrate() failed on fresh db: %v", err)
	}

	want := []string{
		"chat_completion",
		"conversations",
		"goose_db_version",
		"messages",
		"provider",
		"tool_calls",
	}
	for _, name := range want {
		var exists int
		err := conn.QueryRow(
			"SELECT count(*) FROM sqlite_master WHERE type='table' AND name=?", name,
		).Scan(&exists)
		if err != nil {
			t.Fatalf("check table %s: %v", name, err)
		}
		if exists != 1 {
			t.Errorf("expected table %s to exist", name)
		}
	}

	if err := Migrate(ctx, conn); err != nil {
		t.Fatalf("Migrate() failed on second run: %v", err)
	}
}
