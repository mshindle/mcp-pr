//go:build integration

package tests_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mshindle/mcp-pr/internal/git"
	"github.com/mshindle/mcp-pr/internal/review"
)

func TestReviewUnstaged_HappyPath(t *testing.T) {
	prov := mustProvider(t)
	dir := initEmptyRepo(t)

	// Modify a tracked file without staging.
	f := filepath.Join(dir, "README.md")
	if err := os.WriteFile(f, []byte("# updated\n"), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := git.UnstagedDiff(dir)
	if err != nil {
		t.Fatalf("unstaged diff: %v", err)
	}
	if result.Diff == "" {
		t.Fatal("expected non-empty unstaged diff")
	}

	input := review.ReviewInput{
		Code:     result.Diff,
		Language: "unified-diff",
		Provider: prov.Name(),
	}
	res, err := prov.Review(context.Background(), input)
	if err != nil {
		t.Fatalf("review: %v", err)
	}
	if res.Summary == "" {
		t.Error("expected non-empty summary")
	}
}

func TestReviewUnstaged_NothingUnstaged(t *testing.T) {
	dir := initEmptyRepo(t)
	result, err := git.UnstagedDiff(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Diff != "" {
		t.Error("expected empty diff in clean repo")
	}
}
