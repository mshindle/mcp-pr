# Research: MCP Code Review Server

**Branch**: `001-mcp-code-review` | **Phase**: 0

---

## Decision: MCP Transport

**Decision**: stdio transport (not SSE)

**Rationale**: The go-sdk/mcp library supports both stdio and SSE. For a local developer
tool invoked by an MCP host (Claude Desktop, VS Code extension, etc.), stdio is the standard
transport — it requires no port management, no authentication, and starts on demand. SSE is
appropriate for remote/multi-client servers, which is explicitly out of scope for v1.

**Alternatives considered**:
- SSE: Required for remote deployment; adds HTTP server, port config, and auth concerns.
  Rejected for v1 per spec assumption (local process only).

---

## Decision: MCP Server Structure with go-sdk

**Decision**: Use `github.com/modelcontextprotocol/go-sdk/mcp` to create a `Server`,
register tools via `server.AddTool(...)`, and run via `server.Run(ctx, mcp.NewStdioTransport())`.

**Rationale**: The go-sdk provides a typed tool registration API where each tool has a
named schema (defined via Go struct tags or a JSON schema) and a handler function. This maps
cleanly to our four review tools. The server handles all MCP protocol framing; we only
implement handlers.

**Key patterns**:
- Tool input structs tagged with `json` for schema generation
- Handler signature: `func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error)`
- Tool result returned as `mcp.NewToolResultText(jsonString)` for structured output

**Alternatives considered**:
- Manual JSON-RPC server: More control but defeats the purpose of using the SDK.

---

## Decision: AI Provider Abstraction

**Decision**: Define a `Provider` interface with a single `Review(ctx, ReviewInput) (*ReviewResult, error)`
method. Each of the three SDKs gets its own implementation in `internal/provider/`.

**Rationale**: The spec requires providers to be interchangeable with only a parameter
change at call time. A single interface with constructor functions per provider
(`NewAnthropicProvider`, `NewOpenAIProvider`, `NewGoogleProvider`) satisfies this with
minimal abstraction overhead. A `Registry` maps provider name strings to constructors,
enabling runtime selection without reflection.

**Provider-model defaults** (when no model specified):
- Anthropic: `claude-sonnet-4-6` (latest Sonnet, cost/quality balance)
- OpenAI: `gpt-4o` (current flagship)
- Google: `gemini-2.0-flash` (fast, cost-effective)

**Default provider selection** (when no provider specified): iterate
`ANTHROPIC_API_KEY` → `OPENAI_API_KEY` → `GOOGLE_API_KEY`; use first available.

**Alternatives considered**:
- Single struct with a switch: Couples provider logic, violates Constitution Principle II.
- Plugin system: Premature abstraction (Principle IV); 3 providers don't justify it.

---

## Decision: Git Diff Extraction

**Decision**: Invoke `git` as a subprocess using `os/exec`. No Go git library.

**Rationale**: The `go-git` library is a significant dependency (~20MB) and its diff output
formatting differs from `git diff` output that AI models are trained on. Subprocess calls
to the system `git` binary produce exactly the unified diff format models expect, and keep
the dependency footprint minimal (Principle IV: Simplicity).

**Commands**:
- Staged: `git -C <repo_path> diff --cached`
- Unstaged: `git -C <repo_path> diff`
- Commit: `git -C <repo_path> show <sha> --format="" --patch`
- The `-C` flag sets the working directory without requiring `chdir`.

**Alternatives considered**:
- `go-git`: Richer API but large dependency, non-standard diff output, violates Simplicity.
- Shell script wrapper: Harder to test, platform-dependent.

---

## Decision: Review Prompt Structure

**Decision**: Single structured prompt per review invocation. The prompt includes:
1. System instruction: role as expert code reviewer, output format specification
2. User message: the diff or code snippet, with optional language/context hint

**Output format requested from model**: JSON conforming to `ReviewResult` schema:
```json
{
  "summary": "string",
  "findings": [
    {
      "severity": "issue|suggestion|praise",
      "file": "string (optional)",
      "lines": "string (optional, e.g. '12-15')",
      "message": "string"
    }
  ]
}
```

Models are instructed to return only valid JSON (no markdown fences). If JSON parsing fails,
the raw text response is surfaced as a single `suggestion` finding with an empty file/lines.

**Alternatives considered**:
- Streaming responses: Adds complexity; not required for <30s target (Principle IV).
- Separate prompts per file: More granular but multiplies API calls and cost.

---

## Decision: Structured Logging

**Decision**: Use Go's standard `log/slog` package (available since Go 1.21) with JSON handler
for structured output. Log to stderr (not stdout, which is reserved for MCP protocol frames).

**Rationale**: `slog` is zero-dependency, structured, and leveled. Writing to stderr keeps
the stdout stream clean for MCP protocol framing. Log level controlled via `LOG_LEVEL`
environment variable (default: `INFO`; set to `DEBUG` for provider interaction logging).

**Alternatives considered**:
- `zerolog` / `zap`: Better performance but extra dependency for a local dev tool.
- Standard `log`: No structured fields, no levels.

---

## Decision: Integration Test Strategy

**Decision**: Integration tests in `tests/integration/` use build tags to gate execution.
Tests tagged `//go:build integration` are skipped in normal `go test ./...` runs and
require explicit `-tags integration` plus real API keys in environment.

**Rationale**: Constitution Principle III prohibits mocking AI provider behavior. Real
provider calls require API keys that won't be present in CI without secrets. Build tags
allow the test suite to pass in keyless environments while enforcing the no-mock constraint
when keys are available.

**Tag**: `//go:build integration`
**Run command**: `go test ./tests/integration/... -tags integration -v`

**Alternatives considered**:
- Skipping via `t.Skip()` on missing env var: Less explicit about intent; still compiles
  and reports a skip rather than clean pass.
- Separate module: Heavyweight for this scope.
