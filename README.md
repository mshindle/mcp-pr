# mcp-code-review

An MCP (Model Context Protocol) server that performs AI-powered code reviews using multiple LLM providers (Anthropic Claude, OpenAI GPT, Google Gemini). Integrates with any MCP-compatible coding agent such as Claude Code.

## Building

Prerequisites: Go 1.23+ and `git` available on `PATH`.

```bash
# Clone the repository
git clone https://github.com/mshindle/mcp-pr.git
cd mcp-pr

# Build the server binary
go build -o mcp-code-review ./cmd/mcp-code-review

# Run tests
go test ./...

# Format and vet
go fmt ./...
go vet ./...
```

## Starting the Server

The server communicates over **stdio** using the MCP protocol. It is not run directly — instead, you register it with your MCP client (e.g. Claude Code) and the client launches it automatically.

### Registering with Claude Code

Add the following to your Claude Code MCP configuration (`.claude/settings.json` or equivalent):

```json
{
  "mcpServers": {
    "mcp-code-review": {
      "command": "/path/to/mcp-code-review",
      "env": {
        "ANTHROPIC_API_KEY": "sk-ant-...",
        "OPENAI_API_KEY": "sk-...",
        "GOOGLE_API_KEY": "AIza..."
      }
    }
  }
}
```

At least one API key must be set. The server will register only providers whose keys are present.

### Running manually (for debugging)

```bash
ANTHROPIC_API_KEY=sk-ant-... ./mcp-code-review
```

The server reads MCP requests from stdin and writes responses to stdout. Structured logs are written to stderr as JSON.

## Environment Variables

### API Keys

At least one key is required. The server registers providers in the order listed below; the first available provider becomes the default.

| Variable | Provider | Required |
|---|---|---|
| `ANTHROPIC_API_KEY` | Anthropic Claude | At least one of these |
| `OPENAI_API_KEY` | OpenAI GPT | At least one of these |
| `GOOGLE_API_KEY` | Google Gemini | At least one of these |

If no keys are set the server starts but every tool call will return an error.

### Model Overrides

Each provider uses a sensible default model. Override these per-provider with the following variables:

| Variable | Provider | Default Model |
|---|---|---|
| `MCP_REVIEW_ANTHROPIC_MODEL` | Anthropic | `claude-sonnet-4-6` |
| `MCP_REVIEW_OPENAI_MODEL` | OpenAI | `gpt-4o` |
| `MCP_REVIEW_GOOGLE_MODEL` | Google | `gemini-2.0-flash` |

Individual tool calls can also pass a `model` parameter to override the model for that specific request.

### Logging

| Variable | Description | Default |
|---|---|---|
| `LOG_LEVEL` | Verbosity of structured JSON logs written to stderr. One of `DEBUG`, `INFO`, `WARN`, `ERROR`. | `INFO` |

Set `LOG_LEVEL=DEBUG` to see every provider request and response during development.

## Available Tools

The server exposes four MCP tools:

### `review_staged`

Reviews changes that are staged (in the git index) but not yet committed.

| Parameter | Type | Required | Description |
|---|---|---|---|
| `repo_path` | string | No | Absolute path to the git repository root. Defaults to `.`. |
| `provider` | string | No | AI provider to use: `anthropic`, `openai`, or `google`. Defaults to the first available. |
| `model` | string | No | Model ID to use, overriding the provider default. |

### `review_unstaged`

Reviews changes in the working directory that are not yet staged.

| Parameter | Type | Required | Description |
|---|---|---|---|
| `repo_path` | string | No | Absolute path to the git repository root. Defaults to `.`. |
| `provider` | string | No | AI provider to use. |
| `model` | string | No | Model ID to use. |

### `review_commit`

Reviews the diff introduced by a specific git commit.

| Parameter | Type | Required | Description |
|---|---|---|---|
| `sha` | string | Yes | Full or abbreviated commit SHA (minimum 7 characters). |
| `repo_path` | string | No | Absolute path to the git repository root. Defaults to `.`. |
| `provider` | string | No | AI provider to use. |
| `model` | string | No | Model ID to use. |

### `review_code`

Reviews an arbitrary code snippet without requiring a git repository.

| Parameter | Type | Required | Description |
|---|---|---|---|
| `code` | string | Yes | The code to review. |
| `language` | string | No | Language hint (e.g. `go`, `python`, `typescript`). |
| `provider` | string | No | AI provider to use. |
| `model` | string | No | Model ID to use. |

### Response Format

All tools return a JSON object:

