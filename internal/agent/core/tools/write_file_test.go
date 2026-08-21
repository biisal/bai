package tools

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	broker "github.com/biisal/bai/internal/pubsub"
)

func TestWriteFile_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "hello.txt")
	if err := WriteFile(context.Background(), path, "hello world"); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile() error = %v", err)
	}
	if string(b) != "hello world" {
		t.Fatalf("content = %q, want %q", string(b), "hello world")
	}
}

func TestWriteFile_CreatesParentDirs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a", "b", "c", "f.txt")
	if err := WriteFile(context.Background(), path, "nested"); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile() error = %v", err)
	}
	if string(b) != "nested" {
		t.Fatalf("content = %q, want %q", string(b), "nested")
	}
}

func TestWriteFile_OverwritesExisting(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(path, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := WriteFile(context.Background(), path, "new content"); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile() error = %v", err)
	}
	if string(b) != "new content" {
		t.Fatalf("content = %q, want %q", string(b), "new content")
	}
}

func TestWriteFile_EmptyContent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.txt")
	if err := WriteFile(context.Background(), path, ""); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile() error = %v", err)
	}
	if len(b) != 0 {
		t.Fatalf("content length = %d, want 0", len(b))
	}
}

func TestWriteFile_EmptyPath(t *testing.T) {
	if err := WriteFile(context.Background(), "", "x"); err == nil {
		t.Fatal("WriteFile(\"\", ...) expected error")
	}
}

func TestExecuteWriteFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.txt")
	argsJSON := `{"path":` + `"` + path + `"` + `,"content":"test data"}`
	call := Call{ID: "1", Name: WriteFileTool, Args: []byte(argsJSON)}
	content, isError := Execute(t.Context(), call, broker.New())
	if isError {
		t.Fatalf("Execute() unexpected error: %s", content)
	}
	want := "Successfully wrote 9 bytes to " + path
	if content != want {
		t.Fatalf("Execute() = %q, want %q", content, want)
	}
}

func TestExecuteWriteFile_MissingPath(t *testing.T) {
	call := Call{ID: "1", Name: WriteFileTool, Args: []byte(`{"content":"hi"}`)}
	_, isError := Execute(t.Context(), call, broker.New())
	if !isError {
		t.Fatal("expected error for missing path")
	}
}
