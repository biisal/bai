package tools_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"charm.land/fantasy"
	"github.com/biisal/bai/internal/agent/core/tools"
	broker "github.com/biisal/bai/internal/pubsub"
)

func writeManifest(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "tools.json")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func toolNames(ts []fantasy.AgentTool) []string {
	names := make([]string, len(ts))
	for i, t := range ts {
		names[i] = t.Info().Name
	}
	return names
}

func findTool(t *testing.T, ts []fantasy.AgentTool, name string) fantasy.AgentTool {
	t.Helper()
	for _, tool := range ts {
		if tool.Info().Name == name {
			return tool
		}
	}
	t.Fatalf("tool %q not found in %v", name, toolNames(ts))
	return nil
}

func TestNewTools_BuiltinsAndCustom(t *testing.T) {
	manifest := writeManifest(t, `[{"file_name":"echo.sh","description":"Echo test plugin.","input_schema":{"text":"text to echo"}}]`)

	got := tools.NewTools(broker.New(), manifest)
	want := []string{
		tools.ReadFileName,
		tools.WriteFileName,
		tools.EditFileName,
		tools.BashName,
		"echo",
	}
	names := toolNames(got)
	if len(names) != len(want) {
		t.Fatalf("got %d tools (%v), want %d", len(names), names, len(want))
	}
	for i, name := range want {
		if names[i] != name {
			t.Errorf("tool[%d] = %q, want %q", i, names[i], name)
		}
	}
	for _, tool := range got {
		if tool.Info().Description == "" {
			t.Errorf("tool %q has empty description", tool.Info().Name)
		}
	}
}

func TestNewTools_BadManifestKeepsBuiltins(t *testing.T) {
	badJSON := writeManifest(t, `{not json`)
	builtins := []string{
		tools.ReadFileName,
		tools.WriteFileName,
		tools.EditFileName,
		tools.BashName,
	}
	for _, path := range []string{badJSON, filepath.Join(t.TempDir(), "missing.json"), ""} {
		got := tools.NewTools(broker.New(), path)
		names := toolNames(got)
		if len(names) != len(builtins) {
			t.Fatalf("NewTools(%q) = %v, want builtins %v", path, names, builtins)
		}
		for i, name := range builtins {
			if names[i] != name {
				t.Errorf("NewTools(%q)[%d] = %q, want %q", path, i, names[i], name)
			}
		}
	}
}

func TestNewTools_ReadFile(t *testing.T) {
	file := filepath.Join(t.TempDir(), "hello.txt")
	if err := os.WriteFile(file, []byte("line one\nline two\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	ts := tools.NewTools(broker.New(), writeManifest(t, `[]`))
	read := findTool(t, ts, tools.ReadFileName)

	resp, err := read.Run(context.Background(), fantasy.ToolCall{
		ID:    "call-1",
		Name:  tools.ReadFileName,
		Input: `{"path":"` + file + `"}`,
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.IsError {
		t.Fatalf("read_file error: %s", resp.Content)
	}
	if resp.Content != "line one\nline two\n" {
		t.Errorf("content = %q, want file contents", resp.Content)
	}

	resp, err = read.Run(context.Background(), fantasy.ToolCall{
		ID:    "call-2",
		Name:  tools.ReadFileName,
		Input: `{"path":"` + filepath.Join(t.TempDir(), "nope.txt") + `"}`,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !resp.IsError {
		t.Error("expected error response for missing file")
	}
}

func TestReadFileInternalSkillFillsToolPaths(t *testing.T) {
	manifest := writeManifest(t, `[]`)
	abs, err := filepath.Abs(manifest)
	if err != nil {
		t.Fatal(err)
	}

	ts := tools.NewTools(broker.New(), manifest)
	read := findTool(t, ts, tools.ReadFileName)

	resp, err := read.Run(context.Background(), fantasy.ToolCall{
		ID:    "call-1",
		Name:  tools.ReadFileName,
		Input: `{"path":"internal:skills:tools_maker"}`,
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.IsError {
		t.Fatalf("read_file error: %s", resp.Content)
	}
	if strings.Contains(resp.Content, "{{") {
		t.Errorf("unreplaced placeholder left in content: %.200s", resp.Content)
	}
	if !strings.Contains(resp.Content, abs) {
		t.Errorf("content missing manifest path %q", abs)
	}
	if !strings.Contains(resp.Content, filepath.Dir(abs)) {
		t.Errorf("content missing tools dir %q", filepath.Dir(abs))
	}
}

func TestNewTools_Bash(t *testing.T) {
	ts := tools.NewTools(broker.New(), writeManifest(t, `[]`))
	bash := findTool(t, ts, tools.BashName)

	resp, err := bash.Run(context.Background(), fantasy.ToolCall{
		ID:    "call-1",
		Name:  tools.BashName,
		Input: `{"command":"echo hi","purpose":"say hi"}`,
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.IsError {
		t.Fatalf("bash error: %s", resp.Content)
	}
	if resp.Content != "hi\n" {
		t.Errorf("content = %q, want %q", resp.Content, "hi\n")
	}
}
