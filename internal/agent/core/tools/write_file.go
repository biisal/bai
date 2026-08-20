package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

func WriteFile(ctx context.Context, path string, content string) error {
	if path == "" {
		return fmt.Errorf("path is empty")
	}

	resolved, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("cannot resolve %q: %w", path, err)
	}

	return withFileMutationQueue(resolved, func() error {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("write aborted for %q: %w", resolved, err)
		}

		dir := filepath.Dir(resolved)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("cannot create directory %q: %w", dir, err)
		}

		if err := ctx.Err(); err != nil {
			return fmt.Errorf("write aborted for %q: %w", resolved, err)
		}

		if err := os.WriteFile(resolved, []byte(content), 0o644); err != nil {
			return fmt.Errorf("cannot write %q: %w", resolved, err)
		}

		if err := ctx.Err(); err != nil {
			return fmt.Errorf("write aborted for %q: %w", resolved, err)
		}

		return nil
	})
}
