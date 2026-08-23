// Package variant holds optional decorations layered on top of a
// provider format: branded headers, auth fallbacks, session tracking.
//
// Every variant registers itself from its own file via init(), so
// adding or removing one never touches the rest of the codebase.
// Specs are format-agnostic: providers translate them into their own
// client's request options.
package variant

import (
	"fmt"
	"slices"
	"sync"

	"github.com/biisal/bai/internal/config"
)

// Header is a single request header. Value is evaluated once per HTTP
// request, so dynamic values (per-request IDs, rotating sessions) work.
type Header struct {
	Key   string
	Value func() string
}

// Spec is a format-agnostic variant definition.
type Spec struct {
	Name    string
	Headers []Header

	// AuthScheme + AuthFallback form the Authorization value used only
	// when the user configured no API key (e.g. "Bearer public").
	// Empty AuthFallback means: never touch Authorization.
	AuthScheme   string
	AuthFallback string
}

// Factory builds a Spec from provider config. Variants own any state
// they need (e.g. session stores) inside the returned closures.
type Factory func(cfg config.ProviderConfig) (*Spec, error)

var (
	mu        sync.Mutex
	factories = map[string]Factory{}
)

// Register adds a variant factory. Panics on duplicates: a name
// collision is a programming error, best caught at startup.
func Register(name string, f Factory) {
	mu.Lock()
	defer mu.Unlock()
	if _, dup := factories[name]; dup {
		panic(fmt.Sprintf("variant: duplicate registration of %q", name))
	}
	factories[name] = f
}

// Get returns the factory registered under name.
func Get(name string) (Factory, bool) {
	mu.Lock()
	defer mu.Unlock()
	f, ok := factories[name]
	return f, ok
}

// Names lists registered variants, sorted for stable output.
func Names() []string {
	mu.Lock()
	defer mu.Unlock()
	names := make([]string, 0, len(factories))
	for name := range factories {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}
