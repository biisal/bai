//go:build windows

package tools

import (
	"context"
	"os/exec"
)

func setProcessGroup(cmd *exec.Cmd) {
	// Windows doesn't support process groups the same way; skip Setpgid
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
		// On Windows, we can't kill the process group, just kill the process
		_ = cmd.Process.Kill()
		return <-done
	}
}