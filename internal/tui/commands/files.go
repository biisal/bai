package commands

import (
	"charm.land/bubbles/v2/list"
	"github.com/biisal/bai/internal/files"
)

type FileItem struct {
	Name        string
	RelFilePath string
	AbsFilePath string
}

func (t FileItem) Title() string {
	return t.Name
}

func (t FileItem) Description() string {
	return t.RelFilePath
}

func (t FileItem) FilterValue() string { return t.Name }

func FileItems(extraPaths ...string) []list.Item {
	entries := files.List(extraPaths...)
	items := make([]list.Item, 0, len(entries))
	for _, e := range entries {
		items = append(items, FileItem{Name: e.Name, RelFilePath: e.RelPath, AbsFilePath: e.AbsPath})
	}
	return items
}
