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

func TestGitIgnoreFoldersGitSemantics(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []string
	}{
		{
			name:    "negation_reinclude_dir",
			content: "dist/\n!dist/\n",
			want:    nil, // git: !dist/ re-includes, dist is NOT ignored
		},
		{
			name:    "negation_reinclude_no_slash",
			content: "dist/\n!dist\n",
			want:    nil, // git: kept
		},
		{
			name:    "negation_before_exclude",
			content: "!dist/\ndist/\n",
			want:    []string{"dist"}, // git: last match wins -> ignored
		},
		{
			name:    "negation_under_excluded_dir_ignored",
			content: "node_modules/\n!node_modules/keep.txt\n",
			want:    []string{"node_modules"}, // git: cannot re-include under excluded dir
		},
		{
			name:    "escaped_hash",
			content: "\\#build/\n",
			want:    []string{"#build"}, // git: matches dir literally named "#build"
		},
		{
			name:    "escaped_bang",
			content: "\\!build/\n",
			want:    []string{"!build"}, // git: matches dir literally named "!build"
		},
		{
			name:    "escaped_trailing_space",
			content: "foo\\ \n",
			want:    []string{"foo "}, // git: quoted trailing space stays in pattern
		},
		{
			name:    "leading_space_is_literal",
			content: " spaced/\n",
			want:    []string{" spaced"}, // git: ignores " spaced" only, NOT "spaced"
		},
		{
			name:    "plain_names",
			content: "secret\nnode_modules/\n",
			want:    []string{"secret", "node_modules"},
		},
		{
			name:    "comment_and_blank",
			content: "# dist/\n\nbuild\n",
			want:    []string{"build"},
		},
		{
			name:    "inner_hash_is_literal",
			content: "foo#bar/\n",
			want:    []string{"foo#bar"},
		},
		{
			name:    "crlf_line_endings",
			content: "dist/\r\nbuild/\r\n",
			want:    []string{"dist", "build"},
		},
		{
			name:    "duplicate_lines",
			content: "dist/\ndist/\n",
			want:    []string{"dist"}, // duplicates are not folders
		},
		{
			name:    "bang_only",
			content: "!\n",
			want:    nil,
		},
		{
			name:    "slash_only",
			content: "////\n",
			want:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			p := filepath.Join(dir, ".gitignore")
			if err := os.WriteFile(p, []byte(tt.content), 0o644); err != nil {
				t.Fatal(err)
			}
			got, err := GitIgnoreFolders(p)
			if err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("content %q\n got: %q\nwant: %q", tt.content, got, tt.want)
			}
		})
	}
}

func TestGitIgnoreFoldersErrors(t *testing.T) {
	t.Run("empty_file", func(t *testing.T) {
		p := filepath.Join(t.TempDir(), ".gitignore")
		if err := os.WriteFile(p, nil, 0o644); err != nil {
			t.Fatal(err)
		}
		got, err := GitIgnoreFolders(p)
		if err != nil || got != nil {
			t.Errorf("empty file: got (%v, %v), want (nil, nil)", got, err)
		}
	})

	t.Run("path_is_a_directory", func(t *testing.T) {
		got, err := GitIgnoreFolders(t.TempDir())
		if err == nil {
			t.Errorf("directory path: got (%v, nil), want error", got)
		}
	})

	t.Run("unreadable_file", func(t *testing.T) {
		p := filepath.Join(t.TempDir(), ".gitignore")
		if err := os.WriteFile(p, []byte("dist/\n"), 0o000); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = os.Chmod(p, 0o644) })
		if os.Geteuid() == 0 {
			t.Skip("running as root, permissions are bypassed")
		}
		if _, err := GitIgnoreFolders(p); err == nil {
			t.Error("unreadable file: want error, got nil")
		}
	})
}
