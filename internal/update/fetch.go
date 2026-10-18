package update

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/biisal/bai/internal/consts"
)

const (
	checksumsAssetName = "checksums.txt"
	httpTimeout        = 2 * time.Minute
)

type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type Release struct {
	TagName string  `json:"tag_name"`
	Assets  []Asset `json:"assets"`
}

func (r Release) findAsset(match func(name string) bool) (Asset, bool) {
	for _, a := range r.Assets {
		if match(a.Name) {
			return a, true
		}
	}
	return Asset{}, false
}

func (r Release) archiveAsset() (Asset, bool) {
	prefix := fmt.Sprintf("%s_%s_%s.", binaryName, runtime.GOOS, runtime.GOARCH)
	return r.findAsset(func(name string) bool { return strings.HasPrefix(name, prefix) })
}

func (r Release) checksumsAsset() (Asset, bool) {
	return r.findAsset(func(name string) bool { return name == checksumsAssetName })
}

type Fetcher struct {
	client *http.Client
}

func NewFetcher() *Fetcher {
	return &Fetcher{client: &http.Client{Timeout: httpTimeout}}
}

func (f *Fetcher) get(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("GET %s: unexpected status %s", url, resp.Status)
	}
	return resp, nil
}

func (f *Fetcher) copyTo(ctx context.Context, url string, w io.Writer) error {
	resp, err := f.get(ctx, url)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	_, err = io.Copy(w, resp.Body)
	return err
}

func (f *Fetcher) LatestRelease(ctx context.Context) (Release, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", consts.RepositoryName)

	resp, err := f.get(ctx, url)
	if err != nil {
		return Release{}, err
	}
	defer func() { _ = resp.Body.Close() }()

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return Release{}, err
	}
	return release, nil
}

func (f *Fetcher) expectedChecksum(ctx context.Context, url, name string) (string, error) {
	resp, err := f.get(ctx, url)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) == 2 && fields[1] == name {
			return fields[0], nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("no checksum listed for %s", name)
}

func (f *Fetcher) downloadArchive(ctx context.Context, url, wantSum string) (_ string, err error) {
	tmp, err := os.CreateTemp("", "bai-update-*")
	if err != nil {
		return "", err
	}
	defer func() {
		if err != nil {
			_ = os.Remove(tmp.Name())
		}
	}()

	hash := sha256.New()
	copyErr := f.copyTo(ctx, url, io.MultiWriter(tmp, hash))
	if err = errors.Join(copyErr, tmp.Close()); err != nil {
		return "", err
	}

	if got := hex.EncodeToString(hash.Sum(nil)); got != wantSum {
		return "", fmt.Errorf("checksum mismatch: got %s, want %s", got, wantSum)
	}
	return tmp.Name(), nil
}

func (f *Fetcher) Download(ctx context.Context, release Release, targetPath string) error {
	archive, ok := release.archiveAsset()
	if !ok {
		return fmt.Errorf("no asset found for %s/%s", runtime.GOOS, runtime.GOARCH)
	}
	checksums, ok := release.checksumsAsset()
	if !ok {
		return fmt.Errorf("release %s has no %s", release.TagName, checksumsAssetName)
	}

	wantSum, err := f.expectedChecksum(ctx, checksums.BrowserDownloadURL, archive.Name)
	if err != nil {
		return err
	}

	archivePath, err := f.downloadArchive(ctx, archive.BrowserDownloadURL, wantSum)
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(archivePath) }()

	return extractBinary(archivePath, archive.Name, targetPath)
}
