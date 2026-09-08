package commands

import (
	"io/fs"
	"log/slog"
	"path/filepath"
	"slices"

	"charm.land/bubbles/v2/list"
	"github.com/biisal/bai/internal/files"
)

var ignoreFolders = []string{"node_modules", ".venv", ".git"}

type FileItem struct {
	Name     string
	FilePath string
}

func (t FileItem) Title() string {
	return t.Name
}

func (t FileItem) Description() string {
	return t.FilePath
}

func (t FileItem) FilterValue() string { return t.Name }

func FileItems() []list.Item {
	currentDir := files.CurrentDir()

	var items []list.Item
	if err := filepath.WalkDir(currentDir, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			if slices.Contains(ignoreFolders, d.Name()) {
				return fs.SkipDir
			}
			return nil
		}
		relative, err := filepath.Rel(currentDir, path)
		if err != nil {
			return nil
		}

		items = append(items, FileItem{
			Name:     d.Name(),
			FilePath: relative,
		})
		return nil
	}); err != nil {
		slog.Error("error getting file items", "err", err)
		return nil
	}
	return items
}
