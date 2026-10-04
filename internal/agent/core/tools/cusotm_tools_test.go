package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"charm.land/fantasy"
)

func TestCustomPluginParseAndRun(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "echo.sh")
	if err := os.WriteFile(script, []byte("#!/usr/bin/env bash\nprintf 'echo.sh got: %s\\n' \"$0 $*\"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := filepath.Join(dir, "plugins.json")
	content := `[{"file_name":"echo.sh","description":"Echo test plugin.","input_schema":{"text":"text to echo"}}]`
	if err := os.WriteFile(manifest, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	parsed, err := ParseTooolsByPath(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if len(parsed) != 1 {
		t.Fatalf("got %d tools, want 1", len(parsed))
	}
	if _, err := os.Stat(parsed[0].ExecutablePath); err != nil {
		t.Fatalf("plugin executable not found at %q: %v", parsed[0].ExecutablePath, err)
	}

	tool := MakeToolFromCustomTools(parsed[0])
	if got := tool.Info().Name; got != "echo" {
		t.Errorf("tool name = %q, want %q", got, "echo")
	}

	resp, err := tool.Run(context.Background(), fantasy.ToolCall{
		ID:    "test",
		Name:  "echo",
		Input: `{"text":"hello-123"}`,
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.IsError {
		t.Fatalf("tool returned error: %s", resp.Content)
	}
	if !strings.Contains(resp.Content, "text=hello-123") {
		t.Errorf("response %q does not contain the passed arg", resp.Content)
	}
}
