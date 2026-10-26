package web

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"charm.land/fantasy"
	"github.com/biisal/bai/internal/agent"
	"github.com/biisal/bai/internal/config"
	"github.com/biisal/bai/internal/db"
	repo "github.com/biisal/bai/internal/db/sqlc"
	"github.com/biisal/bai/internal/git"
	"github.com/biisal/bai/internal/notifier"
	broker "github.com/biisal/bai/internal/pubsub"
)

func TestSearchFilesHonorsIgnoresAndCaps(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	mustWrite(t, "main.go", "package main")
	mustWrite(t, "node_modules/dep.js", "x")
	mustWrite(t, "ignored.txt", "x")
	mustWrite(t, ".gitignore", "ignored.txt\n")
	for i := range 25 {
		mustWrite(t, fmt.Sprintf("f%02d.go", i), "package main")
	}

	results := searchFiles("")
	if len(results) != fileResultLimit {
		t.Fatalf("got %d results, want %d", len(results), fileResultLimit)
	}
	for _, result := range results {
		if strings.Contains(result.Path, "node_modules") || result.Path == "ignored.txt" {
			t.Fatalf("ignored path returned: %s", result.Path)
		}
		if filepath.IsAbs(result.Path) || result.IsDir {
			t.Fatalf("bad result: %+v", result)
		}
	}
}

