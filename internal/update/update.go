package update

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/biisal/bai/internal/consts"
)

func checkIfUpdateAvailable(ctx context.Context, current string, latest string) (bool, error) {
	const versionPattern = `^v[0-9]+\.[0-9]+\.[0-9]+$`

	matched, err := regexp.MatchString(versionPattern, latest)
	if err != nil {
		return false, err
	}
	if !matched {
		return false, fmt.Errorf("invalid latest version: %s", latest)
	}

	matched, err = regexp.MatchString(versionPattern, current)
	if err != nil {
		return false, err
	}
	if !matched {
		return false, fmt.Errorf("invalid current version: %s", current)
	}

	latestParts := strings.Split(latest[1:], ".")
	currentParts := strings.Split(current[1:], ".")

	for i := range latestParts {
		latestNum, err := strconv.Atoi(latestParts[i])
		if err != nil {
			return false, err
		}

		currentNum, err := strconv.Atoi(currentParts[i])
		if err != nil {
			return false, err
		}

		if latestNum > currentNum {
			return true, nil
		}
		if latestNum < currentNum {
			return false, nil
		}
	}

	return false, nil
}

func Update(ctx context.Context, version string) error {
	binPath, err := os.Executable()
	if err != nil {
		return err
	}
	ext := filepath.Ext(binPath)
	bp := filepath.Base(binPath)

	if err := Cleanup(ctx, bp, ext); err != nil {
		return err
	}

	if (version == "") ||
		(version == consts.VersionLocal) ||
		(version == consts.VersionDev) {
		return nil
	}

	fetcher := NewFetcher()
	latest, err := fetcher.GetLatestVersion(ctx)
	if err != nil {
		return err
	}

	updateAvailable, err := checkIfUpdateAvailable(ctx, consts.Version, latest.TagName)
	if err != nil {
		return err
	}
	if !updateAvailable {
		return nil
	}

	fileName := fmt.Sprintf("%s.downlaoded.%s", bp, ext)
	if err := fetcher.DownloadUpdate(ctx, latest, fileName); err != nil {
		return err
	}

	if err := RenameFile(fileName, binPath, ext); err != nil {
		return err
	}

	return nil
}
