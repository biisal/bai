package variant

import (
	"crypto/rand"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/biisal/bai/internal/config"
)

const OpenCode = "opencode"

// ocIDLength matches opencode's identifier length: 12 hex chars of time
// plus 14 random base62 chars.
const ocIDLength = 26

const ocBase62 = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func init() {
	Register(OpenCode, opencodeFactory)
}

func opencodeFactory(cfg config.ProviderConfig) (*Spec, error) {
	sessions := newSessions()

	return &Spec{
		Headers: []Header{
			{
				Key:   "User-Agent",
				Value: func() string { return "opencode/1.18.34 ai-sdk/provider-utils/4.0.23 runtime/browser" },
			},
			{
				Key:   "x-opencode-request",
				Value: func() string { return ocID("msg", false) },
			},
			{Key: "x-opencode-client", Value: func() string { return "cli" }},
			{Key: "x-opencode-project", Value: func() string { return "global" }},
			{
				Key:   "x-opencode-session",
				Value: func() string { return sessions.get("global") },
			},
			{
				Key:   "x-opencode-session-id",
				Value: func() string { return sessions.get("global") },
			},
		},
		AuthScheme:   "Bearer",
		AuthFallback: "public",
	}, nil
}

// ocID reproduces opencode's identifier format so the free-tier gate
// accepts it: 6 bytes of the timestamp (ms<<12 + counter, optionally
// inverted for descending IDs) as lowercase hex, then 14 base62 chars.
func ocID(prefix string, descending bool) string {
	ts := time.Now().UnixMilli()
	value := ts<<12 + 1
	if descending {
		value = ^value
	}
	value &= (1 << 48) - 1

	var b strings.Builder
	b.Grow(len(prefix) + 1 + ocIDLength)
	b.WriteString(prefix)
	b.WriteByte('_')
	b.WriteString(fmt.Sprintf("%012x", value))

	rnd := make([]byte, ocIDLength-12)
	_, _ = rand.Read(rnd)
	for _, c := range rnd {
		b.WriteByte(ocBase62[int(c)%len(ocBase62)])
	}
	return b.String()
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
			return ocID("ses", true)
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
