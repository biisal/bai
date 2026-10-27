package web

import (
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/biisal/bai/internal/config"
	"github.com/biisal/bai/internal/consts"
	repo "github.com/biisal/bai/internal/db/sqlc"
	"github.com/biisal/bai/internal/domain"
	"github.com/biisal/bai/internal/files"
	"github.com/biisal/bai/internal/git"
	"github.com/biisal/bai/internal/update"
)

func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if t := r.URL.Query().Get("t"); t != "" {
		if !s.equal(t) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:     authCookieName,
			Value:    string(s.token),
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
		})
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	s.serveStatic(w, r)
}

func (s *Server) serveStatic(w http.ResponseWriter, r *http.Request) {
	if s.static == nil {
		s.unbuilt(w)
		return
	}
	name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
	if name == "" || name == "." {
		name = "index.html"
	}
	if !strings.HasPrefix(name, "assets/") && !hasExtension(name) {
		name = "index.html"
	}

	data, err := fs.ReadFile(s.static, name)
	if err != nil {
		if strings.HasPrefix(name, "assets/") || hasExtension(name) {
			http.NotFound(w, r)
			return
		}
		data, err = fs.ReadFile(s.static, "index.html")
		if err != nil {
			s.unbuilt(w)
			return
		}
		name = "index.html"
	}

	if name == "index.html" {
		w.Header().Set("Cache-Control", "no-cache")
	} else if strings.HasPrefix(name, "assets/") {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	}
	if ct := mime.TypeByExtension(path.Ext(name)); ct != "" {
		w.Header().Set("Content-Type", ct)
	}
	http.ServeContent(w, r, name, time.Time{}, bytes.NewReader(data))
}

func (s *Server) unbuilt(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusServiceUnavailable)
	_, _ = w.Write([]byte("bai web UI is not built. Run: make build-web\n"))
}

func hasExtension(name string) bool {
	return path.Ext(name) != ""
}

func (s *Server) handleState(w http.ResponseWriter, r *http.Request) {
	provider, model := s.gateway.Active()
	var running *runningState
	if run := s.registry.Active(); run != nil {
		running = &runningState{ConversationID: run.ConversationID}
	}
	writeJSON(w, http.StatusOK, stateResponse{
		Directory: files.GetBaseDir(),
		Version:   consts.Version,
		Provider:  provider.Name(),
		Model:     model,
		Running:   running,
	})
}

func (s *Server) handleModels(w http.ResponseWriter, r *http.Request) {
	activeProvider, activeModel := s.gateway.Active()
	providers := make([]providerModels, 0, len(s.config.Providers))
	for _, p := range s.config.Providers {
		models := make([]string, 0, len(p.Models))
		for _, m := range p.Models {
			models = append(models, m.ID)
		}
		providers = append(providers, providerModels{ID: p.Name, Name: p.Name, Models: models})
	}
	writeJSON(w, http.StatusOK, modelsResponse{Providers: providers, Provider: activeProvider.Name(), Model: activeModel})
}

func (s *Server) handleSetModel(w http.ResponseWriter, r *http.Request) {
	var req setModelRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if !s.validModel(req.Provider, req.Model) {
		writeError(w, http.StatusBadRequest, "unknown provider or model")
		return
	}
	if err := s.gateway.AddOrUpdateProvider(r.Context(), req.Provider, req.Model); err != nil {
		slog.Error("set model", "error", err)
		writeError(w, http.StatusBadRequest, "could not switch model")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) validModel(providerID, modelID string) bool {
	for _, p := range s.config.Providers {
		if p.Name != providerID {
			continue
		}
		for _, m := range p.Models {
			if m.ID == modelID {
				return true
			}
		}
	}
	return false
}

func (s *Server) handleConversations(w http.ResponseWriter, r *http.Request) {
	conversations, err := s.gateway.GetConversationsByCurrentDir(r.Context())
	if err != nil {
		slog.Error("list conversations", "error", err)
		writeError(w, http.StatusInternalServerError, "could not load conversations")
		return
	}
	items := make([]conversationItem, 0, len(conversations))
	for _, c := range conversations {
		items = append(items, conversationItem{ID: c.ID, Title: c.Title, CreatedAt: c.CreatedAt})
	}
	writeJSON(w, http.StatusOK, items)
}

func (s *Server) handleMessages(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid conversation id")
		return
	}
	conversation, err := s.conversationFor(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "conversation not found")
		return
	}
	messages, err := s.gateway.GetMessageRowsByConversationID(r.Context(), conversation.ID)
	if err != nil {
		slog.Error("load messages", "error", err)
		writeError(w, http.StatusInternalServerError, "could not load messages")
		return
	}
	if run := s.registry.Active(); run != nil && run.ConversationID == conversation.ID {
		messages = trimRunUserMessage(messages)
	}
	writeJSON(w, http.StatusOK, toUIMessages(messages))
}

