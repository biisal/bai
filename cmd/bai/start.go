package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"

	tea "charm.land/bubbletea/v2"
	fantasy "charm.land/fantasy"

	"github.com/biisal/bai/internal/agent"
	"github.com/biisal/bai/internal/config"
	"github.com/biisal/bai/internal/db"
	repo "github.com/biisal/bai/internal/db/sqlc"
	"github.com/biisal/bai/internal/logger"
	broker "github.com/biisal/bai/internal/pubsub"
	tui "github.com/biisal/bai/internal/tui/core"
)

func start(configPath string, dev bool) error {
	config, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	logLevel := slog.LevelInfo
	if dev {
		slog.Info("Starting in dev mode")
		logLevel = slog.LevelDebug
	}

	file, err := logger.SetUpLogger(config.LogFilePath, logLevel)
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

	providers := make(map[string]fantasy.Provider)
	for _, cfg := range config.Providers {
		p, err := buildProvider(cfg)
		if err != nil {
			return fmt.Errorf("failed to create provider: %w", err)
		}
		providers[cfg.Name] = p
	}

	conn, err := db.Connect(config.DatabasePath)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	if err := db.Migrate(ctx, conn); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}
	dbService := repo.New(conn)

	activeProvider, activeModel, err := resolveProvider(ctx, dbService, config.Providers)
	if err != nil {
		return fmt.Errorf("failed to resolve provider: %w", err)
	}

	b := broker.New()
	gateway := agent.NewGateway(dbService, b, providers, activeProvider, activeModel)

	p := tea.NewProgram(tui.InitModel(ctx, gateway, b, config.Providers))
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Oof: %v\n", err)
	}
	return nil
}
