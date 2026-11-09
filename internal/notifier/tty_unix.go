//go:build !windows

package notifier

import "os"

func openTTY() (*os.File, error) {
	return os.OpenFile("/dev/tty", os.O_WRONLY, 0)
}
