package tools

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

func TestExecuteBash_BasicOutput(t *testing.T) {
	content, err := executeBash(context.Background(), "echo hello", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(content) != "hello" {
		t.Fatalf("got %q, want %q", content, "hello")
	}
}

func TestExecuteBash_NoOutput(t *testing.T) {
	content, err := executeBash(context.Background(), "true", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if content != "(no output)" {
		t.Fatalf("got %q, want %q", content, "(no output)")
	}
}

func TestExecuteBash_ExitCode(t *testing.T) {
	_, err := executeBash(context.Background(), "exit 42", nil)
	if err == nil {
		t.Fatal("expected error for non-zero exit")
	}
	if !strings.Contains(err.Error(), "Command exited with code 42") {
		t.Fatalf("got %q, expected exit code 42 message", err.Error())
	}
}

func TestExecuteBash_Timeout(t *testing.T) {
	timeout := 1
	_, err := executeBash(context.Background(), "sleep 10", &timeout)
	if err == nil {
		t.Fatal("expected error for timeout")
	}
	if !strings.Contains(err.Error(), "Command timed out after 1 seconds") {
		t.Fatalf("got %q, expected timeout message", err.Error())
	}
}

func TestExecuteBash_StderrAndStdout(t *testing.T) {
	content, err := executeBash(context.Background(), "echo out; echo err >&2", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(content, "out") || !strings.Contains(content, "err") {
		t.Fatalf("expected both stdout and stderr, got %q", content)
	}
}

func TestExecuteBash_TruncationSavesFullOutput(t *testing.T) {
	content, err := executeBash(context.Background(), "seq 1 2500", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(content, "Output truncated") {
		t.Fatalf("expected truncation notice, got first 200 bytes: %.200s", content)
	}
	idx := strings.Index(content, "Full output: ")
	if idx < 0 {
		t.Fatal("expected 'Full output:' in truncation notice")
	}
	lineEnd := strings.IndexByte(content[idx:], '\n')
	path := strings.TrimRight(strings.TrimSpace(content[idx+len("Full output: "):idx+lineEnd]), "]")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("cannot read full output file %q: %v", path, err)
	}
	full := string(b)
	if !strings.Contains(full, "1\n") || !strings.Contains(full, "2500\n") {
		t.Fatalf("full output missing expected lines (first 80 bytes: %.80s)", full)
	}
}

func TestExecuteBash_BytesTruncation(t *testing.T) {
	// Use dd to generate ~300KB of output without a huge command string
	cmd := "dd if=/dev/zero bs=1024 count=300 2>/dev/null | tr '\\0' 'a'"
	content, err := executeBash(context.Background(), cmd, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(content, "Output truncated") {
		t.Fatalf("expected truncation notice, got first 100 bytes: %.100s", content)
	}
	if !strings.Contains(content, "Full output:") {
		t.Fatal("expected 'Full output:' in byte-level truncation notice")
	}
}

func TestExecuteBash_ContextCancel(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	_, err := executeBash(ctx, "sleep 10", nil)
	if err == nil {
		t.Fatal("expected error for context cancel")
	}
	if !strings.Contains(err.Error(), "aborted") && !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("got %q, expected abort or timeout message", err.Error())
	}
}
