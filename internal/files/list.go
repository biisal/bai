package files

import (
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

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

type Entry struct {
	Name    string
	RelPath string
	AbsPath string
}

// List walks the current directory, skipping hard-ignored folders and
// anything matched by .gitignore, and returns every remaining file.
func List(extraPaths ...string) []Entry {
	currentDir := CurrentDir()

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

	items := make([]Entry, 0, 1024)

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

		items = append(items, Entry{Name: d.Name(), RelPath: rel, AbsPath: path})
		return nil
	})
	if err != nil {
		slog.Error("error getting file items", "err", err)
		return nil
	}
	return items
}
