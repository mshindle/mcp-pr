package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/mshindle/mcp-pr/internal/git"
	"github.com/mshindle/mcp-pr/internal/provider"
	"github.com/mshindle/mcp-pr/internal/review"
)

// ReviewStagedInput defines the input schema for the review_staged tool.
type ReviewStagedInput struct {
	RepoPath string `json:"repo_path,omitempty"`
	Provider string `json:"provider,omitempty"`
	Model    string `json:"model,omitempty"`
}

// ReviewUnstagedInput defines the input schema for the review_unstaged tool.
type ReviewUnstagedInput struct {
	RepoPath string `json:"repo_path,omitempty"`
	Provider string `json:"provider,omitempty"`
	Model    string `json:"model,omitempty"`
}

// ReviewCommitInput defines the input schema for the review_commit tool.
type ReviewCommitInput struct {
	SHA      string `json:"sha"`
	RepoPath string `json:"repo_path,omitempty"`
	Provider string `json:"provider,omitempty"`
	Model    string `json:"model,omitempty"`
}

// ReviewCodeInput defines the input schema for the review_code tool.
type ReviewCodeInput struct {
	Code     string `json:"code"`
	Language string `json:"language,omitempty"`
	Provider string `json:"provider,omitempty"`
	Model    string `json:"model,omitempty"`
}

// handleReviewStaged handles the review_staged tool.
func (s *Server) handleReviewStaged(ctx context.Context, _ *mcp.CallToolRequest, args ReviewStagedInput) (*mcp.CallToolResult, any, error) {
	repoPath := resolveRepoPath(args.RepoPath)
	slog.InfoContext(ctx, "review_staged invoked", "repo_path", repoPath, "provider", args.Provider)

	result, err := git.StagedDiff(repoPath)
	if err != nil {
		return toolError(err.Error()), nil, nil
	}
	if result.Diff == "" {
		return toolError("no staged changes to review"), nil, nil
	}

	prov, err := s.resolveProvider(args.Provider, args.Model)
	if err != nil {
		return toolError(err.Error()), nil, nil
	}

	slog.DebugContext(ctx, "calling provider", "provider", prov.Name(), "model", prov.DefaultModel())

	input := review.ReviewInput{
		Code:     result.Diff,
		Language: "unified-diff",
		Provider: prov.Name(),
		Model:    args.Model,
	}
	if result.IsBinary {
		input.Context = "(note: binary files were skipped)"
	}

	res, err := prov.Review(ctx, input)
	if err != nil {
		return toolError(err.Error()), nil, nil
	}
	if result.IsBinary {
		res.Summary = "[Binary files in diff were skipped] " + res.Summary
	}

	return toolResult(res), nil, nil
}

// handleReviewUnstaged handles the review_unstaged tool.
func (s *Server) handleReviewUnstaged(ctx context.Context, req *mcp.CallToolRequest, args ReviewUnstagedInput) (*mcp.CallToolResult, any, error) {
	repoPath := resolveRepoPath(args.RepoPath)
	slog.InfoContext(ctx, "review_unstaged invoked", "repo_path", repoPath, "provider", args.Provider)

	result, err := git.UnstagedDiff(repoPath)
	if err != nil {
		return toolError(err.Error()), nil, nil
	}
	if result.Diff == "" {
		return toolError("no unstaged changes to review"), nil, nil
	}

	prov, err := s.resolveProvider(args.Provider, args.Model)
	if err != nil {
		return toolError(err.Error()), nil, nil
	}

	slog.DebugContext(ctx, "calling provider", "provider", prov.Name())

	input := review.ReviewInput{
		Code:     result.Diff,
		Language: "unified-diff",
		Provider: prov.Name(),
		Model:    args.Model,
	}
	if result.IsBinary {
		input.Context = "(note: binary files were skipped)"
	}

	res, err := prov.Review(ctx, input)
	if err != nil {
		return toolError(err.Error()), nil, nil
	}
	if result.IsBinary {
		res.Summary = "[Binary files in diff were skipped] " + res.Summary
	}

	return toolResult(res), nil, nil
}