func trimRunUserMessage(messages []repo.Message) []repo.Message {
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == domain.RoleUser {
			return messages[:i+1]
		}
	}
	return messages
}

func (s *Server) handleChat(w http.ResponseWriter, r *http.Request) {
	var req chatRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	text := strings.TrimSpace(req.Message.text())
	if text == "" {
		writeError(w, http.StatusBadRequest, "message is required")
		return
	}
	if s.registry.Active() != nil {
		writeError(w, http.StatusConflict, "a run is already active")
		return
	}

	conversation, isNew, err := s.resolveConversation(r.Context(), req.ID, text)
	if err != nil {
		slog.Error("resolve conversation", "error", err)
		writeError(w, http.StatusInternalServerError, "could not start conversation")
		return
	}

	run, sub, err := s.registry.Start(conversation.ID)
	if err != nil {
		if errors.Is(err, ErrRunActive) {
			writeError(w, http.StatusConflict, "a run is already active")
			return
		}
		writeError(w, http.StatusServiceUnavailable, "could not start run")
		return
	}
	s.gateway.SetConversation(&conversation)

	run.Emit(fStart(newMessageID()))
	if isNew {
		run.Emit(fConversation(conversation.ID, conversation.Title))
	}

	go run.Execute(func(ctx context.Context, run *Run) error {
		return s.gateway.StreamChatDecorated(ctx, text, decorate(run))
	})

	s.stream(w, r, sub)
}

func (s *Server) handleResume(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid conversation id")
		return
	}
	sub, err := s.registry.Subscribe(id)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "could not resume stream")
		return
	}
	if sub == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	s.stream(w, r, sub)
}

func (s *Server) handleStop(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid conversation id")
		return
	}
	s.registry.Stop(id)
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) stream(w http.ResponseWriter, r *http.Request, sub *Subscription) {
	defer sub.Close()

	h := w.Header()
	h.Set("Content-Type", "text/event-stream")
	h.Set("Cache-Control", "no-cache, no-transform")
	h.Set("x-vercel-ai-ui-message-stream", "v1")
	h.Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	rc := http.NewResponseController(w)
	_ = rc.Flush()

	for _, frame := range sub.Replay() {
		if !writeFrame(w, rc, frame) {
			return
		}
	}

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			if _, err := fmt.Fprint(w, ": ping\n\n"); err != nil {
				return
			}
			if err := rc.Flush(); err != nil {
				return
			}
		case frame, ok := <-sub.Frames():
			if !ok {
				return
			}
			if !writeFrame(w, rc, frame) {
				return
			}
			if string(frame) == frameDone {
				return
			}
		}
	}
}

func writeFrame(w http.ResponseWriter, rc *http.ResponseController, frame []byte) bool {
	if _, err := fmt.Fprintf(w, "data: %s\n\n", frame); err != nil {
		return false
	}
	return rc.Flush() == nil
}

