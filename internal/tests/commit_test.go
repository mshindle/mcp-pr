//go:build integration

package tests_test

import (
	"context"
	"os/exec"
	"strings"
	"testing"

	"github.com/mshindle/mcp-pr/internal/git"
	"github.com/mshindle/mcp-pr/internal/review"
)

func TestReviewCommit_ValidSHA(t *testing.T) {
	prov := mustProvider(t)
	dir := initEmptyRepo(t)
	sha := headSHA(t, dir)

	result, err := git.CommitDiff(dir, sha)
	if err != nil {
		t.Fatalf("commit diff: %v", err)
	}
	if result.Diff == "" {
		t.Fatal("expected non-empty commit diff")
	}

	input := review.ReviewInput{
		Code:     result.Diff,
		Language: "unified-diff",
		Context:  "Reviewing commit: " + sha,
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

func TestReviewCommit_AbbreviatedSHA(t *testing.T) {
	prov := mustProvider(t)
	dir := initEmptyRepo(t)
	sha := headSHA(t, dir)[:7]

	result, err := git.CommitDiff(dir, sha)
	if err != nil {
		t.Fatalf("commit diff with abbreviated SHA: %v", err)
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

func TestReviewCommit_InvalidSHA(t *testing.T) {
	dir := initEmptyRepo(t)
	_, err := git.CommitDiff(dir, "deadbeefdeadbeef")
	if err == nil {
		t.Error("expected error for invalid SHA")
	}
}

func headSHA(t *testing.T, dir string) string {
	t.Helper()
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("rev-parse HEAD: %v", err)
	}
	return strings.TrimSpace(string(out))
}
