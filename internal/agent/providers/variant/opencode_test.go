package variant

import (
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
	id := ocID("msg")

	if len(id) <= len("msg_") {
		t.Fatalf("ocID() too short: %q", id)
	}
	if id[:4] != "msg_" {
		t.Errorf("ocID() prefix = %q, want %q", id[:4], "msg_")
	}

	// Two calls should practically never collide.
	if ocID("msg") == ocID("msg") {
		t.Error("ocID() produced identical IDs twice")
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
