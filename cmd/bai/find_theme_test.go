package main

import (
	"context"
	"database/sql"
	"testing"

	"github.com/biisal/bai/internal/config"
	"github.com/biisal/bai/internal/db"
	repo "github.com/biisal/bai/internal/db/sqlc"
)

func getTx(t *testing.T, conn *sql.DB) repo.Querier {
	t.Helper()
	tx, err := conn.Begin()
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Cleanup(func() {
		if err := tx.Rollback(); err != nil {
			t.Fatal(err.Error())
		}
	})
	return repo.New(tx)
}

func assertThemeResult(t *testing.T, gotName, gotDir, wantName, wantDir string) {
	t.Helper()
	if wantName != gotName {
		t.Errorf("want theme name %q, got %q", wantName, gotName)
	}
	if wantDir != gotDir {
		t.Errorf("want theme dir %q, got %q", wantDir, gotDir)
	}
}

func TestResolveTheme(t *testing.T) {
	conn, err := db.Connect(":memory:")
	if err != nil {
		t.Fatal(err.Error())
	}
	if err := db.Migrate(context.Background(), conn); err != nil {
		t.Fatalf("migrate db: %v", err)
	}

	tests := []struct {
		name          string
		themes        []config.ThemeFile
		setup         func(ctx context.Context, t *testing.T) repo.Querier
		wantThemeName string
		wantThemeDir  string
	}{
		{
			name: "returns empty when no themes",
			setup: func(ctx context.Context, t *testing.T) repo.Querier {
				return getTx(t, conn)
			},
			wantThemeName: "",
			wantThemeDir:  "",
		},
		{
			name: "saves and returns first theme when db has no settings",
			themes: []config.ThemeFile{
				{Name: "default.json", FilePath: "/fake/path/themes/default.json"},
				{Name: "dark.json", FilePath: "/fake/path/themes/dark.json"},
			},
			setup: func(ctx context.Context, t *testing.T) repo.Querier {
				return getTx(t, conn)
			},
			wantThemeName: "default.json",
			wantThemeDir:  "/fake/path/themes/default.json",
		},
		{
			name: "returns saved theme from db when it exists in filesystem",
			themes: []config.ThemeFile{
				{Name: "default.json", FilePath: "/fake/path/themes/default.json"},
				{Name: "dark.json", FilePath: "/fake/path/themes/dark.json"},
			},
			setup: func(ctx context.Context, t *testing.T) repo.Querier {
				t.Helper()
				tx := getTx(t, conn)
				if err := tx.AddOrUpdateSettings(ctx, sql.NullString{
					Valid:  true,
					String: "dark.json",
				}); err != nil {
					t.Fatal(err)
				}
				return tx
			},
			wantThemeName: "dark.json",
			wantThemeDir:  "/fake/path/themes/dark.json",
		},
		{
			name: "returns first theme when saved theme no longer exists in filesystem",
			themes: []config.ThemeFile{
				{Name: "default.json", FilePath: "/fake/path/themes/default.json"},
				{Name: "light.json", FilePath: "/fake/path/themes/light.json"},
			},
			setup: func(ctx context.Context, t *testing.T) repo.Querier {
				t.Helper()
				tx := getTx(t, conn)
				if err := tx.AddOrUpdateSettings(ctx, sql.NullString{
					Valid:  true,
					String: "dark.json",
				}); err != nil {
					t.Fatal(err)
				}
				return tx
			},
			wantThemeName: "default.json",
			wantThemeDir:  "/fake/path/themes/default.json",
		},
		{
			name: "saves first theme to db when falling back",
			themes: []config.ThemeFile{
				{Name: "nord.json", FilePath: "/fake/path/themes/nord.json"},
			},
			setup: func(ctx context.Context, t *testing.T) repo.Querier {
				t.Helper()
				tx := getTx(t, conn)
				// verify no settings exist yet
				_, err := tx.GetSettings(ctx)
				if err == nil {
					t.Fatal("expected no settings row")
				}
				return tx
			},
			wantThemeName: "nord.json",
			wantThemeDir:  "/fake/path/themes/nord.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			tx := tt.setup(ctx, t)
			gotName, gotDir := resolveTheme(ctx, tx, tt.themes)
			assertThemeResult(t, gotName, gotDir, tt.wantThemeName, tt.wantThemeDir)
		})
	}
}

func TestResolveTheme_updatesDbOnFallback(t *testing.T) {
	conn, err := db.Connect(":memory:")
	if err != nil {
		t.Fatal(err.Error())
	}
	if err := db.Migrate(context.Background(), conn); err != nil {
		t.Fatalf("migrate db: %v", err)
	}

	ctx := context.Background()
	tx := getTx(t, conn)
	themes := []config.ThemeFile{
		{Name: "ocean.json", FilePath: "/fake/ocean.json"},
	}

	gotName, gotDir := resolveTheme(ctx, tx, themes)
	assertThemeResult(t, gotName, gotDir, "ocean.json", "/fake/ocean.json")

	// Verify the theme was persisted to the db
	settings, err := tx.GetSettings(ctx)
	if err != nil {
		t.Fatalf("expected settings to be saved: %v", err)
	}
	if !settings.Theme.Valid || settings.Theme.String != "ocean.json" {
		t.Errorf("expected db to have theme 'ocean.json', got %v", settings.Theme)
	}
}
