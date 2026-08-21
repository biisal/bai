package tools

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

func executeBash(ctx context.Context, command string, timeoutSecs *int) (content string, isError bool) {
	if timeoutSecs != nil && *timeoutSecs > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(*timeoutSecs)*time.Second)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, "bash", "-c", command)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	err := runCommand(ctx, cmd)
	output := truncateOutput(buf.String())

	switch {
	case errors.Is(ctx.Err(), context.DeadlineExceeded):
		secs := 0
		if timeoutSecs != nil {
			secs = *timeoutSecs
		}
		return appendStatus(output, fmt.Sprintf("Command timed out after %d seconds", secs)), true
	case errors.Is(ctx.Err(), context.Canceled):
		return appendStatus(output, "Command aborted"), true
	}

	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return appendStatus(output, fmt.Sprintf("Command exited with code %d", exitErr.ExitCode())), true
		}
		return fmt.Sprintf("error executing command: %v", err), true
	}

	if strings.TrimSpace(output) == "" {
		return "(no output)", false
	}
	return output, false
}

func runCommand(ctx context.Context, cmd *exec.Cmd) error {
	if err := cmd.Start(); err != nil {
		return err
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		return <-done
	}
}

func truncateOutput(out string) string {
	if len(out) <= DefaultMaxBytes && strings.Count(out, "\n") <= DefaultMaxLines {
		return out
	}

	s := out
	for len(s) > DefaultMaxBytes || strings.Count(s, "\n") > DefaultMaxLines {
		i := strings.IndexByte(s, '\n')
		if i < 0 || i == len(s)-1 {
			s = s[len(s)-DefaultMaxBytes:]
			break
		}
		s = s[i+1:]
	}
	if s == out {
		return out
	}

	path, err := saveFullOutput(out)
	fullOutput := ""
	if err == nil {
		fullOutput = fmt.Sprintf(" Full output: %s", path)
	}

	if !strings.Contains(s, "\n") {
		lastLine := strings.TrimSuffix(out, "\n")
		if i := strings.LastIndex(lastLine, "\n"); i >= 0 {
			lastLine = lastLine[i+1:]
		}
		return fmt.Sprintf("[Output truncated — showing last %s of a %s line.%s]\n%s",
			formatSize(len(s)), formatSize(len(lastLine)), fullOutput, s)
	}

	return fmt.Sprintf("[Output truncated — showing last %d of %d lines (%s limit).%s]\n%s",
		strings.Count(s, "\n"), strings.Count(out, "\n"), formatSize(DefaultMaxBytes), fullOutput, s)
}

func saveFullOutput(out string) (string, error) {
	f, err := os.CreateTemp("", "bai-bash-*")
	if err != nil {
		return "", err
	}
	name := f.Name()
	if _, err := f.WriteString(out); err != nil {
		_ = f.Close()
		_ = os.Remove(name)
		return "", err
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(name)
		return "", err
	}
	return name, nil
}

func formatSize(b int) string {
	const (
		kb = 1024
		mb = 1024 * kb
	)
	switch {
	case b >= mb:
		return fmt.Sprintf("%.1fMB", float64(b)/float64(mb))
	case b >= kb:
		return fmt.Sprintf("%.0fKB", float64(b)/float64(kb))
	default:
		return fmt.Sprintf("%dB", b)
	}
}

func appendStatus(text, status string) string {
	if text == "" {
		return status
	}
	return text + "\n\n" + status
}
