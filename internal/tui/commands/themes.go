package commands

import (
	"charm.land/bubbles/v2/list"
	"github.com/biisal/bai/internal/config"
)

type ThemeItem struct {
	Name     string
	FilePath string
}

func (t ThemeItem) Title() string {
	return t.Name
}

func (t ThemeItem) Description() string {
	return t.FilePath
}

func (t ThemeItem) FilterValue() string { return t.Name }

func ThemeFiles() []list.Item {
	names, err := config.GetAllThemes()
	if err != nil {
		return nil
	}
	var items []list.Item
	for name, path := range names {
		items = append(items, ThemeItem{Name: name, FilePath: path})
	}
	return items
}
