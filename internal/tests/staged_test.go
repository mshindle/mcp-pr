//go:build integration

package tests_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/mshindle/mcp-pr/internal/git"
	"github.com/mshindle/mcp-pr/internal/provider"
	"github.com/mshindle/mcp-pr/internal/review"
	"github.com/mshindle/mcp-pr/internal/server"
)

func TestReviewStaged_HappyPath(t *testing.T) {
	prov := mustProvider(t)
	dir := initRepoWithStagedChange(t)

	result, err := git.StagedDiff(dir)
	if err != nil {
		t.Fatalf("staged diff: %v", err)
	}
	if result.Diff == "" {
		t.Fatal("expected non-empty staged diff")
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

func TestReviewStaged_NothingStaged(t *testing.T) {
	dir := initEmptyRepo(t)
	result, err := git.StagedDiff(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Diff != "" {
		t.Error("expected empty diff in clean repo")
	}
}

func TestReviewStaged_LargeDiff(t *testing.T) {
	prov := mustProvider(t)
	dir := initRepoWithStagedChange(t)

	// Build a larger diff by adding more content.
	var big []byte
	for i := 0; i < 500; i++ {
		big = append(big, []byte("// line comment\n")...)
	}
	f := filepath.Join(dir, "large.go")
	if err := os.WriteFile(f, append([]byte("package main\n"), big...), 0644); err != nil {
		t.Fatal(err)
	}
	run(t, dir, "git", "add", "large.go")

	result, _ := git.StagedDiff(dir)
	input := review.ReviewInput{Code: result.Diff, Language: "unified-diff"}
	res, err := prov.Review(context.Background(), input)
	if err != nil {
		t.Logf("large diff error (acceptable): %v", err)
		return
	}
	if res.Summary == "" {
		t.Error("expected summary even for large diff")
	}
}

// --- helpers ---

func mustProvider(t *testing.T) provider.Provider {
	t.Helper()
	reg := server.BuildRegistry()
	p, err := reg.DefaultProvider("")
	if err != nil {
		t.Skipf("no provider available: %v", err)
	}
	return p
}

func initEmptyRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	run(t, dir, "git", "init")
	run(t, dir, "git", "config", "user.email", "test@test.com")
	run(t, dir, "git", "config", "user.name", "Test")
	f := filepath.Join(dir, "README.md")
	_ = os.WriteFile(f, []byte("# test\n"), 0644)
	run(t, dir, "git", "add", "README.md")
	run(t, dir, "git", "commit", "-m", "init")
	return dir
}

func initRepoWithStagedChange(t *testing.T) string {
	t.Helper()
	dir := initEmptyRepo(t)
	f := filepath.Join(dir, "main.go")
	_ = os.WriteFile(f, []byte("package main\n\nfunc main() {}\n"), 0644)
	run(t, dir, "git", "add", "main.go")
	return dir
}

func run(t *testing.T, dir string, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%s %v: %v\n%s", name, args, err, out)
	}
}
