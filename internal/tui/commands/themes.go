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
	themes, err := config.GetAllThemes()
	if err != nil {
		return nil
	}
	items := make([]list.Item, 0, len(themes))
	for _, tf := range themes {
		items = append(items, ThemeItem{Name: tf.Name, FilePath: tf.FilePath})
	}
	return items
}
