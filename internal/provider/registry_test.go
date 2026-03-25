package provider

import (
	"context"
	"errors"
	"testing"

	"github.com/mshindle/mcp-pr/internal/review"
)

// stubProvider is a test double that satisfies the provider.Provider interface.
type stubProvider struct {
	name         string
	defaultModel string
}

func (s *stubProvider) Review(_ context.Context, _ review.ReviewInput) (*review.ReviewResult, error) {
	return &review.ReviewResult{Summary: "stub"}, nil
}
func (s *stubProvider) Name() string         { return s.name }
func (s *stubProvider) DefaultModel() string { return s.defaultModel }

func TestRegistry_ResolveRegistered(t *testing.T) {
	r := NewRegistry()
	r.Register("testprov", func(model string) (Provider, error) {
		return &stubProvider{name: "testprov", defaultModel: model}, nil
	})

	p, err := r.Resolve("testprov", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Name() != "testprov" {
		t.Errorf("expected 'testprov', got %s", p.Name())
	}
}

func TestRegistry_ResolveUnknown(t *testing.T) {
	r := NewRegistry()
	_, err := r.Resolve("unknown", "")
	if err == nil {
		t.Error("expected error for unknown provider")
	}
	if !errors.Is(err, ErrUnknownProvider) {
		t.Errorf("expected ErrUnknownProvider, got: %v", err)
	}
}

func TestRegistry_DefaultProvider_FirstAvailable(t *testing.T) {
	r := NewRegistry()
	r.Register("alpha", func(model string) (Provider, error) {
		return &stubProvider{name: "alpha"}, nil
	})
	r.Register("beta", func(model string) (Provider, error) {
		return &stubProvider{name: "beta"}, nil
	})

	// Set alpha as having a key (simulated via the constructor success).
	// DefaultProvider returns the first provider whose constructor succeeds.
	p, err := r.DefaultProvider("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p == nil {
		t.Error("expected a provider, got nil")
	}
}

func TestRegistry_DefaultProvider_NoneAvailable(t *testing.T) {
	r := NewRegistry()
	// Register one that always fails (simulates missing API key).
	r.Register("failing", func(model string) (Provider, error) {
		return nil, errors.New("no key")
	})

	_, err := r.DefaultProvider("")
	if err == nil {
		t.Error("expected error when no providers are available")
	}
}

func TestRegistry_ResolvePassesModelOverride(t *testing.T) {
	r := NewRegistry()
	var gotModel string
	r.Register("myprov", func(model string) (Provider, error) {
		gotModel = model
		return &stubProvider{name: "myprov", defaultModel: model}, nil
	})

	_, err := r.Resolve("myprov", "my-custom-model")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotModel != "my-custom-model" {
		t.Errorf("expected model 'my-custom-model', got %s", gotModel)
	}
}
