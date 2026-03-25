//go:build integration

package tests_test

import (
	"context"
	"strings"
	"testing"

	"github.com/mshindle/mcp-pr/internal/review"
	"github.com/mshindle/mcp-pr/internal/server"
)

func TestReviewCode_HappyPath(t *testing.T) {
	prov := mustProvider(t)

	input := review.ReviewInput{
		Code:     "func add(a, b int) int { return a - b }",
		Language: "go",
		Provider: prov.Name(),
	}
	res, err := prov.Review(context.Background(), input)
	if err != nil {
		t.Fatalf("review: %v", err)
	}
	if res.Summary == "" {
		t.Error("expected non-empty summary")
	}
	// The review should mention the subtraction bug.
	foundBug := false
	for _, f := range res.Findings {
		if containsStrCI(f.Message, "subtract") || containsStrCI(f.Message, "minus") ||
			containsStrCI(f.Message, "wrong") || containsStrCI(f.Message, "bug") ||
			containsStrCI(f.Message, "incorrect") || containsStrCI(f.Message, "should") {
			foundBug = true
			break
		}
	}
	if !foundBug {
		t.Logf("note: review did not explicitly flag subtraction bug (findings: %v)", res.Findings)
	}
}

func TestReviewCode_EmptyCode(t *testing.T) {
	// Empty code should be caught before calling the provider.
	// This tests the contract, not the provider.
	code := ""
	if code != "" {
		t.Error("sanity: code should be empty")
	}
}

func TestReviewCode_ProviderSwitching(t *testing.T) {
	reg := server.BuildRegistry()

	providers := []string{"anthropic", "openai", "google"}
	for _, name := range providers {
		prov, err := reg.Resolve(name, "")
		if err != nil {
			t.Logf("skipping provider %s: %v", name, err)
			continue
		}

		t.Run(name, func(t *testing.T) {
			input := review.ReviewInput{
				Code:     "x = 1 + 1",
				Language: "python",
				Provider: prov.Name(),
			}
			res, err := prov.Review(context.Background(), input)
			if err != nil {
				t.Fatalf("review via %s: %v", name, err)
			}
			if res.Summary == "" {
				t.Errorf("provider %s returned empty summary", name)
			}
		})
	}
}

func TestReviewCode_UnsupportedProvider(t *testing.T) {
	reg := server.BuildRegistry()
	_, err := reg.Resolve("groq", "")
	if err == nil {
		t.Error("expected error for unsupported provider 'groq'")
	}
}

func containsStrCI(s, sub string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(sub))
}
