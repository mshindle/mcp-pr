package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/mshindle/mcp-pr/internal/server"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	s := server.NewServer()

	if err := s.MCP().Run(ctx, &mcp.StdioTransport{}); err != nil {
		slog.Error("server error", "err", err)
		os.Exit(1)
	}
}
