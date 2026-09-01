package main

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"github.com/biisal/bai/internal/config"
	repo "github.com/biisal/bai/internal/db/sqlc"
)

func resolveTheme(ctx context.Context, svc repo.Querier, themes []config.ThemeFile) (themeName, themeDir string) {
	if len(themes) == 0 {
		return "", ""
	}

	// Get the saved theme from DB, or use the first theme from the filesystem
	settings, err := svc.GetSettings(ctx)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		slog.Error("failed to get settings", "err", err)
		return "", ""
	}

	// Check if the saved theme still exists in the filesystem
	if settings.Theme.Valid {
		for _, tf := range themes {
			if tf.Name == settings.Theme.String {
				return tf.Name, tf.FilePath
			}
		}
	}

	// Fallback: save the first theme to DB and return it
	first := themes[0]

	if err := svc.AddOrUpdateSettings(ctx, sql.NullString{
		Valid:  true,
		String: first.Name,
	}); err != nil {
		slog.Error("failed to add/update settings", "theme", first.Name, "err", err)
	}

	return first.Name, first.FilePath
}
