package git

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestGitIgnoreFolders(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, ".gitignore")
	content := "# comment\n\nnode_modules/\n/dist/\nbuild\n*.log\n!keep/\ndocs/private/\n.env\n"
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := GitIgnoreFolders(p)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"node_modules", "dist", "build", ".env"}
	if !slices.Equal(got, want) {
		t.Errorf("GitIgnoreFolders() = %v, want %v", got, want)
	}

	got, err = GitIgnoreFolders(filepath.Join(dir, "missing", ".gitignore"))
	if err != nil || got != nil {
		t.Errorf("missing .gitignore: got (%v, %v), want (nil, nil)", got, err)
	}
}
