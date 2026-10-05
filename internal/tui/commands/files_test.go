package commands

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"charm.land/bubbles/v2/list"
)

func buildTree(t *testing.T, root string, files []string) {
	t.Helper()
	for _, f := range files {
		p := filepath.Join(root, f)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func writeIgnore(t *testing.T, root, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(root, ".gitignore"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func itemPaths(t *testing.T, items []list.Item) []string {
	t.Helper()
	got := make([]string, 0, len(items))
	for _, it := range items {
		fi, ok := it.(FileItem)
		if !ok {
			t.Fatalf("unexpected item type %T", it)
		}
		got = append(got, fi.RelFilePath)
	}
	slices.Sort(got)
	return got
}

func runFileItems(t *testing.T, root string) []string {
	t.Helper()
	t.Chdir(root)
	return itemPaths(t, FileItems())
}

// wants were verified against real git (git check-ignore): each list is what
// git would keep visible in the file finder.
func TestFileItemsGitIgnore(t *testing.T) {
	t.Run("basic_skip", func(t *testing.T) {
		root := t.TempDir()
		writeIgnore(t, root, "dist/\nsecret.txt\n")
		buildTree(
			t, root,
			[]string{"keep.txt", "secret.txt", "src/main.go", "dist/deep.txt", ".git/HEAD"},
		)
		want := []string{".gitignore", "keep.txt", "src/main.go"}
		if got := runFileItems(t, root); !slices.Equal(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("no_gitignore_uses_hardcoded_folders", func(t *testing.T) {
		root := t.TempDir()
		buildTree(
			t, root,
			[]string{"keep.txt", "node_modules/pkg/index.js", ".venv/lib.py", ".git/HEAD"},
		)
		want := []string{"keep.txt"}
		if got := runFileItems(t, root); !slices.Equal(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("dir_only_pattern_keeps_same_named_file", func(t *testing.T) {
		// git: "outdir/" ignores the dir only; the FILE tools/outdir stays.
		// Names avoid ignoreFolders, which is checked before the matcher.
		root := t.TempDir()
		writeIgnore(t, root, "outdir/\n")
		buildTree(
			t, root,
			[]string{"keep.txt", "outdir/out.txt", "tools/outdir"},
		)
		want := []string{".gitignore", "keep.txt", "tools/outdir"}
		if got := runFileItems(t, root); !slices.Equal(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("anchored_pattern_keeps_nested_dir", func(t *testing.T) {
		// git: "/gen/" ignores root gen only; lib/gen stays.
		root := t.TempDir()
		writeIgnore(t, root, "/gen/\n")
		buildTree(
			t, root,
			[]string{"keep.txt", "gen/a.txt", "lib/gen/b.txt"},
		)
		want := []string{".gitignore", "keep.txt", "lib/gen/b.txt"}
		if got := runFileItems(t, root); !slices.Equal(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("negation_reinclude", func(t *testing.T) {
		// git: "!cfg/" re-includes; cfg/keep.txt stays visible.
		root := t.TempDir()
		writeIgnore(t, root, "cfg/\n!cfg/\n")
		buildTree(
			t, root,
			[]string{"top.txt", "cfg/keep.txt"},
		)
		want := []string{".gitignore", "cfg/keep.txt", "top.txt"}
		if got := runFileItems(t, root); !slices.Equal(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("glob_pattern_hides_file", func(t *testing.T) {
		// git: "*.log" ignores debug.log.
		root := t.TempDir()
		writeIgnore(t, root, "*.log\n")
		buildTree(
			t, root,
			[]string{"debug.log", "src/main.go"},
		)
		want := []string{".gitignore", "src/main.go"}
		if got := runFileItems(t, root); !slices.Equal(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("path_pattern_hides_dir", func(t *testing.T) {
		// git: "docs/private/" ignores docs/private.
		root := t.TempDir()
		writeIgnore(t, root, "docs/private/\n")
		buildTree(
			t, root,
			[]string{"docs/private/key.txt", "docs/public/readme.md"},
		)
		want := []string{".gitignore", "docs/public/readme.md"}
		if got := runFileItems(t, root); !slices.Equal(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("nested_gitignore_hides_dir", func(t *testing.T) {
		// git reads sub/.gitignore too; "sub/cache" stays ignored.
		root := t.TempDir()
		buildTree(
			t, root,
			[]string{"sub/.gitignore", "sub/cache/x.txt", "keep.txt"},
		)
		writeIgnore(t, filepath.Join(root, "sub"), "cache/\n")
		want := []string{"keep.txt", "sub/.gitignore"}
		if got := runFileItems(t, root); !slices.Equal(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("unreadable_dir_keeps_rest", func(t *testing.T) {
		if os.Geteuid() == 0 {
			t.Skip("running as root, permissions are bypassed")
		}
		root := t.TempDir()
		buildTree(
			t, root,
			[]string{"ok.txt", "noperm/secret.txt"},
		)
		bad := filepath.Join(root, "noperm")
		if err := os.Chmod(bad, 0o000); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = os.Chmod(bad, 0o755) })

		want := []string{"ok.txt"}
		if got := runFileItems(t, root); !slices.Equal(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})
}
