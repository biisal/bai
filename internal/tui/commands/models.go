package commands

import (
	"log/slog"

	"charm.land/bubbles/v2/list"
	"github.com/biisal/bai/internal/config"
)

type ModelItem struct {
	Name       string
	ModelID    string
	ProviderID string
}

func (m ModelItem) Title() string {
	return m.Name
}

func (m ModelItem) Description() string {
	return m.ModelID
}

func (m ModelItem) FilterValue() string {
	return m.ModelID
}

func parseModels(providers []config.ProviderConfig) []list.Item {
	slog.Info("persing models", "providers count", len(providers))
	var items []list.Item
	for _, provider := range providers {
		for _, model := range provider.Models {
			items = append(items, ModelItem{
				Name:    provider.Name,
				ModelID: model.ID,
			})
		}
	}
	return items
}
