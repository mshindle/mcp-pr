# MCP Tool Contracts: Code Review Server

**Branch**: `001-mcp-code-review` | **Phase**: 1

All tools return a JSON-encoded `ReviewResult` as MCP text content.
All input parameters are optional unless marked **required**.

---

## Tool: `review_staged`

Review the staged (index) changes in a git repository.

### Input Schema

| Parameter   | Type   | Required | Description |
|-------------|--------|----------|-------------|
| `repo_path` | string | No       | Absolute path to git repo root. Defaults to server CWD. |
| `provider`  | string | No       | AI provider: `anthropic` \| `openai` \| `google`. Defaults to first available API key. |
| `model`     | string | No       | Model identifier. Defaults to provider's recommended model. |

### Output (success)

```json
{
  "summary": "Overall assessment of the staged changes.",
  "findings": [
    {
      "severity": "issue",
      "file": "internal/server/tools.go",
      "lines": "42-47",
      "message": "Error from provider.Review is ignored; propagate it."
    },
    {
      "severity": "suggestion",
      "file": "internal/git/git.go",
      "lines": "18",
      "message": "Consider adding a timeout to the exec.Command context."
    }
  ]
}
```

### Error responses

| Condition | Error message |
|-----------|---------------|
| Not a git repository | `"not a git repository: <repo_path>"` |
| No staged changes | `"no staged changes to review"` |
| git binary not found | `"git executable not found in PATH"` |
| Provider key missing | `"no API key found for provider 'anthropic'; set ANTHROPIC_API_KEY"` |
| Provider call failed | `"provider 'anthropic' returned error: <upstream message>"` |
| Diff exceeds context | `"diff exceeds model context limit (<N> tokens); reduce staged changes or split the commit"` |

---

## Tool: `review_unstaged`

Review the unstaged (working directory) changes in a git repository.

### Input Schema

| Parameter   | Type   | Required | Description |
|-------------|--------|----------|-------------|
| `repo_path` | string | No       | Absolute path to git repo root. Defaults to server CWD. |
| `provider`  | string | No       | AI provider. Defaults to first available API key. |
| `model`     | string | No       | Model identifier. Defaults to provider's recommended model. |

### Output (success)

Same `ReviewResult` JSON structure as `review_staged`.

### Error responses

| Condition | Error message |
|-----------|---------------|
| Not a git repository | `"not a git repository: <repo_path>"` |
| No unstaged changes | `"no unstaged changes to review"` |
| git binary not found | `"git executable not found in PATH"` |
| Provider call failed | `"provider '<name>' returned error: <upstream message>"` |

---

## Tool: `review_commit`

Review the changes introduced by a specific git commit.

### Input Schema

| Parameter   | Type   | Required | Description |
|-------------|--------|----------|-------------|
| `sha`       | string | **Yes**  | Commit SHA: full (40 chars) or abbreviated (≥7 chars). |
| `repo_path` | string | No       | Absolute path to git repo root. Defaults to server CWD. |
| `provider`  | string | No       | AI provider. Defaults to first available API key. |
| `model`     | string | No       | Model identifier. Defaults to provider's recommended model. |

### Output (success)

Same `ReviewResult` JSON structure as `review_staged`, with the `context` field in
the review prompt populated with the commit message.

### Error responses

| Condition | Error message |
|-----------|---------------|
| `sha` missing or empty | `"'sha' is required for review_commit"` |
| Commit not found | `"commit '<sha>' not found in repository"` |
| Ambiguous abbreviated SHA | `"abbreviated SHA '<sha>' is ambiguous; provide a longer prefix"` |
| Not a git repository | `"not a git repository: <repo_path>"` |
| git binary not found | `"git executable not found in PATH"` |
| Provider call failed | `"provider '<name>' returned error: <upstream message>"` |

---

## Tool: `review_code`

Review an arbitrary code snippet provided directly as text.

### Input Schema

| Parameter   | Type   | Required | Description |
|-------------|--------|----------|-------------|
| `code`      | string | **Yes**  | Source code or text content to review. |
| `language`  | string | No       | Language hint for the reviewer (e.g., `"go"`, `"python"`, `"sql"`). |
| `provider`  | string | No       | AI provider. Defaults to first available API key. |
| `model`     | string | No       | Model identifier. Defaults to provider's recommended model. |

### Output (success)

Same `ReviewResult` JSON structure. `file` fields in findings will be empty (no file context).

### Error responses

| Condition | Error message |
|-----------|---------------|
| `code` missing or empty | `"'code' is required and must not be empty"` |
| Provider key missing | `"no API key found; set at least one of: ANTHROPIC_API_KEY, OPENAI_API_KEY, GOOGLE_API_KEY"` |
| Provider call failed | `"provider '<name>' returned error: <upstream message>"` |

---

## Provider Default Models

| Provider    | Default Model           | Environment Variable  |
|-------------|-------------------------|-----------------------|
| `anthropic` | `claude-sonnet-4-6`     | `ANTHROPIC_API_KEY`   |
| `openai`    | `gpt-4o`                | `OPENAI_API_KEY`      |
| `google`    | `gemini-2.0-flash`      | `GOOGLE_API_KEY`      |

Default model overrides via environment variables (optional):

| Variable                       | Applies to  |
|--------------------------------|-------------|
| `MCP_REVIEW_ANTHROPIC_MODEL`   | Anthropic   |
| `MCP_REVIEW_OPENAI_MODEL`      | OpenAI      |
| `MCP_REVIEW_GOOGLE_MODEL`      | Google      |

---

## Server Environment Variables

| Variable                     | Required | Default | Description |
|------------------------------|----------|---------|-------------|
| `ANTHROPIC_API_KEY`          | No*      | —       | Anthropic API key |
| `OPENAI_API_KEY`             | No*      | —       | OpenAI API key |
| `GOOGLE_API_KEY`             | No*      | —       | Google Generative AI API key |
| `LOG_LEVEL`                  | No       | `INFO`  | Logging verbosity: `DEBUG`, `INFO`, `WARN`, `ERROR` |
| `MCP_REVIEW_ANTHROPIC_MODEL` | No       | (above) | Override default Anthropic model |
| `MCP_REVIEW_OPENAI_MODEL`    | No       | (above) | Override default OpenAI model |
| `MCP_REVIEW_GOOGLE_MODEL`    | No       | (above) | Override default Google model |

\* At least one API key MUST be present at runtime or all tool calls will fail.
