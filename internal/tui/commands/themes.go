package commands

import (
	"os"
	"path/filepath"

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

func checkIfConfigFile(entry os.DirEntry) bool {
	if entry.IsDir() {
		return false
	}

	if filepath.Ext(entry.Name()) != ".json" {
		return false
	}

	return true
}

func ThemeFiles() []list.Item {
	entries, err := os.ReadDir(config.ThemeConfigDir())
	if err != nil {
		return nil
	}
	var items []list.Item
	for _, entry := range entries {
		if !checkIfConfigFile(entry) {
			continue
		}
		items = append(items, ThemeItem{Name: entry.Name(), FilePath: filepath.Join(config.ThemeConfigDir(), entry.Name())})
	}
	return items
}
