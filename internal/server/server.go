package server

import (
	"log/slog"
	"os"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/mshindle/mcp-pr/internal/provider"
)

// Server holds the MCP server and its dependencies.
type Server struct {
	mcp      *mcp.Server
	registry *provider.Registry
}

// BuildRegistry creates and populates a provider Registry from environment variables.
// Registration order defines default provider resolution priority.
func BuildRegistry() *provider.Registry {
	reg := provider.NewRegistry()
	reg.Register("anthropic", func(model string) (provider.Provider, error) {
		return provider.NewAnthropicProvider(model)
	})
	reg.Register("openai", func(model string) (provider.Provider, error) {
		return provider.NewOpenAIProvider(model)
	})
	reg.Register("google", func(model string) (provider.Provider, error) {
		return provider.NewGoogleProvider(model)
	})
	return reg
}

// NewServer creates and configures the MCP server with all review tools registered.
func NewServer() *Server {
	setupLogging()

	reg := BuildRegistry()

	s := &Server{registry: reg}
	s.mcp = mcp.NewServer(&mcp.Implementation{
		Name:    "mcp-code-review",
		Version: "1.0.0",
	}, nil)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "review_staged",
		Description: "Review the staged (index) changes in a git repository using an AI provider.",
	}, s.handleReviewStaged)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "review_unstaged",
		Description: "Review the unstaged (working directory) changes in a git repository using an AI provider.",
	}, s.handleReviewUnstaged)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "review_commit",
		Description: "Review the changes introduced by a specific git commit SHA using an AI provider.",
	}, s.handleReviewCommit)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "review_code",
		Description: "Review an arbitrary code snippet provided as text using an AI provider.",
	}, s.handleReviewCode)

	return s
}

// MCP returns the underlying mcp.Server for transport connection.
func (s *Server) MCP() *mcp.Server {
	return s.mcp
}

// setupLogging initialises a slog handler writing to stderr at the level
// specified by the LOG_LEVEL environment variable (default: INFO).
func setupLogging() {
	level := slog.LevelInfo
	switch strings.ToUpper(os.Getenv("LOG_LEVEL")) {
	case "DEBUG":
		level = slog.LevelDebug
	case "WARN":
		level = slog.LevelWarn
	case "ERROR":
		level = slog.LevelError
	}
	handler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	slog.SetDefault(slog.New(handler))
}
