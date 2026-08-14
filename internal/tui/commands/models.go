package commands

import (
	"charm.land/bubbles/v2/list"
	"github.com/biisal/bai/internal/config"
)

type ModelList struct {
	ModelID    string
	ProviderID string
}

func parseModels(providers []config.ProviderConfig) []list.Item {
	var items []list.Item
	for _, provider := range providers {
		for _, model := range provider.Models {
			items = append(items, ListItem[ModelList]{Fields: ModelList{
				ModelID:    model.ID,
				ProviderID: provider.ID,
			}})
		}
	}
	return items
}
