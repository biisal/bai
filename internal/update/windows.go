//go:build windows

package update

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const oldSuffix = ".old"

func cleanup(binPath string) {
	dir := filepath.Dir(binPath)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	prefix := filepath.Base(binPath) + "."
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, prefix) && strings.HasSuffix(name, oldSuffix) {
			_ = os.Remove(filepath.Join(dir, name))
		}
	}
}

func replaceBinary(newPath, binPath string) error {
	oldPath := fmt.Sprintf("%s.%d%s", binPath, time.Now().UnixNano(), oldSuffix)

	if err := os.Rename(binPath, oldPath); err != nil {
		return err
	}
	if err := os.Rename(newPath, binPath); err != nil {
		return errors.Join(err, os.Rename(oldPath, binPath))
	}
	return nil
}
