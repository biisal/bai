//go:build !windows

package update

import "os"

func cleanup(string) {}

func replaceBinary(newPath, binPath string) error {
	return os.Rename(newPath, binPath)
}