// handleReviewCommit handles the review_commit tool.
func (s *Server) handleReviewCommit(ctx context.Context, req *mcp.CallToolRequest, args ReviewCommitInput) (*mcp.CallToolResult, any, error) {
	if args.SHA == "" {
		return toolError("'sha' is required for review_commit"), nil, nil
	}
	if len(args.SHA) < 7 {
		return toolError(fmt.Sprintf("abbreviated SHA '%s' is ambiguous; provide a longer prefix", args.SHA)), nil, nil
	}

	repoPath := resolveRepoPath(args.RepoPath)
	slog.InfoContext(ctx, "review_commit invoked", "sha", args.SHA, "repo_path", repoPath, "provider", args.Provider)

	result, err := git.CommitDiff(repoPath, args.SHA)
	if err != nil {
		if isCommitNotFound(err) {
			return toolError(fmt.Sprintf("commit '%s' not found in repository", args.SHA)), nil, nil
		}
		return toolError(err.Error()), nil, nil
	}
	if result.Diff == "" {
		return toolError(fmt.Sprintf("commit '%s' introduced no file changes", args.SHA)), nil, nil
	}

	prov, err := s.resolveProvider(args.Provider, args.Model)
	if err != nil {
		return toolError(err.Error()), nil, nil
	}

	slog.DebugContext(ctx, "calling provider", "provider", prov.Name())

	input := review.ReviewInput{
		Code:     result.Diff,
		Language: "unified-diff",
		Context:  fmt.Sprintf("Reviewing commit: %s", args.SHA),
		Provider: prov.Name(),
		Model:    args.Model,
	}

	res, err := prov.Review(ctx, input)
	if err != nil {
		return toolError(err.Error()), nil, nil
	}

	return toolResult(res), nil, nil
}

// handleReviewCode handles the review_code tool.
func (s *Server) handleReviewCode(ctx context.Context, req *mcp.CallToolRequest, args ReviewCodeInput) (*mcp.CallToolResult, any, error) {
	if args.Code == "" {
		return toolError("'code' is required and must not be empty"), nil, nil
	}

	slog.InfoContext(ctx, "review_code invoked", "language", args.Language, "provider", args.Provider)

	prov, err := s.resolveProvider(args.Provider, args.Model)
	if err != nil {
		return toolError(err.Error()), nil, nil
	}

	slog.DebugContext(ctx, "calling provider", "provider", prov.Name())

	input := review.ReviewInput{
		Code:     args.Code,
		Language: args.Language,
		Provider: prov.Name(),
		Model:    args.Model,
	}

	res, err := prov.Review(ctx, input)
	if err != nil {
		return toolError(err.Error()), nil, nil
	}

	return toolResult(res), nil, nil
}

// resolveProvider selects and instantiates the requested provider (or the default).
func (s *Server) resolveProvider(name, model string) (provider.Provider, error) {
	if name == "" {
		p, err := s.registry.DefaultProvider(model)
		if err != nil {
			return nil, errors.New("no API key found; set at least one of: ANTHROPIC_API_KEY, OPENAI_API_KEY, GOOGLE_API_KEY")
		}
		return p, nil
	}
	p, err := s.registry.Resolve(name, model)
	if err != nil {
		if errors.Is(err, provider.ErrUnknownProvider) {
			return nil, fmt.Errorf("unsupported provider %q; supported providers: anthropic, openai, google", name)
		}
		return nil, err
	}
	return p, nil
}

// toolResult marshals a ReviewResult into an MCP tool result.
func toolResult(r *review.ReviewResult) *mcp.CallToolResult {
	data, _ := json.Marshal(r)
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(data)},
		},
	}
}

// toolError returns an MCP tool result containing an error message as plain text.
func toolError(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: msg},
		},
		IsError: true,
	}
}

// resolveRepoPath returns repoPath if non-empty, otherwise the current working directory.
func resolveRepoPath(repoPath string) string {
	if repoPath != "" {
		return repoPath
	}
	// Default to CWD; ignore errors (git will report if not a repo).
	return "."
}

// isCommitNotFound returns true when the error message indicates an unknown commit.
func isCommitNotFound(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return containsAny(msg, "not found in repository", "bad object", "unknown revision", "ambiguous argument")
}

func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if len(sub) > 0 && len(s) >= len(sub) {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
		}
	}
	return false
}
