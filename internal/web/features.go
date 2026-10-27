package web

import (
	"path/filepath"
	"strings"

	"github.com/biisal/bai/internal/config"
	"github.com/biisal/bai/internal/files"
)

const (
	fileResultLimit = 20
	maxDiffBytes    = 1 << 20
)

func validFileQuery(q string) bool {
	if filepath.IsAbs(q) || filepath.VolumeName(q) != "" {
		return false
	}
	if len(q) >= 2 && q[1] == ':' && ((q[0] >= 'a' && q[0] <= 'z') || (q[0] >= 'A' && q[0] <= 'Z')) {
		return false
	}
	if strings.HasPrefix(q, "/") || strings.HasPrefix(q, `\`) {
		return false
	}
	for _, part := range strings.FieldsFunc(q, func(r rune) bool { return r == '/' || r == '\\' }) {
		if part == ".." {
			return false
		}
	}
	return true
}

func searchFiles(q string) []fileResult {
	lower := strings.ToLower(q)
	out := make([]fileResult, 0, fileResultLimit)
	for _, f := range files.List() {
		if !strings.Contains(strings.ToLower(f.RelPath), lower) {
			continue
		}
		out = append(out, fileResult{Path: f.RelPath})
		if len(out) == fileResultLimit {
			break
		}
	}
	return out
}

func settingsProviders(providers []config.ProviderConfig) []settingsProvider {
	out := make([]settingsProvider, 0, len(providers))
	for _, p := range providers {
		models := make([]settingsModel, 0, len(p.Models))
		for _, m := range p.Models {
			models = append(models, settingsModel{ID: m.ID, Context: m.Context, MaxOutput: m.MaxOutput})
		}
		out = append(out, settingsProvider{
			ID:         p.Name,
			Name:       p.Name,
			Configured: strings.TrimSpace(p.APIKey) != "",
			Models:     models,
		})
	}
	return out
}

func capDiff(diff string) (string, bool) {
	if len(diff) <= maxDiffBytes {
		return strings.ToValidUTF8(diff, ""), false
	}
	return strings.ToValidUTF8(diff[:maxDiffBytes], ""), true
}
