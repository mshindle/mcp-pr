package provider

import (
	"errors"
	"fmt"
	"strings"
)

// ErrUnknownProvider is returned when a requested provider name is not registered.
var ErrUnknownProvider = errors.New("unknown provider")

// Registry maps provider name strings to ConstructorFuncs.
type Registry struct {
	constructors map[string]ConstructorFunc
	order        []string // insertion order for default resolution
}

// NewRegistry returns an empty Registry.
func NewRegistry() *Registry {
	return &Registry{
		constructors: make(map[string]ConstructorFunc),
	}
}

// Register adds a provider constructor under the given name.
func (r *Registry) Register(name string, fn ConstructorFunc) {
	if _, exists := r.constructors[name]; !exists {
		r.order = append(r.order, name)
	}
	r.constructors[name] = fn
}

// Resolve creates and returns the named provider with an optional model override.
// Returns ErrUnknownProvider if the name is not registered.
func (r *Registry) Resolve(name, model string) (Provider, error) {
	fn, ok := r.constructors[name]
	if !ok {
		supported := strings.Join(r.order, ", ")
		return nil, fmt.Errorf("%w: %q; supported providers: %s", ErrUnknownProvider, name, supported)
	}
	return fn(model)
}

// DefaultProvider returns the first registered provider whose constructor succeeds.
// This implements the env-var-based resolution order defined at registration time.
func (r *Registry) DefaultProvider(model string) (Provider, error) {
	for _, name := range r.order {
		fn := r.constructors[name]
		p, err := fn(model)
		if err == nil {
			return p, nil
		}
	}
	return nil, errors.New("no AI provider available; set at least one of: ANTHROPIC_API_KEY, OPENAI_API_KEY, GOOGLE_API_KEY")
}
