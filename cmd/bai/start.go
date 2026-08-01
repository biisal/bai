package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/biisal/bai/internal/agent"
	"github.com/biisal/bai/internal/agent/providers"
	"github.com/biisal/bai/internal/config"
	"github.com/biisal/bai/internal/db"
	"github.com/biisal/bai/internal/tui"
)

func start(configPath string) error {
	config, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	providersMap := make(map[string]providers.Provider)

	for _, providerConfig := range config.Providers {
		provider, err := providers.NewFromConfig(providerConfig)
		if err != nil {
			return fmt.Errorf("failed to create provider: %w", err)
		}
		providersMap[providerConfig.ID] = provider
	}

	conn, err := db.Connect(config.DatabasePath)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	dbService := db.New(conn)
	gateway := agent.NewGateway(dbService, providersMap)

	p := tea.NewProgram(tui.InitModel(gateway))
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Oof: %v\n", err)
	}
	return nil
}
