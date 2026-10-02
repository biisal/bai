package commands

import (
	"io/fs"
	"log/slog"
	"path/filepath"
	"strings"

	"charm.land/bubbles/v2/list"
	"github.com/biisal/bai/internal/files"
	"github.com/biisal/bai/internal/git"
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

	gitIgnored, err := git.GitIgnoreFolders(filepath.Join(currentDir, ".gitignore"))
	if err != nil {
		slog.Error("git ignore folders", "error", err)
	}

	ignored := make(map[string]struct{}, len(gitIgnored)+len(ignoreFolders))
	for _, n := range gitIgnored {
		ignored[n] = struct{}{}
	}
	for _, n := range ignoreFolders {
		ignored[n] = struct{}{}
	}

	prefix := currentDir + string(filepath.Separator)
	items := make([]list.Item, 0, 1024)

	err = filepath.WalkDir(currentDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == currentDir {
			return nil // skip root
		}
		if _, skip := ignored[d.Name()]; skip {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			return nil
		}

		items = append(items, FileItem{
			Name:     d.Name(),
			FilePath: strings.TrimPrefix(path, prefix),
		})
		return nil
	})
	if err != nil {
		slog.Error("error getting file items", "err", err)
		return nil
	}
	return items
}
