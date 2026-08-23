package commands

import (
	"charm.land/bubbles/v2/list"
	"github.com/biisal/bai/internal/config"
)

type ModelItem struct {
	ProviderName string
	ModelID      string
}

func (m ModelItem) Title() string {
	return m.ProviderName
}

func (m ModelItem) Description() string {
	return m.ModelID
}

func (m ModelItem) FilterValue() string {
	return m.ModelID
}

func parseModels(providers []config.ProviderConfig) []list.Item {
	var items []list.Item
	for _, provider := range providers {
		for _, model := range provider.Models {
			items = append(items, ModelItem{
				ProviderName: provider.Name,
				ModelID:      model.ID,
			})
		}
	}
	return items
}
