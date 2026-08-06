package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"

	tea "charm.land/bubbletea/v2"
	"github.com/biisal/bai/internal/agent"
	"github.com/biisal/bai/internal/agent/providers"
	"github.com/biisal/bai/internal/config"
	"github.com/biisal/bai/internal/db"
	"github.com/biisal/bai/internal/logger"
	broker "github.com/biisal/bai/internal/pubsub"
	"github.com/biisal/bai/internal/tui"
)

func start(configPath string) error {
	config, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	file, err := logger.SetUpLogger(config.LogFilePath)
	if err != nil {
		return fmt.Errorf("failed to set up logger: %w", err)
	}

	defer func() {
		if err := file.Close(); err != nil {
			slog.Error(err.Error())
			return
		}
	}()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	providersMap := make(map[string]providers.Provider)
	b := broker.New()
	for _, providerConfig := range config.Providers {
		provider, err := providers.NewFromConfig(providerConfig, b)
		if err != nil {
			return fmt.Errorf("failed to create provider: %w", err)
		}
		providersMap[providerConfig.ID] = provider
	}

	conn, err := db.Connect(config.DatabasePath)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	if err := db.Migrate(ctx, conn); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}
	dbService := db.New(conn)

	activeProvider, activeModel, err := resolveProvider(ctx, dbService, config.Providers)
	if err != nil {
		return fmt.Errorf("failed to resolve provider: %w", err)
	}

	gateway := agent.NewGateway(dbService, providersMap, activeProvider, activeModel)
	p := tea.NewProgram(tui.InitModel(ctx, gateway, b))
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Oof: %v\n", err)
	}
	return nil
}
