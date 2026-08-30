package main

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	repo "github.com/biisal/bai/internal/db/sqlc"
)

func resolveTheme(ctx context.Context, svc repo.Querier, themes map[string]string) (name, dir string) {
	if len(themes) == 0 {
		return "", ""
	}
	settings, err := svc.GetSettings(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// TODO
			for theme, dir := range themes {
				err := svc.AddOrUpdateSettings(ctx, sql.NullString{
					Valid:  true,
					String: theme,
				})
				if err != nil {
					slog.Error("failed to add/update seettings", "theme", theme, "err", err)
				}
				return theme, dir
			}
		}
		return "", ""
	}

	themeDir, ok := themes[settings.Theme.String]
	if !ok {
		for theme, dir := range themes {
			err := svc.AddOrUpdateSettings(ctx, sql.NullString{
				Valid:  true,
				String: theme,
			})
			if err != nil {
				slog.Error("failed to add/update seettings", "theme", theme, "err", err)
			}
			return theme, dir
		}
		return "", ""
	}

	return settings.Theme.String, themeDir
}
