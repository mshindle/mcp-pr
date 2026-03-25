# Quickstart: MCP Code Review Server

**Branch**: `001-mcp-code-review` | **Phase**: 1

This document describes how to build, configure, and validate the MCP code review server
end-to-end. Use it as the acceptance checklist after implementation.

---

## Prerequisites

- Go 1.26+
- `git` in PATH
- At least one AI provider API key

---

## Build

```bash
go build -o mcp-code-review ./cmd/mcp-code-review
```

Verify the binary starts without error:

```bash
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1"}}}' \
  | ./mcp-code-review
```

Expected: JSON response with `"result"` containing `"protocolVersion"` and `"capabilities"`.

---

## Configuration

Set at least one API key in your environment:

```bash
export ANTHROPIC_API_KEY="sk-ant-..."   # Anthropic
export OPENAI_API_KEY="sk-..."          # OpenAI
export GOOGLE_API_KEY="AIza..."         # Google
```

Optional overrides:

```bash
export LOG_LEVEL=DEBUG                        # Verbose provider logging
export MCP_REVIEW_ANTHROPIC_MODEL="claude-opus-4-6"  # Override model
```

---

## MCP Host Configuration

### Claude Desktop

Add to `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "code-review": {
      "command": "/path/to/mcp-code-review",
      "env": {
        "ANTHROPIC_API_KEY": "sk-ant-..."
      }
    }
  }
}
```

### VS Code / other MCP hosts

Refer to your host's MCP server configuration documentation. The server reads from stdin
and writes to stdout using the MCP stdio transport protocol.

---

## Validation Scenarios

### 1. Review staged changes (happy path)

```bash
cd /your/go/project
echo "// TODO: fix this" >> main.go
git add main.go

# Via MCP (use your host's tool invocation UI or a test harness)
# Tool: review_staged
# Input: {} (no params needed)
```

**Expected**: ReviewResult JSON with `summary` and at least one `finding`.

### 2. Review staged changes — nothing staged

```bash
cd /clean/repo   # no staged files
# Tool: review_staged, Input: {}
```

**Expected**: Error message `"no staged changes to review"`.

### 3. Review a commit

```bash
# Tool: review_commit
# Input: {"sha": "a5b6cbf", "repo_path": "/path/to/mcp-pr"}
```

**Expected**: ReviewResult for the initial commit.

### 4. Review arbitrary code

```bash
# Tool: review_code
# Input: {"code": "func add(a, b int) int { return a - b }", "language": "go"}
```

**Expected**: ReviewResult with a finding noting the subtraction bug.

### 5. Switch providers

```bash
# Tool: review_code
# Input: {"code": "SELECT * FROM users", "language": "sql", "provider": "openai"}
```

**Expected**: ReviewResult from OpenAI (verify via LOG_LEVEL=DEBUG logs on stderr).

### 6. Invalid provider

```bash
# Tool: review_code
# Input: {"code": "x = 1", "provider": "groq"}
```

**Expected**: Error message listing supported providers.

---

## Running Tests

```bash
# Unit tests (no API keys required)
go test ./...

# Integration tests (requires API keys in environment)
go test ./tests/integration/... -tags integration -v
```

All unit tests MUST pass with no API keys present.
Integration tests MUST pass when at least one provider key is set.

---

## Linting

```bash
go vet ./...
go fmt ./...
```

Both MUST complete with no output before any PR merge.