func (s *Server) resolveConversation(ctx context.Context, rawID json.RawMessage, text string) (repo.Conversation, bool, error) {
	var idStr string
	if len(rawID) > 0 {
		if err := json.Unmarshal(rawID, &idStr); err != nil {
			var idNum int64
			if err := json.Unmarshal(rawID, &idNum); err == nil {
				idStr = strconv.FormatInt(idNum, 10)
			}
		}
	}
	if id, err := strconv.ParseInt(idStr, 10, 64); err == nil {
		conversation, err := s.conversationFor(ctx, id)
		if err == nil {
			return conversation, false, nil
		}
		if !errors.Is(err, sql.ErrNoRows) && !errors.Is(err, errWrongDirectory) {
			return repo.Conversation{}, false, err
		}
	}
	conversation, err := s.gateway.AddNewConversation(ctx, titleFor(text))
	if err != nil {
		return repo.Conversation{}, false, err
	}
	return conversation, true, nil
}

var errWrongDirectory = errors.New("conversation belongs to another directory")

func (s *Server) conversationFor(ctx context.Context, id int64) (repo.Conversation, error) {
	conversation, err := s.gateway.GetConversation(ctx, id)
	if err != nil {
		return repo.Conversation{}, err
	}
	if conversation.Directory != files.CurrentDir() {
		return repo.Conversation{}, errWrongDirectory
	}
	return conversation, nil
}

func titleFor(text string) string {
	const max = 80
	text = strings.TrimSpace(text)
	if utf8.RuneCountInString(text) <= max {
		return text
	}
	runes := []rune(text)
	return string(runes[:max])
}

func newMessageID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 10)
	}
	return hex.EncodeToString(b)
}

func decodeJSON(w http.ResponseWriter, r *http.Request, v any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	return json.NewDecoder(r.Body).Decode(v)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("write json response", "error", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{Error: message})
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
	var req renameConversationRequest
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
	writeJSON(w, http.StatusOK, renameConversationResponse{ID: conversation.ID, Title: conversation.Title})
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

	writeJSON(w, http.StatusOK, settingsResponse{
		Models: providers,
		Active: settingsActive{Provider: activeProvider.Name(), Model: activeModel},
		Git: settingsGit{
			Initialized: initialized,
			Dirty:       dirty,
			Dir:         s.config.GitDirName,
			Hint:        "Run: git --git-dir=" + s.config.GitDirName + " init",
		},
		About: settingsAbout{
			Version:      consts.Version,
			Directory:    files.CurrentDir(),
			ConfigPath:   config.DefaultConfigPath(),
			LogPath:      s.config.LogFilePath,
			LatestUpdate: latest,
		},
	})
}

func (s *Server) handleSkills(w http.ResponseWriter, r *http.Request) {
	list := s.gateway.Skills
	items := make([]skillItem, 0, len(list))
	for _, skill := range list {
		items = append(items, skillItem{Name: skill.Name, Description: skill.Description, Source: skill.Location})
	}
	writeJSON(w, http.StatusOK, items)
}

func (s *Server) handlePlugins(w http.ResponseWriter, r *http.Request) {
	items := make([]pluginItem, 0)
	data, err := os.ReadFile(s.config.PluginsPath)
	if err == nil {
		var manifest []pluginManifest
		if json.Unmarshal(data, &manifest) == nil {
			for _, plugin := range manifest {
				items = append(items, pluginItem{
					Name:        strings.TrimSuffix(plugin.FileName, filepath.Ext(plugin.FileName)),
					Description: plugin.Description,
					Source:      filepath.Join(filepath.Dir(s.config.PluginsPath), plugin.FileName),
				})
			}
		}
	}
	writeJSON(w, http.StatusOK, items)
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
	gitRepo := git.New(s.config.GitDirName)
	if _, err := gitRepo.RunGitCommandContext(r.Context(), "rev-parse", "--verify", "HEAD"); err != nil {
		writeJSON(w, http.StatusOK, diffResponse{})
		return
	}
	diff, err := gitRepo.RunGitCommandContext(r.Context(), diffArgs...)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}
		writeError(w, http.StatusInternalServerError, "could not read changes")
		return
	}
	diff, truncated := capDiff(diff)
	writeJSON(w, http.StatusOK, diffResponse{Diff: diff, Truncated: truncated})
}
