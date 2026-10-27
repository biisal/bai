package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	repo "github.com/biisal/bai/internal/db/sqlc"
	"github.com/biisal/bai/internal/domain"
)

func TestGuard(t *testing.T) {
	ok := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	tests := []struct {
		name        string
		method      string
		target      string
		host        string
		origin      string
		contentType string
		want        int
	}{
		{name: "api get", method: http.MethodGet, target: "/api/state", host: "127.0.0.1:9999", want: http.StatusNoContent},
		{name: "root", method: http.MethodGet, target: "/", host: "127.0.0.1:9999", want: http.StatusNoContent},
		{name: "bad host", method: http.MethodGet, target: "/", host: "evil.example:9999", want: http.StatusForbidden},
		{name: "cross origin post", method: http.MethodPost, target: "/api/chat", host: "127.0.0.1:9999", origin: "http://evil.example", contentType: "application/json", want: http.StatusForbidden},
		{name: "cross origin get", method: http.MethodGet, target: "/api/settings", host: "127.0.0.1:9999", origin: "http://evil.example", want: http.StatusForbidden},
		{name: "same origin post", method: http.MethodPost, target: "/api/chat", host: "127.0.0.1:9999", origin: "http://127.0.0.1:9999", contentType: "application/json", want: http.StatusNoContent},
		{name: "post without json", method: http.MethodPost, target: "/api/chat", host: "127.0.0.1:9999", want: http.StatusUnsupportedMediaType},
		{name: "files", method: http.MethodGet, target: "/api/files?q=main", host: "127.0.0.1:9999", want: http.StatusNoContent},
		{name: "delete", method: http.MethodDelete, target: "/api/conversations/1", host: "127.0.0.1:9999", want: http.StatusNoContent},
		{name: "cross origin delete", method: http.MethodDelete, target: "/api/conversations/1", host: "127.0.0.1:9999", origin: "http://evil.example", want: http.StatusForbidden},
		{name: "patch without json", method: http.MethodPatch, target: "/api/conversations/1", host: "127.0.0.1:9999", want: http.StatusUnsupportedMediaType},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := &Server{hosts: allowedHosts("127.0.0.1:9999")}
			req := httptest.NewRequest(tt.method, tt.target, nil)
			req.Host = tt.host
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}
			rec := httptest.NewRecorder()
			server.guard(ok).ServeHTTP(rec, req)
			if rec.Code != tt.want {
				t.Errorf("got %d, want %d", rec.Code, tt.want)
			}
		})
	}
}

func TestGuardAllowsDevViteOrigin(t *testing.T) {
	server := &Server{
		hosts: allowedHosts("127.0.0.1:9999"),
		dev:   true,
	}
	req := httptest.NewRequest(http.MethodPost, "/api/chat", nil)
	req.Host = "127.0.0.1:9999"
	req.Header.Set("Origin", viteOrigin)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	server.guard(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})).ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("got %d, want %d", rec.Code, http.StatusNoContent)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != viteOrigin {
		t.Fatalf("allow origin = %q, want %q", got, viteOrigin)
	}
}

func TestStaticFallbackAndMissingAsset(t *testing.T) {
	static := fstest.MapFS{
		"index.html":    {Data: []byte("<html>ok</html>")},
		"assets/app.js": {Data: []byte("console.log(1)")},
	}
	server := &Server{static: static}

	req := httptest.NewRequest(http.MethodGet, "/c/42", nil)
	rec := httptest.NewRecorder()
	server.serveStatic(rec, req)
	if rec.Code != http.StatusOK || rec.Body.String() != "<html>ok</html>" {
		t.Fatalf("fallback got %d %q", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/assets/missing.js", nil)
	rec = httptest.NewRecorder()
	server.serveStatic(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("missing asset got %d, want %d", rec.Code, http.StatusNotFound)
	}

	req = httptest.NewRequest(http.MethodGet, "/assets/app.js", nil)
	rec = httptest.NewRecorder()
	server.serveStatic(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("asset got %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get("Cache-Control"); got != "public, max-age=31536000, immutable" {
		t.Fatalf("asset cache = %q", got)
	}
}

func TestTrimRunUserMessageKeepsPrompt(t *testing.T) {
	messages := []repo.Message{
		{Role: domain.RoleUser},
		{Role: domain.RoleAssistant},
		{Role: domain.RoleUser},
		{Role: domain.RoleAssistant},
	}
	got := trimRunUserMessage(messages)
	if len(got) != 3 {
		t.Fatalf("got %d messages, want 3", len(got))
	}
	if got[len(got)-1].Role != domain.RoleUser {
		t.Fatalf("last role = %q, want user", got[len(got)-1].Role)
	}
}
