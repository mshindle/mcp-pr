# Implementation Plan: MCP Code Review Server

**Branch**: `001-mcp-code-review` | **Date**: 2026-03-24 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/001-mcp-code-review/spec.md`

## Summary

Build a Go-based MCP server exposing four tools (`review_staged`, `review_unstaged`,
`review_commit`, `review_code`) that extract diffs via git subprocess calls and submit them
to interchangeable AI providers (Anthropic, OpenAI, Google) for structured code review.
The server runs via stdio MCP transport, uses the `modelcontextprotocol/go-sdk` for
protocol handling, and returns structured JSON review results.

## Technical Context

**Language/Version**: Go 1.26
**Primary Dependencies**: `modelcontextprotocol/go-sdk` v1.4.1 (MCP), `anthropics/anthropic-sdk-go` v1.27.1, `openai/openai-go/v3` v3.29.0, `google.golang.org/genai` v1.51.0, `log/slog` (stdlib)
**Storage**: N/A (stateless)
**Testing**: `go test ./...` (unit), `go test ./tests/integration/... -tags integration` (integration)
**Target Platform**: Local developer workstation, any OS with Go 1.26 and git in PATH
**Project Type**: MCP server binary
**Performance Goals**: Review completion <30s for diffs ≤200 changed lines on standard internet connection
**Constraints**: Stateless; no persistent storage; git binary required in PATH; at least one AI provider API key required at runtime
**Scale/Scope**: Single-user local tool; one concurrent review per server instance

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Check | Status |
|-----------|-------|--------|
| I. MCP-First | All review functionality exposed exclusively as MCP tools; no HTTP/gRPC added | ✅ PASS |
| II. Multi-Provider AI | `Provider` interface; all three SDKs implement it; runtime selection via parameter | ✅ PASS |
| III. Test-First | Integration tests use real provider APIs (build tag `integration`); unit tests cover git and review logic | ✅ PASS |
| IV. Simplicity | No new dependencies added; flat `internal/` package layout; no unnecessary abstraction layers | ✅ PASS |
| V. Observability | `log/slog` to stderr on all tool invocations; DEBUG level for provider interactions | ✅ PASS |

**Post-design re-check**: All five principles satisfied. No violations requiring complexity
justification.

## Project Structure

### Documentation (this feature)

```text
specs/001-mcp-code-review/
├── plan.md              # This file
├── research.md          # Phase 0 decisions
├── data-model.md        # Phase 1 Go types
├── quickstart.md        # Phase 1 validation guide
├── contracts/
│   └── mcp-tools.md     # Phase 1 tool schemas and error contracts
└── tasks.md             # Phase 2 output (/speckit.tasks — not yet created)
```

### Source Code (repository root)

```text
cmd/
└── mcp-code-review/
    └── main.go             # Entry point: build registry, create server, run stdio transport

internal/
├── provider/
│   ├── provider.go         # Provider interface + Registry type
│   ├── registry.go         # Registry implementation + default resolution logic
│   ├── anthropic.go        # AnthropicProvider implementation
│   ├── openai.go           # OpenAIProvider implementation
│   └── google.go           # GoogleProvider implementation
├── git/
│   └── git.go              # Staged/unstaged/commit diff extraction via os/exec
├── review/
│   ├── types.go            # ReviewInput, ReviewResult, Finding
│   └── prompt.go           # Prompt construction (system instruction + diff formatting)
└── server/
    ├── server.go           # MCP Server setup, tool registration, startup
    └── tools.go            # Tool input structs + handler functions (one per tool)

tests/
├── integration/            # //go:build integration — require real API keys
│   ├── staged_test.go
│   ├── unstaged_test.go
│   ├── commit_test.go
│   └── arbitrary_test.go
└── unit/
    ├── git_test.go         # Subprocess behavior, error cases, binary file handling
    ├── review_test.go      # Prompt construction, ReviewResult parsing, edge cases
    └── registry_test.go    # Provider resolution order, missing key errors
```

**Structure Decision**: Single project layout (`cmd/` + `internal/` + `tests/`). Go
convention for a standalone binary. No `src/` wrapper — idiomatic Go. Three provider
implementations do not justify a plugin system (Principle IV); they live as plain files
in `internal/provider/`.

## Complexity Tracking

> No constitution violations — table not required.
