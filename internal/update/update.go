package update

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/biisal/bai/internal/consts"
)

const binaryName = "bai"

var versionPattern = regexp.MustCompile(`^v(\d+)\.(\d+)\.(\d+)$`)

func parseVersion(v string) ([3]int, error) {
	m := versionPattern.FindStringSubmatch(v)
	if m == nil {
		return [3]int{}, fmt.Errorf("invalid version: %q", v)
	}

	var parts [3]int
	for i := range parts {
		n, err := strconv.Atoi(m[i+1])
		if err != nil {
			return [3]int{}, err
		}
		parts[i] = n
	}
	return parts, nil
}

func isNewer(current, latest string) (bool, error) {
	cur, err := parseVersion(current)
	if err != nil {
		return false, err
	}
	lat, err := parseVersion(latest)
	if err != nil {
		return false, err
	}
	return slices.Compare(lat[:], cur[:]) > 0, nil
}

// Update installs the latest release if it is newer than version and returns
// its tag. It returns an empty tag when nothing was installed.
func Update(ctx context.Context, version string) (string, error) {
	if version == "" || version == consts.VersionLocal || version == consts.VersionDev {
		return "", nil
	}

	binPath, err := os.Executable()
	if err != nil {
		return "", err
	}
	binPath, err = filepath.EvalSymlinks(binPath)
	if err != nil {
		return "", err
	}

	cleanup(binPath)

	fetcher := NewFetcher()
	release, err := fetcher.LatestRelease(ctx)
	if err != nil {
		return "", err
	}

	newer, err := isNewer(version, release.TagName)
	if err != nil {
		return "", err
	}
	if !newer {
		return "", nil
	}

	ext := filepath.Ext(binPath)
	newPath := strings.TrimSuffix(binPath, ext) + ".update" + ext
	defer func() { _ = os.Remove(newPath) }()

	if err := fetcher.Download(ctx, release, newPath); err != nil {
		return "", err
	}
	if err := replaceBinary(newPath, binPath); err != nil {
		return "", err
	}
	return release.TagName, nil
}
