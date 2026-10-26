package web

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/subtle"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"mime"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/biisal/bai/internal/agent"
	"github.com/biisal/bai/internal/config"
	"github.com/biisal/bai/internal/consts"
	repo "github.com/biisal/bai/internal/db/sqlc"
	"github.com/biisal/bai/internal/domain"
	"github.com/biisal/bai/internal/files"
)

const (
	authCookieName = "bai_web_token"
	viteOrigin     = "http://localhost:5173"
)

type Options struct {
	Gateway  *agent.Gateway
	Config   *config.Config
	Registry *Registry
	Token    string
	Addr     string
	Dev      bool
	Static   fs.FS
}

type Server struct {
	gateway  *agent.Gateway
	config   *config.Config
	registry *Registry
	token    []byte
	addr     string
	dev      bool
	static   fs.FS
	mux      *http.ServeMux
	hosts    map[string]bool
}

func NewServer(opts Options) *Server {
	s := &Server{
		gateway:  opts.Gateway,
		config:   opts.Config,
		registry: opts.Registry,
		token:    []byte(opts.Token),
		addr:     opts.Addr,
		dev:      opts.Dev,
		static:   opts.Static,
		mux:      http.NewServeMux(),
		hosts:    allowedHosts(opts.Addr),
	}
	s.routes()
	return s
}

func (s *Server) Handler() http.Handler {
	return s.guard(s.mux)
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /api/state", s.handleState)
	s.mux.HandleFunc("GET /api/models", s.handleModels)
	s.mux.HandleFunc("PUT /api/model", s.handleSetModel)
	s.mux.HandleFunc("GET /api/conversations", s.handleConversations)
	s.mux.HandleFunc("GET /api/conversations/{id}/messages", s.handleMessages)
	s.mux.HandleFunc("DELETE /api/conversations/{id}", s.handleDeleteConversation)
	s.mux.HandleFunc("PATCH /api/conversations/{id}", s.handleRenameConversation)
	s.mux.HandleFunc("GET /api/conversations/{id}/diff", s.handleDiff)
	s.mux.HandleFunc("GET /api/files", s.handleFiles)
	s.mux.HandleFunc("GET /api/settings", s.handleSettings)
	s.mux.HandleFunc("GET /api/skills", s.handleSkills)
	s.mux.HandleFunc("GET /api/plugins", s.handlePlugins)
	s.mux.HandleFunc("POST /api/chat", s.handleChat)
	s.mux.HandleFunc("GET /api/chat/{id}/stream", s.handleResume)
	s.mux.HandleFunc("POST /api/chat/{id}/stop", s.handleStop)
	s.mux.HandleFunc("/", s.handleRoot)
}

func allowedHosts(addr string) map[string]bool {
	hosts := map[string]bool{strings.ToLower(addr): true}
	_, port, err := splitHostPort(addr)
	if err != nil {
		return hosts
	}
	if port != "" {
		for _, h := range []string{"localhost", "127.0.0.1", "[::1]"} {
			hosts[h+":"+port] = true
		}
	}
	return hosts
}

func splitHostPort(addr string) (string, string, error) {
	i := strings.LastIndex(addr, ":")
	if i < 0 {
		return addr, "", nil
	}
	return addr[:i], addr[i+1:], nil
}

func (s *Server) guard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self'")
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("X-Frame-Options", "DENY")

		if !s.hostOK(r.Host) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		origin := r.Header.Get("Origin")
		if origin == viteOrigin && s.dev {
			h.Set("Access-Control-Allow-Origin", viteOrigin)
			h.Set("Access-Control-Allow-Credentials", "true")
			h.Set("Vary", "Origin")
			if r.Method == http.MethodOptions {
				h.Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
				h.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
				w.WriteHeader(http.StatusNoContent)
				return
			}
		} else if origin != "" && !s.originMatchesHost(origin, r.Host) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		if strings.HasPrefix(r.URL.Path, "/api/") {
			if !s.tokenOK(r) {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
				if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
					http.Error(w, "content type must be application/json", http.StatusUnsupportedMediaType)
					return
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) hostOK(host string) bool {
	return s.hosts[strings.ToLower(host)]
}

func (s *Server) originMatchesHost(origin, host string) bool {
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	return strings.EqualFold(u.Host, host)
}

func (s *Server) tokenOK(r *http.Request) bool {
	if strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") {
		if s.equal(r.Header.Get("Authorization")[len("Bearer "):]) {
			return true
		}
	}
	if c, err := r.Cookie(authCookieName); err == nil && s.equal(c.Value) {
		return true
	}
	return false
}

func (s *Server) equal(got string) bool {
	return subtle.ConstantTimeCompare([]byte(got), s.token) == 1
}

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
	var running *struct {
		ConversationID int64 `json:"conversationId"`
	}
	if run := s.registry.Active(); run != nil {
		running = &struct {
			ConversationID int64 `json:"conversationId"`
		}{ConversationID: run.ConversationID}
	}
	writeJSON(w, http.StatusOK, struct {
		Directory string `json:"directory"`
		Version   string `json:"version"`
		Provider  string `json:"provider"`
		Model     string `json:"model"`
		Running   any    `json:"running"`
	}{
		Directory: files.GetBaseDir(),
		Version:   consts.Version,
		Provider:  provider.Name(),
		Model:     model,
		Running:   running,
	})
}

func (s *Server) handleModels(w http.ResponseWriter, r *http.Request) {
	activeProvider, activeModel := s.gateway.Active()
	type providerModels struct {
		ID     string   `json:"id"`
		Name   string   `json:"name"`
		Models []string `json:"models"`
	}
	providers := make([]providerModels, 0, len(s.config.Providers))
	for _, p := range s.config.Providers {
		models := make([]string, 0, len(p.Models))
		for _, m := range p.Models {
			models = append(models, m.ID)
		}
		providers = append(providers, providerModels{ID: p.Name, Name: p.Name, Models: models})
	}
	writeJSON(w, http.StatusOK, struct {
		Providers []providerModels `json:"providers"`
		Provider  string           `json:"provider"`
		Model     string           `json:"model"`
	}{Providers: providers, Provider: activeProvider.Name(), Model: activeModel})
}

func (s *Server) handleSetModel(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Provider string `json:"provider"`
		Model    string `json:"model"`
	}
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
	type item struct {
		ID        int64     `json:"id"`
		Title     string    `json:"title"`
		CreatedAt time.Time `json:"createdAt"`
	}
	items := make([]item, 0, len(conversations))
	for _, c := range conversations {
		items = append(items, item{ID: c.ID, Title: c.Title, CreatedAt: c.CreatedAt})
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
	var req struct {
		ID      json.RawMessage `json:"id"`
		Message uiMessage       `json:"message"`
	}
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

type uiMessage struct {
	ID    string   `json:"id"`
	Role  string   `json:"role"`
	Text  string   `json:"text"`
	Parts []uiPart `json:"parts"`
}

type uiPart struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func (m uiMessage) text() string {
	if len(m.Parts) == 0 {
		return m.Text
	}
	var b strings.Builder
	for _, p := range m.Parts {
		if p.Type == "text" || p.Type == "" {
			b.WriteString(p.Text)
		}
	}
	return b.String()
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
	writeJSON(w, status, struct {
		Error string `json:"error"`
	}{Error: message})
}
