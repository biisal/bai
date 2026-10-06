package agent

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	fantasy "charm.land/fantasy"
	"github.com/biisal/bai/internal/config"
)

// TestOpenCodeVariantHeadersAgainstCapture stands up a local stand-in for
// opencode.ai/zen/v1 and records the exact headers bai sends, so they can be
// compared with a real captured opencode request.
func TestOpenCodeVariantHeadersAgainstCapture(t *testing.T) {
	var got http.Header
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Clone()
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)

		w.Header().Set("content-type", "text/event-stream")
		w.WriteHeader(200)
		chunks := []map[string]any{
			{"id": "c1", "object": "chat.completion.chunk", "created": time.Now().Unix(), "model": "m", "choices": []map[string]any{{"index": 0, "delta": map[string]any{"role": "assistant", "content": "pong"}, "finish_reason": nil}}},
			{"id": "c1", "object": "chat.completion.chunk", "created": time.Now().Unix(), "model": "m", "choices": []map[string]any{{"index": 0, "delta": map[string]any{}, "finish_reason": "stop"}}},
		}
		for _, c := range chunks {
			b, _ := json.Marshal(c)
			_, _ = w.Write([]byte("data: " + string(b) + "\n\n"))
		}
		_, _ = w.Write([]byte("data: [DONE]\n\n"))
	}))
	defer srv.Close()

	cfg := config.ProviderConfig{
		Name:    "OpenCode",
		Format:  config.FormatOpenAI,
		Variant: "opencode",
		APIKey:  "",
		BaseURL: srv.URL,
		Models:  []config.ModelConfig{{ID: "grok-code", Context: 256000, MaxOutput: 32768}},
	}
	p, err := buildProvider(cfg)
	if err != nil {
		t.Fatal(err)
	}
	model, err := p.LanguageModel(context.Background(), "grok-code")
	if err != nil {
		t.Fatal(err)
	}
	ag := fantasy.NewAgent(model, fantasy.WithMaxRetries(1))
	_, err = ag.Stream(context.Background(), fantasy.AgentStreamCall{
		Messages: []fantasy.Message{{
			Role:    fantasy.MessageRoleUser,
			Content: []fantasy.MessagePart{fantasy.TextPart{Text: "hi"}},
		}},
	})
	if err != nil {
		t.Fatalf("stream: %v", err)
	}

	for _, k := range []string{"User-Agent", "X-Opencode-Client", "X-Opencode-Project", "X-Opencode-Request", "X-Opencode-Session", "X-Opencode-Session-Id", "Authorization"} {
		t.Logf("%-24s = %q", k, got.Get(k))
	}
	t.Logf("model = %v", body["model"])

	ua := got.Get("User-Agent")
	if !strings.HasPrefix(ua, "opencode/") {
		t.Errorf("User-Agent = %q, want prefix opencode/", ua)
	}
	ses := got.Get("X-Opencode-Session")
	if !validSessionID(ses) {
		t.Errorf("session = %q, want ses_ + 12 lowercase hex + 14 base62", ses)
	}
	if got.Get("X-Opencode-Session-Id") != ses {
		t.Errorf("session-id %q != session %q", got.Get("X-Opencode-Session-Id"), ses)
	}
}

func validSessionID(id string) bool {
	const hex = "0123456789abcdef"
	body, ok := strings.CutPrefix(id, "ses_")
	if !ok || len(body) != 26 {
		return false
	}
	for _, c := range body[:12] {
		if !strings.ContainsRune(hex, c) {
			return false
		}
	}
	for _, c := range body[12:] {
		if !strings.ContainsRune(ocBase62Chars, c) {
			return false
		}
	}
	return true
}

const ocBase62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
