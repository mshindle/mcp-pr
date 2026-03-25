package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// initTestRepo creates a temporary git repository for testing.
func initTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}

	run("init")
	run("config", "user.email", "test@test.com")
	run("config", "user.name", "Test")

	// Create an initial commit so HEAD exists.
	f := filepath.Join(dir, "README.md")
	if err := os.WriteFile(f, []byte("# test\n"), 0644); err != nil {
		t.Fatal(err)
	}
	run("add", "README.md")
	run("commit", "-m", "initial commit")

	return dir
}

func TestStagedDiff_Empty(t *testing.T) {
	dir := initTestRepo(t)
	result, err := StagedDiff(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Diff != "" {
		t.Errorf("expected empty diff, got: %s", result.Diff)
	}
}

func TestStagedDiff_WithChanges(t *testing.T) {
	dir := initTestRepo(t)

	f := filepath.Join(dir, "main.go")
	if err := os.WriteFile(f, []byte("package main\n"), 0644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("git", "add", "main.go")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git add: %v\n%s", err, out)
	}

	result, err := StagedDiff(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Diff == "" {
		t.Error("expected non-empty diff")
	}
	if !strings.Contains(result.Diff, "main.go") {
		t.Errorf("diff does not mention main.go: %s", result.Diff)
	}
}

func TestUnstagedDiff_Empty(t *testing.T) {
	dir := initTestRepo(t)
	result, err := UnstagedDiff(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Diff != "" {
		t.Errorf("expected empty diff, got: %s", result.Diff)
	}
}

func TestUnstagedDiff_WithChanges(t *testing.T) {
	dir := initTestRepo(t)

	f := filepath.Join(dir, "README.md")
	if err := os.WriteFile(f, []byte("# updated\n"), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := UnstagedDiff(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Diff == "" {
		t.Error("expected non-empty diff")
	}
}

func TestCommitDiff_ValidSHA(t *testing.T) {
	dir := initTestRepo(t)

	// Get the initial commit SHA.
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("rev-parse HEAD: %v", err)
	}
	sha := strings.TrimSpace(string(out))

	result, err := CommitDiff(dir, sha)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Diff == "" {
		t.Error("expected non-empty diff for initial commit")
	}
}

func TestCommitDiff_InvalidSHA(t *testing.T) {
	dir := initTestRepo(t)
	_, err := CommitDiff(dir, "deadbeefdeadbeef")
	if err == nil {
		t.Error("expected error for invalid SHA")
	}
}

func TestCommitDiff_NotARepo(t *testing.T) {
	dir := t.TempDir()
	_, err := CommitDiff(dir, "HEAD")
	if err == nil {
		t.Error("expected error for non-repo directory")
	}
}

func TestStagedDiff_NotARepo(t *testing.T) {
	dir := t.TempDir()
	_, err := StagedDiff(dir)
	if err == nil {
		t.Error("expected error for non-repo directory")
	}
}