```json
{
  "summary": "Brief overall assessment of the code",
  "findings": [
    {
      "severity": "issue",
      "file": "internal/server/tools.go",
      "lines": "42-55",
      "message": "SQL query is built via string concatenation — use parameterized queries to prevent injection."
    },
    {
      "severity": "suggestion",
      "file": "internal/server/tools.go",
      "lines": "60",
      "message": "Consider extracting the error message into a named constant for reuse."
    },
    {
      "severity": "praise",
      "message": "Comprehensive input validation with clear, user-facing error messages."
    }
  ]
}
```

Severity values:
- `issue` — bugs, security vulnerabilities, correctness problems
- `suggestion` — style, maintainability, or performance improvements
- `praise` — noteworthy positive aspects

## Example Conversations

Below are example conversations showing how a coding agent (Claude Code) uses this server in practice.

---

### Example 1: Reviewing staged changes before a commit

```
User: Review my staged changes before I commit them.

Agent: I'll review your staged changes now.
[calls review_staged with repo_path="/home/user/myproject"]

The review found 3 findings:

**Summary:** The staged changes add a new user authentication endpoint. Overall structure
is solid, but there is one security issue that should be addressed before committing.

**Findings:**

🔴 Issue — `internal/auth/handler.go` lines 78-82
The password comparison uses `==` on raw strings. Use `bcrypt.CompareHashAndPassword`
instead to prevent timing attacks.

💡 Suggestion — `internal/auth/handler.go` line 45
The 30-second token expiry is hardcoded. Consider making it configurable via an
environment variable or config struct.

✅ Praise
Input validation is thorough and error messages are appropriately generic to avoid
leaking internal details.

Fix the timing attack issue before committing — that's a real security vulnerability.
```

---

### Example 2: Reviewing a specific commit

```
User: Can you review commit a3f8c21? I want to make sure the refactor is clean.

Agent: I'll review that commit for you.
[calls review_commit with sha="a3f8c21", repo_path="/home/user/myproject"]

**Summary:** The refactor cleanly extracts the database layer into its own package.
Two minor issues worth noting.

**Findings:**

🔴 Issue — `internal/db/connection.go` lines 12-15
`db.SetMaxOpenConns` is called but `db.SetMaxIdleConns` is not. Under high load this
can exhaust connections. Set idle conns to a fraction of max open conns.

💡 Suggestion — `internal/db/query.go` line 88
The fallback timeout of 30s is reasonable, but add a comment explaining why that
value was chosen so future maintainers don't change it arbitrarily.

✅ Praise
The package boundary is well-defined with a clean interface — all callers use the
interface type, not the concrete struct. Makes testing straightforward.
```

---

### Example 3: Reviewing a code snippet directly

```
User: Review this Python function I'm about to add:

def get_user(user_id):
    query = "SELECT * FROM users WHERE id = " + str(user_id)
    return db.execute(query).fetchone()

Agent: I'll review that snippet.
[calls review_code with language="python", code="def get_user(user_id):\n    ..."]

**Summary:** The function is dangerously vulnerable to SQL injection and should not
be used as written.

**Findings:**

🔴 Issue — lines 2-3
String concatenation in SQL queries allows SQL injection. Any caller who passes
untrusted input (e.g. from a URL parameter) can read, modify, or delete arbitrary
data. Use parameterized queries:

    query = "SELECT * FROM users WHERE id = %s"
    return db.execute(query, (user_id,)).fetchone()

💡 Suggestion
Return `None` explicitly and document what the caller should expect when no user
is found. `fetchone()` already returns `None`, but making it explicit in a docstring
prevents surprises.

Replace the query construction before merging — this is a critical vulnerability.
```

---

### Example 4: Using a specific provider

```
User: Review my unstaged changes, but use OpenAI instead of the default provider.

Agent: I'll run the review with OpenAI.
[calls review_unstaged with provider="openai", repo_path="/home/user/myproject"]

**Summary:** Minor cleanup changes. No issues found.

**Findings:**

💡 Suggestion — `cmd/server/main.go` line 14
`os.Exit(1)` is called directly in main. Consider logging the error first so the
exit reason appears in the logs when running under a process supervisor.

✅ Praise
The signal handling using `context.WithCancel` and `signal.NotifyContext` is
idiomatic Go and handles both SIGINT and SIGTERM correctly.
```

## Provider Priority

When no `provider` parameter is given, the server uses the first provider for which an API key is configured, checked in this order:

1. Anthropic
2. OpenAI
3. Google

To use a specific provider for all calls, either set only that provider's API key or always pass the `provider` parameter explicitly.
