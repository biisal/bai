//go:build windows

package update

import (
	"context"
	"os"
	"path/filepath"
	"uuid"
)

func patternOrPath(ext string, prefix ...string) string {
	if len(prefix) == 0 {
		return "*.bai.old" + ext
	}
	return prefix[0] + ".bai.old" + ext
}

func Cleanup(ctx context.Context, binFolderPath string, ext string) error {
	files, err := filepath.Glob(filepath.Join(binFolderPath, patternOrPath(ext)))
	if err != nil {
		return err
	}

	for _, file := range files {
		if err := os.Remove(file); err != nil {
			return err
		}
	}
	return nil
}

func RenameFile(latestFilePath, binPath, ext string) error {
	uniquePath := patternOrPath(ext, uuid.New().String())
	if err := os.Rename(binPath, uniquePath); err != nil {
		_ = os.Rename(uniquePath, binPath) // roolback try
		return err
	}
	return nil
}
