package variant

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/biisal/bai/internal/config"
)

func TestSessionsGet(t *testing.T) {
	now := time.Unix(1_000_000, 0)
	ids := []string{"ses_aaa", "ses_bbb"}
	i := 0

	s := newSessions()
	s.now = func() time.Time { return now }
	s.next = func() string {
		id := ids[i%len(ids)]
		i++
		return id
	}

	got := s.get("global")
	if got != "ses_aaa" {
		t.Fatalf("first get() = %q, want %q", got, "ses_aaa")
	}

	// Within TTL: same session returned, no rotation.
	now = now.Add(s.ttl - time.Second)
	if got := s.get("global"); got != "ses_aaa" {
		t.Errorf("get() within TTL = %q, want %q", got, "ses_aaa")
	}

	// Past TTL: rotated.
	now = now.Add(2 * time.Second)
	if got := s.get("global"); got != "ses_bbb" {
		t.Errorf("get() after TTL = %q, want %q", got, "ses_bbb")
	}
}

func TestSessionsKeysAreIndependent(t *testing.T) {
	s := newSessions()

	a := s.get("proj-a")
	b := s.get("proj-b")
	if a == b {
		t.Errorf("different keys share a session: %q", a)
	}
	if a != s.get("proj-a") {
		t.Errorf("same key did not reuse session")
	}
}

func TestOcIDFormat(t *testing.T) {
	id := ocID("msg", false)

	// opencode IDs: prefix + "_" + 12 lowercase hex time + 14 base62 chars.
	body := strings.TrimPrefix(id, "msg_")
	if len(body) != 26 {
		t.Fatalf("ocID() body length = %d, want 26: %q", len(body), id)
	}
	if _, err := strconv.ParseUint(body[:12], 16, 64); err != nil {
		t.Errorf("ocID() time = %q, want 12 lowercase hex chars", body[:12])
	}
	for _, c := range body[12:] {
		if !strings.ContainsRune(ocBase62, c) {
			t.Errorf("ocID() suffix contains non-base62 char %q in %q", c, id)
		}
	}

	// Two calls should practically never collide.
	first, second := ocID("msg", false), ocID("msg", false)
	if first == second {
		t.Errorf("ocID() produced identical IDs twice: %q", first)
	}
}

func TestOcIDDescendingSession(t *testing.T) {
	id := ocID("ses", true)
	body := strings.TrimPrefix(id, "ses_")
	if len(body) != 26 {
		t.Fatalf("ocID() body length = %d, want 26: %q", len(body), id)
	}
	// Descending IDs are bitwise-inverted, so a fresh one starts with a
	// high hex nibble and sorts lexicographically above ascending IDs.
	if body[0] < '8' {
		t.Errorf("ocID(ses, descending) = %q, want inverted time prefix", id)
	}
}

func TestOpenCodeHeadersMatchCapturedRequest(t *testing.T) {
	f, _ := Get(OpenCode)
	spec, err := f(config.ProviderConfig{})
	if err != nil {
		t.Fatalf("factory error: %v", err)
	}

	got := map[string]string{}
	for _, h := range spec.Headers {
		got[h.Key] = h.Value()
	}

	want := map[string]string{
		"User-Agent":            "opencode/1.18.34 ai-sdk/provider-utils/4.0.23 runtime/browser",
		"x-opencode-client":     "cli",
		"x-opencode-project":    "global",
		"x-opencode-session":    "",
		"x-opencode-session-id": "",
		"x-opencode-request":    "",
	}
	for k, v := range want {
		gv, ok := got[k]
		if !ok {
			t.Errorf("missing header %q", k)
			continue
		}
		if v != "" && gv != v {
			t.Errorf("%s = %q, want %q", k, gv, v)
		}
	}

	// session and session-id must be the same rotating value.
	if got["x-opencode-session"] != got["x-opencode-session-id"] {
		t.Errorf("session %q != session-id %q", got["x-opencode-session"], got["x-opencode-session-id"])
	}
	if !strings.HasPrefix(got["x-opencode-session"], "ses_") {
		t.Errorf("session = %q, want ses_ prefix", got["x-opencode-session"])
	}
	if !strings.HasPrefix(got["x-opencode-request"], "msg_") {
		t.Errorf("request = %q, want msg_ prefix", got["x-opencode-request"])
	}
	if spec.AuthScheme != "Bearer" || spec.AuthFallback != "public" {
		t.Errorf("auth = %q %q, want Bearer public", spec.AuthScheme, spec.AuthFallback)
	}
}

func TestRegistry(t *testing.T) {
	f, ok := Get(OpenCode)
	if !ok {
		t.Fatalf("Get(%q) not registered", OpenCode)
	}

	spec, err := f(config.ProviderConfig{})
	if err != nil {
		t.Fatalf("factory error: %v", err)
	}
	if spec.Name != OpenCode && spec.AuthFallback != "public" {
		t.Errorf("unexpected spec: %+v", spec)
	}

	if _, ok := Get("does-not-exist"); ok {
		t.Error("Get() returned factory for unknown variant")
	}

	names := Names()
	found := false
	for _, n := range names {
		if n == OpenCode {
			found = true
		}
	}
	if !found {
		t.Errorf("Names() = %v, missing %q", names, OpenCode)
	}
}
