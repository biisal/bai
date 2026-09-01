package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"

	tea "charm.land/bubbletea/v2"

	"github.com/biisal/bai/internal/agent"
	"github.com/biisal/bai/internal/config"
	"github.com/biisal/bai/internal/db"
	repo "github.com/biisal/bai/internal/db/sqlc"
	"github.com/biisal/bai/internal/logger"
	broker "github.com/biisal/bai/internal/pubsub"
	tui "github.com/biisal/bai/internal/tui/core"
	"github.com/biisal/bai/internal/tui/styles"
)

func startDB(ctx context.Context, path string) (*repo.Queries, error) {
	conn, err := db.Connect(path)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	if err := db.Migrate(ctx, conn); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}
	dbService := repo.New(conn)
	return dbService, nil
}

func SetTheme(ctx context.Context, dbService *repo.Queries, b broker.Service) {
	themes, err := config.GetAllThemes()
	if err != nil {
		// TODO: init model will call content.Rerender which will erase this messages
		b.Publish(ctx, broker.Message{
			Type: broker.EventSystemNoticeError,
			Text: fmt.Sprintf("faild to load themes, continueing with default theme! Error : %s", err),
		})
	}

	_, themeDir := resolveTheme(ctx, dbService, themes)
	theme, err := config.NewTheme(themeDir)
	if err != nil {
		b.Publish(ctx, broker.Message{
			Type: broker.EventSystemNoticeError,
			Text: fmt.Sprintf("faild to load themes, continueing with default theme! Error : %s", err),
		})
		theme = config.DefaultTheme()
	}
	styles.UpdateStylesUsingConfigTheme(theme)
}

func start(configPath string, dev bool) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	logLevel := slog.LevelInfo
	if dev {
		slog.Info("Starting in dev mode")
		logLevel = slog.LevelDebug
	}

	file, err := logger.SetUpLogger(cfg.LogFilePath, logLevel)
	if err != nil {
		return fmt.Errorf("failed to set up logger: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			slog.Error(err.Error())
		}
	}()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	dbService, err := startDB(ctx, cfg.DatabasePath)
	if err != nil {
		return err
	}

	b := broker.New()
	gateway, err := agent.NewGateway(ctx, dbService, b, cfg.Providers)
	if err != nil {
		return err
	}

	SetTheme(ctx, dbService, b)

	p := tea.NewProgram(tui.InitModel(ctx, gateway, b, cfg.Providers))
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Oof: %v\n", err)
	}
	return nil
}
