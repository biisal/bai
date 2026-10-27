package web

import (
	"crypto/subtle"
	"net/http"
	"net/url"
	"strings"
)

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
