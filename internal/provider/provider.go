package provider

import (
	"context"

	"github.com/mshindle/mcp-pr/internal/review"
)

// Provider is the abstraction over AI backends.
type Provider interface {
	// Review submits the given input for review and returns structured findings.
	Review(ctx context.Context, input review.ReviewInput) (*review.ReviewResult, error)

	// Name returns the canonical provider identifier ("anthropic" | "openai" | "google").
	Name() string

	// DefaultModel returns the model used when ReviewInput.Model is empty.
	DefaultModel() string
}

// ConstructorFunc creates a Provider from an optional model override.
type ConstructorFunc func(model string) (Provider, error)
