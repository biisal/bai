package variant

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/biisal/bai/internal/config"
)

// OpenCode layers opencode-specific quirks on top of any provider
// format: branded headers, per-request IDs, a TTL-rotating session
// header, and a "Bearer public" auth fallback when no API key is set.
const OpenCode = "opencode"

func init() {
	Register(OpenCode, opencodeFactory)
}

func opencodeFactory(cfg config.ProviderConfig) (*Spec, error) {
	sessions := newSessions()

	return &Spec{
		Headers: []Header{
			{
				Key:   "User-Agent",
				Value: func() string { return "opencode/1.15.0 ai-sdk/provider-utils/4.0.23 runtime/bun/1.3.13" },
			},
			{
				Key:   "x-opencode-request",
				Value: func() string { return ocID("msg") }, // fresh per request
			},
			{Key: "x-opencode-client", Value: func() string { return "cli" }},
			{Key: "x-opencode-project", Value: func() string { return "global" }},
			{
				Key:   "x-opencode-session",
				Value: func() string { return sessions.get("global") },
			},
		},
		AuthScheme:   "Bearer",
		AuthFallback: "public",
	}, nil
}

// ocID generates an opencode-style ID: hex millisecond timestamp + 16
// chars of random base64url.
func ocID(prefix string) string {
	ts := strconv.FormatInt(time.Now().UnixMilli(), 16)
	rnd := make([]byte, 12)
	_, _ = rand.Read(rnd)
	b64 := base64.RawURLEncoding.EncodeToString(rnd)
	if len(b64) > 16 {
		b64 = b64[:16]
	}
	return fmt.Sprintf("%s_%s%s", prefix, ts, b64)
}

// sessionEntry tracks one session ID and when it was created.
type sessionEntry struct {
	id string
	ts time.Time
}

// sessions issues IDs that rotate after a TTL. State is owned by the
// factory's closures (one instance per provider), never global. The
// clock and ID generator are injectable for tests.
type sessions struct {
	mu   sync.Mutex
	m    map[string]sessionEntry
	now  func() time.Time
	ttl  time.Duration
	next func() string
}

func newSessions() *sessions {
	return &sessions{
		m:   map[string]sessionEntry{},
		now: time.Now,
		ttl: 30 * time.Minute,
		next: func() string {
			return ocID("ses")
		},
	}
}

// get returns the live session for key, rotating it if expired.
func (s *sessions) get(key string) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.now()
	entry, ok := s.m[key]
	if !ok || now.Sub(entry.ts) > s.ttl {
		entry = sessionEntry{id: s.next(), ts: now}
		s.m[key] = entry
	}
	return entry.id
}
