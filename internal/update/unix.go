//go:build !windows

package update

import (
	"context"
	"os"
)

func Cleanup(ctx context.Context, binFolderPath string, ext string) error {
	return nil
}

func RenameFile(latestFilePath, binPath string, ext string) error {
	return os.Rename(latestFilePath, binPath)
}
