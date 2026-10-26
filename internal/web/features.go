package web

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/biisal/bai/internal/config"
	"github.com/biisal/bai/internal/consts"
	"github.com/biisal/bai/internal/files"
	"github.com/biisal/bai/internal/git"
	"github.com/biisal/bai/internal/skills"
	"github.com/biisal/bai/internal/tui/commands"
	"github.com/biisal/bai/internal/update"
)

const (
	fileResultLimit = 20
	maxDiffBytes    = 1 << 20
)

var diffArgs = []string{
	"show", "--format=", "--no-color", "--no-ext-diff", "--no-textconv", "--no-renames", "--unified=3", "HEAD", "--",
}

type fileResult struct {
	Path  string `json:"path"`
	IsDir bool   `json:"isDir"`
}

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
	for _, item := range commands.FileItems() {
		f, ok := item.(commands.FileItem)
		if !ok || !strings.Contains(strings.ToLower(f.RelFilePath), lower) {
			continue
		}
		out = append(out, fileResult{Path: f.RelFilePath})
		if len(out) == fileResultLimit {
			break
		}
	}
	return out
}

func (s *Server) handleFiles(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if !validFileQuery(q) {
		writeError(w, http.StatusBadRequest, "invalid file query")
		return
	}
	writeJSON(w, http.StatusOK, searchFiles(q))
}

func (s *Server) handleDeleteConversation(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid conversation id")
		return
	}
	if run := s.registry.Active(); run != nil && run.ConversationID == id {
		writeError(w, http.StatusConflict, "a run is active on this conversation")
		return
	}
	if _, err := s.conversationFor(r.Context(), id); err != nil {
		writeError(w, http.StatusNotFound, "conversation not found")
		return
	}
	if err := s.gateway.DeleteConversation(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, "could not delete conversation")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleRenameConversation(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid conversation id")
		return
	}
	var req struct {
		Title string `json:"title"`
	}
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	title := strings.TrimSpace(req.Title)
	if title == "" || utf8.RuneCountInString(title) > 200 {
		writeError(w, http.StatusBadRequest, "title must be 1-200 characters")
		return
	}
	if run := s.registry.Active(); run != nil && run.ConversationID == id {
		writeError(w, http.StatusConflict, "a run is active on this conversation")
		return
	}
	if _, err := s.conversationFor(r.Context(), id); err != nil {
		writeError(w, http.StatusNotFound, "conversation not found")
		return
	}
	conversation, err := s.gateway.RenameConversation(r.Context(), id, title)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not rename conversation")
		return
	}
	writeJSON(w, http.StatusOK, struct {
		ID    int64  `json:"id"`
		Title string `json:"title"`
	}{ID: conversation.ID, Title: conversation.Title})
}

type settingsProvider struct {
	ID         string          `json:"id"`
	Name       string          `json:"name"`
	Configured bool            `json:"configured"`
	Models     []settingsModel `json:"models"`
}

