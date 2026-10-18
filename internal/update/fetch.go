package update

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/biisal/bai/internal/consts"
)

type Release struct {
	TagName string `json:"tag_name"`
	URL     string `json:"url"`
	Assets  []struct {
		Name               string `json:"name"`
		URL                string `json:"url"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

type Fetcher struct {
	client *http.Client
}

func NewFetcher() *Fetcher {
	client := &http.Client{}
	return &Fetcher{
		client: client,
	}
}

func (f *Fetcher) GetLatestVersion(ctx context.Context) (Release, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", consts.RepositoryName)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return Release{}, err
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return Release{}, err
	}
	defer func() {
		err = resp.Body.Close()
		slog.Error("closed response body", "err", err)
	}()

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return Release{}, err
	}
	return release, nil
}

func nameIncludeOsAndArch(name string) bool {
	arch := runtime.GOARCH
	osName := runtime.GOOS
	return strings.Contains(name, arch) && strings.Contains(name, osName)
}

func (f *Fetcher) DownloadUpdate(ctx context.Context, release Release, targetPath string) error {
	url := ""
	for _, asset := range release.Assets {
		if nameIncludeOsAndArch(asset.Name) {
			url = asset.BrowserDownloadURL
		}
	}
	if url == "" {
		return fmt.Errorf("no asset found for os/arch: %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		err = resp.Body.Close()
		slog.Error("closed response body", "err", err)
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	out, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer func() {
		err = out.Close()
		slog.Error("closed response body", "err", err)
	}()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
