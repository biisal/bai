package git

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
)

type Git struct {
	Directory string `json:"directory"`
}

func New(directory string) *Git {
	return &Git{Directory: directory}
}

func (g *Git) command(ctx context.Context, args ...string) *exec.Cmd {
	full := append([]string{"--git-dir", g.Directory, "--work-tree", "."}, args...)
	slog.Debug("runGitCommand", "args", full, "directory", g.Directory)
	return exec.CommandContext(ctx, "git", full...)
}

func (g *Git) runGitCommand(args ...string) (string, error) {
	return g.RunGitCommandContext(context.Background(), args...)
}

func (g *Git) RunGitCommandContext(ctx context.Context, args ...string) (string, error) {
	cmd := g.command(ctx, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		slog.Error("runGitCommand", "error", err, "stderr", stderr.String())
		return stdout.String(), fmt.Errorf("git %s: %w: %s",
			strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return stdout.String(), nil
}

func (g *Git) Passthrough(ctx context.Context, args ...string) error {
	cmd := g.command(ctx, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (c *Git) CheckIfGitInitialized() (bool, error) {
	if _, err := c.runGitCommand("status"); err != nil {
		if exitErr, ok := errors.AsType[*exec.ExitError](err); ok {
			if exitErr.ExitCode() == 128 {
				return false, nil
			}
		}
		return false, err
	}
	return true, nil
}

func normalizePath(path string) (s string) {
	s = strings.TrimSpace(path)
	s = strings.TrimPrefix(s, "./")
	s = strings.TrimSuffix(s, "/")
	s = strings.TrimRight(s, "/")
	return
}

func (c *Git) InsertToGitIgnore(paths ...string) error {
	const gitignore = ".gitignore"
	content, err := os.ReadFile(gitignore)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		content = nil
	}
	exitsting := make(map[string]bool)
	for line := range strings.SplitSeq(string(content), "\n") {
		line = normalizePath(line)
		exitsting[line] = true
	}
	toAdd := make([]string, 0, len(paths))
	for _, path := range paths {
		path = normalizePath(path)
		if !exitsting[path] {
			toAdd = append(toAdd, path)
		}
	}

	var b bytes.Buffer
	b.Write(content)
	if len(content) > 0 && !bytes.HasSuffix(content, []byte("\n")) {
		b.WriteString("\n")
	}
	for _, path := range toAdd {
		b.WriteString(path)
		b.WriteString("\n")
	}
	if b.Len() > 0 {
		return os.WriteFile(gitignore, b.Bytes(), 0o644)
	}
	return nil
}

func (c *Git) CheckIfDirty() (bool, string, error) {
	args := []string{"status", "--porcelain"}
	output, err := c.runGitCommand(args...)
	if err != nil {
		return false, "", err
	}
	return len(output) > 0, output, nil
}