type settingsModel struct {
	ID        string `json:"id"`
	Context   int    `json:"context"`
	MaxOutput int    `json:"maxOutput"`
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

func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	activeProvider, activeModel := s.gateway.Active()
	providers := settingsProviders(s.config.Providers)

	gitRepo := git.New(s.config.GitDirName)
	initialized, _ := gitRepo.CheckIfGitInitialized()
	dirty := false
	if initialized {
		dirty, _, _ = gitRepo.CheckIfDirty()
	}

	latest := ""
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if release, err := update.NewFetcher().LatestRelease(ctx); err == nil {
		latest = release.TagName
	}

	writeJSON(w, http.StatusOK, struct {
		Models []settingsProvider `json:"models"`
		Active struct {
			Provider string `json:"provider"`
			Model    string `json:"model"`
		} `json:"active"`
		Git struct {
			Initialized bool   `json:"initialized"`
			Dirty       bool   `json:"dirty"`
			Dir         string `json:"dir"`
			Hint        string `json:"hint"`
		} `json:"git"`
		About struct {
			Version      string `json:"version"`
			Directory    string `json:"directory"`
			ConfigPath   string `json:"configPath"`
			LogPath      string `json:"logPath"`
			LatestUpdate string `json:"latestUpdate"`
		} `json:"about"`
	}{
		Models: providers,
		Active: struct {
			Provider string `json:"provider"`
			Model    string `json:"model"`
		}{Provider: activeProvider.Name(), Model: activeModel},
		Git: struct {
			Initialized bool   `json:"initialized"`
			Dirty       bool   `json:"dirty"`
			Dir         string `json:"dir"`
			Hint        string `json:"hint"`
		}{
			Initialized: initialized,
			Dirty:       dirty,
			Dir:         s.config.GitDirName,
			Hint:        "Run: git --git-dir=" + s.config.GitDirName + " init",
		},
		About: struct {
			Version      string `json:"version"`
			Directory    string `json:"directory"`
			ConfigPath   string `json:"configPath"`
			LogPath      string `json:"logPath"`
			LatestUpdate string `json:"latestUpdate"`
		}{
			Version:      consts.Version,
			Directory:    files.CurrentDir(),
			ConfigPath:   config.DefaultConfigPath(),
			LogPath:      s.config.LogFilePath,
			LatestUpdate: latest,
		},
	})
}

func (s *Server) handleSkills(w http.ResponseWriter, r *http.Request) {
	paths := append([]string{}, s.config.SkillsPaths...)
	paths = append(paths, config.DefaultSkillsPaths()...)
	list := skills.LoadSkills(paths...)
	items := make([]struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Source      string `json:"source"`
	}, 0, len(list))
	for _, skill := range list {
		items = append(items, struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			Source      string `json:"source"`
		}{Name: skill.Name, Description: skill.Description, Source: skill.Location})
	}
	writeJSON(w, http.StatusOK, items)
}

type pluginManifest struct {
	FileName    string `json:"file_name"`
	Description string `json:"description"`
}

func (s *Server) handlePlugins(w http.ResponseWriter, r *http.Request) {
	items := make([]struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Source      string `json:"source"`
	}, 0)
	data, err := os.ReadFile(s.config.PluginsPath)
	if err == nil {
		var manifest []pluginManifest
		if json.Unmarshal(data, &manifest) == nil {
			for _, plugin := range manifest {
				items = append(items, struct {
					Name        string `json:"name"`
					Description string `json:"description"`
					Source      string `json:"source"`
				}{
					Name:        strings.TrimSuffix(plugin.FileName, filepath.Ext(plugin.FileName)),
					Description: plugin.Description,
					Source:      filepath.Join(filepath.Dir(s.config.PluginsPath), plugin.FileName),
				})
			}
		}
	}
	writeJSON(w, http.StatusOK, items)
}

func capDiff(diff string) (string, bool) {
	if len(diff) <= maxDiffBytes {
		return strings.ToValidUTF8(diff, ""), false
	}
	return strings.ToValidUTF8(diff[:maxDiffBytes], ""), true
}

func (s *Server) handleDiff(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid conversation id")
		return
	}
	if _, err := s.conversationFor(r.Context(), id); err != nil {
		writeError(w, http.StatusNotFound, "conversation not found")
		return
	}
	repo := git.New(s.config.GitDirName)
	if _, err := repo.RunGitCommandContext(r.Context(), "rev-parse", "--verify", "HEAD"); err != nil {
		writeJSON(w, http.StatusOK, struct {
			Diff      string `json:"diff"`
			Truncated bool   `json:"truncated"`
		}{})
		return
	}
	diff, err := repo.RunGitCommandContext(r.Context(), diffArgs...)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}
		writeError(w, http.StatusInternalServerError, "could not read changes")
		return
	}
	diff, truncated := capDiff(diff)
	writeJSON(w, http.StatusOK, struct {
		Diff      string `json:"diff"`
		Truncated bool   `json:"truncated"`
	}{Diff: diff, Truncated: truncated})
}
