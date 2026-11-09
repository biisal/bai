//go:build windows

package notifier

import "os"

func openTTY() (*os.File, error) {
	return os.OpenFile("CONOUT$", os.O_WRONLY, 0)
}
