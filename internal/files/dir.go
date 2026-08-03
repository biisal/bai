package files

import "os"

func CurrentDir() string {
	currentWd, err := os.Getwd()
	if err != nil {
		// TODO: show warnign to UI
	}
	return currentWd
}
