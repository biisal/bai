package commands

import (
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"charm.land/bubbles/v2/list"
	"github.com/biisal/bai/internal/files"
	"github.com/biisal/bai/internal/git"
)

var ignoreFolders = []string{
	// Dependencies
	"node_modules",
	"vendor",
	".venv",
	"venv",
	"env",
	".env",

	// Git / version control
	".git",
	".hg",
	".svn",

	// Python
	"__pycache__",
	".pytest_cache",
	".mypy_cache",
	".ruff_cache",

	// JavaScript / TypeScript
	".next",
	".nuxt",
	".turbo",
	".parcel-cache",
	".vite",
	"dist",
	"build",
	"out",

	// Go
	"bin",

	// Rust
	"target",

	// Java / JVM
	".gradle",
	"target",

	// OS
	".DS_Store",

	// Coverage / test output
	"coverage",
	".nyc_output",

	// Caches / temporary
	".cache",
	".tmp",
	"tmp",
	"temp",
}

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
	currentDir := files.CurrentDir()

	hardIgnored := make(map[string]struct{}, len(ignoreFolders))
	for _, n := range ignoreFolders {
		hardIgnored[n] = struct{}{}
	}
	for _, p := range extraPaths {
		hardIgnored[p] = struct{}{}
	}

	var matcher git.Matcher
	loadIgnore := func(dir, domain string) {
		content, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
		if err != nil {
			if !os.IsNotExist(err) {
				slog.Error("read .gitignore", "dir", dir, "error", err)
			}
			return
		}
		matcher.Add(string(content), domain)
	}
	loadIgnore(currentDir, "")

	items := make([]list.Item, 0, 1024)

	err := filepath.WalkDir(currentDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			// Unreadable entry: skip it and keep walking the rest.
			if d != nil && d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if path == currentDir {
			return nil // skip root
		}

		isDir := d.IsDir()
		if _, skip := hardIgnored[d.Name()]; skip {
			if isDir {
				return fs.SkipDir
			}
			return nil
		}

		rel, err := filepath.Rel(currentDir, path)
		if err != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)

		if matcher.Ignored(rel, isDir) {
			if isDir {
				return fs.SkipDir
			}
			return nil
		}
		if isDir {
			loadIgnore(path, rel) // nested .gitignore applies below this dir
			return nil
		}

		items = append(items, FileItem{
			Name:        d.Name(),
			RelFilePath: rel,
			AbsFilePath: path,
		})
		return nil
	})
	if err != nil {
		slog.Error("error getting file items", "err", err)
		return nil
	}
	return items
}
