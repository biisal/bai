package files

import (
	"log/slog"
	"os"
	"os/exec"
	"strings"
)

var (
	CurrentDirWithGitCache string
	CurrentDirCache        string
)

func CurrentDir() string {
	currentWd, err := os.Getwd()
	if err != nil {
		slog.Error("failed to get current dir", "error", err)
		return ""
	}
	CurrentDirCache = currentWd
	return currentWd
}

func gitBranch() string {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	branch := strings.TrimSpace(string(out))
	if branch == "" || branch == "HEAD" {
		return ""
	}
	return branch
}

func CurrentDirWithGit() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}

	// Replace home directory with ~
	if home, err := os.UserHomeDir(); err == nil && strings.HasPrefix(dir, home) {
		dir = "~" + strings.TrimPrefix(dir, home)
	}

	branch := gitBranch()
	if branch == "" {
		return dir
	}

	CurrentDirWithGitCache = dir + " (" + branch + ")"
	return CurrentDirWithGitCache
}

func init() {
	CurrentDirWithGitCache = CurrentDirWithGit()
	CurrentDirCache = CurrentDir()
}
