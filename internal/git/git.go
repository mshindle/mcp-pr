package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// DiffResult is the internal result of a git diff operation.
type DiffResult struct {
	// Diff is the raw unified diff text.
	Diff string

	// Files is the list of files touched (parsed from diff headers).
	Files []string

	// IsBinary reports whether any binary files were encountered (and skipped).
	IsBinary bool
}

// StagedDiff returns the staged (index) diff for the repository at repoPath.
// Returns an empty DiffResult (no error) when nothing is staged.
func StagedDiff(repoPath string) (*DiffResult, error) {
	return runDiff(repoPath, "diff", "--cached")
}

// UnstagedDiff returns the unstaged (working-directory) diff for the repository at repoPath.
// Returns an empty DiffResult (no error) when nothing is modified.
func UnstagedDiff(repoPath string) (*DiffResult, error) {
	return runDiff(repoPath, "diff")
}

// CommitDiff returns the diff introduced by the given commit SHA.
func CommitDiff(repoPath, sha string) (*DiffResult, error) {
	return runDiff(repoPath, "show", sha, "--format=", "--patch")
}

// runDiff executes a git command and parses the unified diff output.
func runDiff(repoPath string, args ...string) (*DiffResult, error) {
	// Prepend -C <repoPath> so git operates in the correct directory.
	fullArgs := append([]string{"-C", repoPath}, args...)
	cmd := exec.Command("git", fullArgs...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		// Detect "not a git repository" and other known errors.
		if strings.Contains(msg, "not a git repository") {
			return nil, fmt.Errorf("not a git repository: %s", repoPath)
		}
		if strings.Contains(msg, "bad object") || strings.Contains(msg, "unknown revision") ||
			strings.Contains(msg, "ambiguous argument") {
			return nil, fmt.Errorf("commit '%s' not found in repository", extractSHA(args))
		}
		// git binary not found.
		if isNotFound(err) {
			return nil, fmt.Errorf("git executable not found in PATH")
		}
		return nil, fmt.Errorf("git error: %s", msg)
	}

	raw := stdout.String()
	return parseDiff(raw), nil
}

// parseDiff converts raw unified diff text into a DiffResult.
func parseDiff(raw string) *DiffResult {
	result := &DiffResult{}

	var filtered strings.Builder
	lines := strings.Split(raw, "\n")
	seenFiles := map[string]bool{}

	for _, line := range lines {
		if strings.HasPrefix(line, "diff --git ") {
			// Extract filename from "diff --git a/foo b/foo"
			parts := strings.Fields(line)
			if len(parts) >= 4 {
				file := strings.TrimPrefix(parts[3], "b/")
				if !seenFiles[file] {
					seenFiles[file] = true
					result.Files = append(result.Files, file)
				}
			}
			filtered.WriteString(line + "\n")
			continue
		}
		if strings.HasPrefix(line, "Binary files") {
			result.IsBinary = true
			// Skip the binary file lines from the diff text.
			continue
		}
		filtered.WriteString(line + "\n")
	}

	result.Diff = strings.TrimRight(filtered.String(), "\n")
	return result
}

// extractSHA attempts to pull a SHA argument from a git args slice.
func extractSHA(args []string) string {
	for _, a := range args {
		if len(a) >= 7 && !strings.HasPrefix(a, "-") && a != "--format=" && a != "--patch" {
			return a
		}
	}
	return "unknown"
}

// isNotFound returns true if the error indicates the executable was not found.
func isNotFound(err error) bool {
	if err == nil {
		return false
	}
	var exitErr *exec.ExitError
	if ok := (err == exec.ErrNotFound); ok {
		return true
	}
	_ = exitErr
	return strings.Contains(err.Error(), "executable file not found")
}