func mustWrite(t *testing.T, name, content string) {
	t.Helper()
	path := filepath.Join(".", name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestFileQueryRejectsTraversalAndAbsolute(t *testing.T) {
	for _, q := range []string{"../etc", "a/../../b", "/etc/passwd", `C:\windows`} {
		if validFileQuery(q) {
			t.Fatalf("query %q should be rejected", q)
		}
	}
	if !validFileQuery("internal/web") {
		t.Fatal("relative query should be allowed")
	}

	req := httptest.NewRequest(http.MethodGet, "/api/files?q=../secret", nil)
	rec := httptest.NewRecorder()
	(&Server{}).handleFiles(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("got %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestSettingsProviderJSONHasNoSecrets(t *testing.T) {
	data, err := json.Marshal(settingsProviders([]config.ProviderConfig{{
		Name:    "private",
		APIKey:  "sk-do-not-leak",
		BaseURL: "https://private.example/v1",
		Models:  []config.ModelConfig{{ID: "m", Context: 10, MaxOutput: 2}},
	}}))
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, bad := range []string{"sk-do-not-leak", "private.example", "api_key", "apiKey", "base_url", "baseURL"} {
		if strings.Contains(text, bad) {
			t.Fatalf("settings JSON leaked %q: %s", bad, text)
		}
	}
	if !strings.Contains(text, `"configured":true`) || !strings.Contains(text, `"context":10`) {
		t.Fatalf("settings JSON missing safe fields: %s", text)
	}
}

func TestDiffUsesFixedArgsAndCaps(t *testing.T) {
	want := []string{"show", "--format=", "--no-color", "--no-ext-diff", "--no-textconv", "--no-renames", "--unified=3", "HEAD", "--"}
	if strings.Join(diffArgs, " ") != strings.Join(want, " ") {
		t.Fatalf("diff args = %v", diffArgs)
	}
	got, truncated := capDiff(strings.Repeat("a", maxDiffBytes+1))
	if !truncated || len(got) != maxDiffBytes {
		t.Fatalf("capDiff() len=%d truncated=%v", len(got), truncated)
	}
}

func TestDiffShowsLastCommit(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("GIT_AUTHOR_NAME", "bai")
	t.Setenv("GIT_AUTHOR_EMAIL", "bai@example.com")
	t.Setenv("GIT_COMMITTER_NAME", "bai")
	t.Setenv("GIT_COMMITTER_EMAIL", "bai@example.com")
	repo := git.New(filepath.Join(t.TempDir(), "gitdir"))
	ctx := context.Background()
	if _, err := repo.RunGitCommandContext(ctx, "init"); err != nil {
		t.Fatal(err)
	}
	mustWrite(t, "hello.txt", "old\n")
	if _, err := repo.RunGitCommandContext(ctx, "add", "."); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.RunGitCommandContext(ctx, "commit", "-m", "before"); err != nil {
		t.Fatal(err)
	}
	mustWrite(t, "hello.txt", "new\n")
	if _, err := repo.RunGitCommandContext(ctx, "add", "."); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.RunGitCommandContext(ctx, "commit", "-m", "last run"); err != nil {
		t.Fatal(err)
	}
	diff, err := repo.RunGitCommandContext(ctx, diffArgs...)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(diff, "-old") || !strings.Contains(diff, "+new") {
		t.Fatalf("last commit diff missing change: %s", diff)
	}
}

func testGateway(t *testing.T) (*agent.Gateway, repo.Querier, *config.Config) {
	t.Helper()
	conn, err := db.Connect(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	if err := db.Migrate(context.Background(), conn); err != nil {
		t.Fatal(err)
	}
	q := repo.New(conn)
	cfg := &config.Config{
		GitDirName:  filepath.Join(t.TempDir(), ".bai_git"),
		PluginsPath: filepath.Join(t.TempDir(), "tools.json"),
		Providers: []config.ProviderConfig{{
			Name:    "test",
			Format:  config.FormatOpenAI,
			BaseURL: "http://127.0.0.1:1",
			APIKey:  "sk-test",
			Models:  []config.ModelConfig{{ID: "model", Context: 10, MaxOutput: 10}},
		}},
	}
	gateway, err := agent.NewGateway(
		context.Background(), q, broker.New(), cfg, notifier.New(nil), nil, git.New(cfg.GitDirName),
	)
	if err != nil {
		t.Fatal(err)
	}
	return gateway, q, cfg
}

func TestConversationMutations(t *testing.T) {
	gateway, q, cfg := testGateway(t)
	ctx := context.Background()
	conversation, err := gateway.AddNewConversation(ctx, "before")
	if err != nil {
		t.Fatal(err)
	}
	if err := gateway.SetActiveConversation(ctx, conversation.ID, &conversation); err != nil {
		t.Fatal(err)
	}
	if err := gateway.AddMessageToDB(ctx, fantasy.Message{
		Role:    fantasy.MessageRoleUser,
		Content: []fantasy.MessagePart{fantasy.TextPart{Text: "hello"}},
	}); err != nil {
		t.Fatal(err)
	}

	registry := NewRegistry(ctx)
	server := &Server{gateway: gateway, config: cfg, registry: registry}

	run, _, err := registry.Start(conversation.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got := mutationStatus(server.handleDeleteConversation, http.MethodDelete, conversation.ID, ""); got != http.StatusConflict {
		t.Fatalf("delete while running = %d", got)
	}
	if got := mutationStatus(server.handleRenameConversation, http.MethodPatch, conversation.ID, `{"title":"after"}`); got != http.StatusConflict {
		t.Fatalf("rename while running = %d", got)
	}
	run.finish()

	if got := mutationStatus(server.handleRenameConversation, http.MethodPatch, conversation.ID, `{"title":"after"}`); got != http.StatusOK {
		t.Fatalf("rename = %d", got)
	}
	renamed, err := q.GetConversation(ctx, conversation.ID)
	if err != nil || renamed.Title != "after" {
		t.Fatalf("renamed conversation = %+v, err=%v", renamed, err)
	}
	if got := mutationStatus(server.handleDeleteConversation, http.MethodDelete, conversation.ID, ""); got != http.StatusNoContent {
		t.Fatalf("delete = %d", got)
	}
	if _, err := q.GetConversation(ctx, conversation.ID); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("conversation still exists: %v", err)
	}
	messages, err := q.GetMessagesByConversation(ctx, conversation.ID)
	if err != nil || len(messages) != 0 {
		t.Fatalf("messages remain: %v, %v", messages, err)
	}

	if got := mutationStatus(server.handleDeleteConversation, http.MethodDelete, 9999, ""); got != http.StatusNotFound {
		t.Fatalf("delete missing = %d", got)
	}
	if got := mutationStatus(server.handleRenameConversation, http.MethodPatch, 9999, `{"title":"x"}`); got != http.StatusNotFound {
		t.Fatalf("rename missing = %d", got)
	}
}

func mutationStatus(handler http.HandlerFunc, method string, id int64, body string) int {
	req := httptest.NewRequest(method, "/api/conversations/1", strings.NewReader(body))
	req.SetPathValue("id", fmt.Sprint(id))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	handler(rec, req)
	return rec.Code
}
